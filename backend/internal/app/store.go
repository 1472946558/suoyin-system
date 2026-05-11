package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	errInvalidCredentials = errors.New("invalid credentials")
	errUnauthorizedStore  = errors.New("store not accessible")
	errCashierNotFound    = errors.New("cashier order not found")
	errRecycleNotFound    = errors.New("recycle order not found")
	errRecycleConflict    = errors.New("recycle order status conflict")
	errMemberNotFound     = errors.New("member not found")
	errProductNotFound    = errors.New("product not found")
)

type MockStore struct {
	mu                  sync.RWMutex
	usersByUsername     map[string]UserAccount
	usersByID           map[string]UserAccount
	roles               []RoleTemplate
	stores              []StoreInfo
	permissionCatalog   []PermissionGroup
	sessions            map[string]Session
	members             []MemberProfile
	catalogProducts     []CatalogProduct
	cashierOrders       []CashierOrder
	recycleOrders       map[string]RecycleOrder
	settings            SystemSettings
	cashierSeq          int
	recycleSeq          int
	adminRoles          []AdminRoleTemplate
	adminAbilityGroups  []AdminAbilityGroup
	adminStores         []AdminStoreRecord
	adminUsers          []AdminUserAccount
	adminProducts       []AdminProductRecord
	adminPaymentConfig  AdminPaymentConfig
	adminPrintTemplate  AdminPrintTemplate
	adminPaymentRecords []AdminPaymentRecord
	adminCashierOrders  []AdminCashierOrderView
	adminRecycleOrders  []AdminRecycleOrderView
	adminSystemProfile  AdminSystemProfile
	adminAuditLogs      []AdminAuditLogRecord
	adminTemplateInit   AdminTemplateInitPlan
}

