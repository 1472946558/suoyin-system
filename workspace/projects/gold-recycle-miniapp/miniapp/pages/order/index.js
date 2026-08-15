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
 * 创建日期: 2026-05-10
 */

const { getDraft, saveDraft, calculateQuote, calculateCashierSummary, createCashierOrderOnline } = require("../../utils/orderStore");
const { listProductsOnline } = require("../../utils/catalogStore");
const { getProfile, canAccessFeature } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");
const { appConfig } = require("../../utils/config");
const { refreshProfileStoreBinding } = require("../../utils/storeSession");

const PRODUCT_STATUS_TEXT = {
  active: "上架中",
  draft: "草稿",
  disabled: "停用"
};

function normalizeText(value) {
  return String(value || "").trim();
}

function normalizeSku(value) {
  return normalizeText(value).toUpperCase();
}

function normalizeStoreOption(store) {
  const source = store || {};
  return {
    id: normalizeText(source.id),
    name: normalizeText(source.name),
    code: normalizeText(source.code)
  };
}

function buildVisibleStoreOptions(profile) {
  const source = profile || {};
  const list = Array.isArray(source.visibleStores) ? source.visibleStores : [];
  const seen = {};
  const options = [];

  list.forEach(function(store) {
    const option = normalizeStoreOption(store);
    const key = option.id || option.name;
    if (!key || seen[key] || !option.name) {
      return;
    }
    seen[key] = true;
    options.push(option);
  });

  const currentStore = normalizeStoreOption({
    id: source.storeId,
    name: source.storeName,
    code: source.storeCode
  });
  const currentKey = currentStore.id || currentStore.name;
  if (currentKey && currentStore.name && !seen[currentKey]) {
    options.unshift(currentStore);
  }

  return options;
}

function matchStoreOption(storeOptions, draft) {
  const options = Array.isArray(storeOptions) ? storeOptions : [];
  const storeId = normalizeText(draft && draft.storeId);
  const storeName = normalizeText(draft && draft.storeName);
  if (storeId) {
    return options.find(function(option) {
      return option.id === storeId;
    }) || null;
  }
  if (storeName) {
    return options.find(function(option) {
      return option.name === storeName;
    }) || null;
  }
  return null;
}

function buildDraftWithStoreContext(form, profile, storeOptions) {
  const nextForm = Object.assign({}, form || {});
  const matchedStore = matchStoreOption(storeOptions, nextForm);
  if (matchedStore) {
    nextForm.storeId = matchedStore.id;
    nextForm.storeName = matchedStore.name;
    nextForm.storeCode = matchedStore.code;
    return nextForm;
  }

  if (!normalizeText(nextForm.storeName) && profile && profile.loggedIn) {
    nextForm.storeId = normalizeText(profile.storeId);
    nextForm.storeName = normalizeText(profile.storeName);
    nextForm.storeCode = normalizeText(profile.storeCode);
  }
  return nextForm;
}

function numberValue(value) {
  const parsed = Number(value || 0);
  return isNaN(parsed) ? 0 : parsed;
}

function cashierUnitPrice(product) {
  const retailPrice = numberValue(product.retailPrice || product.price);
  const benchPrice = numberValue(product.benchPrice);
  const gramWeight = numberValue(product.gramWeight);
  if (retailPrice > 0) {
    return retailPrice;
  }
  if (benchPrice > 0 && gramWeight > 0) {
    return Number((benchPrice * gramWeight).toFixed(2));
  }
  return 0;
}

function normalizeCashierProduct(product) {
  const source = product || {};
  const id = normalizeText(source.productId || source.id);
  const sku = normalizeText(source.sku);
  const storeIds = Array.isArray(source.storeIds) ? source.storeIds : [];
  const stores = Array.isArray(source.stores) ? source.stores : [];
  return Object.assign({}, source, {
    id: id,
    productId: id,
    sku: sku,
    name: normalizeText(source.name),
    status: normalizeText(source.status || "active"),
    category: normalizeText(source.category || source.categoryTab),
    purity: normalizeText(source.purity),
    gramWeight: numberValue(source.gramWeight),
    inventory: numberValue(source.inventory),
    storeIds: storeIds,
    stores: stores,
    unitPrice: cashierUnitPrice(source)
  });
}

