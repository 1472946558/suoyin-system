/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: admin_store.go
 * 功能描述: 状态管理
 * 作者: 廖心慈
 * 创建日期: 2026-06-05
 */

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
	errAdminMemberNotFound  = errors.New("admin member not found")
	errAdminOrderConflict   = errors.New("admin order status conflict")
)

func newAdminAbilityGroups() []AdminAbilityGroup {
	return []AdminAbilityGroup{
		{
			Key:         "dashboard",
			Label:       "首页",
			Description: "查看门店经营概况与常用业务入口。",
			Items: []AdminAbilityOption{
				{Code: "dashboard.view", Label: "查看首页", Description: "查看经营概况、提醒和快捷入口。"},
			},
		},
		{
			Key:         "organization",
			Label:       "门店与账号",
			Description: "查看门店资料和后台账号信息。",
			Items: []AdminAbilityOption{
				{Code: "store.view", Label: "查看门店", Description: "查看门店状态、营业信息和当日数据。"},
				{Code: "store.manage", Label: "管理门店", Description: "维护门店资料、负责人和营业信息。"},
				{Code: "user.view", Label: "查看账号", Description: "查看门店账号、岗位和最近登录情况。"},
				{Code: "user.manage", Label: "管理账号", Description: "维护门店账号和所属门店。"},
				{Code: "role.view", Label: "查看岗位", Description: "查看岗位范围和可用功能。"},
				{Code: "role.manage", Label: "管理岗位", Description: "调整岗位可见范围和功能分配。"},
			},
		},
		{
			Key:         "business",
			Label:       "商品与业务",
			Description: "商品、收银、回收和会员业务。",
			Items: []AdminAbilityOption{
				{Code: "product.view", Label: "查看商品", Description: "查看商品分类、编号、价格和状态。"},
				{Code: "product.manage", Label: "管理商品", Description: "维护商品资料、价格、库存和状态。"},
				{Code: "order.view", Label: "查看订单", Description: "查看收银订单和回收订单信息。"},
				{Code: "recycle.view", Label: "查看回收单", Description: "查看回收客户、照片和确认金额。"},
			},
		},
		{
			Key:         "system",
			Label:       "基础设置",
			Description: "维护品牌信息、拍照规则和打印设置。",
			Items: []AdminAbilityOption{
				{Code: "system.config.view", Label: "查看设置", Description: "查看品牌信息、拍照规则和打印设置。"},
				{Code: "system.config.manage", Label: "管理设置", Description: "维护品牌信息、客服电话和业务规则。"},
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

func newAdminPrintTemplate() AdminPrintTemplate {
	var tmpl AdminPrintTemplate
	tmpl.Receipt.Enabled = true
	tmpl.Receipt.PaperWidth = "80mm"
	tmpl.Receipt.HeaderTitle = "金匠倌收银"
	tmpl.Receipt.FooterNote = "请当面核对金额与留痕信息"
	tmpl.Receipt.ShowStoreName = true
	tmpl.Receipt.ShowOperatorName = true
	tmpl.Receipt.ShowPhotoSummary = true
	tmpl.Receipt.ShowPhotoThumbnails = false
	tmpl.Receipt.Fields = []string{"订单号", "客户信息", "商品信息", "金额", "照片已留存"}
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
		Title:       "基础资料清单",
		Description: "用于查看当前门店已准备好的品牌信息、门店资料、账号岗位和打印配置。",
		Outputs: []string{
			"品牌资料 1 份",
			"门店资料 1 组",
			"账号岗位 1 组",
			"打印与拍照规则 1 组",
		},
	}
	plan.Steps = []struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}{
		{ID: "step-1", Title: "品牌基础信息", Description: "品牌名、客服电话和小票抬头。", Status: "done"},
		{ID: "step-2", Title: "门店资料", Description: "门店名称、营业时间、负责人和联系方式。", Status: "done"},
		{ID: "step-3", Title: "账号岗位", Description: "老板和店长账号已配置。", Status: "done"},
		{ID: "step-4", Title: "业务规则", Description: "回收拍照规则和打印设置。", Status: "current"},
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

func nonNilStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	return append([]string(nil), items...)
}

func cleanUniqueStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(items))
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || value == "待分配门店" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func (s *MockStore) adminSessionUser(user UserAccount) AdminSessionUser {
	dataScope := "assigned_store"
	if user.DataScope == "org_all" {
		dataScope = "all_stores"
	}
	roleKey := normalizeAdminRoleKey(user.RoleCode)
	abilities := []AdminAbilityCode{}
	for _, role := range s.adminRoles {
		if role.Key == roleKey {
			abilities = append(abilities, sanitizeAdminAbilities(role.Abilities)...)
			break
		}
	}
	return AdminSessionUser{
		ID:          user.ID,
		Name:        user.DisplayName,
		Account:     user.Username,
		RoleKey:     roleKey,
		RoleName:    roleNameForKey(roleKey),
		DataScope:   dataScope,
		StoreIDs:    nonNilStrings(user.StoreIDs),
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
		if sessionUser.DataScope == "all_stores" || productVisibleInStores(item, visibleStoreIDs, visibleStoreNames) {
			scopedProducts = append(scopedProducts, item)
		}
	}

	scopedMembers := s.listMembersForUserLocked(user)

	title := "门店经营概况"
	subtitle := "查看本门店的收银、回收、会员和商品情况。"
	if sessionUser.RoleKey == "boss" {
		title = "门店经营概况"
		subtitle = "查看全部门店的收银、回收、会员和商品情况。"
	}

	totalAmount := 0.0
	for _, order := range scopedCashierOrders {
		totalAmount += order.TotalAmount
	}
	for _, order := range scopedRecycleOrders {
		if order.Status == "confirmed" {
			totalAmount += order.ConfirmedAmount
		}
	}

	draftRecycleCount := 0
	for _, order := range scopedRecycleOrders {
		if order.Status == "draft" {
			draftRecycleCount++
		}
	}
	lowStockCount := 0
	for _, product := range scopedProducts {
		if product.StockStatus == "low" || product.StockStatus == "review" || product.Inventory <= 3 {
			lowStockCount++
		}
	}
	todos := []AdminDashboardTodo{}
	if draftRecycleCount > 0 {
		todos = append(todos, AdminDashboardTodo{
			ID:          "task-recycle",
			Title:       fmt.Sprintf("待确认回收单 %d 笔", draftRecycleCount),
			Description: "请核对回收客户信息、照片数量和确认金额。",
			Level:       "high",
			Page:        AdminPageRecycle,
		})
	}
	if lowStockCount > 0 {
		todos = append(todos, AdminDashboardTodo{
			ID:          "task-product",
			Title:       fmt.Sprintf("需关注库存商品 %d 个", lowStockCount),
			Description: "建议检查库存偏低或待确认的商品信息。",
			Level:       "medium",
			Page:        AdminPageProducts,
		})
	}
	if len(scopedMembers) == 0 {
		todos = append(todos, AdminDashboardTodo{
			ID:          "task-member",
			Title:       "会员资料待沉淀",
			Description: "门店成交和回收后，会员资料会逐步沉淀到后台。",
			Level:       "low",
			Page:        AdminPageDashboard,
		})
	}

	notices := []string{
		"请及时核对今日收银、回收留档、会员资料和商品价格。",
	}

	return AdminBootstrap{
		CurrentUser: sessionUser,
		Dashboard: AdminDashboardSummary{
			Title:    title,
			Subtitle: subtitle,
			Metrics: []AdminDashboardMetric{
				{Key: "gross", Label: "业务金额", Value: fmt.Sprintf("¥%.0f", totalAmount), Delta: "按当前可见门店汇总", Tone: "gold"},
				{Key: "orders", Label: "收银订单", Value: fmt.Sprintf("%d", len(scopedCashierOrders)), Delta: "今日与历史订单统一查看", Tone: "emerald"},
				{Key: "recycle", Label: "回收单", Value: fmt.Sprintf("%d", len(scopedRecycleOrders)), Delta: "含待确认与已确认回收单", Tone: "slate"},
				{Key: "members", Label: "会员档案", Value: fmt.Sprintf("%d", len(scopedMembers)), Delta: "当前账号可见会员数量", Tone: "slate"},
			},
			Todos: todos,
			Shortcuts: []AdminDashboardShortcut{
				{ID: "shortcut-orders", Label: "订单对账", Hint: "统一查看收银单和回收单", Page: AdminPageOrders, Ability: "order.view"},
				{ID: "shortcut-products", Label: "商品维护", Hint: "查看商品价格、库存和状态", Page: AdminPageProducts, Ability: "product.view"},
			},
			Notices: notices,
		},
		Stores:        scopedStores,
		Users:         scopedUsers,
		Roles:         sanitizeAdminRoles(s.adminRoles),
		AbilityGroups: newAdminAbilityGroups(),
		Products:      scopedProducts,
		Members:       scopedMembers,
		CashierOrders: scopedCashierOrders,
		RecycleOrders: scopedRecycleOrders,
		PrintTemplate: s.adminPrintTemplate,
		SystemProfile: s.buildDynamicSystemProfileLocked(),
		AuditLogs:     append([]AdminAuditLogRecord(nil), s.adminAuditLogs...),
		ImportLogs:    append([]AdminProductImportLog{}, s.productImportLogs...),
		TemplateInit:  s.adminTemplateInit,
		UpdatedAt:     time.Now().Format("2006-01-02 15:04"),
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
		if mapStoreStatus(store.Status) == "disabled" {
			continue
		}
		contactPhone := store.ContactPhone
		if contactPhone == "" {
			contactPhone = s.settings.Brand.ServicePhone
		}
		businessHours := store.BusinessHours
		if businessHours == "" {
			businessHours = "10:00 - 22:00"
		}
		record := AdminStoreRecord{
			ID:               store.ID,
			Code:             store.Code,
			Name:             store.Name,
			ManagerName:      store.Manager,
			City:             store.City,
			Address:          store.Address,
			ContactPhone:     contactPhone,
			BusinessHours:    businessHours,
			Longitude:        store.Longitude,
			Latitude:         store.Latitude,
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
		if mapUserStatus(user.Status) == "disabled" {
			continue
		}
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
			ID:            user.ID,
			Name:          user.DisplayName,
			Account:       user.Username,
			Phone:         user.Phone,
			MiniAppOpenID: user.WechatOpenID,
			RoleKey:       roleKey,
			RoleName:      roleNameForKey(roleKey),
			DataScope:     mapAdminDataScope(user.DataScope),
			StoreIDs:      nonNilStrings(user.StoreIDs),
			StoreNames:    storeNames,
			Status:        mapUserStatus(user.Status),
			LastLoginAt:   time.Now().Format("2006-01-02 15:04"),
			Abilities:     sanitizeAdminAbilities(role.Abilities),
		})
	}
	return items
}

func (s *MockStore) buildDynamicAdminProductsLocked() []AdminProductRecord {
	items := make([]AdminProductRecord, 0, len(s.catalogProducts))
	for _, product := range s.catalogProducts {
		if product.Status == "disabled" {
			continue
		}
		items = append(items, s.buildAdminProductRecordLocked(product))
	}
	return items
}

func (s *MockStore) buildDynamicCashierOrderViewsLocked() []AdminCashierOrderView {
	items := make([]AdminCashierOrderView, 0, len(s.cashierOrders))
	for i := len(s.cashierOrders) - 1; i >= 0; i-- {
		order := s.cashierOrders[i]
		view := AdminCashierOrderView{
			ID:            order.ID,
			OrderNo:       order.OrderNo,
			StoreID:       order.StoreID,
			StoreName:     order.StoreName,
			Status:        normalizeCashierStatus(order.Status),
			CustomerName:  order.CustomerName,
			CustomerPhone: order.CustomerPhone,
			PaymentMethod: order.PaymentMethod,
			TotalAmount:   order.TotalAmount,
			PaidAmount:    order.PaidAmount,
			ItemCount:     len(order.Items),
			ItemSummary:   buildCashierItemSummary(order.Items),
			Items:         append([]CashierOrderLine(nil), order.Items...),
			CreatedBy:     order.CreatedBy,
			CreatedAt:     order.CreatedAt.Format("2006-01-02 15:04"),
			Remark:        order.Remark,
			VoidReason:    strings.TrimSpace(order.VoidReason),
			VoidedBy:      strings.TrimSpace(order.VoidedBy),
		}
		if order.VoidedAt != nil {
			view.VoidedAt = order.VoidedAt.Format("2006-01-02 15:04")
		}
		items = append(items, view)
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
			ItemSummary:     buildRecycleItemSummary(order.Items),
			Items:           append([]RecycleItem(nil), order.Items...),
			Attachments:     sanitizeAttachmentAssetsForAdmin(order.Attachments),
			CreatedBy:       order.CreatedBy,
			CreatedAt:       order.CreatedAt.Format("2006-01-02 15:04"),
			Remark:          order.Remark,
			CancelReason:    strings.TrimSpace(order.CancelReason),
			CancelledBy:     strings.TrimSpace(order.CancelledBy),
		}
		if len(order.Attachments) > 0 {
			view.PhotoCount = len(order.Attachments)
		}
		if order.ConfirmedAt != nil {
			view.ConfirmedAt = order.ConfirmedAt.Format("2006-01-02 15:04")
		}
		if order.CancelledAt != nil {
			view.CancelledAt = order.CancelledAt.Format("2006-01-02 15:04")
		}
		items = append(items, view)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items
}

func sanitizeAttachmentAssetsForAdmin(items []AttachmentAsset) []AttachmentAsset {
	result := make([]AttachmentAsset, 0, len(items))
	for _, item := range items {
		previewURL := strings.TrimSpace(item.ThumbnailURL)
		if previewURL == "" {
			previewURL = strings.TrimSpace(item.PublicURL)
		}
		item.PreviewURL = previewURL
		item.HasPreview = isAttachmentPreviewURLUsable(previewURL)
		if !item.HasPreview {
			item.PreviewURL = ""
		}
		result = append(result, item)
	}
	return result
}

func isAttachmentPreviewURLUsable(raw string) bool {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return false
	}
	blocked := []string{
		"mock.example.com",
		"example.com/assets/recycle",
		"pending-upload.local",
		"/smoke/",
		"/test/",
	}
	for _, token := range blocked {
		if strings.Contains(value, token) {
			return false
		}
	}
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
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
	if strings.TrimSpace(s.settings.Storage.StatusDescription) != "" {
		profile.OSSStatus = s.settings.Storage.StatusDescription
	}
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

func (s *MockStore) storeHasAnotherActiveManagerLocked(storeID string, accountID string) bool {
	for _, account := range s.usersByID {
		if account.ID == accountID || account.Status == "disabled" || normalizeAdminRoleKey(account.RoleCode) != "shop_manager" {
			continue
		}
		if containsID(account.StoreIDs, storeID) {
			return true
		}
	}
	return false
}

func productVisibleInStores(product AdminProductRecord, visibleStoreIDs []string, visibleStoreNames []string) bool {
	if hasIntersect(product.StoreIDs, visibleStoreIDs) {
		return true
	}
	if len(product.StoreIDs) > 0 {
		return false
	}
	for _, storeName := range product.StoreNames {
		for _, visible := range visibleStoreNames {
			if storeName == visible {
				return true
			}
		}
	}
	return false
}

func buildCashierItemSummary(items []CashierOrderLine) string {
	if len(items) == 0 {
		return "暂无商品明细"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = "未命名商品"
		}
		if item.Quantity > 1 {
			parts = append(parts, fmt.Sprintf("%s x%d", name, item.Quantity))
			continue
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, "、")
}

func buildRecycleItemSummary(items []RecycleItem) string {
	if len(items) == 0 {
		return "未填写品类"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		category := strings.TrimSpace(item.Category)
		if category == "" {
			category = "未分类"
		}
		purity := strings.TrimSpace(item.Purity)
		weight := item.WeightGram
		summary := category
		if purity != "" {
			summary += " · " + purity
		}
		if weight > 0 {
			summary += fmt.Sprintf(" · %.2fg", weight)
		}
		parts = append(parts, summary)
	}
	return strings.Join(parts, "、")
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

func sanitizeAdminAbilities(abilities []AdminAbilityCode) []AdminAbilityCode {
	allowed := map[AdminAbilityCode]struct{}{
		"dashboard.view":       {},
		"store.view":           {},
		"store.manage":         {},
		"user.view":            {},
		"user.manage":          {},
		"role.view":            {},
		"role.manage":          {},
		"product.view":         {},
		"product.manage":       {},
		"order.view":           {},
		"recycle.view":         {},
		"system.config.view":   {},
		"system.config.manage": {},
	}
	items := make([]AdminAbilityCode, 0, len(abilities))
	seen := make(map[AdminAbilityCode]struct{}, len(abilities))
	for _, ability := range abilities {
		if _, ok := allowed[ability]; !ok {
			continue
		}
		if _, ok := seen[ability]; ok {
			continue
		}
		seen[ability] = struct{}{}
		items = append(items, ability)
	}
	return items
}

func sanitizeAdminRoles(roles []AdminRoleTemplate) []AdminRoleTemplate {
	items := make([]AdminRoleTemplate, 0, len(roles))
	for _, role := range roles {
		role.Abilities = sanitizeAdminAbilities(role.Abilities)
		items = append(items, role)
	}
	return items
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
	case "boss":
		return "老板"
	case "shop_manager":
		return "店长"
	default:
		return "店长"
	}
}

func baseRoleCodeFromAdmin(roleKey string) string {
	switch normalizeAdminRoleKey(roleKey) {
	case "boss":
		return "owner"
	case "shop_manager":
		return "manager"
	default:
		return "manager"
	}
}

func baseDataScopeForRole(roleKey string) string {
	switch normalizeAdminRoleKey(roleKey) {
	case "boss":
		return "org_all"
	case "shop_manager":
		return "assigned_stores"
	default:
		return "assigned_stores"
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

func normalizeMemberStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "active":
		return "active"
	case "follow_up":
		return "follow_up"
	case "sleeping":
		return "sleeping"
	case "inactive":
		return "sleeping"
	case "disabled":
		return "disabled"
	default:
		return "active"
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
	contactPhone := store.ContactPhone
	if contactPhone == "" {
		contactPhone = s.settings.Brand.ServicePhone
	}
	businessHours := store.BusinessHours
	if businessHours == "" {
		businessHours = "10:00 - 22:00"
	}
	record := AdminStoreRecord{
		ID:               store.ID,
		Code:             store.Code,
		Name:             store.Name,
		ManagerName:      store.Manager,
		City:             store.City,
		Address:          store.Address,
		ContactPhone:     contactPhone,
		BusinessHours:    businessHours,
		Longitude:        store.Longitude,
		Latitude:         store.Latitude,
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
		ID:            account.ID,
		Name:          account.DisplayName,
		Account:       account.Username,
		Phone:         account.Phone,
		MiniAppOpenID: account.WechatOpenID,
		RoleKey:       roleKey,
		RoleName:      roleNameForKey(roleKey),
		DataScope:     mapAdminDataScope(account.DataScope),
		StoreIDs:      nonNilStrings(account.StoreIDs),
		StoreNames:    storeNames,
		Status:        mapUserStatus(account.Status),
		LastLoginAt:   time.Now().Format("2006-01-02 15:04"),
		Abilities:     sanitizeAdminAbilities(role.Abilities),
	}
}

func (s *MockStore) adminStoreNamesForProductLocked(storeIDs []string, fallback []string) []string {
	names := make([]string, 0, len(storeIDs))
	for _, storeID := range storeIDs {
		if store, ok := s.getStoreByIDLocked(storeID); ok {
			names = append(names, store.Name)
		}
	}
	if len(names) > 0 {
		return names
	}
	return cleanUniqueStrings(fallback)
}

func (s *MockStore) buildAdminProductRecordLocked(product CatalogProduct) AdminProductRecord {
	storeIDs := cleanUniqueStrings(product.StoreIDs)
	return AdminProductRecord{
		ID:          product.ID,
		Name:        product.Name,
		SKU:         product.SKU,
		Category:    product.Category,
		CategoryTab: product.CategoryTab,
		ImageURL:    product.ImageURL,
		Price:       product.RetailPrice,
		GramWeight:  product.GramWeight,
		Status:      product.Status,
		Inventory:   product.Inventory,
		StockStatus: product.StockStatus,
		StoreIDs:    storeIDs,
		StoreNames:  s.adminStoreNamesForProductLocked(storeIDs, product.Stores),
		Tags:        nonNilStrings(product.Tags),
	}
}

func (s *MockStore) resolveAdminProductStoresLocked(user UserAccount, storeIDs []string, storeNames []string) ([]string, []string, error) {
	resolvedIDs := cleanUniqueStrings(storeIDs)
	if len(resolvedIDs) == 0 {
		for _, name := range cleanUniqueStrings(storeNames) {
			matched := false
			for _, store := range s.stores {
				if store.Name == name {
					resolvedIDs = append(resolvedIDs, store.ID)
					matched = true
					break
				}
			}
			if !matched {
				return nil, nil, errUnauthorizedStore
			}
		}
		resolvedIDs = cleanUniqueStrings(resolvedIDs)
	}

	resolvedNames := make([]string, 0, len(resolvedIDs))
	for _, storeID := range resolvedIDs {
		if user.DataScope != "org_all" && !containsID(user.StoreIDs, storeID) {
			return nil, nil, errUnauthorizedStore
		}
		store, ok := s.getStoreByIDLocked(storeID)
		if !ok {
			return nil, nil, errUnauthorizedStore
		}
		resolvedNames = append(resolvedNames, store.Name)
	}
	return resolvedIDs, resolvedNames, nil
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
		if update.Description != "" {
			item.Description = update.Description
		}
		if item.Key != "boss" && update.DataScope != "" {
			item.DataScope = update.DataScope
		}
		if item.Key == "boss" {
			item.DataScope = "all_stores"
			item.Abilities = allAdminAbilities(newAdminAbilityGroups())
		} else {
			item.Abilities = sanitizeAdminAbilities(update.Abilities)
		}
		s.adminRoles[index] = item
		s.refreshUsersForAdminRoleLocked(item)
		if s.persistence != nil {
			_ = s.persistAdminRolesLocked()
			_ = s.persistUsersLocked()
		}
		s.appendAuditLogLocked("岗位权限", "保存岗位设置", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已更新", item.Name))
		return item, nil
	}
	return AdminRoleTemplate{}, errAdminRoleNotFound
}

func (s *MockStore) refreshUsersForAdminRoleLocked(role AdminRoleTemplate) {
	for userID, account := range s.usersByID {
		if normalizeAdminRoleKey(account.RoleCode) != role.Key {
			continue
		}
		account.RoleName = role.Name
		if role.Key == "boss" {
			account.DataScope = "org_all"
			account.StoreIDs = allStoreIDsLocked(s.stores)
		} else if strings.TrimSpace(role.DataScope) != "" {
			account.DataScope = baseDataScopeFromAdmin(role.DataScope)
		}
		account.Permissions = permissionsForRoleLocked(role.Key, s.roles)
		s.usersByID[userID] = account
		s.usersByUsername[account.Username] = account
	}
}

func allStoreIDsLocked(stores []StoreInfo) []string {
	ids := make([]string, 0, len(stores))
	for _, store := range stores {
		ids = append(ids, store.ID)
	}
	return ids
}

func (s *MockStore) listAdminRoles(_ UserAccount) []AdminRoleTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := sanitizeAdminRoles(s.adminRoles)
	memberCounts := make(map[string]int)
	for _, account := range s.usersByID {
		memberCounts[normalizeAdminRoleKey(account.RoleCode)]++
	}
	for index := range items {
		items[index].MemberCount = memberCounts[items[index].Key]
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *MockStore) listAdminStores(user UserAccount, query AdminListQuery) AdminPagedItems[AdminStoreRecord] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]AdminStoreRecord, 0)
	for _, item := range s.buildDynamicAdminStoresLocked() {
		if user.DataScope != "org_all" && !containsID(user.StoreIDs, item.ID) {
			continue
		}
		if strings.TrimSpace(query.StoreID) != "" && item.ID != query.StoreID {
			continue
		}
		if strings.TrimSpace(query.Status) != "" && item.Status != query.Status {
			continue
		}
		if !matchesAdminKeyword(query.Keyword, item.Code, item.Name, item.ManagerName, item.City, item.Address, item.ContactPhone) {
			continue
		}
		items = append(items, item)
	}
	return paginateAdminItems(items, query)
}

