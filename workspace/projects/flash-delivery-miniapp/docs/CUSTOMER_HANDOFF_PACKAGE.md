# 疾蜂同城急送小程序客户交付打包清单

这份文档只解决一件事：把当前模板源码整理成一个可以直接交给客户或接手开发方的交付包。

适用范围：

- 当前项目定位是 `模板源码交付`
- 当前项目可 `导入微信开发者工具演示`
- 当前项目默认不是 `可直接商用上线`

当前真实前提：

- `project.config.json` 当前使用 `touristappid`
- `utils/config.js` 当前是 `mode: "mock"`
- `utils/config.js` 当前 `apiBaseUrl` 仍是占位值 `https://api.your-domain.com`

## 交付前先做 5 分钟检查

按顺序执行，不要跳。

1. 进入项目根目录：

```bash
cd /Users/xiaoliao/Desktop/openclaw/workspace/projects/flash-delivery-miniapp
```

2. 跑一次校验：

```bash
npm run validate
```

预期结果：

- 命令通过
- 输出包含 `Mini program validation passed.`

3. 打开并确认以下事实仍然成立：

- `project.config.json` 里的 `appid` 还是 `touristappid` 或客户指定测试 AppID
- `utils/config.js` 里的 `mode` 是否符合本次交付口径
- `utils/config.js` 里的品牌名、电话、协议地址是否仍是模板默认值

4. 确认本次只交付模板源码，不承诺正式上线：

- 如果仍然是 `touristappid + mock`，交付说明里必须写明“仅供演示/二开”
- 不要在口头或文档里写“已可直接上线”

5. 确认压缩包里不带敏感文件：

- 不带账号密码文本
- 不带云服务器登录信息
- 不带正式支付参数
- 不带任何私钥、证书、密钥文件

## 客户交付包目录怎么组

建议先在本地临时创建一个交付目录，例如：

```text
flash-delivery-miniapp-handoff-2026-05-30/
```

交付目录内只放下面这些内容：

```text
flash-delivery-miniapp-handoff-2026-05-30/
├── source/
├── docs/
└── README-先看这个.txt
```

### `source/` 里放什么

把项目源码复制进去，但按下面规则处理。

必须保留：

- `app.js`
- `app.json`
- `app.wxss`
- `pages/`
- `utils/`
- `assets/`
- `scripts/`
- `project.config.json`
- `package.json`
- `README.md`

不要放进去：

- `node_modules/`
- `.git/`
- 本机 IDE 缓存文件
- `.DS_Store`
- 任何临时录屏、个人截图、聊天记录

实操建议：

1. 复制项目为一个临时目录。
2. 删除 `node_modules/`。
3. 删除 `.git/` 和其他本机垃圾文件。
4. 再放进交付包的 `source/`。

交付给客户时要说明：

- 客户收到源码后先执行 `npm install`
- 然后导入微信开发者工具
- 再按文档改品牌和配置

### `docs/` 里必须附带哪些文档

至少放这 7 份：

- `docs/CUSTOMER_HANDOFF_PACKAGE.md`
- `docs/CUSTOMER_INTEGRATION_GUIDE.md`
- `docs/DEMO_RECORDING_GUIDE.md`
- `docs/CURRENT_STATUS_2026-05-30.md`
- `docs/COMMERCIAL_GO_NO_GO_CHECKLIST.md`
- `docs/OWNER_TEST_GUIDE.md`
- `README.md`

如果客户会继续做商用联调，再额外附带：

- `docs/COMMERCIAL_LAUNCH_CHECKLIST.md`
- `docs/COMMERCIAL_DELIVERY_STANDARD.md`
- `08_go_live/` 下与上线准备有关的资料

### `README-先看这个.txt` 里建议写什么

建议只写 6 行，客户先看这个再看源码：

```text
1. 这是微信小程序模板源码交付包，不是已上线成品。
2. 当前可直接导入微信开发者工具演示。
3. 当前默认配置是 touristappid + mock 模式。
4. 如需接正式后端，请先看 docs/CUSTOMER_INTEGRATION_GUIDE.md。
5. 如需判断能否商用上线，请先看 docs/COMMERCIAL_GO_NO_GO_CHECKLIST.md。
6. 如需客户自己替换品牌、电话、协议地址，请先看本包 docs/CUSTOMER_HANDOFF_PACKAGE.md。
```

## 打包时建议附带的演示说明

交付包里不一定要塞视频文件，但必须告诉客户怎么演示。

直接附这段说明即可：

1. 打开微信开发者工具。
2. 导入 `source/` 里的项目目录。
3. 使用 `touristappid` 或客户自己的测试 AppID 编译。
4. 首页展示品牌、常用急送、价格预估。
5. 进入“我的”页，点击测试资料填充并保存登录。
6. 进入“下单”页，填写寄件信息并提交订单。
7. 进入订单详情、配送跟踪、订单列表，完成 mock 演示闭环。

