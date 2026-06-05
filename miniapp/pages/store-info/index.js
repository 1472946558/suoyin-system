const { appConfig } = require("../../utils/config");
const { getProfile } = require("../../utils/userStore");

function buildBusinessItems(profile) {
  return [
    { title: "小程序名称", value: appConfig.appName, className: "profile-row-app-name" },
    { title: "服务范围", value: appConfig.serviceScope },
    { title: "联系人", value: appConfig.contactName },
    { title: "客服电话", value: appConfig.servicePhone },
    { title: "联系邮箱", value: appConfig.serviceEmail },
    { title: "联系地址", value: appConfig.contactAddress },
    { title: "当前门店", value: profile.storeName || appConfig.storeName },
    { title: "门店编码", value: profile.storeCode || appConfig.defaultStoreCode }
  ];
}

Page({
  data: {
    profile: getProfile(),
    businessItems: []
  },

  onShow() {
    const profile = getProfile();
    this.setData({
      profile,
      businessItems: buildBusinessItems(profile)
    });
  },

  goAccount() {
    wx.navigateTo({ url: "/pages/account/index" });
  }
});
