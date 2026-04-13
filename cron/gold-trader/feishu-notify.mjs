/**
 * 飞书消息推送模块 - App API 方式
 *
 * 使用飞书应用凭证 (App ID + App Secret) 获取 token，
 * 通过 im/v1/messages API 发送消息到指定群聊。
 *
 * 环境变量：
 *   FEISHU_GOLD_APP_ID     - 应用 App ID
 *   FEISHU_GOLD_APP_SECRET - 应用 App Secret
 *   FEISHU_GOLD_CHAT_ID    - 目标群聊 ID（可选，不传则发到第一个群）
 *
 * 或者也可以直接传入 webhook URL：
 *   FEISHU_GOLD_WEBHOOK    - 群自定义机器人 Webhook
 */

const FEISHU_BASE = "https://open.feishu.cn/open-apis";

// ============================================================
//  Token 管理
// ============================================================

let cachedToken = null;
let tokenExpireAt = 0;

async function getTenantAccessToken() {
  if (cachedToken && Date.now() < tokenExpireAt) {
    return cachedToken;
  }

  const appId =
    process.env.FEISHU_GOLD_APP_ID || "cli_a95640c2e0f8dbc9";
  const appSecret =
    process.env.FEISHU_GOLD_APP_SECRET ||
    "CtlGEmyqo6oP0VWFMpUVuftKhkeHS4Ub";

  const res = await fetch(
    `${FEISHU_BASE}/auth/v3/tenant_access_token/internal`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ app_id: appId, app_secret: appSecret }),
    }
  );

  const data = await res.json();
  if (data.code !== 0) {
    throw new Error(`获取 token 失败: ${data.msg}`);
  }

  cachedToken = data.tenant_access_token;
  // 提前 5 分钟过期
  tokenExpireAt = Date.now() + (data.expire - 300) * 1000;
  return cachedToken;
}

// ============================================================
//  群聊管理
// ============================================================

/**
 * 获取机器人所在的群列表
 */
export async function listChats() {
  const token = await getTenantAccessToken();
  const res = await fetch(
    `${FEISHU_BASE}/im/v1/chats?page_size=50`,
    {
      headers: { Authorization: `Bearer ${token}` },
    }
  );
  const data = await res.json();
  if (data.code !== 0) {
    throw new Error(`获取群列表失败: ${data.msg}`);
  }
  return data.data?.items || [];
}

/**
 * 自动找到目标群聊 ID
 * 如果设置了 FEISHU_GOLD_CHAT_ID 则直接用，否则取第一个群
 */
export async function resolveChatId() {
  const envChatId = process.env.FEISHU_GOLD_CHAT_ID;
  if (envChatId) return envChatId;

  const chats = await listChats();
  if (chats.length === 0) {
    throw new Error("机器人未加入任何群聊，请先把机器人拉入群");
  }
  // 返回第一个群
  return chats[0].chat_id;
}

// ============================================================
//  消息发送 - App API
// ============================================================

/**
 * 发送文本消息
 */
export async function sendText(chatId, text) {
  const token = await getTenantAccessToken();
  const res = await fetch(
    `${FEISHU_BASE}/im/v1/messages?receive_id_type=chat_id`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        receive_id: chatId,
        msg_type: "text",
        content: JSON.stringify({ text }),
      }),
    }
  );
  return handleSendResponse(res);
}

/**
 * 发送富文本消息
 */
export async function sendPost(chatId, title, content) {
  const token = await getTenantAccessToken();
  const res = await fetch(
    `${FEISHU_BASE}/im/v1/messages?receive_id_type=chat_id`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        receive_id: chatId,
        msg_type: "post",
        content: JSON.stringify({
          zh_cn: {
            title,
            content: [[{ tag: "text", text: content }]],
          },
        }),
      }),
    }
  );
  return handleSendResponse(res);
}

/**
 * 发送交互式卡片 (Markdown)
 */
