const RUNTIME_CONFIG_KEY = "gr_runtime_config";

const baseAppConfig = {
  appName: "黄金收银",
  brandSlogan: "黄金门店回收收银系统",
  mode: "production",
  apiBaseUrl: "https://example.com",
  tenantId: "gold-recycle-v1",
  requestTimeout: 15000,
  servicePhone: "13800000000",
  serviceEmail: "support@example.com",
  contactName: "示例老板",
  contactAddress: "示例联系地址",
  serviceScope: "收银系统",
  defaultStoreId: "store-demo-main",
  defaultStoreCode: "DEMO-01",
  storeName: "示例门店1",
  storeCity: "佛山",
  recyclePhotoRules: {
    minCount: 3,
    maxCount: 3,
    requireExactThree: false
  },
  referencePrices: [
    { purity: "足金9999", price: 748, trend: "+6", label: "大盘回收参考" },
    { purity: "足金999", price: 742, trend: "+5", label: "门店常用价" },
    { purity: "22K", price: 680, trend: "+4", label: "高成色 K 金" },
    { purity: "18K", price: 558, trend: "+3", label: "常规 K 金" }
  ],
  endpoints: {
    login: "/api/v1/auth/wechat-login",
    passwordLogin: "/api/v1/auth/login",
    logout: "/api/v1/auth/logout",
    listStores: "/api/v1/stores",
    prepareRecyclePhotoUpload: "/api/v1/uploads/recycle-photos/prepare",
    completeRecyclePhotoUpload: "/api/v1/uploads/recycle-photos/complete",
    createCashierOrder: "/api/v1/cashier/orders",
    createRecycleOrder: "/api/v1/recycle/orders",
    confirmRecycleOrder: "/api/v1/recycle/orders/:id/confirm",
    listRecycleOrders: "/api/v1/recycle/orders",
    recycleOrderDetail: "/api/v1/recycle/orders/:id",
    listCashierOrders: "/api/v1/cashier/orders",
    quotePreview: "/api/v1/recycle/quote-preview",
    cashierOrderDetail: "/api/v1/cashier/orders/:id",
    listMembers: "/api/v1/members",
    memberDetail: "/api/v1/members/:id",
    listProducts: "/api/v1/products",
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

  if (!apiBaseUrl || /^(https?:\/\/)?(localhost|127\.|0\.0\.0\.0|198\.18\.)/i.test(apiBaseUrl)) {
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
