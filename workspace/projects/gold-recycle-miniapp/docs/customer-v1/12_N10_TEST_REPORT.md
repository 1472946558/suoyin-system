# N10 - 联调测试与安全修复报告

> 生成日期：2026-08-15
> 测试范围：后端权限隔离、数据越权、前端安全、回归测试
> 测试方式：代码审计 + 自动化测试 + 安全漏洞修复

---

## 一、测试概览

| 类别 | 检查项数 | 通过 | 修复 | 待真机验证 |
|------|---------|------|------|-----------|
| 后端权限隔离 | 8 | 8 | 0 | 0 |
| 后端数据越权 | 5 | 5 | 0 | 0 |
| SQL 注入 | 3 | 3 | 0 | 0 |
| 前端 Token 隔离 | 2 | 2 | 0 | 0 |
| 前端管理入口 | 2 | 2 | 0 | 0 |
| 前端 XSS | 2 | 2 | 0 | 0 |
| 前端请求安全 | 2 | 2 | 0 | 0 |
| 前端登录守卫 | 4 | 2 | 2 | 0 |
| 敏感信息泄露 | 3 | 2 | 0 | 1 |
| 后端回归测试 | 1 | 1 | 0 | 0 |
| **合计** | **34** | **31** | **2** | **1** |

---

## 二、后端权限隔离审计

### 2.1 中间件隔离机制

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 顾客 token 不能通过 `withAuth` | ✅ 通过 | `withAuth` 调用 `getUserByToken` → 查 `s.sessions` map；顾客 token 存储在 `s.customerSessions`，不会命中 |
| 员工 token 不能通过 `withCustomerAuth` | ✅ 通过 | `withCustomerAuth` 调用 `getCustomerByToken` → 查 `s.customerSessions` map；员工 token 存储在 `s.sessions`，不会命中 |
| Redis key 前缀隔离 | ✅ 通过 | 员工：`session:` + token；顾客：`customer_session:` + token |
| 顾客禁用账号拦截 | ✅ 通过 | `withCustomerAuth` 检查 `profile.Status != "active"` → 403 |

### 2.2 路由中间件覆盖

| 路由组 | 中间件 | 结果 |
|--------|--------|------|
| `/api/v1/customer/home, products, stores, recycle-info` | 无（公开） | ✅ 正确 |
| `/api/v1/customer/auth/wechat-login` | 无（公开） | ✅ 正确 |
| `/api/v1/customer/auth/phone, logout, me` | `withCustomerAuth` | ✅ 正确 |
| `/api/v1/customer/appointments` | `withCustomerAuth` | ✅ 正确 |
| `/api/v1/staff/appointments` | `withAuth` | ✅ 正确 |
| `/api/admin/customer-appointments` | `withAuth` + `requireAdminAbility("appointment.manage")` | ✅ 正确 |
| `/api/admin/appointment-rules` | `withAuth` + `requireAdminAbility("customer_content.view")` | ✅ 正确 |

### 2.3 权限隔离验收项映射

| 验收编号 | 验收项 | 代码审计结果 |
|---------|--------|-------------|
| PERM-01 | 顾客 token 调员工接口 → 401/403 | ✅ `withAuth` 查 `s.sessions`，顾客 token 不在其中 → 401 |
| PERM-02 | 员工 token 调顾客接口 → 401/403 | ✅ `withCustomerAuth` 查 `s.customerSessions`，员工 token 不在其中 → 401 |
| PERM-03 | 无 token 调认证接口 → 401 | ✅ `withCustomerAuth`/`withAuth` 检查 `token == ""` → 401 |
| PERM-04 | 无 token 调公开接口 → 正常 | ✅ 公开路由用 `HandleFunc` 无中间件 |
| PERM-05 | 顾客只能看自己预约 | ✅ `listCustomerAppointments(customerID)` / `getCustomerAppointment(customerID, appointmentID)` 均按 customerID 过滤 |
| PERM-06 | 顾客不能访问管理后台 | ✅ `/api/admin/*` 路由用 `withAuth` + `requireAdminAbility`，顾客 token 无法通过 `withAuth` |
| PERM-07 | 前端不暴露管理入口 | ✅ 顾客端仅"我的"页底部"店长入口"（`wx.navigateToMiniProgram`），无其他管理入口 |
| PERM-08 | 顾客调用员工预约状态更新 → 403 | ✅ `/api/v1/staff/appointments/` 用 `withAuth`，顾客 token → 401 |

---

