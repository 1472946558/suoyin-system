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
    this.tryLocate().finally(() => this.loadStores());
  },

  loadHome() {
    api.getHome()
      .then(data => {
        const homeData = Object.assign({}, data || {}, {
          banners: (data && Array.isArray(data.banners) ? data.banners : []).map(banner => Object.assign({}, banner, {
            imageUrl: absUrl(banner.imageUrl)
          }))
        });
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
        const app = getApp();
        const loc = app.globalData.userLocation;
        const enriched = stores.map(store => {
          const item = Object.assign({}, store, {
            thumb: store.imageUrl ? absUrl(store.imageUrl) : '',
            distance: null,
            distanceText: ''
          });
          if (loc && Number(item.latitude) && Number(item.longitude)) {
            item.distance = distanceKm(loc.latitude, loc.longitude, item.latitude, item.longitude);
            item.distanceText = fmtDistance(item.distance);
          }
          return item;
        });
        enriched.sort((a, b) => {
          const da = a.distance == null ? Infinity : a.distance;
          const db = b.distance == null ? Infinity : b.distance;
          return da - db;
        });
        this.setData({ nearestStore: enriched[0] || null });
      })
      .catch(err => {
        console.warn('加载门店失败', err);
        this.setData({ nearestStore: null });
      });
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
        wx.switchTab({ url: '/pages/customer-products/index' });
        break;
      case 'stores':
      case 'store_list':
        wx.switchTab({ url: '/pages/customer-stores/index' });
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
    wx.switchTab({ url: '/pages/customer-products/index' });
  },

  onTapMoreStores() {
    wx.switchTab({ url: '/pages/customer-stores/index' });
  },

  onTapStore(e) {
    const store = e.currentTarget.dataset.store;
    if (!store || !store.id) return;
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
    if (!store || !store.id) return;
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
  },

  tryLocate() {
    const app = getApp();
    if (app.globalData.userLocation || app.globalData.locationAttempted) {
      return Promise.resolve(app.globalData.userLocation);
    }

    app.globalData.locationAttempted = true;
    return new Promise(resolve => {
      wx.getLocation({
        type: 'gcj02',
        success: res => {
          app.globalData.userLocation = {
            latitude: res.latitude,
            longitude: res.longitude
          };
          resolve(app.globalData.userLocation);
        },
        fail: () => {
          app.globalData.locationDenied = true;
          resolve(null);
        }
      });
    });
  }
});
