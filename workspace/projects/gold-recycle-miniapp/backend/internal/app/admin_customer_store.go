/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: admin_customer_store.go
 * 功能描述: 管理后台「顾客端内容管理」数据访问层（Banner、首页文案、款式、门店扩展、预约规则、预约记录、素材）
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
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
	errBannerNotFound          = errors.New("banner not found")
	errBannerLimit             = errors.New("enabled banner limit reached")
	errInvalidAppointmentRules = errors.New("invalid appointment rules")
	errInvalidStyleInput       = errors.New("invalid style input")
	errInvalidRecycleInfo      = errors.New("invalid recycle info")
	errRecycleInfoV1Boundary   = errors.New("recycle info violates v1 boundary")
	errUploadAssetNotFound     = errors.New("upload asset not found")
	errUploadAssetInUse        = errors.New("upload asset still referenced")
)

// --- 首页配置（Banner + 文案） ---

func (s *MockStore) persistCustomerHomeLocked() error {
	if s.persistence == nil {
		return nil
	}
	return s.persistence.saveConfig(context.Background(), configKeyCustomerHome, s.customerHomeConfig)
}

// normalizeCustomerHomeConfigLocked 存量数据兼容：补 Banner ID、默认启用、排序
func normalizeCustomerHomeConfigLocked(cfg *CustomerHomeResponse) {
	for i := range cfg.Banners {
		if cfg.Banners[i].ID == "" {
			cfg.Banners[i].ID = fmt.Sprintf("banner-%d-%s", i+1, generateID())
		}
		if cfg.Banners[i].Enabled == nil {
			enabled := true
			cfg.Banners[i].Enabled = &enabled
		}
	}
	sort.SliceStable(cfg.Banners, func(i, j int) bool {
		return cfg.Banners[i].SortOrder < cfg.Banners[j].SortOrder
	})
	if cfg.NearbyStoreRule == nil {
		cfg.NearbyStoreRule = &NearbyStoreRule{Mode: "distance", Count: 3}
	}
	if cfg.Banners == nil {
		cfg.Banners = []CustomerBanner{}
	}
}

func (s *MockStore) adminListBanners() []CustomerBanner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.customerHomeConfig
	normalizeCustomerHomeConfigLocked(&cfg)
	return append([]CustomerBanner(nil), cfg.Banners...)
}

func (s *MockStore) adminSaveBanner(user UserAccount, bannerID string, req AdminBannerRequest) (CustomerBanner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	enabled := bannerEnabled(req.Enabled)
	if bannerID == "" { // 新建
		if enabled {
			count := 0
			for _, b := range s.customerHomeConfig.Banners {
				if bannerEnabled(b.Enabled) {
					count++
				}
			}
			if count >= 8 {
				return CustomerBanner{}, errBannerLimit
			}
		}
		bannerID = fmt.Sprintf("banner-%s", generateID())
		s.customerHomeConfig.Banners = append(s.customerHomeConfig.Banners, CustomerBanner{
			ID: bannerID, Enabled: &enabled,
		})
	}
	idx := -1
	for i, b := range s.customerHomeConfig.Banners {
		if b.ID == bannerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return CustomerBanner{}, errBannerNotFound
	}
	b := s.customerHomeConfig.Banners[idx]
	b.Title = strings.TrimSpace(req.Title)
	b.Subtitle = strings.TrimSpace(req.Subtitle)
	b.ImageURL = strings.TrimSpace(req.ImageURL)
	b.LinkType = strings.TrimSpace(req.LinkType)
	b.LinkTarget = strings.TrimSpace(req.LinkTarget)
	b.SortOrder = req.SortOrder
	b.Enabled = &enabled
	if b.LinkType == "" {
		b.LinkType = "none"
	}
	if b.LinkType == "none" {
		b.LinkTarget = ""
	}
	s.customerHomeConfig.Banners[idx] = b
	normalizeCustomerHomeConfigLocked(&s.customerHomeConfig)
	if err := s.persistCustomerHomeLocked(); err != nil {
		return CustomerBanner{}, err
	}
	s.appendAuditLogLocked("顾客端配置", "保存首页Banner", user.DisplayName, "success", "low", fmt.Sprintf("Banner「%s」已保存", b.Title))
	return b, nil
}

