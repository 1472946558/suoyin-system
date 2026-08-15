/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: types.go
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-05-10
 */

package app

import "time"

type Config struct {
	Host                   string
	Port                   string
	Mode                   string
	TokenSecret            string
	CORSOrigin             string
	MySQLDSN               string
	RedisAddr              string
	RedisUser              string
	RedisPass              string
	RedisDB                int
	RedisPrefix            string
	WechatMiniAppAppID     string
	WechatMiniAppAppSecret string
	WechatAPIBaseURL       string
	MiniAppAllowMockLogin  bool
	StorageEnabled         bool
	StorageEnabledSet      bool
	StorageProvider        string
	StorageBucket          string
	StorageRegion          string
	StorageEndpoint        string
	StoragePublicBaseURL   string
	StoragePathPrefix      string
	StorageUploadStrategy  string
	StorageAccessKeyID     string
	StorageAccessKeySecret string
	StorageUploadURLTTL    int
	StorageCallbackEnabled bool
	StorageCallbackSet     bool
	StorageStatus          string
	GoldPriceLiveEnabled   bool
	GoldPriceAPIURL        string
	GoldFXAPIURL           string
	GoldPriceCacheTTL      int
	GoldPriceHTTPTimeout   int
	GoldPriceCNYPerGram    float64
	GoldPriceXAUUSD        float64
	GoldPriceUSDCNY        float64
	GoldPriceUpdatedAt     string
	AssetsDir              string // 后台素材上传本地目录（ASSETS_DIR）
	AssetsPublicBaseURL    string // 素材公网地址前缀（ASSETS_PUBLIC_BASE_URL），如 https://jinjiangguan.com
}

func (c Config) ListenAddr() string {
	if c.Host == "" {
		return ":" + c.Port
	}
	return c.Host + ":" + c.Port
}

type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId"`
}

type StoreInfo struct {
	ID        string `json:"id"`
	OrgID     string `json:"orgId"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	City      string `json:"city"`
	Address   string `json:"address"`
	Manager   string `json:"manager"`
	Status    string `json:"status"`
	IsDefault     bool    `json:"isDefault"`
	Longitude     float64 `json:"longitude,omitempty"`
	Latitude      float64 `json:"latitude,omitempty"`
	ContactPhone  string  `json:"contactPhone,omitempty"`
	BusinessHours string  `json:"businessHours,omitempty"`
	// 顾客端扩展（后台「顾客端内容管理」）
	ImageURL           string   `json:"imageUrl,omitempty"`           // 门店图片
	AppointmentEnabled *bool    `json:"appointmentEnabled,omitempty"` // 是否支持预约（nil = 可预约）
	ServiceTags        []string `json:"serviceTags,omitempty"`        // 门店服务标签
	SortOrder          int      `json:"sortOrder,omitempty"`          // 门店排序
}

type PermissionItem struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionGroup struct {
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Items       []PermissionItem `json:"items"`
}

type RoleTemplate struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DataScope   string   `json:"dataScope"`
	Permissions []string `json:"permissions"`
}

type UserAccount struct {
	ID           string   `json:"id"`
	OrgID        string   `json:"orgId"`
	Username     string   `json:"username"`
	DisplayName  string   `json:"displayName"`
	Phone        string   `json:"phone,omitempty"`
	Password     string   `json:"password,omitempty"`
	RoleCode     string   `json:"roleCode"`
	RoleName     string   `json:"roleName"`
	DataScope    string   `json:"dataScope"`
	StoreIDs     []string `json:"storeIds"`
	Permissions  []string `json:"permissions"`
	Status       string   `json:"status"`
	WechatOpenID string   `json:"wechatOpenId,omitempty"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	LoginAt   time.Time `json:"loginAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type CashierOrderLine struct {
	ProductID string  `json:"productId,omitempty"`
	SKU       string  `json:"sku,omitempty"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
	Amount    float64 `json:"amount"`
}

