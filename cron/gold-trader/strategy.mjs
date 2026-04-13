/**
 * 交易策略分析模块
 *
 * 基于技术指标和市场信息生成短线交易建议
 * 注意：这是辅助参考工具，不构成投资建议
 */

/**
 * 根据金价数据生成交易策略分析
 *
 * @param {object} intlData - 国际金价数据
 * @param {object} shuibeiData - 水贝金价数据
 * @param {object} [historicalData] - 可选的历史数据
 * @returns {string} Markdown 格式的策略分析
 */
export function generateStrategy(intlData, shuibeiData, historicalData) {
  const now = new Date();
  const shanghaiTime = now.toLocaleString("zh-CN", { timeZone: "Asia/Shanghai" });
  const dayOfWeek = now.toLocaleDateString("zh-CN", { timeZone: "Asia/Shanghai", weekday: "long" });

  const sections = [];

  // ===== 行情概览 =====
  sections.push(buildMarketOverview(intlData, shuibeiData, shanghaiTime, dayOfWeek));

  // ===== 价差分析 =====
  if (intlData.xauUsd && shuibeiData.goldBar) {
    sections.push(buildSpreadAnalysis(intlData, shuibeiData));
  }

  // ===== 技术面分析框架 =====
  if (intlData.xauUsd) {
    sections.push(buildTechnicalAnalysis(intlData, historicalData));
  }

  // ===== 关键价位 =====
  if (intlData.xauUsd) {
    sections.push(buildKeyLevels(intlData));
  }

  // ===== 交易建议 =====
  if (intlData.xauUsd) {
    sections.push(buildTradingAdvice(intlData));
  }

  // ===== 风控提示 =====
  sections.push(buildRiskWarning());

  return sections.join("\n\n");
}

function buildMarketOverview(intlData, shuibeiData, time, dayOfWeek) {
  let content = `### 行情概览 (${dayOfWeek} ${time})\n`;

  if (intlData.xauUsd) {
    const trend = intlData.change > 0 ? "上涨" : intlData.change < 0 ? "下跌" : "持平";
    content += `- **国际金价**: $${intlData.xauUsd}/盎司 (${trend})\n`;
    if (intlData.change) {
      content += `- **涨跌额**: ${intlData.change > 0 ? "+" : ""}$${intlData.change}\n`;
    }
  }

  if (shuibeiData.goldJewelry) {
    content += `- **水贝首饰金**: ¥${shuibeiData.goldJewelry}/克\n`;
    if (shuibeiData.goldBar) {
      content += `- **水贝金条**: ¥${shuibeiData.goldBar}/克\n`;
    }
    if (shuibeiData.recycle) {
      content += `- **水贝回收价**: ¥${shuibeiData.recycle}/克\n`;
    }
  }

  return content.trim();
}

function buildSpreadAnalysis(intlData, shuibeiData) {
  // 1盎司 = 31.1035 克
  const ozToGram = 31.1035;
  const intlPerGramUsd = parseFloat(intlData.xauUsd) / ozToGram;

  // 粗略使用 7.25 汇率（实际应动态获取）
  const usdCnyRate = 7.25;
  const intlPerGramCny = intlPerGramUsd * usdCnyRate;

  const barPremium = shuibeiData.goldBar
    ? (shuibeiData.goldBar - intlPerGramCny).toFixed(0)
    : "N/A";
  const jewelryPremium = shuibeiData.goldJewelry
    ? (shuibeiData.goldJewelry - intlPerGramCny).toFixed(0)
    : "N/A";

  return `### 价差分析 (汇率参考: 1 USD = ${usdCnyRate} CNY)
- **国际金价折合**: ¥${intlPerGramCny.toFixed(0)}/克
- **金条溢价**: ¥${barPremium}/克
- **首饰金溢价**: ¥${jewelryPremium}/克
- **首饰-金条价差**: ¥${shuibeiData.goldJewelry && shuibeiData.goldBar ? (shuibeiData.goldJewelry - shuibeiData.goldBar) : "N/A"}/克 (工费空间)`;
}

