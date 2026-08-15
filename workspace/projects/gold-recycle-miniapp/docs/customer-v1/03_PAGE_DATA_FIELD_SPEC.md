# 03 - 页面数据字段规格文档

> 本文档列出 8 个顾客端原型页面 + 3 个辅助页面的全部数据字段和接口依赖，作为前后端对接的数据契约。
>
> 所有接口前缀：`/api/v1/customer/`（顾客端）或 `/api/v1/staff/`（员工端）。
>
> 术语约定：管理端入口统一称"店长入口"。

---

## 一、全局数据结构

### 1.1 后端统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "requestId": "xxx"
}
```

- `code === 0` 表示成功，前端 resolve `data`。
- `code !== 0` 表示业务错误，前端 reject `{ code, message, status }`。
- HTTP 状态码非 2xx 时前端 reject `{ status, message }`。

### 1.2 顾客认证数据

| 存储键 | 类型 | 内容 | 用途 |
|--------|------|------|------|
| `customer_token` | string | 顾客 JWT token | 请求头 `Authorization: Bearer {token}` |
| `customer_profile` | object | `{ id, openid, nickname, phone, avatar }` | 本地缓存顾客信息 |

### 1.3 全局状态数据

| 字段 | 类型 | 来源 | 用途 |
|------|------|------|------|
| `app.globalData.apiBase` | string | app.js 配置 | API 基础地址，默认 `https://jinjiangguan.com` |
| `app.globalData.servicePhone` | string | app.js 配置 | 客服电话 |
| `app.globalData.userLocation` | object | `wx.getLocation` | `{ latitude, longitude }`，全局共享 |

### 1.4 服务类型枚举

| code | text | icon |
|------|------|------|
| OLD_FOR_NEW | 旧金换新 | ♻ |
| REPAIR | 黄金维修 | 🔧 |
| CONSULT | 款式工费咨询 | 💬 |
| RECYCLE | 到店回收咨询 | 💎 |

### 1.5 预约状态枚举

| status | text | color |
|--------|------|-------|
| PENDING | 待确认 | #FF9800 |
| CONFIRMED | 已确认 | #4CAF50 |
| ARRIVED | 已到店 | #1976D2 |
| COMPLETED | 已完成 | #8A8F99 |
| CANCELLED | 已取消 | #B0B5BD |
| NO_SHOW | 未到店 | #F44336 |
| TERMINATED | 已终止 | #F44336 |

### 1.6 预约编号格式

- 格式：`YY` + `YYYYMMDD` + 4 位序号
- 示例：`YY26202608150001`

---

## 二、各页面数据字段

### 2.1 首页（pages/home/index）

#### 页面数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| homeData | object | `GET /api/v1/customer/home` | 首页聚合数据 | — |
| homeData.banners | array | 同上 | 轮播图列表 | 每项含 imageUrl/linkType/linkUrl |
| homeData.banners[].imageUrl | string | 同上 | 轮播图图片 | 相对路径，需 absUrl() |
| homeData.banners[].linkType | string | 同上 | 跳转类型 | products/stores/recycle/appointment |
| homeData.banners[].linkUrl | string | 同上 | 自定义跳转 URL | 可为空 |
| homeData.brandName | string | 同上 | 品牌名称 | 降级值 "金匠倌" |
| homeData.brandSlogan1 | string | 同上 | 品牌标语 1 | 降级值 "旧金换打新款" |
| homeData.brandSlogan2 | string | 同上 | 品牌标语 2 | 降级值 "包损耗" |
| nearestStore | object|null | `GET /api/v1/customer/stores` | 附近门店 | — |
| nearestStore.id | string | 同上 | 门店 ID | — |
| nearestStore.name | string | 同上 | 门店名称 | — |
| nearestStore.address | string | 同上 | 门店地址 | — |
| nearestStore.distanceText | string | 前端计算 | 距离文案 | 如 "800m" |
| nearestStore.contactPhone | string | 同上 | 联系电话 | — |
| servicePhone | string | app.globalData | 客服电话 | — |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 获取首页 | GET | `/api/v1/customer/home` | Banner + 品牌文案 | 否 | guest |
| 门店列表 | GET | `/api/v1/customer/stores` | 取附近门店 | 否 | guest |
| 门店详情 | GET | `/api/v1/customer/stores/{id}` | 补经纬度（导航用） | 否 | guest |