type CashierOrder struct {
	ID            string             `json:"id"`
	OrderNo       string             `json:"orderNo"`
	OrgID         string             `json:"orgId"`
	StoreID       string             `json:"storeId"`
	StoreName     string             `json:"storeName"`
	Status        string             `json:"status"`
	CustomerName  string             `json:"customerName"`
	CustomerPhone string             `json:"customerPhone"`
	PaymentMethod string             `json:"paymentMethod"`
	TotalAmount   float64            `json:"totalAmount"`
	PaidAmount    float64            `json:"paidAmount"`
	Remark        string             `json:"remark"`
	Items         []CashierOrderLine `json:"items"`
	CreatedBy     string             `json:"createdBy"`
	CreatedAt     time.Time          `json:"createdAt"`
	VoidReason    string             `json:"voidReason,omitempty"`
	VoidedBy      string             `json:"voidedBy,omitempty"`
	VoidedAt      *time.Time         `json:"voidedAt,omitempty"`
}

type RecycleItem struct {
	Category   string  `json:"category"`
	Purity     string  `json:"purity"`
	WeightGram float64 `json:"weightGram"`
}

type RecycleOrder struct {
	ID              string            `json:"id"`
	OrderNo         string            `json:"orderNo"`
	OrgID           string            `json:"orgId"`
	StoreID         string            `json:"storeId"`
	StoreName       string            `json:"storeName"`
	Status          string            `json:"status"`
	CustomerName    string            `json:"customerName"`
	CustomerPhone   string            `json:"customerPhone"`
	EstimatedAmount float64           `json:"estimatedAmount"`
	ConfirmedAmount float64           `json:"confirmedAmount"`
	Items           []RecycleItem     `json:"items"`
	AttachmentURLs  []string          `json:"attachmentUrls"`
	Attachments     []AttachmentAsset `json:"attachments,omitempty"`
	Remark          string            `json:"remark"`
	CreatedBy       string            `json:"createdBy"`
	CreatedAt       time.Time         `json:"createdAt"`
	ConfirmedAt     *time.Time        `json:"confirmedAt,omitempty"`
	CancelReason    string            `json:"cancelReason,omitempty"`
	CancelledBy     string            `json:"cancelledBy,omitempty"`
	CancelledAt     *time.Time        `json:"cancelledAt,omitempty"`
}

type AttachmentAsset struct {
	ID              string    `json:"id"`
	OrgID           string    `json:"orgId"`
	OrderID         string    `json:"orderId"`
	StoreID         string    `json:"storeId"`
	Category        string    `json:"category"`
	StorageProvider string    `json:"storageProvider"`
	ObjectKey       string    `json:"objectKey"`
	PublicURL       string    `json:"publicUrl"`
	ThumbnailURL    string    `json:"thumbnailUrl"`
	FileName        string    `json:"fileName"`
	ContentType     string    `json:"contentType"`
	SizeBytes       int64     `json:"sizeBytes"`
	Source          string    `json:"source"`
	Status          string    `json:"status"`
	UploadedBy      string    `json:"uploadedBy"`
	UploadedAt      time.Time `json:"uploadedAt"`
	HasPreview      bool      `json:"hasPreview"`
	PreviewURL      string    `json:"previewUrl"`
}

type DashboardSummary struct {
	OrgID                      string  `json:"orgId"`
	VisibleStoreCount          int     `json:"visibleStoreCount"`
	TodayCashierOrderCount     int     `json:"todayCashierOrderCount"`
	TodayCashierAmount         float64 `json:"todayCashierAmount"`
	DraftRecycleOrderCount     int     `json:"draftRecycleOrderCount"`
	ConfirmedRecycleOrderCount int     `json:"confirmedRecycleOrderCount"`
	TodayRecycleWeightGram     float64 `json:"todayRecycleWeightGram"`
}

