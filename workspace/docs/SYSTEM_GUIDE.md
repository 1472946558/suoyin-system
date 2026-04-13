# OpenClaw 系统使用指南

## 快速启动

```bash
# 1. 启动调度器
node scheduler/index.js &

# 2. 打开 Dashboard
open dashboard/index.html

# 3. 打开 OpenClaw 网关（如需要）
# Gateway 默认运行在 localhost:18789
```

## 目录结构

```
openclaw/
├── workspace/           # 核心工作空间
│   ├── user/           # 用户身份文件
│   ├── agents/         # Agent 定义文件
│   ├── skills/         # 技能模块
│   ├── dashboard/      # Dashboard 数据
│   ├── memory/         # 每日记忆
│   ├── docs/           # 系统文档
│   ├── prompts/        # Prompt 模板
│   ├── backups/        # 自动备份
│   ├── SOUL.md         # AI 核心人格
│   ├── AGENTS.md       # Agent 框架
│   └── USER.md         # 用户信息
├── dashboard/          # Dashboard 前端
├── scheduler/          # 调度器
├── extensions/         # 插件（飞书等）
└── openclaw.json       # 主配置
```

## 常用操作

### Dashboard
- **总览页**: 查看系统状态、进行中任务、项目进度
- **Agent页**: 查看 Agent 详情、技能、分配的任务
- **任务页**: 新建/删除任务、手动调度、状态流转
- **项目页**: 各项目完成率和进度
- **日志页**: 系统运行日志

### 调度器 API
```bash
# 查看任务
curl http://localhost:18790/api/tasks

# 新建任务
curl -X POST http://localhost:18790/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"任务标题","agent":"golf-ios","priority":"P1","project":"Golf iOS"}'

# 更新任务状态
curl -X PUT http://localhost:18790/api/tasks/T001 \
  -H 'Content-Type: application/json' \
  -d '{"status":"done"}'

# 手动调度
curl -X POST http://localhost:18790/api/schedule
```

### Git 分支
```bash
git branch              # 查看所有分支
git checkout feat/xxx   # 切换分支
git merge feat/xxx      # 合并到当前分支
```

## Agent 路由表

| 任务类型 | Agent |
|----------|-------|
| Golf iOS 开发 | golf-ios-agent |
| Golf 验收 | golf-qa-agent |
| Amazon 运营 | amazon-de-agent |
| GEO 调研 | geo-product-agent |
| Prompt 生成 | prompt-engineer-agent |
| SOP 沉淀 | sop-builder-agent |
| 记忆归档 | memory-agent |
| 平台维护 | openclaw-agent |

---
_最后更新: 2026-04-14_
