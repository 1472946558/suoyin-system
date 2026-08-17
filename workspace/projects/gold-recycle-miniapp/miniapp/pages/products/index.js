/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: index.js
 * 功能描述: 页面模块
 * 作者: 廖心慈
 * 创建日期: 2026-06-03
 */

const {
  listProductsOnline,
  getProductStats
} = require("../../utils/catalogStore");
const { mergeDraft } = require("../../utils/orderStore");
const { canAccessFeature, getProfile } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");

const CATEGORY_OPTIONS = ["all", "金饰", "金条", "K金", "旧金料"];
const STATUS_FILTER_VALUES = ["all", "active", "draft", "disabled"];
const STATUS_FILTER_LABELS = ["全部状态", "上架中", "草稿", "停用"];
const STATUS_TEXT = {
  active: "上架中",
  draft: "草稿",
  disabled: "停用"
};

function formatCurrency(amount) {
  return `¥${Number(amount || 0).toFixed(2)}`;
}

function normalizeProductView(product) {
  const nextProduct = Object.assign({}, product || {});
  nextProduct.tags = Array.isArray(nextProduct.tags) ? nextProduct.tags : [];
  nextProduct.stores = Array.isArray(nextProduct.stores) ? nextProduct.stores : [];
  nextProduct.imageUrl = nextProduct.imageUrl || "/assets/ui/price.png";
  nextProduct.categoryLabel = nextProduct.categoryTab || nextProduct.category || "未分类";
  nextProduct.storesText = nextProduct.stores.length ? nextProduct.stores.join(" / ") : "未分配";
  nextProduct.benchPriceText = `${formatCurrency(nextProduct.benchPrice)}/g`;
  nextProduct.statusText = STATUS_TEXT[nextProduct.status] || "未设置";
  nextProduct.stockText = nextProduct.status === "disabled"
    ? "停用"
    : (nextProduct.stockStatus === "low" ? "低数量" : `${Number(nextProduct.inventory || 0)} 件`);
  return nextProduct;
}

function filterProducts(products, filters) {
  return (products || []).filter(function(item) {
    const keyword = (filters.keyword || "").toLowerCase();
    const matchesKeyword = !keyword || [item.name, item.sku, item.category, item.purity, (item.tags || []).join(" / ")]
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
    categoryOptions: CATEGORY_OPTIONS,
    statusFilterValues: STATUS_FILTER_VALUES,
    statusFilterLabels: STATUS_FILTER_LABELS,
    statusFilterIndex: 0,
    loading: false
  },

  onShow() {
    if (!canAccessFeature("products")) {
      wx.showToast({ title: "当前账号没有商品查看权限", icon: "none" });
      wx.reLaunch({ url: "/pages/home/index" });
      return;
    }
    this.loadProducts();
  },

  loadProducts() {
    this.setData({ loading: true, profile: getProfile() });
    listProductsOnline().then((result) => {
      let products = result && Array.isArray(result.items) ? result.items : [];
      products = products.map(normalizeProductView);
      this.setData({
        products,
        stats: getProductStats(products)
      });
      this.applyFilters();
    }).catch((error) => {
      this.setData({
        products: [],
        filteredProducts: [],
        selectedId: "",
        selectedProduct: null,
        stats: getProductStats([])
      });
      wx.showToast({ title: formatRequestError(error), icon: "none" });
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
      filteredProducts,
      selectedId,
      selectedProduct
    });
  },

  onKeywordInput(event) {
    const filters = Object.assign({}, this.data.filters, {
      keyword: event.detail.value
    });
    this.setData({ filters });
    this.applyFilters();
  },

  onCategoryChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      category: this.data.categoryOptions[index]
    });
    this.setData({ filters });
    this.applyFilters();
  },

  onCategoryTap(event) {
    const filters = Object.assign({}, this.data.filters, {
      category: event.currentTarget.dataset.category || "all"
    });
    this.setData({ filters });
    this.applyFilters();
  },

  onStatusChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      status: this.data.statusFilterValues[index]
    });
    this.setData({
      filters,
      statusFilterIndex: index
    });
    this.applyFilters();
  },

  selectProduct(event) {
    const id = event.currentTarget.dataset.id;
    const selectedProduct = this.data.filteredProducts.find(function(item) {
      return item.id === id;
    }) || null;
    if (!selectedProduct) {
      wx.showToast({ title: "未找到商品", icon: "none" });
      return;
    }
    this.setData({
      selectedId: id,
      selectedProduct
    });
    wx.navigateTo({ url: `/pages/product-detail/index?id=${encodeURIComponent(id)}` });
  },

  openCreateProduct() {
    wx.navigateTo({ url: "/pages/product-form/index?mode=create" });
  },

  openEditProduct() {
    const product = this.data.selectedProduct;
    if (!product) return;
    wx.navigateTo({ url: `/pages/product-form/index?mode=edit&id=${encodeURIComponent(product.id)}` });
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
    wx.reLaunch({ url: "/pages/track/index" });
  },

  copySku() {
    const product = this.data.selectedProduct;
    if (!product) return;
    wx.setClipboardData({
      data: product.sku
    });
  },

  reloadProducts() {
    if (this.data.loading) return;
    this.loadProducts();
  }
});
