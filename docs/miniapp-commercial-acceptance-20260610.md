# 金匠倌小程序商用交付验收标准

日期：2026-06-10  
范围：微信小程序 `miniapp`，后端 `backend`，生产接口默认域名 `https://jinjiangguan.com`  
目标：今天交付给客户使用，所有页面必须真实接后端、业务逻辑可闭环、交互与状态管理不丢数据。

## 1. 总体验收红线

1. 生产模式不得使用 demo 门店、demo 联系人、mock 订单、本地种子数据作为业务提交来源。
2. 所有需要登录的页面必须使用真实登录态，登录态来自 `/api/v1/auth/login`、`/api/v1/auth/wechat-login`、`/api/v1/me`。
3. 所有门店归属必须来自 `/api/v1/stores` 或登录返回的后端门店数据；无真实 `storeId` 时禁止创建收银单、回收单和图片上传。
4. 表单输入、切页、上传照片、接口失败时，用户已填内容不得被清空；草稿必须按当前账号隔离保存。
5. 设置、门店资料、协议页不得展示演示联系人；客户资料必须来自 `/api/v1/settings` 或显示“请在后台配置”类提示。
6. 图片上传必须走正式上传链路：prepare -> 上传对象存储 -> complete -> 订单确认；上传后页面必须能立即显示缩略图并可预览。
7. 接口失败必须给用户可理解提示，不得静默保存为假数据，不得让用户误以为已经提交成功。
8. 后台、小程序必须使用同一套业务数据源；后台维护的会员、商品、门店、设置，小程序刷新后能同步看到。

## 2. P0 验收项

| 编号 | 问题 | 必须通过的标准 | 本次代码状态 |
| --- | --- | --- | --- |
| P0-1 | 收银单 500，门店未绑定真实后端数据 | 收银页进入时刷新 `/api/v1/me` 和 `/api/v1/stores`；提交 `/api/v1/cashier/orders` 必须带真实 `storeId`；无真实门店时阻止提交并提示重新登录 | 已修复，生产模式不再使用 demo storeId 兜底 |
| P0-2 | 回收录单上传照片后清空全部表单 | 回收录单输入、选择品类/成色、选择照片时实时保存草稿；上传或跳转确认页后客户姓名、手机号、克重、品名、价格、备注不得丢失 | 已修复，`saveDraft` 覆盖输入和照片变更 |
| P0-3 | 回收照片上传后看不到，还要重复上传 | 选择照片后页面立即显示缩略图；点击缩略图可预览；提交时使用本地照片路径完成正式上传并回写公网地址 | 已修复，页面已补照片列表和 `previewPhoto` |
| P0-4 | 设置页/门店资料显示演示客户资料 | 生产配置不再内置演示联系人、演示门店；资料页优先读取 `/api/v1/settings`；接口未配置时显示后台配置提示 | 已修复，默认演示联系人/门店已清空 |

## 3. 页面级商用验收标准

