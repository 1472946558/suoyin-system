const { getLatestOrderOnline, getOrderTrackOnline, statusMap, statusSteps } = require("../../utils/orderStore");

Page({
  data: {
    order: null,
    statusText: "",
    steps: [],
    loading: false,
    errorText: ""
  },

  onShow() {
    this.setTabBarIndex();
    this.loadOrder();
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 2 });
    }
  },

  loadOrder() {
    this.setData({ loading: true, errorText: "" });
    getLatestOrderOnline()
      .then((order) => {
        if (!order) {
          this.setData({ order: null, steps: [] });
          return;
        }
        return getOrderTrackOnline(order).then((trackOrder) => {
          this.applyOrder(trackOrder || order);
        });
      })
      .catch(() => {
        this.setData({ errorText: "配送跟踪加载失败，请检查服务器配置或网络。" });
      })
      .finally(() => {
        this.setData({ loading: false });
      });
  },

  applyOrder(order) {
    const currentIndex = statusSteps.indexOf(order.status);
    const titles = {
      created: "订单已创建",
      accepted: "骑手已接单",
      picked: "骑手已取件",
      delivering: "正在配送",
      completed: "已送达"
    };
    const descs = {
      created: "系统正在为你匹配附近骑手",
      accepted: "骑手将尽快到达取件地址",
      picked: "物品已完成取件",
      delivering: "骑手正在前往收件地址",
      completed: "订单已完成，感谢使用"
    };
    this.setData({
      order,
      statusText: statusMap[order.status],
      steps: statusSteps.map((key, index) => ({
        key,
        title: titles[key],
        desc: descs[key],
        done: index <= currentIndex
      }))
    });
  },

  callRider() {
    wx.showToast({ title: "演示联系骑手", icon: "none" });
  },

  goOrder() {
    wx.switchTab({ url: "/pages/order/index" });
  }
});
