// pkg-customer/product-detail/index.js - 款式详情页
const { api } = require('../../utils/request.js');
const { fmtPrice, absUrl, serviceTypeText } = require('../../utils/customer-services.js');
const { distanceKm, fmtDistance } = require('../../utils/customer-distance.js');

Page({
  data: {
    product: null,
    loading: true,
    error: null
  },

  onLoad(query) {
    this.id = query.id;
    this.loadProduct();
  },

  loadProduct() {
    this.setData({ loading: true, error: null });
    api.getProductDetail(this.id)
      .then(p => this.enrichProduct(p))
      .then(product => {
        this.setData({ product, loading: false });
        this.loadNearestStore();
      })
      .catch(err => {
        this.setData({ loading: false, error: err.message || '网络异常' });
      });
  },

  enrichProduct(p) {
    if (!p) return null;
    // 补全图片 URL（主图 + 详情图合并进轮播）
    const images = Array.isArray(p.images) && p.images.length > 0
      ? p.images.map(absUrl)
      : [absUrl(p.imageUrl)].filter(Boolean);
    if (Array.isArray(p.detailImages) && p.detailImages.length > 0) {
      p.detailImages.forEach(u => {
        const abs = absUrl(u);
        if (abs && images.indexOf(abs) < 0) images.push(abs);
      });
    }
    // 价格格式化
    const priceMain = p.retailPrice > 0 ? fmtPrice(p.retailPrice) : '面议';
    // 适用服务类型文案
    const serviceTypeTexts = Array.isArray(p.applicableServiceTypes) && p.applicableServiceTypes.length > 0
      ? p.applicableServiceTypes.map(serviceTypeText)
      : [];
    return Object.assign({}, p, {
      images,
      priceMain,
      serviceTypeTexts
    });
  },

  loadNearestStore() {
    api.getStores()
      .then(stores => {
        if (!stores || stores.length === 0) return;
        const app = getApp();
        const loc = app.globalData.userLocation;
        let nearest = stores[0];
        let minDist = Infinity;
        stores.forEach(s => {
          if (s.latitude && s.longitude && loc) {
            const d = distanceKm(loc.latitude, loc.longitude, s.latitude, s.longitude);
            if (d < minDist) {
              minDist = d;
              nearest = s;
            }
          }
        });
        if (nearest && loc && nearest.latitude) {
          nearest.distanceText = fmtDistance(minDist) + ' · ';
        } else {
          nearest.distanceText = '';
        }
        // 补全门店图片
        if (nearest.imageUrl) {
          nearest.thumb = absUrl(nearest.imageUrl);
        }
        this.setData({ 'product.nearestStore': nearest });
      })
      .catch(() => {});
  },

  onRetry() {
    this.loadProduct();
  },

  // 咨询门店 → 拨打最近门店电话
  onConsult() {
    const store = this.data.product && this.data.product.nearestStore;
    if (!store || !store.contactPhone) {
      wx.showToast({ title: '暂无门店电话', icon: 'none' });
      return;
    }
    wx.makePhoneCall({ phoneNumber: String(store.contactPhone).replace(/\s/g, '') }).catch(() => {});
  },

  // 到店预约
  onAppointment() {
    const store = this.data.product && this.data.product.nearestStore;
    if (store && store.appointmentEnabled === false) {
      wx.showToast({ title: '该门店暂未开放预约', icon: 'none' });
      return;
    }
    const storeId = store ? store.id : '';
    wx.navigateTo({
      url: '/pkg-customer/appointment-create/index' + (storeId ? '?storeId=' + storeId : '')
    });
  },

  onTapStores() {
    wx.switchTab({ url: '/pages/stores/index' });
  },

  onTapNearestStore(e) {
    const id = e.currentTarget.dataset.id;
    if (id) wx.navigateTo({ url: '/pkg-customer/store-detail/index?id=' + id });
  },

  // 预览大图
  onTapImage(e) {
    const idx = e.currentTarget.dataset.idx;
    const images = this.data.product.images;
    wx.previewImage({ current: images[idx], urls: images });
  }
});
