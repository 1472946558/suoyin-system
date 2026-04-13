/**
 * 金价数据抓取模块
 * 支持：
 *  1. 国际金价 (XAU/USD) - 多源冗余
 *  2. 罗湖水贝实物金价
 */

// ============================================================
//  国际金价 (XAU/USD) - 多数据源
// ============================================================

/**
 * 源1: MetalpriceAPI (免费层 300次/月)
 * API Key 需要在环境变量 METALPRICE_API_KEY 中配置
 */
async function fetchFromMetalPriceAPI() {
  const apiKey = process.env.METALPRICE_API_KEY;
  if (!apiKey) return null;

  try {
    const res = await fetch(
      `https://api.metalpriceapi.com/v1/latest?api_key=${apiKey}&base=XAU&currencies=USD,CNY`
    );
    if (!res.ok) return null;
    const data = await res.json();
    if (!data.success || !data.rates) return null;
    return {
      source: "MetalpriceAPI",
      xauUsd: (1 / data.rates.USD).toFixed(2),
      xauCny: (1 / data.rates.CNY).toFixed(2),
      timestamp: new Date().toISOString(),
    };
  } catch {
    return null;
  }
}

/**
 * 源2: 金价网 (jinjia.com.cn) 爬取水贝数据
 */
async function fetchFromJinjia() {
  try {
    const res = await fetch("https://www.jinjia.com.cn/shuibei/", {
      headers: {
        "User-Agent":
          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
      },
    });
    if (!res.ok) return null;
    const html = await res.text();

    // 从页面提取价格数据
    const prices = {};

    // 尝试匹配金价表格中的数字
    const goldMatch = html.match(/黄金[^]*?(\d{3,4})\s*元\/克/);
    const barMatch = html.match(/金条[^]*?(\d{3,4})\s*元\/克/);
    const recycleMatch = html.match(/回收[^]*?(\d{3,4})\s*元\/克/);

    if (goldMatch) prices.goldJewelry = parseInt(goldMatch[1]);
    if (barMatch) prices.goldBar = parseInt(barMatch[1]);
    if (recycleMatch) prices.recycle = parseInt(recycleMatch[1]);

    if (Object.keys(prices).length === 0) return null;

    return {
      source: "金价网(jinjia.com.cn)",
      ...prices,
      timestamp: new Date().toISOString(),
    };
  } catch {
    return null;
  }
}

/**
 * 源3: 本地宝水贝金价
 * 页面含表格: 类别 | 今日价格 | 回收价格
 * 行: 黄金, 1198 元/克, 1012 元/克
 */
async function fetchFromBendibao() {
  try {
    const res = await fetch("http://sz.bendibao.com/news/2024315/970345.htm", {
      headers: {
        "User-Agent":
          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
      },
    });
    if (!res.ok) return null;
    const html = await res.text();

    const prices = {};

    // 提取所有价格数字 (xxx 元)
    const allPrices = [...html.matchAll(/(\d{3,4})\s*元/g)].map((m) =>
      parseInt(m[1])
    );

    if (allPrices.length >= 6) {
      // 本地宝表格每行有2个价格(今日+回收)，页面重复显示(桌面+移动)
      // 取前6个: 黄金今日, 黄金回收, 金条今日, 金条回收, 铂金今日, 铂金回收
      prices.goldJewelry = allPrices[0];
      prices.recycle = allPrices[1];
      prices.goldBar = allPrices[2];
    } else if (allPrices.length >= 2) {
      prices.goldJewelry = allPrices[0];
      prices.recycle = allPrices[1];
    }

    if (Object.keys(prices).length === 0) return null;
    return {
      source: "深圳本地宝(bendibao.com)",
      ...prices,
      timestamp: new Date().toISOString(),
    };
  } catch {
    return null;
  }
}

/**
 * 源4: 新浪财经 - 实时国际金价 (备用)
 * 通过新浪财经的行情接口获取 XAU/USD
 */
