/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: customer_handlers.go
 * 功能描述: 顾客端 API Handler 及员工预约管理
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

package app

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// --- DTO 转换 ---

func toCustomerProfileDTO(p CustomerProfile) CustomerProfileDTO {
	return CustomerProfileDTO{
		ID:        p.ID,
		Phone:     maskPhone(p.Phone),
		Nickname:  p.Nickname,
		AvatarURL: p.AvatarURL,
	}
}

// --- 顾客认证 ---

func (a *App) handleCustomerWechatLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	var req CustomerWechatLoginRequest
	if err := decodeJSONLenient(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Code) == "" && !a.Config.MiniAppAllowMockLogin {
		a.writeError(w, r, http.StatusBadRequest, 40002, "missing wechat login code")
		return
	}

	var openID, unionID string
	orgID := "org-gold-v1"

	if a.Config.hasMiniAppWechatAuth() {
		session, err := a.exchangeMiniAppCode(r.Context(), req.Code)
		if err != nil {
			a.writeError(w, r, http.StatusUnauthorized, 40104, "wechat login code exchange failed")
			return
		}
		openID = session.OpenID
		unionID = session.UnionID
	} else if a.Config.MiniAppAllowMockLogin {
		openID = fmt.Sprintf("mock-customer-%s", generateID())
	} else {
		a.writeError(w, r, http.StatusServiceUnavailable, 50301, "miniapp wechat auth is not configured")
		return
	}

	profile, _ := a.store.findOrCreateCustomerByOpenID(orgID, openID, unionID, req.Nickname, req.AvatarURL)

	if strings.TrimSpace(req.PhoneCode) != "" {
		phone, err := a.exchangeMiniAppPhoneCode(r.Context(), req.PhoneCode)
		if err != nil {
			a.writeError(w, r, http.StatusUnauthorized, 40105, "wechat phone code exchange failed")
			return
		}
		_ = a.store.updateCustomerPhone(profile.ID, phone)
		profile.Phone = phone
	}

	session, err := a.store.createCustomerSession(a.Config.TokenSecret, profile)
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to create session")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"token":   session.Token,
		"profile": toCustomerProfileDTO(profile),
	})
}

func (a *App) handleCustomerPhoneAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	var req CustomerPhoneAuthRequest
	if err := decodeJSONLenient(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	if strings.TrimSpace(req.PhoneCode) == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "missing phoneCode")
		return
	}

	customer := mustCurrentCustomer(r.Context())

	phone, err := a.exchangeMiniAppPhoneCode(r.Context(), req.PhoneCode)
	if err != nil {
		a.writeError(w, r, http.StatusUnauthorized, 40105, "wechat phone code exchange failed")
		return
	}

	if err := a.store.updateCustomerPhone(customer.ID, phone); err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to update phone")
		return
	}

	customer.Phone = phone
	a.writeJSON(w, r, http.StatusOK, 0, "ok", toCustomerProfileDTO(customer))
}

func (a *App) handleCustomerLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	session, _ := currentCustomerSession(r.Context())
	a.store.deleteCustomerSession(session.Token)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", nil)
}

func (a *App) handleCustomerProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	customer := mustCurrentCustomer(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", toCustomerProfileDTO(customer))
}

// --- 顾客公开数据 ---

func (a *App) handleCustomerHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.customerHomeData())
}

func (a *App) handleCustomerProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	items, total := a.store.customerProductList(category, page, pageSize)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items":    items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"categories": a.store.customerProductCategories(),
	})
}

func (a *App) handleCustomerProductDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/customer/products/")
	if id == "" {
		a.writeError(w, r, http.StatusBadRequest, 40001, "missing product id")
		return
	}

	product, ok := a.store.customerProductDetail(id)
	if !ok {
		a.writeError(w, r, http.StatusNotFound, 40401, "product not found")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", product)
}

func (a *App) handleCustomerStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.customerStoreList())
}

func (a *App) handleCustomerStoreDetail(w http.ResponseWriter, r *http.Request, storeID string) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	if storeID == "" {
		a.writeError(w, r, http.StatusBadRequest, 40001, "missing store id")
		return
	}

	store, ok := a.store.customerStoreDetail(storeID)
	if !ok {
		a.writeError(w, r, http.StatusNotFound, 40401, "store not found")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", store)
}

