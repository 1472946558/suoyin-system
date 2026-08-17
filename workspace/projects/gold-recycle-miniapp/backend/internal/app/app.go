/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: app.go
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-05-10
 */

package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userKey      contextKey = "user"
	sessionKey   contextKey = "session"
)

type App struct {
	Config                     Config
	store                      *MockStore
	requestSeed                uint64
	wechatAccessTokenMu        sync.Mutex
	wechatAccessToken          string
	wechatAccessTokenExpiresAt time.Time
	goldReferenceMu            sync.Mutex
	goldReferenceSnapshot      GoldReferencePriceSnapshot
	goldReferenceFetchedAt     time.Time
}

func New() (*App, error) {
	mode := env("APP_MODE", "memory")
	storageEnabled, storageEnabledSet := envBoolOptional("STORAGE_ENABLED")
	storageCallbackEnabled, storageCallbackSet := envBoolOptional("STORAGE_CALLBACK_ENABLED")
	cfg := Config{
		Host:                   env("HOST", ""),
		Port:                   env("PORT", "8080"),
		Mode:                   mode,
		TokenSecret:            env("TOKEN_SECRET", "gold-recycle-dev-secret"),
		CORSOrigin:             env("CORS_ORIGIN", "*"),
		MySQLDSN:               mysqlDSNFromEnv(),
		RedisAddr:              env("REDIS_ADDR", ""),
		RedisUser:              env("REDIS_USER", ""),
		RedisPass:              env("REDIS_PASSWORD", ""),
		RedisDB:                envInt("REDIS_DB", 0),
		RedisPrefix:            env("REDIS_KEY_PREFIX", "gold:"),
		WechatMiniAppAppID:     env("WECHAT_MINIAPP_APP_ID", ""),
		WechatMiniAppAppSecret: env("WECHAT_MINIAPP_APP_SECRET", ""),
		WechatAPIBaseURL:       env("WECHAT_API_BASE_URL", "https://api.weixin.qq.com"),
		MiniAppAllowMockLogin:  envBool("MINIAPP_ALLOW_MOCK_LOGIN", strings.EqualFold(mode, "memory")),
		StorageEnabled:         storageEnabled,
		StorageEnabledSet:      storageEnabledSet,
		StorageProvider:        env("STORAGE_PROVIDER", ""),
		StorageBucket:          env("STORAGE_BUCKET", ""),
		StorageRegion:          env("STORAGE_REGION", ""),
		StorageEndpoint:        env("STORAGE_ENDPOINT", ""),
		StoragePublicBaseURL:   env("STORAGE_PUBLIC_BASE_URL", ""),
		StoragePathPrefix:      env("STORAGE_PATH_PREFIX", ""),
		StorageUploadStrategy:  env("STORAGE_UPLOAD_STRATEGY", ""),
		StorageAccessKeyID:     env("STORAGE_ACCESS_KEY_ID", ""),
		StorageAccessKeySecret: env("STORAGE_ACCESS_KEY_SECRET", ""),
		StorageUploadURLTTL:    envInt("STORAGE_UPLOAD_URL_TTL_SECONDS", 1800),
		StorageCallbackEnabled: storageCallbackEnabled,
		StorageCallbackSet:     storageCallbackSet,
		StorageStatus:          env("STORAGE_STATUS_DESCRIPTION", ""),
		GoldPriceLiveEnabled:   envBool("GOLD_PRICE_LIVE_ENABLED", true),
		GoldPriceAPIURL:        env("GOLD_PRICE_API_URL", "https://api.gold-api.com/price/XAU"),
		GoldFXAPIURL:           env("GOLD_PRICE_FX_API_URL", "https://api.frankfurter.dev/v1/latest?base=USD&symbols=CNY"),
		GoldPriceCacheTTL:      envInt("GOLD_PRICE_CACHE_TTL_SECONDS", 60),
		GoldPriceHTTPTimeout:   envInt("GOLD_PRICE_HTTP_TIMEOUT_SECONDS", 5),
		GoldPriceCNYPerGram:    envFloat("GOLD_PRICE_CNY_PER_GRAM", 0),
		GoldPriceXAUUSD:        envFloat("GOLD_PRICE_XAU_USD", 0),
		GoldPriceUSDCNY:        envFloat("GOLD_PRICE_USD_CNY", 0),
		GoldPriceUpdatedAt:     env("GOLD_PRICE_UPDATED_AT", ""),
		AssetsDir:              env("ASSETS_DIR", ""),
		AssetsPublicBaseURL:    env("ASSETS_PUBLIC_BASE_URL", ""),
	}

	if err := validatePersistentRuntimeConfig(cfg); err != nil {
		return nil, err
	}

	persistence, err := newPersistence(cfg)
	if err != nil {
		return nil, err
	}

	store, err := newMockStore(persistence)
	if err != nil {
		if persistence != nil {
			_ = persistence.Close()
		}
		return nil, err
	}
	if err := store.applyRuntimeConfig(cfg); err != nil {
		_ = store.close()
		return nil, err
	}

	return &App{
		Config: cfg,
		store:  store,
	}, nil
}

