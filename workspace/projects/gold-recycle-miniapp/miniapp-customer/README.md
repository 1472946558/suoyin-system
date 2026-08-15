# 金匠倌顾客端小程序

顾客预约小程序（V1），与员工收银端（`../miniapp/`）完全独立。

## 项目结构

```
miniapp-customer/
├── app.js / app.json / app.wxss     # 主包入口（4 TabBar + 全局样式）
├── pages/                            # 主包 Tab 页面
│   ├── home/                         # 首页（Tab 1）
│   ├── products/                     # 款式（Tab 2）
│   ├── stores/                       # 门店（Tab 3）
│   └── mine/                         # 我的（Tab 4）
├── pkg-customer/                     # 分包子页面
│   ├── product-detail/               # 款式详情
│   ├── store-detail/                 # 门店详情
│   ├── appointment-create/           # 预约创建
│   ├── appointments/                 # 我的预约列表
│   ├── appointment-detail/           # 预约详情
│   ├── appointment-notes/            # 修改备注
│   └── recycle-info/                 # 回收介绍
├── utils/                            # 工具类
│   ├── request.js                    # 网络请求封装
│   ├── customer-auth.js              # 顾客登录管理
│   ├── customer-date.js              # 日期工具
│   ├── customer-distance.js          # 距离计算
│   └── customer-services.js          # 服务类型/图标映射
└── assets/tabbar/                    # TabBar 图标
```

## 启动

1. 微信开发者工具导入本目录
2. 修改 `project.config.json` 中的 `appid` 为实际的顾客端 AppID
3. 修改 `app.js` 中的 `apiBase` 为后端域名

## 与员工端的关系

- 顾客端和员工端是两个独立小程序
- 同一公众号下关联，互跳通过 `wx.reLaunch`（在「我的」→「员工入口」实现）
- 顾客端独立 AppID（待客户申请）