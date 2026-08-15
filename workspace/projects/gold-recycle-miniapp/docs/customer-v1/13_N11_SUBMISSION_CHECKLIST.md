# N11 - 微信隐私提审清单

> 生成日期：2026-08-15
> 目标：完成隐私协议配置、审核合规标注、提审资料准备
> 依赖：N10 联调测试安全修复已完成

---

## 一、已完成事项

### 1.1 隐私配置

| 项目 | 文件 | 状态 |
|------|------|------|
| `__usePrivacyCheck__: true` | `miniapp-customer/app.json` | ✅ 已添加 |
| `requiredPrivateInfos: ["getLocation"]` | `miniapp-customer/app.json` | ✅ 已有 |
| `permission.scope.userLocation.desc` | `miniapp-customer/app.json` | ✅ 已有 |
| 隐私协议页面 | `miniapp-customer/pages/legal/privacy.*` | ✅ 已创建 |
| "我的"页面隐私入口 | `miniapp-customer/pages/mine/index.wxml` | ✅ 已添加 |

### 1.2 审核合规标注

| 标注 | 位置 | 状态 |
|------|------|------|
| "工费以门店最终确认为准，线上仅作参考" | 款式列表页 notice-banner | ✅ 已有 |
| "工费以店铺确认为准" | 款式详情页 d-price-tip | ✅ 已有 |
| "商品仅供参考，到店咨询详情" | 款式详情页 d-cta-disclaimer | ✅ 已添加 |
| "最终回收价格以门店线下检测为准" | 回收介绍页 ri-cta-disclaimer | ✅ 已添加 |
| "预约后门店将为您安排专属服务" | 回收介绍页 ri-cta-tip | ✅ 已有 |

### 1.3 安全修复

| 项目 | 文件 | 状态 |
|------|------|------|
| 移除硬编码调试账号密码 | `miniapp/utils/userStore.js` | ✅ 密码已清空 |
| 预约页面登录守卫 | `miniapp/pages/appointments/index.js` | ✅ N10 已修复 |
| 预约详情页登录守卫 | `miniapp/pages/appointment-detail/index.js` | ✅ N10 已修复 |
| 预约详情页禁用分享 | `miniapp/pages/appointment-detail/index.js` | ✅ N10 已修复 |

### 1.4 隐私协议内容

隐私协议页面（`pages/legal/privacy`）包含以下信息收集声明：

1. **位置信息** — 用于计算与门店距离，不上传服务器，可拒绝
2. **手机号** — 用于预约联系，通过微信能力获取，存储于服务器
3. **微信昵称和头像** — 用于展示顾客个人信息
4. **预约信息** — 门店/时间/服务类型/备注，保留 2 年

---

## 二、需在微信后台操作的事项

### 2.1 用户隐私保护指引（优先级：高）

在微信公众平台 → 设置 → 服务内容声明 → 用户隐私保护指引中，按以下内容填写：

```
1. 位置信息
   - 使用目的：用于计算您与门店的距离，推荐最近门店
   - 使用场景：进入门店列表页或预约页时
   - 是否上传服务器：否，仅在本地计算

2. 手机号
   - 使用目的：用于到店预约联系
   - 使用场景：提交预约时授权
   - 存储方式：存储于服务器

3. 微信昵称和头像
   - 使用目的：展示顾客个人信息
   - 使用场景：微信登录时获取
   - 存储方式：存储于服务器

4. 预约信息
   - 收集内容：预约门店、预约时间、服务类型、备注
   - 使用目的：到店预约服务
   - 保存期限：预约完成后保留 2 年
```

### 2.2 服务类目（优先级：高）

当前服务类目需确认是否包含以下类目：

| 功能 | 建议类目 | 说明 |
|------|---------|------|
| 商品展示 | 商家自营 → 黄金珠宝 | 展示黄金饰品，不涉及线上交易 |
| 到店预约 | 生活服务 → 珠宝首饰 | 纯预约服务，不涉及支付 |

**操作**：在微信公众平台 → 设置 → 基本设置 → 服务类目中添加上述类目。

### 2.3 小程序基本信息（优先级：中）

确认以下信息是否需要更新：

- 小程序简介：建议包含"黄金回收、到店预约"
- 小程序截图：需包含顾客端首页、款式列表、门店列表、预约页面
- 类目标签：黄金珠宝、生活服务

---

## 三、关键风险：AppID 冲突

### 问题描述

顾客端（`miniapp-customer/`）和员工端（`miniapp/`）**共用同一 AppID `wx3f564355bd5c0526`**。这意味着：

1. 同一 AppID 同时只能有一个线上版本
2. 顾客端"店长入口"使用 `wx.navigateToMiniProgram` 跳转同一 AppID → **在真机上会失败**
3. 提审时只能提交一个项目

