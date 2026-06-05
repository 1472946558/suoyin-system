const { appConfig } = require("../../utils/config");

Page({
  data: {
    appName: appConfig.appName,
    contactName: appConfig.contactName,
    servicePhone: appConfig.servicePhone,
    serviceEmail: appConfig.serviceEmail,
    contactAddress: appConfig.contactAddress,
    sections: [
      {
        title: "一、我们收集的信息",
        items: [
          "门店操作员信息：姓名、手机号、岗位角色、门店名称、门店编码和班次。",
          "客户与会员信息：姓名、手机号、会员等级、回访备注、消费记录和回收记录。",
          "订单信息：商品明细、成色、重量、扣减、报价、订单状态和操作时间。",
          "回收照片：用于记录回收物品、复秤、检测和确认回收过程。"
        ]
      },
      {
        title: "二、信息使用目的",
        items: [
          "用于门店收银、黄金回收录单、会员回访、订单查询、售后处理和经营统计。",
          "用于核验操作员所属门店和岗位权限，避免非授权人员查看或处理门店数据。",
          "用于回收业务留档、主管复核、客户争议处理和内部经营审计。"
        ]
      },
      {
        title: "三、照片用途说明",
        items: [
          "回收照片仅用于黄金回收业务留档，不作为公开展示素材。",
          "回收确认前至少需要上传 3 张现场照片，用于证明物品状态、检测过程和双方确认。",
          "未经客户允许，门店不得将回收照片用于与订单处理无关的用途。"
        ]
      },
      {
        title: "四、信息保护与访问范围",
        items: [
          "老板账号可查看全局经营数据，店长默认查看本店数据，员工仅处理岗位授权范围内的业务。",
          "系统会对门店数据进行权限控制，避免不同门店之间非授权查看。",
          "如需修改、删除或查询个人信息，可通过本页联系方式联系管理员处理。"
        ]
      },
      {
        title: "五、联系我们",
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
