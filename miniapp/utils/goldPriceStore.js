/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: goldPriceStore.js
 * 功能描述: 工具函数
 * 作者: 廖心慈
 * 创建日期: 2026-06-06
 */

const { appConfig, request } = require("./apiClient");

const STORAGE_KEY = "gr_gold_reference_prices";

function canUseStorage() {
  return typeof wx !== "undefined" && typeof wx.getStorageSync === "function";
}

function safeNumber(value) {
  const parsed = parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function formatPriceText(price) {
  if (Math.abs(price - Math.round(price)) < 0.005) {
    return String(Math.round(price));
  }
  return price.toFixed(2);
}

function normalizePriceItem(item) {
  if (!item || !item.purity) {
    return null;
  }
  const price = safeNumber(item.price);
  if (price <= 0) {
    return null;
  }
  return {
    purity: String(item.purity),
    price,
    priceText: formatPriceText(price),
    trend: item.trend ? String(item.trend) : "",
    label: item.label ? String(item.label) : "参考金价"
  };
}

function normalizePriceItems(items) {
  return (items || []).map(normalizePriceItem).filter(Boolean);
}

function buildFallbackPrices() {
  const prices = normalizePriceItems(appConfig.referencePrices);
  if (!prices.some(function(item) { return item.purity === "14K"; })) {
    prices.push({
      purity: "14K",
      price: 436,
      priceText: "436",
      trend: "+2",
      label: "低成色 K 金"
    });
  }
  return prices;
}

function readStoredPrices() {
  if (!canUseStorage()) {
    return [];
  }
  try {
    const stored = wx.getStorageSync(STORAGE_KEY);
    return normalizePriceItems(stored && stored.referencePrices);
  } catch (error) {
    return [];
  }
}

let cachedReferencePrices = readStoredPrices();
if (!cachedReferencePrices.length) {
  cachedReferencePrices = buildFallbackPrices();
}

function writeStoredPrices(snapshot) {
  if (!canUseStorage()) {
    return;
  }
  try {
    wx.setStorageSync(STORAGE_KEY, snapshot);
  } catch (error) {
    // Storage failure should not block quoting.
  }
}

function applyReferenceSnapshot(snapshot) {
  const source = snapshot && snapshot.referencePrices ? snapshot : {};
  const prices = normalizePriceItems(source.referencePrices);
  if (!prices.length) {
    return cachedReferencePrices;
  }
  cachedReferencePrices = prices;
  writeStoredPrices({
    source: source.source || "local_reference",
    sourceText: source.sourceText || "",
    baseCnyPerGram: safeNumber(source.baseCnyPerGram),
    updatedAt: source.updatedAt || "",
    referenceNote: source.referenceNote || "",
    referencePrices: cachedReferencePrices
  });
  return cachedReferencePrices;
}

function getCachedReferencePrices() {
  return cachedReferencePrices.slice();
}

function getVisibleReferencePrices(limit) {
  const prices = getCachedReferencePrices();
  if (limit > 0) {
    return prices.slice(0, limit);
  }
  return prices;
}

function getSuggestedPrice(purity) {
  const match = cachedReferencePrices.find(function(item) {
    return item.purity === purity;
  });
  if (match) {
    return match.price;
  }
  const fallback = buildFallbackPrices().find(function(item) {
    return item.purity === purity;
  });
  return fallback ? fallback.price : 742;
}

function refreshReferencePrices() {
  if (!appConfig.endpoints.goldReferencePrices) {
    return Promise.resolve(getCachedReferencePrices());
  }
  return request({
    endpoint: appConfig.endpoints.goldReferencePrices
  }).then(function(payload) {
    const snapshot = payload && payload.data ? payload.data : payload;
    return applyReferenceSnapshot(snapshot);
  }).catch(function() {
    return getCachedReferencePrices();
  });
}

module.exports = {
  applyReferenceSnapshot,
  getCachedReferencePrices,
  getVisibleReferencePrices,
  getSuggestedPrice,
  refreshReferencePrices
};