---

### 2.2 款式列表页（pages/products/index）

#### 请求参数

| 参数 | 类型 | 必填 | 用途 | 备注 |
|------|------|------|------|------|
| category | string | 否 | 分类筛选 | 空表示全部 |
| keyword | string | 否 | 搜索关键词 | — |
| page | number | 否 | 页码 | 默认 1 |
| pageSize | number | 否 | 每页条数 | 默认 20 |

#### 响应数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| categories | string[] | `GET /api/v1/customer/products` → categories | 分类列表 | 前端拼接"全部" |
| items | array | 同上 → items | 商品列表 | — |
| items[].id | string | 同上 | 商品 ID | 跳转详情 |
| items[].imageUrl | string | 同上 | 商品图片 | 相对路径，需 absUrl() |
| items[].name | string | 同上 | 商品名称 | — |
| items[].category | string | 同上 | 品类 | 如"戒指" |
| items[].purity | string | 同上 | 成色 | 如"足金999" |
| items[].retailPrice | number | 同上 | 零售价 | 0 表示面议 |
| items[].gramWeight | number | 同上 | 克重 | 0 表示无 |
| items[].recommendedScene | string | 同上 | 推荐场景 | 可为空 |
| items[].tags | string[] | 同上 | 标签列表 | 可为空 |
| total | number | 同上 | 总条数 | 判断 hasMore |
| activeCategory | string | 前端状态 | 当前分类 | 默认"全部" |
| keyword | string | 前端状态 | 搜索词 | — |
| page | number | 前端状态 | 当前页码 | — |
| hasMore | boolean | 前端计算 | 是否有更多 | page * 20 < total |
| loading | boolean | 前端状态 | 加载中 | 防重复 |
| error | string|null | 前端状态 | 错误信息 | — |

#### 前端格式化字段

| 字段 | 来源 | 计算方式 | 示例 |
|------|------|----------|------|
| priceText | retailPrice | `>0 ? '¥' + fmtPrice(retailPrice) : '面议'` | "¥1,280" / "面议" |
| gramText | gramWeight | `>0 ? gramWeight + 'g' : ''` | "5.2g" / "" |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 商品列表 | GET | `/api/v1/customer/products` | 分页+分类+搜索 | 否 | guest |

---

### 2.3 款式详情页（pkg-customer/product-detail/index）

#### 请求参数

| 参数 | 类型 | 必填 | 用途 |
|------|------|------|------|
| id | string | 是 | 商品 ID（URL 参数） |

#### 响应数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| product | object | `GET /api/v1/customer/products/{id}` | 商品详情 | — |
| product.id | string | 同上 | 商品 ID | — |
| product.imageUrl | string | 同上 | 商品图片 | 相对路径，需 absUrl() |
| product.name | string | 同上 | 商品名称 | — |
| product.category | string | 同上 | 品类 | — |
| product.purity | string | 同上 | 成色 | — |
| product.retailPrice | number | 同上 | 零售价 | 0 表示面议 |
| product.gramWeight | number | 同上 | 克重 | 0 表示无 |
| product.recommendedScene | string | 同上 | 推荐场景 | — |
| product.tags | string[] | 同上 | 标签列表 | — |
| product.images | array | 前端组装 | 图片列表 | [absUrl(imageUrl)] |
| product.priceMain | string | 前端格式化 | 价格文案 | "¥1,280" / "面议" |
| product.description | string | 前端组装 | 款式说明 | purity + scene + tags |
| product.nearestStore | object|null | `GET /api/v1/customer/stores` | 最近门店 | — |
| product.nearestStore.id | string | 同上 | 门店 ID | — |
| product.nearestStore.name | string | 同上 | 门店名称 | — |
| product.nearestStore.distanceText | string | 前端计算 | 距离文案 | 无定位时"附近" |
| loading | boolean | 前端状态 | 加载中 | — |
| error | string|null | 前端状态 | 错误信息 | — |