func newMockStore() *MockStore {
	permissions := []PermissionGroup{
		{
			Code:        "dashboard",
			Name:        "首页看板",
			Description: "首页统计与经营总览",
			Items: []PermissionItem{
				{Code: "dashboard.summary.view", Name: "查看首页统计", Description: "查看门店经营汇总数据"},
			},
		},
		{
			Code:        "org",
			Name:        "组织与权限",
			Description: "门店、角色、能力目录",
			Items: []PermissionItem{
				{Code: "store.read", Name: "查看门店", Description: "查看可访问门店清单"},
				{Code: "role.read", Name: "查看角色模板", Description: "查看系统角色与默认能力"},
				{Code: "permission.catalog.read", Name: "查看能力目录", Description: "查看能力码与说明"},
			},
		},
		{
			Code:        "cashier",
			Name:        "收银订单",
			Description: "收银单列表与创建",
			Items: []PermissionItem{
				{Code: "cashier.order.read", Name: "查看收银订单", Description: "查看收银订单列表"},
				{Code: "cashier.order.create", Name: "创建收银订单", Description: "创建新的收银订单"},
			},
		},
		{
			Code:        "recycle",
			Name:        "黄金回收",
			Description: "回收单草稿与确认",
			Items: []PermissionItem{
				{Code: "recycle.order.read", Name: "查看回收单", Description: "查看回收单列表或详情"},
				{Code: "recycle.order.draft", Name: "创建回收草稿", Description: "录入黄金回收草稿"},
				{Code: "recycle.order.confirm", Name: "确认回收单", Description: "确认回收成交"},
			},
		},
		{
			Code:        "catalog",
			Name:        "会员与商品",
			Description: "会员档案和商品模板",
			Items: []PermissionItem{
				{Code: "member.read", Name: "查看会员", Description: "查看会员列表和会员详情"},
				{Code: "catalog.product.read", Name: "查看商品模板", Description: "查看商品目录和商品详情"},
			},
		},
		{
			Code:        "settings",
			Name:        "系统配置",
			Description: "系统参数与品牌设置",
			Items: []PermissionItem{
				{Code: "settings.read", Name: "查看系统配置", Description: "查看品牌、拍照与支付配置"},
				{Code: "settings.update", Name: "修改系统配置", Description: "更新系统基础配置"},
			},
		},
	}

	roles := []RoleTemplate{
		{
			Code:        "owner",
			Name:        "老板",
			Description: "默认跨店查看与配置能力",
			DataScope:   "org_all",
			Permissions: flattenPermissions(permissions),
		},
		{
			Code:        "manager",
			Name:        "店长",
			Description: "可管理自己负责门店的看板、收银和回收",
			DataScope:   "assigned_stores",
			Permissions: []string{
				"dashboard.summary.view",
				"store.read",
				"role.read",
				"permission.catalog.read",
				"member.read",
				"catalog.product.read",
				"cashier.order.read",
				"cashier.order.create",
				"recycle.order.read",
				"recycle.order.draft",
				"recycle.order.confirm",
				"settings.read",
			},
		},
		{
			Code:        "cashier",
			Name:        "收银员",
			Description: "门店前台收银与回收录单",
			DataScope:   "assigned_stores",
			Permissions: []string{
				"dashboard.summary.view",
				"store.read",
				"member.read",
				"catalog.product.read",
				"cashier.order.read",
				"cashier.order.create",
				"recycle.order.read",
				"recycle.order.draft",
				"settings.read",
			},
		},
	}

	stores := []StoreInfo{
		{
			ID:        "store-shenzhen-nanshan",
			OrgID:     "org-gold-demo",
			Code:      "SZ-NS",
			Name:      "南山旗舰店",
			City:      "深圳",
			Address:   "深圳市南山区科技园科苑路 18 号",
			Manager:   "李店长",
			Status:    "active",
			IsDefault: true,
		},
		{
			ID:        "store-guangzhou-tianhe",
			OrgID:     "org-gold-demo",
			Code:      "GZ-TH",
			Name:      "天河体验店",
			City:      "广州",
			Address:   "广州市天河区体育东路 88 号",
			Manager:   "王店长",
			Status:    "active",
			IsDefault: false,
		},
	}

	users := []UserAccount{
		{
			ID:          "user-owner-001",
			OrgID:       "org-gold-demo",
			Username:    "boss",
			DisplayName: "陈老板",
			Password:    "Boss123!",
			RoleCode:    "owner",
			RoleName:    "老板",
			DataScope:   "org_all",
			StoreIDs:    []string{stores[0].ID, stores[1].ID},
			Permissions: flattenPermissions(permissions),
			Status:      "active",
		},
		{
			ID:          "user-manager-001",
			OrgID:       "org-gold-demo",
			Username:    "manager.sz",
			DisplayName: "李店长",
			Password:    "Manager123!",
			RoleCode:    "manager",
			RoleName:    "店长",
			DataScope:   "assigned_stores",
			StoreIDs:    []string{stores[0].ID},
			Permissions: roles[1].Permissions,
			Status:      "active",
		},
		{
			ID:          "user-cashier-001",
			OrgID:       "org-gold-demo",
			Username:    "cashier.sz",
			DisplayName: "张收银",
			Password:    "Cashier123!",
			RoleCode:    "cashier",
			RoleName:    "收银员",
			DataScope:   "assigned_stores",
			StoreIDs:    []string{stores[0].ID},
			Permissions: roles[2].Permissions,
			Status:      "active",
		},
	}

	usersByUsername := make(map[string]UserAccount, len(users))
	usersByID := make(map[string]UserAccount, len(users))
	for _, user := range users {
		usersByUsername[user.Username] = user
		usersByID[user.ID] = user
	}

	now := time.Now()
	settings := SystemSettings{
		OrgID: "org-gold-demo",
		Brand: BrandSettings{
			Name:            "金掌柜回收",
			ServicePhone:    "400-888-2026",
			ReceiptTitle:    "黄金回收收银系统",
			SupportMiniApp:  true,
			DefaultCurrency: "CNY",
		},
		Recycle: RecycleSettings{
			MinPhotoCount:     2,
			MaxPhotoCount:     3,
			RequireExactThree: false,
			RequireIDCheck:    true,
		},
		Payments: PaymentSettings{
			EnableCash:         true,
			EnableWechatPay:    true,
			EnableBankTransfer: true,
		},
		FeatureFlags: map[string]bool{
			"mockMode":             true,
			"dataScopePlaceholder": true,
			"wechatCallbackReady":  false,
		},
		UpdatedBy: "system",
		UpdatedAt: now,
	}

	cashierOrders := []CashierOrder{
		{
			ID:            "cashier-order-001",
			OrderNo:       "CO202605100001",
			OrgID:         "org-gold-demo",
			StoreID:       stores[0].ID,
			StoreName:     stores[0].Name,
			Status:        "paid",
			PaymentMethod: "wechat_pay",
			TotalAmount:   1688,
			PaidAmount:    1688,
			Remark:        "开业活动订单",
			Items: []CashierOrderLine{
				{Name: "足金戒指", Quantity: 1, UnitPrice: 1688, Amount: 1688},
			},
			CreatedBy: users[2].DisplayName,
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:            "cashier-order-002",
			OrderNo:       "CO202605100002",
			OrgID:         "org-gold-demo",
			StoreID:       stores[1].ID,
			StoreName:     stores[1].Name,
			Status:        "paid",
			PaymentMethod: "cash",
			TotalAmount:   3200,
			PaidAmount:    3200,
			Remark:        "门店现结",
			Items: []CashierOrderLine{
				{Name: "黄金转运珠", Quantity: 2, UnitPrice: 1600, Amount: 3200},
			},
			CreatedBy: users[0].DisplayName,
			CreatedAt: now.Add(-90 * time.Minute),
		},
	}

	confirmedAt := now.Add(-40 * time.Minute)
	recycleOrders := map[string]RecycleOrder{
		"recycle-order-001": {
			ID:              "recycle-order-001",
			OrderNo:         "RO202605100001",
			OrgID:           "org-gold-demo",
			StoreID:         stores[0].ID,
			StoreName:       stores[0].Name,
			Status:          "confirmed",
			CustomerName:    "王女士",
			CustomerPhone:   "13800001111",
			EstimatedAmount: 5320,
			ConfirmedAmount: 5300,
			Items: []RecycleItem{
				{Category: "戒指", Purity: "AU999", WeightGram: 12.5},
			},
			AttachmentURLs: []string{
				"https://mock.example.com/recycle/1-front.jpg",
				"https://mock.example.com/recycle/1-back.jpg",
			},
			Remark:      "老客复购",
			CreatedBy:   users[1].DisplayName,
			CreatedAt:   now.Add(-75 * time.Minute),
			ConfirmedAt: &confirmedAt,
		},
	}

	members := []MemberProfile{
		{
			ID:                 "member-001",
			OrgID:              "org-gold-demo",
			StoreID:            stores[0].ID,
			StoreName:          stores[0].Name,
			Name:               "张女士",
			Phone:              "13800002026",
			Level:              "VIP",
			Status:             "active",
			TotalOrders:        6,
			TotalRecycleAmount: 28640,
			LastVisitAt:        now.Add(-18 * time.Hour),
			PreferredPurity:    "足金999",
			SourceChannel:      "熟客复购",
			ManagerName:        "李店长",
			IDVerified:         true,
			Tags:               []string{"高净值", "可回访", "到店复秤"},
			Notes:              "偏好工作日下午到店，成交前会要求复秤拍照。",
		},
		{
			ID:                 "member-002",
			OrgID:              "org-gold-demo",
			StoreID:            stores[0].ID,
			StoreName:          stores[0].Name,
			Name:               "陈先生",
			Phone:              "13900008866",
			Level:              "潜力客户",
			Status:             "follow_up",
			TotalOrders:        2,
			TotalRecycleAmount: 9240,
			LastVisitAt:        now.Add(-2 * 24 * time.Hour),
			PreferredPurity:    "18K",
			SourceChannel:      "企业回访",
			ManagerName:        "李店长",
			IDVerified:         true,
			Tags:               []string{"银行卡结算", "待回访"},
			Notes:              "对到账时效敏感，偏好银行卡。",
		},
		{
			ID:                 "member-003",
			OrgID:              "org-gold-demo",
			StoreID:            stores[1].ID,
			StoreName:          stores[1].Name,
			Name:               "王阿姨",
			Phone:              "13700005588",
			Level:              "普通会员",
			Status:             "sleeping",
			TotalOrders:        1,
			TotalRecycleAmount: 3820,
			LastVisitAt:        now.Add(-15 * 24 * time.Hour),
			PreferredPurity:    "22K",
			SourceChannel:      "渠道转介绍",
			ManagerName:        "王店长",
			IDVerified:         false,
			Tags:               []string{"老客家属", "待身份证补录"},
			Notes:              "上次成交后暂无二次回访记录。",
		},
	}

	catalogProducts := []CatalogProduct{
		{
			ID:               "product-001",
			OrgID:            "org-gold-demo",
			Name:             "足金手镯标准款",
			SKU:              "GJG-SZ-001",
			Category:         "金饰",
			Purity:           "足金999",
			BenchPrice:       742,
			RetailPrice:      786,
			GramWeight:       12.68,
			Status:           "active",
			Inventory:        9,
			StoreIDs:         []string{stores[0].ID, stores[1].ID},
			Stores:           []string{stores[0].Name, stores[1].Name},
			Tags:             []string{"主推", "高转化"},
			RecommendedScene: "适合门店当面议价、快速复秤录单。",
			QuoteLeadTime:    "30 秒内可完成报价",
		},
		{
			ID:               "product-002",
			OrgID:            "org-gold-demo",
			Name:             "18K 项链回收模板",
			SKU:              "GJG-KG-018",
			Category:         "K金",
			Purity:           "18K",
			BenchPrice:       558,
			RetailPrice:      612,
			GramWeight:       8.92,
			Status:           "active",
			Inventory:        4,
			StoreIDs:         []string{stores[0].ID},
			Stores:           []string{stores[0].Name},
			Tags:             []string{"K金", "常见回收"},
			RecommendedScene: "适合员工快速带入常见 K 金参数。",
			QuoteLeadTime:    "60 秒内完成复检",
		},
		{
			ID:               "product-003",
			OrgID:            "org-gold-demo",
			Name:             "足金金条 50g",
			SKU:              "GJG-BAR-050",
			Category:         "金条",
			Purity:           "足金9999",
			BenchPrice:       748,
			RetailPrice:      802,
			GramWeight:       50,
			Status:           "draft",
			Inventory:        2,
			StoreIDs:         []string{stores[1].ID},
			Stores:           []string{stores[1].Name},
			Tags:             []string{"大额", "待复核"},
			RecommendedScene: "大额回收建议店长复核并走双人确认。",
			QuoteLeadTime:    "需二次审批",
		},
	}

	adminRoles, adminAbilityGroups, adminStores, adminUsers, adminProducts, adminPaymentConfig, adminPrintTemplate, adminPaymentRecords, adminCashierOrders, adminRecycleOrders, adminSystemProfile, adminAuditLogs, adminTemplateInit := buildAdminFixtures()

	return &MockStore{
		usersByUsername:     usersByUsername,
		usersByID:           usersByID,
		roles:               roles,
		stores:              stores,
		permissionCatalog:   permissions,
		sessions:            make(map[string]Session),
		members:             members,
		catalogProducts:     catalogProducts,
		cashierOrders:       cashierOrders,
		recycleOrders:       recycleOrders,
		settings:            settings,
		cashierSeq:          len(cashierOrders) + 1,
		recycleSeq:          len(recycleOrders) + 1,
		adminRoles:          adminRoles,
		adminAbilityGroups:  adminAbilityGroups,
		adminStores:         adminStores,
		adminUsers:          adminUsers,
		adminProducts:       adminProducts,
		adminPaymentConfig:  adminPaymentConfig,
		adminPrintTemplate:  adminPrintTemplate,
		adminPaymentRecords: adminPaymentRecords,
		adminCashierOrders:  adminCashierOrders,
		adminRecycleOrders:  adminRecycleOrders,
		adminSystemProfile:  adminSystemProfile,
		adminAuditLogs:      adminAuditLogs,
		adminTemplateInit:   adminTemplateInit,
	}
}

