# 金匠倌顾客端小程序

顾客端 V1 的历史开发快照。当前正式发布源已统一为同一 AppID 下的 `../miniapp/`，本目录不再作为独立小程序上传。

## 项目结构

```
miniapp-customer/
├── app.js / app.json / app.wxss     # 主包入口（4 TabBar + 全局样式）
├── pages/                            # 主包 Tab 页面
│   ├── home/                         # 首页（Tab 1）✅ 子阶段1
│   ├── products/                     # 款式（Tab 2）✅ 子阶段2：分类筛选+双列卡片+分页
│   ├── stores/                       # 门店（Tab 3）✅ 子阶段2：定位排序+卡片+拨号
│   └── mine/                         # 我的（Tab 4）⏳ 子阶段3
├── pkg-customer/                     # 分包子页面
│   ├── product-detail/               # 款式详情 ✅ 子阶段2
│   ├── store-detail/                 # 门店详情 ✅ 子阶段2（导航/拨号/预约）
│   ├── appointment-create/           # 预约创建 ⏳ 占位页（子阶段3完整实现）
│   ├── appointments/                 # 我的预约列表 ⏳ 子阶段3
│   ├── appointment-detail/           # 预约详情 ⏳ 子阶段3
│   ├── appointment-notes/            # 修改备注 ⏳ 子阶段3
│   └── recycle-info/                 # 回收介绍 ⏳ 子阶段3
├── utils/                            # 工具类
│   ├── request.js                    # 网络请求封装（14 个 API）
│   ├── customer-auth.js              # 顾客登录管理
│   ├── customer-date.js              # 日期工具
│   ├── customer-distance.js          # Haversine 距离计算
│   └── customer-services.js          # 服务类型/状态/金额格式化/图片绝对地址
└── assets/tabbar/                    # TabBar 图标
```

## 开发进度

| 子阶段 | 内容 | 状态 |
| --- | --- | --- |
| 1 | 骨架 + 公共层 + 首页 | ✅ |
| 2 | 款式 Tab + 门店 Tab（含详情页） | ✅ |
| 3 | 预约模块 + 我的 Tab | ⏳ 待开发 |

## 已知数据约定（与生产 API 对齐）

- 商品列表接口返回 `categories`（首屏取一次即可）；分页 20 条/页
- 商品 `imageUrl` 为相对路径（`/assets/...`），前端用 `absUrl()` 补全域名
- 门店列表接口不含经纬度，门店页并发拉详情补齐后计算距离
- 生产门店经纬度尚未配置（均为 0），导航会提示"暂未配置位置"，需在管理后台编辑

## 历史快照启动（仅用于对照）

1. 正式开发和上传请在微信开发者工具打开 `../miniapp/`
2. 本目录只用于对照历史顾客端页面，不作为独立 AppID 运行
3. 顾客首页为 `../miniapp/pages/customer-home/index`；员工页面仍保留在 `../miniapp/pages/home/index`

## 与员工端的关系

- 顾客端与员工端共用原小程序 AppID，正式发布源为 `../miniapp/`
- 小程序启动首页是顾客首页；底部「我的」→「店长入口」通过同一小程序内部路由进入 `/pages/account/index`
- 店长登录成功后进入原员工工作台 `/pages/home/index`；顾客端不直接展示员工工作台数据
- 禁止把 `miniapp-customer/` 作为独立小程序上传，否则会产生双 AppID、双入口和权限边界不一致
