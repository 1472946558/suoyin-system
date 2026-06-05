package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type miniAppLoginRequest struct {
	Code               string `json:"code"`
	PhoneCode          string `json:"phoneCode"`
	PhoneEncryptedData string `json:"phoneEncryptedData"`
	PhoneIV            string `json:"phoneIv"`
	Profile            struct {
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		RoleKey   string `json:"roleKey"`
		StoreName string `json:"storeName"`
		StoreCode string `json:"storeCode"`
	} `json:"profile"`
}

type cashierCreateRequest struct {
	StoreID       string             `json:"storeId"`
	CustomerName  string             `json:"customerName"`
	CustomerPhone string             `json:"customerPhone"`
	PaymentMethod string             `json:"paymentMethod"`
	Remark        string             `json:"remark"`
	Items         []CashierOrderLine `json:"items"`
}

type recycleDraftRequest struct {
	StoreID         string        `json:"storeId"`
	CustomerName    string        `json:"customerName"`
	CustomerPhone   string        `json:"customerPhone"`
	EstimatedAmount float64       `json:"estimatedAmount"`
	Items           []RecycleItem `json:"items"`
	AttachmentURLs  []string      `json:"attachmentUrls"`
	Remark          string        `json:"remark"`
}

type recycleConfirmRequest struct {
	ConfirmedAmount float64  `json:"confirmedAmount"`
	AttachmentURLs  []string `json:"attachmentUrls"`
	Remark          string   `json:"remark"`
}

type recycleQuotePreviewRequest struct {
	StoreID             string  `json:"storeId"`
	Category            string  `json:"category"`
	Purity              string  `json:"purity"`
	GrossWeightGram     float64 `json:"grossWeightGram"`
	DeductionWeightGram float64 `json:"deductionWeightGram"`
	WeightGram          float64 `json:"weightGram"`
	UnitPrice           float64 `json:"unitPrice"`
}

func attachmentURLCount(urls []string) int {
	count := 0
	for _, rawURL := range urls {
		if strings.TrimSpace(rawURL) != "" {
			count++
		}
	}
	return count
}

type recycleQuotePreviewResult struct {
	StoreID             string  `json:"storeId"`
	Category            string  `json:"category"`
	Purity              string  `json:"purity"`
	GrossWeightGram     float64 `json:"grossWeightGram"`
	DeductionWeightGram float64 `json:"deductionWeightGram"`
	NetWeightGram       float64 `json:"netWeightGram"`
	UnitPrice           float64 `json:"unitPrice"`
	ServiceFeeRate      float64 `json:"serviceFeeRate"`
	SubtotalAmount      float64 `json:"subtotalAmount"`
	ServiceFee          float64 `json:"serviceFee"`
	EstimatedAmount     float64 `json:"estimatedAmount"`
	PriceSource         string  `json:"priceSource"`
	RuleNote            string  `json:"ruleNote"`
}

type uploadPrepareRequest struct {
	StoreID     string `json:"storeId"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

type uploadCompleteRequest struct {
	UploadID     string `json:"uploadId"`
	OrderID      string `json:"orderId"`
	PublicURL    string `json:"publicUrl"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"service":    "gold-recycle-miniapp-backend",
		"version":    "v0-persistent-week2",
		"mode":       a.Config.Mode,
		"serverTime": time.Now().Format(time.RFC3339),
	})
}

func (a *App) handleRecyclePhotoUploadPrepare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	user := currentUser(r.Context())
	var req uploadPrepareRequest
	if err := decodeJSONLenient(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if strings.TrimSpace(req.StoreID) == "" || strings.TrimSpace(req.FileName) == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "storeId and fileName are required")
		return
	}

	preparation, err := a.store.prepareRecycleAttachmentUpload(user, req.StoreID, req.FileName, req.ContentType, req.SizeBytes)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50009, "failed to prepare upload")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", preparation)
	}
}

func (a *App) handleRecyclePhotoUploadComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	user := currentUser(r.Context())
	var req uploadCompleteRequest
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if strings.TrimSpace(req.UploadID) == "" || strings.TrimSpace(req.OrderID) == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "uploadId and orderId are required")
		return
	}

	asset, err := a.store.completeRecycleAttachmentUpload(user, req.UploadID, req.OrderID, req.PublicURL, req.ThumbnailURL)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errUploadNotFound):
		a.writeError(w, r, http.StatusNotFound, 40403, "upload session not found")
	case errors.Is(err, errRecycleNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "recycle order not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50010, "failed to complete upload")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", asset)
	}
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	user, err := a.store.authenticate(strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		a.writeError(w, r, http.StatusUnauthorized, 40103, "username or password is incorrect")
		return
	}

	session, err := a.store.createSession(a.Config.TokenSecret, user)
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50006, "failed to create login session")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"token":        session.Token,
		"expiresAt":    session.ExpiresAt,
		"user":         buildUserProfile(user),
		"accountHints": []string{"boss / Boss123!", "boss.demo / Boss123!", "manager.sz / Manager123!", "cashier.sz / Cashier123!"},
		"dataScopes":   []string{"org_all", "assigned_stores"},
	})
}

func (a *App) handleMiniAppLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	var req miniAppLoginRequest
	if err := decodeJSONLenient(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	user, err := a.authenticateMiniAppUser(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, errMiniAppCodeMissing):
			a.writeError(w, r, http.StatusBadRequest, 40002, "missing wechat login code")
		case errors.Is(err, errMiniAppAuthNotConfigured):
			a.writeError(w, r, http.StatusServiceUnavailable, 50301, "miniapp wechat auth is not configured")
		case errors.Is(err, errMiniAppUserNotBound):
			a.writeJSON(w, r, http.StatusForbidden, 40304, "wechat account is not bound to any operator", map[string]interface{}{
				"nextStep": "get_phone_number",
			})
		case errors.Is(err, errMiniAppCodeExchange):
			a.writeError(w, r, http.StatusUnauthorized, 40104, "wechat login code exchange failed")
		case errors.Is(err, errMiniAppPhoneCodeExchange), errors.Is(err, errMiniAppPhoneDataDecrypt):
			a.writeError(w, r, http.StatusUnauthorized, 40105, "wechat phone code exchange failed")
		default:
			a.writeError(w, r, http.StatusUnauthorized, 40103, "miniapp login failed")
		}
		return
	}

	roleKey := publicRoleCode(user.RoleCode)
	storeName := strings.TrimSpace(req.Profile.StoreName)
	storeCode := strings.TrimSpace(req.Profile.StoreCode)
	storeID := ""
	visibleStores := a.store.listStoresForUser(user)
	if len(visibleStores) > 0 {
		storeID = visibleStores[0].ID
		if storeName == "" {
			storeName = visibleStores[0].Name
		}
		if storeCode == "" {
			storeCode = visibleStores[0].Code
		}
	}

	session, err := a.store.createSession(a.Config.TokenSecret, user)
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50006, "failed to create login session")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"token":       session.Token,
		"userId":      user.ID,
		"id":          user.ID,
		"name":        user.DisplayName,
		"roleKey":     roleKey,
		"roleName":    publicRoleName(user),
		"phone":       user.Phone,
		"storeId":     storeID,
		"storeName":   storeName,
		"storeCode":   storeCode,
		"permissions": user.Permissions,
	})
}

func (a *App) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	user, err := a.store.authenticate(strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		a.writeError(w, r, http.StatusUnauthorized, 40103, "username or password is incorrect")
		return
	}

	session, err := a.store.createSession(a.Config.TokenSecret, user)
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50006, "failed to create login session")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", AdminLoginResult{
		Token:       session.Token,
		User:        a.store.adminSessionUser(user),
		LandingPage: AdminPageDashboard,
	})
}

func (a *App) handleAdminBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.buildAdminBootstrap(user))
}

func (a *App) handleAdminRoleTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/roles/")
	roleID, err := strconv.Atoi(strings.Trim(path, "/"))
	if err != nil || roleID <= 0 {
		a.writeError(w, r, http.StatusBadRequest, 40002, "invalid role id")
		return
	}

	var req AdminRoleTemplate
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	role, err := a.store.updateAdminRoleTemplate(user, roleID, req)
	switch {
	case errors.Is(err, errAdminRoleNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "role template not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to update role template")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", role)
	}
}