#### 前端格式化字段

| 字段 | 来源 | 计算方式 |
|------|------|----------|
| images | imageUrl | `[absUrl(imageUrl)]` |
| priceMain | retailPrice | `>0 ? '¥' + fmtPrice(retailPrice) : '面议'` |
| description | purity + recommendedScene + tags | 排除"新建"/"导入"/"待完善"标签后组装 |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 商品详情 | GET | `/api/v1/customer/products/{id}` | 商品完整信息 | 否 | guest |
| 门店列表 | GET | `/api/v1/customer/stores` | 找最近门店 | 否 | guest |

---

### 2.4 门店列表页（pages/stores/index）

#### 响应数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| stores | array | `GET /api/v1/customer/stores` | 全部门店 | — |
| stores[].id | string | 同上 | 门店 ID | — |
| stores[].name | string | 同上 | 门店名称 | — |
| stores[].city | string | 同上 | 所在城市 | — |
| stores[].address | string | 同上 | 门店地址 | — |
| stores[].contactPhone | string | 同上 | 联系电话 | — |
| stores[].businessHours | string | 同上 | 营业时间 | 如 "09:30-21:30" |
| stores[].longitude | number | `GET /api/v1/customer/stores/{id}` 补齐 | 经度 | 列表接口不返回 |
| stores[].latitude | number | 同上 | 纬度 | 列表接口不返回 |
| stores[].distance | number|null | 前端 Haversine 计算 | 距离 km | 无定位/无坐标为 null |
| stores[].distanceText | string | 前端格式化 | 距离文案 | "800m" / "1.2km" / "" |
| stores[].isOpen | boolean | 前端解析 businessHours | 是否营业中 | — |
| filteredStores | array | 前端过滤 | 当前显示门店 | 按 keyword 过滤 |
| nearestStore | object|null | 前端计算 | 最近门店 | — |
| keyword | string | 用户输入 | 搜索词 | — |
| located | boolean | 前端状态 | 是否已定位 | — |
| locError | string|null | 前端状态 | 定位错误 | — |
| loading | boolean | 前端状态 | 加载中 | — |
| error | string|null | 前端状态 | 接口错误 | — |

#### 距离计算

| 函数 | 输入 | 输出 | 备注 |
|------|------|------|------|
| `distanceKm(lat1, lng1, lat2, lng2)` | 用户定位 + 门店坐标 | 距离 km | Haversine 公式 |
| `fmtDistance(km)` | 距离 km | 距离文案 | <1km → "Xm"；≥1km → "X.Xkm" |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 门店列表 | GET | `/api/v1/customer/stores` | 全部门店 | 否 | guest |
| 门店详情 | GET | `/api/v1/customer/stores/{id}` | 补经纬度 | 否 | guest |

---

### 2.5 门店详情页（pkg-customer/store-detail/index）— 辅助页面

#### 响应数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| store | object | `GET /api/v1/customer/stores/{id}` | 门店详情 | — |
| store.id | string | 同上 | 门店 ID | — |
| store.name | string | 同上 | 门店名称 | — |
| store.city | string | 同上 | 所在城市 | — |
| store.address | string | 同上 | 门店地址 | — |
| store.contactPhone | string | 同上 | 联系电话 | — |
| store.businessHours | string | 同上 | 营业时间 | — |
| store.longitude | number | 同上 | 经度 | 导航用 |
| store.latitude | number | 同上 | 纬度 | 导航用 |
| store.hasCoordinate | boolean | 前端计算 | 是否有坐标 | longitude && latitude |
| store.distanceText | string | 前端计算 | 距离文案 | 有定位+有坐标时 |
| todaySlots | array | `GET /api/v1/customer/stores/{id}/slots?date=today` | 今日时段 | 前 6 个非关闭时段 |
| todaySlots[].time | string | 同上 | 时段开始 | HH:MM |
| todaySlots[].state | string | 同上 | 时段状态 | open/full/closed |
| loading | boolean | 前端状态 | 加载中 | — |
| error | string|null | 前端状态 | 错误信息 | — |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 门店详情 | GET | `/api/v1/customer/stores/{id}` | 门店完整信息 | 否 | guest |
| 时段可用性 | GET | `/api/v1/customer/stores/{id}/slots?date={date}` | 今日时段 | 否 | guest |

