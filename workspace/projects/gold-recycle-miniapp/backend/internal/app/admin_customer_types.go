/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: admin_customer_types.go
 * 功能描述: 管理后台「顾客端内容管理」类型定义（Banner、首页文案、预约规则、素材、后台预约 DTO）
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

package app

import "time"

// NearbyStoreRule 首页附近门店展示规则
type NearbyStoreRule struct {
	Mode  string `json:"mode"`  // distance / city / all
	Count int    `json:"count"` // 1-5
}

// AppointmentRules 预约规则（可配置，替换硬编码）
type AppointmentRules struct {
	BookableServiceTypes []string  `json:"bookableServiceTypes"` // 至少 1 种
	BookableDays         int       `json:"bookableDays"`         // 1-30
	SlotMinutes          int       `json:"slotMinutes"`          // 15 / 30 / 60
	OpenTime             string    `json:"openTime"`             // HH:MM
	CloseTime            string    `json:"closeTime"`            // HH:MM > openTime
	SameDayLeadMinutes   int       `json:"sameDayLeadMinutes"`   // >= 0
	CancelLeadMinutes    int       `json:"cancelLeadMinutes"`    // >= 0（顾客取消；后台取消不受限）
	SlotCapacity         int       `json:"slotCapacity"`         // 1-20
	UpdatedAt            time.Time `json:"updatedAt,omitempty"`
}

// UploadAsset 上传素材登记
type UploadAsset struct {
	ID        string    `json:"id"`        // img-yyyyMMdd-xxxx
	Scene     string    `json:"scene"`     // banner / style_main / style_detail / store / service_intro
	URL       string    `json:"url"`       // 可访问 URL（绝对）
	Path      string    `json:"path"`      // 相对路径（/assets/...）
	Storage   string    `json:"storage"`   // local（V1）/ oss
	MimeType  string    `json:"mimeType"`
	Size      int64     `json:"size"`
	RefID     string    `json:"refId,omitempty"`
	OriginalName string `json:"originalName,omitempty"`
	Deleted   bool      `json:"deleted"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

// AdminCustomerProductRecord 后台款式管理完整投影（含扩展字段）
type AdminCustomerProductRecord struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	SKU                    string   `json:"sku,omitempty"`
	Category               string   `json:"category"`
	CategoryTab            string   `json:"categoryTab,omitempty"`
	ImageURL               string   `json:"imageUrl"`
	Images                 []string `json:"images"`
	DetailImages           []string `json:"detailImages"`
	Purity                 string   `json:"purity,omitempty"`
	RetailPrice            float64  `json:"retailPrice"`
	GramWeight             float64  `json:"gramWeight"`
	LaborFeeRef            string   `json:"laborFeeRef,omitempty"`
	Description            string   `json:"description,omitempty"`
	LaborFeeNote           string   `json:"laborFeeNote,omitempty"`
	ApplicableServiceTypes []string `json:"applicableServiceTypes"`
	RecommendedStoreRule   string   `json:"recommendedStoreRule,omitempty"` // nearest / product_stores / all
	SortOrder              int      `json:"sortOrder"`
	IsRecommended          bool     `json:"isRecommended"`
	IsHot                  bool     `json:"isHot"`
	Status                 string   `json:"status"` // active 上架 / inactive 下架
	Tags                   []string `json:"tags"`
	StoreIDs               []string `json:"storeIds"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

// AdminCustomerStoreConfig 门店顾客端扩展配置
type AdminCustomerStoreConfig struct {
	ImageURL           string   `json:"imageUrl,omitempty"`
	AppointmentEnabled *bool    `json:"appointmentEnabled,omitempty"` // nil = 默认可预约
	ServiceTags        []string `json:"serviceTags"`
	SortOrder          int      `json:"sortOrder"`
	Longitude          *float64 `json:"longitude,omitempty"`
	Latitude           *float64 `json:"latitude,omitempty"`
	ContactPhone       string   `json:"contactPhone,omitempty"`
	BusinessHours      string   `json:"businessHours,omitempty"`
	Address            string   `json:"address,omitempty"`
	Status             string   `json:"status,omitempty"` // active / inactive
}

// AdminAppointmentRecord 后台预约管理 DTO（含内部备注）
type AdminAppointmentRecord struct {
	ID              string     `json:"id"`
	AppointmentNo   string     `json:"appointmentNo"`
	CustomerName    string     `json:"customerName"`
	CustomerPhone   string     `json:"customerPhone"`
	StoreID         string     `json:"storeId"`
	StoreName       string     `json:"storeName"`
	ServiceType     string     `json:"serviceType"`
	ServiceTypeText string     `json:"serviceTypeText"`
	AppointmentDate string     `json:"appointmentDate"`
	AppointmentTime string     `json:"appointmentTime"`
	Status          string     `json:"status"`
	StatusText      string     `json:"statusText"`
	Remark          string     `json:"remark,omitempty"`
	StaffNote       string     `json:"staffNote,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	ConfirmedAt     *time.Time `json:"confirmedAt,omitempty"`
	ConfirmedBy     string     `json:"confirmedBy,omitempty"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
	CancelledAt     *time.Time `json:"cancelledAt,omitempty"`
	CancelReason    string     `json:"cancelReason,omitempty"`
}

// --- 请求体 ---

// AdminBannerRequest Banner 新建/编辑请求
type AdminBannerRequest struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle,omitempty"`
	ImageURL   string `json:"imageUrl"`
	LinkType   string `json:"linkType"`
	LinkTarget string `json:"linkTarget,omitempty"`
	SortOrder  int    `json:"sortOrder"`
	Enabled    *bool  `json:"enabled,omitempty"`
}

