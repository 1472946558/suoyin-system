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
 * 创建日期: 2026-06-04
 */

const { getProfile } = require("../../utils/userStore");
const { getFallbackSystemProfile, getSystemProfile, loadSystemSettingsOnline } = require("../../utils/systemStore");

function buildBusinessItems(profile, systemProfile) {
  const resolvedProfile = systemProfile || getFallbackSystemProfile();
  return [
    { title: "小程序主体名称", value: resolvedProfile.subjectName || resolvedProfile.appName || "--", className: "profile-row-app-name" },
    { title: "小程序名称", value: resolvedProfile.appName || "--", className: "profile-row-app-name" },
    { title: "服务范围", value: resolvedProfile.serviceScope || "--" },
    { title: "联系人", value: resolvedProfile.contactName || "请联系管理员在后台补充" },
    { title: "联系电话", value: resolvedProfile.servicePhone || "请联系管理员在后台补充" },
    { title: "联系邮箱", value: resolvedProfile.serviceEmail || "请联系管理员在后台补充" },
    { title: "联系地址", value: resolvedProfile.contactAddress || "请联系管理员在后台补充", className: "profile-row-address" },
    { title: "当前门店", value: profile.storeName || "请先登录真实门店账号" },
    { title: "门店编码", value: profile.storeCode || "待同步" }
  ].filter(function(item) {
    return item.value !== "";
  });
}

Page({
  data: {
    profile: getProfile(),
    businessItems: [],
    systemProfile: getSystemProfile()
  },

  onShow() {
    const profile = getProfile();
    this.setData({
      profile,
      systemProfile: getSystemProfile(),
      businessItems: buildBusinessItems(profile, getSystemProfile())
    });
    loadSystemSettingsOnline().then((result) => {
      const nextProfile = (result && result.profile) || getSystemProfile();
      this.setData({
        systemProfile: nextProfile,
        businessItems: buildBusinessItems(getProfile(), nextProfile)
      });
    });
  },

  goAccount() {
    wx.navigateTo({ url: "/pages/account/index" });
  }
});
