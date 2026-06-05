const { getDraft, saveDraft, calculateQuote, calculateCashierSummary, createCashierOrderOnline } = require("../../utils/orderStore");
const { getProfile, canAccessFeature } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");

Page({
  data: {
    sourceChannels: ["到店散客", "企业回访", "熟客复购", "渠道转介绍"],
    sourceChannelIndex: 0,
    form: getDraft(),
    quote: calculateQuote(getDraft()),
    cashierSummary: calculateCashierSummary(getDraft()),
    cashierSubmitting: false
  },

  onShow() {
    this.setTabBarIndex();
    if (!canAccessFeature("cashier")) {
      wx.showToast({ title: "当前账号没有收银权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    this.loadDraft();
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 1 });
    }
  },

  loadDraft() {
    const profile = getProfile();
    const nextForm = Object.assign({}, getDraft());
    if (profile.loggedIn) {
      nextForm.operatorName = nextForm.operatorName || profile.name;
      nextForm.storeName = nextForm.storeName || profile.storeName;
    }

    const sourceChannelIndex = this.data.sourceChannels.indexOf(nextForm.sourceChannel);

    this.setData({
      form: nextForm,
      sourceChannelIndex: Math.max(sourceChannelIndex, 0),
      quote: calculateQuote(nextForm),
      cashierSummary: calculateCashierSummary(nextForm)
    });
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const fieldUpdate = {};
    const value = event.detail.value;
    fieldUpdate[key] = value;
    const nextForm = Object.assign({}, this.data.form, fieldUpdate);
    const update = {
      quote: calculateQuote(nextForm),
      cashierSummary: calculateCashierSummary(nextForm)
    };
    update[`form.${key}`] = value;
    this.setData(update);
  },

  onSourceChannelChange(event) {
    const sourceChannelIndex = Number(event.detail.value);
    const nextForm = Object.assign({}, this.data.form, {
      sourceChannel: this.data.sourceChannels[sourceChannelIndex]
    });
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
    const items = (this.data.form.cashierItems || []).map(function(item) {
      return Object.assign({}, item);
    });
    const value = event.detail.value;
    items[index][key] = value;
    const nextForm = Object.assign({}, this.data.form, {
      cashierItems: items
    });
    const update = {
      cashierSummary: calculateCashierSummary(nextForm)
    };
    update[`form.cashierItems[${index}].${key}`] = value;
    this.setData(update);
  },

  addCashierItem() {
    const items = (this.data.form.cashierItems || []).concat([{
      id: `cashier_item_${Date.now()}`,
      name: "",
      quantity: "1",
      unitPrice: ""
    }]);
    const nextForm = Object.assign({}, this.data.form, {
      cashierItems: items
    });
    this.setData({
      form: nextForm,
      cashierSummary: calculateCashierSummary(nextForm)
    });
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
        name: "",
        quantity: "1",
        unitPrice: ""
      }]
    });
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
  },

  goEntry() {
    if (!this.validateCustomer()) return;
    this.saveCurrentDraft(false);
    wx.switchTab({ url: "/pages/track/index" });
  },

  goConfirm() {
    if (!this.validateCustomer()) return;
    this.saveCurrentDraft(false);
    wx.navigateTo({ url: "/pages/order-detail/index?mode=draft" });
  }
});
