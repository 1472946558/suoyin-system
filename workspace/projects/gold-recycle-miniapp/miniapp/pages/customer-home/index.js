// pages/home/index.js
const { api } = require('../../utils/request.js');
const { ensureCustomerSession } = require('../../utils/customer-auth.js');
const { distanceKm, fmtDistance } = require('../../utils/customer-distance.js');
const { absUrl } = require('../../utils/customer-services.js');

// 后台尚未配置 Banner 时，仍保留可点击的顾客首页轮播骨架；配置真实图片后自动替换。
const DEFAULT_BANNERS = [
  {
    id: 'default-style-banner',
    title: '款式图 / 工费',
    subtitle: '海量款式 · 工费透明',
    linkType: 'style_list',
    theme: 'style'
  },
  {
    id: 'default-store-banner',
    title: '附近门店',
    subtitle: '就近选择，到店更方便',
    linkType: 'store_list',
    theme: 'store'
  },
  {
    id: 'default-recycle-banner',
    title: '黄金回收服务',
    subtitle: '到店检测，现场沟通',
    linkType: 'recycle',
    theme: 'recycle'
  }
];

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
        const serverBanners = data && Array.isArray(data.banners) ? data.banners : [];
        const sourceBanners = serverBanners.length > 0 ? serverBanners : DEFAULT_BANNERS;
        const homeData = Object.assign({}, data || {}, {
          banners: sourceBanners.map(banner => Object.assign({}, banner, {
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
        this.setData({
          homeData: {
            brandName: '金匠倌',
            brandSlogan1: '旧金换打新款',
            brandSlogan2: '包损耗',
            banners: DEFAULT_BANNERS
          }
        });
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
            distanceText: '',
            canNavigate: this.hasCoordinates(store)
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
        wx.switchTab({
          url: '/pages/customer-products/index',
          fail: () => this.showActionError('款式页暂不可用')
        });
        break;
      case 'stores':
      case 'store_list':
        wx.switchTab({
          url: '/pages/customer-stores/index',
          fail: () => this.showActionError('门店页暂不可用')
        });
        break;
      case 'recycle':
        wx.navigateTo({
          url: '/pkg-customer/recycle-info/index',
          fail: () => this.showActionError('回收介绍页暂不可用')
        });
        break;
      case 'appointment':
      case 'booking':
        this.goAppointment();
        break;
      case 'style_detail':
        if (target) wx.navigateTo({
          url: '/pkg-customer/product-detail/index?id=' + encodeURIComponent(target),
          fail: () => this.showActionError('款式详情暂不可用')
        });
        break;
      case 'store_detail':
        if (target) wx.navigateTo({
          url: '/pkg-customer/store-detail/index?id=' + encodeURIComponent(target),
          fail: () => this.showActionError('门店详情暂不可用')
        });
        break;
      case 'external_page':
        if (banner.linkUrl) wx.navigateTo({
          url: banner.linkUrl,
          fail: () => this.showActionError('内容页面暂不可用')
        });
        break;
      default:
        if (banner.linkUrl) wx.navigateTo({
          url: banner.linkUrl,
          fail: () => this.showActionError('内容页面暂不可用')
        });
        break;
    }
  },

  onTapProducts() {
    wx.switchTab({
      url: '/pages/customer-products/index',
      fail: () => this.showActionError('款式页暂不可用')
    });
  },

  onTapMoreStores() {
    wx.switchTab({
      url: '/pages/customer-stores/index',
      fail: () => this.showActionError('门店页暂不可用')
    });
  },

  onTapStore(e) {
    const store = e.currentTarget.dataset.store;
    if (!store || !store.id) return;
    wx.navigateTo({
      url: '/pkg-customer/store-detail/index?id=' + encodeURIComponent(store.id),
      fail: () => this.showActionError('门店详情暂不可用')
    });
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
    wx.navigateTo({
      url: '/pkg-customer/appointment-create/index?storeId=' + encodeURIComponent(store.id),
      fail: () => this.showActionError('预约页面暂不可用')
    });
  },

  onTapNavigate(e) {
    const store = e.currentTarget.dataset.store;
    if (!store || !store.id) return;
    if (!this.hasCoordinates(store)) {
      // 详情页获取经纬度
      api.getStoreDetail(store.id).then(detail => {
        if (this.hasCoordinates(detail)) {
          this.openMap(detail);
        } else {
          this.showActionError('该门店暂未配置经纬度');
        }
      }).catch(() => {
        this.showActionError('获取门店位置失败');
      });
      return;
    }
    this.openMap(store);
  },

  openMap(store) {
    wx.openLocation({
      latitude: Number(store.latitude),
      longitude: Number(store.longitude),
      name: store.name,
      address: store.address,
      scale: 16,
      fail: () => this.showActionError('打开地图失败，请稍后重试')
    });
  },

  goAppointment() {
    wx.navigateTo({
      url: '/pkg-customer/appointment-create/index',
      fail: () => this.showActionError('预约页面暂不可用')
    });
  },

  hasCoordinates(store) {
    const longitude = Number(store && store.longitude);
    const latitude = Number(store && store.latitude);
    return isFinite(longitude) && isFinite(latitude) && longitude !== 0 && latitude !== 0;
  },

  showActionError(title) {
    wx.showToast({ title, icon: 'none', duration: 1800 });
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