func (a *App) Close() error {
	return a.store.close()
}

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/api/v1/auth/login", a.handleLogin)
	mux.HandleFunc("/api/admin/login", a.handleAdminLogin)
	mux.Handle("/api/admin/bootstrap", a.withAuth(a.requireAdminAbility("dashboard.view", a.handleAdminBootstrap)))
	mux.Handle("/api/admin/print-template", a.withAuth(a.requireAdminAbility("system.config.manage", a.handleAdminPrintTemplate)))
	mux.Handle("/api/admin/stores", a.withAuth(a.requireAdminAbility("store.manage", a.handleAdminStoreCollection)))
	mux.Handle("/api/admin/roles", a.withAuth(a.requireAdminAbility("role.manage", a.handleAdminRoleCollection)))
	mux.Handle("/api/admin/roles/", a.withAuth(a.requireAdminAbility("role.manage", a.handleAdminRoleTemplate)))
	mux.Handle("/api/admin/stores/", a.withAuth(a.requireAdminAbility("store.manage", a.handleAdminStore)))
	mux.Handle("/api/admin/users", a.withAuth(a.requireAdminAbility("user.manage", a.handleAdminUserCollection)))
	mux.Handle("/api/admin/users/", a.withAuth(a.requireAdminAbility("user.manage", a.handleAdminUser)))
	mux.Handle("/api/admin/miniapp/pending-bindings", a.withAuth(a.requireAdminAbility("user.manage", a.handleAdminMiniAppPendingBindings)))
	mux.Handle("/api/admin/products/import-template", a.withAuth(a.requireAdminAbility("product.manage", a.handleAdminProductImportTemplate)))
	mux.Handle("/api/admin/products/import", a.withAuth(a.requireAdminAbility("product.manage", a.handleAdminProductImport)))
	mux.Handle("/api/admin/products", a.withAuth(a.requireAdminAbility("product.manage", a.handleAdminProductCollection)))
	mux.Handle("/api/admin/products/", a.withAuth(a.requireAdminAbility("product.manage", a.handleAdminProduct)))
	mux.Handle("/api/admin/inventory/items", a.withAuth(a.requireAdminAbility("product.manage", a.handleAdminInventoryItems)))
	mux.Handle("/api/admin/members", a.withAuth(a.requireAdminAbility("user.manage", a.handleAdminMemberCollection)))
	mux.Handle("/api/admin/members/", a.withAuth(a.requireAdminAbility("user.manage", a.handleAdminMember)))
	mux.Handle("/api/admin/cashier-orders", a.withAuth(a.requireAdminAbility("order.view", a.handleAdminCashierOrderCollection)))
	mux.Handle("/api/admin/cashier-orders/", a.withAuth(a.requireAdminAbility("order.view", a.handleAdminCashierOrder)))
	mux.Handle("/api/admin/recycle-orders", a.withAuth(a.requireAdminAbility("recycle.view", a.handleAdminRecycleOrderCollection)))
	mux.Handle("/api/admin/recycle-orders/", a.withAuth(a.requireAdminAbility("recycle.view", a.handleAdminRecycleOrder)))
	mux.Handle("/api/admin/materials", a.withAuth(a.requireAdminAbility("recycle.view", a.handleAdminMaterialItems)))
	mux.Handle("/api/admin/materials/", a.withAuth(a.requireAdminAbility("recycle.view", a.handleAdminMaterialActions)))
	mux.Handle("/api/admin/orders", a.withAuth(a.requireAdminAbility("order.view", a.handleAdminOrderCollection)))
	mux.Handle("/api/admin/system-profile", a.withAuth(a.requireAdminAbility("system.config.manage", a.handleAdminSystemProfile)))

	// 后台「顾客端内容管理」接口（决策：统一沿用 /api/admin/ 前缀）
	mux.Handle("/api/admin/uploads/images", a.withAuth(a.requireAdminAbility("customer_content.view", a.handleAdminUploadImages)))
	mux.Handle("/api/admin/uploads/images/", a.withAuth(a.requireAdminAbility("customer_content.view", a.handleAdminUploadImageActions)))
	mux.Handle("/api/admin/customer-home/banners", a.withAuth(a.requireAdminAbility("customer_content.view", a.handleAdminBannerCollection)))
	mux.Handle("/api/admin/customer-home/banners/", a.withAuth(a.requireAdminAbility("customer_content.view", a.handleAdminBannerItem)))
	mux.Handle("/api/admin/customer-home/config", a.withAuth(a.requireAdminAbility("customer_content.view", a.handleAdminHomeConfig)))
	mux.Handle("/api/admin/customer/recycle-info", a.withAuth(a.requireAdminAbility("customer_content.view", a.handleAdminCustomerRecycleInfo)))
	mux.Handle("/api/admin/customer-products", a.withAuth(a.requireAdminAbility("product.view", a.handleAdminCustomerProductCollection)))
	mux.Handle("/api/admin/customer-products/", a.withAuth(a.requireAdminAbility("product.view", a.handleAdminCustomerProductItem)))
	mux.Handle("/api/admin/appointment-rules", a.withAuth(a.requireAdminAbility("customer_content.view", a.handleAdminAppointmentRules)))
	mux.Handle("/api/admin/customer-appointments", a.withAuth(a.requireAdminAbility("appointment.manage", a.handleAdminCustomerAppointments)))
	mux.Handle("/api/admin/customer-appointments/", a.withAuth(a.requireAdminAbility("appointment.manage", a.handleAdminCustomerAppointmentActions)))

	// 素材静态服务（本地磁盘兜底方案；配置 ASSETS_DIR 后启用）
	if assetsDir := strings.TrimSpace(a.Config.AssetsDir); assetsDir != "" {
		mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))
	}
	mux.Handle("/api/v1/auth/logout", a.withAuth(a.handleLogout))
	mux.Handle("/api/v1/me", a.withAuth(a.handleMe))
	mux.HandleFunc("/api/v1/auth/wechat-login", a.handleMiniAppLogin)
	mux.Handle("/api/v1/stores", a.withAuth(a.requirePermission("store.read", a.handleStores)))
	mux.Handle("/api/v1/roles", a.withAuth(a.requirePermission("role.read", a.handleRoles)))
	mux.Handle("/api/v1/permissions/catalog", a.withAuth(a.requirePermission("permission.catalog.read", a.handlePermissionCatalog)))
	mux.Handle("/api/v1/dashboard/summary", a.withAuth(a.requirePermission("dashboard.summary.view", a.handleDashboardSummary)))
	mux.Handle("/api/v1/reports/daily", a.withAuth(a.requirePermission("dashboard.summary.view", a.handleDailyReport)))
	mux.Handle("/api/v1/members", a.withAuth(a.requirePermission("member.read", a.handleMembers)))
	mux.Handle("/api/v1/members/", a.withAuth(a.requirePermission("member.read", a.handleMemberDetail)))
	mux.Handle("/api/v1/products", a.withAuth(a.requirePermission("catalog.product.read", a.handleCatalogProducts)))
	mux.Handle("/api/v1/products/", a.withAuth(a.requirePermission("catalog.product.read", a.handleCatalogProductDetail)))
	mux.Handle("/api/v1/gold-prices/reference", a.withAuth(a.handleGoldReferencePrices))
	mux.Handle("/api/v1/inventory/summary", a.withAuth(a.requirePermission("catalog.product.read", a.handleInventorySummary)))
	mux.Handle("/api/v1/inventory/items", a.withAuth(a.requirePermission("catalog.product.read", a.handleInventoryItems)))
	mux.Handle("/api/v1/cashier/orders", a.withAuth(a.handleCashierOrders))
	mux.Handle("/api/v1/cashier/orders/", a.withAuth(a.handleCashierOrderDetail))
	mux.Handle("/api/v1/uploads/recycle-photos/prepare", a.withAuth(a.requirePermission("recycle.order.draft", a.handleRecyclePhotoUploadPrepare)))
	mux.Handle("/api/v1/uploads/recycle-photos/complete", a.withAuth(a.requirePermission("recycle.order.draft", a.handleRecyclePhotoUploadComplete)))
	mux.Handle("/api/v1/recycle/quote-preview", a.withAuth(a.requirePermission("recycle.order.draft", a.handleRecycleQuotePreview)))
	mux.Handle("/api/v1/recycle/orders", a.withAuth(a.handleRecycleOrders))
	mux.Handle("/api/v1/recycle/orders/", a.withAuth(a.handleRecycleOrderActions))
	mux.Handle("/api/v1/materials", a.withAuth(a.requirePermission("recycle.order.read", a.handleMaterialItems)))
	mux.Handle("/api/v1/materials/", a.withAuth(a.requirePermission("recycle.order.read", a.handleMaterialActions)))
	mux.Handle("/api/v1/settings", a.withAuth(a.handleSettings))

	// 顾客端公开接口（游客可访问）
	mux.HandleFunc("/api/v1/customer/home", a.handleCustomerHome)
	mux.HandleFunc("/api/v1/customer/products", a.handleCustomerProducts)
	mux.HandleFunc("/api/v1/customer/products/", a.handleCustomerProductDetail)
	mux.HandleFunc("/api/v1/customer/stores", a.handleCustomerStores)
	mux.HandleFunc("/api/v1/customer/stores/", a.handleCustomerStoreActions)
	mux.HandleFunc("/api/v1/customer/recycle-info", a.handleCustomerRecycleInfo)

	// 顾客认证接口
	mux.HandleFunc("/api/v1/customer/auth/wechat-login", a.handleCustomerWechatLogin)
	mux.Handle("/api/v1/customer/auth/phone", a.withCustomerAuth(a.handleCustomerPhoneAuth))
	mux.Handle("/api/v1/customer/auth/logout", a.withCustomerAuth(a.handleCustomerLogout))
	mux.Handle("/api/v1/customer/me", a.withCustomerAuth(a.handleCustomerProfile))

	// 顾客预约接口（需顾客登录）
	mux.Handle("/api/v1/customer/appointments", a.withCustomerAuth(a.handleCustomerAppointments))
	mux.Handle("/api/v1/customer/appointments/", a.withCustomerAuth(a.handleCustomerAppointmentActions))

	// 员工预约管理接口（需员工登录）
	mux.Handle("/api/v1/staff/appointments", a.withAuth(a.handleStaffAppointments))
	mux.Handle("/api/v1/staff/appointments/", a.withAuth(a.handleStaffAppointmentActions))

	mux.HandleFunc("/", a.handleNotFound)

	return a.withCORS(a.withRequestContext(mux))
}

