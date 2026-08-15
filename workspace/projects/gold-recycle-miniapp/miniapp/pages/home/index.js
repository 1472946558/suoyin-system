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

const { appConfig } = require("../../utils/config");
const { getDashboardStatsOnline, getDraft } = require("../../utils/orderStore");
const { getProfile, canAccessFeature } = require("../../utils/userStore");
const { refreshProfileStoreBinding } = require("../../utils/storeSession");
const { listMembersOnline, listProductsOnline, getMemberStats, getProductStats } = require("../../utils/catalogStore");
const { getVisibleReferencePrices, refreshReferencePrices } = require("../../utils/goldPriceStore");
const { formatRequestError } = require("../../utils/apiClient");
const appointmentStore = require("../../utils/appointmentStore.js");

const allQuickActions = [
  { title: "开始收银", desc: "录入客户与商品明细", type: "tab", url: "/pages/order/index", icon: "/assets/ui/home-cashier.png", permission: "cashier" },
  { title: "回收录单", desc: "录入成色重量与报价", type: "tab", url: "/pages/track/index", icon: "/assets/ui/home-recycle.png", permission: "recycle" },
  { title: "查看订单", desc: "浏览历史回收单", type: "tab", url: "/pages/orders/index", icon: "/assets/ui/home-orders.png", permission: "orders" },
  { title: "预约管理", desc: "查看与处理门店预约", type: "nav", url: "/pages/appointments/index", icon: "/assets/ui/home-orders.png", permission: "appointments" },
  { title: "会员管理", desc: "常客档案与带入建单", type: "nav", url: "/pages/members/index", icon: "/assets/ui/home-members.png", permission: "members" },
  { title: "商品目录", desc: "门店商品与报价参考", type: "nav", url: "/pages/products/index", icon: "/assets/ui/home-products.png", permission: "products" },
  { title: "库存系统", desc: "商品入库与库存查询", type: "nav", url: "/pages/inventory/index", icon: "/assets/ui/list.png", permission: "inventory" },
  { title: "旧料管理", desc: "余料和抵押寄存台账", type: "nav", url: "/pages/materials/index", icon: "/assets/ui/weight.png", permission: "materials" },
  { title: "系统设置", desc: "门店与操作员设置", type: "tab", url: "/pages/profile/index", icon: "/assets/ui/home-settings.png", permission: "settings" }
];

function isOwnerProfile(profile) {
  const roleKey = String(profile && profile.roleKey || "").toLowerCase();
  const dataScope = String(profile && profile.dataScope || "").toLowerCase();
  return roleKey === "boss" || roleKey === "owner" || dataScope === "org_all" || dataScope === "all_stores";
}

function getHomeStoreText(profile) {
  if (!profile || !profile.loggedIn) {
    return "未登录门店";
  }
  if (isOwnerProfile(profile)) {
    return "全部门店";
  }
  return profile.storeName || "未登录门店";
}

function normalizeDashboardStats(stats, profile) {
  const nextStats = Object.assign({
    todayCount: 0,
    todayAmount: "0.00",
    todayCashierAmount: "0.00",
    todayRecycleAmount: "0.00",
    pendingCount: 0,
    completedCount: 0,
    scopeText: "未绑定门店",
    visibleStoreCount: 0,
    storeBreakdown: []
  }, stats || {});
  if (isOwnerProfile(profile)) {
    nextStats.scopeText = "全部门店";
  }
  return nextStats;
}

Page({
  data: {
    profile: getProfile(),
    homeStoreText: getHomeStoreText(getProfile()),
    prices: getVisibleReferencePrices(4),
    stats: {
      todayCount: 0,
      todayAmount: "0.00",
      todayCashierAmount: "0.00",
      todayRecycleAmount: "0.00",
      pendingCount: 0,
      completedCount: 0,
      scopeText: "未绑定门店",
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
    pendingAppointments: 0,
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
      homeStoreText: getHomeStoreText(profile),
      prices: getVisibleReferencePrices(4),
      stats: loggedIn ? this.data.stats : this.data.stats,
      resourceStats: loggedIn ? this.data.resourceStats : {
        members: 0,
        vipMembers: 0,
        products: 0,
        activeProducts: 0
      },
      draftReady: loggedIn && Boolean(getDraft().customerName || getDraft().itemName),
      quickActions: loggedIn ? allQuickActions.filter(function(item) {
        return canAccessFeature(item.permission);
      }) : [],
      pendingAppointments: loggedIn ? this.data.pendingAppointments : 0
    });
    if (loggedIn) {
      this.refreshReferencePrices();
      this.loadPendingAppointments();
      refreshProfileStoreBinding().then((nextProfile) => {
        if (!nextProfile || !nextProfile.loggedIn) {
          this.setData({
            profile: getProfile(),
            homeStoreText: "未登录门店",
            quickActions: [],
            resourceStats: {
              members: 0,
              vipMembers: 0,
              products: 0,
              activeProducts: 0
            }
          });
          this.goLogin();
          return;
        }
        this.setData({
          profile: nextProfile,
          homeStoreText: getHomeStoreText(nextProfile)
        });
        this.loadDashboard(nextProfile);
      });
    }
  },

  loadDashboard(profile) {
    Promise.all([
      getDashboardStatsOnline(profile),
      listMembersOnline(),
      listProductsOnline()
    ]).then((results) => {
      const stats = results[0];
      const members = results[1] && Array.isArray(results[1].items) ? results[1].items : [];
      const products = results[2] && Array.isArray(results[2].items) ? results[2].items : [];
      this.setData({
        stats: normalizeDashboardStats(stats, profile),
        resourceStats: {
          members: getMemberStats(members).total,
          vipMembers: getMemberStats(members).vipCount,
          products: getProductStats(products).total,
          activeProducts: getProductStats(products).activeCount
        }
      });
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    });
  },

  refreshReferencePrices() {
    refreshReferencePrices().then((prices) => {
      this.setData({
        prices: prices.slice(0, 4)
      });
    });
  },

  loadPendingAppointments() {
    appointmentStore.listStaffAppointments("PENDING")
      .then((resp) => {
        this.setData({ pendingAppointments: resp.total || 0 });
      })
      .catch(() => {});
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
