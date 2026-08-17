// pages/products/index.js - 款式列表页（款式图 / 工费）
// 分类筛选 + 搜索 + 双列卡片 + 触底分页
const { api } = require('../../utils/request.js');
const { fmtPrice, absUrl } = require('../../utils/customer-services.js');

const PAGE_SIZE = 20;

Page({
  data: {
    categories: ['全部'],
    activeCategory: '全部',
    products: [],
    keyword: '',
    page: 1,
    total: 0,
    hasMore: true,
    loading: false,
    error: null
  },

  onLoad() {
    this.loadProducts(true);
  },

  onPullDownRefresh() {
    this.loadProducts(true).finally(() => wx.stopPullDownRefresh());
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadProducts(false);
    }
  },

  loadProducts(reset) {
    if (this.data.loading) return Promise.resolve();
    const { activeCategory, keyword, page } = this.data;
    const targetPage = reset ? 1 : page + 1;
    const category = activeCategory === '全部' ? '' : activeCategory;
    const search = keyword.trim();

    this.setData({ loading: true, error: null });
    return api.getProducts({ category, keyword: search, page: targetPage, pageSize: PAGE_SIZE })
      .then(data => {
        const items = (data.items || []).map(this.formatProduct);
        const cats = ['全部'].concat(data.categories || []);
        const hasMore = targetPage * PAGE_SIZE < (data.total || 0) && items.length > 0;
        this.setData({
          categories: reset ? cats : this.data.categories,
          products: reset ? items : this.data.products.concat(items),
          page: targetPage,
          total: data.total || 0,
          hasMore,
          loading: false
        });
      })
      .catch(err => {
        this.setData({ loading: false, error: err.message || '网络异常' });
      });
  },

  formatProduct(p) {
    const tags = Array.isArray(p.tags) ? p.tags : [];
    return Object.assign({}, p, {
      imageUrl: absUrl(p.imageUrl),
      priceText: p.retailPrice > 0 ? '¥' + fmtPrice(p.retailPrice) : '面议',
      gramText: p.gramWeight > 0 ? p.gramWeight + 'g' : '',
      laborFeeText: p.laborFeeRef || '',
      showHot: !!p.isHot,
      showRecommended: !!p.isRecommended,
      tagText: tags.length > 0 ? tags.join(' · ') : ''
    });
  },

  onCategoryTap(e) {
    const cat = e.currentTarget.dataset.cat;
    if (cat === this.data.activeCategory) return;
    this.setData({ activeCategory: cat });
    this.loadProducts(true);
  },

  onSearchInput(e) {
    this.setData({ keyword: e.detail.value });
  },

  onSearchConfirm() {
    this.loadProducts(true);
  },

  onClearSearch() {
    this.setData({ keyword: '' });
    this.loadProducts(true);
  },

  onProductTap(e) {
    const id = e.currentTarget.dataset.id;
    wx.navigateTo({ url: '/pkg-customer/product-detail/index?id=' + id });
  },

  onRetry() {
    this.loadProducts(true);
  }
});
