# 黄金回收收银系统模板

这个仓库是一个通用的黄金门店收银与回收留档系统模板，包含：

- `miniapp/`：微信小程序前台
- `admin/`：后台管理端
- `backend/`：Go API 服务

仓库只保留通用代码、演示数据和占位配置；正式客户资料、联系人、服务器、域名、AppID、AppSecret、账号凭证、验收证据和上线资料不应提交到公开代码仓库。

## 本地启动

### Docker Compose

```bash
cp .env.example .env
docker compose up --build -d
```

默认访问：

- 后端健康检查：`http://127.0.0.1:18082/health`
- 后台管理端：`http://127.0.0.1:14176`

默认演示账号：

- 老板：`boss / Boss123!`
- 店长：`manager.sz / Manager123!`
- 收银员：`cashier.sz / Cashier123!`

### 本地开发

后端：

```bash
cd backend
go run ./cmd/server
```

后台：

```bash
cd admin
npm install
npm run dev
```

小程序：

```bash
cd miniapp
npm install
npm run validate
```

## 验证

```bash
cd miniapp && npm run validate
cd backend && go test ./...
cd admin && npm run build
```

API 烟测：

```bash
node scripts/smoke-test.mjs
```

## 配置说明

- `.env.example` 只放占位值。
- 微信小程序 AppID 使用 `touristappid` 或在本地私有配置中覆盖。
- 生产域名、服务器 IP、对象存储、微信密钥、数据库密码等必须放在私有环境变量或私有台账中。
- 模板演示录屏、二维码、提审资料、合同资料、联系人信息和上线记录不进入本仓库。
