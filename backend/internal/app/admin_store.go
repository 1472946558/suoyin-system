package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	errAdminRoleNotFound    = errors.New("admin role not found")
	errAdminStoreNotFound   = errors.New("admin store not found")
	errAdminUserNotFound    = errors.New("admin user not found")
	errAdminProductNotFound = errors.New("admin product not found")
)

func newAdminAbilityGroups() []AdminAbilityGroup {
	return []AdminAbilityGroup{
		{
			Key:         "dashboard",
			Label:       "首页",
			Description: "后台首页概览与快捷入口能力。",
			Items: []AdminAbilityOption{
				{Code: "dashboard.view", Label: "查看首页", Description: "查看经营概览、待办和快捷入口。"},
			},
		},
		{
			Key:         "organization",
			Label:       "组织与权限",
			Description: "门店、账号和角色模板相关能力。",
			Items: []AdminAbilityOption{
				{Code: "store.view", Label: "查看门店", Description: "查看门店资料、状态和营业信息。"},
				{Code: "store.manage", Label: "管理门店", Description: "编辑门店资料、负责人和启停用入口。"},
				{Code: "user.view", Label: "查看账号", Description: "查看后台账号、角色绑定和登录状态。"},
				{Code: "user.manage", Label: "管理账号", Description: "启停用账号、重置密码和绑定门店。"},
				{Code: "role.view", Label: "查看角色权限", Description: "查看固定角色模板和数据范围。"},
				{Code: "role.manage", Label: "管理角色权限", Description: "调整能力项和数据范围配置。"},
			},
		},
		{
			Key:         "business",
			Label:       "商品与业务",
			Description: "商品、订单和回收主线。",
			Items: []AdminAbilityOption{
				{Code: "product.view", Label: "查看商品", Description: "查看商品分类、SKU 和状态。"},
				{Code: "product.manage", Label: "管理商品", Description: "维护分类、价格和上架状态。"},
				{Code: "order.view", Label: "查看订单", Description: "查看收银订单与支付状态。"},
				{Code: "recycle.view", Label: "查看回收单", Description: "查看回收单、图片留痕和支付关联。"},
			},
		},
		{
			Key:         "payment",
			Label:       "支付中心",
			Description: "支付配置、回调状态和流水查询能力。",
			Items: []AdminAbilityOption{
				{Code: "payment.config.view", Label: "查看支付配置", Description: "查看微信支付、现金记账和回调设置。"},
				{Code: "payment.config.manage", Label: "管理支付配置", Description: "修改支付开关和回调重试规则。"},
				{Code: "payment.record.view", Label: "查看支付流水", Description: "按门店、支付方式和状态查询流水。"},
			},
		},
		{
			Key:         "system",
			Label:       "系统与模板",
			Description: "全局配置、模板化初始化和审计留痕。",
			Items: []AdminAbilityOption{
				{Code: "system.config.view", Label: "查看系统配置", Description: "查看拍照规则、编号规则和品牌信息。"},
				{Code: "system.config.manage", Label: "管理系统配置", Description: "修改全局业务规则和基础配置。"},
				{Code: "template.init", Label: "模板初始化", Description: "执行新客户初始化向导。"},
				{Code: "audit.view", Label: "查看操作日志", Description: "查看权限变更、支付修改等审计记录。"},
			},
		},
	}
}

func allAdminAbilities(groups []AdminAbilityGroup) []AdminAbilityCode {
	var abilities []AdminAbilityCode
	for _, group := range groups {
		for _, item := range group.Items {
			abilities = append(abilities, item.Code)
		}
	}
	return abilities
}

func newAdminPaymentConfig() AdminPaymentConfig {
	var cfg AdminPaymentConfig
	cfg.Wechat.Enabled = true
	cfg.Wechat.AppID = "待补充 AppID"
	cfg.Wechat.MerchantID = "待补充商户号"
	cfg.Wechat.SubMerchantID = ""
	cfg.Wechat.CertificateStatus = "已预留接入位，待真实证书联调"
	cfg.Wechat.CallbackURL = "https://jinjiangguan.com/api/payment/wechat/callback"
	cfg.Wechat.SandboxMode = true
	cfg.Wechat.LastVerifiedAt = "2026-05-10 15:30"
	cfg.Cash.Enabled = true
	cfg.Cash.ReceiptRequired = true
	cfg.Cash.ShiftReconciliationRequired = true
	cfg.BankTransfer.Enabled = false
	cfg.BankTransfer.AccountName = "待确认主体账户"
	cfg.BankTransfer.AccountSuffix = "1288"
	cfg.Reconciliation.AutoRetryEnabled = true
	cfg.Reconciliation.RetryMinutes = 15
	cfg.Reconciliation.AbnormalNotify = "支付异常企业群 / ops@gold.local"
	return cfg
}

func newAdminPrintTemplate() AdminPrintTemplate {
	var tmpl AdminPrintTemplate
	tmpl.Receipt.Enabled = true
	tmpl.Receipt.PaperWidth = "80mm"
	tmpl.Receipt.HeaderTitle = "黄金回收收银系统"
	tmpl.Receipt.FooterNote = "请当面核对金额与留痕信息"
	tmpl.Receipt.ShowStoreName = true
	tmpl.Receipt.ShowOperatorName = true
	tmpl.Receipt.ShowPaymentMethod = true
	tmpl.Receipt.ShowPhotoSummary = true
	tmpl.Receipt.ShowPhotoThumbnails = false
	tmpl.Receipt.Fields = []string{"订单号", "客户信息", "商品信息", "金额", "支付方式", "照片已留存"}
	tmpl.Label.Enabled = true
	tmpl.Label.Size = "50x30mm"
	tmpl.Label.Copies = 1
	tmpl.Label.Fields = []string{"标签标题", "门店", "SKU/单号", "金额/克重"}
	tmpl.Label.BarcodeType = "CODE128"
	tmpl.Recycle.Enabled = true
	tmpl.Recycle.PrintPhotoSummary = true
	tmpl.Recycle.PrintPhotoThumbnail = false
	tmpl.Recycle.SummaryText = "已留存现场照片 {photoCount} 张"
	return tmpl
}

func newAdminTemplatePlan() AdminTemplateInitPlan {
	plan := AdminTemplateInitPlan{
		Title:       "模板初始化入口",
		Description: "用于新客户开通时一次性落默认门店、角色模板、支付开关和品牌基础配置。",
		Outputs: []string{
			"默认门店模板 1 套",
			"固定角色模板 3 套",
			"支付方式开关预设 1 组",
			"品牌基础配置包 1 份",
		},
	}
	plan.Steps = []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}{
		{ID: "step-1", Title: "品牌基础信息", Description: "品牌名、客服电话、小票抬头和 Logo 资料。", Status: "done"},
		{ID: "step-2", Title: "默认门店模板", Description: "创建默认门店、营业时间、负责人和设备位。", Status: "current"},
		{ID: "step-3", Title: "固定角色模板", Description: "下发老板、店长、员工模板与初始能力组。", Status: "planned"},
		{ID: "step-4", Title: "支付与系统规则", Description: "回调地址、现金记账、拍照规则、编号规则。", Status: "planned"},
	}
	return plan
}

