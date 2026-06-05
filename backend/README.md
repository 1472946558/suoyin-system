# gold-recycle-miniapp backend

## 当前状态

这个目录现在已经是可启动、可联调的 Go API 服务，不再只是目录规划说明。

当前已落地能力：

- 健康检查：`GET /health`
- 后台登录：`POST /api/admin/login`
- 后台工作台数据：`GET /api/admin/bootstrap`
- 小程序登录：`POST /api/v1/auth/wechat-login`
- 会员列表 / 详情：`GET /api/v1/members`、`GET /api/v1/members/:id`
- 商品目录 / 详情：`GET /api/v1/products`、`GET /api/v1/products/:id`
- 收银单创建 / 查询：`/api/v1/cashier/orders`
- 回收单创建 / 确认 / 详情：`/api/v1/recycle/orders`
- 系统配置、门店、角色、权限目录等首版接口

当前实现仍是“内存数据版 + 持久化骨架”：

- 适合前后端联调、演示、权限模型确认
- 重启服务后数据会重置
- 尚未接入 MySQL / Redis；OSS 已具备正式配置能力

最新进展：

- 已加入 `persistent` 模式配置入口
- 可通过 `MYSQL_*` 和 `REDIS_*` 环境变量接入持久化
- 当前优先持久化登录会话、收银单、回收单
- 可部署为宿主机本地服务，监听 `127.0.0.1:3001` 后由 Nginx 反代
- 小程序登录已支持真实 `code -> openid` 兑换；当配置 `WECHAT_MINIAPP_APP_ID / WECHAT_MINIAPP_APP_SECRET` 且关闭 `MINIAPP_ALLOW_MOCK_LOGIN` 后，只允许已绑定 `miniAppOpenId` 的账号登录

## 本地运行

```bash
cd backend
go run ./cmd/server
```

默认环境变量：

- `PORT=8080`
- `TOKEN_SECRET=gold-recycle-dev-secret`
- `CORS_ORIGIN=*`
- `APP_MODE=memory`
- `MINIAPP_ALLOW_MOCK_LOGIN=true`

持久化模式额外环境变量：

- `APP_MODE=persistent`
- `MYSQL_HOST`
- `MYSQL_PORT`
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`
- `REDIS_ADDR`
- `REDIS_USER`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_KEY_PREFIX`
- `WECHAT_MINIAPP_APP_ID`
- `WECHAT_MINIAPP_APP_SECRET`
- `WECHAT_API_BASE_URL`
- `MINIAPP_ALLOW_MOCK_LOGIN=false`
- `STORAGE_ENABLED=true`
- `STORAGE_PROVIDER=oss`
- `STORAGE_BUCKET`
- `STORAGE_REGION`
- `STORAGE_ENDPOINT`
- `STORAGE_PUBLIC_BASE_URL`
- `STORAGE_PATH_PREFIX`
- `STORAGE_UPLOAD_STRATEGY=oss_signed_put`
- `STORAGE_ACCESS_KEY_ID`
- `STORAGE_ACCESS_KEY_SECRET`
- `STORAGE_UPLOAD_URL_TTL_SECONDS`

健康检查：

```bash
curl http://127.0.0.1:8080/health
```

## Docker

项目根目录已经提供 `docker-compose.yml`，推荐直接在仓库根目录启动：

```bash
cd gold-recycle-miniapp
cp .env.example .env
docker compose up --build -d
```

## 演示账号

- 老板：`boss / Boss123!`
- 店长：`manager.sz / Manager123!`
- 收银员：`cashier.sz / Cashier123!`

小程序联调默认仍允许通过内部联调 code 进入验收账号。
如果要切正式微信身份登录，需要同时满足：

- 配置 `WECHAT_MINIAPP_APP_ID`、`WECHAT_MINIAPP_APP_SECRET`
- 将 `MINIAPP_ALLOW_MOCK_LOGIN=false`
- 给对应操作员账号补 `miniAppOpenId` 绑定值

## 接口约定

- 前缀：`/api/v1`
- 鉴权：`Authorization: Bearer <token>`
- 返回结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "requestId": "req_xxx"
}
```

错误码分层：

- `400xx`：参数错误
- `401xx`：未登录 / token 无效
- `403xx`：权限不足
- `404xx`：资源不存在
- `409xx`：状态冲突
- `500xx`：服务内部错误

## 下一阶段

要进入真正商用，还需要把这套内存数据版继续替换为正式基础设施：

- MySQL 持久化
- Redis 登录态 / 幂等键缓存
- 对象存储图片留档
- 审计日志持久化
- 正式域名、HTTPS、灰度与监控

这些外部依赖准备好后，这个目录可以继续按 `internal/platform` 和 `internal/modules` 方向拆分，不需要推翻现有接口口径。

## 初始化 SQL

当前持久化表结构样例在：

- `backend/migrations/001_persistence_core.sql`