func (a *App) withRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := a.newRequestID()
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin != "" {
			if !a.corsOriginAllowed(origin) {
				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				a.writeError(w, r, http.StatusForbidden, 40303, "origin is not allowed")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) corsOriginAllowed(origin string) bool {
	for _, configured := range strings.Split(a.Config.CORSOrigin, ",") {
		configured = strings.TrimSpace(configured)
		if configured == "*" || configured == origin {
			return true
		}
	}
	return false
}

func validatePersistentRuntimeConfig(cfg Config) error {
	if !strings.EqualFold(strings.TrimSpace(cfg.Mode), "persistent") {
		return nil
	}
	if strings.TrimSpace(cfg.TokenSecret) == "" || cfg.TokenSecret == "gold-recycle-dev-secret" {
		return errors.New("persistent mode requires a non-default TOKEN_SECRET")
	}
	if strings.TrimSpace(cfg.CORSOrigin) == "" || strings.Contains(cfg.CORSOrigin, "*") {
		return errors.New("persistent mode requires explicit CORS_ORIGIN origins")
	}
	if cfg.MiniAppAllowMockLogin {
		return errors.New("persistent mode cannot enable MINIAPP_ALLOW_MOCK_LOGIN")
	}
	if !cfg.hasMiniAppWechatAuth() {
		return errors.New("persistent mode requires WECHAT_MINIAPP_APP_ID and WECHAT_MINIAPP_APP_SECRET")
	}
	return nil
}

func (a *App) withAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" || token == r.Header.Get("Authorization") {
			a.writeError(w, r, http.StatusUnauthorized, 40101, "missing bearer token")
			return
		}

		user, session, ok := a.store.getUserByToken(token)
		if !ok {
			a.writeError(w, r, http.StatusUnauthorized, 40102, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userKey, user)
		ctx = context.WithValue(ctx, sessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := currentUser(r.Context())
		for _, permission := range user.Permissions {
			if permission == code {
				next.ServeHTTP(w, r)
				return
			}
		}
		a.writeError(w, r, http.StatusForbidden, 40301, fmt.Sprintf("missing permission: %s", code))
	}
}

