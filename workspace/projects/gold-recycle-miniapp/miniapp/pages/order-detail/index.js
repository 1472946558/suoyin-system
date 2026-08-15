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

const {
  getDraft,
  clearDraft,
  calculateQuote,
  createRecycleOrderOnline,
  getOrderByIdOnline,
  getCashierOrderByIdOnline,
  statusMap,
  getPhotoValidation
} = require("../../utils/orderStore");
const { appConfig } = require("../../utils/config");
const { formatRequestError } = require("../../utils/apiClient");
const {
  buildReceiptPrintText,
  buildLabelPrintText,
  printReceiptViaBluetooth,
  printLabelViaBluetooth,
  resetSavedPrinter
} = require("../../utils/bluetoothPrinter");
const { isLoggedIn } = require("../../utils/userStore");

Page({
  data: {
    record: null,
    detailType: "recycle",
    quote: calculateQuote(getDraft()),
    isDraft: true,
    statusText: "",
    heroSubtitle: "",
    loading: true,
    errorText: "",
    photoRule: appConfig.recyclePhotoRules,
    photoValidation: getPhotoValidation(getDraft().photos || []),
    complianceItems: [],
    timelineItems: [],
    printing: false,
    printPreviewTitle: "",
    printPreviewText: "",
    printPendingType: "",
    printProtocol: "",
    printStatusText: ""
  },

  onLoad(query) {
    if (!isLoggedIn()) {
      wx.showToast({ title: "请先登录门店账号", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    wx.setNavigationBarTitle({
      title: query && query.type === "cashier" ? "收银单详情" : "订单详情"
    });
    this.loadRecord(query || {});
  },

  loadRecord(query) {
    this.setData({ loading: true, errorText: "" });

    if (query.id && query.type === "cashier") {
      getCashierOrderByIdOnline(query.id)
        .then((record) => {
          if (!record) {
            this.setData({ errorText: "未找到对应收银单，请返回订单页重试。" });
            return;
          }
          this.applyCashierRecord(record);
        })
        .catch(() => {
          this.setData({ errorText: "收银单加载失败，请检查网络或接口配置。" });
        })
        .finally(() => {
          this.setData({ loading: false });
        });
      return;
    }

    if (query.id) {
      getOrderByIdOnline(query.id)
        .then((record) => {
          if (!record) {
            this.setData({ errorText: "未找到对应回收单，请返回订单页重试。" });
            return;
          }
          this.applyRecycleRecord(record, false);
        })
        .catch(() => {
          this.setData({ errorText: "回收单加载失败，请检查网络或接口配置。" });
        })
        .finally(() => {
          this.setData({ loading: false });
        });
      return;
    }

    const draft = getDraft();
    if (!draft.customerName && !draft.itemName) {
      this.setData({
        loading: false,
        errorText: "当前没有可确认的收银草稿，请先完成收银或回收录单。"
      });
      return;
    }

    this.applyRecycleRecord(draft, true);
    this.setData({ loading: false });
  },

  applyRecycleRecord(record, isDraft) {
    const quote = calculateQuote(record);
    const statusText = isDraft ? "待生成" : (statusMap[record.status] || "处理中");
    const photoValidation = getPhotoValidation(record.photos || []);
    const complianceItems = [
      {
        title: "客户信息",
        status: record.customerName && record.customerPhone ? "ok" : "warn",
        text: record.customerName && record.customerPhone ? "客户资料已补齐" : "客户资料仍不完整"
      },
      {
        title: "照片留痕",
        status: photoValidation.ok ? "ok" : "warn",
        text: photoValidation.message
      },
      {
        title: "收银状态",
        status: isDraft ? "pending" : "ok",
        text: isDraft ? "待生成正式回收单" : `当前状态：${statusText}`
      }
    ];
    const timelineItems = [
      {
        label: isDraft ? "草稿更新时间" : "创建时间",
        value: record.updatedAt || record.createdAt || "刚刚"
      },
      {
        label: "操作员",
        value: record.operatorName || record.createdBy || "未填写"
      },
      {
        label: isDraft ? "下一步" : "门店",
        value: isDraft ? "确认并生成回收单" : (record.storeName || "未填写")
      }
    ];
    this.setData({
      record,
      detailType: "recycle",
      quote,
      isDraft,
      statusText,
      heroSubtitle: isDraft ? "确认无误后生成正式回收单并留档照片" : `回收单号 ${record.id}`,
      photoValidation: photoValidation,
      complianceItems: complianceItems,
      timelineItems: timelineItems
    });
  },

  applyCashierRecord(record) {
    const items = record.items || [];
    const statusText = record.statusText || "处理中";
    const complianceItems = [
      {
        title: "商品明细",
        status: items.length ? "ok" : "warn",
        text: items.length ? `当前共 ${record.itemCount || items.length} 件商品` : "当前收银单缺少商品明细"
      },
      {
        title: "收银状态",
        status: record.status === "paid" || record.status === "completed" || record.status === "refunded" || record.status === "cancelled" ? "ok" : "pending",
        text: `当前状态：${statusText}`
      },
      {
        title: "留痕要求",
        status: "ok",
        text: "收银单默认不要求回收照片，重点检查商品明细与金额。"
      }
    ];
    const timelineItems = [
      {
        label: "创建时间",
        value: record.createdAt || "刚刚"
      },
      {
        label: "操作员",
        value: record.operatorName || "未填写"
      },
      {
        label: "门店",
        value: record.storeName || "未填写"
      }
    ];
    this.setData({
      record,
      detailType: "cashier",
      quote: {
        amountText: record.totalAmountText || record.amountText || "0.00",
        subtotalText: record.totalAmountText || record.amountText || "0.00",
        serviceFeeText: "0.00",
        netWeightText: "--"
      },
      isDraft: false,
      statusText,
      heroSubtitle: `收银单号 ${record.orderNo || record.id}`,
      photoValidation: {
        ok: true,
        count: 0,
        message: "收银单默认不要求回收照片留痕。"
      },
      complianceItems,
      timelineItems
    });
  },

  confirmDraft() {
    if (!this.data.quote.readyForSubmit) {
      wx.showToast({ title: "请补齐客户与回收明细后再确认", icon: "none" });
      return;
    }

    if (!this.data.photoValidation.ok) {
      wx.showToast({ title: this.data.photoValidation.message, icon: "none" });
      return;
    }

    this.setData({ loading: true });
    createRecycleOrderOnline(this.data.record)
      .then((order) => {
        wx.showToast({ title: "回收单已生成", icon: "success" });
        setTimeout(() => {
          wx.redirectTo({ url: `/pages/order-detail/index?id=${order.id}` });
        }, 350);
      })
      .catch((error) => {
        this.setData({ loading: false });
        wx.showToast({ title: formatRequestError(error), icon: "none" });
      });
  },

  editCashier() {
    wx.switchTab({ url: "/pages/order/index" });
  },

  editEntry() {
    wx.switchTab({ url: "/pages/track/index" });
  },

  goOrders() {
    wx.switchTab({ url: "/pages/orders/index" });
  },

  startNew() {
    clearDraft();
    wx.switchTab({ url: "/pages/order/index" });
  },

  previewPhoto(event) {
    const url = event.currentTarget.dataset.url;
    const urls = (this.data.record && this.data.record.photos ? this.data.record.photos : []).map(function(item) {
      return item.path;
    }).filter(function(item) {
      return !!item;
    });
    if (!url || !urls.length) return;
    wx.previewImage({
      current: url,
      urls: urls
    });
  },

  copyOrderNo() {
    const orderNo = this.data.record && (this.data.record.id || this.data.record.orderNo);
    if (!orderNo) return;
    wx.setClipboardData({
      data: String(orderNo)
    });
  },

  getPrintText(printType) {
    if (printType === "label") {
      return buildLabelPrintText(this.data.record, this.data.detailType, this.data.quote);
    }
    return buildReceiptPrintText(this.data.record, this.data.detailType, this.data.quote, this.data.photoValidation);
  },

  runBluetoothPrint(printType) {
    if (!this.data.record || this.data.isDraft) {
      wx.showToast({ title: "请先生成正式单据", icon: "none" });
      return;
    }

    const title = printType === "label" ? "标签打印预览" : "小票打印预览";
    const text = this.getPrintText(printType);
    this.setData({
      printPreviewTitle: title,
      printPreviewText: text,
      printPendingType: printType,
      printProtocol: printType === "label" ? "TSC 标签指令" : "ESC/POS 小票指令",
      printStatusText: "打印任务已生成，请确认预览后发送。"
    });
  },

  printReceipt() {
    this.runBluetoothPrint("receipt");
  },

  printLabel() {
    this.runBluetoothPrint("label");
  },

  sendPrintTask() {
    const text = this.data.printPreviewText;
    if (!text) {
      wx.showToast({ title: "请先生成打印预览", icon: "none" });
      return;
    }
    this.setData({
      printing: true,
      printStatusText: "正在搜索并连接蓝牙打印机"
    });
    const action = this.data.printPendingType === "label"
      ? printLabelViaBluetooth(this.data.record, this.data.detailType, this.data.quote)
      : printReceiptViaBluetooth(text);

    action
      .then(() => {
        this.setData({ printStatusText: "打印任务已发送" });
        wx.showToast({ title: "已发送打印", icon: "success" });
      })
      .catch((error) => {
        wx.setClipboardData({ data: text });
        this.setData({
          printStatusText: "打印任务已生成，等待连接设备"
        });
        wx.showModal({
          title: "打印任务已生成",
          content: (error && error.message)
            ? `${error.message}。打印内容已保留在预览区，并已复制到剪贴板，可在连接打印机后重新发送。`
            : "等待连接设备。打印内容已保留在预览区，并已复制到剪贴板，可在连接打印机后重新发送。",
          showCancel: false
        });
      })
      .finally(() => {
        this.setData({ printing: false });
      });
  },

  resetPrinter() {
    resetSavedPrinter();
    wx.showToast({ title: "已清除打印机", icon: "success" });
  }
});
