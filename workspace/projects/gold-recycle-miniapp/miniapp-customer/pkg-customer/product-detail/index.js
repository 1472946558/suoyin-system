// pkg-customer/product-detail/index.js - 款式详情页
const { api } = require('../../../utils/request.js');
const { fmtPrice, absUrl } = require('../../../utils/customer-services.js');
const { distanceKm, fmtDistance } = require('../../../utils/customer-distance.js');

Page({
  data: {
    product: null,
    loading: true,
    error: null
  },

  onLoad(query) {
    this.id = query.id;
    this.loadDetail();
  },

  loadDetail() {
    this.setData({ loading: true, error: null });
    Promise.all([
      api.getProductDetail(this.id),
      api.getStores().catch(() => [])
    ])
      .then(([product, stores]) => {
        const app = getApp();
        const loc = app.globalData.userLocation;
        const nearest = this.findNearestStore(stores, loc);
        const enriched = this.formatProduct(product, nearest);
        this.setData({ product: enriched, loading: false });
      })
      .catch(err => {
        this.setData({ loading: false, error: err.message || '网络异常' });
      });
  },

  formatProduct(p, nearestStore) {
    const images = [];
    if (p.imageUrl) images.push(absUrl(p.imageUrl));
    const priceMain = p.retailPrice > 0 ? fmtPrice(p.retailPrice) : '面议';
    const description = this.buildDescription(p);
    const nearest = nearestStore ? {
      ...nearestStore,
      distanceText: nearestStore.distanceText || '附近'
    } : null;
    return {
      ...p,
      images,
      priceMain,
      description,
      nearestStore: nearest
    };
  },

  // 基于已有字段生成款式说明文本
  buildDescription(p) {
    const parts = [];
    if (p.purity) parts.push(p.purity + '材质');
    if (p.recommendedScene) parts.push(p.recommendedScene);
    if (p.tags && p.tags.length) parts.push(p.tags.filter(t => t !== '新建' && t !== '导入' && t !== '待完善').join('、'));
    if (parts.length === 0) {
      return '金匠倌匠心出品，采用传统工艺与时尚设计结合，每一件作品都经过精工细作，承载匠人精神。';
    }
    return parts.join('，') + '，工艺精湛，款式典雅，适合日常佩戴及送礼，支持旧金换新与定制参考。';
  },

  // 找到最近的门店（带距离文案）
  findNearestStore(stores, loc) {
    if (!stores || stores.length === 0) return null;
    if (!loc) {
      // 无定位：返回第一个有联系方式的门店
      return stores[0];
    }
    // 列表接口可能不含经纬度，直接复用已有距离数据（如有）
    const withDist = stores.map(s => {
      if (s.longitude && s.latitude) {
        const d = distanceKm(loc.latitude, loc.longitude, s.latitude, s.longitude);
        return { ...s, distance: d, distanceText: fmtDistance(d) };
      }
      return { ...s, distance: null, distanceText: '' };
    }).sort((a, b) => {
      const da = a.distance == null ? Infinity : a.distance;
      const db = b.distance == null ? Infinity : b.distance;
      return da - db;
    });
    return withDist[0];
  },

  onRetry() {
    this.loadDetail();
  },

  onTapStores() {
    wx.switchTab({ url: '/pages/stores/index' });
  },

  onTapNearestStore(e) {
    const id = e.currentTarget.dataset.id;
    if (id) wx.navigateTo({ url: '/pkg-customer/store-detail/index?id=' + id });
  },

  // 咨询门店 → 门店列表
  onConsult() {
    wx.switchTab({ url: '/pages/stores/index' });
  },

  // 到店预约 → 选门店并创建预约
  onAppointment() {
    const store = this.data.product && this.data.product.nearestStore;
    const url = store
      ? '/pkg-customer/appointment-create/index?storeId=' + store.id
      : '/pkg-customer/appointment-create/index';
    wx.navigateTo({ url });
  },

  // 拨打客服电话
  onCall() {
    const app = getApp();
    const phone = app.globalData.servicePhone;
    if (!phone) {
      wx.showToast({ title: '暂无联系电话', icon: 'none' });
      return;
    }
    wx.makePhoneCall({ phoneNumber: String(phone).replace(/\s/g, '') }).catch(() => {});
  }
});