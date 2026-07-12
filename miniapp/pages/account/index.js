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
  getProfile,
  loginWithPassword,
  loginOnlineWithPhoneCode,
  loginDevtoolsAccount,
  isMiniAppBindingRequired,
  isDevtoolsRuntime,
  logoutOnline,
  getDevtoolsOwnerProfile
} = require("../../utils/userStore");

function formatLoginError(error) {
  if (error && error.type === "WX_LOGIN_ERROR") {
    return "微信身份登录超时，请重新打开小程序后重试。";
  }

  if (isMiniAppBindingRequired(error)) {
    return "当前微信还没有绑定后台账号。首次使用可以授权手机号，系统会按后台账号手机号完成匹配。";
  }

  const message = error && error.message ? error.message : "";
  const responseMessage = error && error.response && error.response.message ? error.response.message : "";
  if (/phone code|手机号|phone/i.test(responseMessage)) {
    return "手机号授权校验失败，请重新点击授权手机号。";
  }
  if (/openid|appsecret|wx\.login|code|invalid request body|开发者|api/i.test(message)) {
    return "当前微信还没有绑定后台账号，请联系管理员完成绑定后重试。";
  }

  if (message) {
    return message;
  }

  return "登录失败，请稍后重试。";
}

function openHome() {
  wx.reLaunch({
    url: "/pages/home/index",
    fail() {
      wx.switchTab({ url: "/pages/home/index" });
    }
  });
}

function trim(value) {
  return String(value || "").trim();
}

function buildPhoneAuthPayload(detail) {
  const phoneCode = trim(detail.code);
  if (phoneCode) {
    return { phoneCode };
  }

  const encryptedData = trim(detail.encryptedData);
  const iv = trim(detail.iv);
  if (encryptedData && iv) {
    return {
      phoneEncryptedData: encryptedData,
      phoneIv: iv
    };
  }

  return null;
}

function formatPhoneAuthorizeMessage(detail) {
  const errMsg = trim(detail.errMsg);
  if (/deny|cancel|refuse|user deny/i.test(errMsg)) {
    return "你刚才没有允许手机号授权，请重新点击并选择手机号。";
  }

  if (/no permission|permission|not support|unsupported|invalid/i.test(errMsg)) {
    return "当前小程序没有返回手机号授权码。请确认已开通手机号能力，并用手机微信扫码预览或体验版测试。";
  }

  if (errMsg && errMsg !== "getPhoneNumber:ok") {
    return `手机号授权未完成：${errMsg}`;
  }

  return "当前调试方式没有返回手机号授权码。请不要用模拟器测试手机号授权，改用手机微信扫码预览或体验版。";
}

