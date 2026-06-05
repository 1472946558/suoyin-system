package app

type AdminPageID string

const (
	AdminPageDashboard    AdminPageID = "dashboard"
	AdminPageStores       AdminPageID = "stores"
	AdminPageUsers        AdminPageID = "users"
	AdminPageRoles        AdminPageID = "roles"
	AdminPageProducts     AdminPageID = "products"
	AdminPageOrders       AdminPageID = "orders"
	AdminPageRecycle      AdminPageID = "recycle"
	AdminPageSystemConfig AdminPageID = "system-config"
	AdminPageTemplateInit AdminPageID = "template-init"
	AdminPageAudit        AdminPageID = "audit"
)

type AdminAbilityCode string

type AdminSessionUser struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Account     string             `json:"account"`
	RoleKey     string             `json:"roleKey"`
	RoleName    string             `json:"roleName"`
	DataScope   string             `json:"dataScope"`
	StoreIDs    []string           `json:"storeIds"`
	Abilities   []AdminAbilityCode `json:"abilities"`
	LastLoginAt string             `json:"lastLoginAt"`
}

type AdminLoginResult struct {
	Token       string           `json:"token"`
	User        AdminSessionUser `json:"user"`
	LandingPage AdminPageID      `json:"landingPage"`
}

type AdminDashboardMetric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`
	Tone  string `json:"tone"`
}

type AdminDashboardTodo struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Level       string      `json:"level"`
	Page        AdminPageID `json:"page"`
}

type AdminDashboardShortcut struct {
	ID      string           `json:"id"`
	Label   string           `json:"label"`
	Hint    string           `json:"hint"`
	Page    AdminPageID      `json:"page"`
	Ability AdminAbilityCode `json:"ability"`
}

type AdminDashboardSummary struct {
	Title     string                   `json:"title"`
	Subtitle  string                   `json:"subtitle"`
	Metrics   []AdminDashboardMetric   `json:"metrics"`
	Todos     []AdminDashboardTodo     `json:"todos"`
	Shortcuts []AdminDashboardShortcut `json:"shortcuts"`
	Notices   []string                 `json:"notices"`
}

type AdminStoreRecord struct {
	ID               string   `json:"id"`
	Code             string   `json:"code"`
	Name             string   `json:"name"`
	ManagerName      string   `json:"managerName"`
	City             string   `json:"city"`
	Address          string   `json:"address"`
	ContactPhone     string   `json:"contactPhone"`
	BusinessHours    string   `json:"businessHours"`
	Status           string   `json:"status"`
	CashierDevices   int      `json:"cashierDevices"`
	PendingTasks     int      `json:"pendingTasks"`
	TodayAmount      float64  `json:"todayAmount"`
	TodayOrders      int      `json:"todayOrders"`
	LastSettlementAt string   `json:"lastSettlementAt"`
	Tags             []string `json:"tags"`
}

type AdminUserAccount struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Account       string             `json:"account"`
	Phone         string             `json:"phone"`
	MiniAppOpenID string             `json:"miniAppOpenId,omitempty"`
	RoleKey       string             `json:"roleKey"`
	RoleName      string             `json:"roleName"`
	DataScope     string             `json:"dataScope"`
	StoreIDs      []string           `json:"storeIds"`
	StoreNames    []string           `json:"storeNames"`
	Status        string             `json:"status"`
	LastLoginAt   string             `json:"lastLoginAt"`
	Abilities     []AdminAbilityCode `json:"abilities"`
}

type PendingMiniAppBinding struct {
	ID            string `json:"id"`
	MiniAppOpenID string `json:"miniAppOpenId"`
	UnionID       string `json:"unionId,omitempty"`
	OperatorName  string `json:"operatorName"`
	Phone         string `json:"phone"`
	RoleKey       string `json:"roleKey"`
	RoleName      string `json:"roleName"`
	StoreName     string `json:"storeName"`
	StoreCode     string `json:"storeCode"`
	CreatedAt     string `json:"createdAt"`
}

type AdminAbilityOption struct {
	Code        AdminAbilityCode `json:"code"`
	Label       string           `json:"label"`
	Description string           `json:"description"`
}

type AdminAbilityGroup struct {
	Key         string               `json:"key"`
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Items       []AdminAbilityOption `json:"items"`
}

type AdminRoleTemplate struct {
	ID          int                `json:"id"`
	Key         string             `json:"key"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	DataScope   string             `json:"dataScope"`
	MemberCount int                `json:"memberCount"`
	Locked      bool               `json:"locked"`
	Abilities   []AdminAbilityCode `json:"abilities"`
}

type AdminProductRecord struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	SKU         string   `json:"sku"`
	Category    string   `json:"category"`
	CategoryTab string   `json:"categoryTab"`
	ImageURL    string   `json:"imageUrl"`
	Price       float64  `json:"price"`
	GramWeight  float64  `json:"gramWeight"`
	Status      string   `json:"status"`
	Inventory   int      `json:"inventory"`
	StockStatus string   `json:"stockStatus"`
	StoreNames  []string `json:"storeNames"`
	Tags        []string `json:"tags"`
}

