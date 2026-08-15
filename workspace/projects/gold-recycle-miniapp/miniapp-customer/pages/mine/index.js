// pages/mine/index.js - 我的页面（Tab 4）
const { api } = require('../../utils/request.js');
const { getCustomerToken, getStoredProfile, clearCustomerSession, customerLogout, ensureCustomerSession } = require('../../utils/customer-auth.js');

Page({
  data: {
    isLoggedIn: false,
    profile: null,
    servicePhone: ''
  },

  onLoad() {
    const app = getApp();
    this.setData({ servicePhone: app.globalData.servicePhone });
  },

  onShow() {
    this.refreshProfile();
  },

  refreshProfile() {
    const token = getCustomerToken();
    const stored = getStoredProfile();
    if (token && stored) {
      this.setData({ isLoggedIn: true, profile: stored });
      // 静默刷新 profile
      api.getMe()
        .then(profile => {
          if (profile) {
            this.setData({ profile, isLoggedIn: true });
            try { wx.setStorageSync('customer_profile', profile); } catch (e) {}
          }
        })
        .catch(() => {});
    } else {
      this.setData({ isLoggedIn: false, profile: null });
    }
  },

  // 去预约/登录
  onLogin() {
    wx.navigateTo({ url: '/pkg-customer/appointment-create/index' });
  },

  // 我的预约
  onMyAppointments() {
    if (!this.data.isLoggedIn) {
      wx.navigateTo({ url: '/pkg-customer/appointment-create/index' });
      return;
    }
    wx.navigateTo({ url: '/pkg-customer/appointments/index' });
  },

  // 回收介绍
  onRecycleInfo() {
    wx.navigateTo({ url: '/pkg-customer/recycle-info/index' });
  },

  // 店长入口
  onStaffEntrance() {
    // N08 员工端改造时对接
    // 当前提示店长入口功能
    wx.showModal({
      title: '店长入口',
      content: '门店管理功能正在升级中，请咨询门店管理人员或使用原管理端小程序。',
      showCancel: false,
      confirmText: '知道了'
    });
  },

  // 拨打客服电话
  onCallService() {
    const phone = this.data.servicePhone;
    if (!phone) {
      wx.showToast({ title: '暂无客服电话', icon: 'none' });
      return;
    }
    const num = String(phone).replace(/[*\s]/g, '');
    // 脱敏号无法拨打，提示完整号码
    if (phone.indexOf('*') !== -1) {
      wx.showToast({ title: '请通过门店页面拨打', icon: 'none' });
      return;
    }
    wx.makePhoneCall({ phoneNumber: num }).catch(() => {});
  },

  // 登出
  onLogout() {
    wx.showModal({
      title: '退出登录',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (!res.confirm) return;
        customerLogout()
          .then(() => {
            this.setData({ isLoggedIn: false, profile: null });
            wx.showToast({ title: '已退出', icon: 'success' });
          })
          .catch(() => {
            clearCustomerSession();
            this.setData({ isLoggedIn: false, profile: null });
          });
      }
    });
  },

  // 手机号授权
  onGetPhone(e) {
    if (e.detail.errMsg !== 'getPhoneNumber:ok') return;
    const phoneCode = e.detail.code;
    if (!phoneCode) return;

    if (!getCustomerToken()) {
      ensureCustomerSession()
        .then(() => this.doPhoneAuth(phoneCode))
        .catch(() => wx.showToast({ title: '登录失败', icon: 'none' }));
    } else {
      this.doPhoneAuth(phoneCode);
    }
  },

  doPhoneAuth(phoneCode) {
    wx.showLoading({ title: '验证中...' });
    const { customerPhoneAuth } = require('../../utils/customer-auth.js');
    customerPhoneAuth(phoneCode)
      .then(profile => {
        wx.hideLoading();
        this.setData({ isLoggedIn: true, profile });
        wx.showToast({ title: '授权成功', icon: 'success' });
      })
      .catch(err => {
        wx.hideLoading();
        wx.showToast({ title: err.message || '授权失败', icon: 'none' });
      });
  }
});
