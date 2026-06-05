const { appConfig } = require("../../utils/config");
const { getDashboardStats, getDraft } = require("../../utils/orderStore");
const { getProfile, canAccessFeature } = require("../../utils/userStore");
const { getMemberStats, getProductStats } = require("../../utils/catalogStore");

const allQuickActions = [
  { title: "开始收银", desc: "录入客户与商品明细", type: "tab", url: "/pages/order/index", icon: "/assets/ui/home-cashier.png", permission: "cashier" },
  { title: "回收录单", desc: "录入成色重量与报价", type: "tab", url: "/pages/track/index", icon: "/assets/ui/home-recycle.png", permission: "recycle" },
  { title: "查看订单", desc: "浏览历史回收单", type: "tab", url: "/pages/orders/index", icon: "/assets/ui/home-orders.png", permission: "orders" },
  { title: "会员管理", desc: "常客档案与带入建单", type: "nav", url: "/pages/members/index", icon: "/assets/ui/home-members.png", permission: "members" },
  { title: "商品目录", desc: "门店商品与报价参考", type: "nav", url: "/pages/products/index", icon: "/assets/ui/home-products.png", permission: "products" },
  { title: "系统设置", desc: "门店与操作员设置", type: "tab", url: "/pages/profile/index", icon: "/assets/ui/home-settings.png", permission: "settings" }
];

Page({
  data: {
    profile: getProfile(),
    prices: appConfig.referencePrices,
    stats: {
      todayCount: 0,
      todayAmount: "0.00",
      todayCashierAmount: "0.00",
      todayRecycleAmount: "0.00",
      pendingCount: 0,
      completedCount: 0,
      scopeText: "本店",
      visibleStoreCount: 0,
      storeBreakdown: []
    },
    resourceStats: {
      members: 0,
      vipMembers: 0,
      products: 0,
      activeProducts: 0
    },
    quickActions: [],
    checklist: [
      { title: "登录门店账号", desc: "确认操作员、门店与班次信息" },
      { title: "创建收银草稿", desc: "先登记客户与商品明细" },
      { title: "录入回收细项", desc: "记录品类、成色、克重和扣减" },
      { title: "确认正式回收单", desc: "在回收确认页完成留档并生成正式回收单" }
    ],
    draftReady: false
  },

  onShow() {
    this.setTabBarIndex();
    const profile = getProfile();
    const loggedIn = Boolean(profile.loggedIn);
    this.setData({
      profile: profile,
      stats: loggedIn ? getDashboardStats(profile) : this.data.stats,
      resourceStats: {
        members: loggedIn ? getMemberStats().total : 0,
        vipMembers: loggedIn ? getMemberStats().vipCount : 0,
        products: loggedIn ? getProductStats().total : 0,
        activeProducts: loggedIn ? getProductStats().activeCount : 0
      },
      draftReady: loggedIn && Boolean(getDraft().customerName || getDraft().itemName),
      quickActions: loggedIn ? allQuickActions.filter(function(item) {
        return canAccessFeature(item.permission);
      }) : []
    });
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 0 });
    }
  },

  goAction(event) {
    if (!getProfile().loggedIn) {
      this.goLogin();
      return;
    }
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