// handleCustomerStoreActions 路由 /api/v1/customer/stores/{id}/slots
func (a *App) handleCustomerStoreActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/customer/stores/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		a.writeError(w, r, http.StatusBadRequest, 40001, "missing store id")
		return
	}

	parts := strings.SplitN(path, "/", 2)
	storeID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "":
		a.handleCustomerStoreDetail(w, r, storeID)
	case "slots":
		a.customerStoreSlots(w, r, storeID)
	default:
		a.writeError(w, r, http.StatusNotFound, 40402, "unknown action")
	}
}

func (a *App) customerStoreSlots(w http.ResponseWriter, r *http.Request, storeID string) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	date := strings.TrimSpace(r.URL.Query().Get("date"))
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	slots, err := a.store.listCustomerAppointmentSlots(storeID, date)
	if err != nil {
		if errors.Is(err, errAppointmentNotFound) {
			a.writeError(w, r, http.StatusNotFound, 40401, "store not found")
			return
		}
		a.writeError(w, r, http.StatusBadRequest, 40001, err.Error())
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"storeId": storeID,
		"date":    date,
		"slots":   slots,
	})
}

func (a *App) handleCustomerRecycleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.customerRecycleInfoData())
}

// --- 顾客预约 ---

func (a *App) handleCustomerAppointments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.customerListAppointments(w, r)
	case http.MethodPost:
		a.customerCreateAppointment(w, r)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) customerListAppointments(w http.ResponseWriter, r *http.Request) {
	customer := mustCurrentCustomer(r.Context())
	appointments := a.store.listCustomerAppointments(customer.ID)

	items := make([]CustomerAppointmentDTO, 0, len(appointments))
	for _, appt := range appointments {
		items = append(items, buildCustomerAppointmentDTO(appt))
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

func (a *App) customerCreateAppointment(w http.ResponseWriter, r *http.Request) {
	customer := mustCurrentCustomer(r.Context())

	var req CreateAppointmentRequest
	if err := decodeJSONLenient(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	if strings.TrimSpace(req.StoreID) == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "missing storeId")
		return
	}
	if strings.TrimSpace(req.AppointmentDate) == "" || strings.TrimSpace(req.AppointmentTime) == "" {
		a.writeError(w, r, http.StatusBadRequest, 40003, "missing appointment date or time")
		return
	}
	if strings.TrimSpace(req.ContactName) == "" {
		a.writeError(w, r, http.StatusBadRequest, 40004, "missing contactName")
		return
	}
	if strings.TrimSpace(req.ContactPhone) == "" {
		a.writeError(w, r, http.StatusBadRequest, 40005, "missing contactPhone")
		return
	}

	appt, err := a.store.createCustomerAppointment(customer, req)
	if err != nil {
		switch {
		case errors.Is(err, errInvalidAppointmentTime):
			a.writeError(w, r, http.StatusBadRequest, 40010, "invalid appointment time: must be at least 1 hour from now and within 7 days")
		case errors.Is(err, errDuplicateAppointment):
			a.writeError(w, r, http.StatusConflict, 40901, "duplicate appointment for same store and time slot")
		case errors.Is(err, errAppointmentNotFound):
			a.writeError(w, r, http.StatusNotFound, 40401, "store not found")
		default:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to create appointment")
		}
		return
	}

	a.writeJSON(w, r, http.StatusCreated, 0, "ok", buildCustomerAppointmentDTO(appt))
}

func (a *App) handleCustomerAppointmentActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/customer/appointments/")
	if path == "" {
		a.writeError(w, r, http.StatusBadRequest, 40001, "missing appointment id")
		return
	}

	parts := strings.SplitN(path, "/", 2)
	appointmentID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch {
	case action == "" && r.Method == http.MethodGet:
		a.customerGetAppointment(w, r, appointmentID)
	case action == "cancel" && r.Method == http.MethodPost:
		a.customerCancelAppointment(w, r, appointmentID)
	case action == "notes" && r.Method == http.MethodPut:
		a.customerUpdateAppointmentNotes(w, r, appointmentID)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) customerGetAppointment(w http.ResponseWriter, r *http.Request, appointmentID string) {
	customer := mustCurrentCustomer(r.Context())
	appt, ok := a.store.getCustomerAppointment(customer.ID, appointmentID)
	if !ok {
		a.writeError(w, r, http.StatusNotFound, 40401, "appointment not found")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", buildCustomerAppointmentDTO(appt))
}

func (a *App) customerCancelAppointment(w http.ResponseWriter, r *http.Request, appointmentID string) {
	customer := mustCurrentCustomer(r.Context())

	var req CancelAppointmentRequest
	_ = decodeJSONLenient(r, &req)

	err := a.store.cancelCustomerAppointment(customer.ID, appointmentID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, errAppointmentNotFound):
			a.writeError(w, r, http.StatusNotFound, 40401, "appointment not found")
		case errors.Is(err, errAppointmentCancelled):
			a.writeError(w, r, http.StatusConflict, 40902, "appointment already cancelled or completed")
		case errors.Is(err, errAppointmentTimeTooLate):
			a.writeError(w, r, http.StatusBadRequest, 40011, "cannot cancel within 2 hours of appointment")
		case errors.Is(err, errAppointmentStatusFlow):
			a.writeError(w, r, http.StatusConflict, 40903, "appointment status transition not allowed")
		default:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to cancel appointment")
		}
		return
	}

	appt, _ := a.store.getCustomerAppointment(customer.ID, appointmentID)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", buildCustomerAppointmentDTO(appt))
}

