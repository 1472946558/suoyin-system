/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: admin_customer_handlers.go
 * 功能描述: 管理后台「顾客端内容管理」HTTP handler（上传、Banner、首页文案、款式、门店扩展、预约规则、预约记录）
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// hasAdminAbility handler 内二次校验 ability（用于同一路由不同方法不同权限的场景）
func (a *App) hasAdminAbility(r *http.Request, code AdminAbilityCode) bool {
	user := currentUser(r.Context())
	if user.ID == "" {
		return false
	}
	sessionUser := a.store.adminSessionUser(user)
	for _, ability := range sessionUser.Abilities {
		if ability == code {
			return true
		}
	}
	return false
}

// requireAdminAbilityMethod 写操作要求更高 ability，否则 403
func (a *App) requireAdminAbilityMethod(w http.ResponseWriter, r *http.Request, code AdminAbilityCode) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	if a.hasAdminAbility(r, code) {
		return true
	}
	a.writeError(w, r, http.StatusForbidden, 40311, fmt.Sprintf("missing admin ability: %s", code))
	return false
}

// --- 图片上传 ---

const adminUploadMaxBytes = 5 << 20 // 5MB

var adminUploadScenes = map[string]bool{
	"banner": true, "style_main": true, "style_detail": true, "store": true, "service_intro": true,
}

// sniffImageFormat 按文件头魔数判断图片格式，返回扩展名（不信任扩展名/Content-Type）
func sniffImageFormat(head []byte) string {
	if len(head) >= 3 && head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF {
		return "jpg"
	}
	if len(head) >= 8 && head[0] == 0x89 && head[1] == 0x50 && head[2] == 0x4E && head[3] == 0x47 {
		return "png"
	}
	if len(head) >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WEBP" {
		return "webp"
	}
	return ""
}

func (a *App) handleAdminUploadImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, adminUploadMaxBytes+64<<10)
	if err := r.ParseMultipartForm(adminUploadMaxBytes + 64<<10); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40012, "image exceeds 5MB limit")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40010, "missing file field")
		return
	}
	defer file.Close()

	scene := strings.TrimSpace(r.FormValue("scene"))
	if !adminUploadScenes[scene] {
		a.writeError(w, r, http.StatusBadRequest, 40002, "invalid scene")
		return
	}
	refID := strings.TrimSpace(r.FormValue("refId"))

	head := make([]byte, 12)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to read file")
		return
	}
	ext := sniffImageFormat(head[:n])
	if ext == "" {
		a.writeError(w, r, http.StatusBadRequest, 40011, "unsupported image format (jpg/jpeg/png/webp only)")
		return
	}

	assetsDir := strings.TrimSpace(a.Config.AssetsDir)
	if assetsDir == "" {
		a.writeError(w, r, http.StatusInternalServerError, 50002, "assets dir not configured")
		return
	}
	now := time.Now()
	relPath := fmt.Sprintf("%s/%s/img-%s-%04d.%s", scene, now.Format("200601"), now.Format("20060102"), now.Nanosecond()%10000, ext)
	absPath := filepath.Join(assetsDir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to create assets dir")
		return
	}
	dst, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to write file")
		return
	}
	if _, err := dst.Write(head[:n]); err != nil {
		_ = dst.Close()
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to write file")
		return
	}
	size := int64(n)
	remaining, err := io.Copy(dst, io.LimitReader(file, adminUploadMaxBytes+1))
	if err != nil {
		_ = dst.Close()
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to write file")
		return
	}
	size += remaining
	if err := dst.Close(); err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to write file")
		return
	}
	if size > adminUploadMaxBytes {
		_ = os.Remove(absPath)
		a.writeError(w, r, http.StatusBadRequest, 40012, "image exceeds 5MB limit")
		return
	}

	publicBase := strings.TrimRight(strings.TrimSpace(a.Config.AssetsPublicBaseURL), "/")
	path := "/assets/" + relPath
	url := path
	if publicBase != "" {
		url = publicBase + path
	}

	user := currentUser(r.Context())
	asset := UploadAsset{
		ID:           fmt.Sprintf("img-%s-%04d", now.Format("20060102"), now.Nanosecond()%10000),
		Scene:        scene,
		URL:          url,
		Path:         path,
		Storage:      "local",
		MimeType:     "image/" + map[string]string{"jpg": "jpeg", "png": "png", "webp": "webp"}[ext],
		Size:         size,
		RefID:        refID,
		OriginalName: filepath.Base(header.Filename),
		CreatedBy:    user.DisplayName,
		CreatedAt:    now,
	}
	if err := a.store.registerUploadAsset(asset); err != nil {
		_ = os.Remove(absPath)
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to register asset")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", asset)
}

