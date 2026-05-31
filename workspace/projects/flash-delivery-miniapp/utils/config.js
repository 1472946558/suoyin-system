const appConfig = {
  appName: "疾蜂急送",
  brandSlogan: "同城专人急送",
  mode: "mock", // mock | production
  apiBaseUrl: "https://api.your-domain.com",
  tenantId: "demo-tenant",
  requestTimeout: 15000,
  servicePhone: "400-800-2026",
  mapProvider: "tencent",
  privacyUrl: "https://your-domain.com/privacy",
  userAgreementUrl: "https://your-domain.com/agreement",
  endpoints: {
    createOrder: "/api/v1/orders",
    listOrders: "/api/v1/orders",
    orderDetail: "/api/v1/orders/:id",
    orderTrack: "/api/v1/orders/:id/track",
    estimatePrice: "/api/v1/pricing/estimate",
    uploadFile: "/api/v1/files",
    login: "/api/v1/auth/wechat-login"
  }
};

module.exports = {
  appConfig
};