Page({
  data: {
    form: getProfile(),
    credentials: {
      username: "",
      password: ""
    },
    submitting: false,
    passwordSubmitting: false,
    passwordButtonText: "登录",
    phoneSubmitting: false,
    phoneButtonText: "手机号快捷登录",
    agreementAccepted: false,
    needsPhoneBinding: false,
    devtoolsMode: isDevtoolsRuntime(),
    loginError: ""
  },

  onShow() {
    const form = getProfile();
    this.setData({
      form: form,
      needsPhoneBinding: false,
      devtoolsMode: isDevtoolsRuntime(),
      loginError: ""
    });
  },

  toggleAgreement() {
    this.setData({
      agreementAccepted: !this.data.agreementAccepted,
      loginError: ""
    });
  },

  ensureAgreement() {
    if (this.data.agreementAccepted) {
      return true;
    }
    this.setData({ loginError: "请先阅读并同意用户协议和隐私协议。" });
    wx.showToast({ title: "请先勾选协议", icon: "none" });
    return false;
  },

  handleCredentialInput(event) {
    const key = event.currentTarget.dataset.key;
    const value = event.detail.value;
    const update = {
      loginError: ""
    };
    update[`credentials.${key}`] = value;
    this.setData(update);
  },

  handlePasswordLogin() {
    if (!this.ensureAgreement()) return;

    const username = trim(this.data.credentials.username);
    const password = this.data.credentials.password;
    if (!username || !password) {
      this.setData({ loginError: "请输入账号和密码。" });
      wx.showToast({ title: "请输入账号密码", icon: "none" });
      return;
    }

    this.setData({
      passwordSubmitting: true,
      passwordButtonText: "正在登录",
      loginError: ""
    });

    loginWithPassword(username, password)
      .then(() => {
        wx.showToast({ title: "登录成功", icon: "success" });
        this.setData({
          form: getProfile(),
          needsPhoneBinding: false,
          credentials: {
            username,
            password: ""
          }
        });
        setTimeout(() => {
          openHome();
        }, 350);
      })
      .catch((error) => {
        this.setData({ loginError: formatLoginError(error) });
        wx.showToast({ title: "登录失败", icon: "none" });
      })
      .finally(() => {
        this.setData({
          passwordSubmitting: false,
          passwordButtonText: "登录"
        });
      });
  },

  handlePhoneAuthorize(event) {
    if (!this.ensureAgreement()) return;

    if (isDevtoolsRuntime()) {
      this.loginDevtoolsDemoOwner();
      return;
    }

    const detail = event && event.detail ? event.detail : {};
    const phoneAuthPayload = buildPhoneAuthPayload(detail);
    if (!phoneAuthPayload) {
      const message = formatPhoneAuthorizeMessage(detail);
      this.setData({ loginError: message });
      wx.showToast({ title: "未拿到授权码", icon: "none" });
      return;
    }

    const form = this.data.form;
    this.setData({
      phoneSubmitting: true,
      phoneButtonText: "正在登录",
      loginError: ""
    });

    loginOnlineWithPhoneCode(form, phoneAuthPayload)
      .then(() => {
        wx.showToast({ title: "绑定成功", icon: "success" });
        this.setData({
          needsPhoneBinding: false,
          form: getProfile()
        });
        setTimeout(() => {
          openHome();
        }, 350);
      })
      .catch((error) => {
        const message = isMiniAppBindingRequired(error)
          ? "授权手机号未匹配到后台账号，请管理员先在后台录入或核对手机号。"
          : formatLoginError(error);
        this.setData({
          loginError: message,
          needsPhoneBinding: isMiniAppBindingRequired(error)
        });
        wx.showToast({ title: "绑定失败", icon: "none" });
      })
      .finally(() => {
        this.setData({
          phoneSubmitting: false,
          phoneButtonText: "手机号快捷登录"
        });
      });
  },

  loginDevtoolsDemoOwner() {
    const form = getDevtoolsOwnerProfile();
    this.setData({
      phoneSubmitting: true,
      phoneButtonText: "正在登录",
      loginError: ""
    });

    loginDevtoolsAccount(form)
      .then(() => {
        wx.showToast({ title: "登录成功", icon: "success" });
        this.setData({
          form: getProfile(),
          needsPhoneBinding: false
        });
        setTimeout(() => {
          openHome();
        }, 350);
      })
      .catch((error) => {
        this.setData({ loginError: formatLoginError(error) });
        wx.showToast({ title: "登录失败", icon: "none" });
      })
      .finally(() => {
        this.setData({
          phoneSubmitting: false,
          phoneButtonText: "手机号快捷登录"
        });
      });
  },

  logout() {
    this.setData({ submitting: true });
    logoutOnline()
      .then(() => {
        wx.showToast({ title: "已退出登录", icon: "none" });
      })
      .finally(() => {
        this.setData({
          submitting: false,
          form: getProfile(),
          needsPhoneBinding: false,
          loginError: "",
          credentials: {
            username: "",
            password: ""
          }
        });
      });
  },

  goHome() {
    openHome();
  },

  openTerms() {
    wx.navigateTo({ url: "/pages/legal/terms" });
  },

  openPrivacy() {
    wx.navigateTo({ url: "/pages/legal/privacy" });
  }
});