type InventoryItem struct {
	ProductID            string   `json:"productId"`
	Name                 string   `json:"name"`
	SKU                  string   `json:"sku"`
	Category             string   `json:"category"`
	CategoryTab          string   `json:"categoryTab"`
	ImageURL             string   `json:"imageUrl"`
	Status               string   `json:"status"`
	StockStatus          string   `json:"stockStatus"`
	Inventory            int      `json:"inventory"`
	Stores               []string `json:"stores"`
	BenchPrice           float64  `json:"benchPrice"`
	RetailPrice          float64  `json:"retailPrice"`
	GramWeight           float64  `json:"gramWeight"`
	EstimatedRetailValue float64  `json:"estimatedRetailValue"`
}

type InventorySummary struct {
	OrgID                string          `json:"orgId"`
	VisibleStoreCount    int             `json:"visibleStoreCount"`
	TotalSKU             int             `json:"totalSku"`
	ActiveSKU            int             `json:"activeSku"`
	LowStockSKU          int             `json:"lowStockSku"`
	OutOfStockSKU        int             `json:"outOfStockSku"`
	TotalInventory       int             `json:"totalInventory"`
	EstimatedRetailValue float64         `json:"estimatedRetailValue"`
	Items                []InventoryItem `json:"items"`
}

type InventoryLedgerItem struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"orgId"`
	StoreID    string    `json:"storeId"`
	StoreName  string    `json:"storeName"`
	StyleNo    string    `json:"styleNo"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Purity     string    `json:"purity"`
	PieceCount int       `json:"pieceCount"`
	WeightGram float64   `json:"weightGram"`
	CostAmount float64   `json:"costAmount"`
	Status     string    `json:"status"`
	Source     string    `json:"source"`
	Remark     string    `json:"remark"`
	CreatedBy  string    `json:"createdBy"`
	CreatedAt  time.Time `json:"createdAt"`
}

type InventoryLedgerSummary struct {
	OrgID             string                `json:"orgId"`
	VisibleStoreCount int                   `json:"visibleStoreCount"`
	TotalStyleCount   int                   `json:"totalStyleCount"`
	TotalPieceCount   int                   `json:"totalPieceCount"`
	TotalWeightGram   float64               `json:"totalWeightGram"`
	TotalCostAmount   float64               `json:"totalCostAmount"`
	Items             []InventoryLedgerItem `json:"items"`
}

type MaterialLedgerItem struct {
	ID                  string     `json:"id"`
	OrgID               string     `json:"orgId"`
	StoreID             string     `json:"storeId"`
	StoreName           string     `json:"storeName"`
	Type                string     `json:"type"`
	OrderNo             string     `json:"orderNo"`
	CustomerName        string     `json:"customerName"`
	Category            string     `json:"category"`
	Purity              string     `json:"purity"`
	WeightGram          float64    `json:"weightGram"`
	Amount              float64    `json:"amount"`
	RemainingWeightGram float64    `json:"remainingWeightGram"`
	Status              string     `json:"status"`
	DueDate             string     `json:"dueDate"`
	Remark              string     `json:"remark"`
	Source              string     `json:"source"`
	CreatedBy           string     `json:"createdBy"`
	CreatedAt           time.Time  `json:"createdAt"`
	OutboundAt          *time.Time `json:"outboundAt,omitempty"`
	OutboundBy          string     `json:"outboundBy,omitempty"`
	OutboundRemark      string     `json:"outboundRemark,omitempty"`
}

type MaterialPurityStat struct {
	Purity     string  `json:"purity"`
	Count      int     `json:"count"`
	WeightGram float64 `json:"weightGram"`
	Amount     float64 `json:"amount"`
}

