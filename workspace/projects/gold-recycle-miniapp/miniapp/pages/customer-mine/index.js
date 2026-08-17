// pages/mine/index.js - 我的页面（Tab 4）
const { api } = require('../../utils/request.js');
const { getCustomerToken, getStoredProfile, clearCustomerSession, customerLogout, ensureCustomerSession } = require('../../utils/customer-auth.js');

Page({
  data: {
    isLoggedIn: false,
    profile: null,
    profileTail: '顾',
    servicePhone: '',
    serviceIntro: '',
    pendingCount: 0
  },

  onLoad() {
    const app = getApp();
    const homeConfig = app.globalData.homeConfig || {};
    this.setData({
      servicePhone: homeConfig.servicePhone || app.globalData.servicePhone || '',
      serviceIntro: homeConfig.serviceIntro || ''
    });
  },

  onShow() {
    this.refreshProfile();
  },

  refreshProfile() {
    const token = getCustomerToken();
    const stored = getStoredProfile();
    // 同步首页配置中的 servicePhone / serviceIntro
    const app = getApp();
    const homeConfig = app.globalData.homeConfig || {};
    if (homeConfig.servicePhone || app.globalData.servicePhone) {
      this.setData({
        servicePhone: homeConfig.servicePhone || app.globalData.servicePhone || '',
        serviceIntro: homeConfig.serviceIntro || this.data.serviceIntro
      });
    }
    if (token && stored) {
      this.setData({
        isLoggedIn: true,
        profile: stored,
        profileTail: stored.phone ? String(stored.phone).slice(-4) : '顾'
      });
      // 静默刷新 profile
      api.getMe()
        .then(profile => {
          if (profile) {
            this.setData({
              profile,
              isLoggedIn: true,
              profileTail: profile.phone ? String(profile.phone).slice(-4) : '顾'
            });
            try { wx.setStorageSync('customer_profile', profile); } catch (e) {}
          }
        })
        .catch(() => {});
      // 静默加载待处理预约数
      this.loadPendingCount();
    } else {
      this.setData({ isLoggedIn: false, profile: null, profileTail: '顾', pendingCount: 0 });
    }
  },

  loadPendingCount() {
    api.getAppointments()
      .then(resp => {
        var list = [];
        if (Array.isArray(resp)) {
          list = resp;
        } else if (resp && Array.isArray(resp.items)) {
          list = resp.items;
        }
        var count = list.filter(function(item) {
          return item.status === 'PENDING' || item.status === 'CONFIRMED';
        }).length;
        this.setData({ pendingCount: count });
      })
      .catch(() => {});
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

  // 隐私保护指引
  onPrivacy() {
    wx.navigateTo({ url: '/pages/customer-privacy/index' });
  },

  // 店长入口
  onStaffEntrance() {
    // 顾客端和员工端已合并到同一个 AppID，员工入口必须走同一小程序内部路由。
    wx.reLaunch({
      url: '/pages/account/index',
      fail: function() {
        wx.showToast({ title: '员工入口暂不可用', icon: 'none' });
      }
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
            this.setData({ isLoggedIn: false, profile: null, profileTail: '顾' });
            wx.showToast({ title: '已退出', icon: 'success' });
          })
          .catch(() => {
            clearCustomerSession();
            this.setData({ isLoggedIn: false, profile: null, profileTail: '顾' });
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
        this.setData({
          isLoggedIn: true,
          profile,
          profileTail: profile && profile.phone ? String(profile.phone).slice(-4) : '顾'
        });
        wx.showToast({ title: '授权成功', icon: 'success' });
      })
      .catch(err => {
        wx.hideLoading();
        wx.showToast({ title: err.message || '授权失败', icon: 'none' });
      });
  }
});