func (s *MockStore) listAdminUsers(user UserAccount, query AdminListQuery) AdminPagedItems[AdminUserAccount] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]AdminUserAccount, 0)
	for _, item := range s.buildDynamicAdminUsersLocked() {
		if user.DataScope != "org_all" && item.ID != user.ID && !hasIntersect(item.StoreIDs, user.StoreIDs) {
			continue
		}
		if strings.TrimSpace(query.StoreID) != "" && !containsID(item.StoreIDs, query.StoreID) {
			continue
		}
		if strings.TrimSpace(query.Status) != "" && item.Status != query.Status {
			continue
		}
		if !matchesAdminKeyword(query.Keyword, item.Name, item.Account, item.Phone, item.RoleName, strings.Join(item.StoreNames, " ")) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastLoginAt > items[j].LastLoginAt
	})
	return paginateAdminItems(items, query)
}

func (s *MockStore) listAdminProducts(user UserAccount, query AdminListQuery) AdminPagedItems[AdminProductRecord] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storeID := strings.TrimSpace(query.StoreID)
	storeName := ""
	if storeID != "" {
		if store, ok := s.getStoreByIDLocked(storeID); ok {
			storeName = store.Name
		}
	}
	visibleStores := make(map[string]struct{}, len(user.StoreIDs))
	for _, storeID := range user.StoreIDs {
		if store, ok := s.getStoreByIDLocked(storeID); ok {
			visibleStores[store.Name] = struct{}{}
		}
	}

	items := make([]AdminProductRecord, 0)
	for _, item := range s.buildDynamicAdminProductsLocked() {
		if user.DataScope != "org_all" {
			visible := hasIntersect(item.StoreIDs, user.StoreIDs)
			if !visible && len(item.StoreIDs) == 0 {
				for _, name := range item.StoreNames {
					if _, ok := visibleStores[name]; ok {
						visible = true
						break
					}
				}
			}
			if !visible {
				continue
			}
		}
		if storeID != "" && !containsID(item.StoreIDs, storeID) {
			if len(item.StoreIDs) > 0 || storeName == "" || !containsText(item.StoreNames, storeName) {
				continue
			}
		}
		if strings.TrimSpace(query.Status) != "" && item.Status != query.Status {
			continue
		}
		if !matchesAdminKeyword(query.Keyword, item.Name, item.SKU, item.Category, item.CategoryTab, strings.Join(item.Tags, " "), strings.Join(item.StoreNames, " ")) {
			continue
		}
		items = append(items, item)
	}
	return paginateAdminItems(items, query)
}

