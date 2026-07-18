/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: admin_types.go
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-05-10
 */

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
	Password      string             `json:"password,omitempty"`
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

type AdminProductImportFailure struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

type AdminProductImportResult struct {
	ImportID     string                      `json:"importId"`
	StoreID      string                      `json:"storeId"`
	StoreName    string                      `json:"storeName"`
	SuccessCount int                         `json:"successCount"`
	FailureCount int                         `json:"failureCount"`
	Failures     []AdminProductImportFailure `json:"failures"`
	ImportedAt   string                      `json:"importedAt"`
}

type AdminProductImportLog struct {
	ID           string                      `json:"id"`
	StoreID      string                      `json:"storeId"`
	StoreName    string                      `json:"storeName"`
	FileName     string                      `json:"fileName"`
	SuccessCount int                         `json:"successCount"`
	FailureCount int                         `json:"failureCount"`
	Failures     []AdminProductImportFailure `json:"failures"`
	ImportedBy   string                      `json:"importedBy"`
	ImportedAt   string                      `json:"importedAt"`
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
	CostPrice   float64  `json:"costPrice"`
	Price       float64  `json:"price"`
	GramWeight  float64  `json:"gramWeight"`
	Status      string   `json:"status"`
	Inventory   int      `json:"inventory"`
	StockStatus string   `json:"stockStatus"`
	StoreIDs    []string `json:"storeIds"`
	StoreNames  []string `json:"storeNames"`
	Tags        []string `json:"tags"`
}

type AdminCashierOrderView struct {
	ID            string             `json:"id"`
	OrderNo       string             `json:"orderNo"`
	StoreID       string             `json:"storeId"`
	StoreName     string             `json:"storeName"`
	Status        string             `json:"status"`
	CustomerName  string             `json:"customerName"`
	CustomerPhone string             `json:"customerPhone"`
	PaymentMethod string             `json:"paymentMethod"`
	TotalAmount   float64            `json:"totalAmount"`
	PaidAmount    float64            `json:"paidAmount"`
	ItemCount     int                `json:"itemCount"`
	ItemSummary   string             `json:"itemSummary"`
	Items         []CashierOrderLine `json:"items,omitempty"`
	CreatedBy     string             `json:"createdBy"`
	CreatedAt     string             `json:"createdAt"`
	Remark        string             `json:"remark"`
	VoidReason    string             `json:"voidReason,omitempty"`
	VoidedBy      string             `json:"voidedBy,omitempty"`
	VoidedAt      string             `json:"voidedAt,omitempty"`
	RefundReason  string             `json:"refundReason,omitempty"`
	RefundedBy    string             `json:"refundedBy,omitempty"`
	RefundedAt    string             `json:"refundedAt,omitempty"`
}

type AdminRecycleOrderView struct {
	ID              string            `json:"id"`
	OrderNo         string            `json:"orderNo"`
	StoreID         string            `json:"storeId"`
	StoreName       string            `json:"storeName"`
	Status          string            `json:"status"`
	CustomerName    string            `json:"customerName"`
	CustomerPhone   string            `json:"customerPhone"`
	EstimatedAmount float64           `json:"estimatedAmount"`
	ConfirmedAmount float64           `json:"confirmedAmount"`
	PhotoCount      int               `json:"photoCount"`
	ItemSummary     string            `json:"itemSummary"`
	Items           []RecycleItem     `json:"items,omitempty"`
	Attachments     []AttachmentAsset `json:"attachments,omitempty"`
	CreatedBy       string            `json:"createdBy"`
	CreatedAt       string            `json:"createdAt"`
	ConfirmedAt     string            `json:"confirmedAt,omitempty"`
	Remark          string            `json:"remark"`
	CancelReason    string            `json:"cancelReason,omitempty"`
	CancelledBy     string            `json:"cancelledBy,omitempty"`
	CancelledAt     string            `json:"cancelledAt,omitempty"`
}

type AdminListQuery struct {
	Page     int
	PageSize int
	StoreID  string
	Status   string
	DateFrom string
	DateTo   string
	Keyword  string
}

type AdminPagedItems[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type AdminBootstrapSummary struct {
	CurrentUser   AdminSessionUser      `json:"currentUser"`
	Dashboard     AdminDashboardSummary `json:"dashboard"`
	Stores        []AdminStoreRecord    `json:"stores"`
	Roles         []AdminRoleTemplate   `json:"roles"`
	SystemProfile AdminSystemProfile    `json:"systemProfile"`
	UpdatedAt     string                `json:"updatedAt"`
}

type AdminOrderRow struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	OrderNo       string  `json:"orderNo"`
	StoreID       string  `json:"storeId"`
	StoreName     string  `json:"storeName"`
	Status        string  `json:"status"`
	CustomerName  string  `json:"customerName"`
	CustomerPhone string  `json:"customerPhone"`
	Amount        float64 `json:"amount"`
	ItemSummary   string  `json:"itemSummary"`
	CreatedBy     string  `json:"createdBy"`
	CreatedAt     string  `json:"createdAt"`
}

type AdminActionReasonRequest struct {
	Reason string `json:"reason"`
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
	Members       []MemberProfile         `json:"members"`
	CashierOrders []AdminCashierOrderView `json:"cashierOrders"`
	RecycleOrders []AdminRecycleOrderView `json:"recycleOrders"`
	PrintTemplate AdminPrintTemplate      `json:"printTemplate"`
	SystemProfile AdminSystemProfile      `json:"systemProfile"`
	AuditLogs     []AdminAuditLogRecord   `json:"auditLogs"`
	ImportLogs    []AdminProductImportLog `json:"importLogs"`
	TemplateInit  AdminTemplateInitPlan   `json:"templateInit"`
	UpdatedAt     string                  `json:"updatedAt"`
}
