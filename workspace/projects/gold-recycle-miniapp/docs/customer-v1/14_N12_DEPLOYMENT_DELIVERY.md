# N12 - 上线部署和交付文档

> 生成日期：2026-08-15
> 目标：完成生产环境部署、管理后台发布、小程序提审上线、交付全部项目资产
> 依赖：N11 微信隐私提审已完成

---

## 一、生产环境总览

### 1.1 服务器信息

| 项目 | 值 |
|------|------|
| 云厂商 | 阿里云 ECS |
| 公网 IP | 8.163.54.124 |
| 规格 | 2C / 1.8G / 40G |
| 操作系统 | Alibaba Cloud Linux (kernel 5.10) |
| 域名 | jinjiangguan.com |
| HTTPS | Let's Encrypt 证书，443 端口 |
| HTTP | 80 端口（自动跳转 HTTPS） |

### 1.2 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Nginx | 80 / 443 | 反向代理 + 静态资源 |
| 后端 API | 3001 | Go HTTP 服务（systemd 管理） |
| MariaDB | 3306 | 数据库（仅监听 127.0.0.1） |
| Redis | 6379 | 缓存 + 会话（仅监听 127.0.0.1，有密码） |

### 1.3 Nginx 路由规则

| 路径 | 目标 | 说明 |
|------|------|------|
| `/api/*` | `127.0.0.1:3001` | 后端 API 反向代理 |
| `/` | `/usr/share/nginx/jinjiangguan-static/` | 管理后台静态资源 |

### 1.4 Systemd 服务

| 服务名 | 状态 | 说明 |
|--------|------|------|
| `miniapp-backend.service` | active (running) | Go 后端 API |
| `mariadb.service` | active (running) | MariaDB 数据库 |
| `redis.service` | active (running) | Redis 缓存 |
| `nginx.service` | active (running) | Nginx 反向代理 |

---

## 二、后端部署状态

### 2.1 当前状态：已部署到生产环境

后端代码已于 2026-08-15 部署到生产服务器，以下接口已验证通过：

| 接口 | 方法 | 验证结果 |
|------|------|---------|
| `/api/v1/customer/home` | GET | ✅ |
| `/api/v1/customer/products` | GET | ✅ 306 款商品 |
| `/api/v1/customer/stores` | GET | ✅ 10+ 家门店 |
| `/api/v1/customer/stores/{id}` | GET | ✅ |
| `/api/v1/customer/stores/{id}/slots` | GET | ✅ 25 个时段 |
| `/api/v1/customer/recycle-info` | GET | ✅ |
| `/api/v1/customer/auth/wechat-login` | POST | ✅ |
| `/api/v1/customer/appointments` | POST | ✅ |
| `/api/v1/staff/appointments` | GET | ✅（需员工 token） |
| HTTPS 全链路 | — | ✅ jinjiangguan.com |

### 2.2 后端部署路径

| 项目 | 路径 |
|------|------|
| 二进制 | `/srv/miniapp/backend/gold-recycle-backend` |
| 环境配置 | `/srv/miniapp/config/.env.production` |
| 旧备份 | `/srv/miniapp/backend/backups/`（保留 30 天） |
| systemd 配置 | `/etc/systemd/system/miniapp-backend.service` |

### 2.3 后端更新部署步骤

如果需要重新部署后端（例如代码有更新）：

```bash
# 1. 在开发机交叉编译
cd backend/
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gold-recycle-backend ./cmd/server/

# 2. 上传到服务器
scp gold-recycle-backend root@8.163.54.124:/srv/miniapp/backend/gold-recycle-backend.new

# 3. 在服务器上替换
ssh root@8.163.54.124
cd /srv/miniapp/backend/
cp gold-recycle-backend gold-recycle-backend.bak.$(date +%Y%m%d%H%M)
mv gold-recycle-backend.new gold-recycle-backend
chmod +x gold-recycle-backend
systemctl restart miniapp-backend.service
systemctl status miniapp-backend.service

# 4. 验证
curl -s https://jinjiangguan.com/api/v1/customer/home | head -c 200
```

### 2.4 关键环境变量（.env.production）

以下变量需要在生产环境确认：

| 变量名 | 期望值 | 说明 |
|--------|--------|------|
| `MINIAPP_MODE` | `production` | 生产模式 |
| `MINIAPP_ALLOW_MOCK_LOGIN` | `false` | **必须为 false**，禁止模拟登录 |
| `MINIAPP_DB_HOST` | `127.0.0.1` | 数据库地址 |
| `MINIAPP_DB_PORT` | `3306` | 数据库端口 |
| `MINIAPP_DB_NAME` | `miniapp_prod` | 数据库名 |
| `MINIAPP_DB_USER` | `miniapp_app` | 数据库用户 |
| `MINIAPP_DB_PASSWORD` | （保密） | 数据库密码 |
| `MINIAPP_REDIS_ADDR` | `127.0.0.1:6379` | Redis 地址 |
| `MINIAPP_REDIS_PASSWORD` | （保密） | Redis 密码 |
| `MINIAPP_WX_APPID` | `wx3f564355bd5c0526` | 微信小程序 AppID |
| `MINIAPP_WX_SECRET` | （保密） | 微信小程序密钥 |

