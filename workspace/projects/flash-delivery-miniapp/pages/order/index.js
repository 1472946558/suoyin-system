const { createOrderOnline, estimatePrice } = require("../../utils/orderStore");
const { getProfile } = require("../../utils/userStore");
const { appConfig } = require("../../utils/config");

Page({
  data: {
    itemTypes: ["文件证件", "数码配件", "鲜花蛋糕", "生鲜食品", "其它物品"],
    weights: ["1kg以内", "1-3kg", "3-5kg", "5kg以上"],
    itemTypeIndex: 0,
    weightIndex: 0,
    form: {
      fromAddress: "",
      toAddress: "",
      itemType: "文件证件",
      weight: "1kg以内",
      note: ""
    },
    estimate: {
      price: 18,
      distance: "4.8km",
      time: 22
    },
    submitting: false,
    submitButtonText: "提交订单",
    demoNotice: appConfig.mode === "mock" ? "当前为本地演示环境：提交后订单会保存到本机，用于完整走通客户验收流程。" : "当前已连接客户服务器：订单会提交到正式后端，请确认信息无误。"
  },

  onLoad(query) {
    const nextForm = Object.assign({}, this.data.form);
    if (query.from) nextForm.fromAddress = decodeURIComponent(query.from);
    if (query.to) nextForm.toAddress = decodeURIComponent(query.to);
    if (query.itemType && this.data.itemTypes.includes(decodeURIComponent(query.itemType))) {
      nextForm.itemType = decodeURIComponent(query.itemType);
    }
    const itemTypeIndex = this.data.itemTypes.indexOf(nextForm.itemType);
    this.setData({ form: nextForm, itemTypeIndex: Math.max(itemTypeIndex, 0) }, () => this.refreshEstimate());
  },

  onShow() {
    this.setTabBarIndex();
    const pending = wx.getStorageSync("jf_pending_order_form");
    if (!pending) {
      this.applyProfileDefaults();
      return;
    }
    wx.removeStorageSync("jf_pending_order_form");
    const nextForm = Object.assign({}, this.data.form, pending);
    const itemTypeIndex = this.data.itemTypes.indexOf(nextForm.itemType);
    const weightIndex = this.data.weights.indexOf(nextForm.weight);
    this.setData({
      form: nextForm,
      itemTypeIndex: Math.max(itemTypeIndex, 0),
      weightIndex: Math.max(weightIndex, 0)
    }, () => this.refreshEstimate());
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 1 });
    }
  },

  applyProfileDefaults() {
    const profile = getProfile();
    if (!profile.loggedIn) return;
    const nextForm = Object.assign({}, this.data.form);
    if (!nextForm.fromAddress && profile.defaultFromAddress) {
      nextForm.fromAddress = profile.defaultFromAddress;
    }
    if (!nextForm.toAddress && profile.defaultToAddress) {
      nextForm.toAddress = profile.defaultToAddress;
    }
    this.setData({ form: nextForm }, () => this.refreshEstimate());
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const update = {};
    update[`form.${key}`] = event.detail.value;
    this.setData(update, () => this.refreshEstimate());
  },

  onItemTypeChange(event) {
    const index = Number(event.detail.value);
    this.setData({
      itemTypeIndex: index,
      "form.itemType": this.data.itemTypes[index]
    }, () => this.refreshEstimate());
  },

  onWeightChange(event) {
    const index = Number(event.detail.value);
    this.setData({
      weightIndex: index,
      "form.weight": this.data.weights[index]
    }, () => this.refreshEstimate());
  },

  refreshEstimate() {
    const result = estimatePrice(this.data.form);
    this.setData({
      estimate: {
        price: result.price,
        distance: result.distance,
        time: Math.max(22, Math.round(parseFloat(result.distance) * 4))
      }
    });
  },

  submitOrder() {
    const { fromAddress, toAddress } = this.data.form;
    if (!fromAddress || !toAddress) {
      wx.showToast({ title: "请填写取件和收件地址", icon: "none" });
      return;
    }
    this.setData({ submitting: true, submitButtonText: "正在提交" });
    createOrderOnline(this.data.form)
      .then((order) => {
        wx.showToast({ title: "下单成功", icon: "success" });
        setTimeout(() => {
          wx.navigateTo({ url: `/pages/order-detail/index?id=${order.id}` });
        }, 450);
      })
      .catch(() => {
        wx.showToast({ title: "下单失败，请检查服务器配置", icon: "none" });
      })
      .finally(() => {
        this.setData({ submitting: false, submitButtonText: "提交订单" });
      });
  }
});
