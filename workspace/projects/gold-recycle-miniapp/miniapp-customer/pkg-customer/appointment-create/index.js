// pkg-customer/appointment-create/index.js - 到店预约创建页
const { api } = require('../../utils/request.js');
const { ensureCustomerSession, customerPhoneAuth, getStoredProfile, getCustomerToken } = require('../../utils/customer-auth.js');
const { SERVICE_TYPES, serviceTypeText } = require('../../utils/customer-services.js');
const { next7Days, fmtDate, isPastTime, timePlusMin, parseTimeMin } = require('../../utils/customer-date.js');

Page({
  data: {
    // 门店
    storeId: '',
    store: null,
    storeLoading: true,

    // 服务类型
    serviceTypes: SERVICE_TYPES,
    selectedService: '',

    // 日期
    days: [],
    selectedDate: '',

    // 时段
    slots: [],
    selectedSlot: '',

    // 备注
    remark: '',

    // 顾客
    isLoggedIn: false,
    phoneVerified: false,
    profile: null,

    // 提交
    submitting: false
  },

  onLoad(query) {
    this.setData({
      storeId: query.storeId || '',
      days: next7Days(),
      selectedDate: fmtDate(new Date())
    });

    // 检查登录状态
    const token = getCustomerToken();
    const profile = getStoredProfile();
    if (token && profile) {
      this.setData({
        isLoggedIn: true,
        phoneVerified: !!profile.phone,
        profile: profile
      });
    }

    if (this.data.storeId) {
      this.loadStore();
    }
  },

  onShow() {
    // 返回时刷新登录状态
    const token = getCustomerToken();
    const profile = getStoredProfile();
    if (token && profile) {
      this.setData({
        isLoggedIn: true,
        phoneVerified: !!profile.phone,
        profile: profile
      });
    }
  },

  // 加载门店信息
  loadStore() {
    this.setData({ storeLoading: true });
    api.getStoreDetail(this.data.storeId)
      .then(store => {
        this.setData({ store, storeLoading: false });
        this.loadSlots();
      })
      .catch(err => {
        this.setData({ storeLoading: false });
        wx.showToast({ title: err.message || '门店加载失败', icon: 'none' });
      });
  },

  // 加载时段
  loadSlots() {
    if (!this.data.storeId || !this.data.selectedDate) return;
    api.getStoreSlots(this.data.storeId, this.data.selectedDate)
      .then(slots => {
        // 标记已过时段
        const now = new Date();
        const todayStr = fmtDate(now);
        const processed = (slots || []).map(s => {
          const isPast = (this.data.selectedDate === todayStr) && isPastTime(this.data.selectedDate, s.startTime);
          return {
            ...s,
            label: s.startTime + ' - ' + (s.endTime || timePlusMin(s.startTime, 30)),
            isPast: isPast,
            isFull: s.available !== undefined && s.available <= 0,
            disabled: isPast || (s.available !== undefined && s.available <= 0)
          };
        });
        this.setData({ slots: processed, selectedSlot: '' });
      })
      .catch(err => {
        console.warn('加载时段失败', err);
        this.setData({ slots: [], selectedSlot: '' });
      });
  },

  // 选服务类型
  onSelectService(e) {
    this.setData({ selectedService: e.currentTarget.dataset.code });
  },

  // 选日期
  onSelectDate(e) {
    const date = e.currentTarget.dataset.date;
    this.setData({ selectedDate: date, selectedSlot: '' });
    this.loadSlots();
  },

  // 选时段
  onSelectSlot(e) {
    const slot = e.currentTarget.dataset.slot;
    if (this.data.slots.find(s => s.startTime === slot && s.disabled)) return;
    this.setData({ selectedSlot: slot });
  },

  // 输入备注
  onRemarkInput(e) {
    this.setData({ remark: e.detail.value });
  },

  // 手机号授权
  onGetPhone(e) {
    if (e.detail.errMsg !== 'getPhoneNumber:ok') {
      wx.showToast({ title: '需要手机号授权才能预约', icon: 'none' });
      return;
    }
    const phoneCode = e.detail.code;
    if (!phoneCode) {
      wx.showToast({ title: '授权失败，请重试', icon: 'none' });
      return;
    }
    wx.showLoading({ title: '验证中...' });
    // 确保已登录
    const doAuth = () => {
      customerPhoneAuth(phoneCode)
        .then(profile => {
          wx.hideLoading();
          this.setData({
            isLoggedIn: true,
            phoneVerified: true,
            profile: profile
          });
          wx.showToast({ title: '授权成功', icon: 'success' });
        })
        .catch(err => {
          wx.hideLoading();
          wx.showToast({ title: err.message || '手机号验证失败', icon: 'none' });
        });
    };

    if (!getCustomerToken()) {
      ensureCustomerSession()
        .then(doAuth)
        .catch(() => {
          wx.hideLoading();
          wx.showToast({ title: '登录失败，请重试', icon: 'none' });
        });
    } else {
      doAuth();
    }
  },

  // 提交预约
  onSubmit() {
    const { storeId, selectedService, selectedDate, selectedSlot, remark, phoneVerified, submitting } = this.data;

    if (submitting) return;

    if (!storeId) {
      wx.showToast({ title: '请选择门店', icon: 'none' });
      return;
    }
    if (!selectedService) {
      wx.showToast({ title: '请选择服务类型', icon: 'none' });
      return;
    }
    if (!selectedDate) {
      wx.showToast({ title: '请选择日期', icon: 'none' });
      return;
    }
    if (!selectedSlot) {
      wx.showToast({ title: '请选择时段', icon: 'none' });
      return;
    }
    if (!phoneVerified) {
      wx.showToast({ title: '请先授权手机号', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });
    const data = {
      storeId: parseInt(storeId, 10),
      serviceType: selectedService,
      appointmentDate: selectedDate,
      startTime: selectedSlot,
      remark: (remark || '').trim()
    };

    api.createAppointment(data)
      .then(result => {
        wx.showToast({ title: '预约提交成功', icon: 'success' });
        setTimeout(() => {
          // 跳到预约详情
          wx.redirectTo({
            url: '/pkg-customer/appointment-detail/index?id=' + result.id
          });
        }, 1000);
      })
      .catch(err => {
        this.setData({ submitting: false });
        const msg = err.message || '预约失败';
        if (msg.includes('duplicate') || msg.includes('重复')) {
          wx.showModal({ title: '提示', content: '该时段已有有效预约，请选择其他时段', showCancel: false });
        } else if (msg.includes('capacity') || msg.includes('已满')) {
          wx.showModal({ title: '提示', content: '该时段已约满，请选择其他时段', showCancel: false });
        } else {
          wx.showToast({ title: msg, icon: 'none' });
        }
      });
  },

  // 选择门店（跳到门店 Tab）
  onSelectStore() {
    wx.navigateTo({
      url: '/pkg-customer/store-detail/index?id=' + this.data.storeId,
      fail: () => wx.switchTab({ url: '/pages/stores/index' })
    });
  }
});