type MaterialLedgerSummary struct {
	OrgID               string               `json:"orgId"`
	VisibleStoreCount   int                  `json:"visibleStoreCount"`
	TodayWeightGram     float64              `json:"todayWeightGram"`
	TodayAmount         float64              `json:"todayAmount"`
	MonthWeightGram     float64              `json:"monthWeightGram"`
	MonthAmount         float64              `json:"monthAmount"`
	RemainingWeightGram float64              `json:"remainingWeightGram"`
	PledgeCount         int                  `json:"pledgeCount"`
	PledgeAmount        float64              `json:"pledgeAmount"`
	Items               []MaterialLedgerItem `json:"items"`
	PurityStats         []MaterialPurityStat `json:"purityStats"`
}

type GoldReferencePriceItem struct {
	Purity string  `json:"purity"`
	Price  float64 `json:"price"`
	Trend  string  `json:"trend"`
	Label  string  `json:"label"`
}

type GoldReferencePriceSnapshot struct {
	Source          string                   `json:"source"`
	SourceText      string                   `json:"sourceText"`
	BaseCNYPerGram  float64                  `json:"baseCnyPerGram"`
	XAUUSD          float64                  `json:"xauUsd,omitempty"`
	USDCNY          float64                  `json:"usdCny,omitempty"`
	UpdatedAt       string                   `json:"updatedAt"`
	ReferenceNote   string                   `json:"referenceNote"`
	ReferencePrices []GoldReferencePriceItem `json:"referencePrices"`
}

type DailyReportStoreMetric struct {
	StoreID           string  `json:"storeId"`
	StoreName         string  `json:"storeName"`
	CashierOrderCount int     `json:"cashierOrderCount"`
	CashierAmount     float64 `json:"cashierAmount"`
	RecycleOrderCount int     `json:"recycleOrderCount"`
	RecycleAmount     float64 `json:"recycleAmount"`
	RecycleWeightGram float64 `json:"recycleWeightGram"`
}

type DailyReportSummary struct {
	OrgID                      string                   `json:"orgId"`
	Date                       string                   `json:"date"`
	DataScope                  string                   `json:"dataScope"`
	VisibleStoreCount          int                      `json:"visibleStoreCount"`
	CashierOrderCount          int                      `json:"cashierOrderCount"`
	CashierAmount              float64                  `json:"cashierAmount"`
	RecycleDraftOrderCount     int                      `json:"recycleDraftOrderCount"`
	RecycleConfirmedOrderCount int                      `json:"recycleConfirmedOrderCount"`
	RecycleConfirmedAmount     float64                  `json:"recycleConfirmedAmount"`
	RecycleWeightGram          float64                  `json:"recycleWeightGram"`
	StoreMetrics               []DailyReportStoreMetric `json:"storeMetrics"`
}

type MemberProfile struct {
	ID                 string    `json:"id"`
	OrgID              string    `json:"orgId"`
	StoreID            string    `json:"storeId"`
	StoreName          string    `json:"storeName"`
	Name               string    `json:"name"`
	Phone              string    `json:"phone"`
	Level              string    `json:"level"`
	Status             string    `json:"status"`
	TotalOrders        int       `json:"totalOrders"`
	TotalRecycleAmount float64   `json:"totalRecycleAmount"`
	LastVisitAt        time.Time `json:"lastVisitAt"`
	PreferredPurity    string    `json:"preferredPurity"`
	SourceChannel      string    `json:"sourceChannel"`
	ManagerName        string    `json:"managerName"`
	IDVerified         bool      `json:"idVerified"`
	Tags               []string  `json:"tags"`
	Notes              string    `json:"notes"`
}