func (s *MockStore) adminStoreIDsForUser(user UserAccount) []string {
	if user.DataScope == "org_all" {
		ids := make([]string, 0, len(s.adminStores))
		for _, item := range s.adminStores {
			ids = append(ids, item.ID)
		}
		return ids
	}
	return append([]string(nil), user.StoreIDs...)
}

func containsID(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func (s *MockStore) adminSessionUser(user UserAccount) AdminSessionUser {
	dataScope := "assigned_store"
	if user.DataScope == "org_all" {
		dataScope = "all_stores"
	}
	roleKey := "clerk"
	if user.RoleCode == "owner" {
		roleKey = "owner"
	} else if user.RoleCode == "manager" {
		roleKey = "manager"
	}
	abilities := []AdminAbilityCode{}
	for _, role := range s.adminRoles {
		if role.Key == roleKey {
			abilities = append(abilities, role.Abilities...)
			break
		}
	}
	return AdminSessionUser{
		ID:          user.ID,
		Name:        user.DisplayName,
		Account:     user.Username,
		RoleKey:     roleKey,
		RoleName:    user.RoleName,
		DataScope:   dataScope,
		StoreIDs:    append([]string(nil), user.StoreIDs...),
		Abilities:   abilities,
		LastLoginAt: time.Now().Format("2006-01-02 15:04"),
	}
}

func (s *MockStore) buildAdminBootstrap(user UserAccount) AdminBootstrap {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionUser := s.adminSessionUser(user)
	visibleStoreIDs := s.adminStoreIDsForUser(user)
	stores := s.buildDynamicAdminStoresLocked()
	scopedStores := make([]AdminStoreRecord, 0, len(stores))
	for _, item := range stores {
		if sessionUser.DataScope == "all_stores" || containsID(visibleStoreIDs, item.ID) {
			scopedStores = append(scopedStores, item)
		}
	}

	users := s.buildDynamicAdminUsersLocked()
	scopedUsers := make([]AdminUserAccount, 0, len(users))
	for _, item := range users {
		if sessionUser.DataScope == "all_stores" || item.ID == sessionUser.ID || hasIntersect(item.StoreIDs, visibleStoreIDs) {
			scopedUsers = append(scopedUsers, item)
		}
	}

	records := s.buildDynamicPaymentRecordsLocked()
	scopedRecords := make([]AdminPaymentRecord, 0, len(records))
	for _, item := range records {
		if sessionUser.DataScope == "all_stores" || containsID(visibleStoreIDs, item.StoreID) {
			scopedRecords = append(scopedRecords, item)
		}
	}

	cashierOrders := s.buildDynamicCashierOrderViewsLocked()
	scopedCashierOrders := make([]AdminCashierOrderView, 0, len(cashierOrders))
	for _, item := range cashierOrders {
		if sessionUser.DataScope == "all_stores" || containsID(visibleStoreIDs, item.StoreID) {
			scopedCashierOrders = append(scopedCashierOrders, item)
		}
	}

	recycleOrders := s.buildDynamicRecycleOrderViewsLocked()
	scopedRecycleOrders := make([]AdminRecycleOrderView, 0, len(recycleOrders))
	for _, item := range recycleOrders {
		if sessionUser.DataScope == "all_stores" || containsID(visibleStoreIDs, item.StoreID) {
			scopedRecycleOrders = append(scopedRecycleOrders, item)
		}
	}

	products := s.buildDynamicAdminProductsLocked()
	scopedProducts := make([]AdminProductRecord, 0, len(products))
	visibleStoreNames := make([]string, 0, len(scopedStores))
	for _, store := range scopedStores {
		visibleStoreNames = append(visibleStoreNames, store.Name)
	}
	for _, item := range products {
		if sessionUser.DataScope == "all_stores" || productVisibleInStores(item, visibleStoreNames) {
			scopedProducts = append(scopedProducts, item)
		}
	}

	title := "门店运营台"
	subtitle := "已按本门店视角收敛数据与可见菜单，可直接作为店长受限后台演示。"
	if sessionUser.RoleKey == "owner" {
		title = "老板主控台"
		subtitle = "覆盖门店、账号、角色权限、商品、订单、支付中心和模板初始化的正式后台首版。"
	}

	totalAmount := 0.0
	pendingCount := 0
	for _, record := range scopedRecords {
		if record.Status != "failed" {
			totalAmount += record.Amount
		}
		if record.CallbackStatus != "delivered" || record.Status == "refunding" {
			pendingCount++
		}
	}

	todos := []AdminDashboardTodo{
		{ID: "todo-print", Title: "打印设备型号待确认", Description: "需确定小票机和标签机型号，后续做模板适配。", Level: "high", Page: AdminPageSystemConfig},
		{ID: "todo-domain", Title: "域名实名认证审核中", Description: "审核通过后再做正式解析和 HTTPS 配置。", Level: "medium", Page: AdminPageSystemConfig},
	}
	if sessionUser.RoleKey == "owner" {
		todos = append([]AdminDashboardTodo{
			{ID: "todo-role", Title: "固定角色模板待最终冻结", Description: "建议确认店长是否保留商品维护与回收查看权限。", Level: "medium", Page: AdminPageRoles},
		}, todos...)
	}

	notices := []string{
		"当前后台已经优先连接真实 API，接口不可用时才回退到 mock 数据。",
		"客户要求优先，ERPNext 仅用于补齐权限、打印、配置化的结构思路。",
	}

	return AdminBootstrap{
		CurrentUser: sessionUser,
		Dashboard: AdminDashboardSummary{
			Title:    title,
			Subtitle: subtitle,
			Metrics: []AdminDashboardMetric{
				{Key: "gross", Label: "近期开单金额", Value: fmt.Sprintf("¥%.0f", totalAmount), Delta: "按当前可见门店聚合", Tone: "gold"},
				{Key: "orders", Label: "收银订单", Value: fmt.Sprintf("%d", len(scopedCashierOrders)), Delta: "当前后台已接收订单视图", Tone: "emerald"},
				{Key: "recycle", Label: "回收单", Value: fmt.Sprintf("%d", len(scopedRecycleOrders)), Delta: "拍照留痕规则已纳入", Tone: "slate"},
				{Key: "pending", Label: "待处理事项", Value: fmt.Sprintf("%d", pendingCount), Delta: "支付回调 / 云资源 / 打印适配", Tone: "danger"},
			},
			Todos: todos,
			Shortcuts: []AdminDashboardShortcut{
				{ID: "shortcut-store", Label: "门店巡检", Hint: "检查门店状态、负责人和设备位", Page: AdminPageStores, Ability: "store.view"},
				{ID: "shortcut-payment", Label: "支付异常复核", Hint: "核对回调状态和异常流水", Page: AdminPagePaymentRecord, Ability: "payment.record.view"},
				{ID: "shortcut-config", Label: "系统配置", Hint: "查看云资源、拍照规则和打印准备情况", Page: AdminPageSystemConfig, Ability: "system.config.view"},
			},
			Notices: notices,
		},
		Stores:         scopedStores,
		Users:          scopedUsers,
		Roles:          append([]AdminRoleTemplate(nil), s.adminRoles...),
		AbilityGroups:  append([]AdminAbilityGroup(nil), s.adminAbilityGroups...),
		Products:       scopedProducts,
		CashierOrders:  scopedCashierOrders,
		RecycleOrders:  scopedRecycleOrders,
		PaymentConfig:  s.adminPaymentConfig,
		PrintTemplate:  s.adminPrintTemplate,
		PaymentRecords: scopedRecords,
		SystemProfile:  s.buildDynamicSystemProfileLocked(),
		AuditLogs:      append([]AdminAuditLogRecord(nil), s.adminAuditLogs...),
		TemplateInit:   s.adminTemplateInit,
		UpdatedAt:      time.Now().Format("2006-01-02 15:04"),
	}
}

func hasIntersect(left, right []string) bool {
	for _, a := range left {
		for _, b := range right {
			if a == b {
				return true
			}
		}
	}
	return false
}

func (s *MockStore) buildDynamicAdminStoresLocked() []AdminStoreRecord {
	items := make([]AdminStoreRecord, 0, len(s.stores))
	today := time.Now().Format("2006-01-02")
	for _, store := range s.stores {
		record := AdminStoreRecord{
			ID:               store.ID,
			Code:             store.Code,
			Name:             store.Name,
			ManagerName:      store.Manager,
			City:             store.City,
			Address:          store.Address,
			ContactPhone:     s.settings.Brand.ServicePhone,
			BusinessHours:    "10:00 - 22:00",
			Status:           mapStoreStatus(store.Status),
			CashierDevices:   1,
			PendingTasks:     0,
			TodayAmount:      0,
			TodayOrders:      0,
			LastSettlementAt: "尚未结算",
			Tags:             []string{"真实门店", store.Code},
		}
		for _, order := range s.cashierOrders {
			if order.StoreID != store.ID {
				continue
			}
			if order.CreatedAt.Format("2006-01-02") == today {
				record.TodayOrders++
				record.TodayAmount += order.TotalAmount
			}
			if order.CreatedAt.After(parseAdminTime(record.LastSettlementAt)) {
				record.LastSettlementAt = order.CreatedAt.Format("2006-01-02 15:04")
			}
		}
		for _, order := range s.recycleOrders {
			if order.StoreID == store.ID && order.Status == "draft" {
				record.PendingTasks++
			}
		}
		items = append(items, record)
	}
	return items
}

func (s *MockStore) buildDynamicAdminUsersLocked() []AdminUserAccount {
	items := make([]AdminUserAccount, 0, len(s.usersByID))
	for _, user := range s.usersByID {
		roleKey := normalizeAdminRoleKey(user.RoleCode)
		role := s.findAdminRoleLocked(roleKey)
		storeNames := make([]string, 0, len(user.StoreIDs))
		for _, storeID := range user.StoreIDs {
			if store, ok := s.getStoreByIDLocked(storeID); ok {
				storeNames = append(storeNames, store.Name)
			}
		}
		if len(storeNames) == 0 && user.DataScope == "org_all" {
			storeNames = []string{"全部门店"}
		}
		items = append(items, AdminUserAccount{
			ID:          user.ID,
			Name:        user.DisplayName,
			Account:     user.Username,
			Phone:       "待补充",
			RoleKey:     roleKey,
			RoleName:    user.RoleName,
			DataScope:   mapAdminDataScope(user.DataScope),
			StoreIDs:    append([]string(nil), user.StoreIDs...),
			StoreNames:  storeNames,
			Status:      mapUserStatus(user.Status),
			LastLoginAt: time.Now().Format("2006-01-02 15:04"),
			Abilities:   append([]AdminAbilityCode(nil), role.Abilities...),
		})
	}
	return items
}

func (s *MockStore) buildDynamicAdminProductsLocked() []AdminProductRecord {
	items := make([]AdminProductRecord, 0, len(s.catalogProducts))
	for _, product := range s.catalogProducts {
		items = append(items, AdminProductRecord{
			ID:         product.ID,
			Name:       product.Name,
			SKU:        product.SKU,
			Category:   product.Category,
			Price:      product.RetailPrice,
			GramWeight: product.GramWeight,
			Status:     product.Status,
			StoreNames: append([]string(nil), product.Stores...),
			Tags:       append([]string(nil), product.Tags...),
		})
	}
	return items
}

func (s *MockStore) buildDynamicCashierOrderViewsLocked() []AdminCashierOrderView {
	items := make([]AdminCashierOrderView, 0, len(s.cashierOrders))
	for i := len(s.cashierOrders) - 1; i >= 0; i-- {
		order := s.cashierOrders[i]
		items = append(items, AdminCashierOrderView{
			ID:            order.ID,
			OrderNo:       order.OrderNo,
			StoreID:       order.StoreID,
			StoreName:     order.StoreName,
			PaymentMethod: normalizePaymentMethod(order.PaymentMethod),
			Status:        normalizeCashierStatus(order.Status),
			TotalAmount:   order.TotalAmount,
			ItemCount:     len(order.Items),
			CreatedBy:     order.CreatedBy,
			CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04"),
			Remark:        order.Remark,
		})
	}
	return items
}

func (s *MockStore) buildDynamicRecycleOrderViewsLocked() []AdminRecycleOrderView {
	items := make([]AdminRecycleOrderView, 0, len(s.recycleOrders))
	for _, order := range s.recycleOrders {
		view := AdminRecycleOrderView{
			ID:              order.ID,
			OrderNo:         order.OrderNo,
			StoreID:         order.StoreID,
			StoreName:       order.StoreName,
			Status:          order.Status,
			CustomerName:    order.CustomerName,
			CustomerPhone:   order.CustomerPhone,
			EstimatedAmount: order.EstimatedAmount,
			ConfirmedAmount: order.ConfirmedAmount,
			PhotoCount:      len(order.AttachmentURLs),
			CreatedBy:       order.CreatedBy,
			CreatedAt:       order.CreatedAt.Format("2006-01-02 15:04"),
			Remark:          order.Remark,
		}
		if len(order.Attachments) > 0 {
			view.PhotoCount = len(order.Attachments)
		}
		if order.ConfirmedAt != nil {
			view.ConfirmedAt = order.ConfirmedAt.Format("2006-01-02 15:04")
		}
		items = append(items, view)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items
}

func (s *MockStore) buildDynamicPaymentRecordsLocked() []AdminPaymentRecord {
	items := make([]AdminPaymentRecord, 0, len(s.paymentTransactions))
	for i := len(s.paymentTransactions) - 1; i >= 0; i-- {
		payment := s.paymentTransactions[i]
		paidAt := payment.CreatedAt.Format("2006-01-02 15:04")
		if payment.PaidAt != nil {
			paidAt = payment.PaidAt.Format("2006-01-02 15:04")
		}
		items = append(items, AdminPaymentRecord{
			ID:             payment.ID,
			PaymentNo:      payment.PaymentNo,
			OrderNo:        payment.OrderNo,
			BizType:        payment.BizType,
			StoreID:        payment.StoreID,
			StoreName:      payment.StoreName,
			Amount:         payment.Amount,
			Method:         payment.Method,
			Status:         payment.Status,
			CallbackStatus: payment.CallbackStatus,
			PaidAt:         paidAt,
			OperatorName:   payment.OperatorName,
			CustomerLabel:  payment.CustomerLabel,
			Remark:         payment.Remark,
			Anomaly:        payment.Anomaly,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].PaidAt > items[j].PaidAt
	})
	return items
}

func (s *MockStore) buildDynamicSystemProfileLocked() AdminSystemProfile {
	profile := s.adminSystemProfile
	profile.BrandName = s.settings.Brand.Name
	profile.ServicePhone = s.settings.Brand.ServicePhone
	profile.ReceiptTitle = s.settings.Brand.ReceiptTitle
	profile.MinPhotoCount = s.settings.Recycle.MinPhotoCount
	profile.MaxPhotoCount = s.settings.Recycle.MaxPhotoCount
	profile.RequireExactThree = s.settings.Recycle.RequireExactThree
	profile.RequireIDCheck = s.settings.Recycle.RequireIDCheck
	profile.WechatPayEnabled = s.settings.Payments.EnableWechatPay
	profile.CashEnabled = s.settings.Payments.EnableCash
	profile.BankTransferEnabled = s.settings.Payments.EnableBankTransfer
	return profile
}

func (s *MockStore) findAdminRoleLocked(roleKey string) AdminRoleTemplate {
	for _, role := range s.adminRoles {
		if role.Key == roleKey {
			return role
		}
	}
	return AdminRoleTemplate{}
}

func (s *MockStore) getStoreByIDLocked(storeID string) (StoreInfo, bool) {
	for _, store := range s.stores {
		if store.ID == storeID {
			return store, true
		}
	}
	return StoreInfo{}, false
}

func productVisibleInStores(product AdminProductRecord, visibleStoreNames []string) bool {
	for _, storeName := range product.StoreNames {
		for _, visible := range visibleStoreNames {
			if storeName == visible {
				return true
			}
		}
	}
	return false
}

func mapAdminDataScope(scope string) string {
	switch scope {
	case "org_all":
		return "all_stores"
	case "assigned_stores":
		return "assigned_store"
	default:
		return "self"
	}
}

func mapStoreStatus(status string) string {
	switch status {
	case "active":
		return "active"
	case "disabled":
		return "disabled"
	default:
		return "pending"
	}
}

func mapUserStatus(status string) string {
	switch status {
	case "active":
		return "enabled"
	case "disabled":
		return "disabled"
	default:
		return "invited"
	}
}

func normalizeBaseStoreStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "active":
		return "active"
	case "disabled":
		return "disabled"
	default:
		return "pending"
	}
}

