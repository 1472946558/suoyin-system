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

const { listOrdersOnline, listCashierOrdersOnline, statusMap } = require("../../utils/orderStore");
const { canAccessFeature } = require("../../utils/userStore");

const combinedStatusMap = Object.assign({}, statusMap, {
  paid: "已完成",
  pending: "待复核",
  refunded: "已取消"
});

Page({
  data: {
    filter: "all",
    bizFilter: "all",
    keyword: "",
    filters: [
      { key: "all", label: "全部" },
      { key: "pending", label: "待处理" },
      { key: "completed", label: "已完成" }
    ],
    bizFilters: [
      { key: "all", label: "全部单据" },
      { key: "recycle", label: "回收单" },
      { key: "cashier", label: "收银单" }
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
    Promise.all([listOrdersOnline(), listCashierOrdersOnline()])
      .then(([recycleOrders, cashierOrders]) => {
        const nextOrders = recycleOrders.concat(cashierOrders).map((order) => {
          const statusText = combinedStatusMap[order.status] || "处理中";
          const isRecycle = order.detailType === "recycle";
          return Object.assign({}, order, {
            statusText: statusText,
            routeSummary: isRecycle
              ? `${order.itemCategory} · ${order.itemName} · ${order.purity}`
              : `${order.itemSummary || order.itemName}`,
            metricText: isRecycle
              ? `净重 ${order.netWeightText}g`
              : `共 ${order.itemCount || 0} 件商品`,
            footerText: isRecycle
              ? `${order.photoCount || 0} 张留痕`
              : "门店收银"
          });
        }).sort(function(left, right) {
          return String(right.createdAt || "").localeCompare(String(left.createdAt || ""));
        });
        this.setData({ orders: nextOrders });
        this.applyFilter(this.data.filter, this.data.bizFilter, nextOrders);
      })
      .catch(() => {
        this.setData({ errorText: "订单列表加载失败，请检查本地数据或接口配置。" });
      })
      .finally(() => {
        this.setData({ loading: false });
      });
  },

  applyFilter(filter, bizFilter, orders) {
    const source = orders || this.data.orders;
    const keyword = (this.data.keyword || "").toLowerCase();
    const bizMatchedOrders = (bizFilter || this.data.bizFilter) === "all"
      ? source
      : source.filter((item) => item.detailType === (bizFilter || this.data.bizFilter));
        const statusMatchedOrders = (filter || this.data.filter) === "all"
      ? bizMatchedOrders
      : bizMatchedOrders.filter(function(item) {
        if ((filter || "") === "pending") {
          return item.status === "draft" || item.status === "pending_confirm" || item.status === "pending";
        }
        return item.status === "completed" || item.status === "confirmed" || item.status === "paid" || item.status === "refunded";
      });
    const filteredOrders = statusMatchedOrders.filter(function(item) {
      if (!keyword) return true;
      return [item.id, item.orderNo, item.customerName, item.customerPhone, item.itemName, item.itemSummary, item.storeName]
        .join(" ")
        .toLowerCase()
        .indexOf(keyword) > -1;
    });
    const totalAmount = filteredOrders.reduce(function(sum, item) {
      return sum + Number(item.amount || 0);
    }, 0);
    const pendingCount = filteredOrders.filter(function(item) {
      return item.status === "draft" || item.status === "pending_confirm" || item.status === "pending";
    }).length;
    this.setData({
      filter: filter || this.data.filter,
      bizFilter: bizFilter || this.data.bizFilter,
      filteredOrders,
      summary: {
        count: filteredOrders.length,
        pendingCount: pendingCount,
        totalAmount: totalAmount.toFixed(2)
      }
    });
  },

  changeFilter(event) {
    this.applyFilter(event.currentTarget.dataset.key, this.data.bizFilter);
  },

  changeBizFilter(event) {
    this.applyFilter(this.data.filter, event.currentTarget.dataset.key);
  },

  onKeywordInput(event) {
    this.setData({
      keyword: event.detail.value
    });
    this.applyFilter(this.data.filter, this.data.bizFilter);
  },

  goDetail(event) {
    wx.navigateTo({
      url: `/pages/order-detail/index?id=${event.currentTarget.dataset.id}&type=${event.currentTarget.dataset.type}`
    });
  },

  goOrder() {
    wx.switchTab({ url: "/pages/order/index" });
  }
});