func (a *App) requireAdminAbility(code AdminAbilityCode, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := currentUser(r.Context())
		sessionUser := a.store.adminSessionUser(user)
		for _, ability := range sessionUser.Abilities {
			if ability == code {
				next.ServeHTTP(w, r)
				return
			}
		}
		a.writeError(w, r, http.StatusForbidden, 40311, fmt.Sprintf("missing admin ability: %s", code))
	}
}

func currentUser(ctx context.Context) UserAccount {
	user, _ := ctx.Value(userKey).(UserAccount)
	return user
}

func currentSession(ctx context.Context) Session {
	session, _ := ctx.Value(sessionKey).(Session)
	return session
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}

func (a *App) newRequestID() string {
	seed := atomic.AddUint64(&a.requestSeed, 1)
	return fmt.Sprintf("req_%s_%04d", time.Now().Format("20060102150405"), seed)
}

func (a *App) writeJSON(w http.ResponseWriter, r *http.Request, status, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: requestIDFromContext(r.Context()),
	})
}

func (a *App) writeError(w http.ResponseWriter, r *http.Request, status, code int, message string) {
	a.writeJSON(w, r, status, code, message, nil)
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBoolOptional(key string) (bool, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
}

func mysqlDSNFromEnv() string {
	if dsn := env("MYSQL_DSN", ""); dsn != "" {
		return dsn
	}

	host := env("MYSQL_HOST", "")
	database := env("MYSQL_DATABASE", "")
	user := env("MYSQL_USER", "")
	if host == "" || database == "" || user == "" {
		return ""
	}

	password := env("MYSQL_PASSWORD", "")
	port := env("MYSQL_PORT", "3306")
	charset := env("MYSQL_CHARSET", "utf8mb4")
	loc := url.QueryEscape("Local")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=%s&loc=%s",
		user,
		password,
		host,
		port,
		database,
		charset,
		loc,
	)
}