func (a *App) handleAdminPrintTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	var req AdminPrintTemplate
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	updated := a.store.updateAdminPrintTemplate(user, req)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", updated)
}

func (a *App) handleAdminStoreCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	user := currentUser(r.Context())
	store, err := a.store.createAdminStore(user)
	if errors.Is(err, errUnauthorizedStore) {
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		return
	}
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to create store")
		return
	}
	a.writeJSON(w, r, http.StatusCreated, 0, "ok", store)
}

func (a *App) handleAdminStore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	storeID := strings.TrimPrefix(r.URL.Path, "/api/admin/stores/")
	storeID = strings.Trim(storeID, "/")
	if storeID == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "invalid store id")
		return
	}

	var req AdminStoreRecord
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	store, err := a.store.updateAdminStore(user, storeID, req)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errAdminStoreNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "store not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to update store")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", store)
	}
}

func (a *App) handleAdminUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	userID := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	userID = strings.Trim(userID, "/")
	if userID == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "invalid user id")
		return
	}

	var req AdminUserAccount
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	account, err := a.store.updateAdminUser(user, userID, req)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errAdminUserNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "user not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to update user")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", account)
	}
}

func (a *App) handleAdminUserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	user := currentUser(r.Context())
	account, err := a.store.createAdminUser(user)
	if errors.Is(err, errUnauthorizedStore) {
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		return
	}
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to create user")
		return
	}
	a.writeJSON(w, r, http.StatusCreated, 0, "ok", account)
}

func (a *App) handleAdminMiniAppPendingBindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items": a.store.listPendingMiniAppBindings(),
	})
}

func (a *App) handleAdminProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	productID := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
	productID = strings.Trim(productID, "/")
	if productID == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "invalid product id")
		return
	}

	var req AdminProductRecord
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	product, err := a.store.updateAdminProduct(user, productID, req)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errAdminProductNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "product not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to update product")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", product)
	}
}

func (a *App) handleAdminProductCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	user := currentUser(r.Context())
	product, err := a.store.createAdminProduct(user)
	if errors.Is(err, errUnauthorizedStore) {
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		return
	}
	if err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to create product")
		return
	}
	a.writeJSON(w, r, http.StatusCreated, 0, "ok", product)
}

func (a *App) handleAdminSystemProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	var req AdminSystemProfile
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	updated := a.store.updateAdminSystemProfile(user, req)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", updated)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	session := currentSession(r.Context())
	if err := a.store.deleteSession(session.Token); err != nil {
		a.writeError(w, r, http.StatusInternalServerError, 50007, "failed to clear session")
		return
	}
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]bool{"loggedOut": true})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", buildUserProfile(user))
}

func (a *App) handleStores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items":     a.store.listStoresForUser(user),
		"dataScope": user.DataScope,
	})
}

func (a *App) handleRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items": a.store.listRoles(),
	})
}

func (a *App) handlePermissionCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items": a.store.permissionGroups(),
	})
}

func (a *App) handleDashboardSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.dashboardSummary(user))
}

func (a *App) handleDailyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	date := strings.TrimSpace(r.URL.Query().Get("date"))
	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.dailyReport(user, date))
}

func (a *App) handleInventorySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.inventorySummary(user))
}

func (a *App) handleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items":     a.store.listMembersForUser(user),
		"dataScope": user.DataScope,
	})
}

func (a *App) handleMemberDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	memberID := strings.TrimPrefix(r.URL.Path, "/api/v1/members/")
	memberID = strings.Trim(memberID, "/")
	if memberID == "" {
		a.writeError(w, r, http.StatusNotFound, 40401, "resource not found")
		return
	}

	member, err := a.store.getMemberForUser(user, memberID)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errMemberNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "member not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to load member")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", member)
	}
}

func (a *App) handleCatalogProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"items":     a.store.listCatalogProductsForUser(user),
		"dataScope": user.DataScope,
	})
}

