/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: systemStore.js
 * 功能描述: 工具函数
 * 作者: 廖心慈
 * 创建日期: 2026-06-10
 */

const { appConfig, request } = require("./apiClient");
const { getScopedStorageSync, setScopedStorageSync } = require("./sessionStorage");

const SETTINGS_KEY = "gr_system_settings";
const SETTINGS_META_KEY = "gr_system_settings_meta";

function deepCopy(value) {
  return JSON.parse(JSON.stringify(value));
}

function canUseStorage() {
  return typeof wx !== "undefined" && typeof wx.getStorageSync === "function";
}

function formatTimestamp(date) {
  const current = date || new Date();
  const pad = function(value) {
    return value < 10 ? `0${value}` : String(value);
  };
  return [
    current.getFullYear(),
    pad(current.getMonth() + 1),
    pad(current.getDate())
  ].join("-") + " " + [
    pad(current.getHours()),
    pad(current.getMinutes()),
    pad(current.getSeconds())
  ].join(":");
}

function getDefaultSystemSettings() {
  return {
    orgId: "",
    brand: {
      name: appConfig.appName,
      subjectName: appConfig.subjectName || appConfig.appName,
      contactName: appConfig.contactName,
      servicePhone: appConfig.servicePhone,
      serviceEmail: appConfig.serviceEmail,
      contactAddress: appConfig.contactAddress,
      receiptTitle: `${appConfig.appName}门店小票`,
      supportMiniApp: true,
      defaultCurrency: "CNY"
    },
    recycle: {
      minPhotoCount: appConfig.recyclePhotoRules.minCount,
      maxPhotoCount: appConfig.recyclePhotoRules.maxCount,
      requireExactThree: !!appConfig.recyclePhotoRules.requireExactThree,
      requireIdCheck: true
    },
    storage: {
      enabled: false,
      provider: "",
      bucket: "",
      region: "",
      publicBaseUrl: "",
      pathPrefix: "",
      uploadStrategy: "",
      callbackEnabled: false,
      statusDescription: ""
    },
    featureFlags: {},
    updatedBy: "",
    updatedAt: ""
  };
}

function getFallbackSystemProfile() {
  return {
    appName: appConfig.appName,
    subjectName: appConfig.subjectName || appConfig.appName,
    serviceScope: appConfig.serviceScope,
    contactName: appConfig.contactName,
    servicePhone: appConfig.servicePhone,
    serviceEmail: appConfig.serviceEmail,
    contactAddress: appConfig.contactAddress,
    receiptTitle: `${appConfig.appName}门店小票`,
    supportMiniApp: true
  };
}

function getStoredSettings() {
  if (!canUseStorage()) {
    return getDefaultSystemSettings();
  }
  const stored = getScopedStorageSync(SETTINGS_KEY, {});
  return Object.assign(getDefaultSystemSettings(), stored || {});
}

function saveStoredSettings(settings) {
  const nextSettings = Object.assign(getDefaultSystemSettings(), settings || {});
  if (canUseStorage()) {
    setScopedStorageSync(SETTINGS_KEY, nextSettings);
  }
  return nextSettings;
}

function getStoredSettingsMeta() {
  if (!canUseStorage()) {
    return {};
  }
  const stored = getScopedStorageSync(SETTINGS_META_KEY, {});
  return stored && typeof stored === "object" ? stored : {};
}

function saveStoredSettingsMeta(meta) {
  if (canUseStorage()) {
    setScopedStorageSync(SETTINGS_META_KEY, meta || {});
  }
}

function buildSystemProfile(settings) {
  const source = Object.assign(getDefaultSystemSettings(), settings || {});
  const brand = Object.assign({}, getDefaultSystemSettings().brand, source.brand || {});
  return {
    appName: brand.name || appConfig.appName,
    subjectName: brand.subjectName || appConfig.subjectName || brand.name || appConfig.appName,
    serviceScope: appConfig.serviceScope,
    contactName: brand.contactName || appConfig.contactName,
    servicePhone: brand.servicePhone || appConfig.servicePhone,
    serviceEmail: brand.serviceEmail || appConfig.serviceEmail,
    contactAddress: brand.contactAddress || appConfig.contactAddress,
    receiptTitle: brand.receiptTitle || `${appConfig.appName}门店小票`,
    supportMiniApp: typeof brand.supportMiniApp === "boolean" ? brand.supportMiniApp : true
  };
}

function getSystemProfile() {
  return buildSystemProfile(getStoredSettings());
}

function loadSystemSettingsOnline() {
  if (appConfig.mode === "offline") {
    const fallback = getDefaultSystemSettings();
    saveStoredSettings(fallback);
    saveStoredSettingsMeta({
      sourceType: "local",
      updatedAt: formatTimestamp(),
      lastError: ""
    });
    return Promise.resolve({
      settings: deepCopy(fallback),
      profile: buildSystemProfile(fallback),
      meta: deepCopy(getStoredSettingsMeta())
    });
  }

  return request({
    endpoint: appConfig.endpoints.settings
  }).then(function(payload) {
    const data = payload && payload.data ? payload.data : payload;
    const settings = saveStoredSettings(data || {});
    const meta = {
      sourceType: "api",
      updatedAt: formatTimestamp(),
      lastError: ""
    };
    saveStoredSettingsMeta(meta);
    return {
      settings: deepCopy(settings),
      profile: buildSystemProfile(settings),
      meta
    };
  }).catch(function(error) {
    const meta = Object.assign({}, getStoredSettingsMeta(), {
      lastError: error && error.message ? error.message : "failed to load settings",
      lastErrorAt: formatTimestamp()
    });
    saveStoredSettingsMeta(meta);
    const settings = getStoredSettings();
    return {
      settings: deepCopy(settings),
      profile: buildSystemProfile(settings),
      meta
    };
  });
}

module.exports = {
  getDefaultSystemSettings,
  getFallbackSystemProfile,
  getSystemProfile,
  getStoredSettings,
  loadSystemSettingsOnline
};