### 解决方案（三选一）

| 方案 | 操作 | 优缺点 |
|------|------|--------|
| **A. 申请新 AppID** | 为员工端申请新的小程序 AppID | 彻底解决，但需要客户在微信后台注册新小程序 |
| **B. 合并为一个项目** | 将员工端页面合并到顾客端项目，用页面路由区分 | 一个 AppID 两个入口，但包体积可能增大 |
| **C. 顾客端替换员工端** | 顾客端成为唯一线上版本，员工通过顾客端"店长入口"进入员工功能 | 需要 N08 改为内部路由跳转而非 navigateToMiniProgram |

**建议**：方案 A（申请新 AppID）最干净，方案 C 最快落地。

### 当前降级处理

顾客端 `onStaffEntrance` 已有 `fail` 回调，跳转失败时提示用户"请扫描门店管理端小程序码进入"。

---

## 四、提审资料清单

### 4.1 代码包

| 项目 | 路径 | AppID | 说明 |
|------|------|-------|------|
| 顾客端 | `miniapp-customer/` | wx3f564355bd5c0526 | 提审主包 |

### 4.2 测试账号

| 角色 | 用户名 | 密码 | 说明 |
|------|--------|------|------|
| 老板 | boss | （联系管理员获取） | 有全部权限 |
| 店长 | manager.sz | （联系管理员获取） | 有预约管理权限 |

> 注意：硬编码调试密码已从代码中移除，测试时使用后端真实账号。

### 4.3 审核重点说明

供审核人员参考的功能说明：

1. **商品展示**：本小程序不涉及在线交易，所有商品仅供展示参考，到店咨询详情
2. **预约功能**：纯到店预约服务，不涉及支付，顾客到店后线下完成交易
3. **位置权限**：仅用于计算与门店距离，不上传服务器，拒绝授权可手动选店
4. **手机号授权**：仅预约时需要，浏览商品和门店不需要授权
5. **店长入口**：在"我的"页面底部，需账号密码登录，不对公众开放

### 4.4 审核前检查清单

- [ ] 微信后台隐私保护指引已更新
- [ ] 微信后台服务类目已添加（黄金珠宝 / 生活服务）
- [ ] `MINIAPP_ALLOW_MOCK_LOGIN=false` 已确认（生产环境配置）
- [ ] 代码中无硬编码密码
- [ ] 顾客端 app.json 含 `__usePrivacyCheck__: true`
- [ ] 顾客端 app.json 含 `requiredPrivateInfos: ["getLocation"]`
- [ | 商品页有"到店咨询"标注
- [ ] 回收页有"最终价格以门店线下检测为准"
- [ ] 款式列表有"工费以门店最终确认为准"
- [ ] 隐私协议页面可访问
- [ ] AppID 冲突问题已解决（三选一）

---

## 五、修改文件清单

| 文件 | 改动类型 | 说明 |
|------|---------|------|
| `miniapp-customer/app.json` | 修改 | 添加 `__usePrivacyCheck__: true` + 注册 privacy 页面 |
| `miniapp-customer/pages/legal/privacy.js` | 新增 | 隐私协议页面逻辑 |
| `miniapp-customer/pages/legal/privacy.json` | 新增 | 隐私协议页面配置 |
| `miniapp-customer/pages/legal/privacy.wxml` | 新增 | 隐私协议页面结构（6 章节） |
| `miniapp-customer/pages/legal/privacy.wxss` | 新增 | 隐私协议页面样式 |
| `miniapp-customer/pages/mine/index.wxml` | 修改 | 添加隐私协议入口 |
| `miniapp-customer/pages/mine/index.js` | 修改 | 添加 `onPrivacy` 方法 |
| `miniapp-customer/pkg-customer/product-detail/index.wxml` | 修改 | 添加"到店咨询"标注 |
| `miniapp-customer/pkg-customer/product-detail/index.wxss` | 修改 | 添加 disclaimer 样式 |
| `miniapp-customer/pkg-customer/recycle-info/index.wxml` | 修改 | 添加"最终价格以门店线下检测为准" |
| `miniapp-customer/pkg-customer/recycle-info/index.wxss` | 修改 | 添加 disclaimer 样式 |
| `miniapp/utils/userStore.js` | 修改 | 移除硬编码调试密码（清空为空字符串） |

---

## 六、下一步

### N12 上线部署（待 N11 审核通过后）

1. 确认生产环境配置（`MINIAPP_ALLOW_MOCK_LOGIN=false`）
2. 部署最终代码包到微信开发者工具
3. 提交审核
4. 审核通过后发布上线
5. 交付：源码、数据库说明、账号信息、运维文档
