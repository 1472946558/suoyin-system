// pkg-customer/appointment-create/index.js - 到店预约创建页
const { api } = require('../../utils/request.js');
const { ensureCustomerSession, customerPhoneAuth, getStoredProfile, getCustomerToken } = require('../../utils/customer-auth.js');
const { SERVICE_TYPES, serviceTypeText } = require('../../utils/customer-services.js');
const { nextNDays, fmtDate, isPastTime, timePlusMin } = require('../../utils/customer-date.js');

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
    slotMinutes: 30,

    // 联系人
    contactName: '',
    contactPhone: '',

    // 备注
    remark: '',

    // 协议
    agreed: false,

    // 顾客
    isLoggedIn: false,
    phoneVerified: false,
    profile: null,

    // 提交
    submitting: false,

    // 预约须知文案（从后端配置取）
    appointmentNotes: ''
  },

  onLoad(query) {
    // 从首页配置读取预约规则
    const app = getApp();
    const homeConfig = app.globalData.homeConfig || {};
    const rules = homeConfig.appointmentRules || {};
    const bookableDays = rules.bookableDays || 7;
    const slotMinutes = rules.slotMinutes || 30;
    const bookableServiceTypes = rules.bookableServiceTypes || [];

    // 过滤服务类型（后端配置了可预约类型时仅展示配置项）
    let serviceTypes = SERVICE_TYPES;
    if (bookableServiceTypes.length > 0) {
      serviceTypes = SERVICE_TYPES.filter(s => bookableServiceTypes.indexOf(s.code) >= 0);
    }

    // 预约须知
    const appointmentNotes = homeConfig.appointmentNotes || '';

    this.setData({
      storeId: query.storeId || '',
      days: nextNDays(bookableDays),
      selectedDate: fmtDate(new Date()),
      serviceTypes,
      slotMinutes,
      appointmentNotes
    });

    // 检查登录状态
    const token = getCustomerToken();
    const profile = getStoredProfile();
    if (token && profile) {
      this.applyProfile(profile);
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
      this.applyProfile(profile);
    }
  },

  applyProfile(profile) {
    this.setData({
      isLoggedIn: true,
      phoneVerified: profile.phoneVerified === true || !!profile.phone,
      profile: profile,
      contactName: profile.nickname || this.data.contactName,
      contactPhone: profile.phone || this.data.contactPhone
    });
  },

  // 加载门店信息
  loadStore() {
    this.setData({ storeLoading: true });
    api.getStoreDetail(this.data.storeId)
      .then(store => {
        // 检查门店预约开关
        if (store.appointmentEnabled === false) {
          wx.showModal({
            title: '提示',
            content: '该门店暂未开放预约',
            showCancel: false,
            confirmText: '返回',
            success: () => wx.navigateBack()
          });
          return;
        }
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
      .then(resp => {
        // 后端返回 {date, slots: [...]}
        const raw = (resp && Array.isArray(resp)) ? resp : (resp && resp.slots) || [];
        const todayStr = fmtDate(new Date());
        const processed = raw.map(s => {
          const startTime = s.time;
          const endTime = timePlusMin(startTime, this.data.slotMinutes);
          const isPast = (this.data.selectedDate === todayStr) && isPastTime(this.data.selectedDate, startTime);
          const isFull = s.available === false || s.state === 'full';
          const isClosed = s.state === 'closed';
          return {
            startTime: startTime,
            endTime: endTime,
            label: startTime + '-' + endTime,
            isPast,
            isFull,
            isClosed,
            disabled: isPast || isFull || isClosed
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
    const target = this.data.slots.find(s => s.startTime === slot);
    if (target && target.disabled) return;
    this.setData({ selectedSlot: slot });
  },

  // 输入联系人
  onNameInput(e) {
    this.setData({ contactName: e.detail.value });
  },

  // 输入手机号
  onPhoneInput(e) {
    this.setData({ contactPhone: e.detail.value });
    // 用户手动修改手机号时撤销验证状态
    if (this.data.phoneVerified) {
      this.setData({ phoneVerified: false });
    }
  },

  // 点击获取手机号按钮（弹起授权）
  onGetPhoneTap() {
    // 触发子组件 button open-type=getPhoneNumber，由 onGetPhone 处理
    // 这里仅提示，实际授权在 WXML 的 button 中
    wx.showToast({ title: '请点击页面内"获取手机号"按钮', icon: 'none' });
  },

  // 切换协议勾选
  onToggleAgree() {
    this.setData({ agreed: !this.data.agreed });
  },

  // 查看协议
  onTapAgreement() {
    const notes = this.data.appointmentNotes || '1. 预约成功后，门店将通过电话与您确认。\n2. 请按时到店，如需取消请提前操作。\n3. 同一时段同一门店仅可预约一次。';
    wx.showModal({
      title: '预约须知',
      content: notes,
      showCancel: false,
      confirmText: '我知道了'
    });
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
          this.applyProfile(profile);
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
    const {
      storeId, selectedService, selectedDate, selectedSlot,
      contactName, remark, phoneVerified, profile, submitting, agreed
    } = this.data;

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
    if (!contactName.trim()) {
      wx.showToast({ title: '请输入联系人姓名', icon: 'none' });
      return;
    }
    if (!phoneVerified || !profile || (profile.phoneVerified !== true && !profile.phone)) {
      wx.showToast({ title: '请先完成微信手机号授权', icon: 'none' });
      return;
    }
    if (!agreed) {
      wx.showToast({ title: '请先阅读并同意《预约须知》', icon: 'none' });
      return;
    }

    this.setData({ submitting: true });
    const data = {
      storeId: storeId,
      serviceType: selectedService,
      appointmentDate: selectedDate,
      appointmentTime: selectedSlot,
      contactName: contactName.trim(),
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
        if (msg.indexOf('duplicate') >= 0 || msg.indexOf('重复') >= 0) {
          wx.showModal({ title: '提示', content: '该时段已有有效预约，请选择其他时段', showCancel: false });
        } else if (msg.indexOf('capacity') >= 0 || msg.indexOf('已满') >= 0) {
          wx.showModal({ title: '提示', content: '该时段已约满，请选择其他时段', showCancel: false });
        } else {
          wx.showToast({ title: msg, icon: 'none' });
        }
      });
  },

  // 选择门店（跳到门店 Tab）
  onSelectStore() {
    if (this.data.storeId) {
      wx.navigateTo({
        url: '/pkg-customer/store-detail/index?id=' + this.data.storeId
      });
    } else {
      wx.switchTab({ url: '/pages/stores/index' });
    }
  }
});
