# 黄金收银小程序

原生微信小程序技术栈的“黄金回收收银系统”首版产品工程。

## 当前范围

- 登录
- 首页
- 收银
- 回收录单
- 回收确认
- 订单
- 会员管理
- 商品目录
- 设置

## 技术栈

- WXML
- WXSS
- JavaScript
- 微信小程序原生框架
- `wx.setStorageSync` 本地草稿与缓存

## 页面路径

| 页面 | 路径 |
| --- | --- |
| 登录 | `pages/account/index` |
| 首页 | `pages/home/index` |
| 收银 | `pages/order/index` |
| 回收录单 | `pages/track/index` |
| 回收确认 | `pages/order-detail/index` |
| 订单 | `pages/orders/index` |
| 会员管理 | `pages/members/index` |
| 商品目录 | `pages/products/index` |
| 设置 | `pages/profile/index` |
| 门店资料 | `pages/store-info/index` |

## 本地验收

```bash
npm run validate
```

## 真联调前仍缺

- 微信真实登录与门店员工权限
- 总部金价 / 报价接口
- 电子秤、影像留档、小票打印
- 订单后端与风控审核流程