func (a *App) handleCatalogProductDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	productID := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	productID = strings.Trim(productID, "/")
	if productID == "" {
		a.writeError(w, r, http.StatusNotFound, 40401, "resource not found")
		return
	}

	product, err := a.store.getCatalogProductForUser(user, productID)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errProductNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "product not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to load product")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", product)
	}
}

func (a *App) handleRecycleQuotePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	var req recycleQuotePreviewRequest
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	storeID := strings.TrimSpace(req.StoreID)
	if storeID != "" && !a.store.canAccessStore(user, storeID) {
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		return
	}

	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "金饰"
	}
	purity := strings.TrimSpace(req.Purity)
	if purity == "" {
		purity = "足金999"
	}

	grossWeight := round2(req.GrossWeightGram)
	deductionWeight := round2(req.DeductionWeightGram)
	netWeight := round2(req.WeightGram)
	if netWeight <= 0 {
		netWeight = round2(grossWeight - deductionWeight)
	}
	if grossWeight <= 0 && netWeight > 0 {
		grossWeight = netWeight
	}
	if netWeight <= 0 {
		a.writeError(w, r, http.StatusBadRequest, 40003, "weight must be greater than 0")
		return
	}
	if deductionWeight < 0 {
		a.writeError(w, r, http.StatusBadRequest, 40003, "deductionWeightGram cannot be negative")
		return
	}

	unitPrice := round2(req.UnitPrice)
	priceSource := "request"
	if unitPrice <= 0 {
		unitPrice, priceSource = a.lookupRecycleUnitPrice(user, category, purity)
	}
	serviceFeeRate := recycleServiceFeeRate(category)
	subtotalAmount := round2(netWeight * unitPrice)
	serviceFee := round2(netWeight * serviceFeeRate)
	estimatedAmount := subtotalAmount - serviceFee
	if estimatedAmount < 0 {
		estimatedAmount = 0
	}
	estimatedAmount = round2(estimatedAmount)

	a.writeJSON(w, r, http.StatusOK, 0, "ok", recycleQuotePreviewResult{
		StoreID:             storeID,
		Category:            category,
		Purity:              purity,
		GrossWeightGram:     grossWeight,
		DeductionWeightGram: deductionWeight,
		NetWeightGram:       netWeight,
		UnitPrice:           unitPrice,
		ServiceFeeRate:      serviceFeeRate,
		SubtotalAmount:      subtotalAmount,
		ServiceFee:          serviceFee,
		EstimatedAmount:     estimatedAmount,
		PriceSource:         priceSource,
		RuleNote:            "按目录基准金价和门店服务费率预估，最终金额以复秤和确认单为准",
	})
}

func (a *App) lookupRecycleUnitPrice(user UserAccount, category, purity string) (float64, string) {
	products := a.store.listCatalogProductsForUser(user)
	var categoryFallback float64
	var purityFallback float64
	for _, product := range products {
		if product.BenchPrice <= 0 || product.Status == "disabled" {
			continue
		}
		if product.Category == category && product.Purity == purity {
			return round2(product.BenchPrice), "catalog"
		}
		if product.Purity == purity && purityFallback <= 0 {
			purityFallback = product.BenchPrice
		}
		if product.Category == category && categoryFallback <= 0 {
			categoryFallback = product.BenchPrice
		}
	}
	if purityFallback > 0 {
		return round2(purityFallback), "catalog_purity"
	}
	if categoryFallback > 0 {
		return round2(categoryFallback), "catalog_category"
	}
	return round2(defaultRecyclePrice(purity)), "reference"
}

func recycleServiceFeeRate(category string) float64 {
	if strings.Contains(category, "金条") || strings.Contains(category, "投资金") {
		return 1.2
	}
	return 2.5
}

func defaultRecyclePrice(purity string) float64 {
	switch purity {
	case "足金9999":
		return 748
	case "22K":
		return 680
	case "18K":
		return 558
	case "14K":
		return 436
	default:
		return 742
	}
}

