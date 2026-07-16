/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: orderStore.js
 * 功能描述: 工具函数
 * 作者: 廖心慈
 * 创建日期: 2026-06-05
 */

const STORAGE_KEY = "gr_recycle_orders";
const CASHIER_STORAGE_KEY = "gr_cashier_orders";
const DRAFT_KEY = "gr_recycle_draft";
const { appConfig, request } = require("./apiClient");
const { getSuggestedPrice, refreshReferencePrices } = require("./goldPriceStore");
const { getScopedStorageSync, setScopedStorageSync, removeScopedStorageSync } = require("./sessionStorage");
const photoRules = appConfig.recyclePhotoRules;

const statusMap = {
  pending_confirm: "待确认",
  completed: "已完成",
  draft: "待确认",
  confirmed: "已完成",
  cancelled: "已作废"
};

const cashierStatusMap = {
  paid: "已完成",
  pending: "待复核",
  refunded: "已退单",
  completed: "已完成",
  cancelled: "已退单"
};

function getOnlineStoreContext(profile) {
  const source = profile || wx.getStorageSync("gr_operator_profile") || {};
  return {
    storeId: String(source.storeId || "").trim(),
    storeName: String(source.storeName || "").trim(),
    storeCode: String(source.storeCode || "").trim()
  };
}