| 页面 | 后端真实接口 | 业务逻辑验收 | 交互与状态验收 | 通过标准 |
| --- | --- | --- | --- | --- |
| `pages/account/index` 登录页 | `POST /api/v1/auth/login`、`POST /api/v1/auth/wechat-login`、`POST /api/v1/auth/logout` | 账号密码登录、微信登录、手机号绑定失败提示明确；登录成功后保存 token、角色、权限、门店 | 必须勾选协议；登录失败不清空账号；退出登录清空当前账号草稿和缓存 | 真账号可登录；未绑定微信提示联系管理员；无 token 不能进入业务提交 |
| `pages/home/index` 首页 | `GET /api/v1/dashboard/summary`、`GET /api/v1/recycle/orders`、`GET /api/v1/cashier/orders`、`GET /api/v1/gold-prices/reference` | 展示今日单量、金额、待处理、门店范围、实时金价 | 每次进入刷新登录态和门店绑定；无权限入口不可用 | 数据和后台订单一致；老板看全部门店，员工按权限范围显示 |
| `pages/order/index` 收银页 | `GET /api/v1/products`、`POST /api/v1/cashier/orders`、`GET /api/v1/me`、`GET /api/v1/stores` | SKU 查询商品，带出名称、售价、库存；客户和商品必填；提交创建真实收银单 | 输入实时保存草稿；切页返回不丢；提交中按钮防重复；失败保留表单 | 后端产生收银单；无真实门店时不能提交；500 不再由 demo storeId 触发 |
| `pages/track/index` 回收录单页 | `POST /api/v1/recycle/quote-preview`、`POST /api/v1/uploads/recycle-photos/prepare`、`POST /api/v1/uploads/recycle-photos/complete`、`POST /api/v1/recycle/orders`、`POST /api/v1/recycle/orders/:id/confirm` | 客户、手机号、品名、毛重、扣重、成色、照片完整后才能确认；报价以服务端为准 | 输入实时存草稿；上传/选择照片不清空表单；照片立即显示并可预览；失败保留草稿 | 3 张照片可正式上传，确认后生成回收单；无真实门店或对象存储未配置时明确阻止 |
| `pages/orders/index` 订单列表 | `GET /api/v1/recycle/orders`、`GET /api/v1/cashier/orders` | 合并展示收银单和回收单，状态、金额、客户、门店正确 | 下拉/进入刷新；点击进入详情；接口失败提示但不伪造新数据 | 与后台订单列表数量和核心字段一致 |
| `pages/order-detail/index` 订单详情/回收确认 | `GET /api/v1/recycle/orders/:id`、`GET /api/v1/cashier/orders/:id`、`POST /api/v1/recycle/orders/:id/confirm` | 根据类型展示收银/回收详情；草稿确认走正式回收确认链路 | 从草稿进入不丢照片；确认中防重复；失败可返回继续修改 | 详情字段与后端一致，回收确认后状态变为已完成 |
| `pages/members/index` 会员页 | `GET /api/v1/members`、`POST /api/v1/members`、`PUT /api/v1/members/:id` | 会员列表、新增、编辑、标签、备注写入后端 | 表单校验手机号/姓名；保存中防重复；失败不清表单 | 后台新增/编辑会员后，小程序刷新同步 |
| `pages/products/index` 商品页 | `GET /api/v1/products` | 商品按状态、库存、SKU 展示；收银页可查询同一商品 | 刷新后列表同步；空列表提示去后台维护 | 后台商品变更后小程序展示一致 |
| `pages/product-detail/index` 商品详情 | `GET /api/v1/products/:id` | 展示 SKU、价格、成色、库存、状态等后端字段 | 返回列表状态不乱；接口失败提示 | 详情字段与后台商品一致 |
| `pages/product-form/index` 商品表单 | `POST /api/v1/products`、`PUT /api/v1/products/:id` | 新增/编辑商品写入后端，SKU、价格、状态可维护 | 必填校验；保存中防重复；保存后返回列表刷新 | 后台和小程序均能看到新增/编辑结果 |
| `pages/profile/index` 设置首页 | `GET /api/v1/me`、`GET /api/v1/stores`、`GET /api/v1/members`、`GET /api/v1/products`、`GET /api/v1/recycle/orders`、`GET /api/v1/cashier/orders` | 展示当前角色、门店、订单/会员/商品统计；提供业务入口 | 进入时刷新真实门店绑定；清空草稿只影响当前账号 | 统计与后端数据同源，不显示演示门店 |
| `pages/store-info/index` 门店资料 | `GET /api/v1/settings` | 展示小程序名称、服务范围、客服电话、当前门店、门店编码 | 接口失败使用已缓存真实配置；无配置显示后台补充提示 | 不出现演示联系人、演示电话、demo 门店 |
| `pages/legal/privacy` 隐私协议 | `GET /api/v1/settings` | 协议内客户主体/服务电话读取后端配置 | 配置缺失时不显示演示主体 | 客户资料和设置页一致 |
| `pages/legal/terms` 用户协议 | `GET /api/v1/settings` | 协议内服务主体读取后端配置 | 配置缺失时不显示演示主体 | 客户资料和设置页一致 |

## 4. 状态管理验收

| 状态对象 | 存储位置 | 验收标准 |
| --- | --- | --- |
| 登录态 | `gr_operator_profile` | token、角色、权限、门店来自后端；退出登录清空；无 token 不允许提交业务 |
| 收银/回收草稿 | `gr_recycle_draft`，按账号/门店 scope 隔离 | 输入后立即保存；切页、上传照片、接口失败不丢；提交成功后清空 |
| 订单缓存 | `gr_recycle_orders`、`gr_cashier_orders`，按账号/门店 scope 隔离 | 只作为后端列表缓存；生产模式不得凭缓存伪造提交成功 |
| 会员/商品缓存 | `gr_member_records`、`gr_product_records` | 只缓存最近一次 API 快照；生产模式新增/编辑必须写后端 |
| 系统配置缓存 | `gr_system_settings` | 优先 `/api/v1/settings`；接口失败使用上次真实缓存；不得回退演示联系人 |

