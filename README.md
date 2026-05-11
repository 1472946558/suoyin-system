# 金匠倌收银系统

当前仓库包含三部分：

- `miniapp/`：微信小程序前台
- `admin/`：老板/店长后台管理端
- `backend/`：Go API 服务，当前为可联调的内存数据版

当前状态：

- 后台管理端可通过 Docker 直接启动并登录演示
- Go API 已打通登录、会员、商品、收银单、回收单主链路
- 后端已具备 `persistent` 模式，可接 MySQL / Redis 保存会话与主业务订单
- 小程序保留 mock 演示能力，同时新增“联调模式 / 接口地址”运行时切换

## 本地启动

### 方式一：Docker Compose

```bash
cp .env.example .env
docker compose up --build -d
```

启动后默认访问：

- 后端健康检查：`http://127.0.0.1:18082/health`
- 后台管理端：`http://127.0.0.1:14176`
- 如果宿主机已有 Nginx，可把 `80` 端口反代到 `14176`

默认演示账号：

- 老板：`boss / Boss123!`
- 店长：`manager.sz / Manager123!`
- 收银员：`cashier.sz / Cashier123!`

### 方式二：本地开发

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

如果要在微信开发者工具里接本地后端：

1. 打开“我的”页
2. 进入 `联调模式`
3. 切到 `production`
4. 把 `接口地址` 设为当前可访问的 API Base URL
   - 本机开发者工具联调可先使用 `http://127.0.0.1:18082`

## Smoke Test

```bash
node scripts/smoke-test.mjs
```

默认会验证：

- 健康检查
- 后台登录
- 后台 bootstrap
- 小程序登录
- 会员列表
- 商品列表
- 收银单创建
- 回收单创建与确认

我已经在本机执行通过一轮 `docker compose up --build -d` 和 `node scripts/smoke-test.mjs`。

## 服务器部署备注

- 仓库已固定镜像名：
  - `suoyin-system-backend:latest`
  - `suoyin-system-admin:latest`
- 对于只能使用老版 `docker-compose` 的服务器，当前 `docker-compose.yml` 也可直接运行
- 如果服务器公网 `80/443` 已由宿主机 Nginx 接管，可参考 `deploy/nginx/suoyin-system.conf`
- 如果后端需要脱离容器独立运行，可参考 `deploy/systemd/miniapp-backend.service`

## 当前可商用前还需要的外部资源

- 微信小程序正式 `AppID`
- 微信支付商户号、证书、回调域名
- OSS / COS 存储桶和访问域名
- 正式域名与 HTTPS 证书
- 打印机实际型号与模板联调数据

这些外部资源一到位，就可以继续从“本地可联调版本”推进到“正式可商用版本”。更细的交付清单见 `docs/commercial-readiness.md`。