func flattenPermissions(groups []PermissionGroup) []string {
	var permissions []string
	for _, group := range groups {
		for _, item := range group.Items {
			permissions = append(permissions, item.Code)
		}
	}
	return permissions
}

func (s *MockStore) authenticate(username, password string) (UserAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByUsername[username]
	if !ok || user.Password != password || user.Status != "active" {
		return UserAccount{}, errInvalidCredentials
	}
	return user, nil
}

func (s *MockStore) createSession(secret string, user UserAccount) Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	payload := fmt.Sprintf("%s:%d", user.ID, now.UnixNano())
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	token := hex.EncodeToString(mac.Sum(nil))

	session := Session{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		LoginAt:   now,
		ExpiresAt: now.Add(12 * time.Hour),
	}
	s.sessions[token] = session
	return session
}

func (s *MockStore) getUserByToken(token string) (UserAccount, Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[token]
	if !ok || session.ExpiresAt.Before(time.Now()) {
		return UserAccount{}, Session{}, false
	}

	user, ok := s.usersByID[session.UserID]
	if !ok {
		return UserAccount{}, Session{}, false
	}
	return user, session, true
}

func (s *MockStore) deleteSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *MockStore) listStoresForUser(user UserAccount) []StoreInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if user.DataScope == "org_all" {
		return append([]StoreInfo(nil), s.stores...)
	}

	visible := make([]StoreInfo, 0, len(user.StoreIDs))
	allow := make(map[string]struct{}, len(user.StoreIDs))
	for _, storeID := range user.StoreIDs {
		allow[storeID] = struct{}{}
	}
	for _, store := range s.stores {
		if _, ok := allow[store.ID]; ok {
			visible = append(visible, store)
		}
	}
	return visible
}

