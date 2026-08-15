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

---

# 附：管理后台「顾客端内容管理」接口草案（2026-08-15 补充）

> 详见 12_ADMIN_CONFIG_SPEC.md、13_UPLOAD_ASSET_SPEC.md。
> **前缀定案（2026-08-15）**：后台管理接口统一沿用现有项目前缀 `/api/admin/`（不新增 `/api/admin/`）；顾客端接口继续用 `/api/v1/customer/*`。理由：现有权限中间件、路由与前端请求封装均按 `/api/admin/` 实现，沿用零兼容成本。
> 所有接口 Auth 均为**管理员 token**（`session:` 前缀），顾客 token 一律 401/403。

## 11. 图片上传（详见 13 号文档）

```
POST   /api/admin/uploads/images          multipart(file, scene, refId) → {id, url, path, ...}
DELETE /api/admin/uploads/images/:id      软删除，被引用时 40013
GET    /api/admin/uploads/images?scene=   素材库列表（分页）
```

## 12. Banner 管理

```
GET    /api/admin/customer-home/banners        列表（含禁用项，按 sortOrder）
POST   /api/admin/customer-home/banners         新建
PUT    /api/admin/customer-home/banners/:id     编辑（title/subtitle/imageUrl/linkType/linkTarget/sortOrder/enabled）
DELETE /api/admin/customer-home/banners/:id     删除（二次确认由前端保证）
```

Banner 对象：`{id, title, subtitle, imageUrl, linkType(none|style_list|style_detail|store_list|store_detail|booking|external_page), linkTarget, sortOrder, enabled, updatedAt}`

权限：GET=customer_content.view；POST/PUT/DELETE=customer_content.manage（仅 boss）。

## 13. 首页文案配置

```
GET /api/admin/customer-home/config
PUT /api/admin/customer-home/config
```

PUT Body：
```json
{
  "brandName": "金匠倌",
  "brandSlogan": "旧金换打新款 · 包损耗",
  "serviceCopy": "黄金维修 · 到店回收 · 款式定制",
  "entryStyleText": "款式图",
  "entryFeeText": "工费",
  "nearbyStoreRule": {"mode": "distance", "count": 3},
  "servicePhone": "400-xxx-xxxx",
  "appointmentNotes": "预约须知文本...",
  "serviceIntro": "服务说明文本...",
  "locationPermissionNote": "用于展示附近门店..."
}
```

**不提供**"旧金回收独立入口"任何字段。店长入口文案不在此配置（固定值）。

## 14. 款式/工费管理（复用商品管理扩展）

```
GET    /api/admin/customer-products?category=&status=&keyword=&page=&pageSize=
POST   /api/admin/customer-products
PUT    /api/admin/customer-products/:id
DELETE /api/admin/customer-products/:id
```

> 实现说明：与现有 `/api/admin/products` 同一数据源（app_configs `catalog_products`）。建议直接扩展现有接口的请求/响应字段，`/api/admin/customer-products` 作为别名路由；禁止两套独立数据。

对象扩展字段（在原商品字段之上）：
```json
{
  "laborFeeRef": "35元/克 起",
  "description": "款式说明...",
  "laborFeeNote": "工费说明...",
  "images": ["https://...主图", "https://...详情图2"],
  "applicableServiceTypes": ["OLD_FOR_NEW", "REPAIR"],
  "recommendedStoreRule": "nearest",
  "sortOrder": 100,
  "isRecommended": false,
  "isHot": false
}
```

辅助接口：
```
GET  /api/admin/customer-products/categories     款式分类列表
POST /api/admin/customer-products/categories     新建分类 {name, sortOrder}
PUT  /api/admin/customer-products/categories/:id 重命名/排序（删除时校验非空）
GET/PUT /api/admin/customer-fee-note             全局工费说明 {content}
```

## 15. 门店扩展管理（复用门店管理扩展）

```
GET /api/admin/stores                            原有列表 + 扩展字段
GET /api/admin/stores/:id/customer-config        读取顾客端扩展配置（2026-08-15 开发期补充）
PUT /api/admin/stores/:id/customer-config        顾客端扩展配置
```

PUT Body（customer-config 仅顾客端相关字段；名称等原字段仍走原 `/api/admin/stores/:id`）：
```json
{
  "imageUrl": "https://...门店图",
  "appointmentEnabled": true,
  "serviceTags": ["以旧换新", "维修"],
  "sortOrder": 100,
  "longitude": 116.404,
  "latitude": 39.915,
  "contactPhone": "010-12345678",
  "businessHours": "09:30-21:30",
  "address": "朝阳区xxx",
  "status": "active"
}
```

权限：boss 全部门店；shop_manager 仅本店（visibleStoreIDs 校验），且可改字段限于图片/电话/营业时间/经纬度/地址/预约开关/服务标签。

## 16. 预约规则配置

```
GET /api/admin/appointment-rules
PUT /api/admin/appointment-rules
```

PUT Body（全部带范围校验，越界 40014）：
```json
{
  "bookableServiceTypes": ["OLD_FOR_NEW", "REPAIR", "CONSULT", "RECYCLE"],
  "bookableDays": 7,
  "slotMinutes": 30,
  "openTime": "09:30",
  "closeTime": "21:30",
  "sameDayLeadMinutes": 60,
  "cancelLeadMinutes": 120,
  "slotCapacity": 2
}
```

> 实现说明：当前这些规则硬编码在 customer_store.go，本接口落地时抽为配置（存 app_configs `appointment_rules`），slots 计算逻辑读配置。

## 17. 预约记录管理（后台）

```
GET  /api/admin/customer-appointments?status=&storeId=&phone=&dateFrom=&dateTo=&serviceType=&page=&pageSize=
GET  /api/admin/customer-appointments/:id
POST /api/admin/customer-appointments/:id/confirm     PENDING → CONFIRMED
POST /api/admin/customer-appointments/:id/arrive      PENDING/CONFIRMED → ARRIVED
POST /api/admin/customer-appointments/:id/complete    ARRIVED → COMPLETED
POST /api/admin/customer-appointments/:id/cancel      PENDING/CONFIRMED → CANCELLED（body: reason，必填；不受 2 小时限制）
POST /api/admin/customer-appointments/:id/no-show     PENDING/CONFIRMED → NO_SHOW
PUT  /api/admin/customer-appointments/:id/staff-note  内部备注（顾客端不可见）
```

状态机与 07 号文档第 3 节一致；非法转换 40008；越权门店 40301；预约不存在 40404。

权限：appointment.manage（boss 全部；shop_manager 仅本店）。

## 18. 新增错误码汇总

| code | 说明 |
|------|------|
| 40010 | 未选择上传文件 |
| 40011 | 图片格式不支持（魔数校验失败） |
| 40012 | 图片超过大小限制（5MB） |
| 40013 | 图片仍被业务引用，禁止删除 |
| 40014 | 预约规则取值越界 |
| 40015 | Banner 数量超上限（8 张启用中） |
| 40301 | 无权操作（角色/门店越权） |
