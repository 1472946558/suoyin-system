/*
 * Copyright (c) 2020-2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: home.js
 * 功能描述: 页面模块
 * 作者: 廖心慈
 * 创建日期: 2020-04-29
 */

// pages/blueconn/blueconn.js
var app = getApp()
Page({

  /**
   * 页面的初始数据
   */
  data: {
    list: [],
  },
  blueTooth:function(){
    wx.navigateTo({
      url: '../bleConnect/bleConnect',
    })
  },
  //打印指令页面
  printTest: function () {
    wx.navigateTo({
      url: '../sendCommand/sendCommand',
    })
  },
  //小票案例
  recipt: function () {
    wx.navigateTo({
      url: '../receipt/receipt',
    })
  },
  //标签案例
  label: function () {
    wx.navigateTo({
      url: '../label/label',
    })
  },
  //蓝牙配网页面
  blueToothNet: function () {
    wx.navigateTo({
      url: '../net/net',
    })
  },
  //打印彩票
  lotteryTicket: function () {
    wx.navigateTo({
      url: '../ticket/ticket',
    })
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad: function (options) {
    app.BLEInformation.platform = app.getPlatform()
  },

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady: function () {

  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow: function () {

  },

  /**
   * 生命周期函数--监听页面隐藏
   */
  onHide: function () {

  },

  /**
   * 生命周期函数--监听页面卸载
   */
  onUnload: function () {
  wx.closeBLEConnection({
      deviceId: app.BLEInformation.deviceId,
      success: function(res) {
        console.log("关闭蓝牙成功")
      },
    })
  },

  // /**
  //  * 页面相关事件处理函数--监听用户下拉动作
  //  */
  // onPullDownRefresh: function () {
  //   var that = this
  //   wx.startPullDownRefresh({})
  //   that.startSearch()
  // },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom: function () {

  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage: function () {

  }
})