func (s *MockStore) canAccessStore(user UserAccount, storeID string) bool {
	if user.DataScope == "org_all" {
		return true
	}
	for _, id := range user.StoreIDs {
		if id == storeID {
			return true
		}
	}
	return false
}

func (s *MockStore) getStore(storeID string) (StoreInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, store := range s.stores {
		if store.ID == storeID {
			return store, true
		}
	}
	return StoreInfo{}, false
}

func (s *MockStore) listRoles() []RoleTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]RoleTemplate(nil), s.roles...)
}

func (s *MockStore) listMembersForUser(user UserAccount) []MemberProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]MemberProfile, 0, len(s.members))
	for _, member := range s.members {
		if s.canAccessStore(user, member.StoreID) {
			items = append(items, member)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].LastVisitAt.After(items[j].LastVisitAt)
	})
	return items
}

func (s *MockStore) getMemberForUser(user UserAccount, memberID string) (MemberProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, member := range s.members {
		if member.ID == memberID {
			if !s.canAccessStore(user, member.StoreID) {
				return MemberProfile{}, errUnauthorizedStore
			}
			return member, nil
		}
	}
	return MemberProfile{}, errMemberNotFound
}

func (s *MockStore) listCatalogProductsForUser(user UserAccount) []CatalogProduct {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]CatalogProduct, 0, len(s.catalogProducts))
	for _, product := range s.catalogProducts {
		if user.DataScope == "org_all" {
			items = append(items, product)
			continue
		}
		for _, storeID := range product.StoreIDs {
			if s.canAccessStore(user, storeID) {
				items = append(items, product)
				break
			}
		}
	}
	return items
}

