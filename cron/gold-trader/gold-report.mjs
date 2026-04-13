#!/usr/bin/env node
/**
 * 黄金交易日报 - 主入口
 *
 * 功能：
 *   1. 抓取国际金价 (XAU/USD)
 *   2. 抓取罗湖水贝实物金价
 *   3. 生成交易策略分析
 *   4. 通过飞书 App API 推送到群聊
 *
 * 用法：
 *   node gold-report.mjs                    # 完整报告 + 飞书推送
 *   node gold-report.mjs --dry-run          # 仅打印，不推送
 *   node gold-report.mjs --list-chats       # 列出机器人所在群聊
 *
 * 环境变量（可选，不设则用内置凭证）：
 *   FEISHU_GOLD_CHAT_ID    - 目标群聊 chat_id
 */

import { fetchAllGoldData } from "./fetch-gold.mjs";
import {
  sendGoldPriceCard,
  sendText,
  listChats,
  resolveChatId,
} from "./feishu-notify.mjs";
import { generateStrategy } from "./strategy.mjs";

const args = process.argv.slice(2);
const isDryRun = args.includes("--dry-run");

async function main() {
  console.log("========================================");
  console.log("  黄金交易日报 - Gold Trader Report");
  console.log(
    "  " +
      new Date().toLocaleString("zh-CN", { timeZone: "Asia/Shanghai" })
  );
  console.log("========================================\n");

  // 特殊命令: 列出群聊
  if (args.includes("--list-chats")) {
    console.log("正在获取群聊列表...\n");
    try {
      const chats = await listChats();
      if (chats.length === 0) {
        console.log("机器人未加入任何群聊。");
        console.log("请先在飞书中把「金价提醒」机器人拉入群。");
      } else {
        console.log(`找到 ${chats.length} 个群聊：\n`);
        for (const chat of chats) {
          console.log(`  名称: ${chat.name}`);
          console.log(`  chat_id: ${chat.chat_id}`);
          console.log("");
        }
      }
    } catch (err) {
      console.error("获取群聊失败:", err.message);
      console.log(
        "\n需要在飞书开放平台开通以下权限："
      );
      console.log("  - im:chat:readonly (获取群信息)");
      console.log("  - im:message (发送消息)");
      console.log(
        "\n请前往: https://open.feishu.cn/app/cli_a95640c2e0f8dbc9/auth"
      );
    }
    return;
  }

  // Step 1: 抓取金价数据
  console.log("[1/3] 抓取金价数据...");
  const { intl, shuibei } = await fetchAllGoldData();

  console.log(
    "  国际金价:",
    intl.xauUsd ? `$${intl.xauUsd}` : "获取失败"
  );
  console.log("  数据源:", intl.source);
  console.log(
    "  水贝金价:",
    shuibei.goldJewelry ? `¥${shuibei.goldJewelry}/克` : "获取失败"
  );
  console.log("  数据源:", shuibei.source);

  // Step 2: 生成策略分析
  console.log("\n[2/3] 生成交易策略...");
  const strategy = generateStrategy(intl, shuibei);

  if (isDryRun) {
    console.log("\n========================================");
    console.log("  [DRY RUN] 报告预览:");
    console.log("========================================\n");
    console.log(strategy);
    console.log("\n========================================");
    return;
  }

  // Step 3: 推送到飞书
  console.log("\n[3/3] 推送到飞书...");

  try {
    const result = await sendGoldPriceCard({ intl, shuibei, strategy });

    if (result.ok) {
      console.log("✅ 推送成功！");
    } else {
      console.log("❌ 卡片推送失败:", result.reason);
      console.log("   尝试发送纯文本...");
      try {
        const chatId = await resolveChatId();
        const textResult = await sendText(
          chatId,
          `【黄金日报】\n国际金价: $${intl.xauUsd || "N/A"}\n水贝金价: ¥${shuibei.goldJewelry || "N/A"}/克\n\n${strategy}`
        );
        if (textResult.ok) {
          console.log("✅ 纯文本推送成功！");
        }
      } catch (fallbackErr) {
        console.error("   纯文本也失败:", fallbackErr.message);
      }
    }
  } catch (err) {
    console.error("❌ 推送失败:", err.message);
    console.log("\n报告内容：\n");
    console.log(strategy);
  }
}

main().catch((err) => {
  console.error("❌ 执行失败:", err);
  process.exit(1);
});
