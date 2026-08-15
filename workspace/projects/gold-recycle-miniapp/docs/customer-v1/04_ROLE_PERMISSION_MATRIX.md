# 04 - 角色权限矩阵

## 1. 角色定义

| 角色 | 标识 | 认证方式 | 数据范围 |
|------|------|----------|----------|
| 游客 | guest | 无 | 公开数据 |
| 顾客 | customer | 微信 openid + 手机号 | 本人数据 |
| 门店员工 | store_staff | 账号密码 | 本门店数据 |
| 店长 | store_manager | 账号密码 | 本门店数据 + 管理 |
| 总管理员 | super_admin | 账号密码 | 全部门店 + 系统配置 |

## 2. 前端页面权限

| 页面 | guest | customer | store_staff | store_manager | super_admin |
|------|:-----:|:--------:|:-----------:|:-------------:|:-----------:|
| 顾客首页 | ✓ | ✓ | ✓ | ✓ | ✓ |
| 商品列表/详情 | ✓ | ✓ | ✓ | ✓ | ✓ |
| 回收介绍 | ✓ | ✓ | ✓ | ✓ | ✓ |
| 门店列表/详情 | ✓ | ✓ | ✓ | ✓ | ✓ |
| 到店预约 | ✗ | ✓ | ✗ | ✗ | ✗ |
| 我的预约 | ✗ | ✓ | ✗ | ✗ | ✗ |
| 预约详情 | ✗ | ✓ | ✗ | ✗ | ✗ |
| 我的（顾客） | ✓ | ✓ | ✓ | ✓ | ✓ |
| 员工入口 | ✓ | ✓ | ✓ | ✓ | ✓ |
| 员工登录 | ✓ | ✓ | ✓ | ✓ | ✓ |
| 员工工作台 | ✗ | ✗ | ✓ | ✓ | ✓ |
| 预约管理（员工） | ✗ | ✗ | ✓ | ✓ | ✓ |
| 管理后台 | ✗ | ✗ | ✗ | ✓ | ✓ |
| 门店管理 | ✗ | ✗ | ✗ | ✗ | ✓ |
| 员工管理 | ✗ | ✗ | ✗ | ✗ | ✓ |
| 系统设置 | ✗ | ✗ | ✗ | ✗ | ✓ |

## 3. 后端接口权限

### 3.1 顾客端接口（/api/v1/customer/）

| 接口 | 权限 | 说明 |
|------|------|------|
| GET /api/v1/customer/home | guest | 顾客首页数据 |
| GET /api/v1/customer/products | guest | 商品列表 |
| GET /api/v1/customer/products/:id | guest | 商品详情 |
| GET /api/v1/customer/products/categories | guest | 商品分类 |
| GET /api/v1/customer/stores | guest | 门店列表 |
| GET /api/v1/customer/stores/:id | guest | 门店详情（含经纬度） |
| GET /api/v1/customer/recycle-info | guest | 回收介绍 |
| POST /api/v1/customer/auth/wechat-login | guest | 微信登录 |
| POST /api/v1/customer/auth/phone-auth | customer | 手机号授权 |
| GET /api/v1/customer/profile | customer | 顾客信息 |
| POST /api/v1/customer/appointments | customer | 创建预约 |
| GET /api/v1/customer/appointments | customer | 我的预约列表 |
| GET /api/v1/customer/appointments/:id | customer | 预约详情 |
| POST /api/v1/customer/appointments/:id/cancel | customer | 取消预约 |
| POST /api/v1/customer/auth/logout | customer | 退出登录 |

### 3.2 员工端接口（/api/v1/ 和 /api/admin/）

| 接口 | 权限 | 说明 |
|------|------|------|
| POST /api/v1/auth/login | guest | 员工登录 |
| POST /api/v1/auth/wechat-login | guest | 微信员工登录 |
| GET /api/v1/dashboard/summary | store_staff+ | 工作台首页 |
| GET /api/v1/appointments | store_staff+ | 预约列表 |
| PUT /api/v1/appointments/:id/status | store_staff+ | 更新预约状态 |
| /api/admin/stores | super_admin | 门店管理 |
| /api/admin/users | super_admin | 员工管理 |
| /api/admin/products | super_admin | 商品管理 |
| /api/admin/system-profile | super_admin | 系统设置 |

## 4. 数据访问边界

### 4.1 顾客数据

| 数据 | guest | customer | store_staff | store_manager | super_admin |
|------|:-----:|:--------:|:-----------:|:-------------:|:-----------:|
| 商品公开信息 | R | R | R | R | R |
| 门店公开信息 | R | R | R | R | R |
| 门店经纬度 | R | R | R | R | R |
| 顾客本人信息 | ✗ | R(本人) | ✗ | ✗ | ✗ |
| 顾客本人预约 | ✗ | R(本人) | ✗ | ✗ | ✗ |
| 门店预约列表 | ✗ | ✗ | R(本店) | R(本店) | R(全部) |
| 顾客手机号 | ✗ | R(本人) | R(本店预约) | R(本店预约) | R(全部) |
| 商品库存/成本 | ✗ | ✗ | R | R | R |
| 员工信息 | ✗ | ✗ | R(本人) | R(本店) | R(全部) |

### 4.2 关键隔离规则

1. **顾客 token ≠ 员工 token**：Redis key 前缀不同（`customer_session:` vs `session:`）
2. **顾客接口前缀**：`/api/v1/customer/`
3. **员工接口前缀**：`/api/v1/`（不含 customer）和 `/api/admin/`
4. **顾客不能调用员工接口**：后端中间件校验 token 类型
5. **员工不能调用顾客接口**：后端中间件校验 token 类型
6. **门店员工只能看本店预约**：`canAccessStore(user, storeID)` 校验
7. **商品 DTO 白名单**：顾客端不返回 benchPrice/inventory/stockStatus/storeIDs/sku
8. **门店经纬度仅在详情接口返回**：列表接口不返回经纬度