func (s *MockStore) getCatalogProductForUser(user UserAccount, productID string) (CatalogProduct, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, product := range s.catalogProducts {
		if product.ID != productID && product.SKU != productID {
			continue
		}
		if user.DataScope == "org_all" {
			return product, nil
		}
		for _, storeID := range product.StoreIDs {
			if s.canAccessStore(user, storeID) {
				return product, nil
			}
		}
		return CatalogProduct{}, errUnauthorizedStore
	}
	return CatalogProduct{}, errProductNotFound
}

func (s *MockStore) permissionGroups() []PermissionGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]PermissionGroup(nil), s.permissionCatalog...)
}

func (s *MockStore) listCashierOrders(user UserAccount, storeID string) ([]CashierOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]CashierOrder, 0, len(s.cashierOrders))
	for i := len(s.cashierOrders) - 1; i >= 0; i-- {
		order := s.cashierOrders[i]
		if storeID != "" && order.StoreID != storeID {
			continue
		}
		if !s.canAccessStore(user, order.StoreID) {
			continue
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (s *MockStore) getCashierOrder(user UserAccount, orderID string) (CashierOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, order := range s.cashierOrders {
		if order.ID == orderID || order.OrderNo == orderID {
			if !s.canAccessStore(user, order.StoreID) {
				return CashierOrder{}, errUnauthorizedStore
			}
			return order, nil
		}
	}
	return CashierOrder{}, errCashierNotFound
}

func (s *MockStore) createCashierOrder(user UserAccount, storeID string, items []CashierOrderLine, paymentMethod, remark string) (CashierOrder, error) {
	if !s.canAccessStore(user, storeID) {
		return CashierOrder{}, errUnauthorizedStore
	}
	store, ok := s.getStore(storeID)
	if !ok {
		return CashierOrder{}, errUnauthorizedStore
	}

	var normalized []CashierOrderLine
	var total float64
	for _, item := range items {
		amount := round2(item.UnitPrice * float64(item.Quantity))
		normalized = append(normalized, CashierOrderLine{
			Name:      strings.TrimSpace(item.Name),
			Quantity:  item.Quantity,
			UnitPrice: round2(item.UnitPrice),
			Amount:    amount,
		})
		total += amount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := CashierOrder{
		ID:            fmt.Sprintf("cashier-order-%03d", s.cashierSeq),
		OrderNo:       fmt.Sprintf("CO%s%04d", time.Now().Format("20060102"), s.cashierSeq),
		OrgID:         user.OrgID,
		StoreID:       store.ID,
		StoreName:     store.Name,
		Status:        "paid",
		PaymentMethod: paymentMethod,
		TotalAmount:   round2(total),
		PaidAmount:    round2(total),
		Remark:        remark,
		Items:         normalized,
		CreatedBy:     user.DisplayName,
		CreatedAt:     time.Now(),
	}
	s.cashierSeq++
	s.cashierOrders = append(s.cashierOrders, order)
	return order, nil
}

func (s *MockStore) listRecycleOrders(user UserAccount) []RecycleOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]RecycleOrder, 0, len(s.recycleOrders))
	for _, order := range s.recycleOrders {
		if s.canAccessStore(user, order.StoreID) {
			orders = append(orders, order)
		}
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})
	return orders
}