func roleNameForKey(roleKey string) string {
	switch normalizeAdminRoleKey(roleKey) {
	case "owner":
		return "老板"
	case "manager":
		return "店长"
	default:
		return "员工"
	}
}

func baseRoleCodeFromAdmin(roleKey string) string {
	switch normalizeAdminRoleKey(roleKey) {
	case "owner":
		return "owner"
	case "manager":
		return "manager"
	default:
		return "cashier"
	}
}

func baseDataScopeForRole(roleKey string) string {
	switch normalizeAdminRoleKey(roleKey) {
	case "owner":
		return "org_all"
	case "manager":
		return "assigned_stores"
	default:
		return "self"
	}
}

func baseDataScopeFromAdmin(scope string) string {
	switch strings.TrimSpace(scope) {
	case "all_stores":
		return "org_all"
	case "assigned_store":
		return "assigned_stores"
	default:
		return "self"
	}
}

func baseUserStatusFromAdmin(status string) string {
	switch strings.TrimSpace(status) {
	case "enabled":
		return "active"
	case "disabled":
		return "disabled"
	default:
		return "pending"
	}
}

func permissionsForRoleLocked(roleKey string, roles []RoleTemplate) []string {
	baseRoleCode := baseRoleCodeFromAdmin(roleKey)
	for _, role := range roles {
		if role.Code == baseRoleCode {
			return append([]string(nil), role.Permissions...)
		}
	}
	return nil
}

