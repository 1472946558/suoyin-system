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

const { listProductsOnline } = require("../../utils/catalogStore");
const { canAccessFeature, getProfile } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");
const {
  getInventoryEntries,
  deleteInventoryEntry,
  listInventoryOnline,
  createInventoryOnline,
  buildProductInventoryRows,
  getInventoryStats
} = require("../../utils/inventoryStore");

function buildForm(profile) {
  return {
    styleNo: `KC-${Date.now() % 100000}`,
    name: "",
    category: "黄金",
    purity: "足金9999",
    pieceCount: "1",
    weightGram: "",
    costAmount: "",
    storeName: (profile && (profile.storeName || profile.visibleStores && profile.visibleStores[0] && profile.visibleStores[0].name)) || "当前门店",
    remark: ""
  };
}

function normalizeRow(row) {
  const statusText = row.status === "low" ? "库存较低" : (row.status === "out" ? "无库存" : (row.status === "review" ? "待盘点" : "库存正常"));
  return Object.assign({}, row, {
    statusText,
    weightText: `${row.weightGram || 0}g`,
    amountText: Number(row.costAmount || 0).toFixed(2),
    canDelete: row.source !== "商品目录"
  });
}

function filterRows(rows, keyword) {
  const key = String(keyword || "").trim().toLowerCase();
  if (!key) return rows;
  return rows.filter(function(item) {
    return [item.styleNo, item.name, item.category, item.purity, item.storeName, item.source].join(" ").toLowerCase().indexOf(key) > -1;
  });
}

Page({
  data: {
    profile: getProfile(),
    form: buildForm(getProfile()),
    keyword: "",
    rows: [],
    filteredRows: [],
    stats: getInventoryStats([]),
    loading: false
  },

  onShow() {
    if (!canAccessFeature("inventory")) {
      wx.showToast({ title: "当前账号没有库存权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    this.loadInventory();
  },

  loadInventory() {
    this.setData({ loading: true, profile: getProfile() });
    Promise.all([listInventoryOnline(), listProductsOnline()]).then(([ledger, result]) => {
      const products = result && Array.isArray(result.items) ? result.items : [];
      const ledgerRows = ledger && Array.isArray(ledger.items) ? ledger.items : getInventoryEntries();
      const rows = ledgerRows.concat(buildProductInventoryRows(products)).map(normalizeRow);
      this.setData({
        rows,
        stats: getInventoryStats(rows),
        form: Object.assign({}, this.data.form, {
          storeName: this.data.form.storeName || buildForm(getProfile()).storeName
        })
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

  saveEntry() {
    const form = this.data.form;
    if (!form.styleNo || !form.name) {
      wx.showToast({ title: "请填写款式编号和商品名称", icon: "none" });
      return;
    }
    this.setData({ loading: true });
    createInventoryOnline(Object.assign({}, form, {
      source: "miniapp",
      status: "in_stock"
    })).then(() => {
      wx.showToast({ title: "已入库", icon: "success" });
      this.setData({
        form: buildForm(getProfile())
      });
      this.loadInventory();
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  deleteEntry(event) {
    const id = event.currentTarget.dataset.id;
    deleteInventoryEntry(id);
    wx.showToast({ title: "已删除", icon: "none" });
    this.loadInventory();
  },

  showImportTip() {
    wx.showModal({
      title: "文件导入",
      content: "小程序端先支持手动入库和查询。批量文件导入请在后台库存页面下载 CSV 模板并导入。",
      confirmText: "知道了",
      showCancel: false
    });
  }
});