func (s *MockStore) adminDeleteBanner(user UserAccount, bannerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, b := range s.customerHomeConfig.Banners {
		if b.ID == bannerID {
			s.customerHomeConfig.Banners = append(s.customerHomeConfig.Banners[:i], s.customerHomeConfig.Banners[i+1:]...)
			if err := s.persistCustomerHomeLocked(); err != nil {
				return err
			}
			s.appendAuditLogLocked("顾客端配置", "删除首页Banner", user.DisplayName, "success", "medium", fmt.Sprintf("Banner「%s」已删除", b.Title))
			return nil
		}
	}
	return errBannerNotFound
}

func (s *MockStore) adminGetHomeConfig() CustomerHomeResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.customerHomeConfig
	normalizeCustomerHomeConfigLocked(&cfg)
	return cfg
}

func (s *MockStore) adminUpdateHomeConfig(user UserAccount, req AdminHomeConfigRequest) (CustomerHomeResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := s.customerHomeConfig
	cfg.BrandName = strings.TrimSpace(req.BrandName)
	cfg.BrandSlogan1 = strings.TrimSpace(req.BrandSlogan1)
	cfg.BrandSlogan2 = strings.TrimSpace(req.BrandSlogan2)
	cfg.ServiceCopy = strings.TrimSpace(req.ServiceCopy)
	cfg.EntryStyleText = strings.TrimSpace(req.EntryStyleText)
	cfg.EntryFeeText = strings.TrimSpace(req.EntryFeeText)
	cfg.ServicePhone = strings.TrimSpace(req.ServicePhone)
	cfg.AppointmentNotes = strings.TrimSpace(req.AppointmentNotes)
	cfg.ServiceIntro = strings.TrimSpace(req.ServiceIntro)
	cfg.LocationPermissionNote = strings.TrimSpace(req.LocationPermissionNote)
	if req.NearbyStoreRule != nil {
		mode := req.NearbyStoreRule.Mode
		if mode != "distance" && mode != "city" && mode != "all" {
			mode = "distance"
		}
		count := req.NearbyStoreRule.Count
		if count < 1 || count > 5 {
			count = 3
		}
		cfg.NearbyStoreRule = &NearbyStoreRule{Mode: mode, Count: count}
	}
	normalizeCustomerHomeConfigLocked(&cfg)
	s.customerHomeConfig = cfg
	if err := s.persistCustomerHomeLocked(); err != nil {
		return CustomerHomeResponse{}, err
	}
	s.appendAuditLogLocked("顾客端配置", "更新首页文案", user.DisplayName, "success", "low", fmt.Sprintf("品牌文案已更新：%s", cfg.BrandName))
	return cfg, nil
}

// --- 款式（商品）管理：含扩展字段，与 catalog_products 同源 ---

func buildAdminCustomerProduct(p CatalogProduct) AdminCustomerProductRecord {
	images := p.Images
	if images == nil {
		images = []string{}
	}
	detailImages := p.DetailImages
	if detailImages == nil {
		detailImages = []string{}
	}
	serviceTypes := p.ApplicableServiceTypes
	if serviceTypes == nil {
		serviceTypes = []string{}
	}
	tags := p.Tags
	if tags == nil {
		tags = []string{}
	}
	storeIDs := p.StoreIDs
	if storeIDs == nil {
		storeIDs = []string{}
	}
	status := p.Status
	if status != "active" {
		status = "inactive"
	}
	return AdminCustomerProductRecord{
		ID: p.ID, Name: p.Name, SKU: p.SKU, Category: p.Category, CategoryTab: p.CategoryTab,
		ImageURL: p.ImageURL, Images: images, DetailImages: detailImages, Purity: p.Purity,
		RetailPrice: p.RetailPrice, GramWeight: p.GramWeight,
		LaborFeeRef: p.LaborFeeRef, Description: p.Description, LaborFeeNote: p.LaborFeeNote,
		ApplicableServiceTypes: serviceTypes, RecommendedStoreRule: p.RecommendedStoreRule,
		SortOrder: p.SortOrder, IsRecommended: p.IsRecommended, IsHot: p.IsHot,
		Status: status, Tags: tags, StoreIDs: storeIDs,
	}
}

