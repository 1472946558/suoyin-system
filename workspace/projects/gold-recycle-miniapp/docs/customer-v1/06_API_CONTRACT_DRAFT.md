# 06 - 接口文档草案

> 基于后端实际代码（customer_types.go, customer_store.go, miniapp_auth.go）编写。
> 数据层已实现，HTTP handler + 路由注册待 N06 完成。

## 1. 顾客登录模块

### 1.1 微信登录

```
POST /api/v1/customer/auth/wechat-login
Auth: 无（公开接口）
Permission: guest
```

**Request:**
```json
{
  "code": "wx.logincode",
  "phoneCode": "optional_phonecode",
  "nickname": "可选",
  "avatarUrl": "可选"
}
```

**Response 200:**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "customer_xxx",
    "customerId": "customer-xxx",
    "phone": "",
    "isNew": true
  },
  "requestId": "req-xxx"
}
```

**Error Codes:**
| code | 说明 |
|------|------|
| 40001 | code 缺失 |
| 40002 | 微信 code 交换失败 |
| 50001 | 服务器内部错误 |

### 1.2 手机号授权

```
POST /api/v1/customer/auth/phone-auth
Auth: customer token
Permission: customer
```

**Request:**
```json
{
  "phoneCode": "phone_code_from_getPhoneNumber"
}
```

**Response 200:**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "phone": "138****1234"
  },
  "requestId": "req-xxx"
}
```

### 1.3 退出登录

```
POST /api/v1/customer/auth/logout
Auth: customer token
Permission: customer
```

## 2. 顾客资料模块

### 2.1 获取顾客信息

```
GET /api/v1/customer/profile
Auth: customer token
Permission: customer
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "customer-xxx",
    "phone": "138****1234",
    "nickname": "微信昵称",
    "avatarUrl": "https://..."
  }
}
```

## 3. 商品展示模块

### 3.1 商品列表

```
GET /api/v1/customer/products?category=项链&page=1&pageSize=20
Auth: 无（公开接口）
Permission: guest
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": "prod-001",
        "name": "足金项链",
        "imageUrl": "https://...",
        "category": "项链",
        "purity": "足金999",
        "retailPrice": 3580.00,
        "gramWeight": 10.5,
        "recommendedScene": "日常佩戴",
        "tags": ["热销"]
      }
    ],
    "total": 50,
    "page": 1,
    "pageSize": 20
  }
}
```

**DTO 白名单**（不返回的字段）：benchPrice, inventory, stockStatus, storeIDs, sku

### 3.2 商品详情

```
GET /api/v1/customer/products/:id
Auth: 无（公开接口）
Permission: guest
```

### 3.3 商品分类

```
GET /api/v1/customer/products/categories
Auth: 无（公开接口）
Permission: guest
```

**Response 200:**
```json
{
  "code": 0,
  "data": ["项链", "手镯", "戒指", "耳饰"]
}
```

## 4. 门店模块

### 4.1 门店列表

```
GET /api/v1/customer/stores
Auth: 无（公开接口）
Permission: guest
```

**Response 200:**
```json
{
  "code": 0,
  "data": [
    {
      "id": "store-001",
      "name": "金匠馆总店",
      "city": "北京",
      "address": "朝阳区xxx",
      "contactPhone": "010-12345678",
      "businessHours": "09:00-18:00",
      "distance": 2.5
    }
  ]
}
```

**注意**：列表接口不返回经纬度。

### 4.2 门店详情

```
GET /api/v1/customer/stores/:id
Auth: 无（公开接口）
Permission: guest
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "store-001",
    "name": "金匠馆总店",
    "city": "北京",
    "address": "朝阳区xxx",
    "contactPhone": "010-12345678",
    "businessHours": "09:00-18:00",
    "longitude": 116.404,
    "latitude": 39.915
  }
}
```

**注意**：详情接口返回经纬度（用于地图导航）。

## 5. 首页和回收介绍

### 5.1 首页数据

```
GET /api/v1/customer/home
Auth: 无（公开接口）
Permission: guest
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "banners": [
      {"imageUrl": "https://...", "linkType": "products", "linkUrl": ""}
    ],
    "brandName": "金匠馆",
    "brandSlogan1": "旧金换打新款",
    "brandSlogan2": "包损耗",
    "servicePhone": "400-xxx-xxxx"
  }
}
```

### 5.2 回收介绍

```
GET /api/v1/customer/recycle-info
Auth: 无（公开接口）
Permission: guest
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "title": "黄金回收服务",
    "content": "金匠馆专业黄金回收服务...",
    "process": "1. 到店咨询\n2. 黄金检测\n3. 确认价格\n4. 完成回收",
    "notes": "最终回收价格以门店线下检测为准。"
  }
}
```

## 6. 预约模块（顾客）

### 6.1 创建预约

```
POST /api/v1/customer/appointments
Auth: customer token
Permission: customer
```

