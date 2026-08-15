# 变更日志

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
- 我的页面员工入口当前为占位提示（N08 员工端改造时对接）
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
