const { appConfig } = require("../../utils/config");
const { getDraft, saveDraft, calculateQuote, getSuggestedPrice, getPhotoValidation } = require("../../utils/orderStore");
const { canAccessFeature } = require("../../utils/userStore");

Page({
  data: {
    itemCategories: ["金饰", "金条", "K金", "旧金料"],
    purities: ["足金9999", "足金999", "22K", "18K", "14K"],
    itemCategoryIndex: 0,
    purityIndex: 1,
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
    this.loadDraft();
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 2 });
    }
  },

  loadDraft() {
    const form = getDraft();
    const itemCategoryIndex = this.data.itemCategories.indexOf(form.itemCategory);
    const purityIndex = this.data.purities.indexOf(form.purity);

    if (!form.recyclePrice) {
      form.recyclePrice = String(getSuggestedPrice(form.purity));
    }

    this.setData({
      form,
      itemCategoryIndex: Math.max(itemCategoryIndex, 0),
      purityIndex: Math.max(purityIndex, 0),
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
    this.setData({
      itemCategoryIndex,
      form,
      quote: calculateQuote(form),
      photoValidation: getPhotoValidation(form.photos || [])
    });
  },

  onPurityChange(event) {
    const purityIndex = Number(event.detail.value);
    const purity = this.data.purities[purityIndex];
    const form = Object.assign({}, this.data.form, {
      purity,
      recyclePrice: String(getSuggestedPrice(purity))
    });
    this.setData({
      purityIndex,
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
    this.setData({
      form: form,
      photoValidation: getPhotoValidation(nextPhotos)
    });
  },

  saveOnly() {
    saveDraft(this.data.form);
    wx.showToast({ title: "录单草稿已保存", icon: "none" });
  },

  saveAndConfirm() {
    const { itemName } = this.data.form;
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
  },

  goCashier() {
    saveDraft(this.data.form);
    wx.switchTab({ url: "/pages/order/index" });
  }
});