// AdminHomeConfigRequest 首页文案配置请求
type AdminHomeConfigRequest struct {
	BrandName              string          `json:"brandName"`
	BrandSlogan1           string          `json:"brandSlogan1"`
	BrandSlogan2           string          `json:"brandSlogan2"`
	ServiceCopy            string          `json:"serviceCopy,omitempty"`
	EntryStyleText         string          `json:"entryStyleText,omitempty"`
	EntryFeeText           string          `json:"entryFeeText,omitempty"`
	NearbyStoreRule        *NearbyStoreRule `json:"nearbyStoreRule,omitempty"`
	ServicePhone           string          `json:"servicePhone,omitempty"`
	AppointmentNotes       string          `json:"appointmentNotes,omitempty"`
	ServiceIntro           string          `json:"serviceIntro,omitempty"`
	LocationPermissionNote string          `json:"locationPermissionNote,omitempty"`
}

// AdminAppointmentRulesRequest 预约规则保存请求
type AdminAppointmentRulesRequest struct {
	BookableServiceTypes []string `json:"bookableServiceTypes"`
	BookableDays         int      `json:"bookableDays"`
	SlotMinutes          int      `json:"slotMinutes"`
	OpenTime             string   `json:"openTime"`
	CloseTime            string   `json:"closeTime"`
	SameDayLeadMinutes   int      `json:"sameDayLeadMinutes"`
	CancelLeadMinutes    int      `json:"cancelLeadMinutes"`
	SlotCapacity         int      `json:"slotCapacity"`
}

// AdminCancelAppointmentRequest 后台取消预约请求（需填原因，不受提前量限制）
type AdminCancelAppointmentRequest struct {
	Reason string `json:"reason"`
}

// AdminStaffNoteRequest 内部备注请求（顾客端不可见）
type AdminStaffNoteRequest struct {
	StaffNote string `json:"staffNote"`
}

// 预约规则默认值（与原硬编码保持一致：09:30-21:30 / 30 分钟 / 容量 2 / 提前 1h / 7 天 / 取消提前 2h）
func defaultAppointmentRules() AppointmentRules {
	return AppointmentRules{
		BookableServiceTypes: []string{ServiceTypeOldForNew, ServiceTypeRepair, ServiceTypeConsult, ServiceTypeRecycle},
		BookableDays:         7,
		SlotMinutes:          30,
		OpenTime:             "09:30",
		CloseTime:            "21:30",
		SameDayLeadMinutes:   60,
		CancelLeadMinutes:    120,
		SlotCapacity:         2,
	}
}

// validateAppointmentRules 校验预约规则取值范围，返回归一化后的规则或错误
func validateAppointmentRules(rules AppointmentRules) (AppointmentRules, error) {
	if len(rules.BookableServiceTypes) == 0 {
		return rules, errInvalidAppointmentRules
	}
	seen := make(map[string]bool, len(rules.BookableServiceTypes))
	cleaned := make([]string, 0, len(rules.BookableServiceTypes))
	for _, t := range rules.BookableServiceTypes {
		if !validServiceType(t) || seen[t] {
			continue
		}
		seen[t] = true
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return rules, errInvalidAppointmentRules
	}
	rules.BookableServiceTypes = cleaned
	if rules.BookableDays < 1 || rules.BookableDays > 30 {
		return rules, errInvalidAppointmentRules
	}
	switch rules.SlotMinutes {
	case 15, 30, 60:
	default:
		return rules, errInvalidAppointmentRules
	}
	openTime, err := time.Parse("15:04", rules.OpenTime)
	if err != nil {
		return rules, errInvalidAppointmentRules
	}
	closeTime, err := time.Parse("15:04", rules.CloseTime)
	if err != nil || !closeTime.After(openTime) {
		return rules, errInvalidAppointmentRules
	}
	if rules.SameDayLeadMinutes < 0 || rules.CancelLeadMinutes < 0 {
		return rules, errInvalidAppointmentRules
	}
	if rules.SlotCapacity < 1 || rules.SlotCapacity > 20 {
		return rules, errInvalidAppointmentRules
	}
	return rules, nil
}

// bannerEnabled Banner 启用状态（nil 视为启用，兼容存量数据）
func bannerEnabled(enabled *bool) bool {
	return enabled == nil || *enabled
}
