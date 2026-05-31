# 后端接口规划

## 说明

当前版本使用本地 mock 数据。正式商用时建议接入 REST API。

## 基础约定

```text
Base URL: https://api.your-domain.com
Auth: Bearer <token>
Content-Type: application/json
```

## 接口清单

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/v1/pricing/estimate` | 价格预估 |
| `POST` | `/api/v1/auth/wechat-login` | 微信登录/注册 |
| `POST` | `/api/v1/orders` | 创建订单 |
| `GET` | `/api/v1/orders` | 订单列表 |
| `GET` | `/api/v1/orders/{id}` | 订单详情 |
| `GET` | `/api/v1/orders/{id}/track` | 配送跟踪 |
| `POST` | `/api/v1/payments/wechat` | 微信支付下单 |
| `GET` | `/api/v1/profile` | 用户资料 |
| `GET` | `/api/v1/addresses` | 常用地址 |
| `POST` | `/api/v1/support/tickets` | 客服工单 |

## 微信登录/注册

请求：

```json
{
  "code": "wx.login 返回的临时 code",
  "profile": {
    "name": "廖先生",
    "phone": "13800002026",
    "wechat": "liaoliao-demo",
    "qq": "1002026",
    "company": "疾蜂急送测试客户",
    "city": "深圳",
    "defaultFromAddress": "南山区 科技园 A 座",
    "defaultToAddress": "福田区 购物公园 B 口"
  }
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "userId": "U20260509001",
    "token": "server-issued-token"
  }
}
```

## 创建订单

请求：

```json
{
  "fromAddress": "南山区 科技园 A 座",
  "toAddress": "福田区 购物公园 B 口",
  "itemType": "文件证件",
  "weight": "1kg以内",
  "note": "请当面签收",
  "clientEstimate": {
    "price": 32,
    "distance": "8.6km"
  }
}
```

## 订单列表

请求：

```text
GET /api/v1/orders
Authorization: Bearer <token>
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "JF20260509001",
        "fromAddress": "南山区 科技园 A 座",
        "toAddress": "福田区 购物公园 B 口",
        "itemType": "文件证件",
        "weight": "1kg以内",
        "note": "请当面签收",
        "distance": "8.6km",
        "price": 32,
        "status": "accepted",
        "riderName": "李师傅",
        "riderPhone": "138****2026",
        "createdAt": "2026-05-09 18:00"
      }
    ]
  }
}
```

## 订单详情

请求：

```text
GET /api/v1/orders/{id}
Authorization: Bearer <token>
```

响应结构同订单列表里的单条订单对象。

## 配送跟踪

请求：

```text
GET /api/v1/orders/{id}/track
Authorization: Bearer <token>
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "JF20260509001",
    "status": "accepted",
    "riderName": "李师傅",
    "riderPhone": "138****2026",
    "fromAddress": "南山区 科技园 A 座",
    "toAddress": "福田区 购物公园 B 口"
  }
}
```

响应：

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

## 订单状态

| 状态 | 说明 |
| --- | --- |
| `created` | 已创建 |
| `accepted` | 骑手已接单 |
| `picked` | 已取件 |
| `delivering` | 配送中 |
| `completed` | 已完成 |
| `cancelled` | 已取消 |
| `refunding` | 退款中 |

## 商用后端必须补齐

- 用户登录和手机号授权。
- 地址解析和经纬度。
- 距离与价格规则。
- 订单支付。
- 骑手调度。
- 订单状态实时推送。
- 退款和投诉。
- 风控和异常订单。