func (a *App) handleAdminUploadImageActions(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/admin/uploads/images/")
	id = strings.Trim(id, "/")
	if id == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "missing asset id")
		return
	}
	switch {
	case strings.HasSuffix(id, "/list") || id == "list":
		a.handleAdminUploadImageList(w, r)
		return
	case r.Method == http.MethodDelete:
		if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
			return
		}
		if err := a.store.deleteUploadAsset(currentUser(r.Context()), id); err != nil {
			switch {
			case errors.Is(err, errUploadAssetInUse):
				a.writeError(w, r, http.StatusConflict, 40013, "asset is still referenced")
			case errors.Is(err, errUploadAssetNotFound):
				a.writeError(w, r, http.StatusNotFound, 40401, "asset not found")
			default:
				a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to delete asset")
			}
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", nil)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleAdminUploadImageList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	scene := strings.TrimSpace(r.URL.Query().Get("scene"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	items, total := a.store.listUploadAssets(scene, page, pageSize)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items": items, "total": total, "page": page, "pageSize": pageSize,
	})
}

// --- Banner 管理 ---

func (a *App) handleAdminBannerCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.adminListBanners())
	case http.MethodPost:
		if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
			return
		}
		var req AdminBannerRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if strings.TrimSpace(req.ImageURL) == "" {
			a.writeError(w, r, http.StatusBadRequest, 40004, "imageUrl is required")
			return
		}
		banner, err := a.store.adminSaveBanner(currentUser(r.Context()), "", req)
		switch {
		case errors.Is(err, errBannerLimit):
			a.writeError(w, r, http.StatusBadRequest, 40015, "enabled banners exceed limit (8)")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to save banner")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", banner)
		}
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleAdminBannerItem(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/admin/customer-home/banners/")
	id = strings.Trim(id, "/")
	if id == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "missing banner id")
		return
	}
	switch r.Method {
	case http.MethodPut:
		if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
			return
		}
		var req AdminBannerRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		banner, err := a.store.adminSaveBanner(currentUser(r.Context()), id, req)
		switch {
		case errors.Is(err, errBannerNotFound):
			a.writeError(w, r, http.StatusNotFound, 40401, "banner not found")
		case errors.Is(err, errBannerLimit):
			a.writeError(w, r, http.StatusBadRequest, 40015, "enabled banners exceed limit (8)")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to save banner")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", banner)
		}
	case http.MethodDelete:
		if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
			return
		}
		if err := a.store.adminDeleteBanner(currentUser(r.Context()), id); err != nil {
			if errors.Is(err, errBannerNotFound) {
				a.writeError(w, r, http.StatusNotFound, 40401, "banner not found")
				return
			}
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to delete banner")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", nil)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

// --- 首页文案配置 ---

func (a *App) handleAdminHomeConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.adminGetHomeConfig())
	case http.MethodPut:
		if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
			return
		}
		var req AdminHomeConfigRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if strings.TrimSpace(req.BrandName) == "" {
			a.writeError(w, r, http.StatusBadRequest, 40004, "brandName is required")
			return
		}
		cfg, err := a.store.adminUpdateHomeConfig(currentUser(r.Context()), req)
		if err != nil {
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to update home config")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", cfg)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

// --- 回收服务介绍管理 ---

func (a *App) handleAdminCustomerRecycleInfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.adminGetCustomerRecycleInfo())
	case http.MethodPut:
		if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
			return
		}
		var req CustomerRecycleInfo
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		info, err := a.store.adminUpdateCustomerRecycleInfo(currentUser(r.Context()), req)
		switch {
		case errors.Is(err, errInvalidRecycleInfo):
			a.writeError(w, r, http.StatusBadRequest, 40004, "title is required")
		case errors.Is(err, errRecycleInfoV1Boundary):
			a.writeError(w, r, http.StatusBadRequest, 40006, "recycle info cannot contain gold price, online estimate, settlement, door-to-door or mail-in service content in V1")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to update recycle info")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", info)
		}
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

// --- 款式/工费管理 ---

func (a *App) handleAdminCustomerProductCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		category := strings.TrimSpace(r.URL.Query().Get("category"))
		status := strings.TrimSpace(r.URL.Query().Get("status"))
		keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 20
		}
		items, total := a.store.adminListCustomerProducts(category, status, keyword, page, pageSize)
		a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
			"items": items, "total": total, "page": page, "pageSize": pageSize,
		})
	case http.MethodPost:
		if !a.requireAdminAbilityMethod(w, r, "product.manage") {
			return
		}
		var req AdminCustomerProductRecord
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		record, err := a.store.adminSaveCustomerProduct(currentUser(r.Context()), "", req)
		switch {
		case errors.Is(err, errInvalidStyleInput):
			a.writeError(w, r, http.StatusBadRequest, 40004, "name and imageUrl are required")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to save style")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", record)
		}
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleAdminCustomerProductItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/customer-products/")
	path = strings.Trim(path, "/")
	if path == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "missing product id")
		return
	}
	// 子资源：款式分类
	if path == "categories" {
		a.handleAdminStyleCategories(w, r)
		return
	}
	productID := path

	switch r.Method {
	case http.MethodGet:
		record, ok := a.store.adminGetCustomerProduct(productID)
		if !ok {
			a.writeError(w, r, http.StatusNotFound, 40401, "style not found")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", record)
	case http.MethodPut:
		if !a.requireAdminAbilityMethod(w, r, "product.manage") {
			return
		}
		var req AdminCustomerProductRecord
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		record, err := a.store.adminSaveCustomerProduct(currentUser(r.Context()), productID, req)
		switch {
		case errors.Is(err, errInvalidStyleInput):
			a.writeError(w, r, http.StatusBadRequest, 40004, "name and imageUrl are required")
		case errors.Is(err, errProductNotFound):
			a.writeError(w, r, http.StatusNotFound, 40401, "style not found")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to save style")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", record)
		}
	case http.MethodDelete:
		if !a.requireAdminAbilityMethod(w, r, "product.manage") {
			return
		}
		if err := a.store.adminDeleteCustomerProduct(currentUser(r.Context()), productID); err != nil {
			if errors.Is(err, errProductNotFound) {
				a.writeError(w, r, http.StatusNotFound, 40401, "style not found")
				return
			}
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to delete style")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", nil)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

// handleAdminStyleCategories 款式分类（V1 复用商品 Category 字段聚合）
func (a *App) handleAdminStyleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.customerProductCategories())
}

// handleAdminStoreCustomerConfig 门店顾客端扩展配置（由 handleAdminStore 委托）
func (a *App) handleAdminStoreCustomerConfig(w http.ResponseWriter, r *http.Request, storeID string) {
	switch r.Method {
	case http.MethodGet:
		config, ok := a.store.adminGetStoreCustomerConfig(storeID)
		if !ok {
			a.writeError(w, r, http.StatusNotFound, 40402, "store not found")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", config)
	case http.MethodPut:
		if !a.requireAdminAbilityMethod(w, r, "store.manage") {
			return
		}
		var req AdminCustomerStoreConfig
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		config, err := a.store.adminUpdateStoreCustomerConfig(currentUser(r.Context()), storeID, req)
		switch {
		case errors.Is(err, errUnauthorizedStore):
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		case errors.Is(err, errAdminStoreNotFound):
			a.writeError(w, r, http.StatusNotFound, 40402, "store not found")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to update store config")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", config)
		}
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

// --- 预约规则 ---

func (a *App) handleAdminAppointmentRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.adminGetAppointmentRules())
	case http.MethodPut:
		if !a.requireAdminAbilityMethod(w, r, "customer_content.manage") {
			return
		}
		var req AdminAppointmentRulesRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		rules, err := a.store.adminUpdateAppointmentRules(currentUser(r.Context()), req)
		if err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40014, "invalid appointment rules value")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", rules)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

