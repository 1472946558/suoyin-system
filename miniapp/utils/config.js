const RUNTIME_CONFIG_KEY = "gr_runtime_config";

const baseAppConfig = {
  appName: "金秤台",
  brandSlogan: "黄金回收收银系统",
  mode: "mock", // mock | production
  apiBaseUrl: "http://127.0.0.1:18082",
  tenantId: "gold-recycle-demo",
  requestTimeout: 15000,
  servicePhone: "400-860-2026",
  defaultStoreId: "store-shenzhen-nanshan",
  defaultStoreCode: "SZ-NS",
  storeName: "华强北体验店",
  storeCity: "深圳",
  recyclePhotoRules: {
    minCount: 2,
    maxCount: 3,
    requireExactThree: false
  },
  mockBenchPrices: [
    { purity: "足金9999", price: 748, trend: "+6", label: "大盘回收参考" },
    { purity: "足金999", price: 742, trend: "+5", label: "门店常用价" },
    { purity: "22K", price: 680, trend: "+4", label: "高成色 K 金" },
    { purity: "18K", price: 558, trend: "+3", label: "常规 K 金" }
  ],
  endpoints: {
    login: "/api/v1/auth/wechat-login",
    prepareRecyclePhotoUpload: "/api/v1/uploads/recycle-photos/prepare",
    completeRecyclePhotoUpload: "/api/v1/uploads/recycle-photos/complete",
    createRecycleOrder: "/api/v1/recycle/orders",
    listRecycleOrders: "/api/v1/recycle/orders",
    recycleOrderDetail: "/api/v1/recycle/orders/:id",
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
  return {
    mode: stored.mode || baseAppConfig.mode,
    apiBaseUrl: stored.apiBaseUrl || baseAppConfig.apiBaseUrl
  };
}

function getResolvedAppConfig() {
  const runtimeConfig = getRuntimeConfig();
  return Object.assign({}, baseAppConfig, runtimeConfig);
}

function saveRuntimeConfig(partial) {
  const nextConfig = Object.assign({}, getRuntimeConfig(), partial || {});
  if (canUseStorage()) {
    wx.setStorageSync(RUNTIME_CONFIG_KEY, nextConfig);
  }
  return nextConfig;
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
