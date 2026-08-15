# 03 - 页面开发规格文档

> 本文档覆盖顾客端 8 个原型页面，每页含 10 节：页面目标、页面入口、信息结构、操作说明、交互说明、页面状态、数据字段、接口依赖、跳转关系、验收标准。
>
> 术语约定：管理端入口统一称"店长入口"，不写"员工入口"。

---

## 页面 1：首页

### 1.1 页面目标

让顾客一打开小程序就看到品牌形象、核心服务入口和最近门店，引导顾客浏览款式或发起预约。

### 1.2 页面入口

- App 启动默认 Tab（TabBar 第 1 项）
- 其他 Tab 切回首页
- 分包页面 `wx.navigateBack` 回到首页

### 1.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | Banner 轮播图 | imageUrl, linkType, linkUrl | `GET /api/v1/customer/home` → banners | 否 | 是 | 显示默认品牌占位图 |
| M2 | 品牌文案区 | brandName, brandSlogan1, brandSlogan2 | 同上 | 是 | 否 | 降级文案 "金匠倌 / 旧金换打新款 / 包损耗" |
| M3 | 核心入口 | 款式图 / 工费 / 点击看 | 固定文案 | 是 | 是 | 不适用（固定入口） |
| M4 | 附近门店卡片 | name, address, distanceText, contactPhone | `GET /api/v1/customer/stores` + 定位计算 | 否 | 是 | 不显示卡片 |
| M5 | 底部操作区 | 到店预约按钮、地图导航按钮 | 固定入口 | 是 | 是 | 不适用 |
| M6 | TabBar | 首页/款式/门店/我的 | app.json 配置 | 是 | 是 | 不适用 |

### 1.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击 Banner | M1 轮播图 | 无 | 按 linkType 跳转：products→款式 Tab，stores→门店 Tab，recycle→回收介绍，appointment→预约创建 | 无效链接：不跳转 |
| 点击"款式图" | M3 核心入口 | 无 | `wx.switchTab` 到款式列表页 | 不适用 |
| 点击"工费" | M3 核心入口 | 无 | `wx.switchTab` 到款式列表页 | 不适用 |
| 点击"点击看" | M3 核心入口 | 无 | `wx.switchTab` 到款式列表页 | 不适用 |
| 点击附近门店卡片 | M4 | 有门店数据 | `wx.navigateTo` 门店详情页 | 无门店：不显示卡片 |
| 点击"到店预约" | M5 | 无 | `wx.navigateTo` 预约创建页（带 storeId） | 无 |
| 点击"地图导航" | M5 | 门店有经纬度 | `wx.openLocation` 打开微信地图 | 无经纬度：toast "该门店暂未配置位置" |
| 切换 Tab | M6 | 无 | 切换到对应 Tab 页 | 不适用 |

### 1.5 交互说明

- Banner 自动轮播：间隔 4 秒，可手动滑动，点击跳转。
- 品牌文案区为静态展示，不响应点击。
- 附近门店卡片在 `onShow` 时重新拉取门店列表并计算距离。
- 首次进入时调用 `ensureCustomerSession()` 静默登录（不弹窗），失败不阻塞。
- "旧金回收"仅作为品牌文案出现，不做独立按钮。

### 1.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态 | 展示 Banner + 品牌文案 + 核心入口 + 附近门店（如有） |
| 加载中 | Banner 区域空白，品牌文案区显示降级文案 |
| 空数据 | 无 Banner 时显示默认占位图；无门店时不显示门店卡片 |
| 接口失败 | 品牌文案降级为 "金匠倌 / 旧金换打新款 / 包损耗"；门店卡片不显示 |
| 未登录 | 不阻塞浏览，静默登录在后台进行 |
| 未授权手机号 | 不阻塞浏览 |
| 定位失败 | 门店卡片不显示距离，仍可点击进入门店列表手动选择 |
| 无权限 | 不适用（guest 可访问） |

### 1.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| banners | array | home 接口 | 轮播图列表 | 每项含 imageUrl/linkType/linkUrl |
| brandName | string | home 接口 | 品牌名称 | 降级值 "金匠倌" |
| brandSlogan1 | string | home 接口 | 品牌标语 1 | 降级值 "旧金换打新款" |
| brandSlogan2 | string | home 接口 | 品牌标语 2 | 降级值 "包损耗" |
| servicePhone | string | app.globalData | 客服电话 | 全局配置 |
| nearestStore | object | stores 接口 | 附近门店卡片 | 含 name/address/distanceText |
| nearestStore.id | string | stores 接口 | 跳转门店详情 | — |
| nearestStore.name | string | stores 接口 | 门店名称 | — |
| nearestStore.address | string | stores 接口 | 门店地址 | — |
| nearestStore.distanceText | string | 定位计算 | 距离文案 | 如 "800m" / "1.2km" |
| nearestStore.contactPhone | string | stores 接口 | 联系电话 | — |

### 1.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/home` | GET | 首页 Banner + 品牌文案 | 否 | guest |
| `/api/v1/customer/stores` | GET | 门店列表（取第一家或最近） | 否 | guest |
| `/api/v1/customer/stores/{id}` | GET | 门店详情（补经纬度用于导航） | 否 | guest |

### 1.9 跳转关系

```
首页 → 款式列表页（switchTab）
首页 → 门店列表页（switchTab）
首页 → 门店详情页（navigateTo，带 id）
首页 → 回收介绍页（navigateTo，Banner 跳转）
首页 → 预约创建页（navigateTo，带 storeId）
```

### 1.10 验收标准

1. 用户打开小程序后默认进入首页，看到 Banner 轮播图、品牌文案、核心入口。
2. Banner 可自动轮播和手动滑动，点击后跳转到对应页面。
3. 首页不出现"旧金回收"独立按钮，"旧金回收"仅作为品牌文案展示。
4. 用户点击"款式图 / 工费 / 点击看"后，进入款式列表页。
5. 附近门店卡片显示门店名称、地址和距离，点击进入门店详情。
6. 定位失败时门店卡片不显示距离但不报错，用户可点击进入门店列表手动选择。
7. 接口失败时品牌文案降级显示"金匠倌 / 旧金换打新款 / 包损耗"。

---

## 页面 2：款式图 / 工费列表页

### 2.1 页面目标

让顾客浏览门店在售款式，按分类筛选和搜索，找到感兴趣的款式后进入详情。

### 2.2 页面入口

- TabBar 第 2 项"款式"
- 首页核心入口"款式图 / 工费 / 点击看" `switchTab`
- 其他页面 `switchTab` 到款式 Tab

