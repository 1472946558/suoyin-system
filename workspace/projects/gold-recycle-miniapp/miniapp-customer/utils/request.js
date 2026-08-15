// utils/request.js - 网络请求封装
const { getCustomerToken } = require('./customer-auth.js');

function getApiBase() {
  const app = getApp();
  return app && app.globalData ? app.globalData.apiBase : 'https://jinjiangguan.com';
}

/**
 * 通用请求
 * @param {Object} options
 * @param {string} options.path  不含 /api 前缀
 * @param {string} [options.method='GET']
 * @param {Object} [options.data]
 * @param {boolean} [options.withAuth=true] 是否需要顾客 token
 * @returns {Promise<Object>} 业务 data
 */
function request({ path, method = 'GET', data, withAuth = true, header }) {
  const url = getApiBase() + path;
  const finalHeader = { 'Content-Type': 'application/json', ...(header || {}) };
  if (withAuth) {
    const token = getCustomerToken();
    if (token) {
      finalHeader['Authorization'] = 'Bearer ' + token;
    }
  }

  return new Promise((resolve, reject) => {
    wx.request({
      url,
      method,
      data,
      header: finalHeader,
      success(res) {
        if (res.statusCode < 200 || res.statusCode >= 300) {
          reject({
            status: res.statusCode,
            message: '网络异常(' + res.statusCode + ')'
          });
          return;
        }
        const body = res.data || {};
        // 后端统一格式：{ code, message, data, requestId }
        if (body.code !== undefined && body.code !== 0) {
          reject({
            code: body.code,
            message: body.message || '业务错误',
            status: res.statusCode
          });
          return;
        }
        resolve(body.data);
      },
      fail(err) {
        reject({ message: err.errMsg || '网络失败' });
      }
    });
  });
}

const api = {
  // 公开
  getHome: () => request({ path: '/api/v1/customer/home', withAuth: false }),
  getProducts: (params) => request({ path: '/api/v1/customer/products?' + qs(params), withAuth: false }),
  getProductDetail: (id) => request({ path: '/api/v1/customer/products/' + id, withAuth: false }),
  getStores: () => request({ path: '/api/v1/customer/stores', withAuth: false }),
  getStoreDetail: (id) => request({ path: '/api/v1/customer/stores/' + id, withAuth: false }),
  getStoreSlots: (id, date) => request({ path: '/api/v1/customer/stores/' + id + '/slots?date=' + date, withAuth: false }),
  getRecycleInfo: () => request({ path: '/api/v1/customer/recycle-info', withAuth: false }),

  // 认证
  wechatLogin: (data) => request({ path: '/api/v1/customer/auth/wechat-login', method: 'POST', data, withAuth: false }),
  phoneAuth: (data) => request({ path: '/api/v1/customer/auth/phone', method: 'POST', data }),
  logout: () => request({ path: '/api/v1/customer/auth/logout', method: 'POST' }),

  // 顾客
  getMe: () => request({ path: '/api/v1/customer/me' }),

  // 预约
  getAppointments: () => request({ path: '/api/v1/customer/appointments' }),
  createAppointment: (data) => request({ path: '/api/v1/customer/appointments', method: 'POST', data }),
  getAppointment: (id) => request({ path: '/api/v1/customer/appointments/' + id }),
  cancelAppointment: (id, reason) => request({ path: '/api/v1/customer/appointments/' + id + '/cancel', method: 'POST', data: { reason } }),
  updateAppointmentNotes: (id, remark) => request({ path: '/api/v1/customer/appointments/' + id + '/notes', method: 'PUT', data: { remark } })
};

function qs(obj) {
  if (!obj) return '';
  return Object.keys(obj)
    .filter(k => obj[k] !== undefined && obj[k] !== null && obj[k] !== '')
    .map(k => encodeURIComponent(k) + '=' + encodeURIComponent(obj[k]))
    .join('&');
}

module.exports = { request, api, qs };
