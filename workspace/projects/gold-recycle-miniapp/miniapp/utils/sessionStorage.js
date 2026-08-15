/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: sessionStorage.js
 * 功能描述: 工具函数
 * 作者: 廖心慈
 * 创建日期: 2026-06-02
 */

const PROFILE_KEY = "gr_operator_profile";

const BUSINESS_KEYS = [
  "gr_recycle_orders",
  "gr_cashier_orders",
  "gr_recycle_draft",
  "gr_member_records",
  "gr_member_records_meta",
  "gr_product_records",
  "gr_product_records_meta"
];

function canUseStorage() {
  return typeof wx !== "undefined" && typeof wx.getStorageSync === "function";
}

function readStoredProfile() {
  if (!canUseStorage()) {
    return {};
  }
  const profile = wx.getStorageSync(PROFILE_KEY);
  return profile && typeof profile === "object" ? profile : {};
}

function normalizeScopeValue(value, fallback) {
  const trimmed = String(value || "").trim();
  return trimmed || fallback;
}

function getSessionScope(profile) {
  const source = profile || readStoredProfile();
  return {
    storeId: normalizeScopeValue(source.storeId || source.storeCode, "guest-store"),
    userId: normalizeScopeValue(source.id || source.phone || source.roleKey, "guest-user")
  };
}

function buildScopedKey(baseKey, profile) {
  const scope = getSessionScope(profile);
  return `${baseKey}::${scope.storeId}::${scope.userId}`;
}

function hasStoredValue(value) {
  return !(typeof value === "undefined" || value === "");
}

function getScopedStorageSync(baseKey, fallbackValue, profile) {
  if (!canUseStorage()) {
    return fallbackValue;
  }

  const scopedValue = wx.getStorageSync(buildScopedKey(baseKey, profile));
  if (hasStoredValue(scopedValue)) {
    return scopedValue;
  }

  const legacyValue = wx.getStorageSync(baseKey);
  return hasStoredValue(legacyValue) ? legacyValue : fallbackValue;
}

function setScopedStorageSync(baseKey, value, profile) {
  if (canUseStorage()) {
    wx.setStorageSync(buildScopedKey(baseKey, profile), value);
  }
}

function removeScopedStorageSync(baseKey, profile) {
  if (canUseStorage()) {
    wx.removeStorageSync(buildScopedKey(baseKey, profile));
  }
}

function clearScopedBusinessData(profile) {
  BUSINESS_KEYS.forEach(function(key) {
    removeScopedStorageSync(key, profile);
  });
}

function clearLegacyBusinessData() {
  if (!canUseStorage()) {
    return;
  }
  BUSINESS_KEYS.forEach(function(key) {
    wx.removeStorageSync(key);
  });
}

function clearAllSessionData() {
  if (!canUseStorage()) {
    return;
  }
  const profile = readStoredProfile();
  clearScopedBusinessData(profile);
  clearScopedBusinessData({});
  clearLegacyBusinessData();
  wx.removeStorageSync(PROFILE_KEY);
}

module.exports = {
  buildScopedKey,
  getScopedStorageSync,
  setScopedStorageSync,
  removeScopedStorageSync,
  clearScopedBusinessData,
  clearLegacyBusinessData,
  clearAllSessionData,
  getSessionScope
};