---

### 2.6 到店预约页（pkg-customer/appointment-create/index）

#### 请求参数

| 参数 | 类型 | 必填 | 用途 | 备注 |
|------|------|------|------|------|
| storeId | string | 否 | 门店 ID（URL 参数） | 可为空，手动选择 |

#### 页面数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| storeId | string | URL 参数 | 门店 ID | — |
| store | object|null | `GET /api/v1/customer/stores/{id}` | 门店信息 | — |
| store.name | string | 同上 | 门店名称 | — |
| store.address | string | 同上 | 门店地址 | — |
| store.contactPhone | string | 同上 | 联系电话 | — |
| storeLoading | boolean | 前端状态 | 门店加载中 | — |
| serviceTypes | array | 前端常量 SERVICE_TYPES | 服务类型列表 | 4 种 |
| selectedService | string | 用户选择 | 选中的服务类型 code | — |
| days | array | `next7Days()` | 可选日期 | 7 天 |
| days[].date | string | 同上 | 日期 | YYYY-MM-DD |
| days[].label | string | 同上 | 日期标签 | "今天"/"明天"/"周X" |
| days[].shortLabel | string | 同上 | 短日期 | MM/DD |
| selectedDate | string | 用户选择 | 选中日期 | YYYY-MM-DD |
| slots | array | `GET /api/v1/customer/stores/{id}/slots?date=` | 时段列表 | — |
| slots[].startTime | string | 同上 | 开始时间 | HH:MM |
| slots[].endTime | string | 前端计算 | 结束时间 | startTime + 30min |
| slots[].label | string | 前端组装 | 时段文案 | "09:30-10:00" |
| slots[].isPast | boolean | 前端计算 | 是否已过去 | 仅当天 |
| slots[].isFull | boolean | 接口/前端 | 是否约满 | available=false |
| slots[].isClosed | boolean | 接口 | 是否关闭 | state=closed |
| slots[].disabled | boolean | 前端计算 | 是否禁用 | isPast||isFull||isClosed |
| selectedSlot | string | 用户选择 | 选中时段 | HH:MM |
| contactName | string | 用户输入/profile | 联系人姓名 | — |
| contactPhone | string | 手机号授权/profile | 联系人手机号 | 11 位 |
| remark | string | 用户输入 | 备注 | 最多 200 字 |
| isLoggedIn | boolean | 前端状态 | 是否已登录 | — |
| phoneVerified | boolean | 前端状态 | 手机号已授权 | — |
| profile | object | 本地存储 | 顾客信息 | — |
| submitting | boolean | 前端状态 | 提交中 | 防重复 |

#### 提交数据（POST body）

| 字段 | 类型 | 必填 | 用途 | 备注 |
|------|------|------|------|------|
| storeId | string | 是 | 门店 ID | — |
| serviceType | string | 是 | 服务类型 | OLD_FOR_NEW/REPAIR/CONSULT/RECYCLE |
| appointmentDate | string | 是 | 预约日期 | YYYY-MM-DD |
| appointmentTime | string | 是 | 预约时段 | HH:MM（开始时间） |
| contactName | string | 是 | 联系人姓名 | — |
| contactPhone | string | 是 | 联系人手机号 | 11 位 |
| remark | string | 否 | 备注 | 最多 200 字 |

#### 提交响应