func (s *MockStore) adminListCustomerProducts(category, status, keyword string, page, pageSize int) ([]AdminCustomerProductRecord, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []AdminCustomerProductRecord
	kw := strings.ToLower(keyword)
	for _, p := range s.catalogProducts {
		if category != "" && p.Category != category {
			continue
		}
		if status != "" {
			want := status == "active"
			is := p.Status == "active"
			if want != is {
				continue
			}
		}
		if kw != "" {
			hit := strings.Contains(strings.ToLower(p.Name), kw) ||
				strings.Contains(strings.ToLower(p.Category), kw) ||
				strings.Contains(strings.ToLower(p.SKU), kw)
			if !hit {
				continue
			}
		}
		items = append(items, buildAdminCustomerProduct(p))
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].Name < items[j].Name
	})
	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return []AdminCustomerProductRecord{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

func (s *MockStore) adminGetCustomerProduct(productID string) (AdminCustomerProductRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.catalogProducts {
		if p.ID == productID {
			return buildAdminCustomerProduct(p), true
		}
	}
	return AdminCustomerProductRecord{}, false
}

func (s *MockStore) adminSaveCustomerProduct(user UserAccount, productID string, req AdminCustomerProductRecord) (AdminCustomerProductRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := strings.TrimSpace(req.Name)
	if name == "" || strings.TrimSpace(req.ImageURL) == "" {
		return AdminCustomerProductRecord{}, errInvalidStyleInput
	}
	if req.Status != "active" && req.Status != "inactive" {
		req.Status = "active"
	}
	if len(req.Images) > 0 && strings.TrimSpace(req.ImageURL) != "" && req.Images[0] != req.ImageURL {
		req.Images = append([]string{req.ImageURL}, req.Images...)
	}
	if productID == "" {
		productID = fmt.Sprintf("prod-%s", generateID())
		s.catalogProducts = append(s.catalogProducts, CatalogProduct{ID: productID, OrgID: "org-gold-v1", Status: "active"})
	}
	idx := -1
	for i, p := range s.catalogProducts {
		if p.ID == productID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return AdminCustomerProductRecord{}, errProductNotFound
	}
	p := s.catalogProducts[idx]
	p.Name = name
	p.Category = strings.TrimSpace(req.Category)
	p.CategoryTab = strings.TrimSpace(req.CategoryTab)
	p.ImageURL = strings.TrimSpace(req.ImageURL)
	p.Images = req.Images
	p.DetailImages = req.DetailImages
	p.Purity = strings.TrimSpace(req.Purity)
	p.RetailPrice = req.RetailPrice
	p.GramWeight = req.GramWeight
	p.LaborFeeRef = strings.TrimSpace(req.LaborFeeRef)
	p.Description = req.Description
	p.LaborFeeNote = req.LaborFeeNote
	p.ApplicableServiceTypes = req.ApplicableServiceTypes
	p.RecommendedStoreRule = strings.TrimSpace(req.RecommendedStoreRule)
	p.SortOrder = req.SortOrder
	p.IsRecommended = req.IsRecommended
	p.IsHot = req.IsHot
	p.Status = req.Status
	p.Tags = req.Tags
	p.StoreIDs = req.StoreIDs
	s.catalogProducts[idx] = p
	if s.persistence != nil {
		_ = s.persistProductsLocked()
	}
	s.appendAuditLogLocked("款式管理", "保存款式", user.DisplayName, "success", "medium", fmt.Sprintf("款式「%s」已保存", p.Name))
	return buildAdminCustomerProduct(p), nil
}

func (s *MockStore) adminDeleteCustomerProduct(user UserAccount, productID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.catalogProducts {
		if p.ID == productID {
			s.catalogProducts = append(s.catalogProducts[:i], s.catalogProducts[i+1:]...)
			if s.persistence != nil {
				_ = s.persistProductsLocked()
			}
			s.appendAuditLogLocked("款式管理", "删除款式", user.DisplayName, "warning", "medium", fmt.Sprintf("款式「%s」已删除", p.Name))
			return nil
		}
	}
	return errProductNotFound
}

// --- 门店顾客端扩展配置 ---

// adminGetStoreCustomerConfig 读取门店顾客端扩展配置
func (s *MockStore) adminGetStoreCustomerConfig(storeID string) (AdminCustomerStoreConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, store := range s.stores {
		if store.ID != storeID {
			continue
		}
		return AdminCustomerStoreConfig{
			ImageURL: store.ImageURL, AppointmentEnabled: store.AppointmentEnabled,
			ServiceTags: store.ServiceTags, SortOrder: store.SortOrder,
			Longitude: &store.Longitude, Latitude: &store.Latitude,
			ContactPhone: store.ContactPhone, BusinessHours: store.BusinessHours,
			Address: store.Address, Status: store.Status,
		}, true
	}
	return AdminCustomerStoreConfig{}, false
}

func (s *MockStore) adminUpdateStoreCustomerConfig(user UserAccount, storeID string, req AdminCustomerStoreConfig) (AdminCustomerStoreConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.canAccessStore(user, storeID) {
		return AdminCustomerStoreConfig{}, errUnauthorizedStore
	}
	for index, store := range s.stores {
		if store.ID != storeID {
			continue
		}
		if req.ImageURL != "" {
			store.ImageURL = strings.TrimSpace(req.ImageURL)
		}
		if req.AppointmentEnabled != nil {
			store.AppointmentEnabled = req.AppointmentEnabled
		}
		if req.ServiceTags != nil {
			store.ServiceTags = req.ServiceTags
		}
		if req.SortOrder != 0 {
			store.SortOrder = req.SortOrder
		}
		if req.Longitude != nil {
			store.Longitude = *req.Longitude
		}
		if req.Latitude != nil {
			store.Latitude = *req.Latitude
		}
		if req.ContactPhone != "" {
			store.ContactPhone = strings.TrimSpace(req.ContactPhone)
		}
		if req.BusinessHours != "" {
			store.BusinessHours = strings.TrimSpace(req.BusinessHours)
		}
		if req.Address != "" {
			store.Address = strings.TrimSpace(req.Address)
		}
		if req.Status == "active" || req.Status == "inactive" {
			store.Status = req.Status
		}
		s.stores[index] = store
		if s.persistence != nil {
			_ = s.persistStoresLocked()
		}
		enabled := store.AppointmentEnabled == nil || *store.AppointmentEnabled
		s.appendAuditLogLocked("门店管理", "更新门店顾客端配置", user.DisplayName, "success", "medium",
			fmt.Sprintf("%s 已更新（预约开关：%v）", store.Name, enabled))
		return AdminCustomerStoreConfig{
			ImageURL: store.ImageURL, AppointmentEnabled: store.AppointmentEnabled,
			ServiceTags: store.ServiceTags, SortOrder: store.SortOrder,
			Longitude: &store.Longitude, Latitude: &store.Latitude,
			ContactPhone: store.ContactPhone, BusinessHours: store.BusinessHours,
			Address: store.Address, Status: store.Status,
		}, nil
	}
	return AdminCustomerStoreConfig{}, errAdminStoreNotFound
}

// --- 预约规则 ---

func (s *MockStore) appointmentRulesSnapshot() AppointmentRules {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.appointmentRulesLocked()
}

// appointmentRulesLocked 返回预约规则快照。调用方必须已经持有 s.mu 的读锁或写锁。
// 单独抽出该方法，避免在持锁状态下再次调用 appointmentRulesSnapshot 造成锁重入死锁。
func (s *MockStore) appointmentRulesLocked() AppointmentRules {
	rules := s.appointmentRules
	if rules.SlotMinutes == 0 {
		rules = defaultAppointmentRules()
	}
	return rules
}

func (s *MockStore) adminGetAppointmentRules() AppointmentRules {
	rules := s.appointmentRulesSnapshot()
	rules.BookableServiceTypes = append([]string(nil), rules.BookableServiceTypes...)
	return rules
}

func (s *MockStore) adminUpdateAppointmentRules(user UserAccount, req AdminAppointmentRulesRequest) (AppointmentRules, error) {
	normalized, err := validateAppointmentRules(AppointmentRules{
		BookableServiceTypes: req.BookableServiceTypes,
		BookableDays:         req.BookableDays,
		SlotMinutes:          req.SlotMinutes,
		OpenTime:             req.OpenTime,
		CloseTime:            req.CloseTime,
		SameDayLeadMinutes:   req.SameDayLeadMinutes,
		CancelLeadMinutes:    req.CancelLeadMinutes,
		SlotCapacity:         req.SlotCapacity,
	})
	if err != nil {
		return AppointmentRules{}, err
	}
	normalized.UpdatedAt = time.Now()

	s.mu.Lock()
	previous := s.appointmentRules
	s.appointmentRules = normalized
	if s.persistence != nil {
		if err := s.persistence.saveConfig(context.Background(), configKeyAppointmentRules, normalized); err != nil {
			s.appointmentRules = previous
			return AppointmentRules{}, err
		}
	}
	s.appendAuditLogLocked("顾客端配置", "更新预约规则", user.DisplayName, "success", "high",
		fmt.Sprintf("预约规则已更新（粒度 %d 分钟 / 容量 %d / 可约 %d 天 / 取消提前 %d 分钟）",
			normalized.SlotMinutes, normalized.SlotCapacity, normalized.BookableDays, normalized.CancelLeadMinutes))
	s.mu.Unlock()
	return normalized, nil
}

// --- 回收服务介绍管理 ---

// adminGetCustomerRecycleInfo 读取回收服务介绍（空数据时返回默认值）
func (s *MockStore) adminGetCustomerRecycleInfo() CustomerRecycleInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info := s.customerRecycleInfo
	if info.Title == "" || (len(info.Process) == 0 && len(info.Services) == 0) {
		return defaultCustomerRecycleInfo()
	}
	return info
}

