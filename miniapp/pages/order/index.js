const { getDraft, saveDraft, calculateQuote } = require("../../utils/orderStore");
const { getProfile, canAccessFeature } = require("../../utils/userStore");
const { appConfig } = require("../../utils/config");

Page({
  data: {
    paymentMethods: ["微信零钱", "银行卡", "线下转账", "暂不支付"],
    sourceChannels: ["到店散客", "企业回访", "熟客复购", "渠道转介绍"],
    paymentMethodIndex: 0,
    sourceChannelIndex: 0,
    form: getDraft(),
    quote: calculateQuote(getDraft()),
    demoNotice: appConfig.mode === "mock" ? "当前是本地 mock 收银草稿，确认前不会触发真实支付。" : "当前已连接真实服务，请确认收银与客户信息。"
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

    const paymentMethodIndex = this.data.paymentMethods.indexOf(nextForm.paymentMethod);
    const sourceChannelIndex = this.data.sourceChannels.indexOf(nextForm.sourceChannel);

    this.setData({
      form: nextForm,
      paymentMethodIndex: Math.max(paymentMethodIndex, 0),
      sourceChannelIndex: Math.max(sourceChannelIndex, 0),
      quote: calculateQuote(nextForm)
    });
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const fieldUpdate = {};
    fieldUpdate[key] = event.detail.value;
    const nextForm = Object.assign({}, this.data.form, fieldUpdate);
    this.setData({
      form: nextForm,
      quote: calculateQuote(nextForm)
    });
  },

  onPaymentMethodChange(event) {
    const paymentMethodIndex = Number(event.detail.value);
    const nextForm = Object.assign({}, this.data.form, {
      paymentMethod: this.data.paymentMethods[paymentMethodIndex]
    });
    this.setData({
      paymentMethodIndex,
      form: nextForm,
      quote: calculateQuote(nextForm)
    });
  },

  onSourceChannelChange(event) {
    const sourceChannelIndex = Number(event.detail.value);
    const nextForm = Object.assign({}, this.data.form, {
      sourceChannel: this.data.sourceChannels[sourceChannelIndex]
    });
    this.setData({
      sourceChannelIndex,
      form: nextForm,
      quote: calculateQuote(nextForm)
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
