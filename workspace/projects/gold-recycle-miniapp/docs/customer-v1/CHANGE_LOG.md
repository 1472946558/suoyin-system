# 变更日志

## 2026-08-17 功能点 05：同 AppID 顾客首页与店长入口

### 本次交付范围
- 将顾客首页、商品、门店、我的及顾客分包页面合并到正式小程序源目录 `miniapp/`，保留原员工页面和员工接口调用路径。
- 将 `miniapp/app.json` 首屏和底部 TabBar 设为顾客端：首页、款式、门店、我的。
- 将「我的」页面的「店长入口」改为同一小程序内部路由 `/pages/account/index`，不再调用同 AppID 的跨小程序跳转。
- 员工端原有页面因不再位于顾客 TabBar，内部返回/切换路径改为 `wx.reLaunch`，避免跳转到不存在的员工 TabBar 项。
- 将顾客公共 WXSS 导入移到每个页面样式文件顶部；将 WXML 中的手机号截取改为 JS 预计算字段，降低微信运行时编译兼容风险。

### 验收证据
- `miniapp/npm run validate`：通过，23 个主包页面和 4 个顾客 TabBar 项配置有效。
- 全量 `miniapp` JavaScript `node --check`：通过。
- 全量 `miniapp` JSON 解析检查：通过。
- 顾客主包/分包页面文件矩阵：通过。
- 顾客首屏路由、店长入口路由、旧跨小程序调用清理断言：通过。
- 微信开发者工具已打开正式源目录 `miniapp/`；普通编译后实际渲染 `pages/customer-home/index`，控制台为 0 错误、0 警告；点击「我的」进入 `pages/customer-mine/index`，点击「店长入口」进入 `pages/account/index` 登录页。

### 交付状态
- 状态：**本地代码、静态检查和开发者工具关键路径验收通过，待真机/真实环境发布确认**。
- 未执行：生产数据库迁移、生产服务器修改、微信体验版/正式版上传、提审和生产部署。
- `miniapp-customer/` 仅保留为历史参考快照，正式开发和上传必须使用 `miniapp/`。

## 2026-08-17 功能点 04：我的预约与预约详情安全交付候选

### 本次交付范围
- 顾客预约列表在未登录时不再触发请求；预约详情、备注编辑页增加参数与登录守卫。
- 顾客预约 DTO 增加服务端计算的 `canCancel`、`canEditNotes`，前端以服务端规则为准展示操作按钮，兼容旧接口时才使用本地兜底。
- 预约详情取消增加重复提交保护；备注页增加重复保存保护、异常 URL 参数安全解码和 200 字限制。
- 预约详情不再尝试拨打脱敏手机号，门店电话继续使用后端白名单字段。
- 后端所有预约详情、取消、备注接口继续以当前顾客 ID 查询，补充顾客间越权访问/操作回归测试。
- 取消截止时间错误文案改为通用表述，避免后台规则调整后仍显示固定“2 小时”。

### 验收证据
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go test ./internal/app -run 'TestCustomerAppointment|TestCustomerAppointmentOwnerIsolation|TestCustomerProfileDTO' -count=1 -timeout=90s`：通过。
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go test ./... -count=1 -timeout=120s`：通过。
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go vet ./...`：通过。
- 顾客端全部 JavaScript/JSON 检查：通过。
- `miniapp/npm run validate`：通过。

### 交付状态
- 状态：**代码与自动化检查通过，待真机/真实环境发布确认**。
- 未执行：生产数据库迁移、生产服务器修改、微信体验版/正式版上传和提审。

## 2026-08-17 功能点 03：黄金回收服务介绍 V1 合规交付候选

### 本次交付范围
- 修复前后端默认文案中“实时金价、估价、现场结算、即时到账”等超出 V1 范围的内容。
- 默认流程调整为“到店咨询 → 黄金检测 → 服务说明 → 到店办理”。
- 后台保存回收介绍时增加 V1 红线校验，拒绝金价、估价、报价、结算、上门、邮寄、支付等内容。
- 顾客公开接口读取存量违规配置时自动返回安全默认内容，避免旧配置继续泄露到顾客端；不执行数据库迁移。
- 保留“最终服务内容以门店现场检测与沟通结果为准”和预约到店 CTA，不增加在线估价或在线结算。

### 验收证据
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go test ./internal/app -run 'TestCustomerRecycleInfoV1Boundary|TestCustomerBrowseEndpointsExposePublicSafeData|TestCustomerAppointment|TestCustomerProfileDTO' -count=1 -timeout=60s`：通过。
- `node --check miniapp-customer/pkg-customer/recycle-info/index.js`：通过。
- `TestCustomerRecycleInfoV1Boundary` 覆盖默认文案、后台拒绝违规文案、存量违规配置公开接口降级。

