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
	Code    string `json:"code"`
	Profile struct {
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		RoleKey   string `json:"roleKey"`
		StoreName string `json:"storeName"`
		StoreCode string `json:"storeCode"`
	} `json:"profile"`
}

type cashierCreateRequest struct {
	StoreID       string             `json:"storeId"`
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

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"service":    "gold-recycle-miniapp-backend",
		"version":    "v0-mock-week1",
		"mode":       "memory",
		"serverTime": time.Now().Format(time.RFC3339),
	})
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

	session := a.store.createSession(a.Config.TokenSecret, user)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"token":      session.Token,
		"expiresAt":  session.ExpiresAt,
		"user":       buildUserProfile(user),
		"mockHints":  []string{"boss / Boss123!", "manager.sz / Manager123!", "cashier.sz / Cashier123!"},
		"dataScopes": []string{"org_all", "assigned_stores"},
	})
}

func (a *App) handleMiniAppLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	var req miniAppLoginRequest
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	username := "cashier.sz"
	switch strings.TrimSpace(req.Profile.RoleKey) {
	case "owner":
		username = "boss"
	case "manager":
		username = "manager.sz"
	}

	user, err := a.store.authenticate(username, map[string]string{
		"boss":       "Boss123!",
		"manager.sz": "Manager123!",
		"cashier.sz": "Cashier123!",
	}[username])
	if err != nil {
		a.writeError(w, r, http.StatusUnauthorized, 40103, "mock miniapp login failed")
		return
	}

	roleKey := strings.TrimSpace(req.Profile.RoleKey)
	if roleKey == "" {
		roleKey = user.RoleCode
	}
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

	session := a.store.createSession(a.Config.TokenSecret, user)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", map[string]interface{}{
		"token":       session.Token,
		"userId":      user.ID,
		"id":          user.ID,
		"roleKey":     roleKey,
		"roleName":    user.RoleName,
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

	session := a.store.createSession(a.Config.TokenSecret, user)
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

func (a *App) handleAdminPaymentConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}

	user := currentUser(r.Context())
	var req AdminPaymentConfig
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}

	updated := a.store.updateAdminPaymentConfig(user, req)
	a.writeJSON(w, r, http.StatusOK, 0, "ok", updated)
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
	store := a.store.createAdminStore(user)
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
	account := a.store.createAdminUser(user)
	a.writeJSON(w, r, http.StatusCreated, 0, "ok", account)
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
	product := a.store.createAdminProduct(user)
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
	a.store.deleteSession(session.Token)
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
		if strings.TrimSpace(req.StoreID) == "" || len(req.Items) == 0 {
			a.writeError(w, r, http.StatusBadRequest, 40002, "storeId and items are required")
			return
		}
		for _, item := range req.Items {
			if strings.TrimSpace(item.Name) == "" || item.Quantity <= 0 || item.UnitPrice < 0 {
				a.writeError(w, r, http.StatusBadRequest, 40003, "invalid cashier order item")
				return
			}
		}
		if req.PaymentMethod == "" {
			req.PaymentMethod = "cash"
		}

		order, err := a.store.createCashierOrder(user, req.StoreID, req.Items, req.PaymentMethod, req.Remark)
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
	if len(attachments) < settings.Recycle.MinPhotoCount {
		a.writeError(w, r, http.StatusConflict, 40901, "attachment count below recycle requirement")
		return
	}
	if settings.Recycle.RequireExactThree && len(attachments) != 3 {
		a.writeError(w, r, http.StatusConflict, 40902, "recycle confirmation requires exactly 3 photos")
		return
	}
	if len(attachments) > settings.Recycle.MaxPhotoCount {
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
		"roleCode":    user.RoleCode,
		"roleName":    user.RoleName,
		"dataScope":   user.DataScope,
		"storeIds":    user.StoreIDs,
		"permissions": user.Permissions,
		"status":      user.Status,
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
