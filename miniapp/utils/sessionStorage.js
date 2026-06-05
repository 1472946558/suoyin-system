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

module.exports = {
  buildScopedKey,
  getScopedStorageSync,
  setScopedStorageSync,
  removeScopedStorageSync,
  clearScopedBusinessData,
  clearLegacyBusinessData,
  getSessionScope
};