// adminUpdateCustomerRecycleInfo 保存回收服务介绍（归一化后持久化到 app_configs）
func (s *MockStore) adminUpdateCustomerRecycleInfo(user UserAccount, req CustomerRecycleInfo) (CustomerRecycleInfo, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return CustomerRecycleInfo{}, errInvalidRecycleInfo
	}
	normalized := CustomerRecycleInfo{
		Title:    title,
		Intro:    strings.TrimSpace(req.Intro),
		ImageURL: strings.TrimSpace(req.ImageURL),
	}
	for _, step := range req.Process {
		t := strings.TrimSpace(step.Title)
		if t == "" {
			continue
		}
		normalized.Process = append(normalized.Process, RecycleProcessStep{
			Step:  step.Step,
			Title: t,
			Desc:  strings.TrimSpace(step.Desc),
		})
	}
	for i := range normalized.Process {
		normalized.Process[i].Step = i + 1
	}
	for _, svc := range req.Services {
		t := strings.TrimSpace(svc.Title)
		if t == "" {
			continue
		}
		normalized.Services = append(normalized.Services, RecycleServiceItem{
			Icon:  strings.TrimSpace(svc.Icon),
			Title: t,
			Desc:  strings.TrimSpace(svc.Desc),
		})
	}
	for _, n := range req.Notices {
		line := strings.TrimSpace(n)
		if line != "" {
			normalized.Notices = append(normalized.Notices, line)
		}
	}
	if normalized.Process == nil {
		normalized.Process = []RecycleProcessStep{}
	}
	if normalized.Services == nil {
		normalized.Services = []RecycleServiceItem{}
	}
	if normalized.Notices == nil {
		normalized.Notices = []string{}
	}
	if err := validateCustomerRecycleInfoV1(normalized); err != nil {
		return CustomerRecycleInfo{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	previous := s.customerRecycleInfo
	s.customerRecycleInfo = normalized
	if s.persistence != nil {
		if err := s.persistence.saveConfig(context.Background(), configKeyCustomerRecycleInfo, normalized); err != nil {
			s.customerRecycleInfo = previous
			return CustomerRecycleInfo{}, err
		}
	}
	s.appendAuditLogLocked("顾客端配置", "更新回收服务介绍", user.DisplayName, "success", "medium",
		fmt.Sprintf("回收服务介绍已更新：%s（流程 %d 步 / 服务 %d 项 / 须知 %d 条）",
			normalized.Title, len(normalized.Process), len(normalized.Services), len(normalized.Notices)))
	return normalized, nil
}

// validateCustomerRecycleInfoV1 防止 V1 回收介绍变成金价、估价或线上结算入口。
// 顾客端 V1 只允许展示服务说明和到店流程；后台内容即使由管理员填写，也不能绕过产品边界。
func validateCustomerRecycleInfoV1(info CustomerRecycleInfo) error {
	forbiddenTerms := []string{
		"金价",
		"估价",
		"报价",
		"估值",
		"结算",
		"即时到账",
		"上门",
		"邮寄",
		"在线估",
		"在线回收",
		"支付",
		"付款",
	}
	values := []string{info.Title, info.Intro}
	for _, step := range info.Process {
		values = append(values, step.Title, step.Desc)
	}
	for _, service := range info.Services {
		values = append(values, service.Title, service.Desc)
	}
	values = append(values, info.Notices...)
	for _, value := range values {
		for _, forbidden := range forbiddenTerms {
			if strings.Contains(value, forbidden) {
				return errRecycleInfoV1Boundary
			}
		}
	}
	return nil
}

// --- 后台预约记录管理 ---

func buildAdminAppointmentRecord(a CustomerAppointment) AdminAppointmentRecord {
	return AdminAppointmentRecord{
		ID: a.ID, AppointmentNo: a.AppointmentNo,
		CustomerName: a.CustomerName, CustomerPhone: a.CustomerPhone,
		StoreID: a.StoreID, StoreName: a.StoreName,
		ServiceType: a.ServiceType, ServiceTypeText: serviceTypeText(a.ServiceType),
		AppointmentDate: a.AppointmentDate, AppointmentTime: a.AppointmentTime,
		Status: a.Status, StatusText: appointmentStatusText(a.Status),
		Remark: a.Remark, StaffNote: a.StaffNote,
		CreatedAt: a.CreatedAt, ConfirmedAt: a.ConfirmedAt, ConfirmedBy: a.ConfirmedBy,
		CompletedAt: a.CompletedAt, CancelledAt: a.CancelledAt, CancelReason: a.CancelReason,
	}
}

// AdminAppointmentFilter 后台预约筛选条件
type AdminAppointmentFilter struct {
	Status      string
	StoreID     string
	Phone       string
	ServiceType string
	DateFrom    string
	DateTo      string
	Page        int
	PageSize    int
}

func (s *MockStore) adminListAppointments(user UserAccount, filter AdminAppointmentFilter) ([]AdminAppointmentRecord, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []AdminAppointmentRecord
	phone := strings.TrimSpace(filter.Phone)
	for _, a := range s.customerAppointments {
		if !s.canAccessStore(user, a.StoreID) {
			continue
		}
		if filter.Status != "" && a.Status != filter.Status {
			continue
		}
		if filter.StoreID != "" && a.StoreID != filter.StoreID {
			continue
		}
		if phone != "" && !strings.Contains(a.CustomerPhone, phone) {
			continue
		}
		if filter.ServiceType != "" && a.ServiceType != filter.ServiceType {
			continue
		}
		if filter.DateFrom != "" && a.AppointmentDate < filter.DateFrom {
			continue
		}
		if filter.DateTo != "" && a.AppointmentDate > filter.DateTo {
			continue
		}
		items = append(items, buildAdminAppointmentRecord(a))
	}
	// 预约时间倒序
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].AppointmentDate != items[j].AppointmentDate {
			return items[i].AppointmentDate > items[j].AppointmentDate
		}
		return items[i].AppointmentTime > items[j].AppointmentTime
	})
	total := len(items)
	start := (filter.Page - 1) * filter.PageSize
	if start >= total {
		return []AdminAppointmentRecord{}, total
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

func (s *MockStore) adminGetAppointment(user UserAccount, appointmentID string) (AdminAppointmentRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.customerAppointments {
		if a.ID == appointmentID {
			if !s.canAccessStore(user, a.StoreID) {
				return AdminAppointmentRecord{}, errUnauthorizedStore
			}
			return buildAdminAppointmentRecord(a), nil
		}
	}
	return AdminAppointmentRecord{}, errAppointmentNotFound
}

// adminCancelAppointment 后台取消预约：需填原因，不受提前量限制
func (s *MockStore) adminCancelAppointment(user UserAccount, appointmentID, reason string) error {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.customerAppointments {
		if a.ID == appointmentID {
			if !s.canAccessStore(user, a.StoreID) {
				return errUnauthorizedStore
			}
			if a.Status != AppointmentStatusPending && a.Status != AppointmentStatusConfirmed {
				return errAppointmentStatusFlow
			}
			previous := s.customerAppointments[i]
			s.customerAppointments[i].Status = AppointmentStatusCancelled
			s.customerAppointments[i].CancelledAt = &now
			s.customerAppointments[i].CancelReason = reason
			s.customerAppointments[i].UpdatedAt = now
			if err := s.persistCustomerAppointmentLocked(s.customerAppointments[i]); err != nil {
				s.customerAppointments[i] = previous
				return err
			}
			s.appendAuditLogLocked("预约管理", "后台取消预约", user.DisplayName, "warning", "medium",
				fmt.Sprintf("预约 %s 已由后台取消：%s", a.AppointmentNo, reason))
			return nil
		}
	}
	return errAppointmentNotFound
}

// adminUpdateAppointmentStaffNote 更新内部备注（顾客端不可见）
func (s *MockStore) adminUpdateAppointmentStaffNote(user UserAccount, appointmentID, staffNote string) error {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.customerAppointments {
		if a.ID == appointmentID {
			if !s.canAccessStore(user, a.StoreID) {
				return errUnauthorizedStore
			}
			if a.Status == AppointmentStatusCancelled || a.Status == AppointmentStatusCompleted {
				return errAppointmentStatusFlow
			}
			previous := s.customerAppointments[i]
			s.customerAppointments[i].StaffNote = staffNote
			s.customerAppointments[i].UpdatedAt = now
			if err := s.persistCustomerAppointmentLocked(s.customerAppointments[i]); err != nil {
				s.customerAppointments[i] = previous
				return err
			}
			return nil
		}
	}
	return errAppointmentNotFound
}

// --- 上传素材登记 ---

func (s *MockStore) persistUploadAssetsLocked() error {
	if s.persistence == nil {
		return nil
	}
	return s.persistence.saveConfig(context.Background(), configKeyUploadAssets, s.uploadAssets)
}

func (s *MockStore) registerUploadAsset(asset UploadAsset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploadAssets = append([]UploadAsset{asset}, s.uploadAssets...)
	return s.persistUploadAssetsLocked()
}

func (s *MockStore) listUploadAssets(scene string, page, pageSize int) ([]UploadAsset, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []UploadAsset
	for _, a := range s.uploadAssets {
		if a.Deleted {
			continue
		}
		if scene != "" && a.Scene != scene {
			continue
		}
		items = append(items, a)
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return []UploadAsset{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

func (s *MockStore) deleteUploadAsset(user UserAccount, assetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.uploadAssets {
		if a.ID == assetID {
			// 引用检查：Banner / 款式 / 门店 是否仍使用该 URL
			for _, b := range s.customerHomeConfig.Banners {
				if b.ImageURL == a.URL {
					return errUploadAssetInUse
				}
			}
			for _, p := range s.catalogProducts {
				if p.ImageURL == a.URL {
					return errUploadAssetInUse
				}
				for _, img := range p.Images {
					if img == a.URL {
						return errUploadAssetInUse
					}
				}
				for _, img := range p.DetailImages {
					if img == a.URL {
						return errUploadAssetInUse
					}
				}
			}
			if s.customerRecycleInfo.ImageURL == a.URL {
				return errUploadAssetInUse
			}
			for _, st := range s.stores {
				if st.ImageURL == a.URL {
					return errUploadAssetInUse
				}
			}
			s.uploadAssets[i].Deleted = true
			if err := s.persistUploadAssetsLocked(); err != nil {
				return err
			}
			s.appendAuditLogLocked("顾客端配置", "删除素材", user.DisplayName, "success", "low", fmt.Sprintf("素材 %s 已删除", a.ID))
			return nil
		}
	}
	return errUploadAssetNotFound
}