function productVisibleForProfile(product, profile) {
  if (!product || !product.sku || !product.name || product.status === "disabled") {
    return false;
  }
  const storeId = normalizeText(profile && profile.storeId);
  const storeName = normalizeText(profile && profile.storeName);
  if (storeId && product.storeIds.length) {
    return product.storeIds.indexOf(storeId) > -1;
  }
  if (storeName === "全部门店") {
    return true;
  }
  return true;
}

function uniqueCashierProducts(products, profile) {
  const seen = {};
  const items = [];
  (products || []).forEach(function(product) {
    const nextProduct = normalizeCashierProduct(product);
    const key = normalizeSku(nextProduct.sku);
    if (!key || seen[key] || !productVisibleForProfile(nextProduct, profile)) {
      return;
    }
    seen[key] = true;
    items.push(nextProduct);
  });
  return items;
}

function productMatchesKeyword(product, keyword) {
  const normalizedKeyword = normalizeSku(keyword);
  if (!normalizedKeyword) {
    return true;
  }
  const candidate = [
    product.sku,
    product.name,
    product.category,
    product.purity
  ].join(" ").toUpperCase();
  return candidate.indexOf(normalizedKeyword) > -1;
}

function mapDropdownProduct(product) {
  return Object.assign({}, product, {
    priceText: product.unitPrice > 0 ? `¥${product.unitPrice.toFixed(2)}` : "未定价"
  });
}

function findProductBySku(products, sku) {
  const target = normalizeSku(sku);
  if (!target) {
    return null;
  }
  return (products || []).find(function(product) {
    return normalizeSku(product.sku) === target;
  }) || null;
}

function applyProductToCashierItem(item, product) {
  const quantity = Math.max(parseInt(item.quantity, 10) || 0, 0);
  const overStock = quantity > product.inventory;
  const statusText = PRODUCT_STATUS_TEXT[product.status] || "未设置";
  const meta = [product.category, product.purity].filter(function(value) {
    return !!value;
  }).join(" · ");
  const statusSuffix = product.status === "active" ? "" : ` · ${statusText}`;
  return Object.assign({}, item, {
    productId: product.productId || product.id || "",
    sku: product.sku,
    name: product.name,
    inventory: String(product.inventory),
    unitPrice: product.unitPrice > 0 ? product.unitPrice.toFixed(2) : String(item.unitPrice || ""),
    productHint: `库存 ${product.inventory} 件${overStock ? "，当前数量超库存" : ""}${meta ? ` · ${meta}` : ""}${statusSuffix}`,
    productHintTone: product.status === "active" && product.inventory > 0 && !overStock ? "ok" : "warn"
  });
}

function refreshCashierInventoryHint(item) {
  const inventoryText = normalizeText(item.inventory);
  if (!inventoryText) {
    return item;
  }
  const inventory = numberValue(inventoryText);
  const quantity = Math.max(parseInt(item.quantity, 10) || 0, 0);
  const overStock = quantity > inventory;
  return Object.assign({}, item, {
    productHint: `库存 ${inventory} 件${overStock ? "，当前数量超库存" : ""}`,
    productHintTone: overStock || inventory <= 0 ? "warn" : "ok"
  });
}

function cloneCashierItems(form) {
  return (form.cashierItems || []).map(function(item) {
    return Object.assign({}, item);
  });
}

