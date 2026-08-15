# 01 - 产品需求文档（PRD）

## 1. 项目目标

在现有金匠倌员工/门店管理小程序中，新增面向普通顾客的功能模块。同一个小程序、同一 AppID，默认进入顾客端，员工从"我的→员工入口"登录。

## 2. 用户角色

| 角色 | 身份标识 | 登录方式 |
|------|----------|----------|
| guest | 无 | 无需登录 |
| customer | 微信 openid + 手机号 | 微信授权 |
| store_staff | 员工账号 + 密码 | 账号密码 |
| store_manager | 店长账号 + 密码 | 账号密码 |
| super_admin | 管理员账号 + 密码 | 账号密码 |

## 3. 业务范围（V1 做）

1. 顾客首页（轮播图 + 品牌文案 + 核心入口 + 底部门店预约）
2. 商品展示（分类、列表、详情）— 只展示，不交易
3. 黄金回收服务介绍 — 只介绍，不在线估价/结算
4. 门店列表 + 最近门店推荐
5. 到店预约（创建、查看、取消）
6. 我的预约（列表、详情、取消）
7. 我的页面（顾客信息 + 员工入口）
8. 员工入口（账号密码登录 → 按角色进入工作台）

## 4. 不做内容（V1 排除）

- 购物车、在线下单、微信支付、物流发货、退款售后
- 优惠券、会员积分
- 在线估价、在线回收结算、上门回收、邮寄回收
- 参考金价展示（顾客端不展示）

## 5. 页面清单

### 顾客端页面（独立分包 pkg-customer/）

| 页面 | 路径 | 权限 |
|------|------|------|
| 首页 | pkg-customer/pages/home/index | guest |
| 商品列表 | pkg-customer/pages/products/list | guest |
| 商品详情 | pkg-customer/pages/products/detail | guest |
| 黄金回收介绍 | pkg-customer/pages/recycle/info | guest |
| 门店列表 | pkg-customer/pages/stores/list | guest |
| 门店详情 | pkg-customer/pages/stores/detail | guest |
| 到店预约 | pkg-customer/pages/appointment/create | customer |
| 我的预约 | pkg-customer/pages/appointment/list | customer |
| 预约详情 | pkg-customer/pages/appointment/detail | customer |
| 我的 | pkg-customer/pages/mine/index | guest |

### 员工入口

| 页面 | 路径 | 权限 |
|------|------|------|
| 员工登录 | pages/account/index | guest |
| 员工工作台 | pages/home/index | store_staff+ |
| 预约管理 | pages/appointments/index | store_staff+ |

## 6. 关键业务流程

### 6.1 顾客浏览流程
```
打开小程序 → 顾客首页 → 浏览商品/门店/回收介绍 → 选择门店 → 到店预约
```

### 6.2 顾客预约流程
```
选择门店 → 选择日期时间 → 填写联系人信息 → 微信登录 + 手机号授权 → 提交预约
→ 状态 PENDING → 门店确认 → CONFIRMED → 到店 → ARRIVED → COMPLETED
```

### 6.3 员工登录流程
```
我的 → 员工入口 → 账号密码登录 → 后端校验角色权限 → 进入工作台/管理端
```

## 7. 首页布局

```
┌─────────────────────────┐
│    轮播图 Banner         │  ← 多张图片，可滑动，可后台维护
├─────────────────────────┤
│    金匠馆                │  ← 品牌文案区
│    旧金换打新款           │
│    包损耗                │
├─────────────────────────┤
│  [款式图]  [工费]  [点击看] │  ← 核心功能入口
├─────────────────────────┤
│  [地图门店] [附近门店]     │  ← 底部门店入口
│  [到店预约]               │
└─────────────────────────┘
```

## 8. 预约规则（已确认）

- 时间粒度：30 分钟
- 可预约范围：未来 7 天
- 当天预约至少提前 1 小时
- 取消预约至少提前 2 小时
- 同一手机号 + 同一门店 + 同一时间段不能重复提交有效预约
- 预约需要门店确认（PENDING → CONFIRMED）

## 9. 预约状态机（7 态）

```
PENDING ──门店确认──→ CONFIRMED ──到店──→ ARRIVED ──完成──→ COMPLETED
   │                     │
   │                     ├──未到店──→ NO_SHOW
   │                     │
   └──取消──→ CANCELLED  ├──取消──→ CANCELLED
                         │
                         └──终止──→ TERMINATED

PENDING 也可直接 → ARRIVED（未确认直接到店）
```

## 10. 门店经纬度规则

- 门店地址、经度、纬度由客户/管理员在后台手动填写
- 小程序获取顾客当前位置后，根据门店经纬度计算距离
- 系统推荐最近的可预约门店
- 如果用户拒绝位置授权，允许手动选择门店
- 顾客位置仅内存计算距离，不上传不落库

## 11. 权限隔离原则

- 顾客 token（customer 命名空间）→ `/api/v1/customer/` 接口
- 员工 token（operator 命名空间）→ `/api/v1/` + `/api/admin/` 接口
- 两套 token 不可互用
- 后端必须校验身份和权限，不依赖前端隐藏页面
- 顾客不能调用员工接口，普通员工不能调用管理员接口