func (a *App) handleCashierOrders(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())

	switch r.Method {
	case http.MethodGet:
		if !hasPermission(user, "cashier.order.read") {
			a.writeError(w, r, http.StatusForbidden, 40301, "missing permission: cashier.order.read")
			return
		}
		storeID := strings.TrimSpace(r.URL.Query().Get("storeId"))
		if storeID != "" && !a.store.canAccessStore(user, storeID) {
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
			return
		}
		items, err := a.store.listCashierOrders(user, storeID)
		if err != nil {
			a.writeError(w, r, http.StatusInternalServerError, 50001, "failed to load cashier orders")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
			"items":     items,
			"storeId":   storeID,
			"dataScope": user.DataScope,
		})
	case http.MethodPost:
		if !hasPermission(user, "cashier.order.create") {
			a.writeError(w, r, http.StatusForbidden, 40301, "missing permission: cashier.order.create")
			return
		}

		var req cashierCreateRequest
		if err := decodeJSON(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if strings.TrimSpace(req.StoreID) == "" || strings.TrimSpace(req.CustomerName) == "" || strings.TrimSpace(req.CustomerPhone) == "" || len(req.Items) == 0 {
			a.writeError(w, r, http.StatusBadRequest, 40002, "storeId, customerName, customerPhone and items are required")
			return
		}
		for _, item := range req.Items {
			if strings.TrimSpace(item.Name) == "" || item.Quantity <= 0 || item.UnitPrice < 0 {
				a.writeError(w, r, http.StatusBadRequest, 40003, "invalid cashier order item")
				return
			}
		}
		order, err := a.store.createCashierOrder(user, req.StoreID, req.CustomerName, req.CustomerPhone, req.PaymentMethod, req.Items, req.Remark)
		if errors.Is(err, errUnauthorizedStore) {
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
			return
		}
		if err != nil {
			a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to create cashier order")
			return
		}
		a.writeJSON(w, r, http.StatusCreated, 0, "ok", order)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleCashierOrderDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	user := currentUser(r.Context())
	orderID := strings.TrimPrefix(r.URL.Path, "/api/v1/cashier/orders/")
	orderID = strings.Trim(orderID, "/")
	if orderID == "" {
		a.writeError(w, r, http.StatusNotFound, 40401, "resource not found")
		return
	}

	order, err := a.store.getCashierOrder(user, orderID)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errCashierNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "cashier order not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50002, "failed to load cashier order")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", order)
	}
}

