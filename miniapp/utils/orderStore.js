const STORAGE_KEY = "gr_recycle_orders";
const DRAFT_KEY = "gr_recycle_draft";
const { appConfig, request } = require("./apiClient");
const photoRules = appConfig.recyclePhotoRules;

const statusMap = {
  pending_confirm: "待确认",
  pending_payment: "待打款",
  completed: "已完成",
  draft: "待确认",
  confirmed: "已完成",
  cancelled: "已作废"
};

const purityPriceMap = {
  "足金9999": 748,
  "足金999": 742,
  "22K": 680,
  "18K": 558,
  "14K": 436
};

function safeNumber(value) {
  const parsed = parseFloat(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function pad(value) {
  return String(value).padStart(2, "0");
}

function nowText() {
  const date = new Date();
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

function getDefaultDraft() {
  return {
    customerName: "",
    customerPhone: "",
    sourceChannel: "到店散客",
    paymentMethod: "微信零钱",
    operatorName: "",
    storeName: appConfig.storeName,
    itemCategory: "金饰",
    itemName: "",
    purity: "足金999",
    grossWeight: "",
    deductionWeight: "0",
    recyclePrice: "",
    remark: "",
    photos: []
  };
}

function normalizeDraft(draft) {
  const nextDraft = Object.assign(getDefaultDraft(), draft || {});
  nextDraft.photos = normalizePhotos(nextDraft.photos || nextDraft.attachmentUrls || []);
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
    path: item.path || item.url || item.tempFilePath || "",
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
      message: `至少需要上传 ${photoRules.minCount} 张现场照片`
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

function getSuggestedPrice(purity) {
  return purityPriceMap[purity] || purityPriceMap["足金999"];
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

function getDraft() {
  return normalizeDraft(wx.getStorageSync(DRAFT_KEY) || {});
}

function saveDraft(draft) {
  const nextDraft = normalizeDraft(draft);
  wx.setStorageSync(DRAFT_KEY, nextDraft);
  return nextDraft;
}

function mergeDraft(partial) {
  return saveDraft(Object.assign({}, getDraft(), partial || {}));
}

function clearDraft() {
  wx.removeStorageSync(DRAFT_KEY);
}

function getOrders() {
  return wx.getStorageSync(STORAGE_KEY) || [];
}

function saveOrders(orders) {
  wx.setStorageSync(STORAGE_KEY, orders);
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
    storeName: source.storeName || source.shopName || appConfig.storeName,
    photos: source.photos || source.attachmentUrls || []
  }));
  const quote = calculateQuote(draft);
  const photoValidation = getPhotoValidation(draft.photos);
  return Object.assign({}, draft, quote, {
    id: source.id || source.orderNo || createOrderNo(),
    status: source.status || (draft.paymentMethod === "暂不支付" ? "pending_confirm" : "completed"),
    createdAt: nowText(),
    updatedAt: nowText(),
    photos: normalizePhotos(draft.photos),
    photoCount: normalizePhotos(draft.photos).length,
    photoValidation: photoValidation
  }, overrides || {});
}

function seedOrders() {
  if (getOrders().length) return;

  const seededOrders = [
    buildOrderRecord({
      customerName: "张女士",
      customerPhone: "13800002026",
      sourceChannel: "到店散客",
      paymentMethod: "微信零钱",
      operatorName: "廖店长",
      storeName: appConfig.storeName,
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
      paymentMethod: "银行卡",
      operatorName: "廖店长",
      storeName: appConfig.storeName,
      itemCategory: "K金",
      itemName: "18K 项链",
      purity: "18K",
      grossWeight: "8.92",
      deductionWeight: "0.15",
      recyclePrice: "558",
      remark: "等待财务打款"
    }, {
      id: "GR20260510002",
      status: "pending_payment",
      createdAt: "2026-05-10 09:42",
      updatedAt: "2026-05-10 09:45"
    })
  ];

  saveOrders(seededOrders);
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

function createRecycleOrderOnline(draft) {
  if (appConfig.mode === "mock") {
    try {
      return Promise.resolve(createRecycleOrder(draft));
    } catch (error) {
      return Promise.reject(error);
    }
  }

  const normalizedDraft = normalizeDraft(draft);
  const validation = getPhotoValidation(normalizedDraft.photos);
  const quote = calculateQuote(normalizedDraft);
  const profile = wx.getStorageSync("gr_operator_profile") || {};
  if (!validation.ok) {
    return Promise.reject(new Error(validation.message));
  }

  return request({
    endpoint: appConfig.endpoints.createRecycleOrder,
    method: "POST",
    data: {
      storeId: profile.storeId || appConfig.defaultStoreId,
      customerName: normalizedDraft.customerName,
      customerPhone: normalizedDraft.customerPhone,
      estimatedAmount: quote.amount,
      items: [{
        category: normalizedDraft.itemCategory,
        purity: normalizedDraft.purity,
        weightGram: quote.netWeight
      }],
      attachmentUrls: normalizedDraft.photos.map(function(item) {
        return item.path;
      }),
      remark: normalizedDraft.remark
    }
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const nextOrder = buildOrderRecord(Object.assign({}, normalizedDraft, data), {
      id: data.id || data.orderNo || createOrderNo(),
      status: data.status || "pending_payment",
      createdAt: data.createdAt || nowText(),
      updatedAt: data.updatedAt || nowText()
    });
    const orders = [nextOrder].concat(getOrders().filter((item) => item.id !== nextOrder.id));
    saveOrders(orders);
    clearDraft();
    return nextOrder;
  });
}

function listOrdersOnline() {
  if (appConfig.mode === "mock") {
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
      status: item.status || "pending_payment",
      createdAt: item.createdAt || nowText(),
      updatedAt: item.updatedAt || nowText()
      });
    });
    saveOrders(orders);
    return orders;
  });
}

function getOrderById(id) {
  return getOrders().find((item) => item.id === id) || null;
}

function getOrderByIdOnline(id) {
  if (appConfig.mode === "mock") {
    return Promise.resolve(getOrderById(id));
  }

  return request({
    endpoint: appConfig.endpoints.recycleOrderDetail,
    params: { id }
  }).then((payload) => {
    const data = payload && payload.data ? payload.data : payload;
    const order = buildOrderRecord(data, {
      id: data.id || data.orderNo || id,
      status: data.status || "pending_payment",
      createdAt: data.createdAt || nowText(),
      updatedAt: data.updatedAt || nowText()
    });
    const orders = [order].concat(getOrders().filter((item) => item.id !== order.id));
    saveOrders(orders);
    return order;
  });
}

function getDashboardStats() {
  const orders = getOrders();
  const today = todayPrefix();
  const todayOrders = orders.filter((item) => (item.createdAt || "").indexOf(today) === 0);
  const completedOrders = orders.filter((item) => item.status === "completed");
  const pendingOrders = orders.filter((item) => item.status === "pending_confirm" || item.status === "pending_payment");

  return {
    todayCount: todayOrders.length,
    todayAmount: todayOrders.reduce((sum, item) => sum + safeNumber(item.amount), 0).toFixed(2),
    pendingCount: pendingOrders.length,
    completedCount: completedOrders.length
  };
}

module.exports = {
  statusMap,
  seedOrders,
  getOrders,
  saveOrders,
  getDefaultDraft,
  getDraft,
  saveDraft,
  mergeDraft,
  clearDraft,
  calculateQuote,
  getSuggestedPrice,
  getPhotoValidation,
  createRecycleOrder,
  createRecycleOrderOnline,
  listOrdersOnline,
  getOrderById,
  getOrderByIdOnline,
  getDashboardStats
};