type CatalogProduct struct {
	ID               string   `json:"id"`
	OrgID            string   `json:"orgId"`
	Name             string   `json:"name"`
	SKU              string   `json:"sku"`
	Category         string   `json:"category"`
	CategoryTab      string   `json:"categoryTab"`
	ImageURL         string   `json:"imageUrl"`
	Purity           string   `json:"purity"`
	BenchPrice       float64  `json:"benchPrice"`
	RetailPrice      float64  `json:"retailPrice"`
	GramWeight       float64  `json:"gramWeight"`
	Status           string   `json:"status"`
	Inventory        int      `json:"inventory"`
	StockStatus      string   `json:"stockStatus"`
	StoreIDs         []string `json:"storeIds"`
	Stores           []string `json:"stores"`
	Tags             []string `json:"tags"`
	RecommendedScene string   `json:"recommendedScene"`
	QuoteLeadTime    string   `json:"quoteLeadTime"`
	// 顾客端款式扩展（后台「款式/工费管理」）
	LaborFeeRef            string   `json:"laborFeeRef,omitempty"`            // 工费参考，如 "35元/克 起"
	Description            string   `json:"description,omitempty"`            // 款式说明
	LaborFeeNote           string   `json:"laborFeeNote,omitempty"`           // 工费说明
	Images                 []string `json:"images,omitempty"`                 // 款式多图（首图=主图）
	DetailImages           []string `json:"detailImages,omitempty"`           // 款式详情图（style_detail 场景上传）
	ApplicableServiceTypes []string `json:"applicableServiceTypes,omitempty"` // 适用服务类型
	RecommendedStoreRule   string   `json:"recommendedStoreRule,omitempty"`   // nearest / product_stores / all
	SortOrder              int      `json:"sortOrder,omitempty"`
	IsRecommended          bool     `json:"isRecommended,omitempty"`
	IsHot                  bool     `json:"isHot,omitempty"`
}

type BrandSettings struct {
	Name            string `json:"name"`
	ServicePhone    string `json:"servicePhone"`
	ReceiptTitle    string `json:"receiptTitle"`
	SupportMiniApp  bool   `json:"supportMiniApp"`
	DefaultCurrency string `json:"defaultCurrency"`
}

type RecycleSettings struct {
	MinPhotoCount     int  `json:"minPhotoCount"`
	MaxPhotoCount     int  `json:"maxPhotoCount"`
	RequireExactThree bool `json:"requireExactThree"`
	RequireIDCheck    bool `json:"requireIdCheck"`
}

type StorageSettings struct {
	Enabled           bool   `json:"enabled"`
	Provider          string `json:"provider"`
	Bucket            string `json:"bucket"`
	Region            string `json:"region"`
	Endpoint          string `json:"endpoint,omitempty"`
	PublicBaseURL     string `json:"publicBaseUrl"`
	PathPrefix        string `json:"pathPrefix"`
	UploadStrategy    string `json:"uploadStrategy"`
	AccessKeyID       string `json:"-"`
	AccessKeySecret   string `json:"-"`
	UploadURLTTL      int    `json:"uploadUrlTtlSeconds,omitempty"`
	CallbackEnabled   bool   `json:"callbackEnabled"`
	StatusDescription string `json:"statusDescription"`
}

type SystemSettings struct {
	OrgID        string          `json:"orgId"`
	Brand        BrandSettings   `json:"brand"`
	Recycle      RecycleSettings `json:"recycle"`
	Storage      StorageSettings `json:"storage"`
	FeatureFlags map[string]bool `json:"featureFlags"`
	UpdatedBy    string          `json:"updatedBy"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

type UploadPreparation struct {
	UploadID      string            `json:"uploadId"`
	StoreID       string            `json:"storeId"`
	Category      string            `json:"category"`
	Provider      string            `json:"provider"`
	Bucket        string            `json:"bucket"`
	Region        string            `json:"region"`
	ObjectKey     string            `json:"objectKey"`
	FileName      string            `json:"fileName"`
	ContentType   string            `json:"contentType"`
	SizeBytes     int64             `json:"sizeBytes"`
	UploadURL     string            `json:"uploadUrl"`
	PublicURL     string            `json:"publicUrl"`
	Headers       map[string]string `json:"headers"`
	FormFields    map[string]string `json:"formFields"`
	StorageReady  bool              `json:"storageReady"`
	UploadMode    string            `json:"uploadMode"`
	ExpiresAt     time.Time         `json:"expiresAt"`
	ReferenceNote string            `json:"referenceNote"`
}
