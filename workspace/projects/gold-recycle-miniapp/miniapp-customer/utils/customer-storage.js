// utils/customer-storage.js - 顾客端本地存储（与登录逻辑解耦，避免循环依赖）
const TOKEN_KEY = 'customer_token';
const PROFILE_KEY = 'customer_profile';

function getCustomerToken() {
  try {
    return wx.getStorageSync(TOKEN_KEY) || '';
  } catch (e) {
    return '';
  }
}

function setCustomerToken(token) {
  try { wx.setStorageSync(TOKEN_KEY, token); } catch (e) {}
}

function clearCustomerSession() {
  try {
    wx.removeStorageSync(TOKEN_KEY);
    wx.removeStorageSync(PROFILE_KEY);
  } catch (e) {}
}

function getStoredProfile() {
  try { return wx.getStorageSync(PROFILE_KEY) || null; } catch (e) { return null; }
}

function setStoredProfile(profile) {
  try { wx.setStorageSync(PROFILE_KEY, profile); } catch (e) {}
}

module.exports = {
  getCustomerToken,
  setCustomerToken,
  clearCustomerSession,
  getStoredProfile,
  setStoredProfile
};
