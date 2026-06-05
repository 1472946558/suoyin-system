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

function readMessage(value) {
  if (!value) {
    return "";
  }

  if (typeof value === "string") {
    return value.trim();
  }

  if (Array.isArray(value)) {
    return value.map(readMessage).filter(Boolean).join(" / ");
  }

  if (typeof value === "object") {
    return readMessage(value.message || value.msg || value.error || value.errMsg || value.detail);
  }

  return "";
}

function formatRequestError(error) {
  if (!error) {
    return "请求失败";
  }

  if (error.type === "HTTP_ERROR") {
    const message = readMessage(error.response);
    return message ? `HTTP ${error.statusCode}: ${message}` : `HTTP ${error.statusCode}`;
  }

  if (error.type === "NETWORK_ERROR") {
    const message = readMessage(error.error);
    return message || "网络请求失败";
  }

  return readMessage(error) || "请求失败";
}

function request(options) {
  const { endpoint, method = "GET", data = {}, header = {}, params = {} } = options;
  const resolvedConfig = getResolvedAppConfig();

  if (resolvedConfig.mode === "offline") {
    return Promise.resolve({
      code: 0,
      message: "local",
      data: null,
      local: true
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
  formatEndpoint,
  formatRequestError
};
