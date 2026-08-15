/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: customer_types.go
 * 功能描述: 顾客端类型定义
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

package app

import "time"

// CustomerProfile 顾客档案
type CustomerProfile struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	OpenID    string    `json:"openId"`
	UnionID   string    `json:"unionId,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Nickname  string    `json:"nickname,omitempty"`
	AvatarURL string    `json:"avatarUrl,omitempty"`
	Status    string    `json:"status"` // active / disabled
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CustomerSession 顾客会话
type CustomerSession struct {
	Token      string    `json:"token"`
	CustomerID string    `json:"customerId"`
	OrgID      string    `json:"orgId"`
	Phone      string    `json:"phone,omitempty"`
	LoginAt    time.Time `json:"loginAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// 预约状态枚举
const (
	AppointmentStatusPending    = "PENDING"
	AppointmentStatusConfirmed  = "CONFIRMED"
	AppointmentStatusArrived    = "ARRIVED"
	AppointmentStatusCompleted  = "COMPLETED"
	AppointmentStatusCancelled  = "CANCELLED"
	AppointmentStatusNoShow     = "NO_SHOW"
	AppointmentStatusTerminated = "TERMINATED"
)

// 预约服务类型枚举
const (
	ServiceTypeOldForNew = "OLD_FOR_NEW" // 旧金换新
	ServiceTypeRepair    = "REPAIR"      // 黄金维修
	ServiceTypeConsult   = "CONSULT"     // 款式工费咨询
	ServiceTypeRecycle   = "RECYCLE"     // 到店回收咨询
)

// CustomerAppointment 到店预约
type CustomerAppointment struct {
	ID              string     `json:"id"`
	AppointmentNo   string     `json:"appointmentNo"` // 预约编号 YY+YYYYMMDD+0001
	OrgID           string     `json:"orgId"`
	CustomerID      string     `json:"customerId"`
	CustomerName    string     `json:"customerName"`
	CustomerPhone   string     `json:"customerPhone"`
	StoreID         string     `json:"storeId"`
	StoreName       string     `json:"storeName"`
	StoreAddress    string     `json:"storeAddress,omitempty"`
	StorePhone      string     `json:"storePhone,omitempty"`
	ServiceType     string     `json:"serviceType"` // OLD_FOR_NEW / REPAIR / CONSULT / RECYCLE
	AppointmentDate string     `json:"appointmentDate"` // YYYY-MM-DD
	AppointmentTime string     `json:"appointmentTime"` // HH:MM
	Status          string     `json:"status"`
	Remark          string     `json:"remark,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	CancelledAt     *time.Time `json:"cancelledAt,omitempty"`
	CancelReason    string     `json:"cancelReason,omitempty"`
	ConfirmedAt     *time.Time `json:"confirmedAt,omitempty"`
	ConfirmedBy     string     `json:"confirmedBy,omitempty"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
}

// --- 顾客端 DTO ---

// CustomerProductDTO 顾客可见的商品白名单 DTO
type CustomerProductDTO struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	ImageURL         string   `json:"imageUrl"`
	Category         string   `json:"category"`
	Purity           string   `json:"purity"`
	RetailPrice      float64  `json:"retailPrice"`
	GramWeight       float64  `json:"gramWeight"`
	RecommendedScene string   `json:"recommendedScene,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// CustomerStoreDTO 顾客可见的门店白名单 DTO（列表）
type CustomerStoreDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	City          string  `json:"city"`
	Address       string  `json:"address"`
	ContactPhone  string  `json:"contactPhone,omitempty"`
	BusinessHours string  `json:"businessHours,omitempty"`
	Distance      float64 `json:"distance,omitempty"` // km，前端计算
}

// CustomerStoreDetailDTO 顾客门店详情（含经纬度，用于导航）
type CustomerStoreDetailDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	City          string  `json:"city"`
	Address       string  `json:"address"`
	ContactPhone  string  `json:"contactPhone,omitempty"`
	BusinessHours string  `json:"businessHours,omitempty"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
}

