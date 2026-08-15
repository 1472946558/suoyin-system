// utils/customer-auth.js - 顾客登录管理
const { api } = require('./request.js');

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

/**
 * 静默登录（用 wx.login 拿 code，调后端 wechat-login）
 * 已有 token 则不重复登录
 */
function ensureCustomerSession() {
  return new Promise((resolve, reject) => {
    const existing = getCustomerToken();
    if (existing) {
      resolve({ token: existing, profile: getStoredProfile(), silent: true });
      return;
    }
    wx.login({
      success(loginRes) {
        if (!loginRes.code) {
          reject(new Error('wx.login 失败'));
          return;
        }
        api.wechatLogin({ code: loginRes.code })
          .then(data => {
            setCustomerToken(data.token);
            setStoredProfile(data.profile);
            resolve({ token: data.token, profile: data.profile, silent: false });
          })
          .catch(reject);
      },
      fail: reject
    });
  });
}

/**
 * 主动登录（带昵称头像）
 */
function customerLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success(loginRes) {
        if (!loginRes.code) {
          reject(new Error('wx.login 失败'));
          return;
        }
        api.wechatLogin({ code: loginRes.code })
          .then(data => {
            setCustomerToken(data.token);
            setStoredProfile(data.profile);
            resolve(data);
          })
          .catch(reject);
      },
      fail: reject
    });
  });
}

/**
 * 手机号授权（需 button open-type="getPhoneNumber" 触发）
 */
function customerPhoneAuth(phoneCode) {
  return api.phoneAuth({ phoneCode })
    .then(profile => {
      setStoredProfile(profile);
      return profile;
    });
}

/**
 * 登出
 */
function customerLogout() {
  const token = getCustomerToken();
  if (!token) {
    clearCustomerSession();
    return Promise.resolve();
  }
  return api.logout()
    .catch(() => {})
    .then(() => {
      clearCustomerSession();
    });
}

module.exports = {
  getCustomerToken,
  setCustomerToken,
  clearCustomerSession,
  getStoredProfile,
  setStoredProfile,
  ensureCustomerSession,
  customerLogin,
  customerPhoneAuth,
  customerLogout
};
