const { getOrders } = require("../../utils/orderStore");
const { getProfile, clearProfile } = require("../../utils/userStore");

Page({
  data: {
    profile: getProfile(),
    displayName: "商用急送账户",
    displaySubtitle: "请先配置你的测试资料，用真实信息跑通演示流程。",
    accountButtonText: "登录 / 注册",
    stats: {
      orders: 0,
      coupons: 3,
      addresses: 2
    },
    menus: [
      { icon: "/assets/ui/pickup.png", title: "常用地址", desc: "管理寄件和收件地址" },
      { icon: "/assets/ui/coupon.png", title: "优惠券", desc: "新用户立减券和企业券" },
      { icon: "/assets/ui/support.png", title: "在线客服", desc: "订单异常和售后咨询" },
      { icon: "/assets/ui/enterprise.png", title: "企业月结", desc: "预留商家高频下单入口" }
    ]
  },

  onShow() {
    this.setTabBarIndex();
    const profile = getProfile();
    this.setData({
      profile,
      displayName: profile.loggedIn ? profile.name : "商用急送账户",
      displaySubtitle: profile.loggedIn ? `${profile.company || "个人用户"} · ${profile.phone}` : "请先配置你的测试资料，用真实信息跑通演示流程。",
      accountButtonText: profile.loggedIn ? "编辑资料" : "登录 / 注册",
      "stats.orders": getOrders().length
    });
  },

  setTabBarIndex() {
    if (typeof this.getTabBar === "function" && this.getTabBar()) {
      this.getTabBar().setData({ selected: 4 });
    }
  },

  goAccount() {
    wx.navigateTo({ url: "/pages/account/index" });
  },

  logout() {
    clearProfile();
    wx.showToast({ title: "已退出登录", icon: "none" });
    const profile = getProfile();
    this.setData({
      profile,
      displayName: "商用急送账户",
      displaySubtitle: "请先配置你的测试资料，用真实信息跑通演示流程。",
      accountButtonText: "登录 / 注册"
    });
  },

  onMenuTap(event) {
    wx.showToast({
      title: `${event.currentTarget.dataset.title} 为演示入口`,
      icon: "none"
    });
  }
});