> **待确认**：请在生产服务器上执行 `grep MINIAPP_ALLOW_MOCK_LOGIN /srv/miniapp/config/.env.production` 确认值为 `false`。

---

## 三、管理后台部署

### 3.1 构建状态

管理后台已于 2026-08-15 完成生产构建：

```
vue-tsc -b && vite build
✓ 14 modules transformed
dist/index.html                   1.11 kB │ gzip:  0.81 kB
dist/assets/index-TN6YoP9O.css   14.51 kB │ gzip:  3.89 kB
dist/assets/index-D8gA1Aps.js   222.90 kB │ gzip: 64.30 kB
✓ built in 276ms
```

编译 0 错误，构建产物在 `admin/dist/`。

### 3.2 部署到生产服务器

```bash
# 1. 上传 dist/ 到服务器
scp -r admin/dist/* root@8.163.54.124:/usr/share/nginx/jinjiangguan-static/

# 2. 验证
curl -s https://jinjiangguan.com/ | head -c 200

# 3. 如果有缓存问题，重启 Nginx
ssh root@8.163.54.124 "nginx -s reload"
```

### 3.3 管理后台访问地址

| 环境 | 地址 |
|------|------|
| 生产 | `https://jinjiangguan.com/` |
| 本地开发 | `http://localhost:5173/`（需后端在 8080 端口） |

### 3.4 管理后台功能清单

| 功能 | 说明 |
|------|------|
| 门店管理 | 门店 CRUD + 经纬度/联系电话/营业时间编辑 |
| 商品管理 | 商品 CRUD + 批量导入 |
| 预约管理 | 预约列表 + 状态操作 + 详情面板 + 内部备注 |
| 预约规则 | 可预约天数/时段粒度/营业时间/提前量/容量/服务类型 |
| 顾客端内容 | 首页 Banner / 款式展示 / 门店展示配置 |
| 会员管理 | 会员列表 + 详情 |
| 订单管理 | 收银订单 + 回收订单 |
| 库存管理 | 商品库存 + 物料出入库 |
| 系统设置 | 角色权限 / 用户管理 |

---

## 四、数据库文档

### 4.1 数据库连接信息

| 项目 | 值 |
|------|------|
| 数据库类型 | MariaDB 10.5 |
| 数据库名 | `miniapp_prod` |
| 用户名 | `miniapp_app` |
| 连接地址 | `127.0.0.1:3306`（仅本地） |
| 字符集 | `utf8mb4` / `utf8mb4_unicode_ci` |

### 4.2 数据表清单

#### 独立关系表（4 张）

| 表名 | 用途 | 迁移脚本 |
|------|------|---------|
| `app_configs` | 系统配置 JSON 存储（用户/门店/商品/会员/预约规则等） | 001_persistence_core.sql |
| `cashier_orders` | 收银订单 | 001_persistence_core.sql |
| `recycle_orders` | 回收订单 | 001_persistence_core.sql |
| `recycle_attachments` | 回收附件 | 001_persistence_core.sql |
| `customer_appointments` | 顾客预约 | 002_customer_v1.sql |

#### app_configs 中的 JSON 数据键

| config_key | 内容 | 说明 |
|------------|------|------|
| `org:gold-recycle-v1` | 组织配置 | 租户信息 |
| `stores:gold-recycle-v1` | 门店列表 JSON | 所有门店数据（含经纬度/电话/营业时间） |
| `products:gold-recycle-v1` | 商品列表 JSON | 所有商品数据 |
| `users:gold-recycle-v1` | 员工账号 JSON | 员工账号 + 角色 + 权限 |
| `members:gold-recycle-v1` | 会员列表 JSON | 会员数据 |
| `customer_profiles:gold-recycle-v1` | 顾客档案 JSON | 顾客微信档案 |
| `appointment_rules:gold-recycle-v1` | 预约规则 JSON | 时段/容量/提前量等 |

### 4.3 customer_appointments 表结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | VARCHAR(64) PK | 预约 ID |
| `org_id` | VARCHAR(64) | 组织 ID |
| `customer_id` | VARCHAR(64) | 顾客 ID |
| `customer_name` | VARCHAR(128) | 顾客姓名 |
| `customer_phone` | VARCHAR(32) | 顾客手机号 |
| `store_id` | VARCHAR(64) | 门店 ID |
| `store_name` | VARCHAR(128) | 门店名称 |
| `store_address` | VARCHAR(512) | 门店地址 |
| `store_phone` | VARCHAR(32) | 门店电话 |
| `appointment_date` | DATE | 预约日期 |
| `appointment_time` | VARCHAR(8) | 预约时段 |
| `service_type` | VARCHAR(32) | 服务类型（OLD_FOR_NEW/REPAIR/CONSULT/RECYCLE） |
| `status` | VARCHAR(32) | 状态（PENDING/CONFIRMED/ARRIVED/COMPLETED/CANCELLED/NO_SHOW/TERMINATED） |
| `remark` | TEXT | 顾客备注 |
| `created_at` | DATETIME(6) | 创建时间 |
| `updated_at` | DATETIME(6) | 更新时间 |
| `cancelled_at` | DATETIME(6) NULL | 取消时间 |
| `cancel_reason` | TEXT | 取消原因 |
| `confirmed_at` | DATETIME(6) NULL | 确认时间 |
| `confirmed_by` | VARCHAR(128) | 确认人 |
| `completed_at` | DATETIME(6) NULL | 完成时间 |