func (s *MockStore) getRecycleOrder(orderID string) (RecycleOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.recycleOrders[orderID]
	return order, ok
}

func (s *MockStore) createRecycleDraft(user UserAccount, storeID, customerName, customerPhone string, estimatedAmount float64, items []RecycleItem, attachmentURLs []string, remark string) (RecycleOrder, error) {
	if !s.canAccessStore(user, storeID) {
		return RecycleOrder{}, errUnauthorizedStore
	}
	store, ok := s.getStore(storeID)
	if !ok {
		return RecycleOrder{}, errUnauthorizedStore
	}

	var normalizedItems []RecycleItem
	for _, item := range items {
		normalizedItems = append(normalizedItems, RecycleItem{
			Category:   strings.TrimSpace(item.Category),
			Purity:     strings.TrimSpace(item.Purity),
			WeightGram: round2(item.WeightGram),
		})
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := RecycleOrder{
		ID:              fmt.Sprintf("recycle-order-%03d", s.recycleSeq),
		OrderNo:         fmt.Sprintf("RO%s%04d", time.Now().Format("20060102"), s.recycleSeq),
		OrgID:           user.OrgID,
		StoreID:         store.ID,
		StoreName:       store.Name,
		Status:          "draft",
		CustomerName:    strings.TrimSpace(customerName),
		CustomerPhone:   strings.TrimSpace(customerPhone),
		EstimatedAmount: round2(estimatedAmount),
		Items:           normalizedItems,
		AttachmentURLs:  append([]string(nil), attachmentURLs...),
		Remark:          strings.TrimSpace(remark),
		CreatedBy:       user.DisplayName,
		CreatedAt:       time.Now(),
	}
	s.recycleSeq++
	s.recycleOrders[order.ID] = order
	return order, nil
}

func (s *MockStore) confirmRecycleOrder(user UserAccount, orderID string, confirmedAmount float64, attachmentURLs []string, remark string) (RecycleOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.recycleOrders[orderID]
	if !ok {
		return RecycleOrder{}, errRecycleNotFound
	}
	if !s.canAccessStore(user, order.StoreID) {
		return RecycleOrder{}, errUnauthorizedStore
	}
	if order.Status != "draft" {
		return RecycleOrder{}, errRecycleConflict
	}

	attachments := order.AttachmentURLs
	if len(attachmentURLs) > 0 {
		attachments = append([]string(nil), attachmentURLs...)
	}
	now := time.Now()
	order.Status = "confirmed"
	order.ConfirmedAmount = round2(confirmedAmount)
	order.AttachmentURLs = attachments
	order.Remark = strings.TrimSpace(remark)
	order.ConfirmedAt = &now
	s.recycleOrders[orderID] = order
	return order, nil
}

func (s *MockStore) getSettings() SystemSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *MockStore) updateSettings(user UserAccount, update SystemSettings) SystemSettings {
	s.mu.Lock()
	defer s.mu.Unlock()

	if update.Brand.Name != "" {
		s.settings.Brand.Name = update.Brand.Name
	}
	if update.Brand.ServicePhone != "" {
		s.settings.Brand.ServicePhone = update.Brand.ServicePhone
	}
	if update.Brand.ReceiptTitle != "" {
		s.settings.Brand.ReceiptTitle = update.Brand.ReceiptTitle
	}
	if update.Brand.DefaultCurrency != "" {
		s.settings.Brand.DefaultCurrency = update.Brand.DefaultCurrency
	}
	s.settings.Brand.SupportMiniApp = update.Brand.SupportMiniApp

	if update.Recycle.MinPhotoCount > 0 {
		s.settings.Recycle.MinPhotoCount = update.Recycle.MinPhotoCount
	}
	if update.Recycle.MaxPhotoCount > 0 {
		s.settings.Recycle.MaxPhotoCount = update.Recycle.MaxPhotoCount
	}
	s.settings.Recycle.RequireExactThree = update.Recycle.RequireExactThree
	s.settings.Recycle.RequireIDCheck = update.Recycle.RequireIDCheck
	s.settings.Payments.EnableCash = update.Payments.EnableCash
	s.settings.Payments.EnableWechatPay = update.Payments.EnableWechatPay
	s.settings.Payments.EnableBankTransfer = update.Payments.EnableBankTransfer

	if update.FeatureFlags != nil {
		for key, value := range update.FeatureFlags {
			s.settings.FeatureFlags[key] = value
		}
	}

	s.settings.UpdatedBy = user.DisplayName
	s.settings.UpdatedAt = time.Now()
	return s.settings
}