### 2.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 搜索栏 | keyword | 用户输入 | 否 | 是 | placeholder "搜索款式" |
| M2 | 分类标签栏 | categories | `GET /api/v1/customer/products` → categories | 是 | 是 | 仅显示"全部" |
| M3 | 商品列表（双列） | imageUrl, name, category, purity, retailPrice, gramWeight | 同上 → items | 是 | 是 | "暂无商品" |
| M4 | 加载更多 | — | 触底触发 | 否 | 否 | 无更多数据时不显示 |
| M5 | TabBar | 同首页 | app.json | 是 | 是 | 不适用 |

### 2.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 输入搜索词 | M1 搜索栏 | 无 | 实时更新 keyword | 不适用 |
| 确认搜索 | M1 键盘确认 | 有 keyword | 重新加载列表（page=1） | 无结果："暂无商品" |
| 清除搜索 | M1 × 按钮 | 有 keyword | 清空 keyword 并重新加载 | 不适用 |
| 切换分类 | M2 标签 | 无 | 重新加载列表（page=1，按分类筛选） | 无结果："暂无商品" |
| 点击商品卡片 | M3 | 有商品 | `wx.navigateTo` 款式详情页（带 id） | 不适用 |
| 触底加载更多 | M3 列表底部 | hasMore=true 且 loading=false | 追加下一页数据 | 无更多：不触发 |
| 下拉刷新 | 页面顶部 | 无 | 重新加载列表（page=1） | 不适用 |

### 2.5 交互说明

- 分类标签栏横向滚动，选中项高亮（金色 #866E23）。
- 切换分类时立即重新请求，不等动画完成。
- 搜索为输入框 + 确认按钮，不实时搜索（防抖由 `onSearchConfirm` 触发）。
- 双列瀑布流卡片，图片宽高比 1:1，下方显示名称/品类/价格/克重。
- 价格格式：有 retailPrice 显示 "¥" + 千分位金额，无则显示"面议"。
- 克重格式：有 gramWeight 显示 "Xg"，无则不显示。
- 商品 imageUrl 为相对路径，前端用 `absUrl()` 补全域名。

### 2.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态 | 展示分类标签 + 商品双列列表 |
| 加载中 | 首次加载显示 loading 提示；加载更多显示底部 loading |
| 空数据 | 列表区域显示"暂无商品" |
| 接口失败 | 列表区域显示错误文案 + "重试"按钮 |
| 未登录 | 不阻塞（guest 可浏览） |
| 未授权手机号 | 不阻塞 |
| 定位失败 | 不适用 |
| 无权限 | 不适用 |

### 2.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| categories | string[] | products 接口 | 分类标签 | 首次加载取一次，前端拼接"全部" |
| activeCategory | string | 前端状态 | 当前选中分类 | 默认"全部" |
| keyword | string | 用户输入 | 搜索关键词 | — |
| products | array | products 接口 → items | 商品列表 | — |
| products[].id | string | 同上 | 跳转详情 | — |
| products[].imageUrl | string | 同上 | 商品图片 | 相对路径，需 absUrl() |
| products[].name | string | 同上 | 商品名称 | — |
| products[].category | string | 同上 | 品类 | 如"戒指" |
| products[].purity | string | 同上 | 成色 | 如"足金999" |
| products[].retailPrice | number | 同上 | 零售价 | 0 表示面议 |
| products[].gramWeight | number | 同上 | 克重 | 0 表示无 |
| products[].priceText | string | 前端格式化 | 价格文案 | "¥1,280" 或 "面议" |
| products[].gramText | string | 前端格式化 | 克重文案 | "5.2g" 或空 |
| page | number | 前端状态 | 当前页码 | 默认 1 |
| total | number | products 接口 | 总条数 | 用于判断 hasMore |
| hasMore | boolean | 前端计算 | 是否还有更多 | page * 20 < total |
| loading | boolean | 前端状态 | 加载中标记 | 防重复请求 |
| error | string|null | 前端状态 | 错误信息 | 有值时显示重试 |

### 2.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/products` | GET | 商品列表（分页+分类+搜索） | 否 | guest |

请求参数：`?category={分类}&keyword={搜索词}&page={页码}&pageSize=20`

### 2.9 跳转关系

```
款式列表页 → 款式详情页（navigateTo，带 id）
款式列表页 → 首页（switchTab）
款式列表页 → 门店列表页（switchTab）
款式列表页 → 我的页（switchTab）
```

### 2.10 验收标准

1. 用户进入款式 Tab 后，看到分类标签栏和商品双列卡片列表。
2. 切换分类后列表内容跟随刷新，分类标签高亮正确。
3. 搜索框输入关键词并确认后，列表按关键词筛选。
4. 列表触底时自动加载下一页，无更多数据时不再触发。
5. 商品图片为相对路径时正确补全为绝对 URL 并显示。
6. 无商品时显示"暂无商品"，接口失败时显示重试按钮。
7. 点击商品卡片进入款式详情页，URL 携带商品 id。
8. 下拉刷新后列表从第一页重新加载。

---

## 页面 3：款式详情页

### 3.1 页面目标

展示单个款式的完整信息（图片、名称、品类、成色、价格、克重、标签、推荐场景），引导顾客到店咨询或预约。

### 3.2 页面入口

- 款式列表页点击商品卡片 `navigateTo`（带 id）
- 首页 Banner 跳转（如 linkType 配置）

### 3.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 商品图片 | imageUrl | `GET /api/v1/customer/products/{id}` | 是 | 否 | 占位图 |
| M2 | 商品信息 | name, category, purity, retailPrice, gramWeight | 同上 | 是 | 否 | 不显示对应行 |
| M3 | 款式说明 | description（前端组装） | purity + recommendedScene + tags | 否 | 否 | 默认文案 |
| M4 | 标签 | tags | 同上 | 否 | 否 | 不显示 |
| M5 | 推荐场景 | recommendedScene | 同上 | 否 | 否 | 不显示 |
| M6 | 最近门店 | name, distanceText, address | stores 接口 + 定位计算 | 否 | 是 | 不显示 |
| M7 | 底部操作栏 | 到店预约 / 咨询门店 / 客服电话 | 固定按钮 | 是 | 是 | 不适用 |

### 3.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击最近门店卡片 | M6 | 有门店数据 | `wx.navigateTo` 门店详情页 | 无门店：不显示 |
| 点击"到店预约" | M7 | 无 | `wx.navigateTo` 预约创建页（带 storeId） | 无 |
| 点击"咨询门店" | M7 | 无 | `wx.switchTab` 门店列表页 | 不适用 |
| 点击"客服电话" | M7 | 有 servicePhone | `wx.makePhoneCall` | 无电话：toast "暂无联系电话" |
| 返回 | 导航栏返回 | 无 | `wx.navigateBack` | 不适用 |

### 3.5 交互说明

- 商品图片为正方形展示，宽度 100%。
- description 由前端组装：purity + recommendedScene + tags（排除"新建"/"导入"/"待完善"标签），以逗号连接。
- 最近门店计算逻辑：拉取门店列表，若有定位则按距离排序取最近；无定位则取第一个有联系方式的门店。
- 底部操作栏固定在页面底部，不随滚动消失。