export async function sendCard(chatId, title, markdownContent) {
  const token = await getTenantAccessToken();
  const card = {
    config: { wide_screen_mode: true },
    header: {
      title: { tag: "plain_text", content: title },
      template: "blue",
    },
    elements: [
      {
        tag: "markdown",
        content: markdownContent,
      },
    ],
  };

  const res = await fetch(
    `${FEISHU_BASE}/im/v1/messages?receive_id_type=chat_id`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        receive_id: chatId,
        msg_type: "interactive",
        content: JSON.stringify(card),
      }),
    }
  );
  return handleSendResponse(res);
}

// ============================================================
//  黄金报告专用发送
// ============================================================

/**
 * 发送黄金交易日报卡片
 */
export async function sendGoldPriceCard(goldData) {
  const { intl, shuibei, strategy } = goldData;

  // 国际金价部分
  let intlSection = "";
  if (intl.xauUsd) {
    const changeSymbol =
      parseFloat(intl.change) > 0
        ? "🔺"
        : parseFloat(intl.change) < 0
        ? "🔻"
        : "➖";
    intlSection = `
## 国际金价 XAU/USD
- **当前价格**: $${intl.xauUsd} / 盎司 ${changeSymbol}`;
    if (intl.change) {
      intlSection += `\n- **涨跌幅**: ${parseFloat(intl.change) > 0 ? "+" : ""}${intl.change} (${intl.changePercent}%)`;
    }
    if (intl.high && intl.low) {
      intlSection += `\n- **日内波幅**: $${intl.low} - $${intl.high}`;
    }
    if (intl.prevClose) {
      intlSection += `\n- **昨收**: $${intl.prevClose}`;
    }
    intlSection += `\n- **数据源**: ${intl.source}`;
  } else {
    intlSection = `\n## 国际金价\n⚠️ ${intl.error || "数据暂不可用"}`;
  }

  // 水贝金价部分
  let shuibeiSection = "";
  if (shuibei.goldJewelry) {
    shuibeiSection = `

## 罗湖水贝实物金价
- **首饰金**: ¥${shuibei.goldJewelry} / 克
- **金条**: ¥${shuibei.goldBar || "N/A"} / 克
- **回收价**: ¥${shuibei.recycle || "N/A"} / 克
- **数据源**: ${shuibei.source}`;
  } else {
    shuibeiSection = `\n## 罗湖水贝金价\n⚠️ ${shuibei.error || "数据暂不可用"}`;
  }

  // 策略部分
  let strategySection = "";
  if (strategy) {
    strategySection = `

${strategy}`;
  }

  const now = new Date();
  const timeStr = now.toLocaleString("zh-CN", {
    timeZone: "Asia/Shanghai",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });

  const title = `黄金交易日报 | ${timeStr}`;
  const markdown = `${intlSection}${shuibeiSection}${strategySection}`;

  const chatId = await resolveChatId();
  return sendCard(chatId, title, markdown);
}

// ============================================================
//  Webhook 方式 (降级备用)
// ============================================================

/**
 * 通过 Webhook 发送卡片 (降级方案)
 */
export async function sendGoldPriceCardViaWebhook(webhookUrl, goldData) {
  const { intl, shuibei, strategy } = goldData;

  let text = "";
  if (intl.xauUsd) text += `国际金价: $${intl.xauUsd}\n`;
  if (intl.change) text += `涨跌: ${intl.change} (${intl.changePercent}%)\n`;
  if (shuibei.goldJewelry) text += `水贝首饰金: ¥${shuibei.goldJewelry}/克\n`;
  if (shuibei.goldBar) text += `水贝金条: ¥${shuibei.goldBar}/克\n`;
  if (shuibei.recycle) text += `水贝回收: ¥${shuibei.recycle}/克\n`;
  if (strategy) text += `\n${strategy}`;

  const body = {
    msg_type: "interactive",
    card: {
      header: {
        title: { tag: "plain_text", content: "黄金交易日报" },
        template: "blue",
      },
      elements: [{ tag: "markdown", content: text }],
    },
  };

  const res = await fetch(webhookUrl, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return res.json();
}

// ============================================================
//  工具函数
// ============================================================

async function handleSendResponse(res) {
  const data = await res.json();
  if (data.code === 0) {
    console.log("[Feishu] 消息发送成功");
    return { ok: true, data };
  } else {
    console.error("[Feishu] 发送失败:", data.msg);
    return { ok: false, reason: data.msg, code: data.code };
  }
}
