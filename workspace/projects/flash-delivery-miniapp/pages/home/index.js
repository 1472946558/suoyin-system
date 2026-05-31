const { estimatePrice } = require("../../utils/orderStore");

Page({
  data: {
    fromAddress: "南山区 科技园",
    toAddress: "福田区 购物公园",
    estimate: {
      price: 32,
      distance: "8.6km",
      time: 36
    },
    metrics: [
      { value: "30min", label: "平均响应" },
      { value: "1v1", label: "专人专送" },
      { value: "24h", label: "状态追踪" }
    ],
    scenes: [
      { icon: "/assets/ui/document.png", name: "文件证件", desc: "合同票据" },
      { icon: "/assets/ui/flower.png", name: "鲜花蛋糕", desc: "轻拿轻放" },
      { icon: "/assets/ui/device.png", name: "数码配件", desc: "专人保护" },
      { icon: "/assets/ui/cold.png", name: "生鲜食品", desc: "即时送达" }
    ],
    promises: [
      { icon: "/assets/ui/rider.png", title: "专人直送", desc: "一单一送，不拼单绕路", code: "01" },
      { icon: "/assets/ui/route.png", title: "过程可见", desc: "接单、取件、配送、送达全链路跟踪", code: "02" },
      { icon: "/assets/ui/price.png", title: "价格透明", desc: "下单前预估费用，异常费用明确提示", code: "03" },
      { icon: "/assets/ui/api.png", title: "商用可扩展", desc: "预留定位、支付、后端订单接口", code: "04" }
    ],
    flow: [
      { icon: "/assets/ui/pickup.png", name: "填写地址" },
      { icon: "/assets/ui/price.png", name: "智能估价" },
      { icon: "/assets/ui/rider.png", name: "骑手接单" },
      { icon: "/assets/ui/route.png", name: "全程追踪" }
    ]
  },

  onShow() {
    this.setTabBarIndex();
    this.refreshEstimate();
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 0 });
    }
  },

  onFromInput(event) {
    this.setData({ fromAddress: event.detail.value }, () => this.refreshEstimate());
  },

  onToInput(event) {
    this.setData({ toAddress: event.detail.value }, () => this.refreshEstimate());
  },

  refreshEstimate() {
    const result = estimatePrice({
      fromAddress: this.data.fromAddress,
      toAddress: this.data.toAddress,
      itemType: "文件证件",
      weight: "1kg以内"
    });
    this.setData({
      estimate: {
        price: result.price,
        distance: result.distance,
        time: Math.max(22, Math.round(parseFloat(result.distance) * 4))
      }
    });
  },

  chooseScene(event) {
    wx.setStorageSync("jf_pending_order_form", {
      itemType: event.currentTarget.dataset.name
    });
    wx.switchTab({ url: "/pages/order/index" });
  },

  goOrder() {
    wx.switchTab({ url: "/pages/order/index" });
  },

  goOrderWithAddress() {
    wx.setStorageSync("jf_pending_order_form", {
      fromAddress: this.data.fromAddress,
      toAddress: this.data.toAddress
    });
    wx.switchTab({ url: "/pages/order/index" });
  },

  callService() {
    wx.showToast({ title: "演示客服入口", icon: "none" });
  }
});
