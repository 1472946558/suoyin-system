/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: inventoryStore.js
 * 功能描述: 库存与旧料本地台账
 * 作者: 廖心慈
 * 创建日期: 2026-07-03
 */

const { appConfig } = require("./config");
const { request } = require("./apiClient");
const { getScopedStorageSync, setScopedStorageSync } = require("./sessionStorage");

const INVENTORY_KEY = "gr_inventory_entries";
const MATERIAL_KEY = "gr_material_entries";

function pad(value) {
  return value < 10 ? `0${value}` : String(value);
}

function nowText() {
  const current = new Date();
  return `${current.getFullYear()}-${pad(current.getMonth() + 1)}-${pad(current.getDate())} ${pad(current.getHours())}:${pad(current.getMinutes())}`;
}

function todayKey() {
  return nowText().slice(0, 10);
}

function monthKey() {
  return nowText().slice(0, 7);
}

function safeNumber(value) {
  const next = Number(value || 0);
  return Number.isFinite(next) ? next : 0;
}

function formatMoney(value) {
  return safeNumber(value).toFixed(2);
}

function formatWeight(value) {
  const next = safeNumber(value);
  return Number.isInteger(next) ? String(next) : next.toFixed(2).replace(/0+$/, "").replace(/\.$/, "");
}

function getStorageArray(key) {
  const stored = getScopedStorageSync(key, []);
  return Array.isArray(stored) ? stored : [];
}

function saveStorageArray(key, list) {
  setScopedStorageSync(key, Array.isArray(list) ? list : []);
}

function createId(prefix) {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
}

function normalizeInventoryEntry(entry) {
  const source = entry || {};
  return {
    id: source.id || createId("inventory"),
    storeId: String(source.storeId || "").trim(),
    styleNo: String(source.styleNo || source.sku || "").trim(),
    name: String(source.name || "").trim(),
    category: String(source.category || "黄金").trim(),
    purity: String(source.purity || "客户自填").trim(),
    pieceCount: safeNumber(source.pieceCount || source.inventory),
    weightGram: safeNumber(source.weightGram || source.gramWeight),
    storeName: String(source.storeName || source.store || "当前门店").trim(),
    costAmount: safeNumber(source.costAmount || source.unitCost || source.price),
    status: String(source.status || "normal").trim(),
    source: String(source.source || "手动入库").trim(),
    createdAt: source.createdAt || nowText(),
    remark: source.remark || ""
  };
}

function normalizeMaterialEntry(entry) {
  const source = entry || {};
  const weightGram = safeNumber(source.weightGram);
  return {
    id: source.id || createId("material"),
    storeId: String(source.storeId || "").trim(),
    orderNo: source.orderNo || `JL-${todayKey().replace(/-/g, "")}-${Math.floor(Math.random() * 900 + 100)}`,
    type: source.type || "pledge",
    customerName: String(source.customerName || "未填客户").trim(),
    category: String(source.category || "黄金").trim(),
    purity: String(source.purity || "客户自填").trim(),
    weightGram,
    amount: safeNumber(source.amount),
    remainingWeightGram: safeNumber(source.remainingWeightGram || weightGram),
    storeName: String(source.storeName || "当前门店").trim(),
    status: String(source.status || "在库").trim(),
    dueDate: source.dueDate || "",
    remark: source.remark || "",
    createdAt: source.createdAt || nowText(),
    source: source.source || "手动登记"
  };
}

function unwrapResponse(response, fallback) {
  if (response && response.local) {
    return fallback;
  }
  if (response && response.code === 0) {
    return response.data;
  }
  return response || fallback;
}

function buildInventoryPayload(entry) {
  const item = normalizeInventoryEntry(entry);
  const payload = {
    storeId: item.storeId,
    styleNo: item.styleNo,
    name: item.name,
    category: item.category,
    purity: item.purity,
    pieceCount: item.pieceCount,
    weightGram: item.weightGram,
    costAmount: item.costAmount,
    status: item.status === "out" ? "outbound" : "in_stock",
    source: item.source === "文件导入" ? "file" : "miniapp",
    remark: item.remark
  };
  if (!payload.storeId) {
    delete payload.storeId;
  }
  return payload;
}