| 字段 | 类型 | 用途 | 备注 |
|------|------|------|------|
| id | string | 预约 ID | 跳转详情用 |
| appointmentNo | string | 预约编号 | YY+YYYYMMDD+4位序号 |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 门店详情 | GET | `/api/v1/customer/stores/{id}` | 门店信息 | 否 | guest |
| 时段可用性 | GET | `/api/v1/customer/stores/{id}/slots?date={date}` | 时段列表 | 否 | guest |
| 微信登录 | POST | `/api/v1/customer/auth/wechat-login` | 微信登录 | 否 | guest |
| 手机号授权 | POST | `/api/v1/customer/auth/phone` | 手机号授权 | 是 | customer |
| 创建预约 | POST | `/api/v1/customer/appointments` | 提交预约 | 是 | customer |

#### 时段接口响应

| 字段 | 类型 | 用途 | 备注 |
|------|------|------|------|
| date | string | 日期 | YYYY-MM-DD |
| slots | array | 时段列表 | — |
| slots[].time | string | 开始时间 | HH:MM |
| slots[].available | boolean | 是否可选 | false 表示约满 |
| slots[].state | string | 状态 | open/full/closed |
| slots[].capacity | number | 容量 | 默认 2 |
| slots[].booked | number | 已预约数 | — |

---

### 2.7 我的预约页（pkg-customer/appointments/index）

#### 响应数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| tabs | array | 前端常量 STATUS_TABS | 状态标签 | 5 个 |
| tabs[].key | string | 同上 | 状态 code | 空表示全部 |
| tabs[].label | string | 同上 | 标签文案 | — |
| activeTab | string | 前端状态 | 当前标签 | 默认空 |
| list | array | `GET /api/v1/customer/appointments` | 全部预约 | — |
| list[].id | string | 同上 | 预约 ID | — |
| list[].appointmentNo | string | 同上 | 预约编号 | — |
| list[].storeName | string | 同上 | 门店名称 | — |
| list[].serviceType | string | 同上 | 服务类型 code | — |
| list[].appointmentDate | string | 同上 | 预约日期 | YYYY-MM-DD |
| list[].startTime | string | 同上 | 开始时间 | HH:MM |
| list[].endTime | string | 同上 | 结束时间 | HH:MM |
| list[].status | string | 同上 | 预约状态 | PENDING/CONFIRMED 等 |
| list[].statusText | string | 前端映射 | 状态文案 | — |
| list[].statusColor | string | 前端映射 | 状态颜色 | — |
| list[].serviceText | string | 前端映射 | 服务类型文案 | — |
| filteredList | array | 前端过滤 | 当前显示 | 按 activeTab 过滤 |
| loading | boolean | 前端状态 | 加载中 | — |
| isLoggedIn | boolean | 前端状态 | 是否已登录 | — |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 预约列表 | GET | `/api/v1/customer/appointments` | 我的预约 | 是 | customer |

---

### 2.8 预约详情页（pkg-customer/appointment-detail/index）

#### 请求参数

| 参数 | 类型 | 必填 | 用途 |
|------|------|------|------|
| id | string | 是 | 预约 ID（URL 参数） |

#### 响应数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| detail | object | `GET /api/v1/customer/appointments/{id}` | 预约详情 | — |
| detail.id | string | 同上 | 预约 ID | — |
| detail.appointmentNo | string | 同上 | 预约编号 | YY+YYYYMMDD+4位序号 |
| detail.status | string | 同上 | 预约状态 | 7 态枚举 |
| detail.storeName | string | 同上 | 门店名称 | — |
| detail.storeAddress | string | 同上 | 门店地址 | — |
| detail.storePhone | string | 同上 | 门店电话 | — |
| detail.storeLongitude | number | 同上 | 门店经度 | 导航用 |
| detail.storeLatitude | number | 同上 | 门店纬度 | 导航用 |
| detail.serviceType | string | 同上 | 服务类型 code | — |
| detail.appointmentDate | string | 同上 | 预约日期 | YYYY-MM-DD |
| detail.startTime | string | 同上 | 开始时间 | HH:MM |
| detail.endTime | string | 同上 | 结束时间 | HH:MM |
| detail.contactName | string | 同上 | 联系人姓名 | — |
| detail.contactPhone | string | 同上 | 联系人手机号 | — |
| detail.remark | string | 同上 | 备注 | 可为空 |
| detail.createdAt | string | 同上 | 创建时间 | ISO 8601 |
| detail.confirmedAt | string | 同上 | 确认时间 | 可为空 |
| detail.cancelledAt | string | 同上 | 取消时间 | 可为空 |
| statusText | string | 前端映射 | 状态文案 | — |
| statusColor | string | 前端映射 | 状态颜色 | — |
| serviceText | string | 前端映射 | 服务类型文案 | — |
| canCancel | boolean | 前端计算 | 是否可取消 | 状态+时间判定 |
| canEditNotes | boolean | 前端计算 | 是否可改备注 | 状态判定 |
| loading | boolean | 前端状态 | 加载中 | — |
| error | string|null | 前端状态 | 错误信息 | — |