Page({
  data: {
    sourceChannels: ["到店散客", "企业回访", "熟客复购", "渠道转介绍"],
    sourceChannelIndex: 0,
    form: getDraft(),
    quote: calculateQuote(getDraft()),
    cashierSummary: calculateCashierSummary(getDraft()),
    cashierSubmitting: false,
    visibleStoreOptions: [],
    filteredStoreOptions: [],
    storeDropdownVisible: false,
    allProductOptions: [],
    productOptions: [],
    cashierProductsLoaded: false,
    activeSkuDropdownIndex: -1,
    filteredProductOptions: []
  },

  onShow() {
    this.setTabBarIndex();
    if (!canAccessFeature("cashier")) {
      wx.showToast({ title: "当前账号没有收银权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    refreshProfileStoreBinding().finally(() => {
      this.loadDraft();
      this.loadCashierProducts();
    });
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 1 });
    }
  },

  loadDraft() {
    const profile = getProfile();
    const storeOptions = buildVisibleStoreOptions(profile);
    const nextForm = buildDraftWithStoreContext(Object.assign({}, getDraft()), profile, storeOptions);
    if (profile.loggedIn) {
      nextForm.operatorName = nextForm.operatorName || profile.name;
    }

    const sourceChannelIndex = this.data.sourceChannels.indexOf(nextForm.sourceChannel);

    this.setData({
      form: nextForm,
      visibleStoreOptions: storeOptions,
      filteredStoreOptions: this.getFilteredStoreOptions(nextForm.storeName, storeOptions),
      sourceChannelIndex: Math.max(sourceChannelIndex, 0),
      quote: calculateQuote(nextForm),
      cashierSummary: calculateCashierSummary(nextForm)
    });
    this.refreshCashierProductsForStore(nextForm);
    saveDraft(nextForm);
  },

  getSelectedStoreContext(formOverride) {
    const form = formOverride || this.data.form || {};
    return {
      storeId: normalizeText(form.storeId),
      storeName: normalizeText(form.storeName)
    };
  },

  loadCashierProducts(done) {
    listProductsOnline().then((result) => {
      const products = result && Array.isArray(result.items) ? result.items : [];
      this.setCashierProducts(products, this.getSelectedStoreContext());
      if (!products.length && appConfig.mode !== "offline") {
        wx.showToast({ title: "暂无可用商品，请先检查后台商品资料", icon: "none" });
      }
    }).catch((error) => {
      this.setCashierProducts([], this.getSelectedStoreContext());
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    }).finally(() => {
      if (typeof done === "function") {
        done();
      }
    });
  },

  setCashierProducts(products, profile) {
    const options = uniqueCashierProducts(products, profile || getProfile());
    this.setData({
      allProductOptions: Array.isArray(products) ? products : [],
      productOptions: options,
      cashierProductsLoaded: true
    });
    this.refreshActiveSkuDropdown();
    this.applyPendingSkuMatches(options);
  },

  refreshCashierProductsForStore(formOverride) {
    if (!this.data.cashierProductsLoaded && !(this.data.allProductOptions || []).length) {
      return;
    }
    this.setCashierProducts(this.data.allProductOptions || [], this.getSelectedStoreContext(formOverride));
  },

  applyPendingSkuMatches(products) {
    const items = cloneCashierItems(this.data.form);
    let changed = false;
    const options = products || this.data.productOptions;
    items.forEach(function(item, index) {
      const product = findProductBySku(options, item.sku);
      if (!product) {
        return;
      }
      if (!item.productId || !item.name || !item.unitPrice) {
        items[index] = applyProductToCashierItem(item, product);
        changed = true;
      }
    });
    if (changed) {
      this.updateCashierItems(items);
    }
  },

  updateCashierItems(items) {
    const nextForm = Object.assign({}, this.data.form, {
      cashierItems: items
    });
    saveDraft(nextForm);
    this.setData({
      form: nextForm,
      cashierSummary: calculateCashierSummary(nextForm)
    });
    this.refreshActiveSkuDropdown(nextForm);
  },

  getFilteredProductOptions(keyword) {
    return (this.data.productOptions || [])
      .filter(function(product) {
        return productMatchesKeyword(product, keyword);
      })
      .slice(0, 8)
      .map(mapDropdownProduct);
  },

  refreshActiveSkuDropdown(formOverride) {
    const activeIndex = Number(this.data.activeSkuDropdownIndex);
    if (activeIndex < 0) {
      return;
    }
    const form = formOverride || this.data.form;
    const items = form && Array.isArray(form.cashierItems) ? form.cashierItems : [];
    const activeItem = items[activeIndex];
    const filteredProductOptions = activeItem
      ? this.getFilteredProductOptions(activeItem.sku)
      : [];
    this.setData({
      filteredProductOptions
    });
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const fieldUpdate = {};
    const value = event.detail.value;
    fieldUpdate[key] = value;
    const nextForm = Object.assign({}, this.data.form, fieldUpdate);
    saveDraft(nextForm);
    const update = {
      quote: calculateQuote(nextForm),
      cashierSummary: calculateCashierSummary(nextForm)
    };
    update[`form.${key}`] = value;
    this.setData(update);
  },

  getFilteredStoreOptions(keyword, storeOptionsOverride) {
    const keywordText = normalizeText(keyword).toUpperCase();
    const storeOptions = Array.isArray(storeOptionsOverride) ? storeOptionsOverride : this.data.visibleStoreOptions;
    return storeOptions.filter(function(option) {
      if (!keywordText) {
        return true;
      }
      const candidate = `${option.name} ${option.code}`.toUpperCase();
      return candidate.indexOf(keywordText) > -1;
    }).slice(0, 20);
  },

  onStoreFocus() {
    this.setData({
      storeDropdownVisible: true,
      filteredStoreOptions: this.getFilteredStoreOptions(this.data.form.storeName)
    });
  },

  onStoreBlur() {
    const that = this;
    setTimeout(function() {
      that.setData({
        storeDropdownVisible: false
      });
    }, 180);
  },

  onStoreInput(event) {
    const value = event.detail.value;
    const matchedStore = matchStoreOption(this.data.visibleStoreOptions, {
      storeName: value
    });
    const nextForm = Object.assign({}, this.data.form, {
      storeId: matchedStore ? matchedStore.id : "",
      storeName: value,
      storeCode: matchedStore ? matchedStore.code : ""
    });
    saveDraft(nextForm);
    this.setData({
      form: nextForm,
      quote: calculateQuote(nextForm),
      cashierSummary: calculateCashierSummary(nextForm),
      storeDropdownVisible: true,
      filteredStoreOptions: this.getFilteredStoreOptions(value)
    });
    this.refreshCashierProductsForStore(nextForm);
  },

  onStoreSelect(event) {
    const storeId = normalizeText(event.currentTarget.dataset.storeId);
    const storeName = normalizeText(event.currentTarget.dataset.storeName);
    const storeCode = normalizeText(event.currentTarget.dataset.storeCode);
    const nextForm = Object.assign({}, this.data.form, {
      storeId,
      storeName,
      storeCode
    });
    saveDraft(nextForm);
    this.setData({
      form: nextForm,
      quote: calculateQuote(nextForm),
      cashierSummary: calculateCashierSummary(nextForm),
      storeDropdownVisible: false,
      filteredStoreOptions: this.getFilteredStoreOptions(storeName)
    });
    this.refreshCashierProductsForStore(nextForm);
  },

  onSourceChannelChange(event) {
    const sourceChannelIndex = Number(event.detail.value);
    const nextForm = Object.assign({}, this.data.form, {
      sourceChannel: this.data.sourceChannels[sourceChannelIndex]
    });
    saveDraft(nextForm);
    this.setData({
      sourceChannelIndex,
      form: nextForm,
      quote: calculateQuote(nextForm),
      cashierSummary: calculateCashierSummary(nextForm)
    });
  },

  onCashierItemInput(event) {
    const index = Number(event.currentTarget.dataset.index);
    const key = event.currentTarget.dataset.key;
    const items = cloneCashierItems(this.data.form);
    const value = event.detail.value;
    if (!items[index]) {
      return;
    }
    items[index][key] = value;
    if (key === "quantity") {
      items[index] = refreshCashierInventoryHint(items[index]);
    }
    this.updateCashierItems(items);
  },

  onCashierSkuInput(event) {
    const index = Number(event.currentTarget.dataset.index);
    const items = cloneCashierItems(this.data.form);
    const value = event.detail.value;
    if (!items[index]) {
      return;
    }
    const nextSku = normalizeSku(value);
    const skuChanged = normalizeSku(items[index].sku) !== nextSku;
    items[index].sku = nextSku;
    if (skuChanged && items[index].productId) {
      items[index].name = "";
      items[index].unitPrice = "";
      items[index].inventory = "";
    }
    items[index].productId = "";
    items[index].productHint = nextSku ? "可直接点下方商品完成带出" : "";
    items[index].productHintTone = "";

    const product = findProductBySku(this.data.productOptions, nextSku);
    if (product) {
      items[index] = applyProductToCashierItem(items[index], product);
    }
    this.setData({
      activeSkuDropdownIndex: index,
      filteredProductOptions: this.getFilteredProductOptions(nextSku)
    });
    this.updateCashierItems(items);
  },

  onCashierSkuFocus(event) {
    const index = Number(event.currentTarget.dataset.index);
    const items = cloneCashierItems(this.data.form);
    const item = items[index];
    this.setData({
      activeSkuDropdownIndex: index,
      filteredProductOptions: this.getFilteredProductOptions(item ? item.sku : "")
    });
  },

  onCashierSkuBlur() {
    const that = this;
    setTimeout(function() {
      that.setData({
        activeSkuDropdownIndex: -1,
        filteredProductOptions: []
      });
    }, 180);
  },

  lookupCashierItemBySku(event) {
    const index = Number(event.currentTarget.dataset.index);
    this.lookupCashierItemBySkuIndex(index, true);
  },

  lookupCashierItemBySkuIndex(index, showToast) {
    const items = cloneCashierItems(this.data.form);
    const item = items[index];
    if (!item) {
      return;
    }
    const sku = normalizeSku(item.sku);
    if (!sku) {
      item.productHint = "请先输入商品编码";
      item.productHintTone = "warn";
      this.updateCashierItems(items);
      if (showToast) {
        wx.showToast({ title: "请先输入商品编码", icon: "none" });
      }
      return;
    }

    if (!this.data.productOptions.length && !this.data.cashierProductsLoaded) {
      const that = this;
      this.loadCashierProducts(function() {
        that.lookupCashierItemBySkuIndex(index, showToast);
      });
      return;
    }

    const product = findProductBySku(this.data.productOptions, sku);
    if (!product) {
      item.sku = sku;
      item.productId = "";
      item.productHint = `未找到商品编码 ${sku}`;
      item.productHintTone = "warn";
      this.updateCashierItems(items);
      if (showToast) {
        wx.showToast({ title: "未找到商品编码", icon: "none" });
      }
      return;
    }

    items[index] = applyProductToCashierItem(item, product);
    this.updateCashierItems(items);
    if (showToast) {
      wx.showToast({ title: "商品信息已带出", icon: "success" });
    }
  },

  onCashierDropdownSelect(event) {
    const index = Number(event.currentTarget.dataset.index);
    const sku = normalizeSku(event.currentTarget.dataset.sku);
    const product = findProductBySku(this.data.productOptions, sku);
    const items = cloneCashierItems(this.data.form);
    if (!product || !items[index]) {
      return;
    }
    items[index] = applyProductToCashierItem(items[index], product);
    this.setData({
      activeSkuDropdownIndex: -1,
      filteredProductOptions: []
    });
    this.updateCashierItems(items);
  },

  removeCashierItem(event) {
    const index = Number(event.currentTarget.dataset.index);
    const currentItems = this.data.form.cashierItems || [];
    const items = currentItems.filter(function(item, itemIndex) {
      return itemIndex !== index;
    });
    const nextForm = Object.assign({}, this.data.form, {
      cashierItems: items.length ? items : [{
        id: `cashier_item_${Date.now()}`,
        productId: "",
        sku: "",
        name: "",
        quantity: "1",
        unitPrice: "",
        inventory: "",
        productHint: "",
        productHintTone: ""
      }]
    });
    saveDraft(nextForm);
    this.setData({
      form: nextForm,
      cashierSummary: calculateCashierSummary(nextForm)
    });
  },

  saveCurrentDraft(showToast) {
    saveDraft(this.data.form);
    if (showToast) {
      wx.showToast({ title: "草稿已保存", icon: "none" });
    }
  },

  validateCustomer() {
    const { customerName, customerPhone } = this.data.form;
    if (!customerName || !customerPhone) {
      wx.showToast({ title: "请先填写客户姓名和手机号", icon: "none" });
      return false;
    }
    return true;
  },

  saveDraftOnly() {
    this.saveCurrentDraft(true);
  },

  submitCashierOrder() {
    if (!this.validateCustomer()) return;
    if (!this.data.form.storeId) {
      wx.showToast({ title: "请从下拉列表选择门店", icon: "none" });
      return;
    }
    this.setData({ cashierSubmitting: true });
    createCashierOrderOnline(this.data.form)
      .then((order) => {
        wx.showToast({ title: "收银单已生成", icon: "success" });
        setTimeout(() => {
          wx.navigateTo({ url: `/pages/order-detail/index?id=${order.id}&type=cashier` });
        }, 350);
      })
      .catch((error) => {
        wx.showToast({ title: formatRequestError(error), icon: "none" });
      })
      .finally(() => {
        this.setData({ cashierSubmitting: false });
      });
  }
});