function buildMaterialPayload(entry) {
  const item = normalizeMaterialEntry(entry);
  const payload = {
    storeId: item.storeId,
    type: item.type,
    orderNo: item.orderNo,
    customerName: item.customerName,
    category: item.category,
    purity: item.purity,
    weightGram: item.weightGram,
    amount: item.amount,
    remainingWeightGram: item.remainingWeightGram || item.weightGram,
    status: item.status === "已出库" ? "outbound" : "in_stock",
    dueDate: item.dueDate,
    remark: item.remark,
    source: "miniapp"
  };
  if (!payload.storeId) {
    delete payload.storeId;
  }
  return payload;
}

function listInventoryOnline() {
  return request({
    endpoint: appConfig.endpoints.inventoryItems,
    method: "GET"
  }).then(function(response) {
    const data = unwrapResponse(response, null);
    if (!data || !Array.isArray(data.items)) {
      return {
        items: getInventoryEntries(),
        totalStyleCount: getInventoryEntries().length,
        totalPieceCount: 0,
        totalWeightGram: 0,
        totalCostAmount: 0
      };
    }
    return data;
  });
}

function createInventoryOnline(entry) {
  return request({
    endpoint: appConfig.endpoints.inventoryItems,
    method: "POST",
    data: buildInventoryPayload(entry)
  }).then(function(response) {
    return normalizeInventoryEntry(unwrapResponse(response, entry));
  });
}

function listMaterialsOnline() {
  return request({
    endpoint: appConfig.endpoints.materialItems,
    method: "GET"
  }).then(function(response) {
    const data = unwrapResponse(response, null);
    if (!data || !Array.isArray(data.items)) {
      return {
        items: getMaterialEntries(),
        purityStats: getPurityStats(getMaterialEntries())
      };
    }
    return data;
  });
}

function createMaterialOnline(entry) {
  return request({
    endpoint: appConfig.endpoints.materialItems,
    method: "POST",
    data: buildMaterialPayload(entry)
  }).then(function(response) {
    return normalizeMaterialEntry(unwrapResponse(response, entry));
  });
}

function outboundMaterialOnline(id, remark) {
  return request({
    endpoint: appConfig.endpoints.materialOutbound,
    params: { id },
    method: "POST",
    data: { remark: remark || "小程序出库" }
  }).then(function(response) {
    return normalizeMaterialEntry(unwrapResponse(response, {}));
  });
}

function getInventoryEntries() {
  return getStorageArray(INVENTORY_KEY).map(normalizeInventoryEntry);
}

function saveInventoryEntries(list) {
  saveStorageArray(INVENTORY_KEY, list.map(normalizeInventoryEntry));
}

function addInventoryEntry(entry) {
  const next = normalizeInventoryEntry(entry);
  const list = [next].concat(getInventoryEntries());
  saveInventoryEntries(list);
  return next;
}

function deleteInventoryEntry(id) {
  const list = getInventoryEntries().filter(function(item) {
    return item.id !== id;
  });
  saveInventoryEntries(list);
  return list;
}

function buildProductInventoryRows(products) {
  return (products || []).map(function(product) {
    const pieceCount = safeNumber(product.inventory);
    return normalizeInventoryEntry({
      id: `product-${product.id || product.sku}`,
      styleNo: product.sku,
      name: product.name,
      category: product.category || product.categoryTab || "未分类",
      purity: product.purity || "客户自填",
      pieceCount,
      weightGram: safeNumber(product.gramWeight) * Math.max(pieceCount, 1),
      storeName: Array.isArray(product.stores) && product.stores.length ? product.stores[0] : "当前门店",
      costAmount: safeNumber(product.retailPrice || product.benchPrice),
      status: product.stockStatus === "low" ? "low" : (pieceCount > 0 ? "normal" : "out"),
      source: "商品目录",
      createdAt: ""
    });
  });
}

function getInventoryStats(list) {
  const rows = list || [];
  return {
    styles: rows.length,
    pieces: rows.reduce(function(sum, item) {
      return sum + safeNumber(item.pieceCount);
    }, 0),
    weightText: formatWeight(rows.reduce(function(sum, item) {
      return sum + safeNumber(item.weightGram);
    }, 0)),
    amountText: formatMoney(rows.reduce(function(sum, item) {
      return sum + safeNumber(item.costAmount) * Math.max(safeNumber(item.pieceCount), 1);
    }, 0))
  };
}