### 4.4 迁移脚本

| 脚本 | 说明 | 状态 |
|------|------|------|
| `backend/migrations/001_persistence_core.sql` | 核心表（app_configs/cashier_orders/recycle_orders/recycle_attachments） | ✅ 已执行 |
| `backend/migrations/002_customer_v1.sql` | 顾客预约表 | ✅ 已执行 |

> 注意：生产服务器上 `customer_appointments` 表的 `service_type` 列是手动 ALTER 添加的，迁移脚本已更新包含此列。如果重新执行迁移脚本不会有问题（CREATE TABLE IF NOT EXISTS）。

### 4.5 数据库备份

```bash
# 手动备份
mysqldump -u miniapp_app -p miniapp_prod > /srv/miniapp/backups/db_$(date +%Y%m%d).sql

# 建议设置 crontab 每日备份
# 0 3 * * * mysqldump -u miniapp_app -pPASSWORD miniapp_prod | gzip > /srv/miniapp/backups/db_$(date +\%Y\%m\%d).sql.gz
```

---

## 五、小程序部署

### 5.1 项目清单

| 项目 | 路径 | AppID | 说明 |
|------|------|-------|------|
| 顾客端 | `miniapp-customer/` | wx3f564355bd5c0526 | 顾客浏览/预约/个人中心 |
| 员工端 | `miniapp/` | wx3f564355bd5c0526 | 员工收银/预约管理 |
| 管理后台 | `admin/` | — | Web 管理后台（非小程序） |

### 5.2 顾客端小程序

**项目配置**：
- AppID: `wx3f564355bd5c0526`
- 基础库: 3.7.0
- 主包: 5 页面（home/products/stores/mine/legal/privacy）
- 分包: `pkg-customer/` 7 页面（product-detail/store-detail/appointment-create/appointments/appointment-detail/appointment-notes/recycle-info）
- API 地址: `https://jinjiangguan.com`
- TabBar: 4 Tab（首页/款式/门店/我的），金色选中色 #866E23
- 隐私合规: `__usePrivacyCheck__: true` + 隐私协议页面

**提审步骤**：
1. 在微信开发者工具中打开 `miniapp-customer/` 项目
2. 确认项目配置中 AppID 为 `wx3f564355bd5c0526`
3. 点击"上传" → 填写版本号和备注
4. 登录微信公众平台 → 版本管理 → 提交审核
5. 审核通过后 → 发布上线

**提审前检查**：
- [ ] 微信后台隐私保护指引已更新
- [ ] 微信后台服务类目已添加（黄金珠宝 / 生活服务）
- [ ] `MINIAPP_ALLOW_MOCK_LOGIN=false` 已确认
- [ ] 代码中无硬编码密码
- [ ] app.json 含 `__usePrivacyCheck__: true`
- [ ] 商品页有"到店咨询"标注
- [ ] 回收页有"最终价格以门店线下检测为准"
- [ ] 隐私协议页面可访问

### 5.3 员工端小程序

**项目配置**：
- AppID: `wx3f564355bd5c0526`（与顾客端相同）
- API 地址: `https://jinjiangguan.com`
- 新增页面: `pages/appointments/index` + `pages/appointment-detail/index`
- 新增工具: `utils/appointmentStore.js`

### 5.4 关键风险：AppID 冲突

**问题描述**：顾客端和员工端共用同一 AppID `wx3f564355bd5c0526`，同一 AppID 同时只能有一个线上版本。顾客端"店长入口"使用 `wx.navigateToMiniProgram` 跳转同一 AppID 在真机上会失败。

**解决方案（三选一）**：

| 方案 | 操作 | 优缺点 | 推荐度 |
|------|------|--------|--------|
| A. 申请新 AppID | 为员工端申请新的小程序 AppID | 彻底解决，但需要客户在微信后台注册新小程序 | 推荐 |
| B. 合并为一个项目 | 将员工端页面合并到顾客端项目，用页面路由区分 | 一个 AppID 两个入口，但包体积可能增大 | 可行 |
| C. 顾客端替换员工端 | 顾客端成为唯一线上版本，员工通过"店长入口"进入员工功能 | 需要 N08 改为内部路由跳转 | 最快落地 |

**当前降级处理**：顾客端 `onStaffEntrance` 已有 `fail` 回调，跳转失败时提示用户"请扫描门店管理端小程序码进入"。