func (a *App) handleRecycleOrders(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())

	switch r.Method {
	case http.MethodGet:
		if !hasPermission(user, "recycle.order.read") {
			a.writeError(w, r, http.StatusForbidden, 40301, "missing permission: recycle.order.read")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
			"items":     a.store.listRecycleOrders(user),
			"dataScope": user.DataScope,
		})
	case http.MethodPost:
		if !hasPermission(user, "recycle.order.draft") {
			a.writeError(w, r, http.StatusForbidden, 40301, "missing permission: recycle.order.draft")
			return
		}

		var req recycleDraftRequest
		if err := decodeJSON(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		if strings.TrimSpace(req.StoreID) == "" || strings.TrimSpace(req.CustomerName) == "" || len(req.Items) == 0 {
			a.writeError(w, r, http.StatusBadRequest, 40002, "storeId, customerName and items are required")
			return
		}
		for _, item := range req.Items {
			if strings.TrimSpace(item.Category) == "" || item.WeightGram <= 0 {
				a.writeError(w, r, http.StatusBadRequest, 40003, "invalid recycle item")
				return
			}
		}
		order, err := a.store.createRecycleDraft(user, req.StoreID, req.CustomerName, req.CustomerPhone, req.EstimatedAmount, req.Items, req.AttachmentURLs, req.Remark)
		if errors.Is(err, errUnauthorizedStore) {
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
			return
		}
		if err != nil {
			a.writeError(w, r, http.StatusInternalServerError, 50003, "failed to create recycle draft")
			return
		}
		a.writeJSON(w, r, http.StatusCreated, 0, "ok", order)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleRecycleOrderActions(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())
	if r.Method == http.MethodGet {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/recycle/orders/")
		orderID := strings.Trim(path, "/")
		if orderID == "" {
			a.writeError(w, r, http.StatusNotFound, 40401, "resource not found")
			return
		}
		order, ok := a.store.getRecycleOrder(orderID)
		if !ok {
			a.writeError(w, r, http.StatusNotFound, 40402, "recycle order not found")
			return
		}
		if !a.store.canAccessStore(user, order.StoreID) {
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", order)
		return
	}
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	if !hasPermission(user, "recycle.order.confirm") {
		a.writeError(w, r, http.StatusForbidden, 40301, "missing permission: recycle.order.confirm")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/recycle/orders/")
	if !strings.HasSuffix(path, "/confirm") {
		a.writeError(w, r, http.StatusNotFound, 40401, "resource not found")
		return
	}
	orderID := strings.TrimSuffix(path, "/confirm")
	orderID = strings.Trim(orderID, "/")
	if orderID == "" {
		a.writeError(w, r, http.StatusNotFound, 40401, "resource not found")
		return
	}

	var req recycleConfirmRequest
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if req.ConfirmedAmount <= 0 {
		a.writeError(w, r, http.StatusBadRequest, 40004, "confirmedAmount must be greater than 0")
		return
	}

	settings := a.store.getSettings()
	attachments := req.AttachmentURLs
	if len(attachments) == 0 {
		if existing, ok := a.store.getRecycleOrder(orderID); ok {
			attachments = existing.AttachmentURLs
		}
	}
	photoCount := attachmentURLCount(attachments)
	if photoCount < settings.Recycle.MinPhotoCount {
		a.writeError(w, r, http.StatusConflict, 40901, "attachment count below recycle requirement")
		return
	}
	if settings.Recycle.RequireExactThree && photoCount != 3 {
		a.writeError(w, r, http.StatusConflict, 40902, "recycle confirmation requires exactly 3 photos")
		return
	}
	if photoCount > settings.Recycle.MaxPhotoCount {
		a.writeError(w, r, http.StatusConflict, 40903, "attachment count exceeds recycle requirement")
		return
	}

	order, err := a.store.confirmRecycleOrder(user, orderID, req.ConfirmedAmount, req.AttachmentURLs, req.Remark)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errRecycleNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "recycle order not found")
	case errors.Is(err, errRecycleConflict):
		a.writeError(w, r, http.StatusConflict, 40904, "recycle order is not in draft status")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50004, "failed to confirm recycle order")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", order)
	}
}

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())

	switch r.Method {
	case http.MethodGet:
		if !hasPermission(user, "settings.read") {
			a.writeError(w, r, http.StatusForbidden, 40301, "missing permission: settings.read")
			return
		}
		a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.getSettings())
	case http.MethodPut:
		if !hasPermission(user, "settings.update") {
			a.writeError(w, r, http.StatusForbidden, 40301, "missing permission: settings.update")
			return
		}
		var req SystemSettings
		if err := decodeJSON(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		updated := a.store.updateSettings(user, req)
		a.writeJSON(w, r, http.StatusOK, 0, "ok", updated)
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleNotFound(w http.ResponseWriter, r *http.Request) {
	a.writeError(w, r, http.StatusNotFound, 40401, "resource not found")
}

func buildUserProfile(user UserAccount) map[string]interface{} {
	return map[string]interface{}{
		"id":          user.ID,
		"orgId":       user.OrgID,
		"username":    user.Username,
		"displayName": user.DisplayName,
		"phone":       user.Phone,
		"roleCode":    publicRoleCode(user.RoleCode),
		"roleName":    publicRoleName(user),
		"dataScope":   user.DataScope,
		"storeIds":    user.StoreIDs,
		"permissions": user.Permissions,
		"status":      user.Status,
	}
}

func publicRoleCode(roleCode string) string {
	switch strings.TrimSpace(roleCode) {
	case "owner":
		return "owner"
	case "manager":
		return "manager"
	default:
		return "staff"
	}
}

func publicRoleName(user UserAccount) string {
	switch publicRoleCode(user.RoleCode) {
	case "owner":
		return "老板"
	case "manager":
		return "店长"
	default:
		return "员工"
	}
}

func hasPermission(user UserAccount, code string) bool {
	for _, permission := range user.Permissions {
		if permission == code {
			return true
		}
	}
	return false
}

func decodeJSON(r *http.Request, target interface{}) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func decodeJSONLenient(r *http.Request, target interface{}) error {
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
