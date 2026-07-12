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

const { appConfig } = require("../../utils/config");
const { getDraft, saveDraft, calculateQuote, getSuggestedPrice, getPhotoValidation, refreshReferencePrices } = require("../../utils/orderStore");
const { canAccessFeature } = require("../../utils/userStore");
const { refreshProfileStoreBinding } = require("../../utils/storeSession");

Page({
  data: {
    itemCategories: ["足金", "K金", "铂金", "钯金", "银饰", "牙金"],
    itemCategoryIndex: 0,
    inlineCustomerFeatureEnabled: !!appConfig.featureFlags.recycleInlineCustomerEnabled,
    form: getDraft(),
    quote: calculateQuote(getDraft()),
    photoRule: appConfig.recyclePhotoRules,
    photoValidation: getPhotoValidation(getDraft().photos || [])
  },

  onShow() {
    this.setTabBarIndex();
    if (!canAccessFeature("recycle")) {
      wx.showToast({ title: "当前账号没有回收录单权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    refreshProfileStoreBinding().finally(() => {
      this.loadDraft();
      refreshReferencePrices().then(() => {
        this.loadDraft();
      });
    });
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 2 });
    }
  },

  loadDraft() {
    const form = getDraft();
    const itemCategoryIndex = this.data.itemCategories.indexOf(form.itemCategory);

    if (!form.recyclePrice) {
      form.recyclePrice = String(getSuggestedPrice(form.purity));
    }

    this.setData({
      form,
      inlineCustomerFeatureEnabled: !!appConfig.featureFlags.recycleInlineCustomerEnabled,
      itemCategoryIndex: Math.max(itemCategoryIndex, 0),
      quote: calculateQuote(form),
      photoValidation: getPhotoValidation(form.photos || [])
    });
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const fieldUpdate = {};
    const value = event.detail.value;
    fieldUpdate[key] = value;
    const form = Object.assign({}, this.data.form, fieldUpdate);
    saveDraft(form);
    const update = {
      quote: calculateQuote(form),
      photoValidation: getPhotoValidation(form.photos || [])
    };
    update[`form.${key}`] = value;
    this.setData(update);
  },

  onCategoryChange(event) {
    const itemCategoryIndex = Number(event.detail.value);
    const form = Object.assign({}, this.data.form, {
      itemCategory: this.data.itemCategories[itemCategoryIndex]
    });
    saveDraft(form);
    this.setData({
      itemCategoryIndex,
      form,
      quote: calculateQuote(form),
      photoValidation: getPhotoValidation(form.photos || [])
    });
  },

  addPhotos() {
    const remain = this.data.photoRule.maxCount - (this.data.form.photos || []).length;
    if (remain <= 0) {
      wx.showToast({ title: `最多上传 ${this.data.photoRule.maxCount} 张`, icon: "none" });
      return;
    }
    wx.chooseImage({
      count: remain,
      sizeType: ["compressed"],
      sourceType: ["camera", "album"],
      success: (result) => {
        const current = this.data.form.photos || [];
        const picked = (result.tempFilePaths || []).map(function(path, index) {
          return {
            id: `photo_${Date.now()}_${index + 1}`,
            path: path,
            name: `现场图 ${current.length + index + 1}`
          };
        });
        const nextPhotos = current.concat(picked);
        const form = Object.assign({}, this.data.form, {
          photos: nextPhotos
        });
        saveDraft(form);
        this.setData({
          form: form,
          photoValidation: getPhotoValidation(nextPhotos)
        });
      }
    });
  },

  removePhoto(event) {
    const index = Number(event.currentTarget.dataset.index);
    const nextPhotos = (this.data.form.photos || []).filter(function(item, photoIndex) {
      return photoIndex !== index;
    });
    const form = Object.assign({}, this.data.form, {
      photos: nextPhotos
    });
    saveDraft(form);
    this.setData({
      form: form,
      photoValidation: getPhotoValidation(nextPhotos)
    });
  },

  previewPhoto(event) {
    const url = event.currentTarget.dataset.url;
    const urls = (this.data.form.photos || []).map(function(item) {
      return item.path;
    }).filter(function(item) {
      return !!item;
    });
    if (!url || !urls.length) {
      return;
    }
    wx.previewImage({
      current: url,
      urls: urls
    });
  },

  saveOnly() {
    saveDraft(this.data.form);
    wx.showToast({ title: "录单草稿已保存", icon: "none" });
  },

  saveAndConfirm() {
    const { customerName, customerPhone, itemName } = this.data.form;
    if (!customerName || !customerPhone) {
      wx.showToast({ title: "请先填写客户姓名和手机号", icon: "none" });
      return;
    }
    if (!itemName || !this.data.quote.netWeight) {
      wx.showToast({ title: "请填写品名并录入有效克重", icon: "none" });
      return;
    }

    const photoValidation = getPhotoValidation(this.data.form.photos || []);
    if (!photoValidation.ok) {
      wx.showToast({ title: photoValidation.message, icon: "none" });
      this.setData({ photoValidation: photoValidation });
      return;
    }

    saveDraft(this.data.form);
    wx.navigateTo({ url: "/pages/order-detail/index?mode=draft" });
  }
});
