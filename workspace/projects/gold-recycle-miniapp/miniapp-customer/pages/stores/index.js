// pages/stores/index.js - 附近门店列表页
const { api } = require('../../utils/request.js');
const { distanceKm, fmtDistance } = require('../../utils/customer-distance.js');
const { absUrl } = require('../../utils/customer-services.js');

Page({
  data: {
    stores: [],
    filteredStores: [],
    nearestStore: null,
    currentCity: '',
    keyword: '',
    loading: true,
    error: null,
    located: false,
    locError: null
  },

  onShow() {
    this.loadStores();
    this.tryLocate();
  },

  onPullDownRefresh() {
    Promise.all([this.loadStores(), this.tryLocate()])
      .finally(() => wx.stopPullDownRefresh());
  },

  loadStores() {
    this.setData({ loading: true, error: null });
    return api.getStores()
      .then(stores => this.enrichStores(stores || []))
      .then(stores => this.applyLocation(stores))
      .catch(err => {
        this.setData({ loading: false, error: err.message || '网络异常' });
      });
  },

  // 列表接口无经纬度，并发拉详情补齐
  enrichStores(stores) {
    const tasks = stores
      .filter(s => !s.longitude)
      .map(s => api.getStoreDetail(s.id)
        .then(d => ({ id: s.id, longitude: d.longitude, latitude: d.latitude }))
        .catch(() => null));

    return Promise.all(tasks).then(list => {
      const map = {};
      (list || []).forEach(m => {
        if (m && m.longitude) map[m.id] = m;
      });
      return stores.map(s => ({
        ...s,
        longitude: s.longitude || (map[s.id] ? map[s.id].longitude : 0),
        latitude: s.latitude || (map[s.id] ? map[s.id].latitude : 0),
        thumb: s.imageUrl ? absUrl(s.imageUrl) : ''
      }));
    });
  },

  // 根据用户定位计算距离 + 当前城市
  applyLocation(stores) {
    const app = getApp();
    const loc = app.globalData.userLocation;
    let enriched = stores.map(s => ({ ...s, distanceText: '' }));
    let currentCity = '';

    if (loc) {
      enriched = enriched.map(s => {
        if (s.longitude && s.latitude) {
          const d = distanceKm(loc.latitude, loc.longitude, s.latitude, s.longitude);
          return { ...s, distance: d, distanceText: fmtDistance(d) };
        }
        return { ...s, distance: null, distanceText: '' };
      });
      // 按距离排序（无坐标的排最后）
      enriched.sort((a, b) => {
        const da = a.distance == null ? Infinity : a.distance;
        const db = b.distance == null ? Infinity : b.distance;
        return da - db;
      });
      currentCity = enriched.find(s => s.city) ? enriched.find(s => s.city).city : '';
      this.setData({ located: true, locError: null });
    }

    // 过滤关键词
    const filtered = this.applyKeyword(enriched);
    // 离你最近
    const nearest = filtered.find(s => s.distance != null) || filtered[0] || null;

    this.setData({
      stores: enriched,
      filteredStores: filtered,
      nearestStore: nearest ? this.tagOpen(nearest) : null,
      currentCity,
      loading: false
    });
  },

  // 关键词过滤
  applyKeyword(stores) {
    const k = (this.data.keyword || '').trim().toLowerCase();
    if (!k) return stores;
    return stores.filter(s =>
      (s.name && s.name.toLowerCase().indexOf(k) >= 0) ||
      (s.address && s.address.toLowerCase().indexOf(k) >= 0) ||
      (s.city && s.city.toLowerCase().indexOf(k) >= 0)
    );
  },

  // 给门店打"可预约"标（基础判定：营业时间内）
  tagOpen(store) {
    if (!store.businessHours) return { ...store, isOpen: true };
    const m = store.businessHours.match(/(\d{1,2}):(\d{2})\s*[-~]\s*(\d{1,2}):(\d{2})/);
    if (!m) return { ...store, isOpen: true };
    const now = new Date();
    const cur = now.getHours() * 60 + now.getMinutes();
    const open = parseInt(m[1]) * 60 + parseInt(m[2]);
    const close = parseInt(m[3]) * 60 + parseInt(m[4]);
    return { ...store, isOpen: cur >= open && cur <= close };
  },

  tryLocate() {
    const app = getApp();
    if (app.globalData.userLocation) return Promise.resolve();
    return new Promise(resolve => {
      wx.getLocation({
        type: 'gcj02',
        success: res => {
          app.globalData.userLocation = { latitude: res.latitude, longitude: res.longitude };
          this.applyLocation(this.data.stores);
          resolve();
        },
        fail: () => {
          const app = getApp();
          const note = (app.globalData.homeConfig && app.globalData.homeConfig.locationPermissionNote) || '定位服务未开启，无法获取附近门店';
          this.setData({ locError: note });
          resolve();
        }
      });
    });
  },

  onSearchInput(e) {
    this.setData({ keyword: e.detail.value }, () => {
      this.applyLocation(this.data.stores);
    });
  },

  onClearSearch() {
    this.setData({ keyword: '' }, () => {
      this.applyLocation(this.data.stores);
    });
  },

  onRelocate() {
    const app = getApp();
    delete app.globalData.userLocation;
    this.setData({ locError: null });
    wx.showLoading({ title: '定位中...', mask: true });
    this.tryLocate().finally(() => wx.hideLoading());
  },

  onRetry() {
    this.loadStores();
  },

  onCall(e) {
    const phone = e.currentTarget.dataset.phone;
    if (!phone) {
      wx.showToast({ title: '暂无联系电话', icon: 'none' });
      return;
    }
    wx.makePhoneCall({ phoneNumber: String(phone).replace(/\s/g, '') }).catch(() => {});
  },

  onTapStore(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({ url: '/pkg-customer/store-detail/index?id=' + id });
  },

  onTapAppointment(e) {
    const id = e.currentTarget.dataset.id;
    const s = this.data.stores.find(x => x.id === id);
    if (s && s.appointmentEnabled === false) {
      wx.showToast({ title: '该门店暂未开放预约', icon: 'none' });
      return;
    }
    wx.navigateTo({ url: '/pkg-customer/appointment-create/index?storeId=' + id });
  },

  onNavigate(e) {
    const id = e.currentTarget.dataset.id;
    const s = this.data.stores.find(x => x.id === id);
    if (!s || !s.longitude || !s.latitude) {
      wx.showToast({ title: '该门店暂未配置坐标', icon: 'none' });
      return;
    }
    wx.openLocation({
      latitude: s.latitude,
      longitude: s.longitude,
      name: s.name,
      address: s.address,
      scale: 16
    });
  },

  onOpenSetting() {
    wx.openSetting({ success: () => this.tryLocate() });
  }
});