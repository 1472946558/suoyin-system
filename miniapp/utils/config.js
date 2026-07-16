/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: config.js
 * 功能描述: 配置模块
 * 作者: 廖心慈
 * 创建日期: 2026-06-05
 */

const RUNTIME_CONFIG_KEY = "gr_runtime_config";
const PRODUCTION_API_BASE_URL = "https://jinjiangguan.com";
const PLACEHOLDER_API_BASE_URLS = ["https://example.com", "http://example.com"];

const baseAppConfig = {
  appName: "金匠倌收银",
  subjectName: "金匠倌收银",
  brandSlogan: "金匠倌收银系统",
  mode: "production",
  apiBaseUrl: PRODUCTION_API_BASE_URL,
  tenantId: "gold-recycle-v1",
  requestTimeout: 15000,
  servicePhone: "13456844944",
  serviceEmail: "3694597562@qq.com",
  contactName: "周银军",
  contactAddress: "佛山南海区大沥黄岐岐西路教师楼A601",
  serviceScope: "收银系统",
  defaultStoreId: "",
  defaultStoreCode: "",
  storeName: "",
  storeCity: "",
  recyclePhotoRules: {
    minCount: 3,
    maxCount: 3,
    requireExactThree: false
  },
  featureFlags: {
    recycleInlineCustomerEnabled: true
  },
  referencePrices: [
    { purity: "足金9999", price: 748, trend: "+6", label: "大盘回收参考" },
    { purity: "足金999", price: 742, trend: "+5", label: "门店常用价" },
    { purity: "22K", price: 680, trend: "+4", label: "高成色 K 金" },
    { purity: "18K", price: 558, trend: "+3", label: "常规 K 金" },
    { purity: "14K", price: 436, trend: "+2", label: "低成色 K 金" }
  ],
  endpoints: {
    login: "/api/v1/auth/wechat-login",
    passwordLogin: "/api/v1/auth/login",
    logout: "/api/v1/auth/logout",
    me: "/api/v1/me",
    listStores: "/api/v1/stores",
    dashboardSummary: "/api/v1/dashboard/summary",
    inventorySummary: "/api/v1/inventory/summary",
    inventoryItems: "/api/v1/inventory/items",
    materialItems: "/api/v1/materials",
    materialOutbound: "/api/v1/materials/:id/outbound",
    settings: "/api/v1/settings",
    prepareRecyclePhotoUpload: "/api/v1/uploads/recycle-photos/prepare",
    completeRecyclePhotoUpload: "/api/v1/uploads/recycle-photos/complete",
    createCashierOrder: "/api/v1/cashier/orders",
    createRecycleOrder: "/api/v1/recycle/orders",
    confirmRecycleOrder: "/api/v1/recycle/orders/:id/confirm",
    listRecycleOrders: "/api/v1/recycle/orders",
    recycleOrderDetail: "/api/v1/recycle/orders/:id",
    listCashierOrders: "/api/v1/cashier/orders",
    quotePreview: "/api/v1/recycle/quote-preview",
    goldReferencePrices: "/api/v1/gold-prices/reference",
    cashierOrderDetail: "/api/v1/cashier/orders/:id",
    refundCashierOrder: "/api/v1/cashier/orders/:id/refund",
    listMembers: "/api/v1/members",
    createMember: "/api/v1/members",
    memberDetail: "/api/v1/members/:id",
    listProducts: "/api/v1/products",
    createProduct: "/api/v1/products",
    productDetail: "/api/v1/products/:id"
  }
};

function canUseStorage() {
  return typeof wx !== "undefined" && typeof wx.getStorageSync === "function";
}

function getRuntimeConfig() {
  if (!canUseStorage()) {
    return {
      mode: baseAppConfig.mode,
      apiBaseUrl: baseAppConfig.apiBaseUrl
    };
  }

  const stored = wx.getStorageSync(RUNTIME_CONFIG_KEY) || {};
  const mode = stored.mode || baseAppConfig.mode;
  return {
    mode,
    apiBaseUrl: normalizeApiBaseUrl(stored.apiBaseUrl, mode)
  };
}

function getResolvedAppConfig() {
  const runtimeConfig = getRuntimeConfig();
  return Object.assign({}, baseAppConfig, runtimeConfig);
}

function saveRuntimeConfig(partial) {
  const nextConfig = Object.assign({}, getRuntimeConfig(), partial || {});
  nextConfig.apiBaseUrl = normalizeApiBaseUrl(nextConfig.apiBaseUrl, nextConfig.mode);
  if (canUseStorage()) {
    wx.setStorageSync(RUNTIME_CONFIG_KEY, nextConfig);
  }
  return nextConfig;
}

function normalizeApiBaseUrl(value, mode) {
  const apiBaseUrl = String(value || baseAppConfig.apiBaseUrl).trim().replace(/\/+$/, "");
  if (mode === "offline") {
    return apiBaseUrl || baseAppConfig.apiBaseUrl;
  }

  if (!apiBaseUrl || PLACEHOLDER_API_BASE_URLS.indexOf(apiBaseUrl) !== -1) {
    return baseAppConfig.apiBaseUrl;
  }

  if (/^(https?:\/\/)?(localhost|127\.|0\.0\.0\.0|198\.18\.)/i.test(apiBaseUrl)) {
    return baseAppConfig.apiBaseUrl;
  }

  if (/^http:\/\//i.test(apiBaseUrl)) {
    return baseAppConfig.apiBaseUrl;
  }

  return apiBaseUrl;
}

function resetRuntimeConfig() {
  if (canUseStorage()) {
    wx.removeStorageSync(RUNTIME_CONFIG_KEY);
  }
  return getRuntimeConfig();
}

const appConfig = new Proxy(baseAppConfig, {
  get(target, key) {
    return getResolvedAppConfig()[key];
  }
});

module.exports = {
  appConfig,
  getRuntimeConfig,
  getResolvedAppConfig,
  saveRuntimeConfig,
  resetRuntimeConfig
};
