// pkg-customer/store-detail/index.js - 门店详情页
const { api } = require('../../utils/request.js');
const { distanceKm, fmtDistance } = require('../../utils/customer-distance.js');

function todayStr() {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

Page({
  data: {
    store: null,
    todaySlots: [],
    loading: true,
    error: null
  },

  onLoad(query) {
    this.id = query.id;
    this.loadStore();
  },

  loadStore() {
    this.setData({ loading: true, error: null });
    api.getStoreDetail(this.id)
      .then(store => this.enrichStore(store))
      .then(store => {
        this.setData({ store, loading: false });
        this.loadTodaySlots();
      })
      .catch(err => {
        this.setData({ loading: false, error: err.message || '网络异常' });
      });
  },

  enrichStore(store) {
    const hasCoordinate = store.longitude && store.latitude;
    let distanceText = '';
    const app = getApp();
    const loc = app.globalData.userLocation;
    if (hasCoordinate && loc) {
      const d = distanceKm(loc.latitude, loc.longitude, store.latitude, store.longitude);
      distanceText = fmtDistance(d);
    }
    return { ...store, hasCoordinate, distanceText };
  },

  loadTodaySlots() {
    const date = todayStr();
    api.getStoreSlots(this.id, date)
      .then(resp => {
        // 后端返回 {date, slots: [...]}，兼容直接返回数组的旧版
        const list = (resp && Array.isArray(resp)) ? resp : (resp && resp.slots) || [];
        // 只展示前 6 个时段
        const slots = list.filter(s => s.state !== 'closed').slice(0, 6);
        this.setData({ todaySlots: slots });
      })
      .catch(() => {
        this.setData({ todaySlots: [] });
      });
  },

  onRetry() {
    this.loadStore();
  },

  onCall() {
    const store = this.data.store;
    if (!store || !store.contactPhone) {
      wx.showToast({ title: '暂无联系电话', icon: 'none' });
      return;
    }
    wx.makePhoneCall({ phoneNumber: String(store.contactPhone).replace(/\s/g, '') }).catch(() => {});
  },

  // 地图导航
  onNavigate() {
    const store = this.data.store;
    if (!store) return;
    if (!store.longitude || !store.latitude) {
      wx.showToast({ title: '该门店暂未配置位置', icon: 'none' });
      return;
    }
    wx.openLocation({
      latitude: store.latitude,
      longitude: store.longitude,
      name: store.name,
      address: store.address,
      scale: 16
    });
  },

  // 预约到店
  onAppointment() {
    const store = this.data.store;
    wx.navigateTo({
      url: '/pkg-customer/appointment-create/index?storeId=' + store.id,
      fail: () => wx.showToast({ title: '预约功能建设中', icon: 'none' })
    });
  }
});