func (s *MockStore) buildAdminStoreRecordLocked(store StoreInfo) AdminStoreRecord {
	today := time.Now().Format("2006-01-02")
	record := AdminStoreRecord{
		ID:               store.ID,
		Code:             store.Code,
		Name:             store.Name,
		ManagerName:      store.Manager,
		City:             store.City,
		Address:          store.Address,
		ContactPhone:     s.settings.Brand.ServicePhone,
		BusinessHours:    "10:00 - 22:00",
		Status:           mapStoreStatus(store.Status),
		CashierDevices:   1,
		PendingTasks:     0,
		TodayAmount:      0,
		TodayOrders:      0,
		LastSettlementAt: "尚未结算",
		Tags:             []string{"真实门店", store.Code},
	}
	for _, order := range s.cashierOrders {
		if order.StoreID != store.ID {
			continue
		}
		if order.CreatedAt.Format("2006-01-02") == today {
			record.TodayOrders++
			record.TodayAmount += order.TotalAmount
		}
		if order.CreatedAt.After(parseAdminTime(record.LastSettlementAt)) {
			record.LastSettlementAt = order.CreatedAt.Format("2006-01-02 15:04")
		}
	}
	for _, order := range s.recycleOrders {
		if order.StoreID == store.ID && order.Status == "draft" {
			record.PendingTasks++
		}
	}
	return record
}