function parseRecycleOrderMaterial(order) {
  const sourceItems = Array.isArray(order.items) && order.items.length ? order.items : [{
    category: order.itemCategory || "旧料",
    purity: order.purity || "客户自填",
    weightGram: order.netWeight || order.grossWeight || 0
  }];
  const itemCount = Math.max(sourceItems.length, 1);
  return sourceItems.map(function(item, index) {
    return normalizeMaterialEntry({
      id: `recycle-${order.id || order.orderNo}-${index}`,
      orderNo: order.orderNo || order.id,
      type: "recycle",
      customerName: order.customerName || "未填客户",
      category: item.category || order.itemCategory || "旧料",
      purity: item.purity || order.purity || "客户自填",
      weightGram: item.weightGram || order.netWeight || order.grossWeight || 0,
      amount: safeNumber(order.amount || order.confirmedAmount || order.estimatedAmount) / itemCount,
      remainingWeightGram: item.weightGram || order.netWeight || order.grossWeight || 0,
      storeName: order.storeName || "当前门店",
      status: order.statusText || (order.status === "confirmed" ? "已确认" : "待确认"),
      createdAt: order.createdAt || nowText(),
      source: "回收单"
    });
  });
}

function getMaterialEntries() {
  return getStorageArray(MATERIAL_KEY).map(normalizeMaterialEntry);
}

function saveMaterialEntries(list) {
  saveStorageArray(MATERIAL_KEY, list.map(normalizeMaterialEntry));
}

function addMaterialEntry(entry) {
  const next = normalizeMaterialEntry(entry);
  const list = [next].concat(getMaterialEntries());
  saveMaterialEntries(list);
  return next;
}

function deleteMaterialEntry(id) {
  const list = getMaterialEntries().filter(function(item) {
    return item.id !== id;
  });
  saveMaterialEntries(list);
  return list;
}

function getMaterialStats(list) {
  const rows = list || [];
  const todayRows = rows.filter(function(item) {
    return String(item.createdAt || "").slice(0, 10) === todayKey();
  });
  const monthRows = rows.filter(function(item) {
    return String(item.createdAt || "").slice(0, 7) === monthKey();
  });
  const pledgeRows = rows.filter(function(item) {
    return item.type === "pledge";
  });
  return {
    todayWeightText: formatWeight(todayRows.reduce(function(sum, item) { return sum + safeNumber(item.weightGram); }, 0)),
    todayAmountText: formatMoney(todayRows.reduce(function(sum, item) { return sum + safeNumber(item.amount); }, 0)),
    monthWeightText: formatWeight(monthRows.reduce(function(sum, item) { return sum + safeNumber(item.weightGram); }, 0)),
    monthAmountText: formatMoney(monthRows.reduce(function(sum, item) { return sum + safeNumber(item.amount); }, 0)),
    remainingWeightText: formatWeight(rows.reduce(function(sum, item) { return sum + safeNumber(item.remainingWeightGram); }, 0)),
    pledgeCount: pledgeRows.length,
    pledgeAmountText: formatMoney(pledgeRows.reduce(function(sum, item) { return sum + safeNumber(item.amount); }, 0))
  };
}

function getPurityStats(list) {
  const map = {};
  (list || []).forEach(function(item) {
    const key = item.purity || "未填成色";
    if (!map[key]) {
      map[key] = { purity: key, count: 0, weight: 0, amount: 0 };
    }
    map[key].count += 1;
    map[key].weight += safeNumber(item.weightGram);
    map[key].amount += safeNumber(item.amount);
  });
  return Object.keys(map).map(function(key) {
    const item = map[key];
    return {
      purity: item.purity,
      count: item.count,
      weightText: formatWeight(item.weight),
      amountText: formatMoney(item.amount)
    };
  }).sort(function(left, right) {
    return safeNumber(right.weightText) - safeNumber(left.weightText);
  });
}

module.exports = {
  formatMoney,
  formatWeight,
  getInventoryEntries,
  addInventoryEntry,
  deleteInventoryEntry,
  listInventoryOnline,
  createInventoryOnline,
  buildProductInventoryRows,
  getInventoryStats,
  getMaterialEntries,
  addMaterialEntry,
  deleteMaterialEntry,
  listMaterialsOnline,
  createMaterialOnline,
  outboundMaterialOnline,
  parseRecycleOrderMaterial,
  getMaterialStats,
  getPurityStats
};
