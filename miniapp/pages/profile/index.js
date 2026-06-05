const { getOrders, getCashierOrders, clearDraft } = require("../../utils/orderStore");
const { getProfile, logoutOnline } = require("../../utils/userStore");
const { getMemberStats, getProductStats } = require("../../utils/catalogStore");

function buildModules(profile) {
  const publicModules = [
    { title: "门店登录", desc: "修改操作员、门店、班次与岗位", key: "account" },
    { title: "门店资料", desc: "查看客户服务、协议展示和门店识别资料", key: "store-info" },
    { title: "打印设备", desc: "查看门店打印机采购与对接建议", key: "printer" },
    { title: "用户协议", desc: "查看小程序用户服务协议", key: "terms" },
    { title: "隐私协议", desc: "查看用户信息与照片留档说明", key: "privacy" }
  ];

  if (!profile || !profile.loggedIn) {
    return publicModules;
  }

  return publicModules.slice(0, 2).concat([
    { title: "会员管理", desc: "新增会员、编辑资料、查看跟进记录", key: "members" },
    { title: "商品目录", desc: "维护商品图片、分类、状态和库存", key: "products" },
    { title: "订单记录", desc: "查看收银与回收记录", key: "orders" },
    publicModules[2],
    publicModules[3],
    publicModules[4],
    { title: "清空草稿", desc: "清空当前账号未完成的收银和录单草稿", key: "draft" }
  ]);
}

Page({
  data: {
    profile: getProfile(),
    modules: buildModules(getProfile()),
    stats: {
      orders: 0,
      storeCode: "--",
      role: "--",
      members: 0,
      products: 0
    }
  },

  onShow() {
    this.setTabBarIndex();
    const profile = getProfile();
    const loggedIn = Boolean(profile.loggedIn);
    this.setData({
      profile,
      modules: buildModules(profile),
      stats: {
        orders: loggedIn ? getOrders().length + getCashierOrders().length : 0,
        storeCode: profile.storeCode || "--",
        role: profile.role || "--",
        members: loggedIn ? getMemberStats().total : 0,
        products: loggedIn ? getProductStats().total : 0
      }
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
    logoutOnline()
      .then(() => {
        wx.showToast({ title: "已退出登录", icon: "none" });
      })
      .finally(() => {
        const profile = getProfile();
        this.setData({
          profile,
          modules: buildModules(profile),
          stats: Object.assign({}, this.data.stats, {
            orders: 0,
            storeCode: "--",
            role: "--",
            members: 0,
            products: 0
          })
        });
        wx.navigateTo({ url: "/pages/account/index" });
      });
  },

  showPrinterAdvice() {
    wx.showModal({
      title: "打印机建议",
      content: "建议客户购买支持蓝牙连接、并提供 ESC/POS、TSPL 或 CPCL 打印指令文档的热敏小票机/标签机。仅能用手机 App 打印、商家明确不支持系统对接的型号，不建议用于自动出单。",
      confirmText: "知道了",
      showCancel: false
    });
  },

  onModuleTap(event) {
    const { key, title } = event.currentTarget.dataset;
    if (key === "account") {
      this.goAccount();
      return;
    }
    if (key === "store-info") {
      wx.navigateTo({ url: "/pages/store-info/index" });
      return;
    }
    if (key === "draft") {
      clearDraft();
      wx.showToast({ title: "当前草稿已清空", icon: "none" });
      return;
    }
    if (key === "members") {
      wx.navigateTo({ url: "/pages/members/index" });
      return;
    }
    if (key === "products") {
      wx.navigateTo({ url: "/pages/products/index" });
      return;
    }
    if (key === "orders") {
      wx.switchTab({ url: "/pages/orders/index" });
      return;
    }
    if (key === "printer") {
      this.showPrinterAdvice();
      return;
    }
    if (key === "terms") {
      wx.navigateTo({ url: "/pages/legal/terms" });
      return;
    }
    if (key === "privacy") {
      wx.navigateTo({ url: "/pages/legal/privacy" });
      return;
    }
    wx.showToast({ title: `${title} 待完善`, icon: "none" });
  }
});
