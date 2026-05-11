const { appConfig } = require("../../utils/config");
const { getDashboardStats, getDraft } = require("../../utils/orderStore");
const { getProfile, canAccessFeature } = require("../../utils/userStore");
const { getMemberStats, getProductStats } = require("../../utils/catalogStore");

const allQuickActions = [
  { title: "开始收银", desc: "录入客户与打款方式", type: "tab", url: "/pages/order/index", icon: "/assets/ui/price.png", permission: "cashier" },
  { title: "回收录单", desc: "录入成色重量与报价", type: "tab", url: "/pages/track/index", icon: "/assets/ui/document.png", permission: "recycle" },
  { title: "查看订单", desc: "浏览历史回收单", type: "tab", url: "/pages/orders/index", icon: "/assets/ui/list.png", permission: "orders" },
  { title: "会员管理", desc: "常客档案与带入建单", type: "nav", url: "/pages/members/index", icon: "/assets/ui/user.png", permission: "members" },
  { title: "商品目录", desc: "门店商品与报价参考", type: "nav", url: "/pages/products/index", icon: "/assets/ui/enterprise.png", permission: "products" },
  { title: "系统设置", desc: "门店与操作员设置", type: "tab", url: "/pages/profile/index", icon: "/assets/ui/support.png", permission: "settings" }
];

Page({
  data: {
    profile: getProfile(),
    prices: appConfig.mockBenchPrices,
    stats: {
      todayCount: 0,
      todayAmount: "0.00",
      pendingCount: 0,
      completedCount: 0
    },
    resourceStats: {
      members: 0,
      vipMembers: 0,
      products: 0,
      activeProducts: 0
    },
    quickActions: [],
    checklist: [
      { title: "登录门店账号", desc: "保存操作员、门店与班次信息" },
      { title: "创建收银草稿", desc: "先登记客户与付款方式" },
      { title: "录入回收细项", desc: "记录品类、成色、克重和扣减" },
      { title: "确认生成回收单", desc: "在回收确认页一键落本地 mock 订单" }
    ],
    draftReady: false
  },

  onShow() {
    this.setTabBarIndex();
    this.setData({
      profile: getProfile(),
      stats: getDashboardStats(),
      resourceStats: {
        members: getMemberStats().total,
        vipMembers: getMemberStats().vipCount,
        products: getProductStats().total,
        activeProducts: getProductStats().activeCount
      },
      draftReady: Boolean(getDraft().customerName || getDraft().itemName),
      quickActions: allQuickActions.filter(function(item) {
        return canAccessFeature(item.permission);
      })
    });
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 0 });
    }
  },

  goAction(event) {
    const { type, url } = event.currentTarget.dataset;
    if (type === "tab") {
      wx.switchTab({ url });
      return;
    }
    wx.navigateTo({ url });
  },

  goLogin() {
    wx.navigateTo({ url: "/pages/account/index" });
  },

  resumeDraft() {
    wx.switchTab({ url: "/pages/order/index" });
  }
});