// --- 预约记录管理（后台） ---

func (a *App) handleAdminCustomerAppointments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	filter := AdminAppointmentFilter{
		Status:      strings.TrimSpace(strings.ToUpper(r.URL.Query().Get("status"))),
		StoreID:     strings.TrimSpace(r.URL.Query().Get("storeId")),
		Phone:       strings.TrimSpace(r.URL.Query().Get("phone")),
		ServiceType: strings.TrimSpace(strings.ToUpper(r.URL.Query().Get("serviceType"))),
		DateFrom:    strings.TrimSpace(r.URL.Query().Get("dateFrom")),
		DateTo:      strings.TrimSpace(r.URL.Query().Get("dateTo")),
		Page:        page,
		PageSize:    pageSize,
	}
	items, total := a.store.adminListAppointments(currentUser(r.Context()), filter)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items": items, "total": total, "page": page, "pageSize": pageSize,
	})
}

func (a *App) handleAdminCustomerAppointmentActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/customer-appointments/")
	path = strings.Trim(path, "/")
	if path == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "missing appointment id")
		return
	}
	parts := strings.SplitN(path, "/", 2)
	appointmentID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	user := currentUser(r.Context())

	if action == "" && r.Method == http.MethodGet {
		record, err := a.store.adminGetAppointment(user, appointmentID)
		switch {
		case errors.Is(err, errUnauthorizedStore):
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		case errors.Is(err, errAppointmentNotFound):
			a.writeError(w, r, http.StatusNotFound, 40401, "appointment not found")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to load appointment")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", record)
		}
		return
	}

	if action == "staff-note" {
		if r.Method != http.MethodPut {
			a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
			return
		}
		var req AdminStaffNoteRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if err := a.store.adminUpdateAppointmentStaffNote(user, appointmentID, req.StaffNote); err != nil {
			a.writeAdminAppointmentActionError(w, r, err)
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", nil)
		return
	}

	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	statusMap := map[string]string{
		"confirm":  AppointmentStatusConfirmed,
		"arrive":   AppointmentStatusArrived,
		"complete": AppointmentStatusCompleted,
		"no-show":  AppointmentStatusNoShow,
	}
	switch action {
	case "cancel":
		var req AdminCancelAppointmentRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if strings.TrimSpace(req.Reason) == "" {
			a.writeError(w, r, http.StatusBadRequest, 40004, "cancel reason is required")
			return
		}
		if err := a.store.adminCancelAppointment(user, appointmentID, req.Reason); err != nil {
			a.writeAdminAppointmentActionError(w, r, err)
			return
		}
	case "confirm", "arrive", "complete", "no-show":
		if err := a.store.staffUpdateAppointmentStatus(user, appointmentID, statusMap[action]); err != nil {
			a.writeAdminAppointmentActionError(w, r, err)
			return
		}
	default:
		a.writeError(w, r, http.StatusNotFound, 40401, "unknown appointment action")
		return
	}

	record, _ := a.store.adminGetAppointment(user, appointmentID)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", record)
}

func (a *App) writeAdminAppointmentActionError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, errAppointmentNotFound):
		a.writeError(w, r, http.StatusNotFound, 40401, "appointment not found")
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errAppointmentStatusFlow):
		a.writeError(w, r, http.StatusConflict, 40903, "appointment status transition not allowed")
	case errors.Is(err, errAppointmentPersistence):
		a.writeError(w, r, http.StatusServiceUnavailable, 50302, "appointment could not be saved; please retry")
	default:
		a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to update appointment")
	}
}
