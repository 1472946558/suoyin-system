#!/usr/bin/env node

/**
 * OpenClaw Agent Scheduler
 *
 * Minimal scheduler that:
 * 1. Reads task definitions from workspace/dashboard/tasks.json
 * 2. Routes tasks to agents based on agent-task mapping
 * 3. Triggers agent execution via Gateway API
 * 4. Updates task status and serves it to Dashboard
 *
 * Usage:
 *   node scheduler/index.js              # Start scheduler + API server
 *   node scheduler/index.js --once       # Run one cycle and exit
 */

const http = require('http');
const net = require('net');
const fs = require('fs');
const path = require('path');

const ROOT = path.resolve(__dirname, '..');
const DASHBOARD_DIR = path.join(ROOT, 'workspace', 'dashboard');
const PORT = 18790; // Dashboard API port (not gateway)

// ---- Config ----
const GATEWAY = {
  host: 'localhost',
  port: 18789,
  token: '7b069612a41d2cdde5e035c05288c02f01c62c7b11db217c',
};

// ---- State ----
let tasks = [];
let agents = [];
let projects = [];
let runLog = [];
let messages = [];
const MESSAGES_FILE = path.join(DASHBOARD_DIR, 'messages.json');

function log(level, source, message) {
  const entry = {
    time: new Date().toISOString(),
    level,
    source,
    message,
  };
  runLog.push(entry);
  if (runLog.length > 200) runLog.shift();
  const prefix = level === 'ERROR' ? '✗' : level === 'WARN' ? '⚠' : '→';
  console.log(`[${entry.time.slice(11, 19)}] ${prefix} [${source}] ${message}`);
}

// ---- Data I/O ----
function loadData() {
  try {
    const tasksFile = path.join(DASHBOARD_DIR, 'tasks.json');
    const agentsFile = path.join(DASHBOARD_DIR, 'agents.json');

    if (fs.existsSync(tasksFile)) {
      const raw = JSON.parse(fs.readFileSync(tasksFile, 'utf8'));
      tasks = raw.tasks || tasks;
      projects = raw.projects || projects;
    }

    if (fs.existsSync(agentsFile)) {
      const raw = JSON.parse(fs.readFileSync(agentsFile, 'utf8'));
      agents = raw.agents || agents;
    }

    if (fs.existsSync(MESSAGES_FILE)) {
      messages = JSON.parse(fs.readFileSync(MESSAGES_FILE, 'utf8'));
    }

    log('INFO', 'loader', `Loaded ${tasks.length} tasks, ${agents.length} agents, ${messages.length} messages`);
  } catch (e) {
    log('ERROR', 'loader', e.message);
  }
}

function saveMessages() {
  fs.writeFileSync(MESSAGES_FILE, JSON.stringify(messages, null, 2));
}