**建议**：如果顾客端和员工端需要同时在线上运行，选择方案 A（申请新 AppID）。如果只有一个能上线，选择方案 C（顾客端替换员工端，店长入口改为内部路由）。

### 5.5 顾客端"店长入口"跳转修改（如选方案 C）

如果选择方案 C，需要修改 `miniapp-customer/pages/mine/index.js` 中的 `onStaffEntrance` 方法：

```javascript
// 方案 C：改为内部路由跳转
onStaffEntrance() {
  wx.navigateTo({
    url: '/pkg-customer/staff-login/index'  // 需新建员工登录页
  });
}
```

但这需要将员工端的核心页面（收银/回收/库存等）也合并到顾客端项目中，工作量大。建议优先方案 A。

---

## 六、账号交付清单

### 6.1 服务器账号

| 项目 | 值 | 说明 |
|------|------|------|
| SSH 地址 | `root@8.163.54.124` | root 用户 |
| SSH 密码 | （联系管理员获取） | |
| SSH 端口 | 22 | |

### 6.2 数据库账号

| 项目 | 值 | 说明 |
|------|------|------|
| 数据库 | `miniapp_prod` | 生产数据库 |
| 用户名 | `miniapp_app` | 应用账号 |
| 密码 | （联系管理员获取） | |
| root 密码 | （联系管理员获取） | MariaDB root |

### 6.3 Redis 账号

| 项目 | 值 | 说明 |
|------|------|------|
| 地址 | `127.0.0.1:6379` | 仅本地访问 |
| 密码 | （联系管理员获取） | |

### 6.4 微信公众平台账号

| 项目 | 值 | 说明 |
|------|------|------|
| AppID | `wx3f564355bd5c0526` | 小程序 AppID |
| AppSecret | （联系管理员获取） | 小程序密钥 |
| 管理员账号 | （联系管理员获取） | 微信公众平台登录 |

### 6.5 管理后台账号

管理后台使用后端员工账号体系，首批账号需在数据库或通过 API 创建。默认角色：

| 角色 | 权限 | 说明 |
|------|------|------|
| `boss` | 全部权限 | 老板/管理员 |
| `shop_manager` | 本店管理 + 预约管理 | 店长 |

> 注意：硬编码调试账号密码已在 N11 中移除（`miniapp/utils/userStore.js` 中 `devtoolAccounts` 密码已清空）。生产环境如需测试账号，通过管理后台或数据库直接创建。

### 6.6 域名/SSL

| 项目 | 值 | 说明 |
|------|------|------|
| 域名 | `jinjiangguan.com` | |
| 域名注册商 | （联系管理员获取） | |
| SSL 证书 | Let's Encrypt | 需定期续期 |
| DNS 解析 | A 记录 → 8.163.54.124 | |

---

## 七、源码组织

### 7.1 项目目录结构

