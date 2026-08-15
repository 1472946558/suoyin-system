// pages/home/index.js
const { api } = require('../../utils/request.js');
const { ensureCustomerSession } = require('../../utils/customer-auth.js');
const { distanceKm, fmtDistance } = require('../../utils/customer-distance.js');

Page({
  data: {
    homeData: {},
    nearestStore: null,
    servicePhone: '155****5010'
  },

  onLoad() {
    const app = getApp();
    this.setData({ servicePhone: app.globalData.servicePhone });
    ensureCustomerSession().catch(() => {});
  },

  onShow() {
    this.loadHome();
    this.loadStores();
  },

  loadHome() {
    api.getHome()
      .then(data => this.setData({ homeData: data || {} }))
      .catch(err => {
        console.warn('加载首页失败', err);
        this.setData({ homeData: { brandName: '金匠倌', brandSlogan1: '旧金换打新款', brandSlogan2: '包损耗' } });
      });
  },

  loadStores() {
    api.getStores()
      .then(stores => {
        if (!stores || stores.length === 0) {
          this.setData({ nearestStore: null });
          return;
        }
        // 取第一家作为最近门店（实际应按距离排序）
        const first = stores[0];
        first.distance = 0.8;
        first.distanceText = fmtDistance(0.8);
        this.setData({ nearestStore: first });
      })
      .catch(err => {
        console.warn('加载门店失败', err);
        this.setData({ nearestStore: null });
      });
  },

  onTapMore() {
    wx.showToast({ title: '更多功能开发中', icon: 'none' });
  },

  onTapScan() {
    wx.scanCode({ success: () => {}, fail: () => {} });
  },

  onBannerTap(e) {
    const idx = e.currentTarget.dataset.index;
    const banner = this.data.homeData.banners[idx];
    if (!banner) return;
    if (banner.linkType === 'products') {
      wx.switchTab({ url: '/pages/products/index' });
    } else if (banner.linkType === 'stores') {
      wx.switchTab({ url: '/pages/stores/index' });
    } else if (banner.linkType === 'recycle') {
      wx.navigateTo({ url: '/pkg-customer/recycle-info/index' });
    } else if (banner.linkType === 'appointment') {
      this.goAppointment();
    } else if (banner.linkUrl) {
      wx.navigateTo({ url: banner.linkUrl });
    }
  },

  onTapProducts() {
    wx.switchTab({ url: '/pages/products/index' });
  },

  onTapMoreStores() {
    wx.switchTab({ url: '/pages/stores/index' });
  },

  onTapStore(e) {
    const store = e.currentTarget.dataset.store;
    wx.navigateTo({ url: '/pkg-customer/store-detail/index?id=' + store.id });
  },

  onTapAppointment(e) {
    const store = e.currentTarget.dataset.store;
    if (!store) {
      this.goAppointment();
      return;
    }
    wx.navigateTo({ url: '/pkg-customer/appointment-create/index?storeId=' + store.id });
  },

  onTapNavigate(e) {
    const store = e.currentTarget.dataset.store;
    if (!store) return;
    if (!store.longitude || !store.latitude) {
      // 详情页获取经纬度
      api.getStoreDetail(store.id).then(detail => {
        if (detail.longitude && detail.latitude) {
          this.openMap(detail);
        } else {
          wx.showToast({ title: '该门店暂未配置位置', icon: 'none' });
        }
      }).catch(() => {
        wx.showToast({ title: '导航失败', icon: 'none' });
      });
      return;
    }
    this.openMap(store);
  },

  openMap(store) {
    wx.openLocation({
      latitude: store.latitude,
      longitude: store.longitude,
      name: store.name,
      address: store.address,
      scale: 16
    });
  },

  goAppointment() {
    wx.navigateTo({ url: '/pkg-customer/appointment-create/index' });
  }
});
