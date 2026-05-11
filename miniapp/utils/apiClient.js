const { appConfig, getResolvedAppConfig, getRuntimeConfig, saveRuntimeConfig, resetRuntimeConfig } = require("./config");

function formatEndpoint(endpoint, params) {
  return Object.keys(params || {}).reduce((path, key) => {
    return path.replace(`:${key}`, encodeURIComponent(params[key]));
  }, endpoint);
}

function buildUrl(endpoint, params) {
  const resolvedConfig = getResolvedAppConfig();
  const base = resolvedConfig.apiBaseUrl.replace(/\/$/, "");
  const formatted = formatEndpoint(endpoint, params);
  const path = formatted.startsWith("/") ? formatted : `/${formatted}`;
  return `${base}${path}`;
}

function request(options) {
  const { endpoint, method = "GET", data = {}, header = {}, params = {} } = options;
  const resolvedConfig = getResolvedAppConfig();

  if (resolvedConfig.mode === "mock") {
    return Promise.resolve({
      code: 0,
      message: "mock",
      data: null,
      mock: true
    });
  }

  return new Promise((resolve, reject) => {
    const storedProfile = wx.getStorageSync("gr_operator_profile") || {};
    const tokenHeader = storedProfile.token ? { Authorization: `Bearer ${storedProfile.token}` } : {};

    wx.request({
      url: buildUrl(endpoint, params),
      method,
      data,
      timeout: resolvedConfig.requestTimeout,
      header: Object.assign({
        "content-type": "application/json",
        "x-tenant-id": resolvedConfig.tenantId
      }, tokenHeader, header),
      success(response) {
        if (response.statusCode >= 200 && response.statusCode < 300) {
          resolve(response.data);
          return;
        }
        reject({
          type: "HTTP_ERROR",
          statusCode: response.statusCode,
          response: response.data
        });
      },
      fail(error) {
        reject({
          type: "NETWORK_ERROR",
          error
        });
      }
    });
  });
}

module.exports = {
  appConfig,
  getRuntimeConfig,
  saveRuntimeConfig,
  resetRuntimeConfig,
  getResolvedAppConfig,
  request,
  buildUrl,
  formatEndpoint
};