func (s *MockStore) listAdminMembers(user UserAccount, query AdminListQuery) AdminPagedItems[MemberProfile] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]MemberProfile, 0)
	for _, item := range s.members {
		if !s.canAccessStore(user, item.StoreID) {
			continue
		}
		if strings.TrimSpace(query.StoreID) != "" && item.StoreID != query.StoreID {
			continue
		}
		if strings.TrimSpace(query.Status) != "" && item.Status != query.Status {
			continue
		}
		if !matchesAdminKeyword(query.Keyword, item.Name, item.Phone, item.Level, item.StoreName, item.ManagerName, item.Notes) {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastVisitAt.After(items[j].LastVisitAt)
	})
	return paginateAdminItems(items, query)
}

func (s *MockStore) listAdminCashierOrders(user UserAccount, query AdminListQuery) AdminPagedItems[AdminCashierOrderView] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]AdminCashierOrderView, 0)
	for _, item := range s.buildDynamicCashierOrderViewsLocked() {
		if !s.canAccessStore(user, item.StoreID) {
			continue
		}
		if strings.TrimSpace(query.StoreID) != "" && item.StoreID != query.StoreID {
			continue
		}
		if strings.TrimSpace(query.Status) != "" && item.Status != query.Status {
			continue
		}
		if !matchesAdminDateRange(item.CreatedAt, query.DateFrom, query.DateTo) {
			continue
		}
		if !matchesAdminKeyword(query.Keyword, item.OrderNo, item.CustomerName, item.CustomerPhone, item.StoreName, item.ItemSummary, item.Remark) {
			continue
		}
		items = append(items, item)
	}
	return paginateAdminItems(items, query)
}