// ---- Message Parser ----
function parseMessage(raw) {
  const now = new Date().toISOString();
  const lines = raw.split('\n').filter(l => l.trim());

  // Extract keywords
  const keywords = {
    deadline: [],
    requirements: [],
    decisions: [],
    contacts: [],
  };

  // Date patterns
  const datePatterns = [
    /(\d{4}[-/年]\d{1,2}[-/月]\d{1,2}[日号]?)/g,
    /(下个?[周月]\d{0,2}[号日]?)/g,
    /(周[一二三四五六日末])/g,
    /(明天|后天|下周|月底|月底前)/g,
    /(\d{1,2}月\d{1,2}[日号])/g,
  ];

  for (const line of lines) {
    for (const p of datePatterns) {
      const m = line.match(p);
      if (m) keywords.deadline.push(...m);
    }
  }

  // Requirement patterns - broadened for real WeChat conversations
  const reqPatterns = [/需要|要求|必须|要做|帮我|开发|实现|修复|完成|做一下|弄一下|改一下|加上|去掉|改了|修改|调整|更新|升级|问题|bug|BUG|崩了|闪退|卡了|不动|不行|报错|拒绝|审核|图标|登录|功能|按钮|页面|接口|协议|对接|验收|测试|上架|提审|重新|还是|没好|还没|进度|什么时候|多久|能不能|可以不|我改|你改|他改|正常了|搞定了|通过了|没通过|被拒|打回/g];
  for (const line of lines) {
    // Skip lines that are just images/links/mentions
    if (/^\[图片\]$|^\[视频\]$|^\[语音\]$|正在studying.*动态$/.test(line.trim())) continue;
    for (const p of reqPatterns) {
      if (p.test(line)) {
        p.lastIndex = 0;
        keywords.requirements.push(line.trim());
        break;
      }
    }
  }

  // Decision patterns
  const decPatterns = [/确定|决定|就这样|同意|没问题|可以的|OK|ok|确认|不改了|用这个|通过了|正常了|搞定了|好了|可以了|行了|就这样吧|没问题了|改好了|已解决/g];
  for (const line of lines) {
    for (const p of decPatterns) {
      if (p.test(line)) {
        p.lastIndex = 0;
        keywords.decisions.push(line.trim());
        break;
      }
    }
  }

  // Extract issues/blockers
  keywords.blockers = [];
  const blockerPatterns = [/拒绝|被拒|审核不|打回|闪退|崩了|不行|报错|失败|bug|BUG|有问题|没好|还没|卡住|阻塞|不通|连不上|登不上|白屏|黑屏|无响应/g];
  for (const line of lines) {
    for (const p of blockerPatterns) {
      if (p.test(line)) {
        p.lastIndex = 0;
        keywords.blockers.push(line.trim());
        break;
      }
    }
  }

  // Extract contact name (first line often has name in WeChat export)
  const contactMatch = raw.match(/^([^\s:：]{2,10})\s*[\[(:：]/m);
  const contact = contactMatch ? contactMatch[1] : '';

  return {
    id: 'M' + Date.now().toString(36).toUpperCase(),
    raw,
    contact,
    time: now,
    lines: lines.length,
    keywords,
    source: 'wechat_paste',
    project: '',
    tags: [],
  };
}

function saveTasks() {
  const data = {
    tasks,
    task_statuses: [
      { id: 'pending', label: '待开始', color: 'gray' },
      { id: 'in_progress', label: '进行中', color: 'blue' },
      { id: 'review', label: '验收中', color: 'yellow' },
      { id: 'done', label: '已完成', color: 'green' },
      { id: 'blocked', label: '阻塞', color: 'red' },
    ],
    projects: updateProjectStats(),
  };
  fs.writeFileSync(
    path.join(DASHBOARD_DIR, 'tasks.json'),
    JSON.stringify(data, null, 2)
  );
}

function updateProjectStats() {
  return projects.map(p => {
    const pTasks = tasks.filter(t => t.project === p.name);
    return {
      name: p.name,
      total: pTasks.length,
      done: pTasks.filter(t => t.status === 'done').length,
      in_progress: pTasks.filter(t => t.status === 'in_progress').length,
      pending: pTasks.filter(t => t.status === 'pending').length,
    };
  });
}

// ---- Gateway Communication ----
function gatewayRequest(method, path, body) {
  return new Promise((resolve, reject) => {
    const data = body ? JSON.stringify(body) : null;
    const options = {
      hostname: GATEWAY.host,
      port: GATEWAY.port,
      path,
      method,
      headers: {
        Authorization: `Bearer ${GATEWAY.token}`,
        'Content-Type': 'application/json',
      },
    };
    if (data) options.headers['Content-Length'] = Buffer.byteLength(data);

    const req = http.request(options, (res) => {
      let buf = '';
      res.on('data', (c) => (buf += c));
      res.on('end', () => {
        try {
          resolve({ status: res.statusCode, data: JSON.parse(buf) });
        } catch {
          resolve({ status: res.statusCode, data: buf });
        }
      });
    });
    req.on('error', reject);
    if (data) req.write(data);
    req.end();
  });
}

async function checkGatewayHealth() {
  // OpenClaw Gateway is a UI server, not REST API
  // Check TCP connectivity instead of HTTP health endpoint
  return new Promise((resolve) => {
    const socket = net.createConnection({ host: GATEWAY.host, port: GATEWAY.port });
    socket.setTimeout(3000);
    socket.on('connect', () => { socket.destroy(); resolve(true); });
    socket.on('timeout', () => { socket.destroy(); resolve(false); });
    socket.on('error', () => { socket.destroy(); resolve(false); });
  });
}

async function triggerAgent(agentId, task) {
  const agent = agents.find(a => a.id === agentId);
  if (!agent) {
    log('ERROR', 'scheduler', `Agent ${agentId} not found`);
    return false;
  }

  log('INFO', 'scheduler', `Triggering ${agent.name} for task ${task.id}: ${task.title}`);

  // Try to send to gateway
  try {
    const payload = {
      agent: agentId,
      task_id: task.id,
      title: task.title,
      description: task.description || '',
      acceptance: task.acceptance || [],
    };
    const res = await gatewayRequest('POST', '/api/task', payload);
    if (res.status === 200 || res.status === 201) {
      log('INFO', 'gateway', `Task ${task.id} dispatched to ${agent.name}`);
      return true;
    } else {
      log('WARN', 'gateway', `Gateway returned ${res.status}, running in offline mode`);
      return false;
    }
  } catch (e) {
    log('WARN', 'gateway', `Gateway unreachable: ${e.message}, running in offline mode`);
    return false;
  }
}

// ---- Scheduler Logic ----
function pickNextTask() {
  // Priority order: P0 > P1 > P2 > P3
  // Only pick pending tasks whose blockers are done
  const priorityOrder = { P0: 0, P1: 1, P2: 2, P3: 3 };
  const pending = tasks
    .filter((t) => t.status === 'pending')
    .sort((a, b) => (priorityOrder[a.priority] || 9) - (priorityOrder[b.priority] || 9));

  for (const task of pending) {
    // Check blocker
    if (task.blocker) {
      const blocker = tasks.find((t) => t.id === task.blocker);
      if (blocker && blocker.status !== 'done') continue;
    }
    // Check if agent already has an in_progress task
    const agentBusy = tasks.some(
      (t) => t.agent === task.agent && t.status === 'in_progress'
    );
    if (agentBusy) continue;

    return task;
  }
  return null;
}

async function runScheduleCycle() {
  log('INFO', 'scheduler', '--- Schedule cycle start ---');

  // Update project stats
  projects = updateProjectStats();

  // Pick and dispatch next task
  const task = pickNextTask();
  if (task) {
    task.status = 'in_progress';
    task.updated = new Date().toISOString().slice(0, 10);
    log('INFO', 'scheduler', `Picked task ${task.id}: ${task.title} → ${task.agent}`);
    await triggerAgent(task.agent, task);
  } else {
    log('INFO', 'scheduler', 'No pending tasks to dispatch');
  }

  saveTasks();
  log('INFO', 'scheduler', '--- Schedule cycle end ---');
}

// ---- Task Commands (exposed via API) ----
function addTask(taskDef) {
  const id = 'T' + String(tasks.length + 1).padStart(3, '0');
  const task = {
    id,
    title: taskDef.title,
    agent: taskDef.agent,
    project: taskDef.project || '',
    priority: taskDef.priority || 'P2',
    status: 'pending',
    created: new Date().toISOString().slice(0, 10),
    updated: new Date().toISOString().slice(0, 10),
    description: taskDef.description || '',
    acceptance: taskDef.acceptance || [],
    blocker: taskDef.blocker || null,
  };
  tasks.push(task);
  saveTasks();
  log('INFO', 'api', `Added task ${id}: ${task.title}`);
  return task;
}

function updateTask(taskId, updates) {
  const task = tasks.find((t) => t.id === taskId);
  if (!task) return null;
  Object.assign(task, updates, { updated: new Date().toISOString().slice(0, 10) });
  saveTasks();
  log('INFO', 'api', `Updated task ${taskId}: ${JSON.stringify(updates)}`);
  return task;
}

// ---- HTTP API Server ----
function createServer() {
  const server = http.createServer(async (req, res) => {
    // CORS
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');
    if (req.method === 'OPTIONS') {
      res.writeHead(204);
      res.end();
      return;
    }

    const url = new URL(req.url, `http://localhost:${PORT}`);
    const route = url.pathname;
    let body = '';

    if (req.method === 'POST' || req.method === 'PUT') {
      for await (const chunk of req) body += chunk;
    }

    try {
      // Routes
      if (route === '/api/agents' && req.method === 'GET') {
        json(res, { agents });
      } else if (route === '/api/tasks' && req.method === 'GET') {
        json(res, { tasks, projects: updateProjectStats() });
      } else if (route === '/api/tasks' && req.method === 'POST') {
        const data = JSON.parse(body);
        const task = addTask(data);
        json(res, task, 201);
      } else if (route.match(/^\/api\/tasks\/T\d+$/) && req.method === 'PUT') {
        const id = route.split('/').pop();
        const data = JSON.parse(body);
        const task = updateTask(id, data);
        if (task) json(res, task);
        else error(res, 'Task not found', 404);
      } else if (route === '/api/schedule' && req.method === 'POST') {
        await runScheduleCycle();
        json(res, { ok: true });
      } else if (route === '/api/gateway' && req.method === 'GET') {
        const healthy = await checkGatewayHealth();
        json(res, { gateway: healthy ? 'online' : 'offline' });
      } else if (route === '/api/logs' && req.method === 'GET') {
        json(res, { logs: runLog.slice(-50) });
      } else if (route === '/api/messages' && req.method === 'GET') {
        json(res, { messages: messages.slice(-100) });
      } else if (route === '/api/messages' && req.method === 'POST') {
        const data = JSON.parse(body);
        if (!data.raw) { error(res, 'Missing raw text'); return; }
        const msg = parseMessage(data.raw);
        if (data.contact) msg.contact = data.contact;
        if (data.project) msg.project = data.project;
        if (data.tags) msg.tags = data.tags;
        messages.push(msg);
        saveMessages();
        log('INFO', 'api', `New message from ${msg.contact || 'unknown'}: ${msg.lines} lines, ${msg.keywords.requirements.length} requirements, ${msg.keywords.deadline.length} deadlines`);
        json(res, msg, 201);
      } else if (route.match(/^\/api\/messages\//) && req.method === 'PUT') {
        const id = route.split('/').pop();
        const msg = messages.find(m => m.id === id);
        if (!msg) { error(res, 'Not found', 404); return; }
        const data = JSON.parse(body);
        if (data.project) msg.project = data.project;
        if (data.tags) msg.tags = data.tags;
        if (data.contact) msg.contact = data.contact;
        saveMessages();
        json(res, msg);
      } else if (route === '/api/messages/parse' && req.method === 'POST') {
        // Preview parse without saving
        const data = JSON.parse(body);
        if (!data.raw) { error(res, 'Missing raw text'); return; }
        const msg = parseMessage(data.raw);
        json(res, msg);
      } else if (route === '/health' && req.method === 'GET') {
        json(res, { status: 'ok', uptime: process.uptime() });
      } else {
        error(res, 'Not found', 404);
      }
    } catch (e) {
      error(res, e.message, 500);
    }
  });

  return server;
}

function json(res, data, status = 200) {
  res.writeHead(status, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify(data));
}

function error(res, msg, status = 400) {
  res.writeHead(status, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify({ error: msg }));
}

// ---- Main ----
async function main() {
  const args = process.argv.slice(2);
  const once = args.includes('--once');

  // Generate initial JSON data from YAML if needed
  generateJsonData();

  loadData();

  if (once) {
    await runScheduleCycle();
    process.exit(0);
  }

  // Start server
  const server = createServer();
  server.listen(PORT, () => {
    log('INFO', 'server', `Dashboard API listening on http://localhost:${PORT}`);
    log('INFO', 'server', `Dashboard UI at file://${path.join(ROOT, 'dashboard', 'index.html')}`);
  });

  // Schedule cycle every 60 seconds
  setInterval(runScheduleCycle, 60000);

  // Reload data every 30 seconds
  setInterval(loadData, 30000);
}

function generateJsonData() {
  const agentsJson = path.join(DASHBOARD_DIR, 'agents.json');
  const tasksJson = path.join(DASHBOARD_DIR, 'tasks.json');

  if (!fs.existsSync(agentsJson)) {
    const data = {
      agents: [
        { id: 'commander', name: 'Commander Agent', role: '总协调', status: 'active', emoji: '🎯', skills: [] },
        { id: 'strategist', name: 'Strategist Agent', role: '优先级排序', status: 'active', emoji: '📊', skills: [] },
        { id: 'golf-ios', name: 'Golf iOS Agent', role: 'iOS Native + Unity + BLE', status: 'active', emoji: '⛳', skills: ['golf_ios_acceptance', 'golf_ble_protocol', 'golf_unity_json_csv'] },
        { id: 'golf-qa', name: 'Golf QA Agent', role: '验收检查', status: 'active', emoji: '✅', skills: ['golf_ios_acceptance'] },
        { id: 'amazon-de', name: 'Amazon DE Agent', role: 'Amazon.de 运营', status: 'active', emoji: '📦', skills: ['amazon_product_selection', 'amazon_ads_optimization', 'amazon_listing_review', 'amazon_profit_judgement'] },
        { id: 'geo-research', name: 'GEO Research Agent', role: '竞品调研', status: 'standby', emoji: '🔍', skills: ['geo_competitor_analysis'] },
        { id: 'prompt-engineer', name: 'Prompt Engineer Agent', role: 'Prompt 生成', status: 'active', emoji: '✍️', skills: ['cursor_prompt_builder'] },
        { id: 'sop-builder', name: 'SOP Builder Agent', role: 'SOP 沉淀', status: 'standby', emoji: '📋', skills: ['sop_generator'] },
        { id: 'memory', name: 'Memory Agent', role: '记忆归档', status: 'active', emoji: '🧠', skills: ['meeting_summary'] },
        { id: 'delivery', name: 'Delivery Agent', role: '项目交付', status: 'active', emoji: '🚀', skills: [] },
        { id: 'research', name: 'Research Agent', role: '调研', status: 'active', emoji: '🔬', skills: [] },
        { id: 'security', name: 'Security Agent', role: '安全校验', status: 'active', emoji: '🛡️', skills: [] },
        { id: 'backup', name: 'Backup Agent', role: '备份', status: 'active', emoji: '💾', skills: [] },
        { id: 'openclaw', name: 'OpenClaw Agent', role: '平台维护', status: 'active', emoji: '⚙️', skills: ['openclaw_skill'] },
      ],
    };
    fs.mkdirSync(DASHBOARD_DIR, { recursive: true });
    fs.writeFileSync(agentsJson, JSON.stringify(data, null, 2));
  }

  if (!fs.existsSync(tasksJson)) {
    const data = {
      tasks: [
        { id: 'T001', title: '修复 Unity CSV/JSON 协议', agent: 'golf-ios', project: 'Golf iOS', priority: 'P0', status: 'in_progress', created: '2026-04-13', updated: '2026-04-13', description: 'Unity 与 Native iOS 之间数据交换协议修复', acceptance: ['JSON 字段名与 Android 一致', 'CSV 格式可正确解析'], blocker: null },
        { id: 'T002', title: '完成相机模式 UI 对齐', agent: 'golf-ios', project: 'Golf iOS', priority: 'P1', status: 'pending', created: '2026-04-13', updated: '2026-04-13', description: '相机模式 UI 对齐 Android', acceptance: ['布局像素级对齐', '交互流程一致'], blocker: 'T001' },
        { id: 'T003', title: '完成二期合同需求拆解', agent: 'golf-ios', project: 'Golf iOS', priority: 'P1', status: 'pending', created: '2026-04-13', updated: '2026-04-13', description: '拆解二期合同功能需求', acceptance: ['需求列表完整', '每个需求有验收标准'], blocker: null },
        { id: 'T004', title: '609 圆灯差评分析', agent: 'amazon-de', project: 'Amazon DE', priority: 'P1', status: 'pending', created: '2026-04-13', updated: '2026-04-13', description: '分析差评原因', acceptance: ['差评分类统计', '根因分析'], blocker: null },
        { id: 'T005', title: '德国站选品调研', agent: 'amazon-de', project: 'Amazon DE', priority: 'P2', status: 'pending', created: '2026-04-13', updated: '2026-04-13', description: '按选品框架评估新品', acceptance: ['至少 3 个候选品', '利润测算'], blocker: null },
        { id: 'T006', title: '补齐系统基础架构', agent: 'openclaw', project: 'OpenClaw', priority: 'P1', status: 'in_progress', created: '2026-04-13', updated: '2026-04-13', description: 'USER/Agent/Skill/Dashboard 四层补齐', acceptance: ['3 个 USER 文件', '全部 Agent', '核心 Skills'], blocker: null },
      ],
      task_statuses: [
        { id: 'pending', label: '待开始', color: 'gray' },
        { id: 'in_progress', label: '进行中', color: 'blue' },
        { id: 'review', label: '验收中', color: 'yellow' },
        { id: 'done', label: '已完成', color: 'green' },
        { id: 'blocked', label: '阻塞', color: 'red' },
      ],
      projects: [
        { name: 'Golf iOS', total: 3, done: 0, pending: 2, in_progress: 1 },
        { name: 'Amazon DE', total: 2, done: 0, pending: 2, in_progress: 0 },
        { name: 'OpenClaw', total: 1, done: 0, pending: 0, in_progress: 1 },
      ],
    };
    fs.writeFileSync(tasksJson, JSON.stringify(data, null, 2));
  }
}

main().catch((e) => {
  console.error('Fatal:', e);
  process.exit(1);
});