### 3.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态 | 展示商品图片 + 信息 + 说明 + 标签 + 最近门店 + 底部操作栏 |
| 加载中 | 全屏 loading 提示 |
| 空数据 | 不适用（有 id 必有数据） |
| 接口失败 | 显示错误文案 + "重试"按钮 |
| 未登录 | 不阻塞（guest 可浏览） |
| 未授权手机号 | 不阻塞（到店预约时才要求授权） |
| 定位失败 | 最近门店不显示距离，仍展示门店名称 |
| 无权限 | 不适用 |

### 3.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| product.id | string | URL 参数 | 请求商品详情 | — |
| product.imageUrl | string | products/{id} 接口 | 商品图片 | 相对路径，需 absUrl() |
| product.name | string | 同上 | 商品名称 | — |
| product.category | string | 同上 | 品类 | — |
| product.purity | string | 同上 | 成色 | 如"足金999" |
| product.retailPrice | number | 同上 | 零售价 | 0 表示面议 |
| product.gramWeight | number | 同上 | 克重 | 0 表示无 |
| product.recommendedScene | string | 同上 | 推荐场景 | — |
| product.tags | string[] | 同上 | 标签列表 | — |
| product.priceMain | string | 前端格式化 | 价格文案 | "¥1,280" 或 "面议" |
| product.description | string | 前端组装 | 款式说明 | — |
| product.nearestStore | object | stores 接口 | 最近门店 | — |
| product.nearestStore.id | string | 同上 | 跳转门店详情 | — |
| product.nearestStore.name | string | 同上 | 门店名称 | — |
| product.nearestStore.distanceText | string | 定位计算 | 距离文案 | 无定位时为"附近" |

### 3.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/products/{id}` | GET | 商品详情 | 否 | guest |
| `/api/v1/customer/stores` | GET | 门店列表（找最近门店） | 否 | guest |

### 3.9 跳转关系

```
款式详情页 → 门店详情页（navigateTo，带 id）
款式详情页 → 门店列表页（switchTab）
款式详情页 → 预约创建页（navigateTo，带 storeId）
款式详情页 → 返回款式列表页（navigateBack）
```

### 3.10 验收标准

1. 用户从款式列表点击商品后，进入详情页看到完整商品信息。
2. 商品图片正确显示（相对路径补全为绝对 URL）。
3. 有零售价时显示 "¥" + 千分位金额，无则显示"面议"。
4. 款式说明由 purity + recommendedScene + tags 组装，排除系统标签。
5. 最近门店卡片显示门店名称和距离，点击进入门店详情。
6. 底部操作栏固定显示"到店预约 / 咨询门店 / 客服电话"三个按钮。
7. 点击"到店预约"跳转预约创建页，URL 携带 storeId。
8. 接口失败时显示错误文案和重试按钮，重试后重新加载。

---

## 页面 4：附近门店 / 门店列表页

### 4.1 页面目标

展示所有门店列表，按距离排序，让顾客找到最近的门店并查看详情、拨打电话、导航或发起预约。

### 4.2 页面入口

- TabBar 第 3 项"门店"
- 首页"地图门店" `switchTab`
- 款式详情页"咨询门店" `switchTab`

### 4.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 定位状态栏 | locError, located | wx.getLocation 结果 | 否 | 是 | 未定位：显示"定位服务未开启" |
| M2 | 搜索栏 | keyword | 用户输入 | 否 | 是 | placeholder "搜索门店" |
| M3 | 门店卡片列表 | name, city, address, contactPhone, businessHours, distanceText, isOpen | `GET /api/v1/customer/stores` + 详情补经纬度 + 定位计算 | 是 | 是 | "暂无门店" |
| M4 | 门店卡片操作 | 拨打电话 / 查看详情 / 预约到店 / 导航 | 固定按钮 | 是 | 是 | 不适用 |
| M5 | TabBar | 同首页 | app.json | 是 | 是 | 不适用 |

### 4.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击"重新定位" | M1 定位状态栏 | locError 有值 | `wx.getLocation` 重新获取定位 | 定位失败：toast "定位失败" |
| 输入搜索词 | M2 搜索栏 | 无 | 实时过滤门店列表 | 无匹配：列表为空 |
| 清除搜索 | M2 × 按钮 | 有 keyword | 清空并恢复完整列表 | 不适用 |
| 点击门店卡片 | M3 | 有门店 | `wx.navigateTo` 门店详情页（带 id） | 不适用 |
| 点击"电话" | M4 | 有 contactPhone | `wx.makePhoneCall` | 无电话：toast "暂无联系电话" |
| 点击"预约" | M4 | 有门店 id | `wx.navigateTo` 预约创建页（带 storeId） | 不适用 |
| 点击"导航" | M4 | 门店有经纬度 | `wx.openLocation` 打开微信地图 | 无经纬度：toast "该门店暂未配置坐标" |
| 下拉刷新 | 页面顶部 | 无 | 重新加载门店 + 重新定位 | 不适用 |
| 点击"去设置" | M1 定位状态栏 | locError 有值 | `wx.openSetting` 引导开启定位 | 不适用 |

### 4.5 交互说明

- 进入页面时先拉门店列表，同时发起 `wx.getLocation` 定位。
- 门店列表接口不含经纬度，前端并发拉取每个门店的详情补齐坐标。
- 定位成功后按距离排序（无坐标的门店排最后），更新 distanceText。
- 定位失败时显示提示"定位服务未开启，无法获取附近门店"，提供"去设置"按钮。
- 拒绝定位后仍可手动浏览和搜索门店，不强制要求定位。
- 营业时间解析：从 businessHours 字段解析上下班时间，判断当前是否营业中（isOpen）。
- 搜索支持按门店名称、地址、城市模糊匹配。

### 4.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态 | 展示门店列表 + 搜索栏 + 定位状态 |
| 加载中 | 列表区域显示 loading |
| 空数据 | "暂无门店" |
| 接口失败 | 错误文案 + "重试"按钮 |
| 未登录 | 不阻塞 |
| 未授权手机号 | 不阻塞 |
| 定位失败 | 顶部提示"定位服务未开启" + "去设置"按钮，列表仍可浏览 |
| 无权限 | 不适用 |

