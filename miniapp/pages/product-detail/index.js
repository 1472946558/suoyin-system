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
  getProductByIdOnline
} = require("../../utils/catalogStore");
const { mergeDraft } = require("../../utils/orderStore");
const { canAccessFeature } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");

const STATUS_TEXT = {
  active: "上架中",
  draft: "草稿",
  disabled: "停用"
};

function formatCurrency(amount) {
  return `¥${Number(amount || 0).toFixed(2)}`;
}

function normalizeProduct(product) {
  const nextProduct = Object.assign({}, product || {});
  nextProduct.tags = Array.isArray(nextProduct.tags) ? nextProduct.tags : [];
  nextProduct.stores = Array.isArray(nextProduct.stores) ? nextProduct.stores : [];
  nextProduct.statusText = STATUS_TEXT[nextProduct.status] || "未设置";
  nextProduct.benchPriceText = `${formatCurrency(nextProduct.benchPrice)}/g`;
  nextProduct.retailPriceText = Number(nextProduct.retailPrice || 0) > 0
    ? `${formatCurrency(nextProduct.retailPrice)}/g`
    : "未设置";
  nextProduct.gramWeightText = Number(nextProduct.gramWeight || 0) > 0
    ? `${nextProduct.gramWeight}g`
    : "按实秤录入";
  nextProduct.storesText = nextProduct.stores.length ? nextProduct.stores.join(" / ") : "未分配";
  nextProduct.recommendedScene = nextProduct.recommendedScene || "暂无使用说明";
  nextProduct.quoteLeadTime = nextProduct.quoteLeadTime || "当场可报价";
  return nextProduct;
}

Page({
  data: {
    productId: "",
    product: normalizeProduct(),
    loading: false
  },

  onLoad(options) {
    if (!canAccessFeature("products")) {
      wx.showToast({ title: "当前账号没有商品查看权限", icon: "none" });
      wx.navigateBack();
      return;
    }
    this.setData({
      productId: decodeURIComponent((options && options.id) || "")
    });
    this.loadProduct();
  },

  onShow() {
    if (this.data.productId) {
      this.loadProduct();
    }
  },

  loadProduct() {
    this.setData({ loading: true });
    getProductByIdOnline(this.data.productId).then((product) => {
      if (!product) {
        wx.showToast({ title: "未找到商品", icon: "none" });
        setTimeout(() => {
          wx.navigateBack();
        }, 300);
        return;
      }
      this.setData({
        product: normalizeProduct(product)
      });
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
      setTimeout(() => {
        wx.navigateBack();
      }, 300);
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  copySku() {
    if (!this.data.product.sku) return;
    wx.setClipboardData({
      data: this.data.product.sku
    });
  },

  editProduct() {
    wx.navigateTo({ url: `/pages/product-form/index?mode=edit&id=${encodeURIComponent(this.data.productId)}` });
  },

  useForRecycle() {
    const product = this.data.product;
    mergeDraft({
      itemCategory: product.category,
      itemName: product.name,
      purity: product.purity,
      recyclePrice: String(product.benchPrice),
      remark: product.recommendedScene
    });
    wx.showToast({ title: "商品参数已带入录单草稿", icon: "none" });
    wx.switchTab({ url: "/pages/track/index" });
  }
});