如果客户要视频版演示，按 `docs/DEMO_RECORDING_GUIDE.md` 录即可。

## 客户收到后第一批要改的内容

这部分要写清楚，不然客户会问很多来回问题。

客户第一批通常只需要改 4 类内容。

### 1. 品牌文案

优先检查：

- `utils/config.js` 里的 `appName`
- `utils/config.js` 里的 `brandSlogan`
- 页面里的品牌标题、服务文案、按钮文案

### 2. 联系方式和协议地址

优先检查：

- `utils/config.js` 里的 `servicePhone`
- `utils/config.js` 里的 `privacyUrl`
- `utils/config.js` 里的 `userAgreementUrl`

### 3. 小程序接入配置

优先检查：

- `project.config.json` 里的 `appid`
- `utils/config.js` 里的 `mode`
- `utils/config.js` 里的 `apiBaseUrl`
- `utils/config.js` 里的 `tenantId`

替换原则：

- 演示阶段：可以继续用 `touristappid + mock`
- 联调阶段：改成客户测试 AppID + 客户测试 API 域名
- 正式上线阶段：改成正式 AppID + 正式 HTTPS 域名 + 正式后端

### 4. 业务字段和流程

客户常见会改：

- 物品类型
- 重量区间
- 价格说明
- 配送承诺文案
- 下单表单字段
- 订单状态文案

如果客户要改这些，不要直接说“只改文字就行”。
要明确说明：这部分可能涉及页面文案、校验规则、接口字段、订单状态映射一起改。

## 哪些敏感信息不能放仓库

下面这些内容严禁提交到公开或共享仓库，也不要塞进交付压缩包。

- 微信小程序正式 AppID 对应的管理员账号密码
- 微信公众平台登录信息
- 云服务器 SSH 账号、密码、私钥
- 数据库账号密码
- Redis 密码
- 短信平台密钥
- 地图平台正式 Key
- 微信支付商户号密钥、API v3 Key、证书
- 生产环境 JWT 密钥、Webhook Secret
- 任何正式环境 `.env` 真值

允许放进仓库的只能是：

- 示例值
- 占位值
- `.example` 模板
- 非敏感的接入说明

正确做法：

1. 仓库里只保留示例配置。
2. 正式密钥只通过私下安全渠道单独交接。
3. 文档里写“谁负责提供、什么时候提供、填到哪里”，不要把真值写进去。

## 如果客户接下来要正式上线，还缺什么

这部分必须说具体，不要只写“还要后端和测试”。

当前从模板演示走到正式上线，至少还缺下面这些：

1. 正式微信小程序 AppID
2. 微信公众平台主体资料和提审主体
3. 正式 HTTPS API 域名
4. request 合法域名配置
5. 真实后端接口，不再依赖 mock
6. 登录鉴权和手机号/微信身份体系
7. 地址解析、距离计算、地图能力
8. 支付、退款、订单取消规则
9. 客服入口、投诉入口、售后流程
10. 隐私协议、用户协议、采集说明
11. 真机联调和异常场景测试
12. 提审材料、截图、类目、测试账号

对外口径建议直接写成这句话：

> 当前交付的是可演示、可二开的模板源码；若要正式商用上线，还需补齐正式 AppID、正式域名、真实后端、支付、地图、客服、合规资料和真机验收。

## 推荐交付动作顺序

按这个顺序发给客户，最省沟通成本。

1. 先发压缩包。
2. 再发一句交付说明：这是模板源码交付，不是直接上线包。
3. 提醒客户先看 `README-先看这个.txt`。
4. 再看 `docs/CUSTOMER_HANDOFF_PACKAGE.md`。
5. 如果客户要自己接后端，再看 `docs/CUSTOMER_INTEGRATION_GUIDE.md`。
6. 如果客户问“能不能直接上线”，直接让他看 `docs/COMMERCIAL_GO_NO_GO_CHECKLIST.md`。

## 交付时建议直接复制给客户的话术

```text
本次交付内容为：微信小程序模板源码 + 接入说明 + 演示说明。
当前版本可直接导入微信开发者工具进行演示，也适合作为二开模板继续开发。
请注意：当前默认仍是 touristappid 和 mock 模式，不代表已经具备正式商用上线条件。
如果您下一步要接正式后端、支付、地图、客服和提审上线，请按文档里的上线缺口清单逐项补齐。
```

## 交付完成标准

满足下面 6 条，才算这次“交付包整理”完成：

- `npm run validate` 已通过
- 交付包里包含源码和必需文档
- 压缩包里不含 `node_modules`
- 压缩包里不含任何敏感信息
- 文档里已明确“当前是模板源码，不是直接上线包”
- 客户收到后知道先改哪些文件、先看哪些文档