### 4.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| stores | array | stores 接口 | 全部门店 | — |
| filteredStores | array | 前端过滤 | 当前显示的门店 | 按关键词过滤 |
| nearestStore | object | 前端计算 | 离你最近的门店 | 有距离的第一个 |
| stores[].id | string | stores 接口 | 跳转详情/预约 | — |
| stores[].name | string | 同上 | 门店名称 | — |
| stores[].city | string | 同上 | 所在城市 | — |
| stores[].address | string | 同上 | 门店地址 | — |
| stores[].contactPhone | string | 同上 | 联系电话 | — |
| stores[].businessHours | string | 同上 | 营业时间 | 如 "09:30-21:30" |
| stores[].longitude | number | 详情接口补齐 | 导航用 | 列表接口不返回 |
| stores[].latitude | number | 详情接口补齐 | 导航用 | 列表接口不返回 |
| stores[].distance | number | 定位计算 | 距离 km | 无定位或无坐标时为 null |
| stores[].distanceText | string | 前端格式化 | 距离文案 | "800m" / "1.2km" / 空 |
| stores[].isOpen | boolean | 前端计算 | 是否营业中 | 解析 businessHours |
| keyword | string | 用户输入 | 搜索关键词 | — |
| located | boolean | 前端状态 | 是否已定位 | — |
| locError | string|null | 前端状态 | 定位错误信息 | 有值时显示提示 |
| loading | boolean | 前端状态 | 加载中标记 | — |
| error | string|null | 前端状态 | 接口错误信息 | — |

### 4.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/stores` | GET | 门店列表 | 否 | guest |
| `/api/v1/customer/stores/{id}` | GET | 门店详情（补经纬度） | 否 | guest |

### 4.9 跳转关系

```
门店列表页 → 门店详情页（navigateTo，带 id）
门店列表页 → 预约创建页（navigateTo，带 storeId）
门店列表页 → 首页（switchTab）
门店列表页 → 款式列表页（switchTab）
门店列表页 → 我的页（switchTab）
```

### 4.10 验收标准

1. 用户进入门店 Tab 后，看到门店列表，每个门店显示名称、地址、电话、营业时间和距离。
2. 定位成功后门店按距离从近到远排序，距离文案正确显示（<1km 显示 m，≥1km 显示 km）。
3. 用户拒绝定位后，页面仍可手动浏览和搜索门店，顶部显示定位提示。
4. 搜索框支持按门店名称、地址、城市模糊搜索，实时过滤列表。
5. 点击"电话"按钮调用 `wx.makePhoneCall`，无电话时 toast 提示。
6. 点击"导航"按钮调用 `wx.openLocation`，门店无经纬度时 toast 提示。
7. 点击"预约"按钮跳转预约创建页，URL 携带 storeId。
8. 下拉刷新后重新加载门店列表和定位。

---

## 页面 5：到店预约页

### 5.1 页面目标

让顾客选择服务类型、门店、日期、时段，填写联系人信息和备注，完成微信登录和手机号授权后提交预约。

### 5.2 页面入口

- 首页"到店预约"按钮 `navigateTo`（可带 storeId）
- 门店列表页"预约"按钮 `navigateTo`（带 storeId）
- 门店详情页"预约到店"按钮 `navigateTo`（带 storeId）
- 款式详情页"到店预约"按钮 `navigateTo`（可带 storeId）
- 回收介绍页"到店预约"按钮 `navigateTo`
- 我的预约页"新建预约"按钮 `navigateTo`

### 5.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 门店信息 | store.name, store.address, store.contactPhone | `GET /api/v1/customer/stores/{id}` | 是 | 是 | 无 storeId：提示"请选择门店" |
| M2 | 服务类型 | SERVICE_TYPES（4 种） | 前端常量 | 是 | 是 | 不适用（固定列表） |
| M3 | 日期选择 | next7Days()（7 天） | 前端日期工具 | 是 | 是 | 不适用 |
| M4 | 时段选择 | slots[].time, slots[].available, slots[].state | `GET /api/v1/customer/stores/{id}/slots?date=` | 是 | 是 | "该日暂无可选时段" |
| M5 | 联系人信息 | contactName, contactPhone | 用户输入 / 微信授权 | 是 | 是 | placeholder |
| M6 | 备注 | remark | 用户输入 | 否 | 是 | placeholder "选填，最多200字" |
| M7 | 手机号授权 | phoneVerified | 微信 getPhoneNumber | 是 | 是 | 显示"授权手机号"按钮 |
| M8 | 提交按钮 | submitting | 前端状态 | 是 | 是 | 提交中禁用 |

### 5.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击门店信息 | M1 | 有 storeId | `wx.navigateTo` 门店详情页 | 无 storeId：`switchTab` 门店列表 |
| 选择服务类型 | M2 | 无 | 高亮选中项 | 不适用 |
| 选择日期 | M3 | 无 | 高亮选中日期，重新加载时段 | 不适用 |
| 选择时段 | M4 | 时段未禁用 | 高亮选中时段 | 已禁用：不响应点击 |
| 输入联系人姓名 | M5 | 无 | 更新 contactName | 不适用 |
| 授权手机号 | M7 button open-type="getPhoneNumber" | 微信登录 | 手机号回填 contactPhone，phoneVerified=true | 拒绝授权：toast "需要手机号授权才能预约" |
| 输入备注 | M6 | 无 | 更新 remark | 超 200 字：截断 |
| 点击"提交预约" | M8 | 所有必填项已填 | toast "预约提交成功" → `wx.redirectTo` 预约详情页 | 校验失败：toast 对应提示 |

### 5.5 交互说明

- 日期选择为横向 7 天滚动条，今天高亮，点击切换后清空已选时段并重新请求 slots。
- 时段网格：09:30-21:30，30 分钟粒度，每行 4 个时段。
- 时段状态：
  - 可选：正常颜色，可点击。
  - 已过去（当天）：置灰，不可点击。
  - 已约满（available=false / state=full）：置灰 + "约满"标记，不可点击。
  - 已关闭（state=closed）：置灰 + "休息"标记，不可点击。
- 手机号授权按钮为 `<button open-type="getPhoneNumber">`，点击后拿到 phoneCode，调 `customerPhoneAuth(phoneCode)` 换取手机号。
- 如果未登录（无 token），先调 `ensureCustomerSession()` 静默登录，再授权手机号。
- 提交校验顺序：门店 → 服务类型 → 日期 → 时段 → 联系人姓名 → 手机号格式 → 手机号已授权。
- 重复预约（同手机号+同门店+同时段）：弹窗 "该时段已有有效预约，请选择其他时段"。
- 时段已满：弹窗 "该时段已约满，请选择其他时段"。
- 提交成功后用 `wx.redirectTo` 跳转预约详情页（不能返回到提交页）。

### 5.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态 | 展示门店信息 + 服务类型 + 日期 + 时段 + 联系人 + 备注 + 授权按钮 + 提交按钮 |
| 加载中（门店） | 门店信息区域 loading |
| 加载中（时段） | 时段网格区域 loading |
| 空数据（时段） | "该日暂无可选时段" |
| 接口失败（门店） | toast "门店加载失败" |
| 接口失败（时段） | 时段列表清空，不报错 |
| 未登录 | 可浏览页面结构，提交时先触发微信登录 |
| 未授权手机号 | 显示"授权手机号"按钮，contactPhone 为空 |
| 定位失败 | 不适用（门店由 URL 参数或手动选择） |
| 无权限 | 不适用（customer 可访问） |
| 提交中 | 提交按钮禁用 + loading 文案 |