func (s *MockStore) getAdminCashierOrder(user UserAccount, orderID string) (AdminCashierOrderView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.buildDynamicCashierOrderViewsLocked() {
		if item.ID != orderID && item.OrderNo != orderID {
			continue
		}
		if !s.canAccessStore(user, item.StoreID) {
			return AdminCashierOrderView{}, errUnauthorizedStore
		}
		return item, nil
	}
	return AdminCashierOrderView{}, errCashierNotFound
}

func (s *MockStore) voidAdminCashierOrder(user UserAccount, orderID, reason string) (AdminCashierOrderView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, order := range s.cashierOrders {
		if order.ID != orderID && order.OrderNo != orderID {
			continue
		}
		if !s.canAccessStore(user, order.StoreID) {
			return AdminCashierOrderView{}, errUnauthorizedStore
		}
		if normalizeCashierStatus(order.Status) == "refunded" {
			return AdminCashierOrderView{}, errAdminOrderConflict
		}
		now := time.Now()
		order.Status = "refunded"
		order.VoidReason = strings.TrimSpace(reason)
		order.VoidedBy = user.DisplayName
		order.VoidedAt = &now
		s.cashierOrders[index] = order
		if s.persistence != nil {
			_ = s.persistence.saveCashierOrder(context.Background(), order)
		}
		s.appendAuditLogLocked("收银单", "作废收银单", user.DisplayName, "warning", "high", fmt.Sprintf("%s 已作废，原因：%s", order.OrderNo, order.VoidReason))
		for _, item := range s.buildDynamicCashierOrderViewsLocked() {
			if item.ID == order.ID {
				return item, nil
			}
		}
		return AdminCashierOrderView{}, errCashierNotFound
	}
	return AdminCashierOrderView{}, errCashierNotFound
}

func (s *MockStore) listAdminRecycleOrders(user UserAccount, query AdminListQuery) AdminPagedItems[AdminRecycleOrderView] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]AdminRecycleOrderView, 0)
	for _, item := range s.buildDynamicRecycleOrderViewsLocked() {
		if !s.canAccessStore(user, item.StoreID) {
			continue
		}
		if strings.TrimSpace(query.StoreID) != "" && item.StoreID != query.StoreID {
			continue
		}
		if strings.TrimSpace(query.Status) != "" && item.Status != query.Status {
			continue
		}
		if !matchesAdminDateRange(item.CreatedAt, query.DateFrom, query.DateTo) {
			continue
		}
		if !matchesAdminKeyword(query.Keyword, item.OrderNo, item.CustomerName, item.CustomerPhone, item.StoreName, item.ItemSummary, item.Remark) {
			continue
		}
		items = append(items, item)
	}
	return paginateAdminItems(items, query)
}

func (s *MockStore) getAdminRecycleOrder(user UserAccount, orderID string) (AdminRecycleOrderView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.buildDynamicRecycleOrderViewsLocked() {
		if item.ID != orderID && item.OrderNo != orderID {
			continue
		}
		if !s.canAccessStore(user, item.StoreID) {
			return AdminRecycleOrderView{}, errUnauthorizedStore
		}
		return item, nil
	}
	return AdminRecycleOrderView{}, errRecycleNotFound
}

