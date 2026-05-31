# 客户服务器接入说明

## 目标

本项目已经预留客户服务器接入层。客户只需要提供 HTTPS 后端服务，并按本文修改配置，即可把当前 mock 前端切换到真实 RESTful API。

在开始改配置前，先填写：

```text
08_go_live/CUSTOMER_LAUNCH_INPUTS.md
```

这份文档用于一次性收齐 AppID、正式 API 域名、HTTPS、协议地址、客服电话、地图 Key、支付和客服入口，避免后端联调做到一半再补基础资料。

## 客户需要准备什么

| 配置项 | 必须 | 说明 |
| --- | --- | --- |
| `apiBaseUrl` | 是 | 客户后端 HTTPS 域名，例如 `https://api.customer.com` |
| `tenantId` | 建议 | 多客户或多城市部署时用于区分租户 |
| 微信小程序 AppID | 是 | 用于正式预览、上传和上线 |
| 服务器域名白名单 | 是 | 在微信公众平台配置 request 合法域名 |
| 腾讯位置服务 Key | 商用必需 | 用于地址解析、路线和距离 |
| 微信支付商户号 | 商用必需 | 用于运费支付、退款 |
| 隐私协议 URL | 商用必需 | 地址、手机号、定位属于敏感信息 |
| 客服入口 | 商用必需 | 用于异常订单、退款、投诉 |

## 交付目录

| 文件/目录 | 说明 |
| --- | --- |
| `app.json` | 页面路由、导航栏、底部 tab 配置 |
| `app.wxss` | 全局视觉变量、按钮、卡片、表单样式 |
| `pages/home` | 首页和价格预估入口 |
| `pages/account` | 登录/注册与测试资料配置 |
| `pages/order` | 下单表单和创建订单 |
| `pages/order-detail` | 订单详情 |
| `pages/track` | 配送跟踪 |
| `pages/orders` | 订单列表 |
| `pages/profile` | 我的页面 |
| `utils/config.js` | 客户服务器、客服电话、协议地址配置 |
| `utils/apiClient.js` | RESTful 请求封装 |
| `utils/orderStore.js` | mock 订单、本地缓存、生产下单适配 |
| `utils/userStore.js` | 本地资料、微信登录、生产登录适配 |

## 前端配置文件

修改文件：

```text
utils/config.js
```

示例：

```js
const appConfig = {
  appName: "客户品牌名",
  brandSlogan: "同城专人急送",
  mode: "production",
  apiBaseUrl: "https://api.customer.com",
  tenantId: "customer-prod",
  requestTimeout: 15000,
  servicePhone: "400-000-0000",
  mapProvider: "tencent",
  privacyUrl: "https://customer.com/privacy",
  userAgreementUrl: "https://customer.com/agreement",
  endpoints: {
    createOrder: "/api/v1/orders",
    listOrders: "/api/v1/orders",
    orderDetail: "/api/v1/orders/:id",
    orderTrack: "/api/v1/orders/:id/track",
    estimatePrice: "/api/v1/pricing/estimate",
    uploadFile: "/api/v1/files",
    login: "/api/v1/auth/wechat-login"
  }
};
```

## 接入步骤

1. 客户部署后端服务，并确认 API 能通过 HTTPS 访问。
2. 在微信公众平台把 `apiBaseUrl` 域名加入 request 合法域名。
3. 修改 `utils/config.js`：把 `mode` 从 `mock` 改成 `production`。
4. 修改 `apiBaseUrl`、`tenantId`、客服电话、隐私协议和用户协议地址。
5. 在微信开发者工具中选择正式 AppID，重新编译。
6. 测试下单、订单列表、订单详情、配送跟踪、异常提示。
7. 通过 QA 后再上传体验版或提交审核。

## 当前已接入生产请求的位置

| 前端动作 | 代码位置 | 生产模式请求 |
| --- | --- | --- |
| 创建订单 | `pages/order/index.js` | `POST /api/v1/orders` |
| 登录/注册 | `pages/account/index.js` | `POST /api/v1/auth/wechat-login` |
| 订单列表 | `pages/orders/index.js` | `GET /api/v1/orders` |
| 订单详情 | `pages/order-detail/index.js` | `GET /api/v1/orders/:id` |
| 配送跟踪 | `pages/track/index.js` | `GET /api/v1/orders/:id/track` |
| 请求封装 | `utils/apiClient.js` | 统一处理 baseUrl、tenantId、超时、HTTP 错误 |
| 配置中心 | `utils/config.js` | 统一维护客户服务器配置 |

## 后端响应格式约定

建议所有接口统一返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

创建订单响应最少需要：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "JF20260509001",
    "status": "created",
    "price": 32,
    "distance": "8.6km",
    "riderName": "等待派单",
    "riderPhone": "待分配",
    "createdAt": "2026-05-09 18:00"
  }
}
```

## 商用验收标准

- `mode=production` 后，下单请求进入客户服务器。
- 后端返回失败时，小程序出现明确失败提示。
- 订单创建后能进入订单详情。
- 订单列表、订单详情和跟踪页能读取同一笔订单。
- 微信开发者工具“问题”面板无项目代码错误。
- 真机测试通过下单、订单、跟踪、我的四条主路径。

## 风险提醒

当前前端已经预留生产 API 封装，但完整商用仍需要客户后端补齐：登录、手机号授权、地址解析、支付、退款、骑手调度、消息通知、客服投诉、隐私协议和风控。