type AdminCashierOrderView struct {
	ID            string  `json:"id"`
	OrderNo       string  `json:"orderNo"`
	StoreID       string  `json:"storeId"`
	StoreName     string  `json:"storeName"`
	Status        string  `json:"status"`
	PaymentMethod string  `json:"paymentMethod"`
	TotalAmount   float64 `json:"totalAmount"`
	ItemCount     int     `json:"itemCount"`
	CreatedBy     string  `json:"createdBy"`
	CreatedAt     string  `json:"createdAt"`
	Remark        string  `json:"remark"`
}

type AdminRecycleOrderView struct {
	ID              string  `json:"id"`
	OrderNo         string  `json:"orderNo"`
	StoreID         string  `json:"storeId"`
	StoreName       string  `json:"storeName"`
	Status          string  `json:"status"`
	CustomerName    string  `json:"customerName"`
	CustomerPhone   string  `json:"customerPhone"`
	EstimatedAmount float64 `json:"estimatedAmount"`
	ConfirmedAmount float64 `json:"confirmedAmount"`
	PhotoCount      int     `json:"photoCount"`
	CreatedBy       string  `json:"createdBy"`
	CreatedAt       string  `json:"createdAt"`
	ConfirmedAt     string  `json:"confirmedAt,omitempty"`
	Remark          string  `json:"remark"`
}

type AdminSystemProfile struct {
	BrandName         string `json:"brandName"`
	ServicePhone      string `json:"servicePhone"`
	ReceiptTitle      string `json:"receiptTitle"`
	MinPhotoCount     int    `json:"minPhotoCount"`
	MaxPhotoCount     int    `json:"maxPhotoCount"`
	RequireExactThree bool   `json:"requireExactThree"`
	RequireIDCheck    bool   `json:"requireIdCheck"`
	DomainName        string `json:"domainName"`
	DomainStatus      string `json:"domainStatus"`
	OSSStatus         string `json:"ossStatus"`
	AppIDStatus       string `json:"appIdStatus"`
	PrinterStatus     string `json:"printerStatus"`
}

type AdminAuditLogRecord struct {
	ID           string `json:"id"`
	Module       string `json:"module"`
	Action       string `json:"action"`
	OperatorName string `json:"operatorName"`
	Result       string `json:"result"`
	RiskLevel    string `json:"riskLevel"`
	Summary      string `json:"summary"`
	CreatedAt    string `json:"createdAt"`
}

type AdminTemplateInitPlan struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Steps       []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	} `json:"steps"`
	Outputs []string `json:"outputs"`
}

type AdminPrintTemplate struct {
	Receipt struct {
		Enabled             bool     `json:"enabled"`
		PaperWidth          string   `json:"paperWidth"`
		HeaderTitle         string   `json:"headerTitle"`
		FooterNote          string   `json:"footerNote"`
		ShowStoreName       bool     `json:"showStoreName"`
		ShowOperatorName    bool     `json:"showOperatorName"`
		ShowPhotoSummary    bool     `json:"showPhotoSummary"`
		ShowPhotoThumbnails bool     `json:"showPhotoThumbnails"`
		Fields              []string `json:"fields"`
	} `json:"receipt"`
	Label struct {
		Enabled     bool     `json:"enabled"`
		Size        string   `json:"size"`
		Copies      int      `json:"copies"`
		Fields      []string `json:"fields"`
		BarcodeType string   `json:"barcodeType"`
	} `json:"label"`
	Recycle struct {
		Enabled             bool   `json:"enabled"`
		PrintPhotoSummary   bool   `json:"printPhotoSummary"`
		PrintPhotoThumbnail bool   `json:"printPhotoThumbnail"`
		SummaryText         string `json:"summaryText"`
	} `json:"recycle"`
}

type AdminBootstrap struct {
	CurrentUser   AdminSessionUser        `json:"currentUser"`
	Dashboard     AdminDashboardSummary   `json:"dashboard"`
	Stores        []AdminStoreRecord      `json:"stores"`
	Users         []AdminUserAccount      `json:"users"`
	Roles         []AdminRoleTemplate     `json:"roles"`
	AbilityGroups []AdminAbilityGroup     `json:"abilityGroups"`
	Products      []AdminProductRecord    `json:"products"`
	CashierOrders []AdminCashierOrderView `json:"cashierOrders"`
	RecycleOrders []AdminRecycleOrderView `json:"recycleOrders"`
	PrintTemplate AdminPrintTemplate      `json:"printTemplate"`
	SystemProfile AdminSystemProfile      `json:"systemProfile"`
	AuditLogs     []AdminAuditLogRecord   `json:"auditLogs"`
	TemplateInit  AdminTemplateInitPlan   `json:"templateInit"`
	UpdatedAt     string                  `json:"updatedAt"`
}