### 交付状态
- 状态：**代码与自动化检查通过，待真机/真实环境发布确认**。
- 未执行：生产数据库迁移、生产服务器修改、微信体验版/正式版上传和提审。

## 2026-08-17 功能点 02：顾客浏览与最近门店推荐交付候选

### 本次交付范围
- 首页和商品详情页不再默认取第一家门店，统一按顾客定位计算距离并推荐最近门店；未授权定位时保留手动选店和第一家门店兜底。
- 顾客门店列表接口直接返回经纬度，前端不再为每家门店额外请求详情；保留旧接口坐标补齐兼容逻辑。
- 首页 Banner、首页门店图和商品详情门店图统一补全相对 `/assets/` 地址。
- 定位失败、重新定位、打开设置后的状态统一处理，避免重复弹窗并保留手动选择门店提示。
- 顾客端公开接口回归覆盖：仅 active 门店、门店经纬度、商品分页、商品详情、员工字段/内部价格字段不泄露。

### 验收证据
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go test ./internal/app -run 'TestCustomerBrowseEndpointsExposePublicSafeData|TestCustomerAppointment|TestCustomerProfileDTO' -count=1 -timeout=60s`：通过。
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go test ./... -count=1 -timeout=120s`：通过。
- `go test -race`（顾客浏览、预约、顾客 DTO）：通过。
- `go vet ./...`：通过。
- 顾客端 JS/JSON 检查、Haversine 距离计算检查：通过。
- Node 模拟微信运行时验证“近门店优先”和“已有坐标不触发详情 N+1 请求”：通过。
- `miniapp/npm run validate`：通过。

### 交付状态
- 状态：**代码与自动化检查通过，待真机/真实环境发布确认**。
- 未执行：生产数据库、生产服务器、微信体验版/正式版上传和提审。

## 2026-08-17 功能点 01：预约核心闭环商用交付候选

### 本次交付范围
- 修复预约取消在持有写锁时再次获取读锁造成的死锁；同时修复预约时段读取路径的同类锁重入风险。
- 将预约容量已满从 500 错误改为可识别的 40904 业务错误，并补充门店关闭预约时的 40905 错误映射。
- 预约手机号改为服务端读取已完成微信手机号授权的顾客档案，拒绝仅靠客户端手填手机号提交预约。
- 顾客端预约页增加 `phoneVerified` 状态显示与授权后输入框锁定；提交请求不再携带客户端手机号作为可信数据。
- 修复员工端校验器将“处理中...”等文本误判为扩展运算符的问题；顾客端实际扩展运算符全部替换为 `Object.assign`。