### 5.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| storeId | string | URL 参数 | 门店 ID | 可为空（手动选择） |
| store | object | stores/{id} 接口 | 门店信息展示 | — |
| store.name | string | 同上 | 门店名称 | — |
| store.address | string | 同上 | 门店地址 | — |
| store.contactPhone | string | 同上 | 联系电话 | — |
| selectedService | string | 用户选择 | 服务类型 code | OLD_FOR_NEW/REPAIR/CONSULT/RECYCLE |
| days | array | next7Days() | 可选日期 | 7 天 |
| selectedDate | string | 用户选择 | 选中的日期 | YYYY-MM-DD |
| slots | array | stores/{id}/slots 接口 | 时段列表 | — |
| slots[].startTime | string | 同上 | 开始时间 | HH:MM |
| slots[].endTime | string | 前端计算 | 结束时间 | startTime + 30min |
| slots[].label | string | 前端组装 | 时段文案 | "09:30-10:00" |
| slots[].isPast | boolean | 前端计算 | 是否已过去 | 仅当天 |
| slots[].isFull | boolean | 接口/前端 | 是否约满 | available=false |
| slots[].isClosed | boolean | 接口 | 是否关闭 | state=closed |
| slots[].disabled | boolean | 前端计算 | 是否禁用 | isPast||isFull||isClosed |
| selectedSlot | string | 用户选择 | 选中的开始时间 | HH:MM |
| contactName | string | 用户输入/profile | 联系人姓名 | — |
| contactPhone | string | 手机号授权/profile | 联系人手机号 | 11 位 |
| remark | string | 用户输入 | 备注 | 最多 200 字 |
| isLoggedIn | boolean | 前端状态 | 是否已登录 | — |
| phoneVerified | boolean | 前端状态 | 手机号是否已授权 | — |
| profile | object | 本地存储 | 顾客信息 | — |
| submitting | boolean | 前端状态 | 提交中标记 | 防重复提交 |

### 5.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/stores/{id}` | GET | 门店详情 | 否 | guest |
| `/api/v1/customer/stores/{id}/slots` | GET | 时段可用性 | 否 | guest |
| `/api/v1/customer/auth/wechat-login` | POST | 微信登录 | 否 | guest |
| `/api/v1/customer/auth/phone` | POST | 手机号授权 | 是 | customer |
| `/api/v1/customer/appointments` | POST | 创建预约 | 是 | customer |

### 5.9 跳转关系

```
预约创建页 → 门店详情页（navigateTo，带 id）
预约创建页 → 门店列表页（switchTab，选择门店）
预约创建页 → 预约详情页（redirectTo，提交成功后）
```

### 5.10 验收标准

1. 用户进入预约页后，看到门店信息、服务类型选择、日期选择和时段选择。
2. 日期选择为未来 7 天，当天可选，点击切换后时段列表刷新。
3. 已过去的时段（当天当前时间之前）置灰不可点击。
4. 已约满的时段置灰并显示"约满"标记，不可点击。
5. 用户必须授权手机号后才能提交预约，拒绝授权时 toast 提示。
6. 提交校验顺序正确：门店 → 服务类型 → 日期 → 时段 → 姓名 → 手机号。
7. 同一手机号+同一门店+同一时段重复预约时，弹窗提示"该时段已有有效预约"。
8. 提交成功后 toast "预约提交成功"，然后 `redirectTo` 到预约详情页。
9. 提交过程中按钮禁用，防止重复提交。
10. 未登录用户提交时先触发微信登录，登录成功后继续提交流程。

---

## 页面 6：我的预约页

### 6.1 页面目标

让顾客查看自己的所有预约记录，按状态筛选，点击进入详情。

### 6.2 页面入口

- 我的页"我的预约" `navigateTo`（需登录）
- 预约详情页返回 `navigateBack`
- 预约创建页提交成功后不直接进入（用 redirectTo 到详情页）

### 6.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 状态筛选标签 | STATUS_TABS（全部/待确认/已确认/已完成/已取消） | 前端常量 | 是 | 是 | 不适用 |
| M2 | 预约卡片列表 | storeName, serviceType, appointmentDate, startTime, endTime, status, statusText, statusColor | `GET /api/v1/customer/appointments` | 是 | 是 | "暂无预约" |
| M3 | 新建预约 FAB | — | 固定按钮 | 是 | 是 | 不适用 |
| M4 | 未登录引导 | — | 前端判断 | 否 | 是 | "请先登录" + "去登录"按钮 |

### 6.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 切换状态标签 | M1 | 无 | filteredList 按状态过滤 | 无匹配：列表为空 |
| 点击预约卡片 | M2 | 有预约 | `wx.navigateTo` 预约详情页（带 id） | 不适用 |
| 点击"新建预约" | M3 FAB | 无 | `wx.navigateTo` 预约创建页 | 不适用 |
| 点击"去登录" | M4 | 未登录 | `wx.navigateTo` 预约创建页（触发登录流程） | 不适用 |
| 下拉刷新 | 页面顶部 | 已登录 | 重新加载预约列表 | 不适用 |

### 6.5 交互说明

- 状态标签栏：全部 / 待确认 / 已确认 / 已完成 / 已取消，选中项高亮。
- 预约卡片按创建时间倒序排列。
- 卡片显示：门店名称、服务类型文案、日期、时段、状态标签（带颜色）。
- 状态颜色：待确认=#FF9800，已确认=#4CAF50，已到店=#1976D2，已完成=#8A8F99，已取消=#B0B5BD，未到店/已终止=#F44336。
- 未登录时（无 token）不请求列表，显示"请先登录"引导。
- 从详情页返回时 `onShow` 重新加载列表（状态可能已变化）。
- 401 错误时标记 isLoggedIn=false，显示登录引导。

### 6.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态 | 展示状态标签 + 预约卡片列表 + 新建 FAB |
| 加载中 | 列表区域 loading |
| 空数据 | "暂无预约" |
| 接口失败 | 401 → 显示未登录引导；其他 → toast 错误 |
| 未登录 | "请先登录" + "去登录"按钮，不显示列表 |
| 未授权手机号 | 不适用（有 token 即可查看列表） |
| 定位失败 | 不适用 |
| 无权限 | 不适用 |

