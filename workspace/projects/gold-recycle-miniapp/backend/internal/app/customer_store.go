/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: customer_store.go
 * 功能描述: 顾客端数据访问层
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errCustomerNotFound         = errors.New("customer not found")
	errAppointmentNotFound      = errors.New("appointment not found")
	errAppointmentCancelled     = errors.New("appointment already cancelled")
	errAppointmentTimeTooLate   = errors.New("cannot cancel within 2 hours of appointment")
	errDuplicateAppointment     = errors.New("duplicate appointment for same store and time slot")
	errAppointmentCapacityFull  = errors.New("time slot is full")
	errInvalidAppointmentTime   = errors.New("invalid appointment time")
	errAppointmentStatusFlow    = errors.New("appointment status transition not allowed")
	errStoreAppointmentDisabled = errors.New("store appointment disabled")
)

// --- 顾客会话 ---

func (s *MockStore) getCustomerByToken(token string) (CustomerSession, CustomerProfile, bool) {
	s.mu.RLock()
	session, ok := s.customerSessions[token]
	profile := s.findCustomerProfileLocked(session.CustomerID)
	s.mu.RUnlock()

	if ok && !session.ExpiresAt.Before(time.Now()) && profile.ID != "" {
		return session, profile, true
	}

	if s.persistence != nil {
		persisted, found, err := s.persistence.loadCustomerSession(context.Background(), token)
		if err != nil || !found || persisted.ExpiresAt.Before(time.Now()) {
			return CustomerSession{}, CustomerProfile{}, false
		}
		s.mu.Lock()
		s.customerSessions[token] = persisted
		profile = s.findCustomerProfileLocked(persisted.CustomerID)
		s.mu.Unlock()
		if profile.ID != "" {
			return persisted, profile, true
		}
	}
	return CustomerSession{}, CustomerProfile{}, false
}

func (s *MockStore) createCustomerSession(secret string, profile CustomerProfile) (CustomerSession, error) {
	payload := fmt.Sprintf("customer:%s:%d", profile.ID, time.Now().UnixNano())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	token := hex.EncodeToString(mac.Sum(nil))

	now := time.Now()
	session := CustomerSession{
		Token:      token,
		CustomerID: profile.ID,
		OrgID:      profile.OrgID,
		Phone:      profile.Phone,
		LoginAt:    now,
		ExpiresAt:  now.Add(12 * time.Hour),
	}

	s.mu.Lock()
	s.customerSessions[token] = session
	s.mu.Unlock()

	if s.persistence != nil {
		if err := s.persistence.saveCustomerSession(context.Background(), session); err != nil {
			return CustomerSession{}, err
		}
	}
	return session, nil
}

func (s *MockStore) deleteCustomerSession(token string) {
	s.mu.Lock()
	delete(s.customerSessions, token)
	s.mu.Unlock()
	if s.persistence != nil {
		_ = s.persistence.deleteCustomerSession(context.Background(), token)
	}
}

// --- 顾客档案 ---

func (s *MockStore) findCustomerProfileLocked(id string) CustomerProfile {
	for _, p := range s.customerProfiles {
		if p.ID == id {
			return p
		}
	}
	return CustomerProfile{}
}

func (s *MockStore) findCustomerByOpenID(openID string) (CustomerProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.customerProfiles {
		if p.OpenID == openID {
			return p, true
		}
	}
	return CustomerProfile{}, false
}

