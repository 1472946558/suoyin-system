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
 * 创建日期: 2026-07-03
 */

const { canAccessFeature, getProfile } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");
const {
  getMaterialEntries,
  deleteMaterialEntry,
  listMaterialsOnline,
  createMaterialOnline,
  outboundMaterialOnline,
  getMaterialStats,
  getPurityStats
} = require("../../utils/inventoryStore");

const TYPE_VALUES = ["pledge", "leftover", "recycle"];
const TYPE_LABELS = ["抵押寄存", "余料", "回收旧料"];

function buildForm(profile) {
  return {
    type: "pledge",
    customerName: "",
    category: "黄金",
    purity: "足金9999",
    weightGram: "",
    amount: "",
    remainingWeightGram: "",
    storeName: (profile && (profile.storeName || profile.visibleStores && profile.visibleStores[0] && profile.visibleStores[0].name)) || "当前门店",
    status: "在库",
    dueDate: "",
    remark: ""
  };
}

function typeLabel(type) {
  return TYPE_LABELS[Math.max(0, TYPE_VALUES.indexOf(type))] || type;
}

function normalizeRow(row) {
  return Object.assign({}, row, {
    typeText: typeLabel(row.type),
    weightText: `${row.weightGram || 0}g`,
    remainingText: `${row.remainingWeightGram || 0}g`,
    amountText: Number(row.amount || 0).toFixed(2),
    canDelete: row.source !== "回收单"
  });
}

function filterRows(rows, keyword) {
  const key = String(keyword || "").trim().toLowerCase();
  if (!key) return rows;
  return rows.filter(function(item) {
    return [item.orderNo, item.customerName, item.category, item.purity, item.storeName, item.typeText].join(" ").toLowerCase().indexOf(key) > -1;
  });
}

Page({
  data: {
    profile: getProfile(),
    typeValues: TYPE_VALUES,
    typeLabels: TYPE_LABELS,
    typeIndex: 0,
    form: buildForm(getProfile()),
    keyword: "",
    rows: [],
    filteredRows: [],
    stats: getMaterialStats([]),
    purityStats: [],
    loading: false
  },

  onShow() {
    if (!canAccessFeature("materials")) {
      wx.showToast({ title: "当前账号没有旧料管理权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    this.loadMaterials();
  },

  loadMaterials() {
    this.setData({ loading: true, profile: getProfile() });
    listMaterialsOnline().then((ledger) => {
      const rows = ((ledger && Array.isArray(ledger.items)) ? ledger.items : getMaterialEntries()).map(normalizeRow);
      this.setData({
        rows,
        stats: ledger && ledger.items ? {
          todayWeightText: String(ledger.todayWeightGram || 0),
          todayAmountText: Number(ledger.todayAmount || 0).toFixed(2),
          monthWeightText: String(ledger.monthWeightGram || 0),
          monthAmountText: Number(ledger.monthAmount || 0).toFixed(2),
          remainingWeightText: String(ledger.remainingWeightGram || 0),
          pledgeCount: ledger.pledgeCount || 0,
          pledgeAmountText: Number(ledger.pledgeAmount || 0).toFixed(2)
        } : getMaterialStats(rows),
        purityStats: ledger && Array.isArray(ledger.purityStats) ? ledger.purityStats.map(function(item) {
          return {
            purity: item.purity,
            count: item.count,
            weightText: String(item.weightGram || 0),
            amountText: Number(item.amount || 0).toFixed(2)
          };
        }) : getPurityStats(rows)
      });
      this.applyFilters();
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  applyFilters() {
    this.setData({
      filteredRows: filterRows(this.data.rows, this.data.keyword)
    });
  },

  onKeywordInput(event) {
    this.setData({ keyword: event.detail.value });
    this.applyFilters();
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const update = {};
    update[`form.${key}`] = event.detail.value;
    this.setData(update);
  },

  onTypeChange(event) {
    const typeIndex = Number(event.detail.value);
    this.setData({
      typeIndex,
      "form.type": TYPE_VALUES[typeIndex] || "pledge"
    });
  },

  saveEntry() {
    const form = this.data.form;
    if (!form.category || !form.purity) {
      wx.showToast({ title: "请填写品类和成色", icon: "none" });
      return;
    }
    this.setData({ loading: true });
    createMaterialOnline(Object.assign({}, form, {
      source: "miniapp",
      remainingWeightGram: form.remainingWeightGram || form.weightGram
    })).then(() => {
      wx.showToast({ title: "已登记", icon: "success" });
      this.setData({
        typeIndex: 0,
        form: buildForm(getProfile())
      });
      this.loadMaterials();
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  deleteEntry(event) {
    const id = event.currentTarget.dataset.id;
    this.setData({ loading: true });
    outboundMaterialOnline(id, "小程序旧料出库").then(() => {
      deleteMaterialEntry(id);
      wx.showToast({ title: "已出库", icon: "success" });
      this.loadMaterials();
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    }).finally(() => {
      this.setData({ loading: false });
    });
  }
});
