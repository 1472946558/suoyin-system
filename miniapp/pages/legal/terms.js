/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: terms.js
 * 功能描述: 页面模块
 * 作者: 廖心慈
 * 创建日期: 2026-06-03
 */

const { appConfig } = require("../../utils/config");
const { getSystemProfile, loadSystemSettingsOnline } = require("../../utils/systemStore");

Page({
  data: {
    appName: appConfig.appName,
    sections: []
  },

  onShow() {
    this.updateSections(getSystemProfile());
    loadSystemSettingsOnline().then((result) => {
      this.updateSections((result && result.profile) || getSystemProfile());
    });
  },

  updateSections(systemProfile) {
    const profile = systemProfile || getSystemProfile();
    this.setData({
      appName: profile.appName || appConfig.appName,
      sections: [
        {
          title: "一、服务内容",
          items: [
            `${profile.appName || appConfig.appName}为黄金门店提供收银、黄金回收录单、会员管理、商品目录、订单查询和经营统计服务。`,
            "系统用于记录门店业务流程，不替代门店对黄金真伪、成色、重量和价格的人工复核。"
          ]
        },
        {
          title: "二、账号与权限",
          items: [
            "操作员应使用本人门店账号登录，不得转借账号或处理非授权门店业务。",
            "老板账号可查看全局数据，店长账号只查看并处理本店收银、录单和订单查询。"
          ]
        },
        {
          title: "三、订单与回收确认",
          items: [
            "门店应如实录入客户信息、商品明细、成色、重量、扣减和报价。",
            "回收确认前至少上传 3 张现场照片，确保物品状态、检测过程和确认结果可追溯。",
            "系统用于门店业务记录、回收留档和订单查询。"
          ]
        },
        {
          title: "四、打印与外设",
          items: [
            "如需自动打印，建议购买支持蓝牙连接并提供 ESC/POS、TSPL 或 CPCL 指令文档的热敏小票机或标签机。",
            "仅能使用手机 App 打印、且商家明确不支持系统对接的设备，不建议用于自动出单。"
          ]
        },
        {
          title: "五、服务支持",
          items: [
            `联系人：${profile.contactName || "请联系门店管理员"}`,
            `联系电话：${profile.servicePhone || "请联系门店管理员"}`,
            `联系邮箱：${profile.serviceEmail || "请联系门店管理员"}`,
            `联系地址：${profile.contactAddress || "请联系门店管理员"}`,
            profile.receiptTitle ? `小票标题：${profile.receiptTitle}` : "门店品牌与服务信息以后台设置为准"
          ]
        }
      ]
    });
  }
});