func (s *MockStore) findOrCreateCustomerByOpenID(orgID, openID, unionID, nickname, avatarURL string) (CustomerProfile, bool) {
	if profile, ok := s.findCustomerByOpenID(openID); ok {
		changed := false
		if unionID != "" && profile.UnionID == "" {
			profile.UnionID = unionID
			changed = true
		}
		if nickname != "" && profile.Nickname == "" {
			profile.Nickname = nickname
			changed = true
		}
		if avatarURL != "" && profile.AvatarURL == "" {
			profile.AvatarURL = avatarURL
			changed = true
		}
		if changed {
			profile.UpdatedAt = time.Now()
			s.mu.Lock()
			s.updateCustomerProfileLocked(profile)
			s.mu.Unlock()
		}
		return profile, false
	}

	now := time.Now()
	profile := CustomerProfile{
		ID:        fmt.Sprintf("customer-%s", generateID()),
		OrgID:     orgID,
		OpenID:    openID,
		UnionID:   unionID,
		Nickname:  nickname,
		AvatarURL: avatarURL,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.mu.Lock()
	s.customerProfiles = append(s.customerProfiles, profile)
	if s.persistence != nil {
		_ = s.persistCustomerProfilesLocked()
	}
	s.mu.Unlock()
	return profile, true
}

func (s *MockStore) updateCustomerPhone(customerID, phone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.customerProfiles {
		if p.ID == customerID {
			s.customerProfiles[i].Phone = phone
			s.customerProfiles[i].UpdatedAt = time.Now()
			if s.persistence != nil {
				return s.persistCustomerProfilesLocked()
			}
			return nil
		}
	}
	return errCustomerNotFound
}

func (s *MockStore) updateCustomerProfileLocked(profile CustomerProfile) {
	for i, p := range s.customerProfiles {
		if p.ID == profile.ID {
			s.customerProfiles[i] = profile
			if s.persistence != nil {
				_ = s.persistCustomerProfilesLocked()
			}
			return
		}
	}
}

func (s *MockStore) persistCustomerProfilesLocked() error {
	return s.persistence.saveConfig(context.Background(), configKeyCustomerProfiles, s.customerProfiles)
}

// --- 顾客公开数据 ---

func (s *MockStore) customerProductList(category, keyword string, page, pageSize int) ([]CustomerProductDTO, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []CustomerProductDTO
	for _, p := range s.catalogProducts {
		if p.Status != "active" {
			continue
		}
		if category != "" && p.Category != category {
			continue
		}
		if keyword != "" {
			k := strings.ToLower(keyword)
			hit := strings.Contains(strings.ToLower(p.Name), k) ||
				strings.Contains(strings.ToLower(p.Category), k) ||
				strings.Contains(strings.ToLower(p.Purity), k) ||
				strings.Contains(strings.ToLower(p.RecommendedScene), k)
			if !hit {
				continue
			}
		}
		items = append(items, buildCustomerProductDTO(p))
	}

	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return []CustomerProductDTO{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

func (s *MockStore) customerProductDetail(id string) (CustomerProductDTO, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.catalogProducts {
		if p.ID == id && p.Status == "active" {
			return buildCustomerProductDTO(p), true
		}
	}
	return CustomerProductDTO{}, false
}

// buildCustomerProductDTO 构造顾客可见的商品 DTO（含后台配置的扩展字段）
func buildCustomerProductDTO(p CatalogProduct) CustomerProductDTO {
	images := p.Images
	if len(images) == 0 && p.ImageURL != "" {
		images = []string{p.ImageURL}
	}
	return CustomerProductDTO{
		ID:                     p.ID,
		Name:                   p.Name,
		ImageURL:               p.ImageURL,
		Category:               p.Category,
		Purity:                 p.Purity,
		RetailPrice:            p.RetailPrice,
		GramWeight:             p.GramWeight,
		RecommendedScene:       p.RecommendedScene,
		Tags:                   p.Tags,
		LaborFeeRef:            p.LaborFeeRef,
		Description:            p.Description,
		LaborFeeNote:           p.LaborFeeNote,
		Images:                 images,
		DetailImages:           p.DetailImages,
		ApplicableServiceTypes: p.ApplicableServiceTypes,
		IsRecommended:          p.IsRecommended,
		IsHot:                  p.IsHot,
	}
}

func (s *MockStore) customerProductCategories() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := make(map[string]bool)
	var categories []string
	for _, p := range s.catalogProducts {
		if p.Status != "active" {
			continue
		}
		if p.Category != "" && !seen[p.Category] {
			seen[p.Category] = true
			categories = append(categories, p.Category)
		}
	}
	return categories
}

func (s *MockStore) customerStoreList() []CustomerStoreDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []CustomerStoreDTO
	for _, st := range s.stores {
		if st.Status != "active" {
			continue
		}
		items = append(items, buildCustomerStoreDTO(st))
	}
	return items
}

// buildCustomerStoreDTO 构造顾客门店 DTO（含后台配置的扩展字段）
func buildCustomerStoreDTO(st StoreInfo) CustomerStoreDTO {
	enabled := st.AppointmentEnabled == nil || *st.AppointmentEnabled
	return CustomerStoreDTO{
		ID:                 st.ID,
		Name:               st.Name,
		City:               st.City,
		Address:            st.Address,
		ContactPhone:       st.ContactPhone,
		BusinessHours:      st.BusinessHours,
		ImageURL:           st.ImageURL,
		ServiceTags:        st.ServiceTags,
		AppointmentEnabled: enabled,
	}
}

func (s *MockStore) customerStoreDetail(id string) (CustomerStoreDetailDTO, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, st := range s.stores {
		if st.ID == id && st.Status == "active" {
			enabled := st.AppointmentEnabled == nil || *st.AppointmentEnabled
			return CustomerStoreDetailDTO{
				ID:                 st.ID,
				Name:               st.Name,
				City:               st.City,
				Address:            st.Address,
				ContactPhone:       st.ContactPhone,
				BusinessHours:      st.BusinessHours,
				ImageURL:           st.ImageURL,
				ServiceTags:        st.ServiceTags,
				AppointmentEnabled: enabled,
				Longitude:          st.Longitude,
				Latitude:           st.Latitude,
			}, true
		}
	}
	return CustomerStoreDetailDTO{}, false
}

func (s *MockStore) customerHomeData() CustomerHomeResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.customerHomeConfig
	if len(cfg.Banners) == 0 && cfg.BrandName == "" {
		cfg = defaultCustomerHomeConfig()
	}
	normalizeCustomerHomeConfigLocked(&cfg)
	// 顾客端只返回启用中的 Banner，按 SortOrder 排序
	visible := make([]CustomerBanner, 0, len(cfg.Banners))
	for _, b := range cfg.Banners {
		if bannerEnabled(b.Enabled) && strings.TrimSpace(b.ImageURL) != "" {
			visible = append(visible, b)
		}
	}
	cfg.Banners = visible
	// 注入预约规则摘要（替代前端硬编码）
	rules := s.appointmentRules
	if rules.SlotMinutes == 0 {
		rules = defaultAppointmentRules()
	}
	cfg.AppointmentRules = AppointmentRulesSummary{
		BookableDays:         rules.BookableDays,
		SlotMinutes:          rules.SlotMinutes,
		CancelLeadMinutes:    rules.CancelLeadMinutes,
		SameDayLeadMinutes:   rules.SameDayLeadMinutes,
		BookableServiceTypes: rules.BookableServiceTypes,
	}
	return cfg
}

