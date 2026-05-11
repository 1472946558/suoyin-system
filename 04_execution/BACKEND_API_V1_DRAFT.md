# 后端 API V1 草案

## 1. 文档定位

本文档定义黄金回收收银小程序首版后端 API 草案，目标是给前端、后台、后端和测试一个统一的联调口径。当前版本聚焦：

- 多门店
- 三级角色模板 + 能力项勾选
- 收银订单
- 黄金回收单
- 图片附件留痕
- 微信支付主路径
- 模板化复用

本稿是 V1 Draft，不等同于最终接口冻结稿；但字段和模块边界会尽量按“可直接开工”的粒度编写。

## 2. 设计原则

### 2.1 多端共用一套 API

- 小程序端和后台管理端共用 `/api/v1`
- 通过登录来源、能力码和数据范围区分可访问内容

### 2.2 角色与权限不写死页面

- 固定角色模板：老板、店长、员工
- 实际放行以能力码 `permission_code` 为准
- 数据范围以 `data_scope` 为准，不以前端入口做最终判断

### 2.3 数据默认带组织和门店归属

- 所有业务表必须有 `org_id`
- 门店型业务数据必须有 `store_id`
- 跨店查询只能在 `ALL_STORES` 范围内开启

### 2.4 支付与业务解耦

- 支付负责交易发起、回调验签、流水更新、幂等处理
- 订单负责业务状态机

### 2.5 拍照规则配置化

- 默认最少 2 张、最多 3 张
- 支持系统配置切换为“必须 3 张”

## 3. 统一约定

### 3.1 Base URL

```text
/api/v1
```

### 3.2 认证方式

```text
Authorization: Bearer <access_token>
```

