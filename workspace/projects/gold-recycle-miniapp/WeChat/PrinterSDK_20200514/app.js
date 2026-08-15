/*
 * Copyright (c) 2020-2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: app.js
 * 功能描述: 应用主模块
 * 作者: 廖心慈
 * 创建日期: 2020-03-28
 */

//app.js
App({
  onLaunch: function () {
    this.globalData.sysinfo = wx.getSystemInfoSync()
  },
  getModel: function () { //获取手机型号
    return this.globalData.sysinfo["model"]
  },
  getVersion: function () { //获取微信版本号
    return this.globalData.sysinfo["version"]
  },
  getSystem: function () { //获取操作系统版本
    return this.globalData.sysinfo["system"]
  },
  getPlatform: function () { //获取客户端平台
    return this.globalData.sysinfo["platform"]
  },
  getSDKVersion: function () { //获取客户端基础库版本
    return this.globalData.sysinfo["SDKVersion"]
  },
  globalData: {
    userInfo: null,
    platform:"",
    screenWidth:wx.getSystemInfoSync().screenWidth,
    screenHeight:wx.getSystemInfoSync().screenHeight,
  },
  BLEInformation:{
    platform: "",
    deviceId:"",
    writeCharaterId: "",    
    writeServiceId: "",
    notifyCharaterId: "",
    notifyServiceId: "",
    readCharaterId: "",
    readServiceId: "",
  }
  
})