func (a *App) customerUpdateAppointmentNotes(w http.ResponseWriter, r *http.Request, appointmentID string) {
	customer := mustCurrentCustomer(r.Context())

	var req UpdateAppointmentNotesRequest
	if err := decodeJSONLenient(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if len(req.Remark) > 200 {
		a.writeError(w, r, http.StatusBadRequest, 40002, "remark too long (max 200)")
		return
	}

	err := a.store.updateCustomerAppointmentNotes(customer.ID, appointmentID, req.Remark)
	if err != nil {
		switch {
		case errors.Is(err, errAppointmentNotFound):
			a.writeError(w, r, http.StatusNotFound, 40401, "appointment not found")
		case errors.Is(err, errAppointmentStatusFlow):
			a.writeError(w, r, http.StatusConflict, 40903, "cannot edit notes: appointment is not pending or confirmed")
		default:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to update appointment notes")
		}
		return
	}

	appt, _ := a.store.getCustomerAppointment(customer.ID, appointmentID)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", buildCustomerAppointmentDTO(appt))
}

// --- 员工预约管理 ---

func (a *App) handleStaffAppointments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	appointments := a.store.listStaffAppointments(user)

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	items := make([]StaffAppointmentDTO, 0, len(appointments))
	for _, appt := range appointments {
		if statusFilter != "" && appt.Status != statusFilter {
			continue
		}
		items = append(items, buildStaffAppointmentDTO(appt))
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items": items,
		"total": len(items),
	})
}

func (a *App) handleStaffAppointmentActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/staff/appointments/")
	if path == "" {
		a.writeError(w, r, http.StatusBadRequest, 40001, "missing appointment id")
		return
	}

	parts := strings.SplitN(path, "/", 2)
	appointmentID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch {
	case action == "" && r.Method == http.MethodGet:
		a.staffGetAppointment(w, r, appointmentID)
	case action == "status" && r.Method == http.MethodPut:
		a.staffUpdateAppointmentStatus(w, r, appointmentID)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) staffGetAppointment(w http.ResponseWriter, r *http.Request, appointmentID string) {
	user := currentUser(r.Context())
	appt, ok := a.store.staffGetAppointment(user, appointmentID)
	if !ok {
		a.writeError(w, r, http.StatusNotFound, 40401, "appointment not found")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", buildStaffAppointmentDTO(appt))
}

type staffUpdateAppointmentStatusRequest struct {
	Status string `json:"status"`
}

func (a *App) staffUpdateAppointmentStatus(w http.ResponseWriter, r *http.Request, appointmentID string) {
	user := currentUser(r.Context())

	var req staffUpdateAppointmentStatusRequest
	if err := decodeJSONLenient(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	req.Status = strings.TrimSpace(strings.ToUpper(req.Status))
	validStatuses := map[string]bool{
		AppointmentStatusConfirmed:  true,
		AppointmentStatusArrived:    true,
		AppointmentStatusCompleted:  true,
		AppointmentStatusNoShow:     true,
		AppointmentStatusTerminated: true,
	}
	if !validStatuses[req.Status] {
		a.writeError(w, r, http.StatusBadRequest, 40006, "invalid status value")
		return
	}

	err := a.store.staffUpdateAppointmentStatus(user, appointmentID, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, errAppointmentNotFound):
			a.writeError(w, r, http.StatusNotFound, 40401, "appointment not found")
		case errors.Is(err, errAppointmentStatusFlow):
			a.writeError(w, r, http.StatusConflict, 40903, "appointment status transition not allowed")
		default:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to update appointment status")
		}
		return
	}

	appt, _ := a.store.staffGetAppointment(user, appointmentID)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", buildStaffAppointmentDTO(appt))
}