// CustomerAppointmentDTO 顾客预约 DTO（仅本人可见）
type CustomerAppointmentDTO struct {
	ID              string     `json:"id"`
	AppointmentNo   string     `json:"appointmentNo"`
	StoreID         string     `json:"storeId"`
	StoreName       string     `json:"storeName"`
	StoreAddress    string     `json:"storeAddress,omitempty"`
	StorePhone      string     `json:"storePhone,omitempty"`
	ServiceType     string     `json:"serviceType"`
	ServiceTypeText string     `json:"serviceTypeText"`
	AppointmentDate string     `json:"appointmentDate"`
	AppointmentTime string     `json:"appointmentTime"`
	TimeRange       string     `json:"timeRange"` // 例 14:00-14:30
	Status          string     `json:"status"`
	StatusText      string     `json:"statusText"`
	Remark          string     `json:"remark,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	CancelledAt     *time.Time `json:"cancelledAt,omitempty"`
	CancelReason    string     `json:"cancelReason,omitempty"`
	ConfirmedAt     *time.Time `json:"confirmedAt,omitempty"`
}

// CustomerProfileDTO 顾客个人信息（脱敏）
type CustomerProfileDTO struct {
	ID        string `json:"id"`
	Phone     string `json:"phone"` // 脱敏：138****1234
	Nickname  string `json:"nickname,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

// CustomerHomeResponse 顾客首页聚合数据
type CustomerHomeResponse struct {
	Banners      []CustomerBanner `json:"banners"`
	BrandName    string           `json:"brandName"`
	BrandSlogan1 string           `json:"brandSlogan1"`
	BrandSlogan2 string           `json:"brandSlogan2"`
	ServicePhone string           `json:"servicePhone,omitempty"`
}

// CustomerBanner 首页轮播图
type CustomerBanner struct {
	ImageURL string `json:"imageUrl"`
	LinkType string `json:"linkType,omitempty"` // products / recycle / appointment / stores
	LinkURL  string `json:"linkUrl,omitempty"`
}

// CustomerRecycleInfo 黄金回收服务介绍
type CustomerRecycleInfo struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Process string `json:"process,omitempty"`
	Notes   string `json:"notes,omitempty"`
}

// --- 请求体 ---

// CustomerWechatLoginRequest 顾客微信登录请求
type CustomerWechatLoginRequest struct {
	Code      string `json:"code"`
	PhoneCode string `json:"phoneCode,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

// CustomerPhoneAuthRequest 手机号授权请求
type CustomerPhoneAuthRequest struct {
	PhoneCode string `json:"phoneCode"`
}

// CustomerStoreSlotDTO 门店某日时段可用性 DTO
type CustomerStoreSlotDTO struct {
	Time      string `json:"time"`      // HH:MM
	Available bool   `json:"available"` // true=可预约 false=已约满
	State     string `json:"state"`     // open=可预约 / full=已约满 / closed=营业外
}

// CreateAppointmentRequest 创建预约请求
type CreateAppointmentRequest struct {
	StoreID         string `json:"storeId"`
	ServiceType     string `json:"serviceType"` // OLD_FOR_NEW / REPAIR / CONSULT / RECYCLE
	AppointmentDate string `json:"appointmentDate"`
	AppointmentTime string `json:"appointmentTime"`
	ContactName     string `json:"contactName"`
	ContactPhone    string `json:"contactPhone"`
	Remark          string `json:"remark,omitempty"`
}

// CancelAppointmentRequest 取消预约请求
type CancelAppointmentRequest struct {
	Reason string `json:"reason,omitempty"`
}

// UpdateAppointmentNotesRequest 修改预约备注请求（仅修改备注，不修改时间）
type UpdateAppointmentNotesRequest struct {
	Remark string `json:"remark"`
}

// --- 员工端预约管理 DTO ---

// StaffAppointmentDTO 员工查看的预约 DTO
type StaffAppointmentDTO struct {
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
	CreatedAt       time.Time  `json:"createdAt"`
	ConfirmedAt     *time.Time `json:"confirmedAt,omitempty"`
	ConfirmedBy     string     `json:"confirmedBy,omitempty"`
}

// appointmentStatusText 状态中文映射
func appointmentStatusText(status string) string {
	switch status {
	case AppointmentStatusPending:
		return "待确认"
	case AppointmentStatusConfirmed:
		return "已确认"
	case AppointmentStatusArrived:
		return "已到店"
	case AppointmentStatusCompleted:
		return "已完成"
	case AppointmentStatusCancelled:
		return "已取消"
	case AppointmentStatusNoShow:
		return "未到店"
	case AppointmentStatusTerminated:
		return "已终止"
	default:
		return status
	}
}

// serviceTypeText 服务类型中文映射
func serviceTypeText(t string) string {
	switch t {
	case ServiceTypeOldForNew:
		return "旧金换新"
	case ServiceTypeRepair:
		return "黄金维修"
	case ServiceTypeConsult:
		return "款式工费咨询"
	case ServiceTypeRecycle:
		return "到店回收咨询"
	default:
		return t
	}
}

// validServiceType 是否是合法的服务类型
func validServiceType(t string) bool {
	switch t {
	case ServiceTypeOldForNew, ServiceTypeRepair, ServiceTypeConsult, ServiceTypeRecycle:
		return true
	}
	return false
}

// canCancelAppointment 判断预约是否可取消（距预约时间 >= 2 小时）
func canCancelAppointment(appt CustomerAppointment, now time.Time) bool {
	if appt.Status != AppointmentStatusPending && appt.Status != AppointmentStatusConfirmed {
		return false
	}
	apptTime, err := time.ParseInLocation("2006-01-02 15:04", appt.AppointmentDate+" "+appt.AppointmentTime, time.Local)
	if err != nil {
		return false
	}
	return now.Add(2 * time.Hour).Before(apptTime) || now.Add(2*time.Hour).Equal(apptTime)
}

// maskPhone 手机号脱敏
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
