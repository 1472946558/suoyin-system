# 疾蜂同城急送小程序

原创同城即时配送微信小程序模板源码，使用原生微信小程序技术栈开发，可直接导入微信开发者工具演示和二次改造。

## 技术栈

- WXML
- WXSS
- JavaScript
- 微信小程序原生框架
- RESTful API 封装
- 本地 mock 数据和 `wx.setStorageSync`

## 页面

| 页面 | 路径 | 状态 |
| --- | --- | --- |
| 首页 | `pages/home/index` | 已完成 |
| 下单 | `pages/order/index` | 已完成 |
| 配送跟踪 | `pages/track/index` | 已完成 |
| 订单列表 | `pages/orders/index` | 已完成 |
| 我的 | `pages/profile/index` | 已完成 |
| 订单详情 | `pages/order-detail/index` | 已完成 |
| 登录/注册 | `pages/account/index` | 已完成 |

## 如何打开

1. 打开微信开发者工具。
2. 导入项目目录：`workspace/projects/flash-delivery-miniapp`。
3. AppID 可先使用测试号或 `touristappid`。
4. 编译运行。

## 如何验收

```bash
npm run validate
```

该命令会检查：

- 页面路由。
- tabBar 入口。
- 页面必需文件。
- JSON 格式。
- JavaScript 语法。
- 运行时代码是否误用第三方品牌词。

## 当前能力

- 首页服务展示。
- 地址填写。
- 物品类型和重量选择。
- 模拟价格预估。
- 模拟订单创建。
- 客户服务器配置入口。
- RESTful 下单请求封装。
- 微信登录/注册接口预留。
- 订单列表。
- 订单详情。
- 配送状态跟踪。
- 个人中心和商用入口预留。

## 模板源码定位

当前项目默认按“模板源码交付”使用，不以当前这一份直接上线为目标。

适合用途：

- 作为同城配送小程序模板卖给客户
- 客户自行替换品牌、配色、文案、协议地址
- 先做演示，再决定是否进入定制开发

如果客户后续要正式上线，再单独进入后端、地图、支付、客服、域名和 AppID 的联调阶段。

模板源码模式说明见：`docs/TEMPLATE_SOURCE_MODE.md`。
模板产品总入口见：`docs/TEMPLATE_PRODUCT_INDEX.md`。
商用放行最短检查表见：`docs/COMMERCIAL_GO_NO_GO_CHECKLIST.md`。
模板产品分档说明见：`docs/TEMPLATE_DELIVERY_PACKAGES.md`。
客户快速改造清单见：`docs/TEMPLATE_QUICK_EDIT_CHECKLIST.md`。
客户交付打包清单见：`docs/CUSTOMER_HANDOFF_PACKAGE.md`。

客户接入说明见：`docs/CUSTOMER_INTEGRATION_GUIDE.md`。
后端骨架说明见：`backend/README.md`。

老板自测说明见：`docs/OWNER_TEST_GUIDE.md`。

演示视频录制说明见：`docs/DEMO_RECORDING_GUIDE.md`。

详细交付状态见：`docs/DELIVERY_STATUS_2026-05-09.md`。
当前模板状态见：`docs/CURRENT_STATUS_2026-05-31.md`。

## 品牌合规

本项目不使用“闪送”商标、素材、截图和文案。业务上参考同城急送通用流程，品牌和 UI 均为原创。
