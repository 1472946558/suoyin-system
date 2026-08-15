# 07 - 数据模型草案

> 基于后端实际代码（types.go, customer_types.go, store.go, persistence.go）编写。
> 数据层已实现，以下为已落地和待落地的模型说明。

## 1. 已有模型（无需修改）

### 1.1 StoreInfo（门店信息）

**存储位置**：`app_configs` 表，JSON key `stores`

```go
type StoreInfo struct {
    ID            string  `json:"id"`
    OrgID         string  `json:"orgId"`
    Code          string  `json:"code"`
    Name          string  `json:"name"`
    City          string  `json:"city"`
    Address       string  `json:"address"`
    Manager       string  `json:"manager"`
    Status        string  `json:"status"`         // active / inactive
    IsDefault     bool    `json:"isDefault"`
    Longitude     float64 `json:"longitude"`      // ← 已添加
    Latitude      float64 `json:"latitude"`       // ← 已添加
    ContactPhone  string  `json:"contactPhone"`   // ← 已添加
    BusinessHours string  `json:"businessHours"`  // ← 已添加
}
```

**状态**：Longitude/Latitude/ContactPhone/BusinessHours 字段已在 types.go 中添加，无需修改。

### 1.2 CatalogProduct（商品信息）

**存储位置**：`app_configs` 表，JSON key `catalog_products`

```go
type CatalogProduct struct {
    ID               string   `json:"id"`
    OrgID            string   `json:"orgId"`
    Name             string   `json:"name"`
    SKU              string   `json:"sku"`           // 顾客端不返回
    Category         string   `json:"category"`
    CategoryTab      string   `json:"categoryTab"`
    ImageURL         string   `json:"imageUrl"`
    Purity           string   `json:"purity"`
    BenchPrice       float64  `json:"benchPrice"`    // 顾客端不返回
    RetailPrice      float64  `json:"retailPrice"`
    GramWeight       float64  `json:"gramWeight"`
    Status           string   `json:"status"`        // active / inactive
    Inventory        int      `json:"inventory"`     // 顾客端不返回
    StockStatus      string   `json:"stockStatus"`   // 顾客端不返回
    StoreIDs         []string `json:"storeIds"`      // 顾客端不返回
    Stores           []string `json:"stores"`
    Tags             []string `json:"tags"`
    RecommendedScene string   `json:"recommendedScene"`
    QuoteLeadTime    string   `json:"quoteLeadTime"`
}
```

**顾客端 DTO 白名单**（CustomerProductDTO）：只返回 id, name, imageUrl, category, purity, retailPrice, gramWeight, recommendedScene, tags。

## 2. 新增模型（已实现）

### 2.1 CustomerProfile（顾客档案）

**存储位置**：`app_configs` 表，JSON key `customer_profiles`