func (s *MockStore) dashboardSummary(user UserAccount) DashboardSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visibleStores := s.listStoresForUser(user)
	storeSet := make(map[string]struct{}, len(visibleStores))
	for _, store := range visibleStores {
		storeSet[store.ID] = struct{}{}
	}

	today := time.Now().Format("2006-01-02")
	summary := DashboardSummary{
		OrgID:             user.OrgID,
		VisibleStoreCount: len(visibleStores),
	}

	for _, order := range s.cashierOrders {
		if _, ok := storeSet[order.StoreID]; !ok {
			continue
		}
		if order.CreatedAt.Format("2006-01-02") == today {
			summary.TodayCashierOrderCount++
			summary.TodayCashierAmount += order.PaidAmount
		}
	}

	for _, order := range s.recycleOrders {
		if _, ok := storeSet[order.StoreID]; !ok {
			continue
		}
		switch order.Status {
		case "draft":
			summary.DraftRecycleOrderCount++
		case "confirmed":
			summary.ConfirmedRecycleOrderCount++
		}
		if order.CreatedAt.Format("2006-01-02") == today {
			for _, item := range order.Items {
				summary.TodayRecycleWeightGram += item.WeightGram
			}
		}
	}

	summary.TodayCashierAmount = round2(summary.TodayCashierAmount)
	summary.TodayRecycleWeightGram = round2(summary.TodayRecycleWeightGram)
	return summary
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}