```
gold-recycle-miniapp/
├── backend/                     # Go 后端
│   ├── cmd/server/main.go       # 程序入口
│   ├── internal/app/            # 业务逻辑
│   │   ├── app.go               # 路由注册 + 应用初始化
│   │   ├── handlers.go          # 员工端 HTTP handler
│   │   ├── customer_handlers.go # 顾客端 HTTP handler
│   │   ├── customer_middleware.go # 顾客认证中间件
│   │   ├── customer_store.go    # 顾客数据访问层
│   │   ├── customer_types.go    # 顾客类型定义
│   │   ├── admin_customer_handlers.go # 后台内容管理 handler
│   │   ├── admin_customer_store.go    # 后台内容管理数据层
│   │   ├── admin_customer_types.go    # 后台内容管理类型
│   │   ├── admin_store.go       # 管理后台数据访问层
│   │   ├── admin_types.go       # 管理后台类型定义
│   │   ├── store.go             # 员工端数据访问层
│   │   ├── types.go             # 通用类型定义
│   │   ├── persistence.go       # 数据库持久化层
│   │   ├── miniapp_auth.go      # 微信登录认证
│   │   ├── gold_price.go        # 金价服务
│   │   ├── product_import.go    # 商品批量导入
│   │   ├── inventory_material_*.go # 库存物料
│   │   └── *_test.go            # 单元测试
│   ├── migrations/              # 数据库迁移
│   │   ├── 001_persistence_core.sql
│   │   └── 002_customer_v1.sql
│   ├── go.mod
│   └── go.sum
├── miniapp/                     # 员工端小程序
│   ├── app.js / app.json / app.wxss
│   ├── pages/
│   │   ├── home/                # 首页（工作台）
│   │   ├── appointments/        # 预约列表（N08 新增）
│   │   ├── appointment-detail/  # 预约详情（N08 新增）
│   │   ├── login/               # 登录
│   │   ├── cashier/             # 收银
│   │   ├── recycle/             # 回收
│   │   ├── inventory/           # 库存
│   │   ├── members/             # 会员
│   │   └── settings/            # 设置
│   ├── utils/
│   │   ├── config.js            # 配置
│   │   ├── request.js           # 请求封装
│   │   ├── userStore.js         # 用户/权限
│   │   └── appointmentStore.js  # 预约 API（N08 新增）
│   └── assets/                  # 静态资源
├── miniapp-customer/            # 顾客端小程序
│   ├── app.js / app.json / app.wxss
│   ├── pages/
│   │   ├── home/                # 首页
│   │   ├── products/            # 款式列表
│   │   ├── stores/              # 门店列表
│   │   ├── mine/                # 我的
│   │   └── legal/privacy/       # 隐私协议（N11 新增）
│   ├── pkg-customer/            # 分包
│   │   ├── product-detail/      # 商品详情
│   │   ├── store-detail/        # 门店详情
│   │   ├── appointment-create/  # 创建预约
│   │   ├── appointments/        # 我的预约
│   │   ├── appointment-detail/  # 预约详情
│   │   ├── appointment-notes/   # 修改备注
│   │   └── recycle-info/        # 回收介绍
│   ├── utils/
│   │   ├── request.js           # 请求封装
│   │   ├── customer-auth.js     # 顾客认证
│   │   ├── customer-storage.js  # 存储工具
│   │   ├── customer-date.js     # 日期工具
│   │   ├── customer-distance.js # 距离计算
│   │   └── customer-services.js # 服务类型工具
│   └── assets/                  # 静态资源（TabBar 图标等）
├── admin/                       # 管理后台
│   ├── src/
│   │   ├── App.vue              # 主组件（~2500 行，单文件架构）
│   │   ├── api.ts               # API 封装（~1150 行）
│   │   ├── localData.ts         # 本地 mock 数据
│   │   ├── styles.css           # 全局样式
│   │   └── vite-env.d.ts        # 类型声明
│   ├── dist/                    # 构建产物
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
├── docs/customer-v1/            # 顾客端 V1 文档
│   ├── 01_PRD.md                # 产品需求文档
│   ├── 02_SCOPE_BOUNDARY.md     # 范围边界
│   ├── 03_PAGE_SPEC.md          # 页面规格
│   ├── 03_PAGE_INTERACTION_SPEC.md # 交互规格
│   ├── 03_PAGE_DATA_FIELD_SPEC.md  # 数据字段
│   ├── 03_PAGE_ACCEPTANCE_CHECKLIST.md # 验收清单
│   ├── 04_ROLE_PERMISSION_MATRIX.md   # 权限矩阵
│   ├── 05_BUSINESS_FLOW.md      # 业务流程
│   ├── 06_API_CONTRACT_DRAFT.md # 接口文档
│   ├── 07_DATA_MODEL_DRAFT.md   # 数据模型
│   ├── 08_DEVELOPMENT_NODE_PLAN.md    # 开发节点计划
│   ├── 09_PRIVACY_COMPLIANCE.md # 隐私合规
│   ├── 11_ACCEPTANCE_CASES.md   # 验收用例
│   ├── 12_ADMIN_CONFIG_SPEC.md  # 后台配置规格
│   ├── 12_N10_TEST_REPORT.md    # N10 测试报告
│   ├── 13_N11_SUBMISSION_CHECKLIST.md # N11 提审清单
│   ├── 13_UPLOAD_ASSET_SPEC.md  # 上传规格
│   ├── 14_ADMIN_ACCEPTANCE_CHECKLIST.md # 后台验收
│   ├── 14_N12_DEPLOYMENT_DELIVERY.md   # 本文档
│   ├── CHANGE_LOG.md            # 变更日志
│   └── CUSTOMER_CONFIRMATION.md # 客户确认
└── .workbuddy/memory/           # 项目记忆
```

### 7.2 技术栈版本

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.25 | 后端语言 |
| MariaDB | 10.5 | 数据库 |
| Redis | 6.2 | 缓存 |
| Node.js | 22+ | 管理后台构建 |
| Vue | 3.5+ | 管理后台框架 |
| Vite | 8.0+ | 管理后台构建工具 |
| TypeScript | 5.9+ | 管理后台语言 |
| 微信小程序基础库 | 3.7.0 | 小程序运行环境 |
| Nginx | 1.20 | 反向代理 |

### 7.3 依赖清单

**后端（Go）**：
- `github.com/go-sql-driver/mysql` v1.10.0 — MySQL 驱动
- `github.com/redis/go-redis/v9` v9.19.0 — Redis 客户端
- `github.com/xuri/excelize/v2` v2.10.1 — Excel 导入导出
- `golang.org/x/crypto` v0.48.0 — 加密

**管理后台（Node）**：
- `vue` ^3.5.34
- `vite` ^8.0.11
- `vue-tsc` ^3.1.8
- `typescript` ^5.9.3

**小程序**：无 npm 依赖，纯原生开发

---

## 八、API 路由全表

### 8.1 顾客端 API（/api/v1/customer/）