```go
type CustomerProfile struct {
    ID        string    `json:"id"`         // customer-xxx
    OrgID     string    `json:"orgId"`
    OpenID    string    `json:"openId"`     // 微信 openid
    UnionID   string    `json:"unionId"`    // 可选
    Phone     string    `json:"phone"`      // 手机号
    Nickname  string    `json:"nickname"`   // 可选
    AvatarURL string    `json:"avatarUrl"`  // 可选
    Status    string    `json:"status"`     // active / disabled
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

**状态**：已在 customer_types.go 中定义，customer_store.go 中已实现 CRUD。

### 2.2 CustomerSession（顾客会话）

**存储位置**：Redis，key 前缀 `customer_session:`

```go
type CustomerSession struct {
    Token      string    `json:"token"`
    CustomerID string    `json:"customerId"`
    OrgID      string    `json:"orgId"`
    Phone      string    `json:"phone"`
    LoginAt    time.Time `json:"loginAt"`
    ExpiresAt  time.Time `json:"expiresAt"`  // 12 小时有效期
}
```

**状态**：已实现。token 生成方式：HMAC-SHA256(secret, "customer:{id}:{timestamp}")。

### 2.3 CustomerAppointment（到店预约）

**存储位置**：MySQL 表 `customer_appointments`（待建表）或 `app_configs` JSON key

```go
type CustomerAppointment struct {
    ID              string     `json:"id"`              // appt-{seq}-{id}
    OrgID           string     `json:"orgId"`
    CustomerID      string     `json:"customerId"`
    CustomerName    string     `json:"customerName"`
    CustomerPhone   string     `json:"customerPhone"`
    StoreID         string     `json:"storeId"`
    StoreName       string     `json:"storeName"`
    StoreAddress    string     `json:"storeAddress"`
    StorePhone      string     `json:"storePhone"`
    AppointmentDate string     `json:"appointmentDate"`  // YYYY-MM-DD
    AppointmentTime string     `json:"appointmentTime"`  // HH:MM
    Status          string     `json:"status"`
    Remark          string     `json:"remark"`
    CreatedAt       time.Time  `json:"createdAt"`
    UpdatedAt       time.Time  `json:"updatedAt"`
    CancelledAt     *time.Time `json:"cancelledAt"`
    CancelReason    string     `json:"cancelReason"`
    ConfirmedAt     *time.Time `json:"confirmedAt"`
    ConfirmedBy     string     `json:"confirmedBy"`
    CompletedAt     *time.Time `json:"completedAt"`
}
```

**状态**：已在 customer_types.go 中定义，customer_store.go 中已实现完整 CRUD。

## 3. 预约状态机

### 7 态状态枚举

```go
const (
    AppointmentStatusPending    = "PENDING"     // 待确认
    AppointmentStatusConfirmed  = "CONFIRMED"   // 已确认
    AppointmentStatusArrived    = "ARRIVED"     // 已到店
    AppointmentStatusCompleted  = "COMPLETED"   // 已完成
    AppointmentStatusCancelled  = "CANCELLED"   // 已取消
    AppointmentStatusNoShow     = "NO_SHOW"     // 未到店
    AppointmentStatusTerminated = "TERMINATED"  // 已终止
)
```

### 状态转换矩阵

| 当前状态 | 可转换到 | 操作人 | 校验规则 |
|----------|----------|--------|----------|
| PENDING | CONFIRMED | 员工+ | — |
| PENDING | ARRIVED | 员工+ | — |
| PENDING | CANCELLED | 顾客/员工+ | 顾客取消需 ≥2h |
| PENDING | NO_SHOW | 员工+ | — |
| PENDING | TERMINATED | 员工+ | — |
| CONFIRMED | ARRIVED | 员工+ | — |
| CONFIRMED | CANCELLED | 顾客/员工+ | 顾客取消需 ≥2h |
| CONFIRMED | NO_SHOW | 员工+ | — |
| CONFIRMED | TERMINATED | 员工+ | — |
| ARRIVED | COMPLETED | 员工+ | — |
| ARRIVED | TERMINATED | 员工+ | — |
| COMPLETED | — | — | 终态 |
| CANCELLED | — | — | 终态 |
| NO_SHOW | — | — | 终态 |
| TERMINATED | — | — | 终态 |

### 取消规则（已实现）

```go
func canCancelAppointment(appt CustomerAppointment, now time.Time) bool {
    if appt.Status != AppointmentStatusPending && appt.Status != AppointmentStatusConfirmed {
        return false
    }
    apptTime, err := time.ParseInLocation("2006-01-02 15:04",
        appt.AppointmentDate+" "+appt.AppointmentTime, time.Local)
    if err != nil {
        return false
    }
    return now.Add(2 * time.Hour).Before(apptTime) ||
           now.Add(2*time.Hour).Equal(apptTime)
}
```

## 4. 配置键清单

存储在 `app_configs` 表中的 JSON 键：

| config_key | 说明 | 状态 |
|------------|------|------|
| `customer_profiles` | 顾客档案列表 | 已实现 |
| `customer_home_config` | 顾客首页配置（轮播图、品牌文案） | 已实现 |
| `customer_recycle_info` | 回收介绍文案 | 已实现 |
| `customer_appointments` | 预约数据（如果不用独立表） | 已实现 |
| `stores` | 门店列表（含经纬度） | 已有 |
| `catalog_products` | 商品列表 | 已有 |

## 5. 建议索引

### customer_appointments 表（如果迁移到独立 MySQL 表）

```sql
CREATE TABLE customer_appointments (
    id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL,
    customer_id VARCHAR(64) NOT NULL,
    customer_name VARCHAR(100) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    store_id VARCHAR(64) NOT NULL,
    store_name VARCHAR(200) NOT NULL,
    store_address VARCHAR(500),
    store_phone VARCHAR(20),
    appointment_date DATE NOT NULL,
    appointment_time VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    remark TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    cancelled_at DATETIME,
    cancel_reason TEXT,
    confirmed_at DATETIME,
    confirmed_by VARCHAR(100),
    completed_at DATETIME,
    INDEX idx_customer (customer_id, created_at DESC),
    INDEX idx_store_date (store_id, appointment_date, appointment_time),
    INDEX idx_phone_store_time (customer_phone, store_id, appointment_date, appointment_time),
    INDEX idx_status (status, appointment_date)
);
```

### customer_profiles 表（如果迁移到独立 MySQL 表）

```sql
CREATE TABLE customer_profiles (
    id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL,
    open_id VARCHAR(100) NOT NULL UNIQUE,
    union_id VARCHAR(100),
    phone VARCHAR(20),
    nickname VARCHAR(200),
    avatar_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_open_id (open_id),
    INDEX idx_phone (phone)
);
```

## 6. ER 关系图

```
┌─────────────────┐     ┌──────────────────────┐     ┌──────────────────┐
│ CustomerProfile │     │ CustomerAppointment  │     │ StoreInfo        │
│─────────────────│     │──────────────────────│     │──────────────────│
│ id (PK)         │◄──┐ │ id (PK)              │ ┌──►│ id (PK)          │
│ org_id          │   │ │ customer_id (FK)     │─┘   │ org_id           │
│ open_id         │   └─│ store_id (FK)        │     │ name             │
│ phone           │     │ appointment_date     │     │ longitude        │
│ nickname        │     │ appointment_time     │     │ latitude         │
│ status          │     │ status               │     │ contact_phone    │
└─────────────────┘     │ confirmed_by         │     │ business_hours   │
                        └──────────────────────┘     └──────────────────┘

┌─────────────────┐
│ CatalogProduct  │     ┌──────────────────────┐
│─────────────────│     │ CustomerSession      │
│ id (PK)         │     │──────────────────────│     Redis
│ name            │     │ token (PK)           │     customer_session:<token>
│ category        │     │ customer_id          │
│ retail_price    │     │ expires_at           │
│ status          │     └──────────────────────┘
└─────────────────┘
```

## 7. 当前存储方式

当前所有顾客数据存储在 `app_configs` 表的 JSON 字段中，与员工/门店/商品数据共享同一表。这种方式的特点：

**优点**：
- 无需数据库迁移
- 快速开发
- 数据加载简单

**风险**：
- 并发写入可能丢失数据（JSON blob 整体覆盖）
- 无行级锁
- 查询效率随数据量增长下降

**建议**：V1 先用 JSON 存储（快速上线），V2 迁移到独立 MySQL 表（如果预约量增长）。