func (s *MockStore) buildAdminUserRecordLocked(account UserAccount) AdminUserAccount {
	roleKey := normalizeAdminRoleKey(account.RoleCode)
	role := s.findAdminRoleLocked(roleKey)
	storeNames := make([]string, 0, len(account.StoreIDs))
	for _, storeID := range account.StoreIDs {
		if store, ok := s.getStoreByIDLocked(storeID); ok {
			storeNames = append(storeNames, store.Name)
		}
	}
	if len(storeNames) == 0 && account.DataScope == "org_all" {
		storeNames = []string{"全部门店"}
	}
	return AdminUserAccount{
		ID:          account.ID,
		Name:        account.DisplayName,
		Account:     account.Username,
		Phone:       "待补充",
		RoleKey:     roleKey,
		RoleName:    account.RoleName,
		DataScope:   mapAdminDataScope(account.DataScope),
		StoreIDs:    append([]string(nil), account.StoreIDs...),
		StoreNames:  storeNames,
		Status:      mapUserStatus(account.Status),
		LastLoginAt: time.Now().Format("2006-01-02 15:04"),
		Abilities:   append([]AdminAbilityCode(nil), role.Abilities...),
	}
}

func buildAdminProductRecord(product CatalogProduct) AdminProductRecord {
	return AdminProductRecord{
		ID:         product.ID,
		Name:       product.Name,
		SKU:        product.SKU,
		Category:   product.Category,
		Price:      product.RetailPrice,
		GramWeight: product.GramWeight,
		Status:     product.Status,
		StoreNames: append([]string(nil), product.Stores...),
		Tags:       append([]string(nil), product.Tags...),
	}
}

func normalizePaymentMethod(method string) string {
	switch method {
	case "wechat_pay":
		return "wechat"
	case "cash":
		return "cash"
	case "bank_transfer":
		return "bank_transfer"
	default:
		return method
	}
}

func normalizeCashierStatus(status string) string {
	switch status {
	case "paid":
		return "paid"
	case "refunded":
		return "refunded"
	default:
		return "pending"
	}
}