func (s *MockStore) cancelAdminRecycleOrder(user UserAccount, orderID, reason string) (AdminRecycleOrderView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.recycleOrders[orderID]
	if !ok {
		for _, item := range s.recycleOrders {
			if item.OrderNo == orderID {
				order = item
				ok = true
				orderID = item.ID
				break
			}
		}
	}
	if !ok {
		return AdminRecycleOrderView{}, errRecycleNotFound
	}
	if !s.canAccessStore(user, order.StoreID) {
		return AdminRecycleOrderView{}, errUnauthorizedStore
	}
	if order.Status != "draft" {
		return AdminRecycleOrderView{}, errAdminOrderConflict
	}
	now := time.Now()
	order.Status = "cancelled"
	order.CancelReason = strings.TrimSpace(reason)
	order.CancelledBy = user.DisplayName
	order.CancelledAt = &now
	s.recycleOrders[orderID] = order
	if s.persistence != nil {
		_ = s.persistence.saveRecycleOrder(context.Background(), order)
	}
	s.appendAuditLogLocked("回收单", "取消回收单", user.DisplayName, "warning", "high", fmt.Sprintf("%s 已取消，原因：%s", order.OrderNo, order.CancelReason))
	for _, item := range s.buildDynamicRecycleOrderViewsLocked() {
		if item.ID == orderID {
			return item, nil
		}
	}
	return AdminRecycleOrderView{}, errRecycleNotFound
}

