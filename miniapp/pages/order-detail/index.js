const {
  getDraft,
  clearDraft,
  calculateQuote,
  createRecycleOrderOnline,
  getOrderByIdOnline,
  statusMap,
  getPhotoValidation
} = require("../../utils/orderStore");
const { appConfig } = require("../../utils/config");

Page({
  data: {
    record: null,
    quote: calculateQuote(getDraft()),
    isDraft: true,
    statusText: "",
    heroSubtitle: "",
    loading: true,
    errorText: "",
    photoRule: appConfig.recyclePhotoRules,
    photoValidation: getPhotoValidation(getDraft().photos || []),
    complianceItems: [],
    timelineItems: []
  },

  onLoad(query) {
    this.loadRecord(query || {});
  },

  loadRecord(query) {
    this.setData({ loading: true, errorText: "" });

    if (query.id) {
      getOrderByIdOnline(query.id)
        .then((record) => {
          if (!record) {
            this.setData({ errorText: "未找到对应回收单，请返回订单页重试。" });
            return;
          }
          this.applyRecord(record, false);
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

    this.applyRecord(draft, true);
    this.setData({ loading: false });
  },

  applyRecord(record, isDraft) {
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
      quote,
      isDraft,
      statusText,
      heroSubtitle: isDraft ? "确认无误后生成本地 mock 回收单" : `回收单号 ${record.id}`,
      photoValidation: photoValidation,
      complianceItems: complianceItems,
      timelineItems: timelineItems
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
        wx.showToast({ title: error && error.message ? error.message : "生成失败，请检查配置", icon: "none" });
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
  }
});
