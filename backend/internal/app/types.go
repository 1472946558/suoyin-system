package app

import "time"

type Config struct {
	Host        string
	Port        string
	Mode        string
	TokenSecret string
	CORSOrigin  string
	MySQLDSN    string
	RedisAddr   string
	RedisUser   string
	RedisPass   string
	RedisDB     int
	RedisPrefix string
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
	IsDefault bool   `json:"isDefault"`
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
	ID          string   `json:"id"`
	OrgID       string   `json:"orgId"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Password    string   `json:"password,omitempty"`
	RoleCode    string   `json:"roleCode"`
	RoleName    string   `json:"roleName"`
	DataScope   string   `json:"dataScope"`
	StoreIDs    []string `json:"storeIds"`
	Permissions []string `json:"permissions"`
	Status      string   `json:"status"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	LoginAt   time.Time `json:"loginAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type CashierOrderLine struct {
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
	PaymentMethod string             `json:"paymentMethod"`
	TotalAmount   float64            `json:"totalAmount"`
	PaidAmount    float64            `json:"paidAmount"`
	Remark        string             `json:"remark"`
	Items         []CashierOrderLine `json:"items"`
	CreatedBy     string             `json:"createdBy"`
	CreatedAt     time.Time          `json:"createdAt"`
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
}

type PaymentTransaction struct {
	ID              string     `json:"id"`
	PaymentNo       string     `json:"paymentNo"`
	OrderNo         string     `json:"orderNo"`
	BizType         string     `json:"bizType"`
	OrgID           string     `json:"orgId"`
	StoreID         string     `json:"storeId"`
	StoreName       string     `json:"storeName"`
	Amount          float64    `json:"amount"`
	Method          string     `json:"method"`
	Status          string     `json:"status"`
	CallbackStatus  string     `json:"callbackStatus"`
	ProviderRef     string     `json:"providerRef"`
	ProviderPayload string     `json:"providerPayload"`
	OperatorName    string     `json:"operatorName"`
	CustomerLabel   string     `json:"customerLabel"`
	Remark          string     `json:"remark"`
	Anomaly         bool       `json:"anomaly"`
	PaidAt          *time.Time `json:"paidAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
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
	Purity           string   `json:"purity"`
	BenchPrice       float64  `json:"benchPrice"`
	RetailPrice      float64  `json:"retailPrice"`
	GramWeight       float64  `json:"gramWeight"`
	Status           string   `json:"status"`
	Inventory        int      `json:"inventory"`
	StoreIDs         []string `json:"storeIds"`
	Stores           []string `json:"stores"`
	Tags             []string `json:"tags"`
	RecommendedScene string   `json:"recommendedScene"`
	QuoteLeadTime    string   `json:"quoteLeadTime"`
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

type PaymentSettings struct {
	EnableCash         bool `json:"enableCash"`
	EnableWechatPay    bool `json:"enableWechatPay"`
	EnableBankTransfer bool `json:"enableBankTransfer"`
}

type StorageSettings struct {
	Enabled           bool   `json:"enabled"`
	Provider          string `json:"provider"`
	Bucket            string `json:"bucket"`
	Region            string `json:"region"`
	PublicBaseURL     string `json:"publicBaseUrl"`
	PathPrefix        string `json:"pathPrefix"`
	UploadStrategy    string `json:"uploadStrategy"`
	CallbackEnabled   bool   `json:"callbackEnabled"`
	StatusDescription string `json:"statusDescription"`
}

type SystemSettings struct {
	OrgID        string          `json:"orgId"`
	Brand        BrandSettings   `json:"brand"`
	Recycle      RecycleSettings `json:"recycle"`
	Payments     PaymentSettings `json:"payments"`
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