**Request:**
```json
{
  "storeId": "store-001",
  "appointmentDate": "2026-08-16",
  "appointmentTime": "14:00",
  "contactName": "张三",
  "contactPhone": "13800138000",
  "remark": "想回收一条金项链"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "appt-1-xxx",
    "storeId": "store-001",
    "storeName": "金匠馆总店",
    "storeAddress": "北京朝阳区xxx",
    "storePhone": "010-12345678",
    "appointmentDate": "2026-08-16",
    "appointmentTime": "14:00",
    "status": "PENDING",
    "statusText": "待确认",
    "remark": "想回收一条金项链",
    "createdAt": "2026-08-15T10:00:00Z"
  }
}
```

**Error Codes:**
| code | 说明 |
|------|------|
| 40002 | 参数缺失（storeId, date, time, contactName, contactPhone 必填） |
| 40003 | 预约时间无效（当天提前 1 小时，7 天范围内） |
| 40009 | 重复预约（同一手机号 + 同一门店 + 同一时间段） |
| 40404 | 门店不存在或未激活 |
| 40101 | 未登录或 token 无效 |

### 6.2 我的预约列表

```
GET /api/v1/customer/appointments
Auth: customer token
Permission: customer
```

**Response 200:**
```json
{
  "code": 0,
  "data": [
    {
      "id": "appt-1-xxx",
      "storeId": "store-001",
      "storeName": "金匠馆总店",
      "storeAddress": "北京朝阳区xxx",
      "storePhone": "010-12345678",
      "appointmentDate": "2026-08-16",
      "appointmentTime": "14:00",
      "status": "PENDING",
      "statusText": "待确认",
      "remark": "...",
      "createdAt": "2026-08-15T10:00:00Z",
      "cancelledAt": null,
      "confirmedAt": null
    }
  ]
}
```

### 6.3 预约详情

```
GET /api/v1/customer/appointments/:id
Auth: customer token
Permission: customer（仅本人预约）
```

### 6.4 取消预约

```
POST /api/v1/customer/appointments/:id/cancel
Auth: customer token
Permission: customer（仅本人预约）
```

**Request:**
```json
{
  "reason": "时间有变"
}
```

**Error Codes:**
| code | 说明 |
|------|------|
| 40005 | 预约已取消 |
| 40006 | 距预约时间不足 2 小时，不可取消 |
| 40007 | 状态不允许取消（非 PENDING/CONFIRMED） |
| 40404 | 预约不存在 |

## 7. 员工预约管理

### 7.1 预约列表

```
GET /api/v1/appointments?date=2026-08-16&status=PENDING
Auth: 员工 token
Permission: store_staff+
```

**Response 200:**
```json
{
  "code": 0,
  "data": [
    {
      "id": "appt-1-xxx",
      "customerName": "张三",
      "customerPhone": "13800138000",
      "storeId": "store-001",
      "storeName": "金匠馆总店",
      "appointmentDate": "2026-08-16",
      "appointmentTime": "14:00",
      "status": "PENDING",
      "statusText": "待确认",
      "remark": "...",
      "createdAt": "2026-08-15T10:00:00Z",
      "confirmedAt": null,
      "confirmedBy": ""
    }
  ]
}
```

**权限**：store_staff/store_manager 只能看本门店预约，super_admin 可看全部。

### 7.2 更新预约状态

```
PUT /api/v1/appointments/:id/status
Auth: 员工 token
Permission: store_staff+
```

**Request:**
```json
{
  "status": "CONFIRMED"
}
```

**可执行的状态转换**：
- PENDING → CONFIRMED, ARRIVED, CANCELLED, NO_SHOW, TERMINATED
- CONFIRMED → ARRIVED, CANCELLED, NO_SHOW, TERMINATED
- ARRIVED → COMPLETED, TERMINATED

**Error Codes:**
| code | 说明 |
|------|------|
| 40008 | 状态转换不允许 |
| 40301 | 无权操作该门店预约 |

## 8. 后台门店经纬度管理

### 8.1 更新门店（含经纬度）

```
PUT /api/admin/stores/:id
Auth: 管理员 token
Permission: super_admin (store.manage)
```

**Request:**
```json
{
  "name": "金匠馆总店",
  "city": "北京",
  "address": "朝阳区xxx",
  "contactPhone": "010-12345678",
  "businessHours": "09:00-18:00",
  "longitude": 116.404,
  "latitude": 39.915,
  "status": "active"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "store-001",
    "name": "金匠馆总店",
    "longitude": 116.404,
    "latitude": 39.915,
    "...": "..."
  }
}
```

## 9. 统一响应格式

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "requestId": "req-xxx"
}
```

| code 范围 | 说明 |
|-----------|------|
| 0 | 成功 |
| 400xx | 客户端错误 |
| 401xx | 认证错误 |
| 403xx | 权限错误 |
| 404xx | 资源不存在 |
| 500xx | 服务器错误 |

## 10. 认证方式

### 顾客接口
```
Authorization: Bearer <customer_token>
```
Redis key: `customer_session:<token>`

### 员工接口
```
Authorization: Bearer <operator_token>
```
Redis key: `session:<token>`

**两套 token 不可互用**，后端中间件校验 token 类型。
