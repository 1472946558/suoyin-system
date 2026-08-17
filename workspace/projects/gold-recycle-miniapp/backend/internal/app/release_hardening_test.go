package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSRejectsUntrustedOriginsAndEchoesAllowedOrigin(t *testing.T) {
	app := newTestAppWithEnv(t, map[string]string{
		"CORS_ORIGIN": "https://admin.example.com,https://console.example.com",
	})
	handler := app.Router()

	blocked := httptest.NewRecorder()
	blockedReq := httptest.NewRequest(http.MethodOptions, "/api/v1/customer/home", nil)
	blockedReq.Header.Set("Origin", "https://evil.example")
	blockedReq.Header.Set("Access-Control-Request-Method", http.MethodGet)
	handler.ServeHTTP(blocked, blockedReq)
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("blocked CORS status = %d, want %d", blocked.Code, http.StatusForbidden)
	}
	if got := blocked.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("blocked CORS response must not allow origin, got %q", got)
	}

	allowed := httptest.NewRecorder()
	allowedReq := httptest.NewRequest(http.MethodOptions, "/api/v1/customer/home", nil)
	allowedReq.Header.Set("Origin", "https://console.example.com")
	allowedReq.Header.Set("Access-Control-Request-Method", http.MethodGet)
	handler.ServeHTTP(allowed, allowedReq)
	if allowed.Code != http.StatusNoContent {
		t.Fatalf("allowed CORS status = %d, want %d", allowed.Code, http.StatusNoContent)
	}
	if got := allowed.Header().Get("Access-Control-Allow-Origin"); got != "https://console.example.com" {
		t.Fatalf("allowed CORS origin = %q", got)
	}
	if !strings.Contains(allowed.Header().Get("Vary"), "Origin") {
		t.Fatalf("allowed CORS response must vary by Origin")
	}
}

func TestValidatePersistentRuntimeConfig(t *testing.T) {
	valid := Config{
		Mode:                   "persistent",
		TokenSecret:            "a-long-random-production-secret",
		CORSOrigin:             "https://jinjiangguan.com",
		MiniAppAllowMockLogin:  false,
		WechatMiniAppAppID:     "wx-production",
		WechatMiniAppAppSecret: "production-secret",
	}
	if err := validatePersistentRuntimeConfig(valid); err != nil {
		t.Fatalf("valid persistent config rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "default token secret", mutate: func(cfg *Config) { cfg.TokenSecret = "gold-recycle-dev-secret" }},
		{name: "wildcard cors", mutate: func(cfg *Config) { cfg.CORSOrigin = "*" }},
		{name: "mock login", mutate: func(cfg *Config) { cfg.MiniAppAllowMockLogin = true }},
		{name: "missing wechat secret", mutate: func(cfg *Config) { cfg.WechatMiniAppAppSecret = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := valid
			tc.mutate(&cfg)
			if err := validatePersistentRuntimeConfig(cfg); err == nil {
				t.Fatalf("expected persistent config to be rejected")
			}
		})
	}
}
