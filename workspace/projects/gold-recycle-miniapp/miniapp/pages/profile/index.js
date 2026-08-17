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
 * 创建日期: 2026-06-03
 */

const { listOrdersOnline, listCashierOrdersOnline, clearDraft } = require("../../utils/orderStore");
const { appConfig } = require("../../utils/config");
const { getProfile, logoutOnline } = require("../../utils/userStore");
const { listMembersOnline, listProductsOnline, getMemberStats, getProductStats } = require("../../utils/catalogStore");
const { formatRequestError } = require("../../utils/apiClient");
const { refreshProfileStoreBinding } = require("../../utils/storeSession");

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

  const businessModules = [
    { title: "会员管理", desc: "新增会员、编辑资料、查看跟进记录", key: "members" },
    { title: "商品目录", desc: "维护商品模板、报价和展示状态", key: "products" },
    { title: "库存系统", desc: "商品入库、款式编号和库存查询", key: "inventory" },
    { title: "旧料管理", desc: "旧料统计、余料和抵押寄存", key: "materials" },
    { title: "订单记录", desc: "查看收银与回收记录", key: "orders" }
  ];

  return publicModules.slice(0, 2).concat(
    businessModules,
    publicModules[2],
    publicModules[3],
    publicModules[4],
    { title: "清空草稿", desc: "清空当前账号未完成的收银和录单草稿", key: "draft" }
  );
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
    refreshProfileStoreBinding().finally(() => {
      const profile = getProfile();
      const loggedIn = Boolean(profile.loggedIn);
      this.setData({
        profile,
        modules: buildModules(profile),
        stats: {
          orders: 0,
          storeCode: profile.storeCode || "--",
          role: profile.role || "--",
          members: 0,
          products: 0
        }
      });
      if (loggedIn) {
        this.loadStats();
      }
    });
  },

  loadStats() {
    Promise.all([
      listOrdersOnline(),
      listCashierOrdersOnline(),
      listMembersOnline(),
      listProductsOnline()
    ]).then((results) => {
      const orders = results[0] || [];
      const cashierOrders = results[1] || [];
      const members = results[2] && Array.isArray(results[2].items) ? results[2].items : [];
      const products = results[3] && Array.isArray(results[3].items) ? results[3].items : [];
      this.setData({
        stats: Object.assign({}, this.data.stats, {
          orders: orders.length + cashierOrders.length,
          members: getMemberStats(members).total,
          products: getProductStats(products).total
        })
      });
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
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
    if (key === "inventory") {
      wx.navigateTo({ url: "/pages/inventory/index" });
      return;
    }
    if (key === "materials") {
      wx.navigateTo({ url: "/pages/materials/index" });
      return;
    }
    if (key === "orders") {
      wx.reLaunch({ url: "/pages/orders/index" });
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
