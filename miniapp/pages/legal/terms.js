const { appConfig } = require("../../utils/config");

Page({
  data: {
    appName: appConfig.appName,
    sections: [
      {
        title: "一、服务内容",
        items: [
          `${appConfig.appName}为黄金门店提供收银、黄金回收录单、会员管理、商品目录、订单查询和经营统计服务。`,
          "系统用于记录门店业务流程，不替代门店对黄金真伪、成色、重量和价格的人工复核。"
        ]
      },
      {
        title: "二、账号与权限",
        items: [
          "操作员应使用本人门店账号登录，不得转借账号或处理非授权门店业务。",
          "老板账号可查看全局数据，店长账号查看本店数据，员工账号按岗位处理收银、录单和订单查询。"
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
          `联系人：${appConfig.contactName}`,
          `联系电话：${appConfig.servicePhone}`,
          `联系邮箱：${appConfig.serviceEmail}`,
          `联系地址：${appConfig.contactAddress}`
        ]
      }
    ]
  }
});