func paymentCallbackStatus(method string) string {
	if normalizePaymentMethod(method) == "wechat" {
		return "delivered"
	}
	return "delivered"
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func parseAdminTime(value string) time.Time {
	parsed, err := time.Parse("2006-01-02 15:04", value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func (s *MockStore) updateAdminRoleTemplate(user UserAccount, roleID int, update AdminRoleTemplate) (AdminRoleTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.adminRoles {
		if item.ID != roleID {
			continue
		}
		if update.Name != "" {
			item.Name = update.Name
		}
		if update.Description != "" {
			item.Description = update.Description
		}
		if update.DataScope != "" {
			item.DataScope = update.DataScope
		}
		item.Abilities = append([]AdminAbilityCode(nil), update.Abilities...)
		s.adminRoles[index] = item
		if s.persistence != nil {
			_ = s.persistAdminRolesLocked()
		}
		s.appendAuditLogLocked("角色权限", "保存角色模板", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已更新", item.Name))
		return item, nil
	}
	return AdminRoleTemplate{}, errAdminRoleNotFound
}

func (s *MockStore) updateAdminPaymentConfig(user UserAccount, update AdminPaymentConfig) AdminPaymentConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adminPaymentConfig = update
	s.settings.Payments.EnableWechatPay = update.Wechat.Enabled
	s.settings.Payments.EnableCash = update.Cash.Enabled
	s.settings.Payments.EnableBankTransfer = update.BankTransfer.Enabled
	s.adminSystemProfile = s.buildDynamicSystemProfileLocked()
	if s.persistence != nil {
		_ = s.persistSettingsLocked()
		_ = s.persistence.saveConfig(context.Background(), configKeyAdminPayment, s.adminPaymentConfig)
		_ = s.persistence.saveConfig(context.Background(), configKeyAdminSystem, s.adminSystemProfile)
	}
	s.appendAuditLogLocked("支付中心", "修改支付配置", user.DisplayName, "warning", "high", "支付配置已更新，待真实联调校验。")
	return s.adminPaymentConfig
}

func (s *MockStore) updateAdminStore(user UserAccount, storeID string, update AdminStoreRecord) (AdminStoreRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.canAccessStore(user, storeID) {
		return AdminStoreRecord{}, errUnauthorizedStore
	}

	for index, store := range s.stores {
		if store.ID != storeID {
			continue
		}
		if strings.TrimSpace(update.Code) != "" {
			store.Code = strings.TrimSpace(update.Code)
		}
		if strings.TrimSpace(update.Name) != "" {
			store.Name = strings.TrimSpace(update.Name)
		}
		if strings.TrimSpace(update.ManagerName) != "" {
			store.Manager = strings.TrimSpace(update.ManagerName)
		}
		if strings.TrimSpace(update.City) != "" {
			store.City = strings.TrimSpace(update.City)
		}
		if strings.TrimSpace(update.Address) != "" {
			store.Address = strings.TrimSpace(update.Address)
		}
		if strings.TrimSpace(update.Status) != "" {
			store.Status = normalizeBaseStoreStatus(update.Status)
		}
		s.stores[index] = store
		if s.persistence != nil {
			_ = s.persistStoresLocked()
		}
		record := s.buildAdminStoreRecordLocked(store)
		s.appendAuditLogLocked("门店管理", "更新门店资料", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已更新", record.Name))
		return record, nil
	}
	return AdminStoreRecord{}, errAdminStoreNotFound
}

func (s *MockStore) updateAdminUser(user UserAccount, userID string, update AdminUserAccount) (AdminUserAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for userMapID, account := range s.usersByID {
		if userMapID != userID {
			continue
		}
		if user.DataScope != "org_all" && account.ID != user.ID && !hasIntersect(account.StoreIDs, user.StoreIDs) {
			return AdminUserAccount{}, errUnauthorizedStore
		}
		if strings.TrimSpace(update.Name) != "" {
			account.DisplayName = strings.TrimSpace(update.Name)
		}
		if strings.TrimSpace(update.Account) != "" {
			delete(s.usersByUsername, account.Username)
			account.Username = strings.TrimSpace(update.Account)
		}
		if strings.TrimSpace(update.RoleKey) != "" {
			roleKey := normalizeAdminRoleKey(update.RoleKey)
			account.RoleCode = baseRoleCodeFromAdmin(roleKey)
			account.RoleName = roleNameForKey(roleKey)
			account.DataScope = baseDataScopeForRole(roleKey)
			account.Permissions = permissionsForRoleLocked(roleKey, s.roles)
		}
		if strings.TrimSpace(update.RoleName) != "" {
			account.RoleName = strings.TrimSpace(update.RoleName)
		}
		if strings.TrimSpace(update.DataScope) != "" {
			account.DataScope = baseDataScopeFromAdmin(update.DataScope)
		}
		if update.StoreIDs != nil {
			account.StoreIDs = append([]string(nil), update.StoreIDs...)
		}
		if strings.TrimSpace(update.Status) != "" {
			account.Status = baseUserStatusFromAdmin(update.Status)
		}
		s.usersByID[userMapID] = account
		s.usersByUsername[account.Username] = account
		if s.persistence != nil {
			_ = s.persistUsersLocked()
		}
		record := s.buildAdminUserRecordLocked(account)
		s.appendAuditLogLocked("账号管理", "更新账号资料", user.DisplayName, "warning", "high", fmt.Sprintf("%s 账号资料已更新", record.Name))
		return record, nil
	}
	return AdminUserAccount{}, errAdminUserNotFound
}

func (s *MockStore) updateAdminProduct(user UserAccount, productID string, update AdminProductRecord) (AdminProductRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, product := range s.catalogProducts {
		if product.ID != productID {
			continue
		}
		if user.DataScope != "org_all" && !hasIntersect(product.StoreIDs, user.StoreIDs) {
			return AdminProductRecord{}, errUnauthorizedStore
		}
		if strings.TrimSpace(update.Name) != "" {
			product.Name = strings.TrimSpace(update.Name)
		}
		if strings.TrimSpace(update.SKU) != "" {
			product.SKU = strings.TrimSpace(update.SKU)
		}
		if strings.TrimSpace(update.Category) != "" {
			product.Category = strings.TrimSpace(update.Category)
		}
		if update.Price >= 0 {
			product.RetailPrice = update.Price
		}
		if update.GramWeight >= 0 {
			product.GramWeight = update.GramWeight
		}
		if strings.TrimSpace(update.Status) != "" {
			product.Status = strings.TrimSpace(update.Status)
		}
		if update.StoreNames != nil {
			product.Stores = append([]string(nil), update.StoreNames...)
		}
		if update.Tags != nil {
			product.Tags = append([]string(nil), update.Tags...)
		}
		s.catalogProducts[index] = product
		if s.persistence != nil {
			_ = s.persistProductsLocked()
		}
		record := buildAdminProductRecord(product)
		s.appendAuditLogLocked("商品管理", "更新商品资料", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已更新", record.Name))
		return record, nil
	}
	return AdminProductRecord{}, errAdminProductNotFound
}

func (s *MockStore) updateAdminSystemProfile(user UserAccount, update AdminSystemProfile) AdminSystemProfile {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(update.BrandName) != "" {
		s.adminSystemProfile.BrandName = strings.TrimSpace(update.BrandName)
	}
	if strings.TrimSpace(update.ServicePhone) != "" {
		s.adminSystemProfile.ServicePhone = strings.TrimSpace(update.ServicePhone)
	}
	if strings.TrimSpace(update.ReceiptTitle) != "" {
		s.adminSystemProfile.ReceiptTitle = strings.TrimSpace(update.ReceiptTitle)
	}
	if update.MinPhotoCount > 0 {
		s.adminSystemProfile.MinPhotoCount = update.MinPhotoCount
	}
	if update.MaxPhotoCount > 0 {
		s.adminSystemProfile.MaxPhotoCount = update.MaxPhotoCount
	}
	s.adminSystemProfile.RequireExactThree = update.RequireExactThree
	s.adminSystemProfile.RequireIDCheck = update.RequireIDCheck
	s.adminSystemProfile.WechatPayEnabled = update.WechatPayEnabled
	s.adminSystemProfile.CashEnabled = update.CashEnabled
	s.adminSystemProfile.BankTransferEnabled = update.BankTransferEnabled
	if strings.TrimSpace(update.DomainName) != "" {
		s.adminSystemProfile.DomainName = strings.TrimSpace(update.DomainName)
	}
	if strings.TrimSpace(update.DomainStatus) != "" {
		s.adminSystemProfile.DomainStatus = strings.TrimSpace(update.DomainStatus)
	}
	if strings.TrimSpace(update.OSSStatus) != "" {
		s.adminSystemProfile.OSSStatus = strings.TrimSpace(update.OSSStatus)
	}
	if strings.TrimSpace(update.AppIDStatus) != "" {
		s.adminSystemProfile.AppIDStatus = strings.TrimSpace(update.AppIDStatus)
	}
	if strings.TrimSpace(update.MerchantStatus) != "" {
		s.adminSystemProfile.MerchantStatus = strings.TrimSpace(update.MerchantStatus)
	}
	if strings.TrimSpace(update.PrinterStatus) != "" {
		s.adminSystemProfile.PrinterStatus = strings.TrimSpace(update.PrinterStatus)
	}
	s.settings.Brand.Name = s.adminSystemProfile.BrandName
	s.settings.Brand.ServicePhone = s.adminSystemProfile.ServicePhone
	s.settings.Brand.ReceiptTitle = s.adminSystemProfile.ReceiptTitle
	s.settings.Recycle.MinPhotoCount = s.adminSystemProfile.MinPhotoCount
	s.settings.Recycle.MaxPhotoCount = s.adminSystemProfile.MaxPhotoCount
	s.settings.Recycle.RequireExactThree = s.adminSystemProfile.RequireExactThree
	s.settings.Recycle.RequireIDCheck = s.adminSystemProfile.RequireIDCheck
	s.settings.Payments.EnableWechatPay = s.adminSystemProfile.WechatPayEnabled
	s.settings.Payments.EnableCash = s.adminSystemProfile.CashEnabled
	s.settings.Payments.EnableBankTransfer = s.adminSystemProfile.BankTransferEnabled
	s.settings.UpdatedBy = user.DisplayName
	s.settings.UpdatedAt = time.Now()
	if s.persistence != nil {
		_ = s.persistSettingsLocked()
		_ = s.persistence.saveConfig(context.Background(), configKeyAdminSystem, s.adminSystemProfile)
	}
	s.appendAuditLogLocked("系统配置", "更新系统配置", user.DisplayName, "warning", "high", "系统配置与打印准备信息已更新。")
	return s.adminSystemProfile
}

func (s *MockStore) createAdminStore(user UserAccount) (AdminStoreRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.DataScope != "org_all" {
		return AdminStoreRecord{}, errUnauthorizedStore
	}

	store := StoreInfo{
		ID:        fmt.Sprintf("store-new-%03d", len(s.stores)+1),
		OrgID:     "org-gold-demo",
		Code:      fmt.Sprintf("NEW-%03d", len(s.stores)+1),
		Name:      fmt.Sprintf("新门店 %d", len(s.stores)+1),
		City:      "待填写",
		Address:   "待填写地址",
		Manager:   "待分配",
		Status:    "pending",
		IsDefault: false,
	}
	s.stores = append([]StoreInfo{store}, s.stores...)
	if s.persistence != nil {
		_ = s.persistStoresLocked()
	}
	record := s.buildAdminStoreRecordLocked(store)
	s.appendAuditLogLocked("门店管理", "新建门店", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已创建", record.Name))
	return record, nil
}

func (s *MockStore) createAdminUser(user UserAccount) (AdminUserAccount, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.DataScope != "org_all" {
		return AdminUserAccount{}, errUnauthorizedStore
	}

	account := UserAccount{
		ID:          fmt.Sprintf("user-new-%03d", len(s.usersByID)+1),
		OrgID:       "org-gold-demo",
		Username:    fmt.Sprintf("new.user.%03d", len(s.usersByID)+1),
		DisplayName: fmt.Sprintf("新账号 %d", len(s.usersByID)+1),
		Password:    "ChangeMe123!",
		RoleCode:    "cashier",
		RoleName:    "员工",
		DataScope:   "self",
		StoreIDs:    []string{},
		Permissions: permissionsForRoleLocked("cashier", s.roles),
		Status:      "pending",
	}
	s.usersByID[account.ID] = account
	s.usersByUsername[account.Username] = account
	if s.persistence != nil {
		_ = s.persistUsersLocked()
	}
	record := s.buildAdminUserRecordLocked(account)
	s.appendAuditLogLocked("账号管理", "新建账号", user.DisplayName, "warning", "high", fmt.Sprintf("%s 已创建，待绑定角色和门店", record.Name))
	return record, nil
}

func (s *MockStore) createAdminProduct(user UserAccount) (AdminProductRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user.DataScope != "org_all" {
		return AdminProductRecord{}, errUnauthorizedStore
	}

	product := CatalogProduct{
		ID:               fmt.Sprintf("prd-new-%03d", len(s.catalogProducts)+1),
		OrgID:            "org-gold-demo",
		Name:             fmt.Sprintf("新商品 %d", len(s.catalogProducts)+1),
		SKU:              fmt.Sprintf("NEW-SKU-%03d", len(s.catalogProducts)+1),
		Category:         "待分类",
		Status:           "draft",
		Stores:           []string{"待分配门店"},
		Tags:             []string{"新建", "待完善"},
		RecommendedScene: "待完善场景",
		QuoteLeadTime:    "待设置",
	}
	s.catalogProducts = append([]CatalogProduct{product}, s.catalogProducts...)
	if s.persistence != nil {
		_ = s.persistProductsLocked()
	}
	record := buildAdminProductRecord(product)
	s.appendAuditLogLocked("商品管理", "新建商品", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已创建", record.Name))
	return record, nil
}

func (s *MockStore) updateAdminPrintTemplate(user UserAccount, update AdminPrintTemplate) AdminPrintTemplate {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.adminPrintTemplate = update
	if s.persistence != nil {
		_ = s.persistence.saveConfig(context.Background(), configKeyAdminPrint, s.adminPrintTemplate)
	}
	s.appendAuditLogLocked("打印配置", "更新打印模板", user.DisplayName, "warning", "high", "小票/标签/回收打印模板结构已更新。")
	return s.adminPrintTemplate
}

func (s *MockStore) appendAuditLogLocked(module, action, operatorName, result, riskLevel, summary string) {
	log := AdminAuditLogRecord{
		ID:           fmt.Sprintf("log-%03d", len(s.adminAuditLogs)+1),
		Module:       module,
		Action:       action,
		OperatorName: operatorName,
		Result:       result,
		RiskLevel:    riskLevel,
		Summary:      summary,
		CreatedAt:    time.Now().Format("2006-01-02 15:04"),
	}
	s.adminAuditLogs = append([]AdminAuditLogRecord{log}, s.adminAuditLogs...)
}

func buildAdminFixtures() (
	[]AdminRoleTemplate,
	[]AdminAbilityGroup,
	[]AdminStoreRecord,
	[]AdminUserAccount,
	[]AdminProductRecord,
	AdminPaymentConfig,
	AdminPrintTemplate,
	[]AdminPaymentRecord,
	[]AdminCashierOrderView,
	[]AdminRecycleOrderView,
	AdminSystemProfile,
	[]AdminAuditLogRecord,
	AdminTemplateInitPlan,
) {
	groups := newAdminAbilityGroups()
	allAbilities := allAdminAbilities(groups)
	roles := []AdminRoleTemplate{
		{ID: 1, Key: "owner", Name: "老板模板", Description: "全局主控权限，覆盖组织、支付、系统与模板初始化。", DataScope: "all_stores", MemberCount: 1, Locked: true, Abilities: allAbilities},
		{ID: 2, Key: "manager", Name: "店长模板", Description: "本门店运营台，可管理门店资料、账号和查看支付流水。", DataScope: "assigned_store", MemberCount: 3, Locked: true, Abilities: []AdminAbilityCode{"dashboard.view", "store.view", "store.manage", "user.view", "user.manage", "product.view", "product.manage", "order.view", "recycle.view", "payment.record.view"}},
		{ID: 3, Key: "clerk", Name: "员工模板", Description: "第一版默认不开放后台入口，仅保留模板位用于后续扩展。", DataScope: "self", MemberCount: 1, Locked: true, Abilities: []AdminAbilityCode{}},
	}
	stores := []AdminStoreRecord{
		{ID: "store-sz-luohu", Code: "SZ-LH-01", Name: "罗湖旗舰店", ManagerName: "李店长", City: "深圳", Address: "罗湖区深南东路 1888 号 1F", ContactPhone: "0755-8899 1201", BusinessHours: "10:00 - 22:00", Status: "active", CashierDevices: 3, PendingTasks: 2, TodayAmount: 18260, TodayOrders: 16, LastSettlementAt: "2026-05-10 10:05", Tags: []string{"黄金回收", "收银台", "支持微信支付"}},
		{ID: "store-sz-nanshan", Code: "SZ-NS-02", Name: "南山体验店", ManagerName: "王店长", City: "深圳", Address: "南山区科苑路 66 号 B1", ContactPhone: "0755-8899 1202", BusinessHours: "10:00 - 21:30", Status: "active", CashierDevices: 2, PendingTasks: 1, TodayAmount: 13680, TodayOrders: 11, LastSettlementAt: "2026-05-10 09:35", Tags: []string{"黄金回收", "体验店", "现金记账"}},
		{ID: "store-gz-tianhe", Code: "GZ-TH-01", Name: "广州天河店", ManagerName: "空缺", City: "广州", Address: "天河区体育西路 88 号", ContactPhone: "020-8899 1201", BusinessHours: "10:00 - 22:00", Status: "disabled", CashierDevices: 0, PendingTasks: 3, TodayAmount: 0, TodayOrders: 0, LastSettlementAt: "2026-05-06 21:30", Tags: []string{"暂停营业", "待重新授权"}},
	}
	users := []AdminUserAccount{
		{ID: "user-owner-001", Name: "陈老板", Account: "boss", Phone: "138****1001", RoleKey: "owner", RoleName: "老板", DataScope: "all_stores", StoreIDs: []string{}, StoreNames: []string{"全部门店"}, Status: "enabled", LastLoginAt: "2026-05-10 09:12", Abilities: allAbilities},
		{ID: "user-manager-001", Name: "李店长", Account: "manager.sz", Phone: "138****2108", RoleKey: "manager", RoleName: "店长", DataScope: "assigned_store", StoreIDs: []string{"store-sz-luohu"}, StoreNames: []string{"罗湖旗舰店"}, Status: "enabled", LastLoginAt: "2026-05-10 08:42", Abilities: roles[1].Abilities},
		{ID: "user-manager-002", Name: "王店长", Account: "manager.gz", Phone: "138****2109", RoleKey: "manager", RoleName: "店长", DataScope: "assigned_store", StoreIDs: []string{"store-sz-nanshan"}, StoreNames: []string{"南山体验店"}, Status: "enabled", LastLoginAt: "2026-05-09 20:16", Abilities: roles[1].Abilities},
	}
	products := []AdminProductRecord{
		{ID: "prd-001", Name: "足金项链", SKU: "GJ-XL-001", Category: "项链", Price: 3298, GramWeight: 8.6, Status: "active", StoreNames: []string{"罗湖旗舰店", "南山体验店"}, Tags: []string{"热卖", "足金"}},
		{ID: "prd-002", Name: "古法手镯", SKU: "GJ-SZ-018", Category: "手镯", Price: 5680, GramWeight: 15.2, Status: "active", StoreNames: []string{"罗湖旗舰店"}, Tags: []string{"回购高", "礼盒装"}},
		{ID: "prd-003", Name: "回收服务单", SKU: "REC-SRV-001", Category: "回收服务", Price: 0, GramWeight: 0, Status: "draft", StoreNames: []string{"罗湖旗舰店", "南山体验店", "广州天河店"}, Tags: []string{"系统单据", "非销售商品"}},
	}
	records := []AdminPaymentRecord{
		{ID: "pay-001", PaymentNo: "PAY202605100001", OrderNo: "ORD202605100001", BizType: "retail", StoreID: "store-sz-luohu", StoreName: "罗湖旗舰店", Amount: 3298, Method: "wechat", Status: "paid", CallbackStatus: "delivered", PaidAt: "2026-05-10 10:08", OperatorName: "李店长", CustomerLabel: "顾客 A", Remark: "足金项链成交", Anomaly: false},
		{ID: "pay-002", PaymentNo: "PAY202605100002", OrderNo: "REC202605100014", BizType: "recycle", StoreID: "store-sz-luohu", StoreName: "罗湖旗舰店", Amount: 5680, Method: "cash", Status: "paid", CallbackStatus: "delivered", PaidAt: "2026-05-10 09:32", OperatorName: "李店长", CustomerLabel: "回收客户 B", Remark: "回收单已签字", Anomaly: false},
		{ID: "pay-003", PaymentNo: "PAY202605090087", OrderNo: "ORD202605090145", BizType: "retail", StoreID: "store-sz-nanshan", StoreName: "南山体验店", Amount: 2680, Method: "wechat", Status: "paid", CallbackStatus: "pending", PaidAt: "2026-05-09 19:42", OperatorName: "王店长", CustomerLabel: "顾客 C", Remark: "回调待补偿", Anomaly: true},
	}
	orders := []AdminCashierOrderView{
		{ID: "ord-001", OrderNo: "ORD202605100001", StoreID: "store-sz-luohu", StoreName: "罗湖旗舰店", PaymentMethod: "wechat", Status: "paid", TotalAmount: 3298, ItemCount: 2, CreatedBy: "李店长", CreatedAt: "2026-05-10 10:08", Remark: "足金项链成交"},
		{ID: "ord-002", OrderNo: "ORD202605090145", StoreID: "store-sz-nanshan", StoreName: "南山体验店", PaymentMethod: "wechat", Status: "pending", TotalAmount: 2680, ItemCount: 1, CreatedBy: "王店长", CreatedAt: "2026-05-09 19:42", Remark: "等待支付补偿确认"},
		{ID: "ord-003", OrderNo: "ORD202605070201", StoreID: "store-gz-tianhe", StoreName: "广州天河店", PaymentMethod: "cash", Status: "refunded", TotalAmount: 980, ItemCount: 1, CreatedBy: "系统迁移", CreatedAt: "2026-05-07 20:20", Remark: "停店前历史订单"},
	}
	recycleOrders := []AdminRecycleOrderView{
		{ID: "rec-001", OrderNo: "REC202605100014", StoreID: "store-sz-luohu", StoreName: "罗湖旗舰店", Status: "confirmed", CustomerName: "周女士", CustomerPhone: "134****2188", EstimatedAmount: 5800, ConfirmedAmount: 5680, PhotoCount: 3, CreatedBy: "李店长", CreatedAt: "2026-05-10 09:18", ConfirmedAt: "2026-05-10 09:32", Remark: "三张现场图已留存"},
		{ID: "rec-002", OrderNo: "REC202605090021", StoreID: "store-sz-nanshan", StoreName: "南山体验店", Status: "draft", CustomerName: "王先生", CustomerPhone: "135****1018", EstimatedAmount: 4320, ConfirmedAmount: 0, PhotoCount: 2, CreatedBy: "王店长", CreatedAt: "2026-05-09 16:18", Remark: "待客户最终确认回收价"},
	}
	systemProfile := AdminSystemProfile{
		BrandName: "金匠馆回收", ServicePhone: "400-888-2026", ReceiptTitle: "黄金回收收银系统",
		MinPhotoCount: 2, MaxPhotoCount: 3, RequireExactThree: false, RequireIDCheck: true,
		WechatPayEnabled: true, CashEnabled: true, BankTransferEnabled: false,
		DomainName: "jinjiangguan.com", DomainStatus: "已购买，实名认证审核中",
		OSSStatus: "已进入 OSS 控制台，Bucket 待正式规划", AppIDStatus: "待补充 AppID",
		MerchantStatus: "待补充商户号与证书", PrinterStatus: "待确定小票机 / 标签机型号，客户倾向彩色小票",
	}
	auditLogs := []AdminAuditLogRecord{
		{ID: "log-001", Module: "支付中心", Action: "修改支付配置", OperatorName: "陈老板", Result: "warning", RiskLevel: "high", Summary: "支付回调地址待联调，当前仍为测试占位。", CreatedAt: "2026-05-10 11:05"},
		{ID: "log-002", Module: "角色权限", Action: "冻结店长模板", OperatorName: "陈老板", Result: "success", RiskLevel: "medium", Summary: "已确认店长默认只能看所属门店数据。", CreatedAt: "2026-05-10 10:26"},
		{ID: "log-003", Module: "系统配置", Action: "调整拍照规则", OperatorName: "系统", Result: "info", RiskLevel: "low", Summary: "当前规则为至少 2 张最多 3 张，保留切换为必须 3 张能力。", CreatedAt: "2026-05-10 09:40"},
	}
	return roles, groups, stores, users, products, newAdminPaymentConfig(), newAdminPrintTemplate(), records, orders, recycleOrders, systemProfile, auditLogs, newAdminTemplatePlan()
}

func normalizeAdminRoleKey(roleCode string) string {
	switch strings.TrimSpace(roleCode) {
	case "owner":
		return "owner"
	case "manager":
		return "manager"
	default:
		return "clerk"
	}
}
