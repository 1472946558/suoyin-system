# flash-delivery-miniapp backend

`backend/` 预留给本项目的 `Node + Fastify` 后端骨架，用于把当前微信小程序从 `mock` 演示模式切到真实 REST API。

## 定位

- 当前小程序前端仍运行在 `utils/config.js` 的 `mode: "mock"`。
- 这个后端目录的目标不是替代前端演示，而是提供后续联调、部署、商用接入的服务端入口。
- 推荐形态：`Fastify + dotenv + REST API`，由 Nginx 反代到本地监听端口。

## 建议目录

```text
backend/
  package.json
  .env
  src/
    app.js
    server.js
    routes/
```

## 本地启动

1. 进入目录：`cd backend`
2. 安装依赖：`npm install`
3. 复制环境变量：`cp .env.example .env`
4. 开发启动：`npm run dev`
5. 生产启动：`npm run start`

建议脚本：

```json
{
  "scripts": {
    "dev": "node --watch src/server.js",
    "start": "node src/server.js"
  }
}
```

默认监听建议为 `127.0.0.1:18092`，与服务器反代规划保持一致。

## 环境变量

最小必需项见 [backend/.env.example](/Users/xiaoliao/Desktop/openclaw/workspace/projects/flash-delivery-miniapp/backend/.env.example)。

关键变量说明：

- `PORT` / `HOST`: Fastify 监听地址，默认建议 `127.0.0.1:18092`
- `APP_URL`: 对外 API 域名，例如 `https://api.your-domain.com`
- `API_PREFIX`: 默认 `/api/v1`
- `APP_MODE`: `memory` / `mysql`，当前建议先用 `memory`
- `JWT_SECRET`: 登录态签名密钥
- `WECHAT_APP_ID` / `WECHAT_APP_SECRET`: 微信小程序登录态换取所需
- `MYSQL_*`: 订单、用户、地址等真实数据存储
- `REDIS_URL`: 可选，用于验证码、会话、限流、队列

## 当前接口范围

当前只建议按前端既有合同先落一版骨架，不要超范围扩展。接口范围以 `docs/API_CONTRACT.md` 为准：

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/healthz` | 健康检查 |
| `POST` | `/api/v1/auth/wechat-login` | 微信登录/注册 |
| `POST` | `/api/v1/pricing/estimate` | 价格预估 |
| `POST` | `/api/v1/orders` | 创建订单 |
| `GET` | `/api/v1/orders` | 订单列表 |
| `GET` | `/api/v1/orders/:id` | 订单详情 |
| `GET` | `/api/v1/orders/:id/track` | 配送跟踪 |
| `POST` | `/api/v1/payments/wechat` | 微信支付占位 |
| `GET` | `/api/v1/profile` | 用户资料 |
| `GET` | `/api/v1/addresses` | 常用地址 |
| `POST` | `/api/v1/support/tickets` | 客服工单 |

建议当前阶段把能力边界写死：

- 可以先返回占位数据或 `501 Not Implemented`
- 不要假装已完成支付、地图、骑手调度、消息通知
- 所有接口先统一成 JSON 响应格式，方便前端切换

## 与小程序的关系

- 小程序当前可在 `mock` 模式完成演示闭环。
- 当后端可用后，把 `utils/config.js` 中的 `mode` 切到 `production`，并把 `apiBaseUrl` 改成真实 HTTPS 域名。
- 小程序负责页面、交互、调用 REST API；后端负责登录态、订单数据、业务规则、支付和第三方集成。
- 没有真实后端前，不应把当前模板视为可直接商用上线版本。

## 服务器部署目标

目标部署位置已在项目资料中明确：

- 代码目录：`/opt/flash-delivery-miniapp/app`
- 日志目录：`/opt/flash-delivery-miniapp/logs`
- 本地监听：`127.0.0.1:18092`

建议部署方式：

1. 将 `backend/` 内容发布到 `/opt/flash-delivery-miniapp/app`
2. 在服务器内执行 `npm ci --omit=dev`
3. 通过 `systemd` 守护 `node src/server.js`
4. 由 Nginx 将正式域名反代到 `127.0.0.1:18092`

## 当前结论

- 现在这份说明服务于“后端骨架落地与联调准备”
- 它不表示后端已经实现，也不表示项目已具备正式上线条件
- 真正进入商用前，还需要补齐数据库、微信登录、支付、地图、日志、监控和发布域名
