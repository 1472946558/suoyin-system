/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: app.js
 * 功能描述: 应用主模块
 * 作者: 廖心慈
 * 创建日期: 2026-05-09
 */

const { seedOrders } = require("./utils/orderStore");
const { seedCatalog } = require("./utils/catalogStore");
const { appConfig } = require("./utils/config");
const { getProfile, clearProfile } = require("./utils/userStore");

function getSystemInfo() {
  if (wx.getDeviceInfo && wx.getWindowInfo && wx.getAppBaseInfo) {
    const deviceInfo = wx.getDeviceInfo();
    const windowInfo = wx.getWindowInfo();
    const appBaseInfo = wx.getAppBaseInfo();
    return Object.assign({}, deviceInfo, windowInfo, appBaseInfo);
  }
  return wx.getSystemInfoSync();
}

function isExpired(expiresAt) {
  const timestamp = Date.parse(expiresAt || "");
  return Number.isFinite(timestamp) && timestamp <= Date.now();
}

App({
  globalData: {
    brandName: appConfig.appName,
    brandSlogan: appConfig.brandSlogan,
    storeName: appConfig.storeName,
    servicePhone: appConfig.servicePhone,
    city: appConfig.storeCity,
    apiBaseUrl: appConfig.apiBaseUrl,
    mode: appConfig.mode,
    sysinfo: null
  },

  BLEInformation: {
    platform: "",
    deviceId: "",
    writeCharaterId: "",
    writeServiceId: "",
    notifyCharaterId: "",
    notifyServiceId: "",
    readCharaterId: "",
    readServiceId: ""
  },

  onLaunch() {
    this.globalData.sysinfo = getSystemInfo();
    this.BLEInformation.platform = this.getPlatform();
    const profile = getProfile();
    if (appConfig.mode !== "offline" && profile.loggedIn && !profile.token) {
      clearProfile();
    }
    if (appConfig.mode !== "offline" && profile.loggedIn && isExpired(profile.expiresAt)) {
      clearProfile();
    }
    if (appConfig.mode === "offline") {
      seedOrders();
      seedCatalog();
    }
  },

  getModel() {
    return this.globalData.sysinfo ? this.globalData.sysinfo.model : "";
  },

  getVersion() {
    return this.globalData.sysinfo ? this.globalData.sysinfo.version : "";
  },

  getSystem() {
    return this.globalData.sysinfo ? this.globalData.sysinfo.system : "";
  },

  getPlatform() {
    return this.globalData.sysinfo ? this.globalData.sysinfo.platform : "";
  },

  getSDKVersion() {
    return this.globalData.sysinfo ? this.globalData.sysinfo.SDKVersion : "";
  }
});
