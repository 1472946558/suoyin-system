package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
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
	errUploadNotFound     = errors.New("upload session not found")
	errMemberNotFound     = errors.New("member not found")
	errProductNotFound    = errors.New("product not found")
)

const (
	configKeyAdminPrint   = "admin_print_template"
	configKeyAdminSystem  = "admin_system_profile"
	configKeyAdminAudit   = "admin_audit_logs"
	configKeyBaseStores   = "base_stores"
	configKeyBaseUsers    = "base_users"
	configKeyBaseRoles    = "base_roles"
	configKeyBaseMembers  = "base_members"
	configKeyBaseProducts = "base_products"
	configKeyBaseSettings = "base_system_settings"
	configKeyAdminRoles   = "admin_roles"
)

type MockStore struct {
	mu                  sync.RWMutex
	persistence         *Persistence
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
	adminPrintTemplate  AdminPrintTemplate
	adminCashierOrders  []AdminCashierOrderView
	adminRecycleOrders  []AdminRecycleOrderView
	adminSystemProfile  AdminSystemProfile
	adminAuditLogs      []AdminAuditLogRecord
	adminTemplateInit   AdminTemplateInitPlan
	uploadSessions      map[string]UploadPreparation
	pendingMiniAppBinds []PendingMiniAppBinding
}

func newMockStore(persistence *Persistence) (*MockStore, error) {
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
				{Code: "settings.read", Name: "查看系统配置", Description: "查看品牌、拍照与对象存储配置"},
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
			OrgID:     "org-gold-v1",
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
			OrgID:     "org-gold-v1",
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
			ID:           "user-owner-001",
			OrgID:        "org-gold-v1",
			Username:     "boss",
			DisplayName:  "廖总",
			Phone:        "13800000001",
			Password:     "Boss123!",
			RoleCode:     "owner",
			RoleName:     "老板",
			DataScope:    "org_all",
			StoreIDs:     []string{stores[0].ID, stores[1].ID},
			Permissions:  flattenPermissions(permissions),
			Status:       "active",
			WechatOpenID: "",
		},
		{
			ID:           "user-owner-002",
			OrgID:        "org-gold-v1",
			Username:     "boss.demo",
			DisplayName:  "示例老板",
			Phone:        "13800000000",
			Password:     "Boss123!",
			RoleCode:     "owner",
			RoleName:     "老板",
			DataScope:    "org_all",
			StoreIDs:     []string{stores[0].ID, stores[1].ID},
			Permissions:  flattenPermissions(permissions),
			Status:       "active",
			WechatOpenID: "",
		},
		{
			ID:           "user-manager-001",
			OrgID:        "org-gold-v1",
			Username:     "manager.sz",
			DisplayName:  "李店长",
			Phone:        "13800002108",
			Password:     "Manager123!",
			RoleCode:     "manager",
			RoleName:     "店长",
			DataScope:    "assigned_stores",
			StoreIDs:     []string{stores[0].ID},
			Permissions:  roles[1].Permissions,
			Status:       "active",
			WechatOpenID: "wx-openid-manager-001",
		},
		{
			ID:           "user-cashier-001",
			OrgID:        "org-gold-v1",
			Username:     "cashier.sz",
			DisplayName:  "张收银",
			Phone:        "13800003001",
			Password:     "Cashier123!",
			RoleCode:     "cashier",
			RoleName:     "收银员",
			DataScope:    "assigned_stores",
			StoreIDs:     []string{stores[0].ID},
			Permissions:  roles[2].Permissions,
			Status:       "active",
			WechatOpenID: "wx-openid-cashier-001",
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
		OrgID: "org-gold-v1",
		Brand: BrandSettings{
			Name:            "金掌柜回收",
			ServicePhone:    "400-888-2026",
			ReceiptTitle:    "黄金回收收银系统",
			SupportMiniApp:  true,
			DefaultCurrency: "CNY",
		},
		Recycle: RecycleSettings{
			MinPhotoCount:     3,
			MaxPhotoCount:     3,
			RequireExactThree: false,
			RequireIDCheck:    true,
		},
		Storage: StorageSettings{
			Enabled:           false,
			Provider:          "unconfigured",
			Bucket:            "",
			Region:            "",
			PublicBaseURL:     "",
			PathPrefix:        "recycle-evidence",
			UploadStrategy:    "manual_url",
			CallbackEnabled:   false,
			StatusDescription: "对象存储待配置，当前允许先登记外部图片链接。",
		},
		FeatureFlags: map[string]bool{
			"localDataMode":        true,
			"dataScopePlaceholder": true,
			"wechatCallbackReady":  false,
		},
		UpdatedBy: "system",
		UpdatedAt: now,
	}
	applySystemSettingDefaults(&settings)

	cashierOrders := []CashierOrder{
		{
			ID:            "cashier-order-001",
			OrderNo:       "CO202605100001",
			OrgID:         "org-gold-v1",
			StoreID:       stores[0].ID,
			StoreName:     stores[0].Name,
			Status:        "paid",
			CustomerName:  "张女士",
			CustomerPhone: "13800138000",
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
			OrgID:         "org-gold-v1",
			StoreID:       stores[1].ID,
			StoreName:     stores[1].Name,
			Status:        "paid",
			CustomerName:  "陈先生",
			CustomerPhone: "13900139000",
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
			OrgID:           "org-gold-v1",
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
				"https://example.com/assets/recycle/1-front.jpg",
				"https://example.com/assets/recycle/1-back.jpg",
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
			OrgID:              "org-gold-v1",
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
			OrgID:              "org-gold-v1",
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
			OrgID:              "org-gold-v1",
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
			OrgID:            "org-gold-v1",
			Name:             "足金手镯标准款",
			SKU:              "GJG-SZ-001",
			Category:         "金饰",
			CategoryTab:      "黄金饰品",
			ImageURL:         "/assets/ui/price.png",
			Purity:           "足金999",
			BenchPrice:       742,
			RetailPrice:      786,
			GramWeight:       12.68,
			Status:           "active",
			Inventory:        9,
			StockStatus:      "normal",
			StoreIDs:         []string{stores[0].ID, stores[1].ID},
			Stores:           []string{stores[0].Name, stores[1].Name},
			Tags:             []string{"主推", "高转化"},
			RecommendedScene: "适合门店当面议价、快速复秤录单。",
			QuoteLeadTime:    "30 秒内可完成报价",
		},
		{
			ID:               "product-002",
			OrgID:            "org-gold-v1",
			Name:             "18K 项链回收模板",
			SKU:              "GJG-KG-018",
			Category:         "K金",
			CategoryTab:      "K金",
			ImageURL:         "/assets/ui/weight.png",
			Purity:           "18K",
			BenchPrice:       558,
			RetailPrice:      612,
			GramWeight:       8.92,
			Status:           "active",
			Inventory:        4,
			StockStatus:      "low",
			StoreIDs:         []string{stores[0].ID},
			Stores:           []string{stores[0].Name},
			Tags:             []string{"K金", "常见回收"},
			RecommendedScene: "适合员工快速带入常见 K 金参数。",
			QuoteLeadTime:    "60 秒内完成复检",
		},
		{
			ID:               "product-003",
			OrgID:            "org-gold-v1",
			Name:             "足金金条 50g",
			SKU:              "GJG-BAR-050",
			Category:         "金条",
			CategoryTab:      "投资金",
			ImageURL:         "/assets/ui/document.png",
			Purity:           "足金9999",
			BenchPrice:       748,
			RetailPrice:      802,
			GramWeight:       50,
			Status:           "draft",
			Inventory:        2,
			StockStatus:      "review",
			StoreIDs:         []string{stores[1].ID},
			Stores:           []string{stores[1].Name},
			Tags:             []string{"大额", "待复核"},
			RecommendedScene: "大额回收建议店长复核并走双人确认。",
			QuoteLeadTime:    "需二次审批",
		},
	}

	adminRoles, adminAbilityGroups, adminStores, adminUsers, adminProducts, adminPrintTemplate, adminCashierOrders, adminRecycleOrders, adminSystemProfile, adminAuditLogs, adminTemplateInit := buildAdminFixtures()

	store := &MockStore{
		persistence:        persistence,
		usersByUsername:    usersByUsername,
		usersByID:          usersByID,
		roles:              roles,
		stores:             stores,
		permissionCatalog:  permissions,
		sessions:           make(map[string]Session),
		members:            members,
		catalogProducts:    catalogProducts,
		cashierOrders:      cashierOrders,
		recycleOrders:      recycleOrders,
		settings:           settings,
		cashierSeq:         len(cashierOrders) + 1,
		recycleSeq:         len(recycleOrders) + 1,
		adminRoles:         adminRoles,
		adminAbilityGroups: adminAbilityGroups,
		adminStores:        adminStores,
		adminUsers:         adminUsers,
		adminProducts:      adminProducts,
		adminPrintTemplate: adminPrintTemplate,
		adminCashierOrders: adminCashierOrders,
		adminRecycleOrders: adminRecycleOrders,
		adminSystemProfile: adminSystemProfile,
		adminAuditLogs:     adminAuditLogs,
		adminTemplateInit:  adminTemplateInit,
		uploadSessions:     make(map[string]UploadPreparation),
	}

	store.rebuildRecycleAttachmentsLocked()

	if persistence != nil {
		store.settings.FeatureFlags["localDataMode"] = false
		if err := store.bootstrapPersistence(); err != nil {
			return nil, err
		}
	}

	return store, nil
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