| 路径 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/v1/customer/home` | GET | 公开 | 首页（Banner + 推荐款式 + 附近门店） |
| `/api/v1/customer/products` | GET | 公开 | 商品列表（分页 20 条/页） |
| `/api/v1/customer/products/{id}` | GET | 公开 | 商品详情 |
| `/api/v1/customer/stores` | GET | 公开 | 门店列表 |
| `/api/v1/customer/stores/{id}` | GET | 公开 | 门店详情（含经纬度） |
| `/api/v1/customer/stores/{id}/slots?date=` | GET | 公开 | 时段可用性（09:30-21:30 30 分钟粒度） |
| `/api/v1/customer/recycle-info` | GET | 公开 | 回收服务介绍 |
| `/api/v1/customer/auth/wechat-login` | POST | 公开 | 微信登录 |
| `/api/v1/customer/auth/phone` | POST | 顾客 | 手机号授权 |
| `/api/v1/customer/auth/logout` | POST | 顾客 | 登出 |
| `/api/v1/customer/me` | GET | 顾客 | 个人信息 |
| `/api/v1/customer/appointments` | GET | 顾客 | 预约列表 |
| `/api/v1/customer/appointments` | POST | 顾客 | 创建预约 |
| `/api/v1/customer/appointments/{id}` | GET | 顾客 | 预约详情 |
| `/api/v1/customer/appointments/{id}/cancel` | POST | 顾客 | 取消预约（提前 2 小时） |
| `/api/v1/customer/appointments/{id}/notes` | PUT | 顾客 | 修改备注（仅 PENDING/CONFIRMED） |

### 8.2 员工端 API（/api/v1/staff/）

| 路径 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/v1/staff/appointments` | GET | 员工 | 预约列表（支持状态筛选） |
| `/api/v1/staff/appointments/{id}` | GET | 员工 | 预约详情 |
| `/api/v1/staff/appointments/{id}/status` | PUT | 员工 | 更新预约状态 |

### 8.3 管理后台 API（/api/admin/）

| 路径 | 方法 | 认证 | 说明 |
|------|------|------|------|
| `/api/admin/appointment-rules` | GET | 管理员 | 获取预约规则 |
| `/api/admin/appointment-rules` | PUT | 管理员 | 更新预约规则 |
| `/api/admin/customer-appointments` | GET | 管理员 | 预约列表 |
| `/api/admin/customer-appointments/{id}` | GET | 管理员 | 预约详情 |
| `/api/admin/customer-appointments/{id}/confirm` | POST | 管理员 | 确认预约 |
| `/api/admin/customer-appointments/{id}/arrive` | POST | 管理员 | 标记到店 |
| `/api/admin/customer-appointments/{id}/complete` | POST | 管理员 | 完成服务 |
| `/api/admin/customer-appointments/{id}/no-show` | POST | 管理员 | 标记未到 |
| `/api/admin/customer-appointments/{id}/cancel` | POST | 管理员 | 取消预约 |
| `/api/admin/customer-appointments/{id}/staff-note` | PUT | 管理员 | 内部备注 |
| `/api/admin/stores/{id}/customer-config` | GET | 管理员 | 门店顾客端配置 |
| `/api/admin/stores/{id}/customer-config` | PUT | 管理员 | 更新门店顾客端配置 |
| `/api/admin/customer-home` | GET/PUT | 管理员 | 首页 Banner 配置 |
| `/api/admin/customer-styles` | GET/PUT | 管理员 | 款式展示配置 |

---

## 九、运维指南

### 9.1 常用命令

```bash
# 后端服务管理
systemctl status miniapp-backend    # 查看状态
systemctl restart miniapp-backend   # 重启
systemctl stop miniapp-backend      # 停止
journalctl -u miniapp-backend -f    # 查看日志

# Nginx
nginx -t                             # 检查配置
nginx -s reload                      # 重新加载
tail -f /var/log/nginx/access.log    # 访问日志
tail -f /var/log/nginx/error.log     # 错误日志

# 数据库
mysql -u miniapp_app -p miniapp_prod # 连接数据库
mysqldump -u miniapp_app -p miniapp_prod > backup.sql  # 备份

# Redis
redis-cli -a PASSWORD                # 连接 Redis
redis-cli -a PASSWORD info           # 查看信息
redis-cli -a PASSWORD keys "session:*"  # 查看员工会话
redis-cli -a PASSWORD keys "customer_session:*"  # 查看顾客会话
```

### 9.2 日志位置

| 服务 | 日志路径 |
|------|---------|
| 后端 | `journalctl -u miniapp-backend` |
| Nginx 访问 | `/var/log/nginx/access.log` |
| Nginx 错误 | `/var/log/nginx/error.log` |
| MariaDB | `/var/log/mariadb/mariadb.log` |
| Redis | `/var/log/redis/redis.log` |

### 9.3 健康检查

```bash
# 后端健康检查
curl -s https://jinjiangguan.com/api/v1/customer/home | python3 -m json.tool

# 检查后端进程
ps aux | grep gold-recycle-backend

# 检查端口
ss -tlnp | grep -E '3001|3306|6379|80|443'
```

### 9.4 备份策略