### 6.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| tabs | array | 前端常量 | 状态筛选标签 | 5 个 |
| activeTab | string | 前端状态 | 当前选中标签 | 默认空（全部） |
| list | array | appointments 接口 | 全部预约 | — |
| filteredList | array | 前端过滤 | 当前显示的预约 | 按 activeTab 过滤 |
| list[].id | string | 同上 | 跳转详情 | — |
| list[].storeName | string | 同上 | 门店名称 | — |
| list[].serviceType | string | 同上 | 服务类型 code | OLD_FOR_NEW 等 |
| list[].serviceText | string | 前端映射 | 服务类型文案 | "旧金换新" 等 |
| list[].appointmentDate | string | 同上 | 预约日期 | YYYY-MM-DD |
| list[].startTime | string | 同上 | 开始时间 | HH:MM |
| list[].endTime | string | 同上 | 结束时间 | HH:MM |
| list[].status | string | 同上 | 预约状态 | PENDING/CONFIRMED 等 |
| list[].statusText | string | 前端映射 | 状态文案 | "待确认" 等 |
| list[].statusColor | string | 前端映射 | 状态颜色 | — |
| loading | boolean | 前端状态 | 加载中标记 | — |
| isLoggedIn | boolean | 前端状态 | 是否已登录 | — |

### 6.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/appointments` | GET | 预约列表 | 是 | customer |

### 6.9 跳转关系

```
我的预约页 → 预约详情页（navigateTo，带 id）
我的预约页 → 预约创建页（navigateTo，FAB 或"去登录"）
我的预约页 → 返回我的页（navigateBack）
```

### 6.10 验收标准

1. 已登录用户进入我的预约页后，看到预约卡片列表，按创建时间倒序。
2. 状态标签栏可切换筛选，列表内容跟随过滤。
3. 预约卡片显示门店名称、服务类型、日期、时段和状态标签（带正确颜色）。
4. 未登录用户看到"请先登录"引导和"去登录"按钮，不显示列表。
5. 点击预约卡片进入预约详情页，URL 携带预约 id。
6. 从详情页返回后列表自动刷新（状态可能已变化）。
7. 无预约时显示"暂无预约"。
8. 下拉刷新后列表重新加载。

---

## 页面 7：预约详情页

### 7.1 页面目标

展示单个预约的完整信息，让顾客查看状态、取消预约、修改备注、联系门店和导航到店。

### 7.2 页面入口

- 我的预约页点击预约卡片 `navigateTo`（带 id）
- 预约创建页提交成功后 `redirectTo`（带 id）
- 预约备注页返回 `navigateBack`

### 7.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 状态头部 | status, statusText, statusColor | `GET /api/v1/customer/appointments/{id}` | 是 | 否 | 不适用 |
| M2 | 预约信息 | storeName, serviceType, serviceText, appointmentDate, startTime, endTime | 同上 | 是 | 否 | 不适用 |
| M3 | 门店信息 | storeName, storeAddress, storePhone | 同上 | 是 | 是 | 不适用 |
| M4 | 联系人信息 | contactName, contactPhone | 同上 | 是 | 否 | 不适用 |
| M5 | 备注 | remark | 同上 | 否 | 是（可编辑时） | "无备注" |
| M6 | 时间信息 | createdAt, confirmedAt, cancelledAt | 同上 | 否 | 否 | 不显示对应行 |
| M7 | 底部操作栏 | 取消预约 / 修改备注 / 拨打电话 / 导航 | 固定按钮 | 是 | 是 | 不适用 |

### 7.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击"取消预约" | M7 | 状态为 PENDING/CONFIRMED 且距预约 >2h | 弹窗二次确认 → 调 cancelAppointment → 刷新详情 | 距预约 ≤2h：不显示按钮 |
| 点击"修改备注" | M7 | 状态为 PENDING/CONFIRMED | `wx.navigateTo` 预约备注页（带 id + remark） | 其他状态：不显示按钮 |
| 点击"拨打电话" | M7 | 有 storePhone | `wx.makePhoneCall` | 无电话：toast "暂无门店电话" |
| 点击"导航到店" | M7 | 有 storeLongitude/latitude | `wx.openLocation` | 无坐标：toast "该门店暂未配置位置" |
| 返回 | 导航栏返回 | 无 | `wx.navigateBack` | 不适用 |

### 7.5 交互说明

- 取消预约需要二次确认：`wx.showModal` "确定要取消此预约吗？取消后不可恢复。"，确认后调接口。
- 取消条件：状态为 PENDING 或 CONFIRMED，且预约时间距当前 >2 小时。
- 不满足取消条件时不显示"取消预约"按钮。
- 修改备注条件：状态为 PENDING 或 CONFIRMED。
- 不满足修改条件时不显示"修改备注"按钮。
- 从备注页返回时 `onShow` 重新加载详情（备注可能已修改）。
- 状态头部用 statusColor 作为背景色，statusText 白色文字。

### 7.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态 | 展示状态头部 + 预约信息 + 门店信息 + 联系人 + 备注 + 时间 + 操作栏 |
| 加载中 | 全屏 loading |
| 空数据 | 不适用（有 id 必有数据） |
| 接口失败 | 错误文案 + "重试"按钮 |
| 未登录 | 401 → toast "请先登录" → 返回上一页 |
| 未授权手机号 | 不适用 |
| 定位失败 | 不适用（导航使用门店坐标） |
| 无权限 | 不适用（customer 只能看自己的预约） |

### 7.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| id | string | URL 参数 | 请求详情 | — |
| detail | object | appointments/{id} 接口 | 预约详情 | — |
| detail.status | string | 同上 | 预约状态 | PENDING/CONFIRMED/ARRIVED/COMPLETED/CANCELLED/NO_SHOW/TERMINATED |
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
| detail.appointmentNo | string | 同上 | 预约编号 | YY+YYYYMMDD+4位序号 |
| statusText | string | 前端映射 | 状态文案 | — |
| statusColor | string | 前端映射 | 状态颜色 | — |
| serviceText | string | 前端映射 | 服务类型文案 | — |
| canCancel | boolean | 前端计算 | 是否可取消 | 状态+时间判定 |
| canEditNotes | boolean | 前端计算 | 是否可改备注 | 状态判定 |
| loading | boolean | 前端状态 | 加载中标记 | — |
| error | string|null | 前端状态 | 错误信息 | — |

### 7.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/appointments/{id}` | GET | 预约详情 | 是 | customer |
| `/api/v1/customer/appointments/{id}/cancel` | POST | 取消预约 | 是 | customer |
| `/api/v1/customer/appointments/{id}/notes` | PUT | 修改备注（跳转备注页执行） | 是 | customer |

### 7.9 跳转关系

```
预约详情页 → 预约备注页（navigateTo，带 id + remark）
预约详情页 → 返回我的预约页（navigateBack）
预约详情页 → 预约创建页（不跳转）
```

### 7.10 验收标准

