"use strict";

const { nanoid } = require("nanoid");
const { estimatePrice } = require("./estimate");

const statusTimeline = {
  created: "订单已创建",
  accepted: "骑手已接单",
  picked: "骑手已取件",
  delivering: "骑手配送中",
  completed: "订单已完成",
  cancelled: "订单已取消",
  refunding: "订单退款中"
};

function nowText() {
  const date = new Date();
  const pad = (value) => String(value).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function toMaskedPhone(phone = "13800002026") {
  if (phone.length < 7) return phone;
  return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
}

function createStore() {
  const users = new Map();
  const orders = new Map();
  const tickets = [];

  const seedOrder = {
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
    createdAt: "2026-05-09 10:18",
    track: [
      { status: "created", label: statusTimeline.created, createdAt: "2026-05-09 10:18" },
      { status: "accepted", label: statusTimeline.accepted, createdAt: "2026-05-09 10:22" },
      { status: "picked", label: statusTimeline.picked, createdAt: "2026-05-09 10:35" },
      { status: "delivering", label: statusTimeline.delivering, createdAt: "2026-05-09 10:48" }
    ]
  };
  orders.set(seedOrder.id, seedOrder);

  return {
    users,
    orders,
    tickets,
    listOrders() {
      return Array.from(orders.values()).sort((a, b) => String(b.createdAt).localeCompare(String(a.createdAt)));
    },
    getOrder(id) {
      return orders.get(id) || null;
    },
    saveUser(profile = {}) {
      const id = profile.id || `U${Date.now()}${Math.floor(Math.random() * 10)}`;
      const user = {
        id,
        name: profile.name || "演示用户",
        phone: profile.phone || "",
        wechat: profile.wechat || "",
        qq: profile.qq || "",
        company: profile.company || "",
        city: profile.city || "深圳",
        defaultFromAddress: profile.defaultFromAddress || "",
        defaultToAddress: profile.defaultToAddress || "",
        createdAt: profile.createdAt || new Date().toISOString()
      };
      users.set(id, user);
      return user;
    },
    getUser(userId) {
      return users.get(userId) || null;
    },
    createOrder(payload = {}) {
      const calculated = estimatePrice(payload);
      const id = `JF${Date.now()}${Math.floor(Math.random() * 10)}`;
      const createdAt = nowText();
      const order = {
        id,
        fromAddress: payload.fromAddress || "",
        toAddress: payload.toAddress || "",
        itemType: payload.itemType || "文件证件",
        weight: payload.weight || "1kg以内",
        note: payload.note || "",
        distance: payload.clientEstimate?.distance || calculated.distance,
        price: payload.clientEstimate?.price || calculated.price,
        status: "created",
        riderName: "等待派单",
        riderPhone: "待分配",
        createdAt,
        track: [
          {
            status: "created",
            label: statusTimeline.created,
            createdAt
          }
        ]
      };
      orders.set(id, order);
      return order;
    },
    createTicket(payload = {}) {
      const ticket = {
        id: `TK_${nanoid(10)}`,
        name: payload.name || "",
        phone: payload.phone || "",
        orderId: payload.orderId || "",
        content: payload.content || "",
        createdAt: new Date().toISOString()
      };
      tickets.unshift(ticket);
      return ticket;
    },
    getDefaultAddresses(userId) {
      const user = users.get(userId);
      if (!user) return [];
      return [
        user.defaultFromAddress ? { type: "from", address: user.defaultFromAddress } : null,
        user.defaultToAddress ? { type: "to", address: user.defaultToAddress } : null
      ].filter(Boolean);
    },
    buildTrack(order) {
      return {
        id: order.id,
        status: order.status,
        riderName: order.riderName,
        riderPhone: order.riderPhone,
        fromAddress: order.fromAddress,
        toAddress: order.toAddress,
        timeline: order.track || []
      };
    },
    buildPayment(order) {
      return {
        paymentId: `PAY_${nanoid(12)}`,
        orderId: order.id,
        status: "pending",
        amount: order.price,
        channel: "wechat",
        prepayToken: `mock_prepay_${nanoid(16)}`
      };
    },
    buildProfile(userId) {
      const user = users.get(userId);
      if (!user) return null;
      return {
        ...user,
        displayPhone: toMaskedPhone(user.phone)
      };
    }
  };
}

module.exports = {
  createStore,
  nowText
};