func (s *MockStore) mockMiniAppUserForRole(roleKey string) (UserAccount, error) {
	username := "cashier.sz"
	switch strings.TrimSpace(roleKey) {
	case "owner":
		username = "boss"
	case "manager":
		username = "manager.sz"
	case "cashier", "clerk":
		username = "cashier.sz"
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.usersByUsername[username]
	if !ok || user.Status != "active" {
		return UserAccount{}, errInvalidCredentials
	}
	return user, nil
}

func (s *MockStore) findUserByWechatOpenID(openID string) (UserAccount, bool) {
	normalized := strings.TrimSpace(openID)
	if normalized == "" {
		return UserAccount{}, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.usersByID {
		if user.Status != "active" {
			continue
		}
		if strings.TrimSpace(user.WechatOpenID) == normalized {
			return user, true
		}
	}
	return UserAccount{}, false
}

func (s *MockStore) bindMiniAppUserByProfile(session wechatMiniAppSession, phone, roleKey string) (UserAccount, bool) {
	openID := strings.TrimSpace(session.OpenID)
	normalizedPhone := strings.TrimSpace(phone)
	if openID == "" || normalizedPhone == "" {
		return UserAccount{}, false
	}

	requestedRole := baseRoleCodeFromAdmin(normalizeAdminRoleKey(roleKey))

	s.mu.Lock()
	defer s.mu.Unlock()

	for userID, user := range s.usersByID {
		if user.Status != "active" {
			continue
		}
		if strings.TrimSpace(user.WechatOpenID) != "" {
			continue
		}
		if strings.TrimSpace(user.Phone) != normalizedPhone {
			continue
		}
		if strings.TrimSpace(roleKey) != "" && user.RoleCode != requestedRole {
			continue
		}

		user.WechatOpenID = openID
		s.usersByID[userID] = user
		s.usersByUsername[user.Username] = user
		s.removePendingMiniAppBindingLocked(openID)
		if s.persistence != nil {
			_ = s.persistUsersLocked()
		}
		s.appendAuditLogLocked("小程序登录", "自动绑定微信身份", user.DisplayName, "success", "medium", fmt.Sprintf("%s 已通过手机号完成微信绑定", user.DisplayName))
		return user, true
	}

	return UserAccount{}, false
}

func (s *MockStore) recordPendingMiniAppBinding(session wechatMiniAppSession, name, phone, roleKey, storeName, storeCode string) {
	openID := strings.TrimSpace(session.OpenID)
	if openID == "" {
		return
	}

	normalizedRoleKey := normalizeAdminRoleKey(roleKey)
	record := PendingMiniAppBinding{
		ID:            "miniapp-bind-" + shortID(openID),
		MiniAppOpenID: openID,
		UnionID:       strings.TrimSpace(session.UnionID),
		OperatorName:  strings.TrimSpace(name),
		Phone:         strings.TrimSpace(phone),
		RoleKey:       normalizedRoleKey,
		RoleName:      roleNameForKey(normalizedRoleKey),
		StoreName:     strings.TrimSpace(storeName),
		StoreCode:     strings.TrimSpace(storeCode),
		CreatedAt:     time.Now().Format("2006-01-02 15:04:05"),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.pendingMiniAppBinds {
		if item.MiniAppOpenID == openID {
			s.pendingMiniAppBinds[index] = record
			return
		}
	}

	s.pendingMiniAppBinds = append([]PendingMiniAppBinding{record}, s.pendingMiniAppBinds...)
	if len(s.pendingMiniAppBinds) > 50 {
		s.pendingMiniAppBinds = s.pendingMiniAppBinds[:50]
	}
}

func (s *MockStore) listPendingMiniAppBindings() []PendingMiniAppBinding {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]PendingMiniAppBinding, len(s.pendingMiniAppBinds))
	copy(items, s.pendingMiniAppBinds)
	return items
}

func (s *MockStore) removePendingMiniAppBindingLocked(openID string) {
	normalized := strings.TrimSpace(openID)
	if normalized == "" {
		return
	}
	items := s.pendingMiniAppBinds[:0]
	for _, item := range s.pendingMiniAppBinds {
		if item.MiniAppOpenID != normalized {
			items = append(items, item)
		}
	}
	s.pendingMiniAppBinds = items
}

func shortID(value string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])[:12]
}