#### canCancel 计算逻辑

```
canCancel = CANCELLABLE.includes(detail.status) && 
            (appointmentTime - now > 2 * 60 * 60 * 1000)

CANCELLABLE = ['PENDING', 'CONFIRMED']
appointmentTime = new Date(detail.appointmentDate + 'T' + detail.startTime + ':00')
```

#### canEditNotes 计算逻辑

```
canEditNotes = NOTES_EDITABLE.includes(detail.status)

NOTES_EDITABLE = ['PENDING', 'CONFIRMED']
```

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 预约详情 | GET | `/api/v1/customer/appointments/{id}` | 预约完整信息 | 是 | customer |
| 取消预约 | POST | `/api/v1/customer/appointments/{id}/cancel` | 取消预约 | 是 | customer |
| 修改备注 | PUT | `/api/v1/customer/appointments/{id}/notes` | 修改备注（跳转备注页执行） | 是 | customer |

#### 取消预约请求

| 字段 | 类型 | 必填 | 用途 |
|------|------|------|------|
| reason | string | 是 | 取消原因 | 默认 "用户取消" |

#### 修改备注请求

| 字段 | 类型 | 必填 | 用途 |
|------|------|------|------|
| remark | string | 否 | 备注 | 最多 200 字 |

---

### 2.9 预约备注页（pkg-customer/appointment-notes/index）— 辅助页面

#### 请求参数

| 参数 | 类型 | 必填 | 用途 |
|------|------|------|------|
| id | string | 是 | 预约 ID（URL 参数） |
| remark | string | 否 | 当前备注（URL 参数，encodeURIComponent） |

#### 页面数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| id | string | URL 参数 | 预约 ID | — |
| remark | string | URL 参数 / 用户输入 | 备注 | 最多 200 字 |
| saving | boolean | 前端状态 | 保存中 | 防重复 |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 修改备注 | PUT | `/api/v1/customer/appointments/{id}/notes` | 保存备注 | 是 | customer |

---

### 2.10 回收介绍页（pkg-customer/recycle-info/index）— 辅助页面

#### 响应数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| recycleInfo | object | `GET /api/v1/customer/recycle-info` | 回收介绍 | 降级为本地默认 |
| recycleInfo.title | string | 同上 | 标题 | 降级 "黄金回收服务" |
| recycleInfo.intro | string | 同上 | 简介 | 降级文案 |
| recycleInfo.process | array | 同上 | 服务流程 | 降级 4 步 |
| recycleInfo.process[].step | number | 同上 | 步骤序号 | 1-4 |
| recycleInfo.process[].title | string | 同上 | 步骤标题 | — |
| recycleInfo.process[].desc | string | 同上 | 步骤描述 | — |
| recycleInfo.services | array | 同上 | 服务类型 | 降级 4 类 |
| recycleInfo.services[].icon | string | 同上 | 图标 | — |
| recycleInfo.services[].title | string | 同上 | 标题 | — |
| recycleInfo.services[].desc | string | 同上 | 描述 | — |
| recycleInfo.notices | array | 同上 | 注意事项 | 降级 4 条 |
| loading | boolean | 前端状态 | 加载中 | — |
| error | string|null | 前端状态 | 错误信息 | 失败时降级，不报错 |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 回收介绍 | GET | `/api/v1/customer/recycle-info` | 回收服务介绍 | 否 | guest |