## 5. 后端接口验收

| 接口 | 验收动作 | 通过标准 |
| --- | --- | --- |
| `/health` | 访问健康检查 | 返回服务名和正常状态 |
| `/api/v1/auth/login` | 用后台员工账号登录 | 返回 token、用户、权限、门店范围 |
| `/api/v1/me` | 登录后刷新用户 | 返回当前真实员工和权限 |
| `/api/v1/stores` | 登录后刷新门店 | 返回该账号可见真实门店；至少一个真实门店才能提交收银/回收 |
| `/api/v1/settings` | 设置页读取 | 返回客户品牌、电话、回收规则、存储配置等 |
| `/api/v1/products` | 商品页/收银 SKU 查询 | 返回后台维护商品，SKU 可在收银页命中 |
| `/api/v1/cashier/orders` | 创建/查询收银单 | 创建成功返回订单号；列表和详情可查 |
| `/api/v1/recycle/quote-preview` | 回收报价 | 返回服务端价格、净重、金额 |
| `/api/v1/uploads/recycle-photos/prepare` | 准备上传照片 | 返回可用上传地址、上传模式、uploadId |
| `/api/v1/uploads/recycle-photos/complete` | 完成上传 | 返回公网图片地址或缩略图地址 |
| `/api/v1/recycle/orders` | 创建/查询回收单 | 草稿创建成功；确认后状态更新 |
| `/api/v1/recycle/orders/:id/confirm` | 照片上传后确认 | 返回已确认状态，附件地址可在详情查看 |
| `/api/v1/members` | 会员新增/查询 | 新增后列表可查，后台同源 |

## 6. 今天交付执行结果

| 验收项 | 命令/方式 | 结果 |
| --- | --- | --- |
| 小程序静态结构与语法 | `cd miniapp && npm run validate` | 通过，14 个页面和核心 JS 语法均通过 |
| 后端单元测试 | `cd backend && env GOCACHE=/private/tmp/gold-recycle-go-build-cache go test ./...` | 通过，`backend/internal/app` 测试通过 |
| 生产演示信息扫描 | `rg "demo|示例|store-demo|13800000000|support@example" miniapp/pages miniapp/utils` | 通过生产展示口径；仅剩旧缓存迁移判断中的历史字符串，会被迁移为“本地测试”文案 |
| 体验版上传 | 微信开发者工具 CLI `upload --version 1.0.1` | 通过，使用 AppID `wx3f564355bd5c0526`，包体约 1.6 MB，CLI 返回 `✔ upload` |
| 预览二维码 | 微信开发者工具 CLI `preview --qr-format image` | 通过，二维码已生成到 `release/acceptance-20260610/preview-qrcode.png` |
| 验收视频 | 微信开发者工具真机预览或模拟器录屏 | 待录制；需按第 8 节脚本跑完整真实业务流程 |

## 7. 上传命令

微信开发者工具开启路径：`微信开发者工具 -> 设置 -> 安全设置 -> 服务端口`。

服务端口开启后执行：

```bash
'/Applications/微信开发者工具.app/Contents/MacOS/cli' upload \
  --project '/Users/xiaoliao/Desktop/openclaw/workspace/projects/gold-recycle-miniapp/miniapp' \
  --version '1.0.1' \
  --desc '商用验收修复：收银真实门店、回收录单照片与设置真实配置'
```

本轮上传结果：

```text
使用 AppID: wx3f564355bd5c0526
TOTAL: 1.6 MB
✔ upload
```

预览二维码路径：

```text
release/acceptance-20260610/preview-qrcode.png
```

## 8. 验收视频脚本

录屏内容按以下顺序，视频中每一步必须可见：

1. 登录页：勾选协议，使用真实员工账号或体验版微信登录成功。
2. 首页：展示真实门店/角色/今日数据，不出现演示门店。
3. 收银页：输入客户信息，输入真实 SKU，自动带出商品，提交成功生成收银单。
4. 订单页：看到刚提交的收银单，进入详情字段正确。
5. 回收录单页：填写客户、手机号、品名、重量、扣重，选择 3 张照片；照片缩略图立即显示并可预览；跳转确认后信息不丢。
6. 回收详情：确认留档成功，订单详情能看到照片/附件状态。
7. 设置页/门店资料：显示真实后端配置或后台配置提示，不出现演示联系人。
