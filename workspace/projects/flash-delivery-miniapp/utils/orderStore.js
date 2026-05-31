const STORAGE_KEY = "jf_orders";
const { appConfig, request } = require("./apiClient");

const statusMap = {
  created: "待接单",
  accepted: "骑手已接单",
  picked: "已取件",
  delivering: "配送中",
  completed: "已送达"
};

const statusSteps = ["created", "accepted", "picked", "delivering", "completed"];

function nowText() {
  const date = new Date();
  const pad = (value) => String(value).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function getOrders() {
  return wx.getStorageSync(STORAGE_KEY) || [];
}

function saveOrders(orders) {
  wx.setStorageSync(STORAGE_KEY, orders);
}

function seedOrders() {
  const current = getOrders();
  if (current.length) return;
  saveOrders([
    {
      id: "JF20260509001",
      fromAddress: "南山区 科技园 A 座",
      toAddress: "福田区 购物公园 B 口",
      itemType: "文件证件",
      weight: "1kg以内",
      note: "请当面签收",
      distance: "8.6km",
      price: 32,
      status: "delivering",
      riderName: "李师傅",
      riderPhone: "138****2026",
      createdAt: "2026-05-09 10:18"
    }
  ]);
}

function estimatePrice(form) {
  const weightIndex = ["1kg以内", "1-3kg", "3-5kg", "5kg以上"].indexOf(form.weight);
  const itemIndex = ["文件证件", "数码配件", "鲜花蛋糕", "生鲜食品", "其它物品"].indexOf(form.itemType);
  const base = 18;
  const weightFee = Math.max(weightIndex, 0) * 8;
  const itemFee = itemIndex === 2 || itemIndex === 3 ? 6 : 0;
  const distance = 4.8 + ((form.fromAddress || "").length + (form.toAddress || "").length) % 8;
  return {
    price: base + weightFee + itemFee + Math.round(distance * 1.6),
    distance: `${distance.toFixed(1)}km`
  };
}

function createOrder(form) {
  const estimate = estimatePrice(form);
  const order = {
    id: `JF${Date.now()}`,
    fromAddress: form.fromAddress,
    toAddress: form.toAddress,
    itemType: form.itemType,
    weight: form.weight,
    note: form.note,
    distance: estimate.distance,
    price: estimate.price,
    status: "accepted",
    riderName: "系统派单中",
    riderPhone: "待分配",
    createdAt: nowText()
  };
  const orders = [order].concat(getOrders());
  saveOrders(orders);
  return order;
}

function normalizeServerOrder(payload, form) {
  const data = (payload && payload.data ? payload.data : payload) || {};
  const source = form || {};
  const fallbackEstimate = estimatePrice(source);
  return {
    id: data.id || data.orderNo || `JF${Date.now()}`,
    fromAddress: data.fromAddress || source.fromAddress || "",
    toAddress: data.toAddress || source.toAddress || "",
    itemType: data.itemType || source.itemType || "文件证件",
    weight: data.weight || source.weight || "1kg以内",
    note: data.note || source.note || "",
    distance: data.distance || fallbackEstimate.distance,
    price: typeof data.price === "number" ? data.price : fallbackEstimate.price,
    status: data.status || "created",
    riderName: data.riderName || "等待派单",
    riderPhone: data.riderPhone || "待分配",
    createdAt: data.createdAt || nowText()
  };
}

function normalizeServerOrders(payload) {
  const data = (payload && payload.data ? payload.data : payload) || {};
  const list = Array.isArray(data) ? data : (data.items || data.orders || []);
  return list.map((item) => normalizeServerOrder(item, item));
}

function cacheCreatedOrder(order) {
  const orders = [order].concat(getOrders().filter((item) => item.id !== order.id));
  saveOrders(orders);
  return order;
}

function createOrderOnline(form) {
  if (appConfig.mode === "mock") {
    return Promise.resolve(createOrder(form));
  }

  const clientEstimate = estimatePrice(form);
  return request({
    endpoint: appConfig.endpoints.createOrder,
    method: "POST",
    data: {
      fromAddress: form.fromAddress,
      toAddress: form.toAddress,
      itemType: form.itemType,
      weight: form.weight,
      note: form.note,
      clientEstimate
    }
  }).then((payload) => cacheCreatedOrder(normalizeServerOrder(payload, form)));
}

function listOrdersOnline() {
  if (appConfig.mode === "mock") {
    return Promise.resolve(getOrders());
  }

  return request({
    endpoint: appConfig.endpoints.listOrders
  }).then((payload) => {
    const orders = normalizeServerOrders(payload);
    saveOrders(orders);
    return orders;
  });
}

function getOrderById(id) {
  return getOrders().find((order) => order.id === id) || null;
}

function getOrderByIdOnline(id) {
  if (appConfig.mode === "mock") {
    return Promise.resolve(getOrderById(id));
  }

  return request({
    endpoint: appConfig.endpoints.orderDetail,
    params: { id }
  }).then((payload) => cacheCreatedOrder(normalizeServerOrder(payload, getOrderById(id) || {})));
}

function getLatestOrder() {
  const orders = getOrders();
  return orders[0] || null;
}

function getLatestOrderOnline() {
  if (appConfig.mode === "mock") {
    return Promise.resolve(getLatestOrder());
  }

  return listOrdersOnline().then((orders) => orders[0] || null);
}

function getOrderTrackOnline(order) {
  if (!order || appConfig.mode === "mock") {
    return Promise.resolve(order);
  }

  return request({
    endpoint: appConfig.endpoints.orderTrack,
    params: { id: order.id }
  }).then((payload) => {
    const data = (payload && payload.data ? payload.data : payload) || {};
    const trackOrder = data.order || data;
    return cacheCreatedOrder(normalizeServerOrder(Object.assign({}, order, trackOrder), order));
  });
}

module.exports = {
  statusMap,
  statusSteps,
  seedOrders,
  getOrders,
  saveOrders,
  estimatePrice,
  createOrder,
  createOrderOnline,
  listOrdersOnline,
  getOrderById,
  getOrderByIdOnline,
  getLatestOrder,
  getLatestOrderOnline,
  getOrderTrackOnline
};