1. 用户从我的预约点击进入后，看到预约完整信息：状态、门店、服务类型、日期、时段、联系人、备注。
2. 状态头部用对应颜色显示状态文案。
3. 状态为 PENDING/CONFIRMED 且距预约 >2h 时显示"取消预约"按钮，否则不显示。
4. 取消预约需要二次确认弹窗，确认后调接口并刷新详情。
5. 状态为 PENDING/CONFIRMED 时显示"修改备注"按钮，点击跳转备注页。
6. 点击"拨打电话"调用 `wx.makePhoneCall`，无电话时 toast 提示。
7. 点击"导航到店"调用 `wx.openLocation`，门店无坐标时 toast 提示。
8. 从备注页返回后自动刷新详情，备注内容更新。
9. 预约编号格式为 YY+YYYYMMDD+4位序号。

---

## 页面 8：我的页 / 店长入口

### 8.1 页面目标

展示顾客个人信息，提供我的预约、回收介绍、客服电话入口，底部提供"店长入口"供门店管理人员登录管理端。

### 8.2 页面入口

- TabBar 第 4 项"我的"
- 其他 Tab 切回我的

### 8.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 顾客头部 | profile.nickname, profile.phone（脱敏）| 本地存储 / `GET /api/v1/customer/me` | 否 | 是 | 未登录："点击登录" |
| M2 | 功能菜单 | 我的预约 / 回收介绍 | 固定入口 | 是 | 是 | 不适用 |
| M3 | 客服电话 | servicePhone | app.globalData | 是 | 是 | "暂无客服电话" |
| M4 | 店长入口 | "店长入口" + 副文案 "门店店长 / 管理员登录" | 固定入口 | 是 | 是 | 不适用 |
| M5 | 退出登录 | — | 前端判断 | 否 | 是 | 未登录不显示 |
| M6 | TabBar | 同首页 | app.json | 是 | 是 | 不适用 |

### 8.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击"点击登录" | M1 头部 | 未登录 | `wx.navigateTo` 预约创建页（触发登录） | 不适用 |
| 手机号授权登录 | M1 button open-type="getPhoneNumber" | 未登录 | 静默登录 → 手机号授权 → 更新头部 | 授权失败：toast |
| 点击"我的预约" | M2 | 已登录 | `wx.navigateTo` 我的预约页 | 未登录：`navigateTo` 预约创建页 |
| 点击"回收介绍" | M2 | 无 | `wx.navigateTo` 回收介绍页 | 不适用 |
| 点击"客服电话" | M3 | 有 servicePhone | `wx.makePhoneCall` | 脱敏号：toast "请通过门店页面拨打" |
| 点击"店长入口" | M4 | 无 | 弹窗提示（N08 对接后跳转店长登录页） | 不适用 |
| 点击"退出登录" | M5 | 已登录 | 二次确认 → 调 logout → 清除本地 → 更新头部 | 不适用 |
| 切换 Tab | M6 | 无 | 切换到对应 Tab | 不适用 |

### 8.5 交互说明

- 顾客头部：
  - 未登录：显示默认头像 + "点击登录"文案 + 手机号授权按钮。
  - 已登录：显示微信昵称（或"顾客"）+ 脱敏手机号（如 138****1234）。
- 手机号脱敏：中间 4 位用 * 替换。
- `onShow` 时从本地存储读取 profile 并展示，同时静默调 `getMe()` 刷新。
- "我的预约"在未登录时点击会跳转预约创建页（间接引导登录）。
- "店长入口"文案必须叫"店长入口"，副文案为"门店店长 / 管理员登录"。
- 店长入口当前为弹窗提示（N08 员工端改造时对接 `wx.reLaunch` 到员工端登录页）。
- 退出登录需二次确认弹窗。

### 8.6 页面状态

| 状态 | 表现 |
|------|------|
| 默认状态（已登录） | 头像 + 昵称 + 脱敏手机号 + 菜单 + 客服 + 店长入口 + 退出登录 |
| 默认状态（未登录） | 默认头像 + "点击登录" + 手机号授权按钮 + 菜单 + 客服 + 店长入口 |
| 加载中 | getMe() 刷新中，头部显示本地缓存的 profile |
| 空数据 | 不适用 |
| 接口失败 | getMe() 失败时使用本地存储的 profile，不报错 |
| 未登录 | 显示"点击登录"和手机号授权按钮 |
| 未授权手机号 | 已登录但无手机号时，仍显示授权按钮 |
| 定位失败 | 不适用 |
| 无权限 | 不适用（guest 可访问） |

### 8.7 数据字段

| 字段 | 类型 | 来源 | 用途 | 备注 |
|------|------|------|------|------|
| isLoggedIn | boolean | 前端状态 | 是否已登录 | — |
| profile | object | 本地存储 / getMe() | 顾客信息 | — |
| profile.nickname | string | 同上 | 昵称 | 无则显示"顾客" |
| profile.phone | string | 同上 | 手机号 | 脱敏显示 |
| servicePhone | string | app.globalData | 客服电话 | 全局配置 |

### 8.8 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/me` | GET | 刷新顾客信息 | 是 | customer |
| `/api/v1/customer/auth/wechat-login` | POST | 微信登录 | 否 | guest |
| `/api/v1/customer/auth/phone` | POST | 手机号授权 | 是 | customer |
| `/api/v1/customer/auth/logout` | POST | 退出登录 | 是 | customer |

### 8.9 跳转关系

```
我的页 → 我的预约页（navigateTo）
我的页 → 预约创建页（navigateTo，未登录时引导登录）
我的页 → 回收介绍页（navigateTo）
我的页 → 店长登录页（N08 对接，当前为弹窗提示）
我的页 → 首页（switchTab）
我的页 → 款式列表页（switchTab）
我的页 → 门店列表页（switchTab）
```

### 8.10 验收标准

1. 未登录用户进入"我的"Tab 后，看到默认头像和"点击登录"文案。
2. 已登录用户看到微信昵称和脱敏手机号（中间 4 位用 * 替换）。
3. 手机号授权按钮为 `<button open-type="getPhoneNumber">`，授权成功后更新头部。
4. "我的预约"在未登录时点击跳转预约创建页（间接引导登录），已登录时跳转我的预约页。
5. "店长入口"文案必须叫"店长入口"，副文案为"门店店长 / 管理员登录"，不叫"员工入口"。
6. 点击"店长入口"后显示弹窗提示（N08 对接后跳转店长登录页）。
7. "客服电话"在号码为脱敏号时 toast 提示"请通过门店页面拨打"。
8. "退出登录"需二次确认，确认后清除本地 token 和 profile。
9. `onShow` 时从本地存储读取 profile 并展示，同时静默调 `getMe()` 刷新。
10. 顾客不能看到任何店长管理功能入口之外的店长管理页面。

---

## 附：门店详情页（辅助页面）

> 门店详情页不属于 8 个核心原型页面之一，但在多个页面跳转链路中出现，此处补充说明。

