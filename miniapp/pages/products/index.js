const { listProductsOnline, getProductStats } = require("../../utils/catalogStore");
const { mergeDraft } = require("../../utils/orderStore");
const { canAccessFeature, getProfile } = require("../../utils/userStore");

function formatCurrency(amount) {
  return `¥${Number(amount || 0).toFixed(2)}`;
}

function filterProducts(products, filters) {
  return (products || []).filter(function(item) {
    const keyword = (filters.keyword || "").toLowerCase();
    const matchesKeyword = !keyword || [item.name, item.sku, item.category, item.purity, item.tags.join(" / ")]
      .join(" ")
      .toLowerCase()
      .indexOf(keyword) > -1;
    const matchesCategory = filters.category === "all" || item.category === filters.category;
    const matchesStatus = filters.status === "all" || item.status === filters.status;
    return matchesKeyword && matchesCategory && matchesStatus;
  });
}

Page({
  data: {
    profile: getProfile(),
    products: [],
    filteredProducts: [],
    selectedId: "",
    selectedProduct: null,
    stats: {
      total: 0,
      activeCount: 0,
      lowStockCount: 0,
      draftCount: 0
    },
    filters: {
      keyword: "",
      category: "all",
      status: "all"
    },
    categoryOptions: ["all", "金饰", "金条", "K金", "旧金料"],
    statusOptions: ["all", "active", "draft", "disabled"],
    loading: false
  },

  onShow() {
    if (!canAccessFeature("products")) {
      wx.showToast({ title: "当前账号没有商品查看权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    this.loadProducts();
  },

  loadProducts() {
    this.setData({ loading: true, profile: getProfile() });
    listProductsOnline().then((products) => {
      this.setData({
        products: products,
        stats: getProductStats(products)
      });
      this.applyFilters();
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  applyFilters() {
    const filteredProducts = filterProducts(this.data.products, this.data.filters);
    const selectedId = filteredProducts.some(function(item) {
      return item.id === this.data.selectedId;
    }, this)
      ? this.data.selectedId
      : (filteredProducts[0] && filteredProducts[0].id) || "";
    const selectedProduct = filteredProducts.find(function(item) {
      return item.id === selectedId;
    }) || null;

    this.setData({
      filteredProducts: filteredProducts,
      selectedId: selectedId,
      selectedProduct: selectedProduct
    });
  },

  onKeywordInput(event) {
    const filters = Object.assign({}, this.data.filters, {
      keyword: event.detail.value
    });
    this.setData({ filters: filters });
    this.applyFilters();
  },

  onCategoryChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      category: this.data.categoryOptions[index]
    });
    this.setData({ filters: filters });
    this.applyFilters();
  },

  onStatusChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      status: this.data.statusOptions[index]
    });
    this.setData({ filters: filters });
    this.applyFilters();
  },

  selectProduct(event) {
    const id = event.currentTarget.dataset.id;
    const selectedProduct = this.data.filteredProducts.find(function(item) {
      return item.id === id;
    }) || null;
    this.setData({
      selectedId: id,
      selectedProduct: selectedProduct
    });
  },

  useForRecycle() {
    const product = this.data.selectedProduct;
    if (!product) return;
    mergeDraft({
      itemCategory: product.category,
      itemName: product.name,
      purity: product.purity,
      recyclePrice: String(product.benchPrice),
      remark: product.recommendedScene
    });
    wx.showToast({ title: "商品参数已带入录单草稿", icon: "none" });
    wx.switchTab({ url: "/pages/track/index" });
  },

  copySku() {
    const product = this.data.selectedProduct;
    if (!product) return;
    wx.setClipboardData({
      data: product.sku
    });
  },

  formatCurrency(amount) {
    return formatCurrency(amount);
  }
});