func (s *MockStore) listAdminOrderRows(user UserAccount, query AdminListQuery) AdminPagedItems[AdminOrderRow] {
	cashier := s.listAdminCashierOrders(user, AdminListQuery{
		Page:     1,
		PageSize: 100000,
		StoreID:  query.StoreID,
		Status:   query.Status,
		DateFrom: query.DateFrom,
		DateTo:   query.DateTo,
		Keyword:  query.Keyword,
	})
	recycle := s.listAdminRecycleOrders(user, AdminListQuery{
		Page:     1,
		PageSize: 100000,
		StoreID:  query.StoreID,
		Status:   query.Status,
		DateFrom: query.DateFrom,
		DateTo:   query.DateTo,
		Keyword:  query.Keyword,
	})

	items := make([]AdminOrderRow, 0, len(cashier.Items)+len(recycle.Items))
	for _, item := range cashier.Items {
		items = append(items, AdminOrderRow{
			ID:            item.ID,
			Type:          "cashier",
			OrderNo:       item.OrderNo,
			StoreID:       item.StoreID,
			StoreName:     item.StoreName,
			Status:        item.Status,
			CustomerName:  item.CustomerName,
			CustomerPhone: item.CustomerPhone,
			Amount:        item.TotalAmount,
			ItemSummary:   item.ItemSummary,
			CreatedBy:     item.CreatedBy,
			CreatedAt:     item.CreatedAt,
		})
	}
	for _, item := range recycle.Items {
		amount := item.ConfirmedAmount
		if amount <= 0 {
			amount = item.EstimatedAmount
		}
		items = append(items, AdminOrderRow{
			ID:            item.ID,
			Type:          "recycle",
			OrderNo:       item.OrderNo,
			StoreID:       item.StoreID,
			StoreName:     item.StoreName,
			Status:        item.Status,
			CustomerName:  item.CustomerName,
			CustomerPhone: item.CustomerPhone,
			Amount:        amount,
			ItemSummary:   item.ItemSummary,
			CreatedBy:     item.CreatedBy,
			CreatedAt:     item.CreatedAt,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return paginateAdminItems(items, query)
}

func paginateAdminItems[T any](items []T, query AdminListQuery) AdminPagedItems[T] {
	total := len(items)
	page := query.Page
	pageSize := query.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	paged := make([]T, 0, end-start)
	if start < end {
		paged = append(paged, items[start:end]...)
	}
	return AdminPagedItems[T]{
		Items:    paged,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

func matchesAdminKeyword(keyword string, fields ...string) bool {
	needle := strings.TrimSpace(strings.ToLower(keyword))
	if needle == "" {
		return true
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(strings.TrimSpace(field)), needle) {
			return true
		}
	}
	return false
}

func containsText(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func matchesAdminDateRange(createdAt, dateFrom, dateTo string) bool {
	if strings.TrimSpace(dateFrom) == "" && strings.TrimSpace(dateTo) == "" {
		return true
	}
	createdDate := createdAt
	if len(createdDate) >= 10 {
		createdDate = createdDate[:10]
	}
	if strings.TrimSpace(dateFrom) != "" && createdDate < strings.TrimSpace(dateFrom) {
		return false
	}
	if strings.TrimSpace(dateTo) != "" && createdDate > strings.TrimSpace(dateTo) {
		return false
	}
	return true
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
		if strings.TrimSpace(update.ContactPhone) != "" {
			store.ContactPhone = strings.TrimSpace(update.ContactPhone)
		}
		if strings.TrimSpace(update.BusinessHours) != "" {
			store.BusinessHours = strings.TrimSpace(update.BusinessHours)
		}
		if update.Longitude != 0 {
			store.Longitude = update.Longitude
		}
		if update.Latitude != 0 {
			store.Latitude = update.Latitude
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
		if user.DataScope != "org_all" && account.ID != user.ID {
			return AdminUserAccount{}, errUnauthorizedStore
		}
		if strings.TrimSpace(update.Name) != "" {
			account.DisplayName = strings.TrimSpace(update.Name)
		}
		if strings.TrimSpace(update.Account) != "" {
			delete(s.usersByUsername, account.Username)
			account.Username = strings.TrimSpace(update.Account)
		}
		if strings.TrimSpace(update.Phone) != "" {
			account.Phone = strings.TrimSpace(update.Phone)
		}
		if strings.TrimSpace(update.Password) != "" {
			account.Password = strings.TrimSpace(update.Password)
		}
		if strings.TrimSpace(update.MiniAppOpenID) != "" {
			account.WechatOpenID = strings.TrimSpace(update.MiniAppOpenID)
		}
		roleKey := normalizeAdminRoleKey(account.RoleCode)
		if strings.TrimSpace(update.RoleKey) != "" {
			roleKey = normalizeAdminRoleKey(update.RoleKey)
			account.RoleCode = baseRoleCodeFromAdmin(roleKey)
			account.RoleName = roleNameForKey(roleKey)
			account.DataScope = baseDataScopeForRole(roleKey)
			account.Permissions = permissionsForRoleLocked(roleKey, s.roles)
		}
		account.RoleName = roleNameForKey(roleKey)
		account.DataScope = baseDataScopeForRole(roleKey)
		if update.StoreIDs != nil {
			if user.DataScope != "org_all" {
				for _, storeID := range update.StoreIDs {
					if !containsID(user.StoreIDs, storeID) {
						return AdminUserAccount{}, errUnauthorizedStore
					}
				}
			}
			account.StoreIDs = cleanUniqueStrings(update.StoreIDs)
		}
		if strings.TrimSpace(update.Status) != "" {
			account.Status = baseUserStatusFromAdmin(update.Status)
		}
		if roleKey == "boss" {
			account.StoreIDs = allStoreIDsLocked(s.stores)
		} else {
			account.StoreIDs = cleanUniqueStrings(account.StoreIDs)
			if len(account.StoreIDs) != 1 {
				return AdminUserAccount{}, errUnauthorizedStore
			}
			if account.Status != "disabled" && s.storeHasAnotherActiveManagerLocked(account.StoreIDs[0], account.ID) {
				return AdminUserAccount{}, errUnauthorizedStore
			}
		}
		if strings.TrimSpace(account.Phone) != "" && strings.TrimSpace(account.Username) == "" {
			account.Username = strings.TrimSpace(account.Phone)
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
		if user.DataScope != "org_all" {
			visibleStoreNames := s.adminStoreNamesForProductLocked(user.StoreIDs, nil)
			if !productVisibleInStores(s.buildAdminProductRecordLocked(product), user.StoreIDs, visibleStoreNames) {
				return AdminProductRecord{}, errUnauthorizedStore
			}
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
		if strings.TrimSpace(update.CategoryTab) != "" {
			product.CategoryTab = strings.TrimSpace(update.CategoryTab)
		}
		if strings.TrimSpace(update.ImageURL) != "" {
			product.ImageURL = strings.TrimSpace(update.ImageURL)
		}
		if update.Price >= 0 {
			product.RetailPrice = update.Price
		}
		if update.GramWeight >= 0 {
			product.GramWeight = update.GramWeight
		}
		if update.Inventory >= 0 {
			product.Inventory = update.Inventory
		}
		if strings.TrimSpace(update.StockStatus) != "" {
			product.StockStatus = strings.TrimSpace(update.StockStatus)
		}
		if strings.TrimSpace(update.Status) != "" {
			product.Status = strings.TrimSpace(update.Status)
		}
		if update.StoreIDs != nil || update.StoreNames != nil {
			storeIDs, storeNames, err := s.resolveAdminProductStoresLocked(user, update.StoreIDs, update.StoreNames)
			if err != nil {
				return AdminProductRecord{}, err
			}
			product.StoreIDs = storeIDs
			product.Stores = storeNames
		}
		if product.Status != "disabled" && len(product.StoreIDs) == 0 {
			return AdminProductRecord{}, errUnauthorizedStore
		}
		if update.Tags != nil {
			product.Tags = cleanUniqueStrings(update.Tags)
		}
		s.catalogProducts[index] = product
		if s.persistence != nil {
			_ = s.persistProductsLocked()
		}
		record := s.buildAdminProductRecordLocked(product)
		s.appendAuditLogLocked("商品管理", "更新商品资料", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已更新", record.Name))
		return record, nil
	}
	return AdminProductRecord{}, errAdminProductNotFound
}

func (s *MockStore) updateAdminMember(user UserAccount, memberID string, update MemberProfile) (MemberProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, member := range s.members {
		if member.ID != memberID {
			continue
		}
		if !s.canAccessStore(user, member.StoreID) {
			return MemberProfile{}, errUnauthorizedStore
		}
		if strings.TrimSpace(update.StoreID) != "" && user.DataScope != "org_all" && !s.canAccessStore(user, update.StoreID) {
			return MemberProfile{}, errUnauthorizedStore
		}
		if strings.TrimSpace(update.StoreID) != "" {
			member.StoreID = strings.TrimSpace(update.StoreID)
			if store, ok := s.getStoreByIDLocked(member.StoreID); ok {
				member.StoreName = store.Name
			}
		}
		if strings.TrimSpace(update.Name) != "" {
			member.Name = strings.TrimSpace(update.Name)
		}
		if strings.TrimSpace(update.Phone) != "" {
			member.Phone = strings.TrimSpace(update.Phone)
		}
		if strings.TrimSpace(update.Level) != "" {
			member.Level = strings.TrimSpace(update.Level)
		}
		if strings.TrimSpace(update.Status) != "" {
			member.Status = normalizeMemberStatus(update.Status)
		}
		if update.TotalOrders >= 0 {
			member.TotalOrders = update.TotalOrders
		}
		if update.TotalRecycleAmount >= 0 {
			member.TotalRecycleAmount = update.TotalRecycleAmount
		}
		if strings.TrimSpace(update.PreferredPurity) != "" {
			member.PreferredPurity = strings.TrimSpace(update.PreferredPurity)
		}
		if strings.TrimSpace(update.SourceChannel) != "" {
			member.SourceChannel = strings.TrimSpace(update.SourceChannel)
		}
		if strings.TrimSpace(update.ManagerName) != "" {
			member.ManagerName = strings.TrimSpace(update.ManagerName)
		}
		member.IDVerified = update.IDVerified
		if update.Tags != nil {
			member.Tags = append([]string(nil), update.Tags...)
		}
		if strings.TrimSpace(update.Notes) != "" || update.Notes == "" {
			member.Notes = strings.TrimSpace(update.Notes)
		}
		if !update.LastVisitAt.IsZero() {
			member.LastVisitAt = update.LastVisitAt
		}
		s.members[index] = member
		if s.persistence != nil {
			_ = s.persistMembersLocked()
		}
		s.appendAuditLogLocked("会员档案", "更新会员资料", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已更新", member.Name))
		return member, nil
	}
	return MemberProfile{}, errAdminMemberNotFound
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
		OrgID:     "org-gold-v1",
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

	if user.DataScope != "org_all" && len(user.StoreIDs) == 0 {
		return AdminUserAccount{}, errUnauthorizedStore
	}

	storeIDs := []string{}
	if user.DataScope != "org_all" {
		storeIDs = append(storeIDs, user.StoreIDs[0])
	}

	account := UserAccount{
		ID:          fmt.Sprintf("user-new-%03d", len(s.usersByID)+1),
		OrgID:       "org-gold-v1",
		Username:    fmt.Sprintf("new.user.%03d", len(s.usersByID)+1),
		DisplayName: fmt.Sprintf("新店长 %d", len(s.usersByID)+1),
		Phone:       "",
		Password:    "ChangeMe123!",
		RoleCode:    "manager",
		RoleName:    "店长",
		DataScope:   "assigned_stores",
		StoreIDs:    storeIDs,
		Permissions: permissionsForRoleLocked("shop_manager", s.roles),
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

	if user.DataScope != "org_all" && len(user.StoreIDs) == 0 {
		return AdminProductRecord{}, errUnauthorizedStore
	}

	storeIDs := []string{}
	storeNames := []string{}
	if user.DataScope != "org_all" {
		for _, storeID := range user.StoreIDs {
			if store, ok := s.getStoreByIDLocked(storeID); ok {
				storeIDs = append(storeIDs, store.ID)
				storeNames = append(storeNames, store.Name)
			}
		}
		if len(storeIDs) == 0 {
			return AdminProductRecord{}, errUnauthorizedStore
		}
	}

	product := CatalogProduct{
		ID:               fmt.Sprintf("prd-new-%03d", len(s.catalogProducts)+1),
		OrgID:            "org-gold-v1",
		Name:             fmt.Sprintf("新商品 %d", len(s.catalogProducts)+1),
		SKU:              fmt.Sprintf("NEW-SKU-%03d", len(s.catalogProducts)+1),
		Category:         "待分类",
		CategoryTab:      "其他",
		ImageURL:         "/assets/ui/document.png",
		Status:           "draft",
		StockStatus:      "review",
		StoreIDs:         storeIDs,
		Stores:           storeNames,
		Tags:             []string{"新建", "待完善"},
		RecommendedScene: "待完善场景",
		QuoteLeadTime:    "待设置",
	}
	s.catalogProducts = append([]CatalogProduct{product}, s.catalogProducts...)
	if s.persistence != nil {
		_ = s.persistProductsLocked()
	}
	record := s.buildAdminProductRecordLocked(product)
	s.appendAuditLogLocked("商品管理", "新建商品", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已创建", record.Name))
	return record, nil
}

func (s *MockStore) createAdminMember(user UserAccount) (MemberProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	storeID := ""
	storeName := ""
	switch {
	case user.DataScope == "org_all" && len(s.stores) > 0:
		storeID = s.stores[0].ID
		storeName = s.stores[0].Name
	case len(user.StoreIDs) > 0:
		storeID = user.StoreIDs[0]
		if store, ok := s.getStoreByIDLocked(storeID); ok {
			storeName = store.Name
		}
	default:
		return MemberProfile{}, errUnauthorizedStore
	}

	member := MemberProfile{
		ID:                 fmt.Sprintf("member-new-%03d", len(s.members)+1),
		OrgID:              user.OrgID,
		StoreID:            storeID,
		StoreName:          storeName,
		Name:               fmt.Sprintf("新会员 %d", len(s.members)+1),
		Phone:              "",
		Level:              "普通会员",
		Status:             "active",
		TotalOrders:        0,
		TotalRecycleAmount: 0,
		LastVisitAt:        time.Now(),
		PreferredPurity:    "",
		SourceChannel:      "后台新增",
		ManagerName:        user.DisplayName,
		IDVerified:         false,
		Tags:               []string{"待完善"},
		Notes:              "",
	}
	s.members = append([]MemberProfile{member}, s.members...)
	if s.persistence != nil {
		_ = s.persistMembersLocked()
	}
	s.appendAuditLogLocked("会员档案", "新建会员", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已创建", member.Name))
	return member, nil
}

func (s *MockStore) disableAdminStore(user UserAccount, storeID string) (AdminStoreRecord, error) {
	return s.updateAdminStore(user, storeID, AdminStoreRecord{Status: "disabled"})
}

func (s *MockStore) disableAdminUser(user UserAccount, userID string) (AdminUserAccount, error) {
	return s.updateAdminUser(user, userID, AdminUserAccount{Status: "disabled"})
}

func (s *MockStore) disableAdminProduct(user UserAccount, productID string) (AdminProductRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, product := range s.catalogProducts {
		if product.ID != productID {
			continue
		}
		if user.DataScope != "org_all" {
			visibleStoreNames := s.adminStoreNamesForProductLocked(user.StoreIDs, nil)
			if !productVisibleInStores(s.buildAdminProductRecordLocked(product), user.StoreIDs, visibleStoreNames) {
				return AdminProductRecord{}, errUnauthorizedStore
			}
		}
		product.Status = "disabled"
		product.StockStatus = "disabled"
		s.catalogProducts[index] = product
		if s.persistence != nil {
			_ = s.persistProductsLocked()
		}
		record := s.buildAdminProductRecordLocked(product)
		s.appendAuditLogLocked("商品管理", "删除商品", user.DisplayName, "warning", "medium", fmt.Sprintf("%s 已删除", record.Name))
		return record, nil
	}
	return AdminProductRecord{}, errAdminProductNotFound
}

func (s *MockStore) disableAdminMember(user UserAccount, memberID string) (MemberProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index, member := range s.members {
		if member.ID != memberID {
			continue
		}
		if !s.canAccessStore(user, member.StoreID) {
			return MemberProfile{}, errUnauthorizedStore
		}
		member.Status = "disabled"
		s.members[index] = member
		if s.persistence != nil {
			_ = s.persistMembersLocked()
		}
		s.appendAuditLogLocked("会员档案", "删除会员", user.DisplayName, "warning", "medium", fmt.Sprintf("%s 已删除", member.Name))
		return member, nil
	}
	return MemberProfile{}, errAdminMemberNotFound
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
	if s.persistence != nil {
		_ = s.persistAdminAuditLogsLocked()
	}
}

func displayPhone(phone string) string {
	normalized := strings.TrimSpace(phone)
	if normalized == "" {
		return "待补充"
	}
	if len(normalized) == 11 {
		return normalized[:3] + "****" + normalized[7:]
	}
	return normalized
}

func buildAdminFixtures() (
	[]AdminRoleTemplate,
	[]AdminAbilityGroup,
	[]AdminStoreRecord,
	[]AdminUserAccount,
	[]AdminProductRecord,
	AdminPrintTemplate,
	[]AdminCashierOrderView,
	[]AdminRecycleOrderView,
	AdminSystemProfile,
	[]AdminAuditLogRecord,
	AdminTemplateInitPlan,
) {
	groups := newAdminAbilityGroups()
	allAbilities := allAdminAbilities(groups)
	roles := []AdminRoleTemplate{
		{ID: 1, Key: "boss", Name: "老板", Description: "可查看并维护全部门店的收银、回收、会员、商品和基础设置。", DataScope: "all_stores", MemberCount: 1, Locked: true, Abilities: allAbilities},
		{ID: 2, Key: "shop_manager", Name: "店长", Description: "只查看和维护自己门店的业务数据。", DataScope: "assigned_store", MemberCount: 3, Locked: true, Abilities: []AdminAbilityCode{"dashboard.view", "store.view", "product.view", "product.manage", "order.view", "recycle.view"}},
	}
	stores := []AdminStoreRecord{
		{ID: "store-sz-luohu", Code: "SZ-LH-01", Name: "罗湖旗舰店", ManagerName: "李店长", City: "深圳", Address: "罗湖区深南东路 1888 号 1F", ContactPhone: "0755-8899 1201", BusinessHours: "10:00 - 22:00", Status: "active", CashierDevices: 3, PendingTasks: 2, TodayAmount: 18260, TodayOrders: 16, LastSettlementAt: "2026-05-10 10:05", Tags: []string{"黄金回收", "收银台", "拍照留档"}},
		{ID: "store-sz-nanshan", Code: "SZ-NS-02", Name: "南山体验店", ManagerName: "王店长", City: "深圳", Address: "南山区科苑路 66 号 B1", ContactPhone: "0755-8899 1202", BusinessHours: "10:00 - 21:30", Status: "active", CashierDevices: 2, PendingTasks: 1, TodayAmount: 13680, TodayOrders: 11, LastSettlementAt: "2026-05-10 09:35", Tags: []string{"黄金回收", "体验店", "业务记账"}},
		{ID: "store-gz-tianhe", Code: "GZ-TH-01", Name: "广州天河店", ManagerName: "空缺", City: "广州", Address: "天河区体育西路 88 号", ContactPhone: "020-8899 1201", BusinessHours: "10:00 - 22:00", Status: "disabled", CashierDevices: 0, PendingTasks: 3, TodayAmount: 0, TodayOrders: 0, LastSettlementAt: "2026-05-06 21:30", Tags: []string{"暂停营业", "待重新授权"}},
	}
	users := []AdminUserAccount{
		{ID: "user-owner-001", Name: "廖总", Account: "boss", Phone: "151****5083", RoleKey: "boss", RoleName: "老板", DataScope: "all_stores", StoreIDs: []string{}, StoreNames: []string{"全部门店"}, Status: "enabled", LastLoginAt: "2026-06-04 16:50", Abilities: allAbilities},
		{ID: "user-owner-002", Name: "示例老板", Account: "boss.demo", Phone: "134****4944", RoleKey: "boss", RoleName: "老板", DataScope: "all_stores", StoreIDs: []string{}, StoreNames: []string{"全部门店"}, Status: "enabled", LastLoginAt: "2026-06-04 16:50", Abilities: allAbilities},
		{ID: "user-manager-001", Name: "李店长", Account: "manager.sz", Phone: "138****2108", RoleKey: "shop_manager", RoleName: "店长", DataScope: "assigned_store", StoreIDs: []string{"store-sz-luohu"}, StoreNames: []string{"罗湖旗舰店"}, Status: "enabled", LastLoginAt: "2026-05-10 08:42", Abilities: roles[1].Abilities},
		{ID: "user-manager-002", Name: "王店长", Account: "manager.gz", Phone: "138****2109", RoleKey: "shop_manager", RoleName: "店长", DataScope: "assigned_store", StoreIDs: []string{"store-sz-nanshan"}, StoreNames: []string{"南山体验店"}, Status: "enabled", LastLoginAt: "2026-05-09 20:16", Abilities: roles[1].Abilities},
	}
	products := []AdminProductRecord{
		{ID: "prd-001", Name: "足金项链", SKU: "GJ-XL-001", Category: "项链", Price: 3298, GramWeight: 8.6, Status: "active", StoreNames: []string{"罗湖旗舰店", "南山体验店"}, Tags: []string{"热卖", "足金"}},
		{ID: "prd-002", Name: "古法手镯", SKU: "GJ-SZ-018", Category: "手镯", Price: 5680, GramWeight: 15.2, Status: "active", StoreNames: []string{"罗湖旗舰店"}, Tags: []string{"回购高", "礼盒装"}},
		{ID: "prd-003", Name: "回收服务单", SKU: "REC-SRV-001", Category: "回收服务", Price: 0, GramWeight: 0, Status: "draft", StoreNames: []string{"罗湖旗舰店", "南山体验店", "广州天河店"}, Tags: []string{"系统单据", "非销售商品"}},
	}
	orders := []AdminCashierOrderView{
		{ID: "ord-001", OrderNo: "ORD202605100001", StoreID: "store-sz-luohu", StoreName: "罗湖旗舰店", Status: "paid", TotalAmount: 3298, ItemCount: 2, CreatedBy: "李店长", CreatedAt: "2026-05-10 10:08", Remark: "足金项链成交"},
		{ID: "ord-002", OrderNo: "ORD202605090145", StoreID: "store-sz-nanshan", StoreName: "南山体验店", Status: "pending", TotalAmount: 2680, ItemCount: 1, CreatedBy: "王店长", CreatedAt: "2026-05-09 19:42", Remark: "待主管复核"},
		{ID: "ord-003", OrderNo: "ORD202605070201", StoreID: "store-gz-tianhe", StoreName: "广州天河店", Status: "refunded", TotalAmount: 980, ItemCount: 1, CreatedBy: "系统迁移", CreatedAt: "2026-05-07 20:20", Remark: "停店前历史订单"},
	}
	recycleOrders := []AdminRecycleOrderView{
		{ID: "rec-001", OrderNo: "REC202605100014", StoreID: "store-sz-luohu", StoreName: "罗湖旗舰店", Status: "confirmed", CustomerName: "周女士", CustomerPhone: "134****2188", EstimatedAmount: 5800, ConfirmedAmount: 5680, PhotoCount: 3, CreatedBy: "李店长", CreatedAt: "2026-05-10 09:18", ConfirmedAt: "2026-05-10 09:32", Remark: "三张现场图已留存"},
		{ID: "rec-002", OrderNo: "REC202605090021", StoreID: "store-sz-nanshan", StoreName: "南山体验店", Status: "draft", CustomerName: "王先生", CustomerPhone: "135****1018", EstimatedAmount: 4320, ConfirmedAmount: 0, PhotoCount: 3, CreatedBy: "王店长", CreatedAt: "2026-05-09 16:18", Remark: "三张现场图待主管确认"},
	}
	systemProfile := AdminSystemProfile{
		BrandName: "金匠倌收银", ServicePhone: "400-888-2026", ReceiptTitle: "金匠倌收银门店小票",
		MinPhotoCount: 3, MaxPhotoCount: 3, RequireExactThree: false, RequireIDCheck: true,
		DomainName: "jinjiangguan.com", DomainStatus: "已启用",
		OSSStatus:     "回收留档图片已接入云端存储",
		AppIDStatus:   "小程序已接入",
		PrinterStatus: "可按门店设备配置小票打印",
	}
	auditLogs := []AdminAuditLogRecord{
		{ID: "log-001", Module: "基础设置", Action: "更新门店资料", OperatorName: "老板", Result: "success", RiskLevel: "low", Summary: "已更新品牌信息和客服电话。", CreatedAt: "2026-05-10 11:05"},
		{ID: "log-002", Module: "商品资料", Action: "调整商品价格", OperatorName: "老板", Result: "success", RiskLevel: "medium", Summary: "已更新重点商品价格和库存状态。", CreatedAt: "2026-05-10 10:26"},
		{ID: "log-003", Module: "回收规则", Action: "调整拍照规则", OperatorName: "系统", Result: "info", RiskLevel: "low", Summary: "当前要求至少上传 3 张现场照片。", CreatedAt: "2026-05-10 09:40"},
	}
	return roles, groups, stores, users, products, newAdminPrintTemplate(), orders, recycleOrders, systemProfile, auditLogs, newAdminTemplatePlan()
}

func normalizeAdminRoleKey(roleCode string) string {
	switch strings.ToLower(strings.TrimSpace(roleCode)) {
	case "boss", "owner":
		return "boss"
	case "shop_manager", "manager", "staff", "cashier", "clerk":
		return "shop_manager"
	default:
		return "shop_manager"
	}
}

func sanitizePersistedAdminRoles(roles []AdminRoleTemplate) ([]AdminRoleTemplate, bool) {
	changed := false
	merged := make(map[string]AdminRoleTemplate, len(roles))
	for _, role := range roles {
		normalizedKey := normalizeAdminRoleKey(role.Key)
		if role.Key != normalizedKey {
			role.Key = normalizedKey
			changed = true
		}
		if strings.Contains(role.Name, "模板") {
			role.Name = strings.ReplaceAll(role.Name, "模板", "")
			changed = true
		}
		switch role.Key {
		case "boss":
			if role.Name != "老板" {
				role.Name = "老板"
				changed = true
			}
			if role.Description != "可查看全部门店的收银、回收、会员、商品和基础设置。" {
				role.Description = "可查看全部门店的收银、回收、会员、商品和基础设置。"
				changed = true
			}
			role.DataScope = "all_stores"
		case "shop_manager":
			if role.Name != "店长" {
				role.Name = "店长"
				changed = true
			}
			if role.Description != "可查看本门店的业务数据，并维护商品和基础资料。" {
				role.Description = "可查看本门店的业务数据，并维护商品和基础资料。"
				changed = true
			}
			role.DataScope = "assigned_store"
		}
		sanitized := sanitizeAdminAbilities(role.Abilities)
		if len(sanitized) != len(role.Abilities) {
			changed = true
		}
		role.Abilities = sanitized
		if existing, ok := merged[role.Key]; ok {
			existing.Abilities = sanitizeAdminAbilities(append(existing.Abilities, role.Abilities...))
			existing.MemberCount += role.MemberCount
			merged[role.Key] = existing
			changed = true
			continue
		}
		merged[role.Key] = role
	}
	items := make([]AdminRoleTemplate, 0, 2)
	if role, ok := merged["boss"]; ok {
		role.ID = 1
		role.Key = "boss"
		role.Locked = true
		items = append(items, role)
	} else {
		items = append(items, AdminRoleTemplate{ID: 1, Key: "boss", Name: "老板", Description: "可查看全部门店的收银、回收、会员、商品和基础设置。", DataScope: "all_stores", Locked: true, Abilities: allAdminAbilities(newAdminAbilityGroups())})
		changed = true
	}
	if role, ok := merged["shop_manager"]; ok {
		role.ID = 2
		role.Key = "shop_manager"
		role.Locked = true
		items = append(items, role)
	} else {
		items = append(items, AdminRoleTemplate{ID: 2, Key: "shop_manager", Name: "店长", Description: "可查看本门店的业务数据，并维护商品和基础资料。", DataScope: "assigned_store", Locked: true, Abilities: []AdminAbilityCode{"dashboard.view", "store.view", "product.view", "product.manage", "order.view", "recycle.view"}})
		changed = true
	}
	return items, changed
}

func sanitizePersistedAdminSystemProfile(profile *AdminSystemProfile, settings *SystemSettings) bool {
	if profile == nil {
		return false
	}
	changed := false
	if strings.TrimSpace(profile.BrandName) == "" || strings.Contains(profile.BrandName, "金掌柜") || profile.BrandName == "金匠倌" {
		profile.BrandName = "金匠倌收银"
		changed = true
	}
	if strings.TrimSpace(profile.ReceiptTitle) == "" || strings.Contains(profile.ReceiptTitle, "黄金回收收银系统") || profile.ReceiptTitle == "金匠倌门店小票" {
		profile.ReceiptTitle = "金匠倌收银门店小票"
		changed = true
	}
	if strings.Contains(profile.DomainStatus, "实名认证审核中") || strings.Contains(profile.DomainStatus, "已购买") {
		profile.DomainStatus = "已启用"
		changed = true
	}
	if strings.Contains(profile.AppIDStatus, "待补充") {
		profile.AppIDStatus = "小程序已接入"
		changed = true
	}
	if strings.Contains(profile.PrinterStatus, "待确定小票机") || strings.Contains(profile.PrinterStatus, "彩色小票") {
		profile.PrinterStatus = "可按门店设备配置小票打印"
		changed = true
	}
	if settings != nil {
		if strings.TrimSpace(settings.Brand.Name) == "" || strings.Contains(settings.Brand.Name, "金掌柜") || settings.Brand.Name == "金匠倌" {
			settings.Brand.Name = "金匠倌收银"
			changed = true
		}
		if strings.TrimSpace(settings.Brand.ReceiptTitle) == "" || strings.Contains(settings.Brand.ReceiptTitle, "黄金回收收银系统") || settings.Brand.ReceiptTitle == "金匠倌门店小票" {
			settings.Brand.ReceiptTitle = "金匠倌收银门店小票"
			changed = true
		}
		if strings.Contains(settings.Storage.StatusDescription, "当前允许先登记外部图片链接") {
			settings.Storage.StatusDescription = "回收留档图片已接入云端存储"
			changed = true
		}
	}
	return changed
}