## 三、后端数据越权检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 顾客预约列表按 customerID 过滤 | ✅ | `listCustomerAppointments(customerID)` — `a.CustomerID == customerID` |
| 顾客预约详情按 customerID 过滤 | ✅ | `getCustomerAppointment(customerID, appointmentID)` — `a.ID == appointmentID && a.CustomerID == customerID` |
| 顾客取消预约按 customerID 过滤 | ✅ | `cancelCustomerAppointment(customerID, appointmentID, reason)` |
| 顾客修改备注按 customerID 过滤 | ✅ | `updateCustomerAppointmentNotes(customerID, appointmentID, remark)` |
| 员工预约按门店权限过滤 | ✅ | `listStaffAppointments(user)` / `staffGetAppointment(user, id)` — `canAccessStore(user, a.StoreID)` |

---

## 四、SQL 注入检查

| 检查项 | 结果 | 说明 |
|--------|------|------|
| 无 `fmt.Sprintf` 拼接 SQL | ✅ | persistence.go 全量搜索无命中 |
| 参数化查询 | ✅ | 所有 SQL 使用 `?` 占位符 + `QueryRowContext` / `ExecContext` |
| 搜索关键词参数化 | ✅ | 顾客商品搜索 `keyword` 通过 Go 字符串 `strings.Contains` 内存过滤，不拼 SQL |

---

## 五、前端安全审查

### 5.1 Token 存储隔离

| 端 | 存储 Key | 隔离方式 | 结果 |
|----|---------|---------|------|
| 顾客端 | `customer_token` / `customer_profile` | 独立小程序（不同 AppID）+ 独立 key | ✅ |
| 员工端 | `gr_operator_profile`（含 token 字段） | 独立小程序 + 独立 key | ✅ |

### 5.2 管理入口审查

- 顾客端全量搜索 `admin|staff|管理|员工`：仅"我的"页"店长入口"卡片，通过 `wx.navigateToMiniProgram` 跳转员工端 ✅
- 无隐藏管理入口、无 API 硬编码管理路径 ✅

### 5.3 XSS 风险

- 顾客端所有 WXML 搜索 `rich-text|v-html|innerHTML`：零命中 ✅
- 所有数据通过 WXML 数据绑定渲染（微信默认转义）✅

### 5.4 请求安全

- 顾客端 API base URL：`https://jinjiangguan.com` ✅
- 员工端 API base URL：`https://jinjiangguan.com` ✅
- 全端无明文 `http://` 请求 ✅

### 5.5 登录守卫（已修复）

| 页面 | 修复前 | 修复后 |
|------|--------|--------|
| `miniapp/pages/appointments/index.js` | `onLoad`/`onShow` 直接调 `loadList()`，无登录检查 | `onLoad`/`onShow` 先调 `isLoggedIn()`，未登录 `reLaunch` 到登录页 |
| `miniapp/pages/appointment-detail/index.js` | `onLoad` 直接 `loadDetail()`，无登录检查 | `onLoad` 先调 `isLoggedIn()`，未登录 `reLaunch` 到登录页 |

### 5.6 分享隐私（已修复）

| 页面 | 修复前 | 修复后 |
|------|--------|--------|
| `miniapp/pages/appointment-detail/index.js` | `onShareAppMessage` 允许分享预约详情（含顾客手机号） | 移除 `onShareAppMessage`，微信默认不显示分享按钮 |

### 5.7 敏感信息（已记录，待处理）

| 风险项 | 级别 | 说明 | 建议 |
|--------|------|------|------|
| 员工端 `userStore.js` 硬编码调试账号密码 | 中 | `devtoolAccounts` 含明文密码 `Boss123!` / `Manager123!`，有 `isNonReleaseRuntime()` 守卫但可被反编译 | 提审前从代码中移除，改为运行时配置注入 |

---

## 六、后端回归测试

```
$ go test ./internal/app/ -count=1 -timeout 120s
ok  gold-recycle-miniapp/backend/internal/app  0.024s
```

- 全部测试通过 ✅
- 无回归 ✅

---

## 七、DTO 信息泄露审查