### 3.3 返回结构

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "requestId": "req_20260510_xxx"
}
```

### 3.4 时间与金额

- 时间统一使用 ISO 8601，例如 `2026-05-10T14:00:00+08:00`
- 金额统一用“分”存储与传输，字段命名后缀建议为 `_fen`
- 重量建议使用克，字段命名后缀建议为 `_gram`

### 3.5 常见错误码

| 错误码 | 含义 |
| --- | --- |
| `40001` | 参数错误 |
| `40101` | 未登录或 token 无效 |
| `40301` | 无接口权限 |
| `40302` | 数据范围越权 |
| `40401` | 资源不存在 |
| `40901` | 重复提交 |
| `40902` | 状态流转冲突 |
| `50001` | 服务内部错误 |

## 4. 角色、能力码、数据范围

### 4.1 角色模板

| 角色 | 默认说明 |
| --- | --- |
| `owner` | 老板，默认全门店视角，可控权限、账号、配置、支付 |
| `manager` | 店长，默认本门店运营视角 |
| `staff` | 员工，默认本门店执行视角 |

### 4.2 数据范围

| 值 | 说明 |
| --- | --- |
| `ALL_STORES` | 全部门店 |
| `CURRENT_STORE` | 当前门店 |
| `SELF` | 仅本人数据 |

### 4.3 首批能力码建议

| 能力码 | 含义 |
| --- | --- |
| `dashboard.view` | 查看首页统计 |
| `store.manage` | 门店管理 |
| `user.manage` | 账号管理 |
| `role.manage` | 角色权限管理 |
| `product.manage` | 商品管理 |
| `member.view` | 查看会员 |
| `member.manage` | 编辑会员 |
| `cashier.order.create` | 创建收银订单 |
| `cashier.order.view` | 查看收银订单 |
| `recycle.order.create` | 创建回收单 |
| `recycle.order.confirm` | 确认回收单 |
| `recycle.order.view` | 查看回收单 |
| `payment.config.manage` | 管理支付配置 |
| `payment.record.view` | 查看支付流水 |
| `system.config.manage` | 管理系统配置 |
| `audit.log.view` | 查看操作日志 |

## 5. 模块划分

### 5.1 认证与会话

负责登录、登出、token 刷新、当前用户身份信息。

### 5.2 组织与门店

负责组织主体、门店信息、员工归属门店、默认当前门店切换。

### 5.3 账号与角色权限

负责账号 CRUD、角色模板、能力项勾选、数据范围配置。

### 5.4 系统配置

负责回收拍照规则、品牌信息、支付方式开关、小票设置等。

### 5.5 商品与会员

负责分类、商品、会员档案，为收银链路提供基础数据。

### 5.6 收银订单

负责购物车落单、优惠备注、订单明细、收款状态。

### 5.7 黄金回收

负责回收单草稿、重量纯度报价、图片附件绑定、确认回收。

### 5.8 附件中心

负责上传凭证初始化、附件确认入库、权限访问控制。

### 5.9 支付

负责统一下单、支付回调、支付流水、现金记账。

### 5.10 统计与审计

负责首页统计、操作日志、关键动作追踪。

## 6. 核心数据对象

### 6.1 User

```json
{
  "id": "usr_001",
  "orgId": "org_001",
  "storeId": "store_001",
  "name": "张三",
  "mobile": "13800000000",
  "roleCode": "manager",
  "status": "active"
}
```

### 6.2 RolePermissionProfile

```json
{
  "roleCode": "manager",
  "dataScope": "CURRENT_STORE",
  "permissions": [
    "dashboard.view",
    "cashier.order.create",
    "recycle.order.create"
  ]
}
```

### 6.3 RecycleOrder

```json
{
  "id": "ro_001",
  "orgId": "org_001",
  "storeId": "store_001",
  "customerName": "李女士",
  "customerMobile": "13900000000",
  "materialType": "gold",
  "weightGram": 12.35,
  "purityRate": 0.916,
  "quotePriceFen": 65000,
  "recycleAmountFen": 80275,
  "photoCount": 2,
  "status": "confirmed"
}
```

### 6.4 PaymentRecord

```json
{
  "id": "pay_001",
  "bizType": "cashier_order",
  "bizId": "co_001",
  "channel": "wechat_pay",
  "amountFen": 128800,
  "status": "paid",
  "transactionNo": "420000xxxx"
}
```

## 7. 关键接口草案

以下仅列 V1 关键接口，字段会在开发时再拆成 request/response schema。

### 7.1 认证与会话

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| `POST` | `/auth/login` | 账号密码登录，支持 `terminal=miniapp/admin` | 公开 |
| `POST` | `/auth/logout` | 登出 | 已登录 |
| `POST` | `/auth/refresh` | 刷新 token | 已登录 |
| `GET` | `/auth/me` | 获取当前用户、角色、能力、可见门店 | 已登录 |
| `POST` | `/auth/switch-store` | 切换当前操作门店 | 已登录且有该门店权限 |

`POST /auth/login` 请求建议：

```json
{
  "terminal": "admin",
  "account": "owner001",
  "password": "******"
}
```

`GET /auth/me` 返回建议包含：

- 用户基础信息
- 角色模板
- 数据范围
- 能力码列表
- 当前门店
- 可访问门店列表

### 7.2 组织与门店

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `GET` | `/stores` | 门店列表 | `store.manage` 或已登录 | `ALL_STORES/CURRENT_STORE` |
| `POST` | `/stores` | 创建门店 | `store.manage` | `ALL_STORES` |
| `GET` | `/stores/{storeId}` | 门店详情 | `store.manage` | 可见门店 |
| `PUT` | `/stores/{storeId}` | 更新门店 | `store.manage` | `ALL_STORES` |
| `PATCH` | `/stores/{storeId}/status` | 启停用门店 | `store.manage` | `ALL_STORES` |

说明：

- 非管理场景下，`GET /stores` 也可作为“当前用户可选择门店列表”接口使用。

### 7.3 账号与角色权限

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `GET` | `/users` | 账号列表 | `user.manage` | `ALL_STORES/CURRENT_STORE` |
| `POST` | `/users` | 创建账号 | `user.manage` | `ALL_STORES/CURRENT_STORE` |
| `GET` | `/users/{userId}` | 账号详情 | `user.manage` | 可见数据 |
| `PUT` | `/users/{userId}` | 更新账号基础信息 | `user.manage` | 可见数据 |
| `PATCH` | `/users/{userId}/status` | 启停用账号 | `user.manage` | 可见数据 |
| `PATCH` | `/users/{userId}/role` | 调整角色模板与数据范围 | `user.manage` | 可见数据 |
| `GET` | `/roles/templates` | 固定角色模板列表 | `role.manage` | `ALL_STORES/CURRENT_STORE` |
| `GET` | `/roles/templates/{roleCode}` | 查看角色模板详情 | `role.manage` | `ALL_STORES/CURRENT_STORE` |
| `PUT` | `/roles/templates/{roleCode}/permissions` | 更新角色模板能力项 | `role.manage` | `ALL_STORES` 优先 |
| `GET` | `/permissions` | 能力项清单 | `role.manage` | `ALL_STORES/CURRENT_STORE` |

说明：

- 第一版不开放完全自定义角色，只允许在固定模板上勾选能力项。
- 店长是否可编辑本店员工权限，可通过能力码单独放行。

### 7.4 系统配置

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `GET` | `/system-config` | 获取系统配置聚合信息 | 已登录 | 按可见范围裁剪 |
| `PUT` | `/system-config/recycle-photo-rule` | 更新回收拍照规则 | `system.config.manage` | `ALL_STORES` |
| `PUT` | `/system-config/brand-profile` | 更新品牌信息 | `system.config.manage` | `ALL_STORES` |
| `PUT` | `/system-config/payment-methods` | 更新支付方式开关 | `payment.config.manage` | `ALL_STORES` |

`GET /system-config` 建议返回：

- 回收拍照规则
- 支付方式开关
- 门店基础配置
- 角色权限入口是否开放

### 7.5 商品与分类

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `GET` | `/product-categories` | 分类列表 | 已登录 | 按门店/组织 |
| `POST` | `/product-categories` | 创建分类 | `product.manage` | 可见数据 |
| `PUT` | `/product-categories/{categoryId}` | 更新分类 | `product.manage` | 可见数据 |
| `GET` | `/products` | 商品列表 | 已登录 | 按门店/组织 |
| `POST` | `/products` | 创建商品 | `product.manage` | 可见数据 |
| `GET` | `/products/{productId}` | 商品详情 | 已登录 | 可见数据 |
| `PUT` | `/products/{productId}` | 更新商品 | `product.manage` | 可见数据 |
| `PATCH` | `/products/{productId}/status` | 启停用商品 | `product.manage` | 可见数据 |

### 7.6 会员

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `GET` | `/members` | 会员列表 | `member.view` | 按门店/组织 |
| `POST` | `/members` | 创建会员 | `member.manage` | `CURRENT_STORE` |
| `GET` | `/members/{memberId}` | 会员详情 | `member.view` | 可见数据 |
| `PUT` | `/members/{memberId}` | 更新会员 | `member.manage` | 可见数据 |
| `GET` | `/members/{memberId}/orders` | 会员消费/回收记录 | `member.view` | 可见数据 |

### 7.7 收银订单

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `POST` | `/cashier-orders` | 创建收银订单 | `cashier.order.create` | `CURRENT_STORE` |
| `GET` | `/cashier-orders` | 收银订单列表 | `cashier.order.view` | 按数据范围 |
| `GET` | `/cashier-orders/{orderId}` | 收银订单详情 | `cashier.order.view` | 可见数据 |
| `POST` | `/cashier-orders/{orderId}/close` | 现金记账并完成订单 | `cashier.order.create` | `CURRENT_STORE` |
| `POST` | `/cashier-orders/{orderId}/cancel` | 取消订单 | `cashier.order.create` | `CURRENT_STORE` |

`POST /cashier-orders` 请求建议：

```json
{
  "storeId": "store_001",
  "memberId": "mem_001",
  "items": [
    {
      "productId": "prod_001",
      "qty": 2,
      "unitPriceFen": 18800
    }
  ],
  "discountFen": 1000,
  "remark": "老会员"
}
```

### 7.8 黄金回收单

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `POST` | `/recycle-orders` | 创建回收单草稿 | `recycle.order.create` | `CURRENT_STORE` |
| `PUT` | `/recycle-orders/{recycleOrderId}` | 更新回收基础信息 | `recycle.order.create` | `CURRENT_STORE` |
| `GET` | `/recycle-orders` | 回收单列表 | `recycle.order.view` | 按数据范围 |
| `GET` | `/recycle-orders/{recycleOrderId}` | 回收单详情 | `recycle.order.view` | 可见数据 |
| `POST` | `/recycle-orders/{recycleOrderId}/confirm` | 确认回收单 | `recycle.order.confirm` | `CURRENT_STORE` |
| `POST` | `/recycle-orders/{recycleOrderId}/cancel` | 取消回收单 | `recycle.order.create` | `CURRENT_STORE` |

`POST /recycle-orders/{id}/confirm` 服务端必须校验：

- 单据状态为可确认
- 图片数量满足配置
- 当前用户对该门店有操作权限

### 7.9 附件上传

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `POST` | `/uploads/presign` | 申请上传凭证 | 已登录 | `CURRENT_STORE` |
| `POST` | `/attachments` | 确认附件入库并绑定业务单据 | 已登录 | 可见数据 |
| `GET` | `/attachments/{attachmentId}` | 获取附件元信息 | 已登录 | 可见数据 |
| `DELETE` | `/attachments/{attachmentId}` | 删除未确认附件 | 已登录 | 可见数据 |

上传流程建议：

1. 前端调用 `/uploads/presign`
2. 直传对象存储
3. 回调 `/attachments` 写入附件记录
4. 业务单据通过 `bizType + bizId` 绑定附件

### 7.10 支付

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `POST` | `/payments/cashier-orders/{orderId}/wechat-prepay` | 发起微信支付统一下单 | `cashier.order.create` | `CURRENT_STORE` |
| `POST` | `/payments/cashier-orders/{orderId}/cash-record` | 现金记账 | `cashier.order.create` | `CURRENT_STORE` |
| `POST` | `/payments/callback/wechat` | 微信支付回调 | 微信签名验签 |
| `GET` | `/payment-records` | 支付流水列表 | `payment.record.view` | 按数据范围 |
| `GET` | `/payment-records/{paymentRecordId}` | 支付流水详情 | `payment.record.view` | 可见数据 |
| `GET` | `/payment-config` | 支付配置详情 | `payment.config.manage` | `ALL_STORES` |
| `PUT` | `/payment-config` | 更新支付配置 | `payment.config.manage` | `ALL_STORES` |

支付约束：

- 同一订单重复发起支付需做幂等
- 微信回调允许重复通知，但状态更新只能成功一次
- 支付成功后由支付模块驱动订单状态转为 `paid`

### 7.11 统计与审计

| 方法 | 路径 | 说明 | 权限 | 数据范围 |
| --- | --- | --- | --- | --- |
| `GET` | `/dashboard/overview` | 首页统计 | `dashboard.view` | 按数据范围 |
| `GET` | `/audit-logs` | 操作日志列表 | `audit.log.view` | 按数据范围 |

`GET /dashboard/overview` V1 可先返回：

- 今日收银额
- 今日订单数
- 今日回收单数
- 回收利润占位字段

## 8. 权限与数据范围约束

### 8.1 通用约束

- 任一读写接口都必须从 token 中解析 `org_id`
- 任一门店型接口都必须校验 `store_id` 是否在用户可见范围
- 禁止仅依赖前端传入的 `storeId` 判权

### 8.2 老板

- 默认 `ALL_STORES`
- 可管理角色能力项、门店、账号、系统配置、支付配置

### 8.3 店长

- 默认 `CURRENT_STORE`
- 可查看并处理本店订单、回收单、商品、会员
- 是否允许管理本店员工，由能力码控制

### 8.4 员工

- 默认 `CURRENT_STORE`
- 以录单、查看本人或本店被授权数据为主
- 默认不可查看跨店统计、支付配置、系统配置、账号管理

### 8.5 附件访问控制

- 附件 URL 不建议永久公开
- 下载或预览应通过服务端鉴权后签发临时访问地址
- 回收单图片属于敏感留档，不允许跨组织访问

### 8.6 支付安全

- 支付回调接口不走用户 token，但必须验签
- 回调处理需记录原始通知报文与处理结果
- 幂等键建议使用微信交易单号或业务单号

## 9. 第一周优先接口

第一周目标是“骨架冻结 + 基础联调可跑通”，优先级如下。

### P0

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/health` | 健康检查 |
| `POST` | `/auth/login` | 小程序/后台登录 |
| `GET` | `/auth/me` | 获取当前身份、角色、能力、门店 |
| `POST` | `/auth/logout` | 登出 |
| `GET` | `/stores` | 当前用户可访问门店列表 |
| `GET` | `/roles/templates` | 角色模板列表 |
| `GET` | `/permissions` | 能力项列表 |
| `GET` | `/system-config` | 拉取系统配置 |