function buildTechnicalAnalysis(intlData, _histData) {
  const price = parseFloat(intlData.xauUsd);
  const change = parseFloat(intlData.change || 0);
  const changePct = parseFloat(intlData.changePercent || 0);

  // 动态计算关键价位（基于当前价的百分比）
  const strongResist = (price * 1.04).toFixed(0);
  const weakResist = (price * 1.02).toFixed(0);
  const weakSupport = (price * 0.98).toFixed(0);
  const strongSupport = (price * 0.96).toFixed(0);

  // 趋势判断
  let trend;
  if (changePct > 1.0) {
    trend = "强势上行，多头主导";
  } else if (changePct > 0.3) {
    trend = "偏多震荡，短线看涨";
  } else if (changePct > -0.3) {
    trend = "窄幅震荡，方向待明";
  } else if (changePct > -1.0) {
    trend = "偏空震荡，短线看跌";
  } else {
    trend = "强势下行，空头主导";
  }

  const analysis = [];
  analysis.push("### 技术面分析");
  analysis.push("");
  analysis.push(`- **趋势**: ${trend}`);
  analysis.push(`- **日内波幅**: ${intlData.high && intlData.low ? `$${intlData.low} - $${intlData.high}` : "暂无数据"}`);
  analysis.push(`- **关键阻力**: $${weakResist}, $${strongResist}`);
  analysis.push(`- **关键支撑**: $${weakSupport}, $${strongSupport}`);
  analysis.push("");
  analysis.push("> 注：精确技术指标(MA/MACD/RSI/布林带/KDJ)需接入K线历史数据");

  return analysis.join("\n");
}

function buildKeyLevels(intlData) {
  const price = parseFloat(intlData.xauUsd);

  // 动态计算关键价位
  const range = price * 0.02; // 2% 区间
  const support1 = (price - range).toFixed(0);
  const support2 = (price - range * 2).toFixed(0);
  const resist1 = (price + range).toFixed(0);
  const resist2 = (price + range * 2).toFixed(0);

  // 止损止盈建议（短线 0.5%-1.5%）
  const stopLoss = (price * 0.985).toFixed(0); // 1.5% 止损
  const takeProfit1 = (price * 1.01).toFixed(0); // 1% 止盈1
  const takeProfit2 = (price * 1.02).toFixed(0); // 2% 止盈2

  return `### 关键价位
| 类型 | 价位 |
|------|------|
| 强阻力 | $${resist2} |
| 弱阻力 | $${resist1} |
| **当前价** | **$${price}** |
| 弱支撑 | $${support1} |
| 强支撑 | $${support2} |

### 短线参考点位
- **止损参考**: $${stopLoss} (-1.5%)
- **止盈1**: $${takeProfit1} (+1.0%)
- **止盈2**: $${takeProfit2} (+2.0%)
- **盈亏比**: 1:1.3 ~ 1:1.5`;
}

function buildTradingAdvice(intlData) {
  const price = parseFloat(intlData.xauUsd);
  const change = parseFloat(intlData.change || 0);
  const changePct = parseFloat(intlData.changePercent || 0);

  let direction = "观望";
  let confidence = "低";
  let advice = "";

  if (changePct > 1.5) {
    direction = "强势上涨";
    confidence = "中";
    advice = `大涨行情，注意追高风险。建议等回踩 $${(price * 0.995).toFixed(0)} 附近再考虑多单，严格止损。`;
  } else if (changePct > 0.5) {
    direction = "偏多";
    confidence = "中";
    advice = `温和上涨，可轻仓试多。入场参考 $${(price * 0.997).toFixed(0)}，止损 $${(price * 0.985).toFixed(0)}。`;
  } else if (changePct > -0.5) {
    direction = "震荡";
    confidence = "低";
    advice = `窄幅震荡，方向不明。建议观望为主，等待突破 $${(price * 1.01).toFixed(0)} 或跌破 $${(price * 0.99).toFixed(0)} 后再入场。`;
  } else if (changePct > -1.5) {
    direction = "偏空";
    confidence = "中";
    advice = `温和下跌，可轻仓试空。入场参考 $${(price * 1.003).toFixed(0)}，止损 $${(price * 1.015).toFixed(0)}。`;
  } else {
    direction = "强势下跌";
    confidence = "中";
    advice = `大跌行情，谨慎做空。建议等反弹 $${(price * 1.005).toFixed(0)} 附近再考虑空单，严格止损。`;
  }

  return `### 交易建议
- **方向判断**: ${direction}
- **信心等级**: ${confidence}
- **操作建议**: ${advice}

> 以上为程序化参考建议，不构成投资建议。实盘交易请结合自身判断和风控规则。`;
}

function buildRiskWarning() {
  const now = new Date();
  const month = now.getMonth() + 1;
  const day = now.getDate();

  // 检查是否临近重要数据发布
  let specialNote = "";
  // 每月第一个周五 - 非农数据
  if (day <= 7 && now.getDay() === 5) {
    specialNote = "\n⚠️ **今日为非农数据发布日，波动可能异常放大，建议减小仓位！**";
  }
  // 每月中旬 - CPI
  if (day >= 10 && day <= 15) {
    specialNote += "\n📅 月中关注美国 CPI 数据发布";
  }

  return `### 风控提醒
- 单笔最大亏损 ≤ 总资金 2%
- 日最大亏损 ≤ 总资金 5%
- 连续3次亏损 → 暂停交易
- 重大数据公布前后30分钟不建议开仓${specialNote}`;
}
