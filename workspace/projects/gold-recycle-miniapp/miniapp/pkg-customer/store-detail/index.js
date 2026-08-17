// pkg-customer/store-detail/index.js - 门店详情页
const { api } = require('../../utils/request.js');
const { distanceKm, fmtDistance } = require('../../utils/customer-distance.js');
const { absUrl, serviceTypeText } = require('../../utils/customer-services.js');

function hasValidCoordinates(store) {
  const longitude = Number(store && store.longitude);
  const latitude = Number(store && store.latitude);
  return isFinite(longitude) && isFinite(latitude) && longitude !== 0 && latitude !== 0;
}

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
        // 预约关闭时不加载时段
        if (store.appointmentEnabled !== false) {
          this.loadTodaySlots();
        }
      })
      .catch(err => {
        this.setData({ loading: false, error: err.message || '网络异常' });
      });
  },

  enrichStore(store) {
    const hasCoordinate = hasValidCoordinates(store);
    let distanceText = '';
    const app = getApp();
    const loc = app.globalData.userLocation;
    if (hasCoordinate && loc) {
      const d = distanceKm(loc.latitude, loc.longitude, store.latitude, store.longitude);
      distanceText = fmtDistance(d);
    }
    // 补全门店图片 URL
    let thumb = '';
    if (store.imageUrl) {
      thumb = absUrl(store.imageUrl);
    }
    // 服务标签：优先使用后端 serviceTags，否则用默认
    let serviceTags = [];
    if (Array.isArray(store.serviceTags) && store.serviceTags.length > 0) {
      serviceTags = store.serviceTags.map(tag => ({
        text: tag,
        icon: this.tagIcon(tag)
      }));
    } else {
      serviceTags = [
        { text: '旧金换新', icon: '♻' },
        { text: '黄金维修', icon: '🔧' },
        { text: '款式咨询', icon: '💬' },
        { text: '到店回收', icon: '💎' }
      ];
    }
    return Object.assign({}, store, { hasCoordinate, distanceText, thumb, serviceTags });
  },

  tagIcon(tag) {
    const t = (tag || '').toLowerCase();
    if (t.indexOf('换') >= 0) return '♻';
    if (t.indexOf('修') >= 0) return '🔧';
    if (t.indexOf('咨询') >= 0 || t.indexOf('工费') >= 0) return '💬';
    if (t.indexOf('回收') >= 0) return '💎';
    return '✦';
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
    if (!hasValidCoordinates(store)) {
      wx.showToast({ title: '该门店暂未配置位置', icon: 'none' });
      return;
    }
    wx.openLocation({
      latitude: Number(store.latitude),
      longitude: Number(store.longitude),
      name: store.name,
      address: store.address,
      scale: 16,
      fail: () => wx.showToast({ title: '打开地图失败，请稍后重试', icon: 'none' })
    });
  },

  // 预约到店
  onAppointment() {
    const store = this.data.store;
    if (!store) return;
    if (store.appointmentEnabled === false) {
      wx.showToast({ title: '该门店暂未开放预约', icon: 'none' });
      return;
    }
    wx.navigateTo({
      url: '/pkg-customer/appointment-create/index?storeId=' + encodeURIComponent(store.id),
      fail: () => wx.showToast({ title: '预约功能建设中', icon: 'none' })
    });
  }
});
