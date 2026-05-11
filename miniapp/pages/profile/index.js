const { appConfig } = require("../../utils/config");
const { getRuntimeConfig, saveRuntimeConfig, resetRuntimeConfig } = require("../../utils/apiClient");
const { getOrders, clearDraft } = require("../../utils/orderStore");
const { getProfile, clearProfile } = require("../../utils/userStore");
const { getMemberStats, getProductStats } = require("../../utils/catalogStore");

function buildEnvItems() {
  const runtimeConfig = getRuntimeConfig();
  return [
    { title: "运行模式", value: runtimeConfig.mode === "mock" ? "mock" : "production" },
    { title: "接口地址", value: runtimeConfig.apiBaseUrl || "--" },
    { title: "品牌名", value: appConfig.appName },
    { title: "服务热线", value: appConfig.servicePhone }
  ];
}

function buildRuntimeSummary() {
  const runtimeConfig = getRuntimeConfig();
  return runtimeConfig.mode === "mock"
    ? "当前使用本地 mock 数据，适合演示 UI 和前台流程。"
    : `当前已切到真实 API 联调模式，接口地址：${runtimeConfig.apiBaseUrl}`;
}

Page({
  data: {
    profile: getProfile(),
    envItems: buildEnvItems(),
    modules: [
      { title: "门店登录", desc: "操作员、门店、班次已支持本地保存", key: "account" },
      { title: "会员管理", desc: "查看常客档案并带入收银", key: "members" },
      { title: "商品目录", desc: "查看商品与报价参考模板", key: "products" },
      { title: "联调模式", desc: "在 mock / 真实 API 之间切换", key: "runtime" },
      { title: "接口地址", desc: "修改小程序联调使用的 API Base URL", key: "api-base" },
      { title: "草稿重置", desc: "清空当前未完成收银与录单数据", key: "draft" },
      { title: "重置联调配置", desc: "恢复默认 mock 模式和本地接口地址", key: "runtime-reset" }
    ],
    runtimeSummary: buildRuntimeSummary(),
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
    this.setData({
      profile,
      envItems: buildEnvItems(),
      runtimeSummary: buildRuntimeSummary(),
      stats: {
        orders: getOrders().length,
        storeCode: profile.storeCode || "--",
        role: profile.role || "--",
        members: getMemberStats().total,
        products: getProductStats().total
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

  refreshRuntimePanel() {
    this.setData({
      envItems: buildEnvItems(),
      runtimeSummary: buildRuntimeSummary()
    });
  },

  onRuntimeModeTap() {
    const runtimeConfig = getRuntimeConfig();
    const isMock = runtimeConfig.mode === "mock";
    wx.showActionSheet({
      itemList: isMock ? ["切到真实 API 联调"] : ["切回本地 mock"],
      success: (result) => {
        if (result.tapIndex !== 0) return;
        saveRuntimeConfig({
          mode: isMock ? "production" : "mock"
        });
        this.refreshRuntimePanel();
        wx.showToast({
          title: isMock ? "已切到真实 API" : "已切回 mock",
          icon: "none"
        });
      }
    });
  },

  onApiBaseTap() {
    const runtimeConfig = getRuntimeConfig();
    wx.showModal({
      title: "设置接口地址",
      content: "请输入当前联调环境的 API Base URL",
      editable: true,
      placeholderText: "http://127.0.0.1:18082",
      confirmText: "保存",
      cancelText: "取消",
      value: runtimeConfig.apiBaseUrl || "",
      success: (result) => {
        if (!result.confirm) return;
        const value = (result.content || "").trim();
        if (!value) {
          wx.showToast({ title: "接口地址不能为空", icon: "none" });
          return;
        }
        saveRuntimeConfig({
          apiBaseUrl: value
        });
        this.refreshRuntimePanel();
        wx.showToast({ title: "接口地址已保存", icon: "none" });
      }
    });
  },

  onRuntimeResetTap() {
    resetRuntimeConfig();
    this.refreshRuntimePanel();
    wx.showToast({ title: "已恢复默认联调配置", icon: "none" });
  },

  logout() {
    clearProfile();
    wx.showToast({ title: "已退出当前门店账号", icon: "none" });
    this.setData({
      profile: getProfile(),
      stats: Object.assign({}, this.data.stats, {
        storeCode: "--",
        role: "--"
      })
    });
  },

  onModuleTap(event) {
    const { key, title } = event.currentTarget.dataset;
    if (key === "account") {
      this.goAccount();
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
    if (key === "runtime") {
      this.onRuntimeModeTap();
      return;
    }
    if (key === "api-base") {
      this.onApiBaseTap();
      return;
    }
    if (key === "runtime-reset") {
      this.onRuntimeResetTap();
      return;
    }
    wx.showToast({ title: `${title} 仍需补充`, icon: "none" });
  }
});
