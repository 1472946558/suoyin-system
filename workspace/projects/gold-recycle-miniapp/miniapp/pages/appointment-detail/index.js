/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: index.js
 * 功能描述: 预约详情页
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

var store = require("../../utils/appointmentStore.js");
var userStore = require("../../utils/userStore.js");

Page({
  data: {
    appointmentId: "",
    appt: null,
    actions: [],
    loading: false,
    updating: false
  },

  onLoad: function(options) {
    if (!userStore.isLoggedIn()) {
      wx.reLaunch({ url: "/pages/account/index" });
      return;
    }
    this.setData({ appointmentId: options.id || "" });
    this.loadDetail();
  },

  onShow: function() {
    if (!userStore.isLoggedIn()) return;
    if (this.data.appointmentId && !this.data.appt) {
      this.loadDetail();
    }
  },

  loadDetail: function() {
    var self = this;
    if (!self.data.appointmentId) return;
    self.setData({ loading: true });

    store.getStaffAppointment(self.data.appointmentId)
      .then(function(appt) {
        var formatted = store.formatAppointmentCard(appt);
        var actions = store.STATUS_ACTIONS[appt.status] || [];
        self.setData({
          appt: formatted,
          actions: actions,
          loading: false
        });
      })
      .catch(function(err) {
        self.setData({ loading: false });
        wx.showToast({
          title: store.formatRequestError(err),
          icon: "none"
        });
      });
  },

  onAction: function(e) {
    var self = this;
    var newStatus = e.currentTarget.dataset.status;
    var label = e.currentTarget.dataset.label;

    if (!newStatus || self.data.updating) return;

    wx.showModal({
      title: "确认操作",
      content: "确定要" + label + "吗？",
      success: function(res) {
        if (!res.confirm) return;
        self.doUpdateStatus(newStatus);
      }
    });
  },

  doUpdateStatus: function(newStatus) {
    var self = this;
    self.setData({ updating: true });
    wx.showLoading({ title: "处理中..." });

    store.updateAppointmentStatus(self.data.appointmentId, newStatus)
      .then(function(appt) {
        wx.hideLoading();
        var formatted = store.formatAppointmentCard(appt);
        var actions = store.STATUS_ACTIONS[appt.status] || [];
        self.setData({
          appt: formatted,
          actions: actions,
          updating: false
        });
        wx.showToast({ title: "操作成功", icon: "success" });
      })
      .catch(function(err) {
        wx.hideLoading();
        self.setData({ updating: false });
        wx.showToast({
          title: store.formatRequestError(err),
          icon: "none"
        });
      });
  },

  onCallCustomer: function() {
    var phone = this.data.appt && this.data.appt.customerPhone;
    if (!phone) {
      wx.showToast({ title: "无联系电话", icon: "none" });
      return;
    }
    wx.makePhoneCall({ phoneNumber: phone }).catch(function() {});
  }
});
