// utils/customer-auth.js - 顾客登录业务
// 拆分说明：把纯 wx.getStorageSync 抽到 customer-storage.js，原因是
//   request.js 和 customer-auth.js 之前互相 require，形成循环依赖，
//   微信开发者工具在分包加载时对该循环解析不佳，会触发
//   "can not find module : , require args is ../../utils/request.js" 报错。
const { api } = require('./request.js');
const {
  getCustomerToken,
  setCustomerToken,
  clearCustomerSession,
  getStoredProfile,
  setStoredProfile
} = require('./customer-storage.js');

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

// 重新导出，保持向后兼容（其它页面可能还在 require customer-auth.js 拿这些方法）
module.exports = {
  // 存储相关（透传给 customer-storage.js）
  getCustomerToken,
  setCustomerToken,
  clearCustomerSession,
  getStoredProfile,
  setStoredProfile,
  // 业务方法（保持原接口）
  ensureCustomerSession,
  customerLogin,
  customerPhoneAuth,
  customerLogout
};
