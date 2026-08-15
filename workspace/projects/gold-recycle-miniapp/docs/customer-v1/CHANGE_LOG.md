# 变更日志

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
