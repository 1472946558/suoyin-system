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
 * 创建日期: 2026-06-06
 */

const {
  getProductByIdOnline,
  saveProductRecord
} = require("../../utils/catalogStore");
const { canAccessFeature } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");

const CATEGORY_OPTIONS = ["金饰", "金条", "K金", "旧金料"];
const STATUS_VALUES = ["active", "draft", "disabled"];
const STATUS_LABELS = ["上架中", "草稿", "停用"];
const STATUS_TEXT = {
  active: "上架中",
  draft: "草稿",
  disabled: "停用"
};

function buildProductForm(product) {
  const source = product || {};
  const category = source.category || "金饰";
  const status = source.status || "active";
  return {
    id: source.id || "",
    name: source.name || "",
    sku: source.sku || `GJG-${Date.now() % 100000}`,
    category,
    categoryTab: source.categoryTab || category,
    imageUrl: source.imageUrl || "",
    purity: source.purity || "足金999",
    benchPrice: String(source.benchPrice || ""),
    retailPrice: String(source.retailPrice || ""),
    gramWeight: String(source.gramWeight || ""),
    status,
    statusText: STATUS_TEXT[status] || "上架中",
    inventory: String(source.inventory || 0),
    tagsText: Array.isArray(source.tags) ? source.tags.join("，") : "",
    recommendedScene: source.recommendedScene || "适合门店快速带入录单。",
    quoteLeadTime: source.quoteLeadTime || "当场可报价"
  };
}

Page({
  data: {
    pageTitle: "新增商品",
    mode: "create",
    categoryOptions: CATEGORY_OPTIONS,
    statusValues: STATUS_VALUES,
    statusLabels: STATUS_LABELS,
    categoryIndex: 0,
    statusIndex: 0,
    saving: false,
    loading: false,
    form: buildProductForm()
  },

  onLoad(options) {
    if (!canAccessFeature("products")) {
      wx.showToast({ title: "当前账号没有商品维护权限", icon: "none" });
      wx.navigateBack();
      return;
    }

    const mode = options && options.mode === "edit" ? "edit" : "create";
    const productId = decodeURIComponent((options && options.id) || "");
    const pageTitle = mode === "edit" ? "编辑商品" : "新增商品";
    wx.setNavigationBarTitle({ title: pageTitle });
    if (mode !== "edit") {
      const form = buildProductForm();
      this.setData({
        mode,
        pageTitle,
        form,
        categoryIndex: Math.max(0, CATEGORY_OPTIONS.indexOf(form.category)),
        statusIndex: Math.max(0, STATUS_VALUES.indexOf(form.status))
      });
      return;
    }

    this.setData({ mode, pageTitle, loading: true });
    getProductByIdOnline(productId).then((product) => {
      if (!product) {
        wx.showToast({ title: "未找到商品", icon: "none" });
        this.goBack();
        return;
      }
      const form = buildProductForm(product);
      this.setData({
        form,
        categoryIndex: Math.max(0, CATEGORY_OPTIONS.indexOf(form.category)),
        statusIndex: Math.max(0, STATUS_VALUES.indexOf(form.status))
      });
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
      this.goBack();
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const update = {};
    update[`form.${key}`] = event.detail.value;
    this.setData(update);
  },

  onCategoryChange(event) {
    const categoryIndex = Number(event.detail.value);
    const category = CATEGORY_OPTIONS[categoryIndex] || CATEGORY_OPTIONS[0];
    this.setData({
      categoryIndex,
      "form.category": category,
      "form.categoryTab": category
    });
  },

  onStatusChange(event) {
    const statusIndex = Number(event.detail.value);
    const status = STATUS_VALUES[statusIndex] || STATUS_VALUES[0];
    this.setData({
      statusIndex,
      "form.status": status,
      "form.statusText": STATUS_TEXT[status] || STATUS_TEXT.active
    });
  },

  uploadProductImage() {
    const setImage = (path) => {
      if (!path) {
        wx.showToast({ title: "未选择图片", icon: "none" });
        return;
      }
      this.setData({
        "form.imageUrl": path
      });
    };

    if (typeof wx.chooseMedia === "function") {
      wx.chooseMedia({
        count: 1,
        mediaType: ["image"],
        sourceType: ["album", "camera"],
        success: (result) => {
          const file = result.tempFiles && result.tempFiles[0];
          setImage(file && file.tempFilePath);
        }
      });
      return;
    }

    wx.chooseImage({
      count: 1,
      sizeType: ["compressed"],
      sourceType: ["album", "camera"],
      success: (result) => {
        setImage(result.tempFilePaths && result.tempFilePaths[0]);
      }
    });
  },

  saveProduct() {
    if (this.data.saving || this.data.loading) return;
    const form = this.data.form;
    if (!form.name || !form.sku) {
      wx.showToast({ title: "请填写商品名称和 SKU", icon: "none" });
      return;
    }

    this.setData({ saving: true });
    saveProductRecord(form).then(() => {
      wx.showToast({ title: "商品已保存", icon: "success" });
      setTimeout(() => {
        this.setData({ saving: false });
        this.goBack();
      }, 300);
    }).catch((error) => {
      this.setData({ saving: false });
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    });
  },

  goBack() {
    const pages = typeof getCurrentPages === "function" ? getCurrentPages() : [];
    if (pages.length > 1) {
      wx.navigateBack();
      return;
    }
    wx.redirectTo({ url: "/pages/products/index" });
  }
});
