const { listOrdersOnline, statusMap } = require("../../utils/orderStore");
const { canAccessFeature } = require("../../utils/userStore");

Page({
  data: {
    filter: "all",
    keyword: "",
    filters: [
      { key: "all", label: "全部" },
      { key: "pending_confirm", label: "待确认" },
      { key: "pending_payment", label: "待打款" },
      { key: "completed", label: "已完成" }
    ],
    orders: [],
    filteredOrders: [],
    summary: {
      count: 0,
      pendingCount: 0,
      totalAmount: "0.00"
    },
    loading: false,
    errorText: ""
  },

  onShow() {
    this.setTabBarIndex();
    if (!canAccessFeature("orders")) {
      wx.showToast({ title: "当前账号没有订单查看权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    this.loadOrders();
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 3 });
    }
  },

  loadOrders() {
    this.setData({ loading: true, errorText: "" });
    listOrdersOnline()
      .then((orders) => {
        const nextOrders = orders.map((order) => Object.assign({}, order, {
          statusText: statusMap[order.status] || "处理中"
        }));
        this.setData({ orders: nextOrders });
        this.applyFilter(this.data.filter, nextOrders);
      })
      .catch(() => {
        this.setData({ errorText: "订单列表加载失败，请检查本地数据或接口配置。" });
      })
      .finally(() => {
        this.setData({ loading: false });
      });
  },

  applyFilter(filter, orders) {
    const source = orders || this.data.orders;
    const keyword = (this.data.keyword || "").toLowerCase();
    const statusMatchedOrders = filter === "all" ? source : source.filter((item) => item.status === filter);
    const filteredOrders = statusMatchedOrders.filter(function(item) {
      if (!keyword) return true;
      return [item.id, item.customerName, item.customerPhone, item.itemName, item.storeName]
        .join(" ")
        .toLowerCase()
        .indexOf(keyword) > -1;
    });
    const totalAmount = filteredOrders.reduce(function(sum, item) {
      return sum + Number(item.amount || 0);
    }, 0);
    const pendingCount = filteredOrders.filter(function(item) {
      return item.status === "pending_confirm" || item.status === "pending_payment";
    }).length;
    this.setData({
      filter,
      filteredOrders,
      summary: {
        count: filteredOrders.length,
        pendingCount: pendingCount,
        totalAmount: totalAmount.toFixed(2)
      }
    });
  },

  changeFilter(event) {
    this.applyFilter(event.currentTarget.dataset.key);
  },

  onKeywordInput(event) {
    this.setData({
      keyword: event.detail.value
    });
    this.applyFilter(this.data.filter);
  },

  goDetail(event) {
    wx.navigateTo({
      url: `/pages/order-detail/index?id=${event.currentTarget.dataset.id}`
    });
  },

  goOrder() {
    wx.switchTab({ url: "/pages/order/index" });
  }
});