| DTO | 暴露字段 | 敏感字段排查 | 结果 |
|-----|---------|-------------|------|
| `CustomerStoreDTO`（公开） | id, name, city, address, contactPhone, businessHours, imageUrl, serviceTags, appointmentEnabled | 不含内部管理字段 | ✅ |
| `CustomerStoreDetailDTO`（公开） | 同上 + longitude, latitude | 经纬度为导航必需，非敏感 | ✅ |
| `CustomerProductDTO`（公开） | id, name, imageUrl, category, purity, retailPrice, gramWeight, recommendedScene, tags, laborFeeRef, description, images, applicableServiceTypes, isRecommended | 不含成本价/进货价 | ✅ |
| `CustomerAppointmentDTO`（顾客本人） | id, appointmentNo, storeId, storeName, storeAddress, storePhone, serviceType, appointmentDate, appointmentTime, timeRange, status, remark, createdAt, cancelledAt, cancelReason, confirmedAt | 不含其他顾客信息 | ✅ |
| `StaffAppointmentDTO`（员工） | id, appointmentNo, customerName, customerPhone, storeId, storeName, serviceType, appointmentDate, appointmentTime, status, remark, createdAt, confirmedAt, confirmedBy | 含顾客姓名/手机号（员工联系顾客必需） | ✅ |

---

## 八、Mock 登录安全

| 检查项 | 结果 | 说明 |
|--------|------|------|
| `MiniAppAllowMockLogin` 默认值 | ✅ | 仅 `mode == "memory"` 时为 `true`；MariaDB 模式默认 `false` |
| 生产环境配置 | ⚠️ 待确认 | 需确认 `/srv/miniapp/config/.env.production` 中 `MINIAPP_ALLOW_MOCK_LOGIN=false` |

---

## 九、修改文件清单

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `miniapp/pages/appointments/index.js` | 安全修复 | 添加 `isLoggedIn()` 登录守卫 |
| `miniapp/pages/appointment-detail/index.js` | 安全修复 | 添加 `isLoggedIn()` 登录守卫 + 移除 `onShareAppMessage` |

---

## 十、待真机验证项

以下验收项需要微信开发者工具真机预览验证，无法通过代码审计确认：

### 10.1 页面功能验证（188 项验收清单）

- 首页 14 项（H-01 ~ H-14）
- 款式列表 15 项（P-01 ~ P-15）
- 款式详情 13 项（D-01 ~ D-13）
- 门店列表 16 项（S-01 ~ S-16）
- 门店详情 8 项（SD-01 ~ SD-08）
- 到店预约 30 项（A-01 ~ A-30）
- 我的预约 12 项（L-01 ~ L-12）
- 预约详情 17 项（AD-01 ~ AD-17）
- 预约备注 6 项（AN-01 ~ AN-06）
- 回收介绍 7 项（R-01 ~ R-07）
- 我的页 17 项（M-01 ~ M-17）

### 10.2 异常场景验证

- 断网降级（EXC-01 ~ EXC-03）
- token 过期处理（EXC-04）
- 门店无经纬度导航（EXC-05）
- 定位拒绝（EXC-06）
- 手机号授权拒绝（EXC-07）
- 重复预约 / 时段已满（EXC-08 ~ EXC-09）
- 不存在的商品/预约（EXC-10 ~ EXC-11）
- 防重复提交（EXC-12）
- 搜索特殊字符（EXC-13）
- 超长备注截断（EXC-14）

### 10.3 回归验证

- 员工端登录/订单/回收单正常（REG-01 ~ REG-03）
- 管理后台门店/商品/员工管理正常（REG-04 ~ REG-06）
- 微信登录流程正常（REG-07）
- 顾客端与员工端数据隔离（REG-08 ~ REG-10）

### 10.4 生产环境配置确认

- `MINIAPP_ALLOW_MOCK_LOGIN=false`
- `CORSOrigin` 配置正确
- HTTPS 证书有效
- Nginx 反向代理正常

---

## 十一、结论

### 安全状态

- **权限隔离**：后端中间件 + 数据层双重保障，顾客/员工/admin 三端 token 完全隔离 ✅
- **数据越权**：所有查询按身份 ID 过滤，无越权风险 ✅
- **SQL 注入**：全参数化查询，无注入风险 ✅
- **前端安全**：Token 隔离 + 无管理入口泄露 + 无 XSS + 全 HTTPS ✅
- **已修复**：预约页面登录守卫 + 分享隐私保护 ✅
- **待处理**：员工端调试账号密码（提审前移除）⚠️

### 进入 N11 的前提条件

1. ✅ 代码审计全部通过
2. ✅ 后端测试全部通过
3. ✅ 安全漏洞已修复
4. ⚠️ 真机验证（需用户在微信开发者工具中完成）
5. ⚠️ 生产环境 `MINIAPP_ALLOW_MOCK_LOGIN=false` 确认
6. ⚠️ 员工端调试账号密码提审前移除