async function fetchFromSina() {
  try {
    const res = await fetch(
      "https://hq.sinajs.cn/list=hf_XAU",
      {
        headers: {
          Referer: "https://finance.sina.com.cn",
          "User-Agent":
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
        },
      }
    );
    if (!res.ok) return null;
    const text = await res.text();

    // 解析新浪行情格式: var hq_str_hf_XAU="..."
    const match = text.match(/"([^"]+)"/);
    if (!match) return null;

    const parts = match[1].split(",");
    if (parts.length < 3) return null;

    // 新浪外盘行情格式 (hf_XAU):
    // [0]昨收, [1]当前价, [2]开盘, [3]买价, [4]卖价,
    // [5]最低(?), [6]时间, [7]当前价2, [8]昨低(?), ...
    // 注意: 新浪字段命名不完全确定，用安全解析
    const prevClose = parseFloat(parts[0]);
    const currentPrice = parseFloat(parts[1]);
    const openPrice = parseFloat(parts[2]);

    // 从所有数字字段中推断 high/low
    const numericFields = parts.slice(0, 9).map(Number).filter(n => !isNaN(n) && n > 0);
    const high = Math.max(...numericFields);
    const low = Math.min(...numericFields);

    if (isNaN(currentPrice) || currentPrice <= 0) return null;

    const change = currentPrice - prevClose;
    const changePercent = prevClose > 0 ? (change / prevClose) * 100 : 0;

    return {
      source: "新浪财经(hq.sinajs.cn)",
      xauUsd: currentPrice.toFixed(2),
      change: change.toFixed(2),
      changePercent: changePercent.toFixed(2),
      prevClose: prevClose.toFixed(2),
      open: isNaN(openPrice) ? null : openPrice.toFixed(2),
      high: high.toFixed(2),
      low: low.toFixed(2),
      timestamp: new Date().toISOString(),
    };
  } catch {
    return null;
  }
}

/**
 * 源5: 金投网水贝黄金 (mobile)
 */
async function fetchFromCngold() {
  try {
    const res = await fetch("https://m.cngold.org/quote/gjs/swhj_shuibei.html", {
      headers: {
        "User-Agent":
          "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15",
      },
    });
    if (!res.ok) return null;
    const html = await res.text();

    const prices = {};
    // 提取价格
    const patterns = [
      { key: "goldJewelry", re: /水贝黄金[^\d]*?(\d{3,4})\s*元/ },
      { key: "goldBar", re: /金条[^\d]*?(\d{3,4})\s*元/ },
    ];
    for (const { key, re } of patterns) {
      const m = html.match(re);
      if (m) prices[key] = parseInt(m[1]);
    }

    if (Object.keys(prices).length === 0) return null;
    return {
      source: "金投网(cngold.org)",
      ...prices,
      timestamp: new Date().toISOString(),
    };
  } catch {
    return null;
  }
}

// ============================================================
//  聚合接口
// ============================================================

/**
 * 获取国际金价 (XAU/USD) - 多源自动降级
 */
export async function fetchInternationalGold() {
  const sources = [fetchFromSina, fetchFromMetalPriceAPI];

  for (const fn of sources) {
    const result = await fn();
    if (result) return result;
  }

  return {
    source: "all-failed",
    error: "所有国际金价数据源均不可用",
    timestamp: new Date().toISOString(),
  };
}

/**
 * 获取罗湖水贝实物金价 - 多源自动降级
 */
export async function fetchShuibeiGold() {
  const sources = [fetchFromJinjia, fetchFromBendibao, fetchFromCngold];

  for (const fn of sources) {
    const result = await fn();
    if (result) return result;
  }

  return {
    source: "all-failed",
    error: "所有水贝金价数据源均不可用",
    timestamp: new Date().toISOString(),
  };
}

/**
 * 获取所有金价数据（并行抓取）
 */
export async function fetchAllGoldData() {
  const [intl, shuibei] = await Promise.all([
    fetchInternationalGold(),
    fetchShuibeiGold(),
  ]);

  return { intl, shuibei };
}

// ============================================================
//  CLI 入口 (测试用)
// ============================================================

const args = process.argv.slice(2);
if (args.includes("--test") || args.includes("--demo")) {
  console.log("=== 抓取金价数据 ===\n");

  const data = await fetchAllGoldData();

  console.log("【国际金价 XAU/USD】");
  console.log(JSON.stringify(data.intl, null, 2));
  console.log();
  console.log("【罗湖水贝金价】");
  console.log(JSON.stringify(data.shuibei, null, 2));
}