### P1

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/users` | 后台账号列表 |
| `POST` | `/users` | 创建账号 |
| `PUT` | `/users/{userId}` | 更新账号 |
| `PATCH` | `/users/{userId}/role` | 调整角色与数据范围 |
| `POST` | `/stores` | 创建门店 |
| `PUT` | `/stores/{storeId}` | 更新门店 |
| `PUT` | `/roles/templates/{roleCode}/permissions` | 更新角色能力项 |
| `PUT` | `/system-config/recycle-photo-rule` | 调整回收拍照规则 |

### 第一周不建议强推

- 微信真实支付
- 回收单图片上传直传链路
- 统计口径复杂聚合
- 审计日志完整检索

这些接口可以先定 contract，第二周再联调。

## 10. 建表与实现顺序建议

建议后端实现顺序：

1. 用户、角色、能力项、门店、用户门店关系
2. 登录态、鉴权中间件、数据范围中间件
3. 系统配置
4. 商品、会员
5. 收银订单
6. 回收单、附件
7. 支付流水与微信支付
8. 统计与审计

## 11. 待确认口径

- 小程序首版是否采用账号密码登录，还是需要微信登录态绑定账号
- 店长是否拥有本店员工管理权
- 回收支付首版是否需要真实出款
- 利润 / 回收收益字段的计算规则
- 品牌、编号规则、初始化脚本是否要在第一版开放后台配置