---

### 2.11 我的页（pages/mine/index）

#### 页面数据

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| isLoggedIn | boolean | 前端状态 | 是否已登录 | — |
| profile | object|null | 本地存储 / `GET /api/v1/customer/me` | 顾客信息 | — |
| profile.nickname | string | 同上 | 昵称 | 无则"顾客" |
| profile.phone | string | 同上 | 手机号 | 脱敏显示 |
| profile.avatar | string | 同上 | 头像 | — |
| servicePhone | string | app.globalData | 客服电话 | — |

#### 接口依赖

| 接口 | 方法 | 路径 | 用途 | 需登录 | 权限 |
|------|------|------|------|--------|------|
| 顾客信息 | GET | `/api/v1/customer/me` | 刷新 profile | 是 | customer |
| 微信登录 | POST | `/api/v1/customer/auth/wechat-login` | 微信登录 | 否 | guest |
| 手机号授权 | POST | `/api/v1/customer/auth/phone` | 手机号授权 | 是 | customer |
| 退出登录 | POST | `/api/v1/customer/auth/logout` | 退出登录 | 是 | customer |

#### 微信登录请求

| 字段 | 类型 | 必填 | 用途 |
|------|------|------|------|
| code | string | 是 | wx.login 拿到的 code | — |

#### 微信登录响应

| 字段 | 类型 | 用途 | 备注 |
|------|------|------|------|
| token | string | 顾客 JWT token | 存入 customer_token |
| profile | object | 顾客信息 | 存入 customer_profile |

#### 手机号授权请求

| 字段 | 类型 | 必填 | 用途 |
|------|------|------|------|
| phoneCode | string | 是 | getPhoneNumber 拿到的 code | — |

#### 手机号授权响应

| 字段 | 类型 | 用途 | 备注 |
|------|------|------|------|
| id | string | 顾客 ID | — |
| openid | string | 微信 openid | — |
| nickname | string | 昵称 | — |
| phone | string | 手机号 | — |
| avatar | string | 头像 | — |

---

## 三、接口总览

### 3.1 顾客端公开接口（无需登录）

| # | 方法 | 路径 | 用途 |
|---|------|------|------|
| 1 | GET | `/api/v1/customer/home` | 首页 Banner + 品牌文案 |
| 2 | GET | `/api/v1/customer/products` | 商品列表（分页+分类+搜索） |
| 3 | GET | `/api/v1/customer/products/{id}` | 商品详情 |
| 4 | GET | `/api/v1/customer/stores` | 门店列表 |
| 5 | GET | `/api/v1/customer/stores/{id}` | 门店详情 |
| 6 | GET | `/api/v1/customer/stores/{id}/slots?date={date}` | 时段可用性 |
| 7 | GET | `/api/v1/customer/recycle-info` | 回收服务介绍 |
| 8 | POST | `/api/v1/customer/auth/wechat-login` | 微信登录 |

### 3.2 顾客端认证接口（需顾客 token）

| # | 方法 | 路径 | 用途 |
|---|------|------|------|
| 9 | POST | `/api/v1/customer/auth/phone` | 手机号授权 |
| 10 | POST | `/api/v1/customer/auth/logout` | 退出登录 |
| 11 | GET | `/api/v1/customer/me` | 顾客个人信息 |
| 12 | GET | `/api/v1/customer/appointments` | 预约列表 |
| 13 | POST | `/api/v1/customer/appointments` | 创建预约 |
| 14 | GET | `/api/v1/customer/appointments/{id}` | 预约详情 |
| 15 | POST | `/api/v1/customer/appointments/{id}/cancel` | 取消预约 |
| 16 | PUT | `/api/v1/customer/appointments/{id}/notes` | 修改备注 |

### 3.3 员工端接口（需员工 token，N08 对接）

| # | 方法 | 路径 | 用途 |
|---|------|------|------|
| 17 | GET | `/api/v1/staff/appointments` | 员工预约列表 |
| 18 | GET | `/api/v1/staff/appointments/{id}` | 员工预约详情 |
| 19 | PUT | `/api/v1/staff/appointments/{id}/status` | 员工更新预约状态 |