func (s *MockStore) createSession(secret string, user UserAccount) (Session, error) {
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
	if s.persistence != nil {
		if err := s.persistence.saveSession(context.Background(), session); err != nil {
			delete(s.sessions, token)
			return Session{}, err
		}
	}
	return session, nil
}

func (s *MockStore) getUserByToken(token string) (UserAccount, Session, bool) {
	s.mu.RLock()
	session, ok := s.sessions[token]
	user := s.usersByID[session.UserID]
	s.mu.RUnlock()

	if ok && !session.ExpiresAt.Before(time.Now()) {
		if user.ID != "" {
			return user, session, true
		}
	}

	if s.persistence != nil {
		persisted, found, err := s.persistence.loadSession(context.Background(), token)
		if err != nil || !found || persisted.ExpiresAt.Before(time.Now()) {
			return UserAccount{}, Session{}, false
		}

		s.mu.Lock()
		s.sessions[token] = persisted
		user, ok = s.usersByID[persisted.UserID]
		s.mu.Unlock()
		if ok {
			return user, persisted, true
		}
	}

	if !ok || user.ID == "" {
		return UserAccount{}, Session{}, false
	}
	return user, session, true
}

func (s *MockStore) deleteSession(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
	if s.persistence != nil {
		return s.persistence.deleteSession(context.Background(), token)
	}
	return nil
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

func attachmentFileName(rawURL string, index int) string {
	cleaned := strings.TrimSpace(rawURL)
	if cleaned == "" {
		return fmt.Sprintf("attachment-%02d.jpg", index+1)
	}
	if idx := strings.LastIndex(cleaned, "/"); idx >= 0 && idx < len(cleaned)-1 {
		cleaned = cleaned[idx+1:]
	}
	if idx := strings.Index(cleaned, "?"); idx >= 0 {
		cleaned = cleaned[:idx]
	}
	if strings.TrimSpace(cleaned) == "" {
		return fmt.Sprintf("attachment-%02d.jpg", index+1)
	}
	return cleaned
}

func buildRecycleAttachmentAssets(order RecycleOrder, urls []string, actor string, uploadedAt time.Time) []AttachmentAsset {
	items := make([]AttachmentAsset, 0, len(urls))
	for index, rawURL := range urls {
		fileName := attachmentFileName(rawURL, index)
		items = append(items, AttachmentAsset{
			ID:              fmt.Sprintf("att-%s-%02d", order.ID, index+1),
			OrgID:           order.OrgID,
			OrderID:         order.ID,
			StoreID:         order.StoreID,
			Category:        "recycle_photo",
			StorageProvider: "manual_url",
			ObjectKey:       fmt.Sprintf("recycle/%s/%s", order.OrderNo, fileName),
			PublicURL:       rawURL,
			ThumbnailURL:    rawURL,
			FileName:        fileName,
			ContentType:     "image/jpeg",
			SizeBytes:       0,
			Source:          "manual",
			Status:          "archived",
			UploadedBy:      actor,
			UploadedAt:      uploadedAt,
		})
	}
	return items
}

func buildRecycleAttachmentAssetsPreservingProvider(order RecycleOrder, urls []string, existing []AttachmentAsset, actor string, uploadedAt time.Time) []AttachmentAsset {
	if len(existing) == 0 {
		return buildRecycleAttachmentAssets(order, urls, actor, uploadedAt)
	}

	existingByURL := make(map[string]AttachmentAsset, len(existing))
	for _, item := range existing {
		for _, key := range []string{item.PublicURL, item.ThumbnailURL} {
			normalized := strings.TrimSpace(key)
			if normalized != "" {
				existingByURL[normalized] = item
			}
		}
	}

	items := make([]AttachmentAsset, 0, len(urls))
	for index, rawURL := range urls {
		normalizedURL := strings.TrimSpace(rawURL)
		if item, ok := existingByURL[normalizedURL]; ok {
			item.OrderID = order.ID
			item.StoreID = order.StoreID
			item.UploadedBy = actor
			item.UploadedAt = uploadedAt
			items = append(items, item)
			continue
		}
		items = append(items, buildRecycleAttachmentAssets(order, []string{rawURL}, actor, uploadedAt)[0])
		if items[len(items)-1].ID == fmt.Sprintf("att-%s-01", order.ID) {
			items[len(items)-1].ID = fmt.Sprintf("att-%s-%02d", order.ID, index+1)
		}
	}
	return items
}

func sanitizeObjectSegment(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "file"
	}
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if valid {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	cleaned := strings.Trim(builder.String(), "-")
	if cleaned == "" {
		return "file"
	}
	return cleaned
}

func applySystemSettingDefaults(settings *SystemSettings) {
	if settings == nil {
		return
	}
	if settings.FeatureFlags == nil {
		settings.FeatureFlags = map[string]bool{}
	}
	if strings.TrimSpace(settings.Storage.Provider) == "" {
		settings.Storage.Provider = "unconfigured"
	}
	if strings.TrimSpace(settings.Storage.PathPrefix) == "" {
		settings.Storage.PathPrefix = "recycle-evidence"
	}
	if strings.TrimSpace(settings.Storage.UploadStrategy) == "" {
		settings.Storage.UploadStrategy = "manual_url"
	}
	if strings.TrimSpace(settings.Storage.StatusDescription) == "" {
		settings.Storage.StatusDescription = "对象存储待配置，当前允许先登记外部图片链接。"
	}
}

func (s *MockStore) applyRuntimeConfig(cfg Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	if cfg.StorageEnabledSet {
		s.settings.Storage.Enabled = cfg.StorageEnabled
		changed = true
	}
	if strings.TrimSpace(cfg.StorageProvider) != "" {
		s.settings.Storage.Provider = strings.TrimSpace(cfg.StorageProvider)
		changed = true
	}
	if strings.TrimSpace(cfg.StorageBucket) != "" {
		s.settings.Storage.Bucket = strings.TrimSpace(cfg.StorageBucket)
		changed = true
	}
	if strings.TrimSpace(cfg.StorageRegion) != "" {
		s.settings.Storage.Region = strings.TrimSpace(cfg.StorageRegion)
		changed = true
	}
	if strings.TrimSpace(cfg.StorageEndpoint) != "" {
		s.settings.Storage.Endpoint = strings.TrimSpace(cfg.StorageEndpoint)
		changed = true
	}
	if strings.TrimSpace(cfg.StoragePublicBaseURL) != "" {
		s.settings.Storage.PublicBaseURL = strings.TrimSpace(cfg.StoragePublicBaseURL)
		changed = true
	}
	if strings.TrimSpace(cfg.StoragePathPrefix) != "" {
		s.settings.Storage.PathPrefix = strings.TrimSpace(cfg.StoragePathPrefix)
		changed = true
	}
	if strings.TrimSpace(cfg.StorageUploadStrategy) != "" {
		s.settings.Storage.UploadStrategy = strings.TrimSpace(cfg.StorageUploadStrategy)
		changed = true
	}
	if strings.TrimSpace(cfg.StorageAccessKeyID) != "" {
		s.settings.Storage.AccessKeyID = strings.TrimSpace(cfg.StorageAccessKeyID)
		changed = true
	}
	if strings.TrimSpace(cfg.StorageAccessKeySecret) != "" {
		s.settings.Storage.AccessKeySecret = strings.TrimSpace(cfg.StorageAccessKeySecret)
		changed = true
	}
	if cfg.StorageUploadURLTTL > 0 {
		s.settings.Storage.UploadURLTTL = cfg.StorageUploadURLTTL
		changed = true
	}
	if cfg.StorageCallbackSet {
		s.settings.Storage.CallbackEnabled = cfg.StorageCallbackEnabled
		changed = true
	}
	if strings.TrimSpace(cfg.StorageStatus) != "" {
		s.settings.Storage.StatusDescription = strings.TrimSpace(cfg.StorageStatus)
		changed = true
	}
	if !changed {
		return nil
	}

	applySystemSettingDefaults(&s.settings)
	s.settings.UpdatedBy = "runtime-env"
	s.settings.UpdatedAt = time.Now()
	s.adminSystemProfile = s.buildDynamicSystemProfileLocked()
	if s.persistence == nil {
		return nil
	}
	if err := s.persistSettingsLocked(); err != nil {
		return err
	}
	return s.persistence.saveConfig(context.Background(), configKeyAdminSystem, s.adminSystemProfile)
}

func buildStoragePublicURL(settings StorageSettings, objectKey string) string {
	base := strings.TrimRight(strings.TrimSpace(settings.PublicBaseURL), "/")
	if base == "" {
		return ""
	}
	return base + "/" + strings.TrimLeft(objectKey, "/")
}

func normalizedStorageUploadMode(settings StorageSettings) string {
	mode := strings.TrimSpace(settings.UploadStrategy)
	if mode == "" {
		return "direct_put"
	}
	return mode
}

func buildOSSSignedPutURL(settings StorageSettings, objectKey, contentType string, now time.Time) (string, time.Time, bool) {
	endpoint := strings.TrimRight(strings.TrimSpace(settings.Endpoint), "/")
	bucket := strings.TrimSpace(settings.Bucket)
	accessKeyID := strings.TrimSpace(settings.AccessKeyID)
	accessKeySecret := strings.TrimSpace(settings.AccessKeySecret)
	if endpoint == "" || bucket == "" || accessKeyID == "" || accessKeySecret == "" {
		return "", time.Time{}, false
	}

	scheme := "https"
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		scheme = parsed.Scheme
		endpoint = parsed.Host
	}
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ttl := settings.UploadURLTTL
	if ttl <= 0 {
		ttl = 1800
	}
	expiresAt := now.Add(time.Duration(ttl) * time.Second)
	expires := expiresAt.Unix()
	canonicalResource := "/" + bucket + "/" + strings.TrimLeft(objectKey, "/")
	stringToSign := fmt.Sprintf("PUT\n\n%s\n%d\n%s", contentType, expires, canonicalResource)
	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	query := url.Values{}
	query.Set("OSSAccessKeyId", accessKeyID)
	query.Set("Expires", strconv.FormatInt(expires, 10))
	query.Set("Signature", signature)
	escapedObjectKey := strings.TrimLeft(objectKey, "/")
	escapedObjectKey = strings.ReplaceAll(url.PathEscape(escapedObjectKey), "%2F", "/")
	signedURL := fmt.Sprintf("%s://%s.%s/%s?%s", scheme, bucket, endpoint, escapedObjectKey, query.Encode())
	return signedURL, expiresAt, true
}

