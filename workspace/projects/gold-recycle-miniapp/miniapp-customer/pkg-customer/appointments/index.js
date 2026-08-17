// pkg-customer/appointments/index.js - 我的预约列表
const { api } = require('../../utils/request.js');
const { appointmentStatusText, appointmentStatusColor, appointmentStatusBgColor, serviceTypeText } = require('../../utils/customer-services.js');
const { getCustomerToken } = require('../../utils/customer-auth.js');

const STATUS_TABS = [
  { key: '', label: '全部' },
  { key: 'PENDING', label: '待确认' },
  { key: 'CONFIRMED', label: '已确认' },
  { key: 'COMPLETED', label: '已完成' },
  { key: 'CANCELLED', label: '已取消' }
];

Page({
  data: {
    tabs: STATUS_TABS,
    activeTab: '',
    list: [],
    filteredList: [],
    loading: true,
    isLoggedIn: false
  },

  onLoad() {
    const token = getCustomerToken();
    if (!token) {
      this.setData({ loading: false, isLoggedIn: false });
      return;
    }
    this.setData({ isLoggedIn: true });
    this.loadList();
  },

  onShow() {
    // 从详情页返回时刷新
    if (this.data.isLoggedIn) {
      this.loadList();
    }
  },

  onPullDownRefresh() {
    this.loadList(() => wx.stopPullDownRefresh());
  },

  loadList(cb) {
    this.setData({ loading: true });
    api.getAppointments()
      .then(resp => {
        // 后端返回 { items: [...], total: N }，兼容直接返回数组
        var list = [];
        if (Array.isArray(resp)) {
          list = resp;
        } else if (resp && Array.isArray(resp.items)) {
          list = resp.items;
        }
        var processed = list.map(item => Object.assign({}, item, {
          // 后端已提供 statusText / serviceTypeText，缺失时本地兜底
          statusText: item.statusText || appointmentStatusText(item.status),
          statusColor: appointmentStatusColor(item.status),
          statusBgColor: appointmentStatusBgColor(item.status),
          serviceText: item.serviceTypeText || serviceTypeText(item.serviceType)
        }));
        // 按预约时间倒序排列（最新的在前）
        processed.sort(function(a, b) {
          var ka = (a.appointmentDate || '') + (a.appointmentTime || '');
          var kb = (b.appointmentDate || '') + (b.appointmentTime || '');
          return kb.localeCompare(ka);
        });
        this.setData({ list: processed, filteredList: this.filterByTab(processed, this.data.activeTab), loading: false });
        if (cb) cb();
      })
      .catch(err => {
        this.setData({ loading: false });
        if (cb) cb();
        if (err.status === 401 || (err.message && err.message.includes('unauthorized'))) {
          this.setData({ isLoggedIn: false });
        }
      });
  },

  onTabChange(e) {
    const key = e.currentTarget.dataset.key;
    this.setData({
      activeTab: key,
      filteredList: this.filterByTab(this.data.list, key)
    });
  },

  filterByTab(list, tab) {
    if (!tab) return list;
    return list.filter(item => item.status === tab);
  },

  onTapItem(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({ url: '/pkg-customer/appointment-detail/index?id=' + id });
  },

  onNewAppointment() {
    wx.navigateTo({ url: '/pkg-customer/appointment-create/index' });
  },

  onLogin() {
    wx.navigateTo({ url: '/pkg-customer/appointment-create/index' });
  }
});
