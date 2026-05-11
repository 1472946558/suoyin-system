const { getProfile, loginOnline, clearProfile, getRolePresets } = require("../../utils/userStore");

Page({
  data: {
    form: getProfile(),
    roleOptions: getRolePresets(),
    roleIndex: 2,
    submitting: false,
    submitText: "保存并进入系统"
  },

  onShow() {
    const form = getProfile();
    this.setData({
      form: form,
      roleIndex: this.findRoleIndex(form.roleKey)
    });
  },

  findRoleIndex(roleKey) {
    const index = this.data.roleOptions.findIndex(function(item) {
      return item.key === roleKey;
    });
    return index > -1 ? index : 2;
  },

  onInput(event) {
    const key = event.currentTarget.dataset.key;
    const update = {};
    update[`form.${key}`] = event.detail.value;
    this.setData(update);
  },

  onRoleChange(event) {
    const roleIndex = Number(event.detail.value);
    const roleOption = this.data.roleOptions[roleIndex];
    this.setData({
      roleIndex: roleIndex,
      form: Object.assign({}, this.data.form, {
        roleKey: roleOption.key,
        role: roleOption.label
      })
    });
  },

  submitProfile() {
    const { name, phone, storeName, storeCode } = this.data.form;
    if (!name || !phone || !storeName || !storeCode) {
      wx.showToast({ title: "请补齐操作员与门店信息", icon: "none" });
      return;
    }

    this.setData({
      submitting: true,
      submitText: "正在登录"
    });

    loginOnline(this.data.form)
      .then(() => {
        wx.showToast({ title: "登录成功", icon: "success" });
        setTimeout(() => {
          wx.switchTab({ url: "/pages/home/index" });
        }, 350);
      })
      .catch(() => {
        wx.showToast({ title: "登录失败，请检查配置", icon: "none" });
      })
      .finally(() => {
        this.setData({
          submitting: false,
          submitText: "保存并进入系统"
        });
      });
  },

  fillDemo() {
    this.setData({
      form: {
        id: "",
        name: "廖店长",
        phone: "13800002026",
        roleKey: "manager",
        role: "店长",
        storeName: "华强北体验店",
        storeCode: "SZ-HQB-01",
        shiftName: "早班",
        token: "",
        loggedIn: false,
        createdAt: ""
      },
      roleIndex: this.findRoleIndex("manager")
    });
  },

  fillBossDemo() {
    this.setData({
      form: {
        id: "",
        name: "陈老板",
        phone: "13800008866",
        roleKey: "owner",
        role: "老板",
        storeName: "华强北体验店",
        storeCode: "SZ-HQB-01",
        shiftName: "总部巡店",
        token: "",
        loggedIn: false,
        createdAt: ""
      },
      roleIndex: this.findRoleIndex("owner")
    });
  },

  logout() {
    clearProfile();
    wx.showToast({ title: "已清空登录信息", icon: "none" });
    this.setData({ form: getProfile() });
  }
});