| 备份项 | 频率 | 保留 | 路径 |
|--------|------|------|------|
| 数据库 | 每日 03:00 | 30 天 | `/srv/miniapp/backups/db_*.sql.gz` |
| 后端二进制 | 每次部署 | 30 天 | `/srv/miniapp/backend/backups/` |
| Nginx 配置 | 每次变更 | 手动 | `/etc/nginx/` |
| .env 配置 | 每次变更 | 手动 | `/srv/miniapp/config/` |

### 9.5 SSL 证书续期

```bash
# 检查证书有效期
openssl s_client -connect jinjiangguan.com:443 2>/dev/null | openssl x509 -noout -dates

# Let's Encrypt 自动续期（certbot）
certbot renew --dry-run  # 测试
certbot renew            # 执行续期
```

---

## 十、验收清单

### 10.1 生产环境验收

- [ ] `https://jinjiangguan.com/` 可访问管理后台
- [ ] `https://jinjiangguan.com/api/v1/customer/home` 返回正常数据
- [ ] `https://jinjiangguan.com/api/v1/customer/products` 返回商品列表
- [ ] `https://jinjiangguan.com/api/v1/customer/stores` 返回门店列表
- [ ] `MINIAPP_ALLOW_MOCK_LOGIN=false` 已确认
- [ ] 管理后台可登录（boss/shop_manager 账号）
- [ ] 管理后台可编辑门店经纬度
- [ ] 管理后台可查看预约列表
- [ ] 管理后台可操作预约状态

### 10.2 小程序验收

- [ ] 顾客端小程序可正常打开
- [ ] 首页 Banner 展示正常
- [ ] 款式列表加载正常（分页/分类筛选）
- [ ] 商品详情页展示正常
- [ ] 门店列表加载正常（距离计算）
- [ ] 门店详情页展示正常
- [ ] 创建预约流程完整（选服务/门店/日期/时段/备注/授权）
- [ ] 我的预约列表展示正常
- [ ] 预约详情页展示正常
- [ ] 取消预约功能正常（2 小时限制）
- [ ] 修改备注功能正常
- [ ] 我的页面展示正常（已登录/游客）
- [ ] 隐私协议页面可访问
- [ ] 店长入口可跳转（或降级提示正常）
- [ ] 员工端预约列表页可加载（需登录）
- [ ] 员工端预约详情页可操作状态

### 10.3 安全验收

- [ ] 顾客 token 不可访问员工接口
- [ ] 员工 token 不可访问顾客接口
- [ ] 顾客只能查看/操作自己的预约
- [ ] 无 SQL 注入风险（全参数化查询）
- [ ] 无 XSS 风险（无 rich-text/innerHTML）
- [ ] 全 HTTPS
- [ ] 无硬编码密码
- [ ] 调试账号密码已移除

### 10.4 微信合规验收

- [ ] 隐私保护指引已在微信后台更新
- [ ] 服务类目已添加（黄金珠宝/生活服务）
- [ ] `__usePrivacyCheck__: true` 已配置
- [ ] `requiredPrivateInfos: ["getLocation"]` 已配置
- [ ] 商品页有"到店咨询"标注
- [ ] 回收页有"最终价格以门店线下检测为准"
- [ ] 款式列表有"工费以门店最终确认为准"
- [ ] 隐私协议页面内容完整（6 章节）

---

## 十一、已知问题和风险

### 11.1 AppID 冲突（高风险）

**状态**：未解决
**影响**：顾客端和员工端无法同时在线上运行
**方案**：详见第五节 5.4
**建议**：方案 A（申请新 AppID）

### 11.2 门店经纬度未配置（中风险）

**状态**：待客户操作
**影响**：顾客端门店距离计算和导航功能不可用
**解决**：在管理后台 → 门店管理 → 编辑门店 → 填写经纬度

### 11.3 TabBar 图标为占位图（低风险）

**状态**：临时方案
**影响**：美观度不足
**解决**：设计师提供正式图标后替换 `miniapp-customer/assets/tabbar/` 下的 PNG 文件

### 11.4 真机验证未完成（中风险）

**状态**：188 项验收清单待真机验证
**影响**：可能存在真机兼容性问题
**解决**：在微信开发者工具中逐项验证

---

## 十二、回滚方案

### 12.1 后端回滚

```bash
# 1. 查看备份列表
ls -la /srv/miniapp/backend/backups/

# 2. 恢复旧版本
cd /srv/miniapp/backend/
cp gold-recycle-backend gold-recycle-backend.failed
cp backups/gold-recycle-backend.YYYYMMDDHHMM gold-recycle-backend
chmod +x gold-recycle-backend
systemctl restart miniapp-backend.service

# 3. 数据库回滚（如有需要）
mysql -u miniapp_app -p miniapp_prod < /srv/miniapp/backups/db_YYYYMMDD.sql
```

### 12.2 管理后台回滚

```bash
# Nginx 静态资源回滚（需保留旧版 dist/）
cp -r /usr/share/nginx/jinjiangguan-static.bak/* /usr/share/nginx/jinjiangguan-static/
nginx -s reload
```