func (s *MockStore) prepareRecycleAttachmentUpload(user UserAccount, storeID, fileName, contentType string, sizeBytes int64) (UploadPreparation, error) {
	if !s.canAccessStore(user, storeID) {
		return UploadPreparation{}, errUnauthorizedStore
	}
	store, ok := s.getStore(storeID)
	if !ok {
		return UploadPreparation{}, errUnauthorizedStore
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	uploadID := fmt.Sprintf("upload-%s-%d", store.Code, now.UnixNano())
	cleanFileName := attachmentFileName(fileName, 0)
	objectKey := strings.Trim(strings.TrimSpace(s.settings.Storage.PathPrefix), "/")
	if objectKey == "" {
		objectKey = "recycle-evidence"
	}
	objectKey = fmt.Sprintf("%s/%s/%s/%s-%s", objectKey, store.Code, now.Format("20060102"), uploadID, sanitizeObjectSegment(cleanFileName))

	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	publicURL := buildStoragePublicURL(s.settings.Storage, objectKey)
	uploadURL := publicURL
	storageReady := s.settings.Storage.Enabled && strings.TrimSpace(s.settings.Storage.PublicBaseURL) != ""
	mode := "manual_url"
	expiresAt := now.Add(30 * time.Minute)
	note := "对象存储未完全配置，可先上传到外部地址后再调用 complete 接口回填。"
	if storageReady {
		mode = normalizedStorageUploadMode(s.settings.Storage)
		if mode == "oss_signed_put" {
			if signedURL, signedExpiresAt, ok := buildOSSSignedPutURL(s.settings.Storage, objectKey, contentType, now); ok {
				uploadURL = signedURL
				expiresAt = signedExpiresAt
				note = "已生成 OSS 签名上传 URL，可由小程序直接 PUT 上传后回填。"
			} else {
				storageReady = false
				uploadURL = ""
				note = "OSS 签名上传配置不完整，请补 STORAGE_ENDPOINT / STORAGE_ACCESS_KEY_ID / STORAGE_ACCESS_KEY_SECRET。"
			}
		} else {
			note = "已生成对象键，可由前端按配置完成直传后回填。"
		}
	}
	preparation := UploadPreparation{
		UploadID:      uploadID,
		StoreID:       storeID,
		Category:      "recycle_photo",
		Provider:      s.settings.Storage.Provider,
		Bucket:        s.settings.Storage.Bucket,
		Region:        s.settings.Storage.Region,
		ObjectKey:     objectKey,
		FileName:      cleanFileName,
		ContentType:   contentType,
		SizeBytes:     sizeBytes,
		UploadURL:     uploadURL,
		PublicURL:     publicURL,
		Headers:       map[string]string{"content-type": contentType},
		FormFields:    map[string]string{},
		StorageReady:  storageReady,
		UploadMode:    mode,
		ExpiresAt:     expiresAt,
		ReferenceNote: note,
	}
	s.uploadSessions[uploadID] = preparation
	return preparation, nil
}

func (s *MockStore) completeRecycleAttachmentUpload(user UserAccount, uploadID, orderID, publicURL, thumbnailURL string) (AttachmentAsset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	preparation, ok := s.uploadSessions[uploadID]
	if !ok || preparation.ExpiresAt.Before(time.Now()) {
		return AttachmentAsset{}, errUploadNotFound
	}
	if !s.canAccessStore(user, preparation.StoreID) {
		return AttachmentAsset{}, errUnauthorizedStore
	}
	if strings.TrimSpace(orderID) == "" {
		return AttachmentAsset{}, errRecycleNotFound
	}
	order, ok := s.recycleOrders[orderID]
	if !ok {
		return AttachmentAsset{}, errRecycleNotFound
	}
	if !s.canAccessStore(user, order.StoreID) {
		return AttachmentAsset{}, errUnauthorizedStore
	}

	finalPublicURL := strings.TrimSpace(publicURL)
	if finalPublicURL == "" {
		finalPublicURL = preparation.PublicURL
	}
	if finalPublicURL == "" {
		finalPublicURL = "https://pending-upload.local/" + strings.TrimLeft(preparation.ObjectKey, "/")
	}
	finalThumbURL := strings.TrimSpace(thumbnailURL)
	if finalThumbURL == "" {
		finalThumbURL = finalPublicURL
	}
	asset := AttachmentAsset{
		ID:              "att-" + uploadID,
		OrgID:           order.OrgID,
		OrderID:         order.ID,
		StoreID:         order.StoreID,
		Category:        "recycle_photo",
		StorageProvider: preparation.Provider,
		ObjectKey:       preparation.ObjectKey,
		PublicURL:       finalPublicURL,
		ThumbnailURL:    finalThumbURL,
		FileName:        preparation.FileName,
		ContentType:     preparation.ContentType,
		SizeBytes:       preparation.SizeBytes,
		Source:          "prepared_upload",
		Status:          "archived",
		UploadedBy:      user.DisplayName,
		UploadedAt:      time.Now(),
	}

	replaced := false
	for index, existing := range order.Attachments {
		if existing.ID == asset.ID {
			order.Attachments[index] = asset
			replaced = true
			break
		}
	}
	if !replaced {
		order.Attachments = append(order.Attachments, asset)
	}
	urlExists := false
	for _, existing := range order.AttachmentURLs {
		if existing == asset.PublicURL {
			urlExists = true
			break
		}
	}
	if !urlExists {
		order.AttachmentURLs = append(order.AttachmentURLs, asset.PublicURL)
	}

	if s.persistence != nil {
		if err := s.persistence.saveRecycleOrder(context.Background(), order); err != nil {
			return AttachmentAsset{}, err
		}
		if err := s.persistence.saveRecycleAttachments(context.Background(), order.ID, order.Attachments); err != nil {
			return AttachmentAsset{}, err
		}
	}

	s.recycleOrders[order.ID] = order
	delete(s.uploadSessions, uploadID)
	return asset, nil
}

func (s *MockStore) rebuildRecycleAttachmentsLocked() {
	for orderID, order := range s.recycleOrders {
		order.Attachments = buildRecycleAttachmentAssets(order, order.AttachmentURLs, order.CreatedBy, order.CreatedAt)
		s.recycleOrders[orderID] = order
	}
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

func (s *MockStore) createCashierOrder(user UserAccount, storeID string, customerName string, customerPhone string, paymentMethod string, items []CashierOrderLine, remark string) (CashierOrder, error) {
	if !s.canAccessStore(user, storeID) {
		return CashierOrder{}, errUnauthorizedStore
	}
	store, ok := s.getStore(storeID)
	if !ok {
		return CashierOrder{}, errUnauthorizedStore
	}

	var normalized []CashierOrderLine
	var total float64
	status := "paid"
	paidAmount := 0.0
	normalizedPaymentMethod := normalizePaymentMethod(paymentMethod)
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
	if normalizedPaymentMethod == "暂不支付" {
		status = "pending"
	} else {
		paidAmount = round2(total)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := CashierOrder{
		ID:            fmt.Sprintf("cashier-order-%03d", s.cashierSeq),
		OrderNo:       fmt.Sprintf("CO%s%04d", time.Now().Format("20060102"), s.cashierSeq),
		OrgID:         user.OrgID,
		StoreID:       store.ID,
		StoreName:     store.Name,
		Status:        status,
		CustomerName:  strings.TrimSpace(customerName),
		CustomerPhone: strings.TrimSpace(customerPhone),
		PaymentMethod: normalizedPaymentMethod,
		TotalAmount:   round2(total),
		PaidAmount:    paidAmount,
		Remark:        remark,
		Items:         normalized,
		CreatedBy:     user.DisplayName,
		CreatedAt:     time.Now(),
	}
	s.cashierSeq++
	if s.persistence != nil {
		if err := s.persistence.saveCashierOrder(context.Background(), order); err != nil {
			return CashierOrder{}, err
		}
	}
	s.cashierOrders = append(s.cashierOrders, order)
	return order, nil
}

func normalizePaymentMethod(value string) string {
	method := strings.TrimSpace(value)
	if method == "" || method == "cash" {
		return "微信转账"
	}
	return method
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
	order.Attachments = buildRecycleAttachmentAssets(order, order.AttachmentURLs, user.DisplayName, order.CreatedAt)
	s.recycleSeq++
	if s.persistence != nil {
		if err := s.persistence.saveRecycleOrder(context.Background(), order); err != nil {
			return RecycleOrder{}, err
		}
		if err := s.persistence.saveRecycleAttachments(context.Background(), order.ID, order.Attachments); err != nil {
			return RecycleOrder{}, err
		}
	}
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
	existingAttachments := append([]AttachmentAsset(nil), order.Attachments...)
	order.Status = "confirmed"
	order.ConfirmedAmount = round2(confirmedAmount)
	order.AttachmentURLs = attachments
	order.Attachments = buildRecycleAttachmentAssetsPreservingProvider(order, attachments, existingAttachments, user.DisplayName, now)
	order.Remark = strings.TrimSpace(remark)
	order.ConfirmedAt = &now
	if s.persistence != nil {
		if err := s.persistence.saveRecycleOrder(context.Background(), order); err != nil {
			return RecycleOrder{}, err
		}
		if err := s.persistence.saveRecycleAttachments(context.Background(), order.ID, order.Attachments); err != nil {
			return RecycleOrder{}, err
		}
	}
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
	if strings.TrimSpace(update.Storage.Provider) != "" {
		s.settings.Storage.Provider = strings.TrimSpace(update.Storage.Provider)
	}
	if strings.TrimSpace(update.Storage.Bucket) != "" {
		s.settings.Storage.Bucket = strings.TrimSpace(update.Storage.Bucket)
	}
	if strings.TrimSpace(update.Storage.Region) != "" {
		s.settings.Storage.Region = strings.TrimSpace(update.Storage.Region)
	}
	if strings.TrimSpace(update.Storage.PublicBaseURL) != "" {
		s.settings.Storage.PublicBaseURL = strings.TrimSpace(update.Storage.PublicBaseURL)
	}
	if strings.TrimSpace(update.Storage.PathPrefix) != "" {
		s.settings.Storage.PathPrefix = strings.TrimSpace(update.Storage.PathPrefix)
	}
	if strings.TrimSpace(update.Storage.UploadStrategy) != "" {
		s.settings.Storage.UploadStrategy = strings.TrimSpace(update.Storage.UploadStrategy)
	}
	if strings.TrimSpace(update.Storage.StatusDescription) != "" {
		s.settings.Storage.StatusDescription = strings.TrimSpace(update.Storage.StatusDescription)
	}
	s.settings.Storage.Enabled = update.Storage.Enabled
	s.settings.Storage.CallbackEnabled = update.Storage.CallbackEnabled

	if update.FeatureFlags != nil {
		for key, value := range update.FeatureFlags {
			s.settings.FeatureFlags[key] = value
		}
	}
	applySystemSettingDefaults(&s.settings)

	s.settings.UpdatedBy = user.DisplayName
	s.settings.UpdatedAt = time.Now()
	s.adminSystemProfile = s.buildDynamicSystemProfileLocked()
	if s.persistence != nil {
		_ = s.persistSettingsLocked()
		_ = s.persistence.saveConfig(context.Background(), configKeyAdminSystem, s.adminSystemProfile)
	}
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

func (s *MockStore) inventorySummary(user UserAccount) InventorySummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visibleStores, storeSet := visibleStoresForUserLocked(s.stores, user)
	summary := InventorySummary{
		OrgID:             user.OrgID,
		VisibleStoreCount: len(visibleStores),
		Items:             make([]InventoryItem, 0, len(s.catalogProducts)),
	}

	for _, product := range s.catalogProducts {
		if !productVisibleToUser(product, user, storeSet) {
			continue
		}
		estimatedValue := estimateCatalogProductValue(product)
		stockStatus := product.StockStatus
		if stockStatus == "" {
			stockStatus = deriveStockStatus(product)
		}
		item := InventoryItem{
			ProductID:            product.ID,
			Name:                 product.Name,
			SKU:                  product.SKU,
			Category:             product.Category,
			CategoryTab:          product.CategoryTab,
			ImageURL:             product.ImageURL,
			Status:               product.Status,
			StockStatus:          stockStatus,
			Inventory:            product.Inventory,
			Stores:               append([]string(nil), product.Stores...),
			BenchPrice:           round2(product.BenchPrice),
			RetailPrice:          round2(product.RetailPrice),
			GramWeight:           round2(product.GramWeight),
			EstimatedRetailValue: estimatedValue,
		}
		summary.Items = append(summary.Items, item)
		summary.TotalSKU++
		summary.TotalInventory += product.Inventory
		summary.EstimatedRetailValue += estimatedValue
		if product.Status == "active" {
			summary.ActiveSKU++
		}
		if product.Inventory <= 0 || stockStatus == "disabled" || stockStatus == "out" {
			summary.OutOfStockSKU++
			continue
		}
		if product.Inventory <= 4 || stockStatus == "low" {
			summary.LowStockSKU++
		}
	}
	summary.EstimatedRetailValue = round2(summary.EstimatedRetailValue)
	return summary
}

func (s *MockStore) dailyReport(user UserAccount, date string) DailyReportSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if strings.TrimSpace(date) == "" {
		date = time.Now().Format("2006-01-02")
	}
	visibleStores, storeSet := visibleStoresForUserLocked(s.stores, user)
	storeMetrics := make(map[string]DailyReportStoreMetric, len(visibleStores))
	for _, store := range visibleStores {
		storeMetrics[store.ID] = DailyReportStoreMetric{
			StoreID:   store.ID,
			StoreName: store.Name,
		}
	}

	report := DailyReportSummary{
		OrgID:             user.OrgID,
		Date:              date,
		DataScope:         user.DataScope,
		VisibleStoreCount: len(visibleStores),
	}

	for _, order := range s.cashierOrders {
		if _, ok := storeSet[order.StoreID]; !ok || order.CreatedAt.Format("2006-01-02") != date {
			continue
		}
		report.CashierOrderCount++
		report.CashierAmount += order.PaidAmount
		metric := storeMetrics[order.StoreID]
		metric.CashierOrderCount++
		metric.CashierAmount += order.PaidAmount
		storeMetrics[order.StoreID] = metric
	}

	for _, order := range s.recycleOrders {
		if _, ok := storeSet[order.StoreID]; !ok || order.CreatedAt.Format("2006-01-02") != date {
			continue
		}
		switch order.Status {
		case "draft":
			report.RecycleDraftOrderCount++
		case "confirmed":
			report.RecycleConfirmedOrderCount++
			report.RecycleConfirmedAmount += order.ConfirmedAmount
		}
		var weight float64
		for _, item := range order.Items {
			weight += item.WeightGram
		}
		report.RecycleWeightGram += weight
		metric := storeMetrics[order.StoreID]
		metric.RecycleOrderCount++
		metric.RecycleAmount += order.ConfirmedAmount
		metric.RecycleWeightGram += weight
		storeMetrics[order.StoreID] = metric
	}

	report.CashierAmount = round2(report.CashierAmount)
	report.RecycleConfirmedAmount = round2(report.RecycleConfirmedAmount)
	report.RecycleWeightGram = round2(report.RecycleWeightGram)
	report.StoreMetrics = make([]DailyReportStoreMetric, 0, len(visibleStores))
	for _, store := range visibleStores {
		metric := storeMetrics[store.ID]
		metric.CashierAmount = round2(metric.CashierAmount)
		metric.RecycleAmount = round2(metric.RecycleAmount)
		metric.RecycleWeightGram = round2(metric.RecycleWeightGram)
		report.StoreMetrics = append(report.StoreMetrics, metric)
	}
	return report
}

func visibleStoresForUserLocked(stores []StoreInfo, user UserAccount) ([]StoreInfo, map[string]struct{}) {
	if user.DataScope == "org_all" {
		storeSet := make(map[string]struct{}, len(stores))
		for _, store := range stores {
			storeSet[store.ID] = struct{}{}
		}
		return append([]StoreInfo(nil), stores...), storeSet
	}
	allow := make(map[string]struct{}, len(user.StoreIDs))
	for _, storeID := range user.StoreIDs {
		allow[storeID] = struct{}{}
	}
	visible := make([]StoreInfo, 0, len(user.StoreIDs))
	storeSet := make(map[string]struct{}, len(user.StoreIDs))
	for _, store := range stores {
		if _, ok := allow[store.ID]; ok {
			visible = append(visible, store)
			storeSet[store.ID] = struct{}{}
		}
	}
	return visible, storeSet
}

func productVisibleToUser(product CatalogProduct, user UserAccount, storeSet map[string]struct{}) bool {
	if user.DataScope == "org_all" {
		return true
	}
	for _, storeID := range product.StoreIDs {
		if _, ok := storeSet[storeID]; ok {
			return true
		}
	}
	return false
}

func estimateCatalogProductValue(product CatalogProduct) float64 {
	if product.Inventory <= 0 {
		return 0
	}
	unitValue := product.RetailPrice
	if unitValue <= 0 && product.BenchPrice > 0 && product.GramWeight > 0 {
		unitValue = product.BenchPrice * product.GramWeight
	}
	return round2(unitValue * float64(product.Inventory))
}

func deriveStockStatus(product CatalogProduct) string {
	if product.Status == "disabled" {
		return "disabled"
	}
	if product.Inventory <= 0 {
		return "out"
	}
	if product.Inventory <= 4 {
		return "low"
	}
	return "normal"
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func (s *MockStore) close() error {
	if s.persistence != nil {
		return s.persistence.Close()
	}
	return nil
}

func (s *MockStore) bootstrapPersistence() error {
	ctx := context.Background()

	if err := s.bootstrapBaseConfigs(ctx); err != nil {
		return err
	}

	cashierOrders, err := s.persistence.loadCashierOrders(ctx)
	if err != nil {
		return err
	}
	if len(cashierOrders) == 0 {
		for _, order := range s.cashierOrders {
			if err := s.persistence.saveCashierOrder(ctx, order); err != nil {
				return err
			}
		}
	} else {
		s.cashierOrders = cashierOrders
	}

	recycleOrders, err := s.persistence.loadRecycleOrders(ctx)
	if err != nil {
		return err
	}
	if len(recycleOrders) == 0 {
		for _, order := range s.recycleOrders {
			if err := s.persistence.saveRecycleOrder(ctx, order); err != nil {
				return err
			}
			if err := s.persistence.saveRecycleAttachments(ctx, order.ID, order.Attachments); err != nil {
				return err
			}
		}
	} else {
		s.recycleOrders = recycleOrders
	}

	attachmentsByOrder, err := s.persistence.loadRecycleAttachments(ctx)
	if err != nil {
		return err
	}
	if len(attachmentsByOrder) == 0 {
		s.rebuildRecycleAttachmentsLocked()
		for orderID, order := range s.recycleOrders {
			if err := s.persistence.saveRecycleAttachments(ctx, orderID, order.Attachments); err != nil {
				return err
			}
		}
	} else {
		for orderID, assets := range attachmentsByOrder {
			order, ok := s.recycleOrders[orderID]
			if !ok {
				continue
			}
			order.Attachments = append([]AttachmentAsset(nil), assets...)
			if len(order.AttachmentURLs) == 0 {
				urls := make([]string, 0, len(assets))
				for _, asset := range assets {
					urls = append(urls, asset.PublicURL)
				}
				order.AttachmentURLs = urls
			}
			s.recycleOrders[orderID] = order
		}
	}

	s.cashierSeq = nextSequenceFromCashierOrders(s.cashierOrders)
	s.recycleSeq = nextSequenceFromRecycleOrders(s.recycleOrders)
	if err := s.bootstrapAdminConfigs(ctx); err != nil {
		return err
	}
	return nil
}

func (s *MockStore) bootstrapBaseConfigs(ctx context.Context) error {
	if s.persistence == nil {
		return nil
	}

	var settings SystemSettings
	found, err := s.persistence.loadConfig(ctx, configKeyBaseSettings, &settings)
	if err != nil {
		return err
	}
	if found {
		s.settings = settings
		applySystemSettingDefaults(&s.settings)
	} else if err := s.persistence.saveConfig(ctx, configKeyBaseSettings, s.settings); err != nil {
		return err
	}

	var stores []StoreInfo
	found, err = s.persistence.loadConfig(ctx, configKeyBaseStores, &stores)
	if err != nil {
		return err
	}
	if found {
		s.stores = stores
	} else if err := s.persistence.saveConfig(ctx, configKeyBaseStores, s.stores); err != nil {
		return err
	}

	var roles []RoleTemplate
	found, err = s.persistence.loadConfig(ctx, configKeyBaseRoles, &roles)
	if err != nil {
		return err
	}
	if found {
		s.roles = roles
	} else if err := s.persistence.saveConfig(ctx, configKeyBaseRoles, s.roles); err != nil {
		return err
	}

	var users []UserAccount
	found, err = s.persistence.loadConfig(ctx, configKeyBaseUsers, &users)
	if err != nil {
		return err
	}
	if found {
		restoredUsers, restored := restoreMissingUserPasswords(users, s.usersByUsername)
		users = restoredUsers
		s.rebuildUserMaps(users)
		if restored {
			if err := s.persistence.saveConfig(ctx, configKeyBaseUsers, s.usersSliceLocked()); err != nil {
				return err
			}
		}
	} else if err := s.persistence.saveConfig(ctx, configKeyBaseUsers, s.usersSliceLocked()); err != nil {
		return err
	}

	var members []MemberProfile
	found, err = s.persistence.loadConfig(ctx, configKeyBaseMembers, &members)
	if err != nil {
		return err
	}
	if found {
		s.members = members
	} else if err := s.persistence.saveConfig(ctx, configKeyBaseMembers, s.members); err != nil {
		return err
	}

	var products []CatalogProduct
	found, err = s.persistence.loadConfig(ctx, configKeyBaseProducts, &products)
	if err != nil {
		return err
	}
	if found {
		s.catalogProducts = products
	} else if err := s.persistence.saveConfig(ctx, configKeyBaseProducts, s.catalogProducts); err != nil {
		return err
	}

	return nil
}

func (s *MockStore) bootstrapAdminConfigs(ctx context.Context) error {
	if s.persistence == nil {
		return nil
	}

	var adminRoles []AdminRoleTemplate
	found, err := s.persistence.loadConfig(ctx, configKeyAdminRoles, &adminRoles)
	if err != nil {
		return err
	}
	if found {
		s.adminRoles = adminRoles
	} else if err := s.persistence.saveConfig(ctx, configKeyAdminRoles, s.adminRoles); err != nil {
		return err
	}

	var printTemplate AdminPrintTemplate
	found, err = s.persistence.loadConfig(ctx, configKeyAdminPrint, &printTemplate)
	if err != nil {
		return err
	}
	if found {
		s.adminPrintTemplate = printTemplate
	} else if err := s.persistence.saveConfig(ctx, configKeyAdminPrint, s.adminPrintTemplate); err != nil {
		return err
	}

	var systemProfile AdminSystemProfile
	found, err = s.persistence.loadConfig(ctx, configKeyAdminSystem, &systemProfile)
	if err != nil {
		return err
	}
	if found {
		s.adminSystemProfile = systemProfile
	} else if err := s.persistence.saveConfig(ctx, configKeyAdminSystem, s.adminSystemProfile); err != nil {
		return err
	}

	var auditLogs []AdminAuditLogRecord
	found, err = s.persistence.loadConfig(ctx, configKeyAdminAudit, &auditLogs)
	if err != nil {
		return err
	}
	if found {
		s.adminAuditLogs = auditLogs
	} else if err := s.persistence.saveConfig(ctx, configKeyAdminAudit, s.adminAuditLogs); err != nil {
		return err
	}
	return nil
}

func (s *MockStore) rebuildUserMaps(users []UserAccount) {
	s.usersByID = make(map[string]UserAccount, len(users))
	s.usersByUsername = make(map[string]UserAccount, len(users))
	for _, user := range users {
		s.usersByID[user.ID] = user
		s.usersByUsername[user.Username] = user
	}
}

func restoreMissingUserPasswords(users []UserAccount, defaults map[string]UserAccount) ([]UserAccount, bool) {
	restored := false
	items := make([]UserAccount, 0, len(users))
	for _, user := range users {
		if strings.TrimSpace(user.Password) == "" {
			if fallback, ok := defaults[user.Username]; ok && strings.TrimSpace(fallback.Password) != "" {
				user.Password = fallback.Password
				restored = true
			}
		}
		items = append(items, user)
	}
	return items, restored
}

func (s *MockStore) usersSliceLocked() []UserAccount {
	items := make([]UserAccount, 0, len(s.usersByID))
	for _, user := range s.usersByID {
		items = append(items, user)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *MockStore) persistStoresLocked() error {
	return s.persistence.saveConfig(context.Background(), configKeyBaseStores, s.stores)
}

func (s *MockStore) persistUsersLocked() error {
	return s.persistence.saveConfig(context.Background(), configKeyBaseUsers, s.usersSliceLocked())
}

func (s *MockStore) persistProductsLocked() error {
	return s.persistence.saveConfig(context.Background(), configKeyBaseProducts, s.catalogProducts)
}

func (s *MockStore) persistSettingsLocked() error {
	return s.persistence.saveConfig(context.Background(), configKeyBaseSettings, s.settings)
}

func (s *MockStore) persistAdminRolesLocked() error {
	return s.persistence.saveConfig(context.Background(), configKeyAdminRoles, s.adminRoles)
}

func (s *MockStore) persistAdminAuditLogsLocked() error {
	return s.persistence.saveConfig(context.Background(), configKeyAdminAudit, s.adminAuditLogs)
}

func nextSequenceFromCashierOrders(orders []CashierOrder) int {
	maxSeq := 1
	for _, order := range orders {
		maxSeq = max(maxSeq, numericSuffix(order.ID)+1)
	}
	return maxSeq
}

func nextSequenceFromRecycleOrders(orders map[string]RecycleOrder) int {
	maxSeq := 1
	for _, order := range orders {
		maxSeq = max(maxSeq, numericSuffix(order.ID)+1)
	}
	return maxSeq
}

func numericSuffix(value string) int {
	if idx := strings.LastIndex(value, "-"); idx >= 0 && idx < len(value)-1 {
		parsed, err := strconv.Atoi(value[idx+1:])
		if err == nil {
			return parsed
		}
	}
	return 0
}