---

## 四、前端工具函数

### 4.1 格式化函数

| 函数 | 文件 | 输入 | 输出 | 用途 |
|------|------|------|------|------|
| `fmtPrice(n)` | customer-services.js | 数字 | 千分位字符串 | "1,280.00" |
| `absUrl(url)` | customer-services.js | 相对路径 | 绝对 URL | 补全域名 |
| `fmtDistance(km)` | customer-distance.js | 距离 km | 距离文案 | "800m" / "1.2km" |
| `distanceKm(lat1,lng1,lat2,lng2)` | customer-distance.js | 4 个坐标 | 距离 km | Haversine |
| `fmtDate(d)` | customer-date.js | Date 对象 | YYYY-MM-DD | 日期格式化 |
| `next7Days()` | customer-date.js | 无 | 7 天数组 | 预约日期选择 |
| `timePlusMin(time, min)` | customer-date.js | HH:MM + 分钟 | HH:MM | 计算结束时间 |
| `isPastTime(date, time)` | customer-date.js | 日期+时间 | boolean | 是否已过去 |

### 4.2 映射函数

| 函数 | 文件 | 输入 | 输出 | 用途 |
|------|------|------|------|------|
| `serviceTypeText(code)` | customer-services.js | 服务类型 code | 中文文案 | "旧金换新" |
| `appointmentStatusText(status)` | customer-services.js | 状态 code | 中文文案 | "待确认" |
| `appointmentStatusColor(status)` | customer-services.js | 状态 code | 颜色值 | "#FF9800" |

### 4.3 认证函数

| 函数 | 文件 | 用途 | 返回 |
|------|------|------|------|
| `getCustomerToken()` | customer-auth.js | 读取本地 token | string |
| `setCustomerToken(token)` | customer-auth.js | 保存 token | void |
| `clearCustomerSession()` | customer-auth.js | 清除 token + profile | void |
| `getStoredProfile()` | customer-auth.js | 读取本地 profile | object|null |
| `ensureCustomerSession()` | customer-auth.js | 静默登录 | Promise<{token, profile}> |
| `customerLogin()` | customer-auth.js | 主动登录 | Promise<{token, profile}> |
| `customerPhoneAuth(phoneCode)` | customer-auth.js | 手机号授权 | Promise<profile> |
| `customerLogout()` | customer-auth.js | 退出登录 | Promise<void> |

---

## 五、数据字段排除清单

### 5.1 商品 DTO 顾客端不可见字段

以下字段在后端 CatalogProduct 中存在，但顾客端 DTO 必须排除：

| 排除字段 | 原因 |
|----------|------|
| benchPrice | 柜台价，仅内部使用 |
| inventory | 库存数，顾客不可见 |
| stockStatus | 库存状态，顾客不可见 |
| storeIDs | 关联门店，顾客不可见 |
| sku | SKU 编号，顾客不可见 |

### 5.2 门店 DTO 字段可见性

| 字段 | 列表接口 | 详情接口 | 备注 |
|------|----------|----------|------|
| id | ✅ | ✅ | — |
| name | ✅ | ✅ | — |
| city | ✅ | ✅ | — |
| address | ✅ | ✅ | — |
| contactPhone | ✅ | ✅ | — |
| businessHours | ✅ | ✅ | — |
| longitude | ❌ | ✅ | 仅详情返回 |
| latitude | ❌ | ✅ | 仅详情返回 |

### 5.3 顾客 token 与员工 token 隔离

| 维度 | 顾客 token | 员工 token |
|------|-----------|-----------|
| Redis key 前缀 | `customer_session:` | `session:` |
| 接口前缀 | `/api/v1/customer/` | `/api/v1/` + `/api/admin/` |
| 中间件 | `withCustomerAuth` | `withAuth` |
| 不可互用 | 顾客 token 不能调员工接口 | 员工 token 不能调顾客接口 |