function safeNumber(value) {
  const parsed = parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function getFileName(filePath, fallbackName) {
  const source = String(filePath || fallbackName || "").trim();
  if (!source) {
    return fallbackName || `photo-${Date.now()}.jpg`;
  }
  const withoutQuery = source.split("?")[0];
  const segments = withoutQuery.split("/");
  return segments[segments.length - 1] || fallbackName || `photo-${Date.now()}.jpg`;
}

function inferContentType(fileName) {
  const lowered = String(fileName || "").toLowerCase();
  if (lowered.endsWith(".png")) return "image/png";
  if (lowered.endsWith(".webp")) return "image/webp";
  if (lowered.endsWith(".heic")) return "image/heic";
  return "image/jpeg";
}

function getFileInfo(filePath) {
  return new Promise((resolve, reject) => {
    wx.getFileInfo({
      filePath,
      success(result) {
        resolve(result || {});
      },
      fail(error) {
        reject(error);
      }
    });
  });
}

function readFileBuffer(filePath) {
  return new Promise((resolve, reject) => {
    wx.getFileSystemManager().readFile({
      filePath,
      success(result) {
        resolve(result.data);
      },
      fail(error) {
        reject(error);
      }
    });
  });
}

function requestUploadPreparation(storeId, fileName, fileInfo) {
  return request({
    endpoint: appConfig.endpoints.prepareRecyclePhotoUpload,
    method: "POST",
    data: {
      storeId,
      fileName,
      contentType: inferContentType(fileName),
      sizeBytes: Number(fileInfo.size || 0)
    }
  }).then((payload) => {
    return payload && payload.data ? payload.data : payload;
  });
}

function uploadWithMultipart(filePath, preparation) {
  return new Promise((resolve, reject) => {
    wx.uploadFile({
      url: preparation.uploadUrl,
      filePath,
      name: "file",
      header: preparation.headers || {},
      formData: preparation.formFields || {},
      success(result) {
        if (Number(result.statusCode) >= 200 && Number(result.statusCode) < 300) {
          resolve();
          return;
        }
        reject(new Error(`图片上传失败，状态码 ${result.statusCode || "--"}`));
      },
      fail(error) {
        reject(error);
      }
    });
  });
}

function uploadWithDirectPut(filePath, preparation) {
  return readFileBuffer(filePath).then((buffer) => {
    return new Promise((resolve, reject) => {
      wx.request({
        url: preparation.uploadUrl,
        method: "PUT",
        data: buffer,
        responseType: "arraybuffer",
        header: preparation.headers || {},
        success(result) {
          if (Number(result.statusCode) >= 200 && Number(result.statusCode) < 300) {
            resolve();
            return;
          }
          reject(new Error(`图片上传失败，状态码 ${result.statusCode || "--"}`));
        },
        fail(error) {
          reject(error);
        }
      });
    });
  });
}

function uploadPreparedFile(filePath, preparation) {
  const mode = String(preparation.uploadMode || "").trim();
  if (!preparation.storageReady || !preparation.uploadUrl) {
    return Promise.reject(new Error("当前服务端未完成对象存储配置，production 模式不能正式留档上传。"));
  }
  if (mode === "direct_put" || mode === "oss_signed_put") {
    return uploadWithDirectPut(filePath, preparation);
  }
  if (mode === "presigned_post" || mode === "form_post" || Object.keys(preparation.formFields || {}).length) {
    return uploadWithMultipart(filePath, preparation);
  }
  return Promise.reject(new Error(`当前上传模式 ${mode || "unknown"} 尚未接入小程序端。`));
}

function completeRecyclePhotoUpload(orderId, preparation) {
  return request({
    endpoint: appConfig.endpoints.completeRecyclePhotoUpload,
    method: "POST",
    data: {
      uploadId: preparation.uploadId,
      orderId,
      publicUrl: preparation.publicUrl || "",
      thumbnailUrl: preparation.publicUrl || ""
    }
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    return data.publicUrl || data.thumbnailUrl || "";
  });
}

function uploadRecyclePhotos(orderId, photos, profile) {
  const list = normalizePhotos(photos);
  const storeContext = getOnlineStoreContext(profile);
  if (!storeContext.storeId) {
    return Promise.reject(new Error("当前登录账号未同步到真实门店，请先重新登录门店账号。"));
  }
  return list.reduce(function(chain, photo) {
    return chain.then(function(collected) {
      const fileName = getFileName(photo.path, `${photo.name || "photo"}.jpg`);
      return getFileInfo(photo.path)
        .then(function(fileInfo) {
          return requestUploadPreparation(storeContext.storeId, fileName, fileInfo);
        })
        .then(function(preparation) {
          return uploadPreparedFile(photo.path, preparation).then(function() {
            return completeRecyclePhotoUpload(orderId, preparation);
          });
        })
        .then(function(publicUrl) {
          if (!publicUrl) {
            throw new Error("图片留档完成后未返回可访问地址。");
          }
          return collected.concat(publicUrl);
        });
    });
  }, Promise.resolve([]));
}

function pad(value) {
  return String(value).padStart(2, "0");
}

function nowText() {
  const date = new Date();
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function normalizeDateText(value) {
  if (!value) return "";
  if (typeof value === "string") {
    if (value.indexOf("T") > -1) {
      return value.replace("T", " ").slice(0, 16);
    }
    return value.slice(0, 16);
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function todayPrefix() {
  const date = new Date();
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

function createOrderNo() {
  const date = new Date();
  const stamp = `${date.getFullYear()}${pad(date.getMonth() + 1)}${pad(date.getDate())}${pad(date.getHours())}${pad(date.getMinutes())}${pad(date.getSeconds())}`;
  return `GR${stamp}`;
}

function hasOperatorToken() {
  try {
    const profile = wx.getStorageSync("gr_operator_profile") || {};
    return Boolean(profile.token);
  } catch (error) {
    return false;
  }
}

function shouldUseLocalBusinessData() {
  return appConfig.mode === "offline";
}

function getDefaultDraft() {
  return {
    storeId: "",
    storeCode: "",
    customerName: "",
    customerPhone: "",
    sourceChannel: "到店散客",
    operatorName: "",
    storeName: "",
    itemCategory: "足金",
    itemName: "",
    purity: "足金999",
    grossWeight: "",
    deductionWeight: "0",
    recyclePrice: "",
    remark: "",
    photos: [],
    cashierItems: [buildEmptyCashierItem(0)]
  };
}

function normalizeRecycleCategory(category) {
  const value = String(category || "").trim();
  if (value === "金饰" || value === "金条" || value === "旧金料") {
    return "足金";
  }
  return value || "足金";
}

function normalizeDraft(draft) {
  const nextDraft = Object.assign(getDefaultDraft(), draft || {});
  nextDraft.itemCategory = normalizeRecycleCategory(nextDraft.itemCategory);
  nextDraft.photos = normalizePhotos(nextDraft.photos || nextDraft.attachments || nextDraft.attachmentUrls || []);
  nextDraft.cashierItems = normalizeCashierItems(nextDraft.cashierItems);
  return nextDraft;
}

function buildPhotoRecord(item, index) {
  if (typeof item === "string") {
    return {
      id: `photo_${index + 1}_${Date.now()}`,
      path: item,
      name: `现场图 ${index + 1}`
    };
  }
  return {
    id: item.id || `photo_${index + 1}_${Date.now()}`,
    path: item.path || item.url || item.publicUrl || item.thumbnailUrl || item.tempFilePath || "",
    name: item.name || `现场图 ${index + 1}`
  };
}

function normalizePhotos(source) {
  return (source || []).filter(function(item) {
    return !!item;
  }).map(function(item, index) {
    return buildPhotoRecord(item, index);
  }).filter(function(item) {
    return !!item.path;
  });
}

function buildEmptyCashierItem(index) {
  return {
    id: `cashier_item_${Date.now()}_${index + 1}`,
    productId: "",
    sku: "",
    name: "",
    quantity: "1",
    unitPrice: "",
    inventory: "",
    productHint: "",
    productHintTone: ""
  };
}

function normalizeCashierItems(source) {
  const list = Array.isArray(source) ? source : [];
  const normalized = list.map(function(item, index) {
    return {
      id: item.id || `cashier_item_${Date.now()}_${index + 1}`,
      productId: String(item.productId || "").trim(),
      sku: String(item.sku || "").trim(),
      name: String(item.name || "").trim(),
      quantity: String(item.quantity || "1").trim() || "1",
      unitPrice: String(item.unitPrice || "").trim(),
      inventory: String(item.inventory || "").trim(),
      productHint: String(item.productHint || "").trim(),
      productHintTone: String(item.productHintTone || "").trim()
    };
  }).filter(function(item) {
    return item.productId || item.sku || item.name || item.unitPrice || item.quantity !== "1";
  });
  return normalized.length ? normalized : [buildEmptyCashierItem(0)];
}

function calculateCashierSummary(draft) {
  const source = normalizeDraft(draft);
  const items = normalizeCashierItems(source.cashierItems);
  const normalizedItems = items.map(function(item) {
    const quantity = Math.max(parseInt(item.quantity, 10) || 0, 0);
    const unitPrice = safeNumber(item.unitPrice);
    const amount = Number((quantity * unitPrice).toFixed(2));
    return Object.assign({}, item, {
      quantity,
      unitPrice,
      amount,
      amountText: amount.toFixed(2),
      unitPriceText: unitPrice.toFixed(2)
    });
  });
  const itemCount = normalizedItems.reduce(function(sum, item) {
    return sum + item.quantity;
  }, 0);
  const totalAmount = normalizedItems.reduce(function(sum, item) {
    return sum + item.amount;
  }, 0);
  const validItems = normalizedItems.filter(function(item) {
    return item.sku && item.name && item.quantity > 0 && item.unitPrice >= 0;
  });
  return {
    items: normalizedItems,
    itemCount,
    totalAmount,
    totalAmountText: totalAmount.toFixed(2),
    readyForSubmit: Boolean(source.customerName && source.customerPhone && validItems.length > 0 && validItems.length === normalizedItems.length)
  };
}

function getPhotoValidation(photos) {
  const count = normalizePhotos(photos).length;
  const requireExactThree = !!photoRules.requireExactThree;
  if (requireExactThree) {
    return {
      ok: count === 3,
      count,
      minCount: photoRules.minCount,
      maxCount: photoRules.maxCount,
      message: count === 3 ? "照片数量满足确认要求" : "确认回收前必须上传 3 张现场照片"
    };
  }
  if (count < photoRules.minCount) {
    return {
      ok: false,
      count,
      minCount: photoRules.minCount,
      maxCount: photoRules.maxCount,
      message: `至少上传${photoRules.minCount}张现场照片`
    };
  }
  if (count > photoRules.maxCount) {
    return {
      ok: false,
      count,
      minCount: photoRules.minCount,
      maxCount: photoRules.maxCount,
      message: `最多只能上传 ${photoRules.maxCount} 张现场照片`
    };
  }
  return {
    ok: true,
    count,
    minCount: photoRules.minCount,
    maxCount: photoRules.maxCount,
    message: "照片数量满足确认要求"
  };
}

function getPrimaryRecycleItem(source) {
  return source && source.items && source.items.length ? source.items[0] : null;
}

function calculateQuote(draft) {
  const source = normalizeDraft(draft);
  const grossWeight = safeNumber(source.grossWeight);
  const deductionWeight = safeNumber(source.deductionWeight);
  const netWeight = Math.max(grossWeight - deductionWeight, 0);
  const benchPrice = getSuggestedPrice(source.purity);
  const recyclePrice = safeNumber(source.recyclePrice) || benchPrice;
  const serviceFeeRate = source.itemCategory === "金条" ? 1.2 : 2.5;
  const subtotal = Number((netWeight * recyclePrice).toFixed(2));
  const serviceFee = Number((netWeight * serviceFeeRate).toFixed(2));
  const amount = Number(Math.max(subtotal - serviceFee, 0).toFixed(2));

  return {
    benchPrice,
    recyclePrice,
    grossWeight,
    deductionWeight,
    netWeight,
    serviceFeeRate,
    subtotal,
    serviceFee,
    amount,
    grossWeightText: grossWeight.toFixed(2),
    deductionWeightText: deductionWeight.toFixed(2),
    netWeightText: netWeight.toFixed(2),
    recyclePriceText: recyclePrice.toFixed(2),
    serviceFeeText: serviceFee.toFixed(2),
    subtotalText: subtotal.toFixed(2),
    amountText: amount.toFixed(2),
    readyForSubmit: Boolean(source.customerName && source.customerPhone && source.itemName && netWeight > 0)
  };
}

function formatQuoteFromServer(localQuote, serverQuote) {
  const unitPrice = safeNumber(serverQuote.unitPrice) || localQuote.recyclePrice;
  const netWeight = safeNumber(serverQuote.netWeightGram) || localQuote.netWeight;
  const grossWeight = safeNumber(serverQuote.grossWeightGram) || localQuote.grossWeight;
  const deductionWeight = safeNumber(serverQuote.deductionWeightGram) || localQuote.deductionWeight;
  const serviceFeeRate = safeNumber(serverQuote.serviceFeeRate) || localQuote.serviceFeeRate;
  const subtotal = safeNumber(serverQuote.subtotalAmount) || Number((netWeight * unitPrice).toFixed(2));
  const serviceFee = safeNumber(serverQuote.serviceFee) || Number((netWeight * serviceFeeRate).toFixed(2));
  const amount = safeNumber(serverQuote.estimatedAmount) || Number(Math.max(subtotal - serviceFee, 0).toFixed(2));

  return Object.assign({}, localQuote, {
    benchPrice: unitPrice,
    recyclePrice: unitPrice,
    grossWeight,
    deductionWeight,
    netWeight,
    serviceFeeRate,
    subtotal,
    serviceFee,
    amount,
    priceSource: serverQuote.priceSource || "server",
    grossWeightText: grossWeight.toFixed(2),
    deductionWeightText: deductionWeight.toFixed(2),
    netWeightText: netWeight.toFixed(2),
    recyclePriceText: unitPrice.toFixed(2),
    serviceFeeText: serviceFee.toFixed(2),
    subtotalText: subtotal.toFixed(2),
    amountText: amount.toFixed(2)
  });
}

function previewRecycleQuoteOnline(draft, profile) {
  const normalizedDraft = normalizeDraft(draft);
  const localQuote = calculateQuote(normalizedDraft);
  if (shouldUseLocalBusinessData()) {
    return Promise.resolve(localQuote);
  }
  const operatorProfile = profile || wx.getStorageSync("gr_operator_profile") || {};
  const storeContext = getOnlineStoreContext(operatorProfile);
  if (!storeContext.storeId) {
    return Promise.reject(new Error("当前登录账号未同步到真实门店，请先重新登录门店账号。"));
  }
  return request({
    endpoint: appConfig.endpoints.quotePreview,
    method: "POST",
    data: {
      storeId: storeContext.storeId,
      category: normalizedDraft.itemCategory,
      purity: normalizedDraft.purity,
      grossWeightGram: localQuote.grossWeight,
      deductionWeightGram: localQuote.deductionWeight,
      unitPrice: safeNumber(normalizedDraft.recyclePrice)
    }
  }).then((payload) => {
    const serverQuote = payload && payload.data ? payload.data : payload;
    return formatQuoteFromServer(localQuote, serverQuote || {});
  }).catch(function() {
    return Object.assign({}, localQuote, {
      priceSource: "local_reference"
    });
  });
}

function getDraft() {
  return normalizeDraft(getScopedStorageSync(DRAFT_KEY, {}));
}

function saveDraft(draft) {
  const nextDraft = normalizeDraft(draft);
  setScopedStorageSync(DRAFT_KEY, nextDraft);
  return nextDraft;
}

function mergeDraft(partial) {
  return saveDraft(Object.assign({}, getDraft(), partial || {}));
}

function clearDraft() {
  removeScopedStorageSync(DRAFT_KEY);
}

function getOrders() {
  if (shouldUseLocalBusinessData()) {
    seedOrders();
  }
  return getScopedStorageSync(STORAGE_KEY, []);
}

function saveOrders(orders) {
  setScopedStorageSync(STORAGE_KEY, orders);
}

function getCashierOrders() {
  if (shouldUseLocalBusinessData()) {
    seedCashierOrders();
  }
  return getScopedStorageSync(CASHIER_STORAGE_KEY, []);
}

function saveCashierOrders(orders) {
  setScopedStorageSync(CASHIER_STORAGE_KEY, orders);
}

function normalizeCashierStatus(status) {
  if (status === "paid" || status === "completed" || status === "refunded" || status === "cancelled") {
    return status;
  }
  return "pending";
}

function normalizeCustomerName(value) {
  const name = String(value || "").trim();
  if (name === "门店散客" || name === "到店散客") {
    return "";
  }
  return name;
}

function buildOrderRecord(source, overrides) {
  const backendItem = getPrimaryRecycleItem(source);
  const grossWeight = safeNumber(source.grossWeight) || safeNumber(backendItem && backendItem.weightGram);
  const deductionWeight = safeNumber(source.deductionWeight);
  const netWeight = Math.max(grossWeight - deductionWeight, 0);
  const settleAmount = safeNumber(source.confirmedAmount) || safeNumber(source.estimatedAmount);
  const recyclePrice = safeNumber(source.recyclePrice) || (netWeight > 0 && settleAmount > 0 ? Number((settleAmount / netWeight).toFixed(2)) : 0);
  const draft = normalizeDraft(Object.assign({}, source, {
    itemCategory: source.itemCategory || (backendItem && backendItem.category) || "",
    itemName: source.itemName || (backendItem && backendItem.category ? `${backendItem.category}回收` : ""),
    purity: source.purity || (backendItem && backendItem.purity) || "足金999",
    grossWeight: source.grossWeight || (grossWeight > 0 ? String(grossWeight) : ""),
    deductionWeight: source.deductionWeight || (deductionWeight > 0 ? String(deductionWeight) : "0"),
    recyclePrice: source.recyclePrice || (recyclePrice > 0 ? String(recyclePrice) : ""),
    storeName: source.storeName || source.shopName || "",
    photos: source.photos || source.attachments || source.attachmentUrls || []
  }));
  const quote = calculateQuote(draft);
  const photoValidation = getPhotoValidation(draft.photos);
  return Object.assign({}, draft, quote, {
    id: source.id || source.orderNo || createOrderNo(),
    orderNo: source.orderNo || source.id || createOrderNo(),
    bizType: "recycle",
    detailType: "recycle",
    badgeText: "回收单",
    status: source.status || "completed",
    createdAt: normalizeDateText(source.createdAt) || nowText(),
    updatedAt: normalizeDateText(source.updatedAt) || normalizeDateText(source.createdAt) || nowText(),
    photos: normalizePhotos(draft.photos),
    photoCount: normalizePhotos(draft.photos).length,
    photoValidation: photoValidation
  }, overrides || {});
}

function buildCashierOrderRecord(source, overrides) {
  const items = (source.items || []).map(function(item) {
    return {
      productId: String(item.productId || "").trim(),
      sku: String(item.sku || "").trim(),
      name: item.name || "未命名商品",
      quantity: Number(item.quantity || 0),
      unitPrice: safeNumber(item.unitPrice),
      amount: safeNumber(item.amount || (Number(item.quantity || 0) * safeNumber(item.unitPrice)))
    };
  });
  const normalizedStatus = normalizeCashierStatus(source.status);
  const totalAmount = safeNumber(source.totalAmount || source.paidAmount || source.amount);
  const itemNames = items.map(function(item) {
    return item.name;
  }).filter(function(item) {
    return !!item;
  });
  const createdAt = normalizeDateText(source.createdAt) || nowText();
  return Object.assign({
    id: source.id || source.orderNo || createOrderNo(),
    orderNo: source.orderNo || source.id || createOrderNo(),
    bizType: "cashier",
    detailType: "cashier",
    badgeText: "收银单",
    status: normalizedStatus,
    statusText: cashierStatusMap[normalizedStatus] || "处理中",
    customerName: normalizeCustomerName(source.customerName || source.customerLabel),
    customerPhone: String(source.customerPhone || "").trim(),
    itemName: itemNames[0] || "商品收银",
    itemSummary: itemNames.join("、") || "未填写商品明细",
    itemCount: Number(source.itemCount || items.reduce(function(sum, item) {
      return sum + Number(item.quantity || 0);
    }, 0)),
    amount: totalAmount,
    amountText: totalAmount.toFixed(2),
    totalAmount: totalAmount,
    totalAmountText: totalAmount.toFixed(2),
    operatorName: source.createdBy || source.operatorName || "未填写",
    refundReason: source.refundReason || source.voidReason || "",
    refundedBy: source.refundedBy || source.voidedBy || "",
    refundedAt: normalizeDateText(source.refundedAt || source.voidedAt) || "",
    voidReason: source.voidReason || source.refundReason || "",
    voidedBy: source.voidedBy || source.refundedBy || "",
    voidedAt: normalizeDateText(source.voidedAt || source.refundedAt) || "",
    sourceChannel: source.sourceChannel || "门店收银",
    storeName: source.storeName || "",
    remark: source.remark || "",
    createdAt: createdAt,
    updatedAt: normalizeDateText(source.updatedAt) || createdAt,
    items: items,
    photoCount: 0
  }, overrides || {});
}

function refundCashierOrder(id, reason) {
  const normalizedId = String(id || "").trim();
  const normalizedReason = String(reason || "").trim();
  const orders = getCashierOrders();
  const index = orders.findIndex((item) => item.id === normalizedId || item.orderNo === normalizedId);
  if (index < 0) {
    throw new Error("未找到收银单");
  }
  if (orders[index].status === "refunded") {
    throw new Error("该收银单已退单");
  }
  const now = nowText();
  const nextOrder = buildCashierOrderRecord(Object.assign({}, orders[index], {
    status: "refunded",
    refundReason: normalizedReason,
    refundedBy: "当前操作员",
    refundedAt: now,
    voidReason: normalizedReason,
    voidedBy: "当前操作员",
    voidedAt: now,
    updatedAt: now
  }));
  orders[index] = nextOrder;
  saveCashierOrders(orders);
  return nextOrder;
}

function normalizeLegacyStoreName(storeName) {
  if (storeName === "示例门店1") return "本地测试门店1";
  if (storeName === "示例门店2") return "本地测试门店2";
  return storeName;
}

function normalizeLegacyOperatorName(operatorName) {
  return operatorName === "示例店长" ? "本地测试店长" : operatorName;
}

function migrateLegacyOrders(list) {
  return (list || []).map(function(item) {
    return Object.assign({}, item, {
      storeName: normalizeLegacyStoreName(item.storeName),
      operatorName: normalizeLegacyOperatorName(item.operatorName || item.createdBy),
      createdBy: normalizeLegacyOperatorName(item.createdBy)
    });
  });
}

function seedOrders() {
  const existingOrders = getScopedStorageSync(STORAGE_KEY, []);
  if (existingOrders.length) {
    const migratedOrders = migrateLegacyOrders(existingOrders);
    if (JSON.stringify(existingOrders) !== JSON.stringify(migratedOrders)) {
      saveOrders(migratedOrders);
    }
    return;
  }

  const seededOrders = [
    buildOrderRecord({
      customerName: "张女士",
      customerPhone: "13800002026",
      sourceChannel: "到店散客",
      operatorName: "本地测试店长",
      storeName: "本地测试门店1",
      itemCategory: "金饰",
      itemName: "足金手镯",
      purity: "足金999",
      grossWeight: "12.68",
      deductionWeight: "0.20",
      recyclePrice: "742",
      remark: "客户当面复秤确认"
    }, {
      id: "GR20260509001",
      status: "completed",
      createdAt: "2026-05-09 10:18",
      updatedAt: "2026-05-09 10:25"
    }),
    buildOrderRecord({
      customerName: "陈先生",
      customerPhone: "13900008866",
      sourceChannel: "企业回访",
      operatorName: "本地测试店长",
      storeName: "本地测试门店1",
      itemCategory: "K金",
      itemName: "18K 项链",
      purity: "18K",
      grossWeight: "8.92",
      deductionWeight: "0.15",
      recyclePrice: "558",
      remark: "等待财务打款"
    }, {
      id: "GR20260510002",
      status: "pending_confirm",
      createdAt: "2026-05-10 09:42",
      updatedAt: "2026-05-10 09:45"
    })
  ];

  saveOrders(seededOrders);
}

function seedCashierOrders() {
  const existingOrders = getScopedStorageSync(CASHIER_STORAGE_KEY, []);
  if (existingOrders.length) {
    const migratedOrders = migrateLegacyOrders(existingOrders);
    if (JSON.stringify(existingOrders) !== JSON.stringify(migratedOrders)) {
      saveCashierOrders(migratedOrders);
    }
    return;
  }

  const seededOrders = [
    buildCashierOrderRecord({
      id: "ORD20260510031",
      orderNo: "ORD20260510031",
      status: "paid",
      totalAmount: 3298,
      createdBy: "李店长",
      createdAt: "2026-05-10 14:18",
      storeName: "本地测试门店1",
      remark: "足金项链到店成交",
      items: [
        { productId: "product-001", sku: "GJG-SZ-001", name: "足金项链", quantity: 1, unitPrice: 3298, amount: 3298 }
      ]
    }),
    buildCashierOrderRecord({
      id: "ORD20260509012",
      orderNo: "ORD20260509012",
      status: "pending",
      totalAmount: 1880,
      createdBy: "张收银",
      createdAt: "2026-05-09 18:06",
      storeName: "本地测试门店1",
      remark: "等待客户完成转账",
      items: [
        { productId: "product-002", sku: "GJG-KG-018", name: "古法耳饰", quantity: 1, unitPrice: 1880, amount: 1880 }
      ]
    })
  ];

  saveCashierOrders(seededOrders);
}

function createRecycleOrder(draft) {
  const validation = getPhotoValidation(draft.photos);
  if (!validation.ok) {
    throw new Error(validation.message);
  }
  const order = buildOrderRecord(draft);
  const orders = [order].concat(getOrders());
  saveOrders(orders);
  clearDraft();
  return order;
}

function createCashierOrder(draft) {
  const source = normalizeDraft(draft);
  const summary = calculateCashierSummary(source);
  if (!summary.readyForSubmit) {
    throw new Error("请补齐客户信息、商品编码和至少一条有效商品明细");
  }
  const order = buildCashierOrderRecord({
    customerName: source.customerName,
    customerPhone: source.customerPhone,
    remark: source.remark,
    storeName: source.storeName || "",
    operatorName: source.operatorName || "未填写",
    sourceChannel: source.sourceChannel,
    items: summary.items
  });
  const orders = [order].concat(getCashierOrders());
  saveCashierOrders(orders);
  clearDraft();
  return order;
}

function createRecycleOrderOnline(draft) {
  if (shouldUseLocalBusinessData()) {
    try {
      return Promise.resolve(createRecycleOrder(draft));
    } catch (error) {
      return Promise.reject(error);
    }
  }

  const normalizedDraft = normalizeDraft(draft);
  const validation = getPhotoValidation(normalizedDraft.photos);
  const profile = wx.getStorageSync("gr_operator_profile") || {};
  const storeContext = getOnlineStoreContext(profile);
  if (!validation.ok) {
    return Promise.reject(new Error(validation.message));
  }
  if (!storeContext.storeId) {
    return Promise.reject(new Error("当前登录账号未同步到真实门店，请先重新登录门店账号。"));
  }

  return previewRecycleQuoteOnline(normalizedDraft, profile).then((quote) => {
    return request({
      endpoint: appConfig.endpoints.createRecycleOrder,
      method: "POST",
      data: {
        storeId: storeContext.storeId,
        customerName: normalizedDraft.customerName,
        customerPhone: normalizedDraft.customerPhone,
        estimatedAmount: quote.amount,
        items: [{
          category: normalizedDraft.itemCategory,
          purity: normalizedDraft.purity,
          weightGram: quote.netWeight
        }],
        attachmentUrls: [],
        remark: normalizedDraft.remark
      }
    }).then((payload) => {
      const createdDraft = payload && payload.data ? payload.data : payload;
      const createdOrderId = createdDraft.id || createdDraft.orderNo || createOrderNo();
      return uploadRecyclePhotos(createdOrderId, normalizedDraft.photos, profile).then((attachmentUrls) => {
        return request({
          endpoint: appConfig.endpoints.confirmRecycleOrder,
          method: "POST",
          params: { id: createdOrderId },
          data: {
            confirmedAmount: quote.amount,
            attachmentUrls,
            remark: normalizedDraft.remark
          }
        });
      }).then((confirmPayload) => {
        const data = confirmPayload && confirmPayload.data ? confirmPayload.data : confirmPayload;
        const nextOrder = buildOrderRecord(Object.assign({}, normalizedDraft, data, {
          recyclePrice: String(quote.recyclePrice),
          grossWeight: String(quote.grossWeight),
          deductionWeight: String(quote.deductionWeight),
          estimatedAmount: quote.amount,
          confirmedAmount: quote.amount
        }), {
          id: data.id || data.orderNo || createdOrderId,
          status: data.status || "confirmed",
          createdAt: data.createdAt || nowText(),
          updatedAt: data.confirmedAt || data.updatedAt || nowText()
        });
        const orders = [nextOrder].concat(getOrders().filter((item) => item.id !== nextOrder.id));
        saveOrders(orders);
        clearDraft();
        return nextOrder;
      }).catch((error) => {
        const message = error && error.message ? error.message : "正式留档失败";
        throw new Error(`草稿单已创建，但正式留档失败：${message}`);
      });
    });
  });
}

function createCashierOrderOnline(draft) {
  if (shouldUseLocalBusinessData()) {
    try {
      return Promise.resolve(createCashierOrder(draft));
    } catch (error) {
      return Promise.reject(error);
    }
  }

  const source = normalizeDraft(draft);
  const summary = calculateCashierSummary(source);
  const profile = wx.getStorageSync("gr_operator_profile") || {};
  const storeContext = {
    storeId: String(source.storeId || profile.storeId || "").trim(),
    storeName: String(source.storeName || profile.storeName || "").trim(),
    storeCode: String(source.storeCode || profile.storeCode || "").trim()
  };
  if (!summary.readyForSubmit) {
    return Promise.reject(new Error("请补齐客户信息、商品编码和至少一条有效商品明细"));
  }
  if (!storeContext.storeId) {
    return Promise.reject(new Error("当前登录账号未同步到真实门店，请先重新登录门店账号。"));
  }

  return request({
    endpoint: appConfig.endpoints.createCashierOrder,
    method: "POST",
    data: {
      storeId: storeContext.storeId,
      customerName: source.customerName,
      customerPhone: source.customerPhone,
      remark: source.remark,
      items: summary.items.map(function(item) {
        return {
          productId: item.productId,
          sku: item.sku,
          name: item.name,
          quantity: item.quantity,
          unitPrice: item.unitPrice
        };
      })
    }
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const nextOrder = buildCashierOrderRecord(Object.assign({}, data, {
      customerName: source.customerName,
      customerPhone: source.customerPhone,
      sourceChannel: source.sourceChannel,
      storeName: storeContext.storeName || source.storeName
    }));
    const orders = [nextOrder].concat(getCashierOrders().filter(function(item) {
      return item.id !== nextOrder.id;
    }));
    saveCashierOrders(orders);
    clearDraft();
    return nextOrder;
  });
}

function listOrdersOnline() {
  if (shouldUseLocalBusinessData()) {
    return Promise.resolve(getOrders());
  }

  return request({
    endpoint: appConfig.endpoints.listRecycleOrders
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const list = Array.isArray(data) ? data : (data.items || []);
    const orders = list.map(function(item) {
      return buildOrderRecord(item, {
      id: item.id || item.orderNo || createOrderNo(),
      status: item.status || "pending_confirm",
      createdAt: item.createdAt || nowText(),
      updatedAt: item.updatedAt || nowText()
      });
    });
    saveOrders(orders);
    return orders;
  });
}

function listCashierOrdersOnline() {
  if (shouldUseLocalBusinessData()) {
    return Promise.resolve(getCashierOrders());
  }

  return request({
    endpoint: appConfig.endpoints.listCashierOrders
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const list = Array.isArray(data) ? data : (data.items || []);
    const orders = list.map(function(item) {
      return buildCashierOrderRecord(item);
    });
    saveCashierOrders(orders);
    return orders;
  });
}

function getOrderById(id) {
  return getOrders().find((item) => item.id === id) || null;
}

function getCashierOrderById(id) {
  return getCashierOrders().find((item) => item.id === id) || null;
}

function getOrderByIdOnline(id) {
  if (shouldUseLocalBusinessData()) {
    return Promise.resolve(getOrderById(id));
  }

  return request({
    endpoint: appConfig.endpoints.recycleOrderDetail,
    params: { id }
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const order = buildOrderRecord(data, {
      id: data.id || data.orderNo || id,
      status: data.status || "pending_confirm",
      createdAt: data.createdAt || nowText(),
      updatedAt: data.updatedAt || nowText()
    });
    const orders = [order].concat(getOrders().filter((item) => item.id !== order.id));
    saveOrders(orders);
    return order;
  });
}

function getCashierOrderByIdOnline(id) {
  if (shouldUseLocalBusinessData()) {
    return Promise.resolve(getCashierOrderById(id));
  }

  return request({
    endpoint: appConfig.endpoints.cashierOrderDetail,
    params: { id }
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const order = buildCashierOrderRecord(data, {
      id: data.id || data.orderNo || id,
      orderNo: data.orderNo || data.id || id
    });
    const orders = [order].concat(getCashierOrders().filter((item) => item.id !== order.id));
    saveCashierOrders(orders);
    return order;
  });
}

function refundCashierOrderOnline(id, reason) {
  if (shouldUseLocalBusinessData()) {
    try {
      return Promise.resolve(refundCashierOrder(id, reason));
    } catch (error) {
      return Promise.reject(error);
    }
  }

  return request({
    endpoint: appConfig.endpoints.refundCashierOrder,
    method: "POST",
    params: { id },
    data: { reason }
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const order = buildCashierOrderRecord(data, {
      id: data.id || data.orderNo || id,
      orderNo: data.orderNo || data.id || id
    });
    const orders = [order].concat(getCashierOrders().filter((item) => item.id !== order.id));
    saveCashierOrders(orders);
    return order;
  });
}

function getOrderAmount(order) {
  if (order && (order.status === "refunded" || order.status === "cancelled")) {
    return 0;
  }
  return safeNumber(order.amount || order.totalAmount || order.estimatedAmount || order.confirmedAmount);
}

function isEffectiveDashboardOrder(order) {
  return order && order.status !== "refunded" && order.status !== "cancelled";
}

function getScopedDashboardOrders(orders, profile) {
  if (!profile || profile.roleKey === "owner" || profile.roleKey === "boss") {
    return orders;
  }
  const storeName = profile.storeName || "";
  const storeCode = profile.storeCode || "";
  return orders.filter(function(item) {
    return !storeName && !storeCode
      ? true
      : item.storeName === storeName || item.storeCode === storeCode;
  });
}

function buildStoreBreakdown(orders) {
  const map = {};
  orders.forEach(function(item) {
    const storeName = item.storeName || "未设置门店";
    if (!map[storeName]) {
      map[storeName] = {
        storeName,
        orderCount: 0,
        amount: 0
      };
    }
    map[storeName].orderCount += 1;
    map[storeName].amount += getOrderAmount(item);
  });
  return Object.keys(map).map(function(key) {
    return Object.assign({}, map[key], {
      amountText: map[key].amount.toFixed(2)
    });
  }).sort(function(a, b) {
    return b.amount - a.amount;
  });
}

function getDashboardStats(profile) {
  const recycleOrders = getOrders();
  const cashierOrders = getCashierOrders();
  const scopedRecycleOrders = getScopedDashboardOrders(recycleOrders, profile);
  const scopedCashierOrders = getScopedDashboardOrders(cashierOrders, profile);
  const orders = scopedRecycleOrders.concat(scopedCashierOrders);
  const effectiveOrders = orders.filter(isEffectiveDashboardOrder);
  const today = todayPrefix();
  const todayOrders = effectiveOrders.filter((item) => (item.createdAt || "").indexOf(today) === 0);
  const todayRecycleOrders = scopedRecycleOrders.filter((item) => isEffectiveDashboardOrder(item) && (item.createdAt || "").indexOf(today) === 0);
  const todayCashierOrders = scopedCashierOrders.filter((item) => isEffectiveDashboardOrder(item) && (item.createdAt || "").indexOf(today) === 0);
  const completedOrders = orders.filter((item) => item.status === "completed" || item.status === "confirmed" || item.status === "paid");
  const pendingOrders = orders.filter((item) => item.status === "pending_confirm" || item.status === "pending" || item.status === "draft");
  const storeBreakdown = buildStoreBreakdown(orders);

  return {
    todayCount: todayOrders.length,
    todayAmount: todayOrders.reduce((sum, item) => sum + getOrderAmount(item), 0).toFixed(2),
    todayCashierAmount: todayCashierOrders.reduce((sum, item) => sum + getOrderAmount(item), 0).toFixed(2),
    todayRecycleAmount: todayRecycleOrders.reduce((sum, item) => sum + getOrderAmount(item), 0).toFixed(2),
    pendingCount: pendingOrders.length,
    completedCount: completedOrders.length,
    scopeText: profile && (profile.roleKey === "owner" || profile.roleKey === "boss") ? "全部门店" : ((profile && profile.storeName) || "未绑定门店"),
    visibleStoreCount: storeBreakdown.length,
    storeBreakdown: storeBreakdown.slice(0, 3)
  };
}

function getDashboardStatsOnline(profile) {
  if (shouldUseLocalBusinessData()) {
    return Promise.resolve(getDashboardStats(profile));
  }

  return Promise.all([
    request({
      endpoint: appConfig.endpoints.dashboardSummary
    }).then((payload) => {
      return payload && payload.data ? payload.data : payload;
    }),
    listOrdersOnline(),
    listCashierOrdersOnline()
  ]).then(function(results) {
    const summary = results[0] || {};
    const recycleOrders = results[1] || [];
    const cashierOrders = results[2] || [];
    const scopedRecycleOrders = getScopedDashboardOrders(recycleOrders, profile);
    const scopedCashierOrders = getScopedDashboardOrders(cashierOrders, profile);
    const allOrders = scopedRecycleOrders.concat(scopedCashierOrders);
    const effectiveOrders = allOrders.filter(isEffectiveDashboardOrder);
    const today = todayPrefix();
    const todayOrders = effectiveOrders.filter(function(item) {
      return String(item.createdAt || "").indexOf(today) === 0;
    });
    const todayRecycleOrders = scopedRecycleOrders.filter(function(item) {
      return isEffectiveDashboardOrder(item) && String(item.createdAt || "").indexOf(today) === 0;
    });
    const todayCashierOrders = scopedCashierOrders.filter(function(item) {
      return isEffectiveDashboardOrder(item) && String(item.createdAt || "").indexOf(today) === 0;
    });
    const completedOrders = allOrders.filter(function(item) {
      return item.status === "completed" || item.status === "confirmed" || item.status === "paid";
    });
    const pendingOrders = allOrders.filter(function(item) {
      return item.status === "pending_confirm" || item.status === "pending" || item.status === "draft";
    });
    const storeBreakdown = buildStoreBreakdown(allOrders);

    return {
      todayCount: todayOrders.length,
      todayAmount: todayOrders.reduce(function(sum, item) {
        return sum + getOrderAmount(item);
      }, 0).toFixed(2),
      todayCashierAmount: todayCashierOrders.reduce(function(sum, item) {
        return sum + getOrderAmount(item);
      }, 0).toFixed(2),
      todayRecycleAmount: todayRecycleOrders.reduce(function(sum, item) {
        return sum + getOrderAmount(item);
      }, 0).toFixed(2),
      pendingCount: typeof summary.draftRecycleOrderCount === "number"
        ? summary.draftRecycleOrderCount
        : pendingOrders.length,
      completedCount: completedOrders.length,
      scopeText: profile && (profile.roleKey === "owner" || profile.roleKey === "boss") ? "全部门店" : ((profile && profile.storeName) || "未绑定门店"),
      visibleStoreCount: typeof summary.visibleStoreCount === "number"
        ? summary.visibleStoreCount
        : storeBreakdown.length,
      storeBreakdown: storeBreakdown.slice(0, 3)
    };
  });
}

module.exports = {
  statusMap,
  seedOrders,
  getOrders,
  saveOrders,
  getCashierOrders,
  saveCashierOrders,
  getDefaultDraft,
  getDraft,
  saveDraft,
  mergeDraft,
  clearDraft,
  calculateQuote,
  previewRecycleQuoteOnline,
  getSuggestedPrice,
  refreshReferencePrices,
  getPhotoValidation,
  createRecycleOrder,
  createCashierOrder,
  createRecycleOrderOnline,
  createCashierOrderOnline,
  listOrdersOnline,
  listCashierOrdersOnline,
  getOrderById,
  getCashierOrderById,
  getOrderByIdOnline,
  getCashierOrderByIdOnline,
  refundCashierOrderOnline,
  getDashboardStats,
  getDashboardStatsOnline,
  calculateCashierSummary
};
