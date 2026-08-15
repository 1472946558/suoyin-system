// pages/home/index.js
const { api } = require('../../utils/request.js');
const { ensureCustomerSession } = require('../../utils/customer-auth.js');
const { distanceKm, fmtDistance } = require('../../utils/customer-distance.js');
const { absUrl } = require('../../utils/customer-services.js');

Page({
  data: {
    homeData: {},
    nearestStore: null,
    servicePhone: ''
  },

  onLoad() {
    ensureCustomerSession().catch(() => {});
  },

  onShow() {
    this.loadHome();
    this.loadStores();
  },

  loadHome() {
    api.getHome()
      .then(data => {
        const homeData = data || {};
        this.setData({ homeData, servicePhone: homeData.servicePhone || '' });
        // 缓存到 globalData 供其他页面使用
        const app = getApp();
        app.globalData.homeConfig = homeData;
        app.globalData.servicePhone = homeData.servicePhone || '';
        app.globalData.brandName = homeData.brandName || app.globalData.brandName;
      })
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
        // 取第一家作为最近门店
        const first = stores[0];
        // 使用真实距离（如果有经纬度）
        const app = getApp();
        const loc = app.globalData.userLocation;
        if (loc && first.latitude && first.longitude) {
          first.distance = distanceKm(loc.latitude, loc.longitude, first.latitude, first.longitude);
          first.distanceText = fmtDistance(first.distance);
        } else {
          first.distance = 0;
          first.distanceText = '';
        }
        // 补全门店图片 URL
        if (first.imageUrl) {
          first.thumb = absUrl(first.imageUrl);
        }
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
    const linkType = banner.linkType;
    const target = banner.linkTarget;
    // 兼容旧值 + 新值
    switch (linkType) {
      case 'products':
      case 'style_list':
        wx.switchTab({ url: '/pages/products/index' });
        break;
      case 'stores':
      case 'store_list':
        wx.switchTab({ url: '/pages/stores/index' });
        break;
      case 'recycle':
        wx.navigateTo({ url: '/pkg-customer/recycle-info/index' });
        break;
      case 'appointment':
      case 'booking':
        this.goAppointment();
        break;
      case 'style_detail':
        if (target) wx.navigateTo({ url: '/pkg-customer/product-detail/index?id=' + target });
        break;
      case 'store_detail':
        if (target) wx.navigateTo({ url: '/pkg-customer/store-detail/index?id=' + target });
        break;
      case 'external_page':
        if (banner.linkUrl) wx.navigateTo({ url: banner.linkUrl });
        break;
      default:
        if (banner.linkUrl) wx.navigateTo({ url: banner.linkUrl });
        break;
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
    // 检查门店预约开关
    if (store.appointmentEnabled === false) {
      wx.showToast({ title: '该门店暂未开放预约', icon: 'none' });
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
