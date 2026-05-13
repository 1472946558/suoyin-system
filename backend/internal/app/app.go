package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	Config      Config
	store       *MockStore
	requestSeed uint64
}

func New() (*App, error) {
	cfg := Config{
		Host:        env("HOST", ""),
		Port:        env("PORT", "8080"),
		Mode:        env("APP_MODE", "memory"),
		TokenSecret: env("TOKEN_SECRET", "gold-recycle-dev-secret"),
		CORSOrigin:  env("CORS_ORIGIN", "*"),
		MySQLDSN:    mysqlDSNFromEnv(),
		RedisAddr:   env("REDIS_ADDR", ""),
		RedisUser:   env("REDIS_USER", ""),
		RedisPass:   env("REDIS_PASSWORD", ""),
		RedisDB:     envInt("REDIS_DB", 0),
		RedisPrefix: env("REDIS_KEY_PREFIX", "gold:"),
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
	mux.HandleFunc("/api/payment/wechat/callback", a.handleWechatPaymentCallback)
	mux.HandleFunc("/api/v1/auth/login", a.handleLogin)
	mux.HandleFunc("/api/admin/login", a.handleAdminLogin)
	mux.Handle("/api/admin/bootstrap", a.withAuth(a.requireAdminAbility("dashboard.view", a.handleAdminBootstrap)))
	mux.Handle("/api/admin/payment-config", a.withAuth(a.requireAdminAbility("payment.config.manage", a.handleAdminPaymentConfig)))
	mux.Handle("/api/admin/print-template", a.withAuth(a.requireAdminAbility("system.config.manage", a.handleAdminPrintTemplate)))
	mux.Handle("/api/admin/stores", a.withAuth(a.requireAdminAbility("store.manage", a.handleAdminStoreCreate)))
	mux.Handle("/api/admin/roles/", a.withAuth(a.requireAdminAbility("role.manage", a.handleAdminRoleTemplate)))
	mux.Handle("/api/admin/stores/", a.withAuth(a.requireAdminAbility("store.manage", a.handleAdminStore)))
	mux.Handle("/api/admin/users", a.withAuth(a.requireAdminAbility("user.manage", a.handleAdminUserCreate)))
	mux.Handle("/api/admin/users/", a.withAuth(a.requireAdminAbility("user.manage", a.handleAdminUser)))
	mux.Handle("/api/admin/products", a.withAuth(a.requireAdminAbility("product.manage", a.handleAdminProductCreate)))
	mux.Handle("/api/admin/products/", a.withAuth(a.requireAdminAbility("product.manage", a.handleAdminProduct)))
	mux.Handle("/api/admin/system-profile", a.withAuth(a.requireAdminAbility("system.config.manage", a.handleAdminSystemProfile)))
	mux.Handle("/api/v1/auth/logout", a.withAuth(a.handleLogout))
	mux.Handle("/api/v1/me", a.withAuth(a.handleMe))
	mux.HandleFunc("/api/v1/auth/wechat-login", a.handleMiniAppLogin)
	mux.Handle("/api/v1/stores", a.withAuth(a.requirePermission("store.read", a.handleStores)))
	mux.Handle("/api/v1/roles", a.withAuth(a.requirePermission("role.read", a.handleRoles)))
	mux.Handle("/api/v1/permissions/catalog", a.withAuth(a.requirePermission("permission.catalog.read", a.handlePermissionCatalog)))
	mux.Handle("/api/v1/dashboard/summary", a.withAuth(a.requirePermission("dashboard.summary.view", a.handleDashboardSummary)))
	mux.Handle("/api/v1/members", a.withAuth(a.requirePermission("member.read", a.handleMembers)))
	mux.Handle("/api/v1/members/", a.withAuth(a.requirePermission("member.read", a.handleMemberDetail)))
	mux.Handle("/api/v1/products", a.withAuth(a.requirePermission("catalog.product.read", a.handleCatalogProducts)))
	mux.Handle("/api/v1/products/", a.withAuth(a.requirePermission("catalog.product.read", a.handleCatalogProductDetail)))
	mux.Handle("/api/v1/cashier/orders", a.withAuth(a.handleCashierOrders))
	mux.Handle("/api/v1/cashier/orders/", a.withAuth(a.handleCashierOrderDetail))
	mux.Handle("/api/v1/uploads/recycle-photos/prepare", a.withAuth(a.requirePermission("recycle.order.draft", a.handleRecyclePhotoUploadPrepare)))
	mux.Handle("/api/v1/uploads/recycle-photos/complete", a.withAuth(a.requirePermission("recycle.order.draft", a.handleRecyclePhotoUploadComplete)))
	mux.Handle("/api/v1/recycle/orders", a.withAuth(a.handleRecycleOrders))
	mux.Handle("/api/v1/recycle/orders/", a.withAuth(a.handleRecycleOrderActions))
	mux.Handle("/api/v1/settings", a.withAuth(a.handleSettings))
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
		w.Header().Set("Access-Control-Allow-Origin", a.Config.CORSOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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