### 验收证据
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go test ./... -count=1 -timeout=120s`：通过。
- `go test -race ./internal/app -run 'TestCustomerAppointment|TestCustomerProfileDTO' -count=1 -timeout=60s`：通过。
- `GOCACHE=/private/tmp/gold-recycle-go-build-cache go vet ./...`：通过。
- 顾客端全部 JavaScript 文件 `node --check`：通过。
- `miniapp/npm run validate`：通过。
- 新增 `backend/internal/app/customer_appointment_test.go`，覆盖取消死锁、手机号授权门槛、容量冲突映射、服务端手机号落库和 DTO 脱敏。

### 交付状态
- 状态：**代码与自动化检查通过，待发布确认**。
- 未执行：生产数据库迁移、生产服务器变更、微信开发者工具真机验收、提交审核、生产部署。
- 本功能点通过后才可进入预约场景的真实环境联调；不代表顾客端全部功能或整包小程序已完成商用验收。

## 2026-08-15 页面开发说明文档（8 页完整规格）

### 变更内容
- 基于顾客端原型图，将 8 个页面拆解为可开发、可验收的页面实现说明
- 每页含 10 节：页面目标、页面入口、信息结构、操作说明、交互说明、页面状态、数据字段、接口依赖、跳转关系、验收标准
- 术语统一：管理端入口全部使用"店长入口"，不写"员工入口"

### Files Created
- `03_PAGE_INTERACTION_SPEC.md` — 交互规格文档（11 章：全局规则 + 8 页交互 + 状态矩阵 + 业务规则汇总）
- `03_PAGE_DATA_FIELD_SPEC.md` — 数据字段文档（11 页字段表 + 19 接口总览 + 工具函数 + DTO 排除清单）
- `03_PAGE_ACCEPTANCE_CHECKLIST.md` — 验收清单（188 条：8 页 + 3 辅助页 + 权限隔离 + 异常场景 + 回归）

### Files Updated
- `03_PAGE_SPEC.md` — 从简版页面清单重写为完整规格（8 页 × 10 节 + 3 辅助页 + 全局跳转图）

### 覆盖页面
1. 首页（pages/home/index）
2. 款式列表页（pages/products/index）
3. 款式详情页（pkg-customer/product-detail/index）
4. 门店列表页（pages/stores/index）
5. 到店预约页（pkg-customer/appointment-create/index）
6. 我的预约页（pkg-customer/appointments/index）
7. 预约详情页（pkg-customer/appointment-detail/index）
8. 我的页/店长入口（pages/mine/index）
- 辅助：门店详情页、预约备注页、回收介绍页

### 验收项统计
| 类别 | 数量 |
|------|------|
| 首页 | 14 |
| 款式列表页 | 15 |
| 款式详情页 | 13 |
| 门店列表页 | 16 |
| 门店详情页 | 8 |
| 到店预约页 | 30 |
| 我的预约页 | 12 |
| 预约详情页 | 17 |
| 预约备注页 | 6 |
| 回收介绍页 | 7 |
| 我的页 | 17 |
| 权限隔离 | 8 |
| 异常场景 | 15 |
| 回归验收 | 10 |
| **合计** | **188** |

---

## 2026-08-15 N07 子阶段 3 顾客端开发完成

### 变更内容
- 完成 N07 子阶段 3：预约模块 + 我的 Tab
- 6 个页面全部实现（JS/WXML/WXSS/JSON 4 件套齐全）：

| 页面 | 路径 | 功能 |
|------|------|------|
| 回收介绍 | pkg-customer/recycle-info/ | 回收服务介绍、流程、注意事项、预约 CTA |
| 预约创建 | pkg-customer/appointment-create/ | 选服务/门店/日期/时段/备注 + 手机号授权 + 提交 |
| 预约列表 | pkg-customer/appointments/ | 状态筛选标签、预约卡片、下拉刷新、FAB 新建 |
| 预约详情 | pkg-customer/appointment-detail/ | 完整信息展示、取消预约(>2h)、改备注、联系门店/导航 |
| 预约备注 | pkg-customer/appointment-notes/ | textarea 编辑、200 字限制、保存 API |
| 我的 Tab | pages/mine/ | 顾客头部/登录登出、我的预约入口、回收介绍、客服电话、店长入口 |

### 技术要点
- 预约创建页对接 `getStoreSlots` API 实时获取可约时段，标记已满/已过时段
- 手机号授权使用 `button open-type="getPhoneNumber"` + `customerPhoneAuth`
- 预约取消校验：PENDING/CONFIRMED 状态 + 距预约 > 2 小时
- 预约备注修改：仅 PENDING/CONFIRMED 状态可改
- 我的页面店长入口当前为占位提示（N08 员工端改造时对接）
- 所有页面复用全局 CSS 变量（金色主题 #866E23）

### N07 子阶段完成状态
- 子阶段 1 ✅ 项目骨架 + 公共层 + 首页 + 3 Tab 占位
- 子阶段 2 ✅ 款式 Tab + 门店 Tab（列表+详情+分页+定位排序）
- 子阶段 3 ✅ 预约模块 + 我的 Tab

### 下一步
- N08 员工端改造
- N09 管理后台改造
- N10 联调测试

---

## 2026-08-15 N02-N05 需求冻结 + 客户确认

### 变更内容
- 客户确认 4 项阻断问题：
  - Q-01：预约需要门店确认（PENDING → 门店确认 → CONFIRMED）
  - Q-02：取消预约截止时间 2 小时
  - Q-03：微信服务类目需要调整
  - Q-04：生产环境安全配置已处理
- 所有阻断问题已解决，N06 后端开发可以启动
- 创建 `CUSTOMER_CONFIRMATION.md` 记录客户确认

### 当前实现状态
- 后端数据层已完成：`customer_types.go`、`customer_store.go`
- StoreInfo 已含经纬度字段
- 微信登录流程已实现
- HTTP 路由和处理函数尚未实现 — 这是 N06 的工作

### 结论
- N06 后端开发：**可以启动**
- 后端数据层已就绪，需补全 HTTP handler + 路由注册 + 顾客 token 中间件

---

## 2026-08-15 N01 审计 + N02 PRD + 开发计划

### 变更内容
- 完成项目审计，输出审计报告
- 确认双身份模型：顾客（微信 openid + 手机号）vs 员工（账号密码 + 角色）
- 确认 13 节点开发计划（N00-N12）
- 确认预约规则：30 分钟粒度、7 天范围、当天提前 1 小时、取消提前 2 小时
- 确认 7 态状态机：PENDING → CONFIRMED → ARRIVED → COMPLETED；CANCELLED；NO_SHOW；TERMINATED
- 确认 5 角色：guest / customer / store_staff / store_manager / super_admin
- 确认首页布局：轮播图 → 品牌文案 → 核心入口 → 底部门店预约入口

### 技术栈
- 后端：Go 标准库 net/http（无框架），MySQL 8.4 + Redis 7.4
- 小程序：微信原生（JS/WXML/WXSS）
- 管理后台：Vue 3 + Vite + TypeScript
- 部署：Docker Compose + Nginx 反代
- AppID：wx3f564355bd5c0526
- 生产域名：https://jinjiangguan.com

---

## 2026-08-15 后台「顾客端内容管理」需求与接口设计（文档补充，未写代码）

### 变更内容
- 新增 12_ADMIN_CONFIG_SPEC.md：后台菜单结构、首页配置/款式工费/门店/预约规则/预约记录五大模块字段规格、页面↔配置映射总表、权限矩阵、红线清单
- 新增 13_UPLOAD_ASSET_SPEC.md：统一图片上传 /api/v1/admin/uploads/images 草案（OSS 优先、本地兜底、魔数校验、5MB、四格式、五场景）
- 新增 14_ADMIN_ACCEPTANCE_CHECKLIST.md：9 大类验收清单（含反写死总检）
- 更新 06_API_CONTRACT_DRAFT.md：追加 §11-§18 后台内容管理接口草案（上传/Banner/首页文案/款式/门店/预约规则/预约管理/错误码）
- 更新 07_DATA_MODEL_DRAFT.md：追加 §8.1-§8.9 新增配置模型（HomeBanner、HomeConfig 扩展、款式/门店扩展字段、AppointmentRules、UploadAsset、预约扩展字段、ability 扩展）

### 关键决策
- 复用现有商品管理/门店管理扩展字段，不建第二套数据
- 预约规则由硬编码迁入 app_configs `appointment_rules`
- 管理后台路由前缀 /api/v1/admin/ vs 现有 /api/admin/ 列为 Open Question，N09 前定案
- 红线：首页无旧金回收独立入口；店长入口文案固定；详情页仅咨询门店+到店预约

---

## 2026-08-15 决策补丁节点：后台前缀定案 + 开放问题关闭

### 定案
1. 后台管理接口统一沿用现有前缀 `/api/admin/`，不新增 `/api/v1/admin/`；顾客端接口继续 `/api/v1/customer/*`
2. Banner 最多 8 张（启用中）；建议尺寸 750×300，单张 ≤2MB
3. 店长可改本店预约开关，仅限本店（visibleStoreIDs 校验）
4. 预约规则变更必须记录操作日志
5. 服务说明 V1 不做富文本，多行文本 + 图片；富文本留 V2

### 文档更新
- 06_API_CONTRACT_DRAFT.md：附注改为定案，全部 `/api/v1/admin/` → `/api/admin/`
- 12_ADMIN_CONFIG_SPEC.md：§10 开放问题改为决策记录表
- 13_UPLOAD_ASSET_SPEC.md、14_ADMIN_ACCEPTANCE_CHECKLIST.md：前缀同步
- CHANGE_LOG.md：历史记录中一处"员工入口"指代改为"店长入口"（各文档中"不得出现员工入口字样"为禁令表述，保留）

### 最终口径
顾客端接口：/api/v1/customer/*；后台管理接口：/api/admin/*

---

## 2026-08-15 后台内容管理开发（后端接口 + 管理后台页面）

### 后端（backend/internal/app/）
- 新增 admin_customer_types.go / admin_customer_store.go / admin_customer_handlers.go
- 新路由（前缀 /api/admin/）：uploads/images、customer-home/banners(+/:id)、customer-home/config、customer-products(+/:id, /categories)、appointment-rules、customer-appointments(+/:id/action)、stores/:id/customer-config（GET+PUT）
- /assets/ 静态文件服务（ASSETS_DIR / ASSETS_PUBLIC_BASE_URL 环境变量）
- 预约规则全部参数化（原硬编码 09:30-21:30/30min/容量2/提前1h/7天/取消2h → AppointmentRules 可配置），规则变更记审计日志
- 顾客端 DTO 暴露后台配置字段：商品(laborFeeRef/description/images/isHot 等)、门店(imageUrl/serviceTags/appointmentEnabled)、首页(ServiceCopy/Entry*/AppointmentNotes 等)
- persistence：customer_appointments 增加 service_type、staff_note 列（ensureColumn 自动补列）
- 权限：新增 customer_content.view/manage、appointment.manage；存量 boss 角色自动补全新权限
- 新增测试 admin_customer_test.go：首页文案/Banner 上限/停用过滤、款式扩展字段、门店预约开关、预约规则校验与时段粒度联动

### 管理后台（admin/）
- api.ts：顾客端内容管理 API 全套（首页配置/Banner/款式/门店扩展/图片上传）；request() 支持 FormData
- App.vue：新增「顾客端配置」「款式工费管理」两个页面 + Banner/款式编辑弹窗 + 图片上传（场景化、前端体积校验）
- 门店页新增「顾客端展示配置」卡片（门店图片/预约开关/服务标签/排序）
- 预约管理页此前已完成（列表/详情/状态流转/内部备注/规则配置）

### 待办
- 顾客端小程序对接新字段（home config 新文案、款式工费、门店服务标签）
- 部署与生产验证（需用户确认后执行）
