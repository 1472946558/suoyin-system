const {
  seedCatalog,
  getProducts,
  listProductsOnline,
  saveProductRecord,
  getProductStats
} = require("../../utils/catalogStore");
const { mergeDraft } = require("../../utils/orderStore");
const { canAccessFeature, getProfile } = require("../../utils/userStore");

const CATEGORY_OPTIONS = ["all", "金饰", "金条", "K金", "旧金料"];
const PRODUCT_CATEGORY_OPTIONS = ["金饰", "金条", "K金", "旧金料"];
const STATUS_FILTER_VALUES = ["all", "active", "draft", "disabled"];
const STATUS_FILTER_LABELS = ["全部状态", "上架中", "草稿", "停用"];
const STATUS_VALUES = ["active", "draft", "disabled"];
const STATUS_LABELS = ["上架中", "草稿", "停用"];
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
    : (nextProduct.stockStatus === "low" ? "低库存" : `${Number(nextProduct.inventory || 0)} 件`);
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

function buildProductEditor(product) {
  const source = product || {};
  const category = source.category || "金饰";
  const status = source.status || "active";
  return {
    id: source.id || "",
    name: source.name || "",
    sku: source.sku || `GJG-${Date.now() % 100000}`,
    category,
    categoryTab: source.categoryTab || category,
    imageUrl: source.imageUrl || "/assets/ui/price.png",
    purity: source.purity || "足金999",
    benchPrice: String(source.benchPrice || ""),
    retailPrice: String(source.retailPrice || ""),
    gramWeight: String(source.gramWeight || ""),
    status,
    inventory: String(source.inventory || 0),
    tagsText: Array.isArray(source.tags) ? source.tags.join("，") : "",
    recommendedScene: source.recommendedScene || "适合门店快速带入录单。",
    quoteLeadTime: source.quoteLeadTime || "当场可报价"
  };
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
    productCategoryOptions: PRODUCT_CATEGORY_OPTIONS,
    statusFilterValues: STATUS_FILTER_VALUES,
    statusFilterLabels: STATUS_FILTER_LABELS,
    statusValues: STATUS_VALUES,
    statusLabels: STATUS_LABELS,
    statusFilterIndex: 0,
    loading: false,
    productEditorVisible: false,
    productEditorTitle: "新增商品",
    productEditor: buildProductEditor(),
    productEditorCategoryIndex: 0,
    productEditorStatusIndex: 0
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
    listProductsOnline().then((result) => {
      let products = result && Array.isArray(result.items) ? result.items : [];
      if (!products.length) {
        seedCatalog();
        products = getProducts();
      }
      products = products.map(normalizeProductView);
      this.setData({
        products,
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
    this.setData({
      selectedId: id,
      selectedProduct
    });
  },

  openCreateProduct() {
    this.setData({
      productEditorVisible: true,
      productEditorTitle: "新增商品",
      productEditor: buildProductEditor(),
      productEditorCategoryIndex: 0,
      productEditorStatusIndex: 0
    });
  },

  openEditProduct() {
    const product = this.data.selectedProduct;
    if (!product) return;
    this.setData({
      productEditorVisible: true,
      productEditorTitle: "编辑商品",
      productEditor: buildProductEditor(product),
      productEditorCategoryIndex: Math.max(0, PRODUCT_CATEGORY_OPTIONS.indexOf(product.category)),
      productEditorStatusIndex: Math.max(0, STATUS_VALUES.indexOf(product.status))
    });
  },

  closeProductEditor() {
    this.setData({ productEditorVisible: false });
  },

  noop() {},

  onProductEditorInput(event) {
    const key = event.currentTarget.dataset.key;
    const update = {};
    update[`productEditor.${key}`] = event.detail.value;
    this.setData(update);
  },

  onProductEditorCategory(event) {
    const index = Number(event.detail.value);
    const category = PRODUCT_CATEGORY_OPTIONS[index] || PRODUCT_CATEGORY_OPTIONS[0];
    this.setData({
      productEditorCategoryIndex: index,
      "productEditor.category": category,
      "productEditor.categoryTab": category
    });
  },

  onProductEditorStatus(event) {
    const index = Number(event.detail.value);
    this.setData({
      productEditorStatusIndex: index,
      "productEditor.status": STATUS_VALUES[index] || "active"
    });
  },

  saveProductEditor() {
    const editor = this.data.productEditor;
    if (!editor.name || !editor.sku) {
      wx.showToast({ title: "请填写商品名称和 SKU", icon: "none" });
      return;
    }
    const savedProduct = normalizeProductView(saveProductRecord(editor));
    const products = [savedProduct].concat(this.data.products.filter(function(item) {
      return item.id !== savedProduct.id;
    }));
    this.setData({
      products,
      stats: getProductStats(products),
      selectedId: savedProduct.id,
      productEditorVisible: false
    });
    this.applyFilters();
    wx.showToast({ title: "商品已保存", icon: "success" });
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

  reloadProducts() {
    if (this.data.loading) return;
    this.loadProducts();
  }
});
