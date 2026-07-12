/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: apiClient.js
 * 功能描述: 接口模块
 * 作者: 廖心慈
 * 创建日期: 2026-05-09
 */

const { appConfig, getResolvedAppConfig, getRuntimeConfig, saveRuntimeConfig, resetRuntimeConfig } = require("./config");
const { clearAllSessionData } = require("./sessionStorage");

let authRedirecting = false;

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
    if (error.statusCode === 401) {
      return "登录已失效，请重新登录";
    }
    return message ? `HTTP ${error.statusCode}: ${message}` : `HTTP ${error.statusCode}`;
  }

  if (error.type === "NETWORK_ERROR") {
    const message = readMessage(error.error);
    return message || "网络请求失败";
  }

  return readMessage(error) || "请求失败";
}

function redirectToLoginAfterUnauthorized() {
  if (authRedirecting || typeof wx === "undefined") {
    return;
  }
  authRedirecting = true;
  wx.showToast({ title: "登录已失效，请重新登录", icon: "none" });
  setTimeout(function() {
    wx.reLaunch({
      url: "/pages/account/index",
      complete() {
        authRedirecting = false;
      }
    });
  }, 250);
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
        if (response.statusCode === 401) {
          clearAllSessionData();
          redirectToLoginAfterUnauthorized();
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
