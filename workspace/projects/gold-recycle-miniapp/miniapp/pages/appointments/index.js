/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: index.js
 * 功能描述: 预约列表页
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

var store = require("../../utils/appointmentStore.js");
var userStore = require("../../utils/userStore.js");
var getProfile = userStore.getProfile;

Page({
  data: {
    activeTab: "",
    tabs: store.STATUS_TABS,
    list: [],
    total: 0,
    loading: false,
    loadingMore: false,
    isEmpty: false,
    profile: getProfile()
  },

  onLoad: function() {
    if (!userStore.isLoggedIn()) {
      wx.reLaunch({ url: "/pages/account/index" });
      return;
    }
    this.loadList();
  },

  onShow: function() {
    if (!userStore.isLoggedIn()) {
      wx.reLaunch({ url: "/pages/account/index" });
      return;
    }
    this.setData({ profile: getProfile() });
    this.loadList();
  },

  onPullDownRefresh: function() {
    this.loadList(function() {
      wx.stopPullDownRefresh();
    });
  },

  onTabChange: function(e) {
    var key = e.currentTarget.dataset.key;
    if (key === this.data.activeTab) return;
    this.setData({ activeTab: key, list: [], isEmpty: false });
    this.loadList();
  },

  loadList: function(cb) {
    var self = this;
    if (self.data.loading) return;
    self.setData({ loading: true });

    store.listStaffAppointments(self.data.activeTab)
      .then(function(resp) {
        var items = (resp.items || []).map(function(appt) {
          return store.formatAppointmentCard(appt);
        });
        self.setData({
          list: items,
          total: resp.total,
          isEmpty: items.length === 0,
          loading: false
        });
      })
      .catch(function(err) {
        self.setData({ loading: false, isEmpty: true });
        wx.showToast({
          title: store.formatRequestError(err),
          icon: "none"
        });
      })
      .then(function() {
        if (typeof cb === "function") cb();
      });
  },

  onTapCard: function(e) {
    var id = e.currentTarget.dataset.id;
    if (!id) return;
    wx.navigateTo({
      url: "/pages/appointment-detail/index?id=" + id
    });
  },

  onCallCustomer: function(e) {
    var phone = e.currentTarget.dataset.phone;
    if (!phone) {
      wx.showToast({ title: "无联系电话", icon: "none" });
      return;
    }
    wx.makePhoneCall({ phoneNumber: phone }).catch(function() {});
  }
});