func (s *MockStore) customerRecycleInfoData() CustomerRecycleInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info := s.customerRecycleInfo
	if info.Title == "" || (len(info.Process) == 0 && len(info.Services) == 0) {
		return defaultCustomerRecycleInfo()
	}
	return info
}

// --- 顾客预约 ---

func (s *MockStore) createCustomerAppointment(customer CustomerProfile, req CreateAppointmentRequest) (CustomerAppointment, error) {
	now := time.Now()

	// 服务类型校验
	if !validServiceType(req.ServiceType) {
		return CustomerAppointment{}, errors.New("invalid service type")
	}

	// 解析预约时间
	apptTime, err := time.ParseInLocation("2006-01-02 15:04", req.AppointmentDate+" "+req.AppointmentTime, time.Local)
	if err != nil {
		return CustomerAppointment{}, errInvalidAppointmentTime
	}

	// 预约规则（可配置，替换原硬编码）
	rules := s.appointmentRulesSnapshot()

	// 当天预约至少提前 sameDayLeadMinutes（默认 60 分钟）
	if apptTime.Before(now.Add(time.Duration(rules.SameDayLeadMinutes) * time.Minute)) {
		return CustomerAppointment{}, errInvalidAppointmentTime
	}

	// 可预约范围：未来 bookableDays（默认 7）天
	if apptTime.After(now.Add(time.Duration(rules.BookableDays) * 24 * time.Hour)) {
		return CustomerAppointment{}, errInvalidAppointmentTime
	}

	// 可预约服务类型受限时校验
	if len(rules.BookableServiceTypes) > 0 {
		allowed := false
		for _, t := range rules.BookableServiceTypes {
			if t == req.ServiceType {
				allowed = true
				break
			}
		}
		if !allowed {
			return CustomerAppointment{}, errors.New("service type not bookable")
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找门店
	var store StoreInfo
	storeFound := false
	for _, st := range s.stores {
		if st.ID == req.StoreID && st.Status == "active" {
			store = st
			storeFound = true
			break
		}
	}
	if !storeFound {
		return CustomerAppointment{}, errAppointmentNotFound
	}

	// 门店预约开关（关闭后不可新预约）
	if store.AppointmentEnabled != nil && !*store.AppointmentEnabled {
		return CustomerAppointment{}, errStoreAppointmentDisabled
	}

	// 重复预约检查：同一手机号 + 同一门店 + 同一日期 + 同一时间
	for _, a := range s.customerAppointments {
		if a.CustomerPhone == req.ContactPhone && a.StoreID == req.StoreID &&
			a.AppointmentDate == req.AppointmentDate && a.AppointmentTime == req.AppointmentTime &&
			(a.Status == AppointmentStatusPending || a.Status == AppointmentStatusConfirmed) {
			return CustomerAppointment{}, errDuplicateAppointment
		}
	}

	// 时段容量检查：同一门店同一时段最多 slotCapacity 单（默认 2）
	slotCount := 0
	for _, a := range s.customerAppointments {
		if a.StoreID == req.StoreID &&
			a.AppointmentDate == req.AppointmentDate &&
			a.AppointmentTime == req.AppointmentTime &&
			(a.Status == AppointmentStatusPending || a.Status == AppointmentStatusConfirmed) {
			slotCount++
		}
	}
	if slotCount >= rules.SlotCapacity {
		return CustomerAppointment{}, errAppointmentCapacityFull
	}

	s.customerSeq++
	apptID := fmt.Sprintf("appt-%d-%s", s.customerSeq, generateID())
	apptNo := generateAppointmentNo(req.AppointmentDate, s.customerSeq)

	appt := CustomerAppointment{
		ID:              apptID,
		AppointmentNo:   apptNo,
		OrgID:           customer.OrgID,
		CustomerID:      customer.ID,
		CustomerName:    req.ContactName,
		CustomerPhone:   req.ContactPhone,
		StoreID:         store.ID,
		StoreName:       store.Name,
		StoreAddress:    store.City + store.Address,
		StorePhone:      store.ContactPhone,
		ServiceType:     req.ServiceType,
		AppointmentDate: req.AppointmentDate,
		AppointmentTime: req.AppointmentTime,
		Status:          AppointmentStatusPending,
		Remark:          req.Remark,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	s.customerAppointments = append(s.customerAppointments, appt)
	if s.persistence != nil {
		_ = s.persistence.saveCustomerAppointment(context.Background(), appt)
	}
	return appt, nil
}

// listCustomerAppointmentSlots 查询门店某日所有时段可用性
func (s *MockStore) listCustomerAppointmentSlots(storeID, date string) ([]CustomerStoreSlotDTO, error) {
	if _, err := time.ParseInLocation("2006-01-02", date, time.Local); err != nil {
		return nil, errors.New("invalid date format")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 查找门店
	var store *StoreInfo
	for i := range s.stores {
		if s.stores[i].ID == storeID && s.stores[i].Status == "active" {
			store = &s.stores[i]
			break
		}
	}
	if store == nil {
		return nil, errAppointmentNotFound
	}

	// 营业时间与粒度来自预约规则（可配置，默认 09:30-21:30 / 30 分钟）
	rules := s.appointmentRulesLocked()
	openTime, err := time.Parse("15:04", rules.OpenTime)
	if err != nil {
		openTime = time.Date(0, 1, 1, 9, 30, 0, 0, time.UTC)
	}
	closeTime, err := time.Parse("15:04", rules.CloseTime)
	if err != nil {
		closeTime = time.Date(0, 1, 1, 21, 30, 0, 0, time.UTC)
	}
	slotMinutes := rules.SlotMinutes
	if slotMinutes <= 0 {
		slotMinutes = 30
	}
	slotCapacity := rules.SlotCapacity
	if slotCapacity <= 0 {
		slotCapacity = 2
	}
	// 门店预约开关关闭时不返回可约时段
	if store.AppointmentEnabled != nil && !*store.AppointmentEnabled {
		return []CustomerStoreSlotDTO{}, nil
	}

	// 统计每个时段已预约数
	slotBooked := map[string]int{}
	for _, a := range s.customerAppointments {
		if a.StoreID != storeID || a.AppointmentDate != date {
			continue
		}
		if a.Status != AppointmentStatusPending && a.Status != AppointmentStatusConfirmed {
			continue
		}
		slotBooked[a.AppointmentTime]++
	}

	// 生成所有时段（按规则粒度递增）
	cur := openTime
	end := closeTime
	var slots []CustomerStoreSlotDTO

	for cur.Before(end) {
		timeStr := cur.Format("15:04")
		booked := slotBooked[timeStr]
		state := "open"
		available := booked < slotCapacity
		if !available {
			state = "full"
		}
		slots = append(slots, CustomerStoreSlotDTO{
			Time:      timeStr,
			Available: available,
			State:     state,
		})
		cur = cur.Add(time.Duration(slotMinutes) * time.Minute)
	}

	return slots, nil
}

// updateCustomerAppointmentNotes 顾客修改预约备注（仅修改备注）
func (s *MockStore) updateCustomerAppointmentNotes(customerID, appointmentID, remark string) error {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.customerAppointments {
		if a.ID == appointmentID && a.CustomerID == customerID {
			if a.Status != AppointmentStatusPending && a.Status != AppointmentStatusConfirmed {
				return errAppointmentStatusFlow
			}
			s.customerAppointments[i].Remark = remark
			s.customerAppointments[i].UpdatedAt = now
			if s.persistence != nil {
				_ = s.persistence.saveCustomerAppointment(context.Background(), s.customerAppointments[i])
			}
			return nil
		}
	}
	return errAppointmentNotFound
}

// buildCustomerAppointmentDTO 构造顾客预约 DTO
func buildCustomerAppointmentDTO(a CustomerAppointment) CustomerAppointmentDTO {
	endMin := addMinutes(a.AppointmentTime, 30)
	return CustomerAppointmentDTO{
		ID:              a.ID,
		AppointmentNo:   a.AppointmentNo,
		StoreID:         a.StoreID,
		StoreName:       a.StoreName,
		StoreAddress:    a.StoreAddress,
		StorePhone:      a.StorePhone,
		ServiceType:     a.ServiceType,
		ServiceTypeText: serviceTypeText(a.ServiceType),
		AppointmentDate: a.AppointmentDate,
		AppointmentTime: a.AppointmentTime,
		TimeRange:       a.AppointmentTime + "-" + endMin,
		Status:          a.Status,
		StatusText:      appointmentStatusText(a.Status),
		Remark:          a.Remark,
		CreatedAt:       a.CreatedAt,
		CancelledAt:     a.CancelledAt,
		CancelReason:    a.CancelReason,
		ConfirmedAt:     a.ConfirmedAt,
	}
}

// buildStaffAppointmentDTO 构造员工端预约 DTO
func buildStaffAppointmentDTO(a CustomerAppointment) StaffAppointmentDTO {
	return StaffAppointmentDTO{
		ID:              a.ID,
		AppointmentNo:   a.AppointmentNo,
		CustomerName:    a.CustomerName,
		CustomerPhone:   a.CustomerPhone,
		StoreID:         a.StoreID,
		StoreName:       a.StoreName,
		ServiceType:     a.ServiceType,
		ServiceTypeText: serviceTypeText(a.ServiceType),
		AppointmentDate: a.AppointmentDate,
		AppointmentTime: a.AppointmentTime,
		Status:          a.Status,
		StatusText:      appointmentStatusText(a.Status),
		Remark:          a.Remark,
		CreatedAt:       a.CreatedAt,
		ConfirmedAt:     a.ConfirmedAt,
		ConfirmedBy:     a.ConfirmedBy,
	}
}

// addMinutes 给 HH:MM 加分钟返回 HH:MM
func addMinutes(t string, minutes int) string {
	parsed, err := time.Parse("15:04", t)
	if err != nil {
		return t
	}
	return parsed.Add(time.Duration(minutes) * time.Minute).Format("15:04")
}

// generateAppointmentNo 生成预约编号 YY+YYYYMMDD+4 位序号
func generateAppointmentNo(date string, seq int) string {
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return fmt.Sprintf("YY%s%04d", date, seq)
	}
	year2 := t.Format("06") // 后两位年份
	day := t.Format("20060102")
	return fmt.Sprintf("YY%s%s%04d", year2, day, seq)
}

func (s *MockStore) listCustomerAppointments(customerID string) []CustomerAppointment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []CustomerAppointment
	for _, a := range s.customerAppointments {
		if a.CustomerID == customerID {
			items = append(items, a)
		}
	}
	// 按时间倒序
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].CreatedAt.Before(items[j].CreatedAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items
}

func (s *MockStore) getCustomerAppointment(customerID, appointmentID string) (CustomerAppointment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.customerAppointments {
		if a.ID == appointmentID && a.CustomerID == customerID {
			return a, true
		}
	}
	return CustomerAppointment{}, false
}

func (s *MockStore) cancelCustomerAppointment(customerID, appointmentID, reason string) error {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.customerAppointments {
		if a.ID == appointmentID && a.CustomerID == customerID {
			if a.Status == AppointmentStatusCancelled || a.Status == AppointmentStatusCompleted {
				return errAppointmentCancelled
			}
			if a.Status != AppointmentStatusPending && a.Status != AppointmentStatusConfirmed {
				return errAppointmentStatusFlow
			}
			rules := s.appointmentRulesLocked()
			cancelLead := rules.CancelLeadMinutes
			if cancelLead <= 0 {
				cancelLead = 120
			}
			if !canCancelAppointment(a, now, cancelLead) {
				return errAppointmentTimeTooLate
			}
			s.customerAppointments[i].Status = AppointmentStatusCancelled
			s.customerAppointments[i].CancelledAt = &now
			s.customerAppointments[i].CancelReason = reason
			s.customerAppointments[i].UpdatedAt = now
			if s.persistence != nil {
				_ = s.persistence.saveCustomerAppointment(context.Background(), s.customerAppointments[i])
			}
			return nil
		}
	}
	return errAppointmentNotFound
}

// --- 员工预约管理 ---

func (s *MockStore) listStaffAppointments(user UserAccount) []CustomerAppointment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []CustomerAppointment
	for _, a := range s.customerAppointments {
		if s.canAccessStore(user, a.StoreID) {
			items = append(items, a)
		}
	}
	return items
}

func (s *MockStore) staffGetAppointment(user UserAccount, appointmentID string) (CustomerAppointment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.customerAppointments {
		if a.ID == appointmentID && s.canAccessStore(user, a.StoreID) {
			return a, true
		}
	}
	return CustomerAppointment{}, false
}

func (s *MockStore) staffUpdateAppointmentStatus(user UserAccount, appointmentID, newStatus string) error {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, a := range s.customerAppointments {
		if a.ID == appointmentID && s.canAccessStore(user, a.StoreID) {
			// 状态机校验
			switch newStatus {
			case AppointmentStatusConfirmed:
				if a.Status != AppointmentStatusPending {
					return errAppointmentStatusFlow
				}
				s.customerAppointments[i].Status = newStatus
				s.customerAppointments[i].ConfirmedAt = &now
				s.customerAppointments[i].ConfirmedBy = user.DisplayName
			case AppointmentStatusArrived:
				if a.Status != AppointmentStatusConfirmed && a.Status != AppointmentStatusPending {
					return errAppointmentStatusFlow
				}
				s.customerAppointments[i].Status = newStatus
			case AppointmentStatusCompleted:
				if a.Status != AppointmentStatusArrived {
					return errAppointmentStatusFlow
				}
				s.customerAppointments[i].Status = newStatus
				s.customerAppointments[i].CompletedAt = &now
			case AppointmentStatusNoShow:
				if a.Status != AppointmentStatusPending && a.Status != AppointmentStatusConfirmed {
					return errAppointmentStatusFlow
				}
				s.customerAppointments[i].Status = newStatus
			case AppointmentStatusTerminated:
				if a.Status == AppointmentStatusCompleted || a.Status == AppointmentStatusCancelled {
					return errAppointmentStatusFlow
				}
				s.customerAppointments[i].Status = newStatus
			default:
				return errAppointmentStatusFlow
			}
			s.customerAppointments[i].UpdatedAt = now
			if s.persistence != nil {
				_ = s.persistence.saveCustomerAppointment(context.Background(), s.customerAppointments[i])
			}
			return nil
		}
	}
	return errAppointmentNotFound
}

// --- 辅助函数 ---

func defaultCustomerHomeConfig() CustomerHomeResponse {
	return CustomerHomeResponse{
		BrandName:      "金匠馆",
		BrandSlogan1:   "旧金换打新款",
		BrandSlogan2:   "包损耗",
		ServiceCopy:    "黄金维修 · 到店回收 · 款式定制",
		EntryStyleText: "款式图",
		EntryFeeText:   "工费",
		Banners:        []CustomerBanner{},
	}
}

func defaultCustomerRecycleInfo() CustomerRecycleInfo {
	return CustomerRecycleInfo{
		Title: "黄金回收服务",
		Intro: "金匠馆专业黄金回收服务，旧金换打新款，包损耗。到店即可享受专业检测和公正估价。",
		Process: []RecycleProcessStep{
			{Step: 1, Title: "到店咨询", Desc: "携带黄金饰品到门店，专业顾问接待"},
			{Step: 2, Title: "黄金检测", Desc: "使用专业仪器检测纯度和克重"},
			{Step: 3, Title: "确认价格", Desc: "根据实时金价和检测结果给出报价"},
			{Step: 4, Title: "完成回收", Desc: "确认无误后现场结算，款项即时到账"},
		},
		Services: []RecycleServiceItem{
			{Icon: "recycle", Title: "旧金换新", Desc: "旧金饰折价换购新款，补差价即可"},
			{Icon: "repair", Title: "黄金维修", Desc: "变形、断裂、损耗等维修修复服务"},
			{Icon: "consult", Title: "款式咨询", Desc: "专业顾问提供款式与工费咨询"},
			{Icon: "recycle", Title: "到店回收", Desc: "黄金饰品现场检测、公正估价、即时结算"},
		},
		Notices: []string{
			"最终回收价格以门店线下检测为准",
			"请携带有效身份证件办理回收业务",
			"回收金价参考当日上海黄金交易所基准价",
		},
	}
}

func nextSequenceFromCustomerAppointments(appointments []CustomerAppointment) int {
	max := 0
	for _, a := range appointments {
		var seq int
		fmt.Sscanf(a.ID, "appt-%d-", &seq)
		if seq > max {
			max = seq
		}
	}
	return max
}

func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