### A.1 页面目标

展示单个门店的完整信息（名称、地址、电话、营业时间、经纬度、今日时段），让顾客拨打电话、导航到店或发起预约。

### A.2 页面入口

- 首页附近门店卡片 `navigateTo`（带 id）
- 门店列表页门店卡片 `navigateTo`（带 id）
- 款式详情页最近门店卡片 `navigateTo`（带 id）
- 预约创建页门店信息 `navigateTo`（带 id）

### A.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 门店基本信息 | name, city, address, contactPhone, businessHours | `GET /api/v1/customer/stores/{id}` | 是 | 否 | 不适用 |
| M2 | 距离信息 | distanceText | 定位计算 | 否 | 否 | 无定位不显示 |
| M3 | 今日时段 | todaySlots[].time, state | `GET /api/v1/customer/stores/{id}/slots?date=today` | 否 | 否 | "今日暂无时段" |
| M4 | 底部操作栏 | 拨打电话 / 导航到店 / 预约到店 | 固定按钮 | 是 | 是 | 不适用 |

### A.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击"拨打电话" | M4 | 有 contactPhone | `wx.makePhoneCall` | 无电话：toast "暂无联系电话" |
| 点击"导航到店" | M4 | 有 longitude/latitude | `wx.openLocation` | 无坐标：toast "该门店暂未配置位置" |
| 点击"预约到店" | M4 | 无 | `wx.navigateTo` 预约创建页（带 storeId） | 不适用 |

### A.5 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/stores/{id}` | GET | 门店详情 | 否 | guest |
| `/api/v1/customer/stores/{id}/slots` | GET | 今日时段 | 否 | guest |

### A.6 跳转关系

```
门店详情页 → 预约创建页（navigateTo，带 storeId）
门店详情页 → 返回上一页（navigateBack）
```

### A.7 验收标准

1. 用户进入门店详情页后，看到门店名称、地址、电话、营业时间。
2. 有定位时显示距离，无定位时不显示距离但不报错。
3. 今日时段展示前 6 个非关闭时段。
4. 门店无经纬度时点击"导航到店"toast 提示"该门店暂未配置位置"。
5. 点击"预约到店"跳转预约创建页，URL 携带 storeId。

---

## 附：预约备注页（辅助页面）

### B.1 页面目标

让顾客修改预约的备注信息（仅 PENDING/CONFIRMED 状态可编辑）。

### B.2 页面入口

- 预约详情页"修改备注" `navigateTo`（带 id + remark）

### B.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 备注 textarea | remark | URL 参数 / 用户输入 | 否 | 是 | placeholder "选填，最多200字" |
| M2 | 保存按钮 | — | 固定按钮 | 是 | 是 | 保存中禁用 |

### B.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 输入备注 | M1 textarea | 无 | 更新 remark | 超 200 字截断 |
| 点击"保存" | M2 | 有修改 | `PUT /api/v1/customer/appointments/{id}/notes` → toast "保存成功" → `navigateBack` | 接口失败：toast |

### B.5 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/appointments/{id}/notes` | PUT | 修改备注 | 是 | customer |

### B.6 验收标准

1. 进入备注页后，textarea 预填当前备注内容。
2. 输入超过 200 字时截断。
3. 点击保存后调 PUT 接口，成功后 toast "保存成功"并返回详情页。
4. 保存中按钮禁用，防止重复提交。
5. 接口失败时 toast 提示错误信息。

---

## 附：回收介绍页（辅助页面）

### C.1 页面目标

向顾客介绍黄金回收服务的流程、服务类型和注意事项，引导顾客到店预约。

### C.2 页面入口

- 首页 Banner `navigateTo`（linkType=recycle）
- 我的页"回收介绍" `navigateTo`

### C.3 信息结构

| 序号 | 模块名称 | 展示字段 | 字段来源 | 必填 | 可点击 | 空数据显示 |
|------|----------|----------|----------|------|--------|------------|
| M1 | 标题 | title | `GET /api/v1/customer/recycle-info` | 是 | 否 | 降级 "黄金回收服务" |
| M2 | 简介 | intro | 同上 | 是 | 否 | 降级文案 |
| M3 | 服务流程 | process[].step, title, desc | 同上 | 是 | 否 | 降级 4 步流程 |
| M4 | 服务类型 | services[].icon, title, desc | 同上 | 是 | 否 | 降级 4 类服务 |
| M5 | 注意事项 | notices[] | 同上 | 是 | 否 | 降级注意事项 |
| M6 | 预约 CTA | — | 固定按钮 | 是 | 是 | 不适用 |

### C.4 操作说明

| 操作 | 触发位置 | 前置条件 | 成功结果 | 失败提示 |
|------|----------|----------|----------|----------|
| 点击"到店预约" | M6 | 无 | `wx.navigateTo` 预约创建页 | 不适用 |
| 返回 | 导航栏返回 | 无 | `wx.navigateBack`（无上一页则 switchTab 首页） | 不适用 |

### C.5 交互说明

- 接口失败时降级为本地默认文案，不报错。
- 服务流程为 4 步：到店预约 → 实物检测 → 确认价格 → 完成回收。
- 服务类型为 4 类：旧金换新 / 黄金维修 / 到店回收 / 款式咨询。
- 注意事项明确："最终回收价格以门店实物检测为准"和"V1版本不提供上门回收和邮寄回收服务"。

### C.6 接口依赖

| 接口 | 方法 | 用途 | 需登录 | 权限 |
|------|------|------|--------|------|
| `/api/v1/customer/recycle-info` | GET | 回收介绍 | 否 | guest |

### C.7 验收标准

1. 用户进入回收介绍页后，看到服务标题、简介、流程、服务类型和注意事项。
2. 接口失败时降级为本地默认文案，不显示错误。
3. 注意事项中明确标注"最终回收价格以门店实物检测为准"。
4. 点击"到店预约"跳转预约创建页。
5. 无上一页时返回首页（switchTab）。

---

## 全局跳转关系图

```
                    ┌──────────┐
                    │  顾客首页  │ ← TabBar
                    └────┬─────┘
           ┌──────┬─────┼─────┬──────┐
           ▼      ▼     ▼     ▼      ▼
        款式列表  回收介绍  门店列表  预约创建  我的页
           │      │       │       │       │
           ▼      │       ▼       │       ▼
        款式详情  │     门店详情    │     我的预约
                   │       │       │       │
                   └───────┴──→ 预约创建  预约详情
                                      │       │
                                      ▼       ▼
                                   微信登录  预约备注
                                      │
                                      ▼
                                   手机号授权
                                      │
                                      ▼
                                   提交预约 → 预约详情

         我的页 ──底部──→ 店长入口 ──→ 店长登录 ──→ 工作台（N08 对接）
```