### 12.3 小程序回滚

在微信公众平台 → 版本管理中，选择上一个审核通过的版本 → 点击"退回旧版本"。

---

## 十三、交付确认

### 13.1 交付物清单

| 交付物 | 状态 | 说明 |
|--------|------|------|
| 后端源码 | ✅ | `backend/` 目录 |
| 顾客端小程序源码 | ✅ | `miniapp-customer/` 目录 |
| 员工端小程序源码 | ✅ | `miniapp/` 目录 |
| 管理后台源码 | ✅ | `admin/` 目录 |
| 管理后台构建产物 | ✅ | `admin/dist/` 目录 |
| 数据库迁移脚本 | ✅ | `backend/migrations/` |
| 项目文档 | ✅ | `docs/customer-v1/` 20 份文档 |
| 生产环境部署 | ✅ | 后端已部署，管理后台待上传 |
| 服务器账号 | 待交付 | 联系管理员 |
| 数据库账号 | 待交付 | 联系管理员 |
| 微信后台账号 | 待交付 | 联系管理员 |

### 13.2 文档清单

| 文档 | 说明 |
|------|------|
| 01_PRD.md | 产品需求文档 |
| 02_SCOPE_BOUNDARY.md | 范围边界 |
| 03_PAGE_SPEC.md | 页面规格 |
| 03_PAGE_INTERACTION_SPEC.md | 交互规格 |
| 03_PAGE_DATA_FIELD_SPEC.md | 数据字段规格 |
| 03_PAGE_ACCEPTANCE_CHECKLIST.md | 验收清单 |
| 04_ROLE_PERMISSION_MATRIX.md | 权限矩阵 |
| 05_BUSINESS_FLOW.md | 业务流程 |
| 06_API_CONTRACT_DRAFT.md | 接口文档 |
| 07_DATA_MODEL_DRAFT.md | 数据模型 |
| 08_DEVELOPMENT_NODE_PLAN.md | 开发节点计划 |
| 09_PRIVACY_COMPLIANCE.md | 隐私合规 |
| 11_ACCEPTANCE_CASES.md | 验收用例 |
| 12_ADMIN_CONFIG_SPEC.md | 后台配置规格 |
| 12_N10_TEST_REPORT.md | N10 测试报告 |
| 13_N11_SUBMISSION_CHECKLIST.md | N11 提审清单 |
| 13_UPLOAD_ASSET_SPEC.md | 上传规格 |
| 14_ADMIN_ACCEPTANCE_CHECKLIST.md | 后台验收清单 |
| 14_N12_DEPLOYMENT_DELIVERY.md | 本文档（部署交付） |
| CHANGE_LOG.md | 变更日志 |
| CUSTOMER_CONFIRMATION.md | 客户确认 |

### 13.3 待客户操作事项

1. **确认 `MINIAPP_ALLOW_MOCK_LOGIN=false`**：登录服务器 `grep MINIAPP_ALLOW_MOCK_LOGIN /srv/miniapp/config/.env.production`
2. **上传管理后台到生产**：将 `admin/dist/` 内容上传到 `/usr/share/nginx/jinjiangguan-static/`
3. **配置门店经纬度**：在管理后台 → 门店管理 → 编辑每个门店的经纬度
4. **微信后台更新隐私保护指引**：微信公众平台 → 设置 → 服务内容声明 → 用户隐私保护指引
5. **微信后台添加服务类目**：微信公众平台 → 设置 → 基本设置 → 服务类目
6. **解决 AppID 冲突**：申请新 AppID 或选择其他方案
7. **真机验证**：在微信开发者工具中验证 188 项验收清单
8. **提交审核**：在微信开发者工具中上传代码包 → 微信公众平台提交审核
9. **审核通过后发布上线**

---

## 十四、项目里程碑总结

| 节点 | 内容 | 状态 | 完成日期 |
|------|------|------|---------|
| N00 | 备份 | ⏳ 跳过（开发阶段无需） | — |
| N01 | 审计 | ✅ | 2026-08-15 |
| N02 | PRD 冻结 | ✅ | 2026-08-15 |
| N03 | 数据/权限设计 | ✅ | 2026-08-15 |
| N04 | 接口文档 | ✅ | 2026-08-15 |
| N05 | 页面流程验收用例 | ✅ | 2026-08-15 |
| N06 | 后端开发 | ✅ | 2026-08-15 |
| N07 | 顾客端开发 | ✅ | 2026-08-15 |
| N08 | 员工端改造 | ✅ | 2026-08-15 |
| N09 | 管理后台改造 | ✅ | 2026-08-15 |
| N10 | 联调测试安全修复 | ✅ | 2026-08-15 |
| N11 | 微信隐私提审 | ✅ | 2026-08-15 |
| N12 | 上线部署交付 | ✅ 文档完成 | 2026-08-15 |

> N12 文档已完成，实际部署上线需客户确认上述待操作事项后执行。
