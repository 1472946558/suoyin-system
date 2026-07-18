/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: app_test.go
 * 功能描述: 测试代码
 * 作者: 廖心慈
 * 创建日期: 2026-06-05
 */

package app

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

type apiEnvelope struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	RequestID string          `json:"requestId"`
}

type listResponse[T any] struct {
	Items     []T    `json:"items"`
	DataScope string `json:"dataScope"`
}

type miniappLoginResult struct {
	Token     string   `json:"token"`
	UserID    string   `json:"userId"`
	ID        string   `json:"id"`
	RoleKey   string   `json:"roleKey"`
	RoleName  string   `json:"roleName"`
	StoreID   string   `json:"storeId"`
	StoreName string   `json:"storeName"`
	StoreCode string   `json:"storeCode"`
	Perms     []string `json:"permissions"`
}

type passwordLoginResult struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		RoleCode string `json:"roleCode"`
	} `json:"user"`
}

func encryptedWechatPhonePayloadForTest(t *testing.T, sessionKey, iv, phone string) (string, string) {
	t.Helper()

	keyBytes := []byte(sessionKey)
	ivBytes := []byte(iv)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	plainBytes, err := json.Marshal(map[string]any{
		"phoneNumber":     "86" + phone,
		"purePhoneNumber": phone,
		"countryCode":     "86",
	})
	if err != nil {
		t.Fatalf("marshal phone payload: %v", err)
	}
	padding := block.BlockSize() - len(plainBytes)%block.BlockSize()
	for i := 0; i < padding; i++ {
		plainBytes = append(plainBytes, byte(padding))
	}
	encryptedBytes := make([]byte, len(plainBytes))
	cipher.NewCBCEncrypter(block, ivBytes).CryptBlocks(encryptedBytes, plainBytes)
	return base64.StdEncoding.EncodeToString(encryptedBytes), base64.StdEncoding.EncodeToString(ivBytes)
}

type recycleQuotePreviewTestResult struct {
	NetWeightGram   float64 `json:"netWeightGram"`
	UnitPrice       float64 `json:"unitPrice"`
	EstimatedAmount float64 `json:"estimatedAmount"`
	PriceSource     string  `json:"priceSource"`
}

type goldReferencePricesTestResult struct {
	Source          string                   `json:"source"`
	BaseCNYPerGram  float64                  `json:"baseCnyPerGram"`
	XAUUSD          float64                  `json:"xauUsd"`
	USDCNY          float64                  `json:"usdCny"`
	ReferencePrices []GoldReferencePriceItem `json:"referencePrices"`
}

type inventorySummaryTestResult struct {
	TotalSKU             int             `json:"totalSku"`
	LowStockSKU          int             `json:"lowStockSku"`
	EstimatedRetailValue float64         `json:"estimatedRetailValue"`
	Items                []InventoryItem `json:"items"`
}

func TestSaveCashierOrderStatementColumnPlaceholderCount(t *testing.T) {
	insertStart := strings.Index(saveCashierOrderStatement, "INSERT INTO cashier_orders (")
	valuesStart := strings.Index(saveCashierOrderStatement, ") VALUES (")
	updateStart := strings.Index(saveCashierOrderStatement, "ON DUPLICATE KEY UPDATE")
	if insertStart < 0 || valuesStart < 0 || updateStart < 0 {
		t.Fatalf("unexpected cashier insert statement shape: %s", saveCashierOrderStatement)
	}

	columnSegment := saveCashierOrderStatement[insertStart+len("INSERT INTO cashier_orders (") : valuesStart]
	valueSegment := saveCashierOrderStatement[valuesStart+len(") VALUES (") : updateStart]
	columnCount := 0
	for _, part := range strings.Split(columnSegment, ",") {
		if strings.TrimSpace(part) != "" {
			columnCount++
		}
	}
	placeholderCount := strings.Count(valueSegment, "?")
	if columnCount != placeholderCount {
		t.Fatalf("cashier insert columns/placeholders mismatch: columns=%d placeholders=%d", columnCount, placeholderCount)
	}
}

type dailyReportSummaryTestResult struct {
	Date              string                   `json:"date"`
	VisibleStoreCount int                      `json:"visibleStoreCount"`
	CashierOrderCount int                      `json:"cashierOrderCount"`
	CashierAmount     float64                  `json:"cashierAmount"`
	StoreMetrics      []DailyReportStoreMetric `json:"storeMetrics"`
}

func newTestApp(t *testing.T) *App {
	return newTestAppWithEnv(t, nil)
}

func newTestAppWithEnv(t *testing.T, overrides map[string]string) *App {
	t.Helper()
	defaults := map[string]string{
		"APP_MODE":                        "memory",
		"PORT":                            "8080",
		"TOKEN_SECRET":                    "test-secret",
		"CORS_ORIGIN":                     "*",
		"WECHAT_MINIAPP_APP_ID":           "",
		"WECHAT_MINIAPP_APP_SECRET":       "",
		"WECHAT_API_BASE_URL":             "",
		"MINIAPP_ALLOW_MOCK_LOGIN":        "true",
		"GOLD_PRICE_LIVE_ENABLED":         "false",
		"GOLD_PRICE_API_URL":              "",
		"GOLD_PRICE_FX_API_URL":           "",
		"GOLD_PRICE_CACHE_TTL_SECONDS":    "",
		"GOLD_PRICE_HTTP_TIMEOUT_SECONDS": "",
		"GOLD_PRICE_CNY_PER_GRAM":         "",
		"GOLD_PRICE_XAU_USD":              "",
		"GOLD_PRICE_USD_CNY":              "",
		"GOLD_PRICE_UPDATED_AT":           "",
	}
	for key, value := range overrides {
		defaults[key] = value
	}
	for key, value := range defaults {
		t.Setenv(key, value)
	}

	app, err := New()
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})
	return app
}

func performRequest(t *testing.T, handler http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func decodeResponse[T any](t *testing.T, rr *httptest.ResponseRecorder, wantStatus int) T {
	t.Helper()

	if rr.Code != wantStatus {
		t.Fatalf("unexpected status: got %d want %d body=%s", rr.Code, wantStatus, rr.Body.String())
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	if envelope.Code != 0 {
		t.Fatalf("unexpected response code: got %d body=%s", envelope.Code, rr.Body.String())
	}

	var data T
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return data
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v body=%s", err, rr.Body.String())
	}
	return data
}

func decodeEnvelope(t *testing.T, rr *httptest.ResponseRecorder, wantStatus int) apiEnvelope {
	t.Helper()

	if rr.Code != wantStatus {
		t.Fatalf("unexpected status: got %d want %d body=%s", rr.Code, wantStatus, rr.Body.String())
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, rr.Body.String())
	}
	return envelope
}

func decodeError(t *testing.T, rr *httptest.ResponseRecorder, wantStatus, wantCode int) apiEnvelope {
	t.Helper()

	envelope := decodeEnvelope(t, rr, wantStatus)
	if envelope.Code != wantCode {
		t.Fatalf("unexpected response code: got %d want %d body=%s", envelope.Code, wantCode, rr.Body.String())
	}
	return envelope
}

func adminLogin(t *testing.T, handler http.Handler, username, password string) AdminLoginResult {
	t.Helper()
	rr := performRequest(t, handler, http.MethodPost, "/api/admin/login", "", map[string]string{
		"username": username,
		"password": password,
	})
	return decodeResponse[AdminLoginResult](t, rr, http.StatusOK)
}

func passwordLogin(t *testing.T, handler http.Handler, username, password string) passwordLoginResult {
	t.Helper()
	rr := performRequest(t, handler, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"username": username,
		"password": password,
	})
	return decodeResponse[passwordLoginResult](t, rr, http.StatusOK)
}

func miniappLogin(t *testing.T, handler http.Handler, roleKey string) miniappLoginResult {
	t.Helper()
	body := map[string]any{
		"code": "mock-wechat-code",
		"profile": map[string]any{
			"name":      "测试店员",
			"roleKey":   roleKey,
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}
	rr := performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", body)
	return decodeResponse[miniappLoginResult](t, rr, http.StatusOK)
}

func TestDefaultBossCredentialRestoresLoginAndOwnerCanUpdateSelf(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	app.store.mu.Lock()
	boss := app.store.usersByUsername[defaultBossUsername]
	boss.Password = "lost-password"
	boss.Status = "disabled"
	app.store.usersByID[boss.ID] = boss
	app.store.usersByUsername[boss.Username] = boss
	app.store.mu.Unlock()

	admin := adminLogin(t, handler, defaultBossUsername, defaultBossPassword)
	if admin.User.RoleKey != "boss" || admin.User.Account != defaultBossUsername {
		t.Fatalf("default boss login did not restore owner account: %#v", admin.User)
	}

	mini := passwordLogin(t, handler, defaultBossUsername, defaultBossPassword)
	if mini.User.RoleCode != "boss" || mini.Token == "" {
		t.Fatalf("miniapp password login did not use restored boss: %#v", mini)
	}

	updated := decodeResponse[AdminUserAccount](t, performRequest(t, handler, http.MethodPut, "/api/admin/users/"+admin.User.ID, admin.Token, map[string]any{
		"id":       admin.User.ID,
		"name":     admin.User.Name,
		"account":  "boss.main",
		"password": "BossMain456!",
		"roleKey":  "boss",
		"status":   "enabled",
	}), http.StatusOK)
	if updated.Account != "boss.main" || updated.RoleKey != "boss" {
		t.Fatalf("owner self account update not applied: %#v", updated)
	}

	updatedAdmin := adminLogin(t, handler, "boss.main", "BossMain456!")
	if updatedAdmin.User.ID != admin.User.ID || updatedAdmin.User.RoleKey != "boss" {
		t.Fatalf("updated owner credentials cannot log in admin: %#v", updatedAdmin.User)
	}

	updatedMini := passwordLogin(t, handler, "boss.main", "BossMain456!")
	if updatedMini.User.RoleCode != "boss" || updatedMini.Token == "" {
		t.Fatalf("updated owner credentials cannot log in miniapp: %#v", updatedMini)
	}
}

func TestDefaultManagerCredentialRestoresSingleStoreLogin(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	app.store.mu.Lock()
	manager := app.store.usersByUsername[defaultManagerUsername]
	manager.Password = "lost-password"
	manager.Status = "disabled"
	app.store.usersByID[manager.ID] = manager
	app.store.usersByUsername[manager.Username] = manager
	app.store.mu.Unlock()

	admin := adminLogin(t, handler, defaultManagerUsername, defaultManagerPassword)
	if admin.User.RoleKey != "shop_manager" || admin.User.Account != defaultManagerUsername {
		t.Fatalf("default manager login did not restore shop manager account: %#v", admin.User)
	}

	mini := passwordLogin(t, handler, defaultManagerUsername, defaultManagerPassword)
	if mini.User.RoleCode != "shop_manager" || mini.Token == "" {
		t.Fatalf("miniapp password login did not use restored shop manager: %#v", mini)
	}

	stores := decodeResponse[listResponse[StoreInfo]](t, performRequest(t, handler, http.MethodGet, "/api/v1/stores", mini.Token, nil), http.StatusOK)
	if len(stores.Items) != 1 {
		t.Fatalf("default manager should see exactly one store, got %#v", stores.Items)
	}
}

func TestMiniappMainFlowInProcess(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	health := decodeResponse[map[string]any](t, performRequest(t, handler, http.MethodGet, "/health", "", nil), http.StatusOK)
	if health["service"] != "gold-recycle-miniapp-backend" {
		t.Fatalf("unexpected health service: %#v", health)
	}

	admin := adminLogin(t, handler, "boss", "Boss123!")
	bootstrap := decodeResponse[AdminBootstrap](t, performRequest(t, handler, http.MethodGet, "/api/admin/bootstrap", admin.Token, nil), http.StatusOK)
	if len(bootstrap.Stores) == 0 || len(bootstrap.Roles) == 0 {
		t.Fatalf("bootstrap missing stores or roles: %#v", bootstrap)
	}

	mini := miniappLogin(t, handler, "manager")

	members := decodeResponse[listResponse[MemberProfile]](t, performRequest(t, handler, http.MethodGet, "/api/v1/members", mini.Token, nil), http.StatusOK)
	if len(members.Items) == 0 {
		t.Fatal("expected seeded members for miniapp flow")
	}

	products := decodeResponse[listResponse[CatalogProduct]](t, performRequest(t, handler, http.MethodGet, "/api/v1/products", mini.Token, nil), http.StatusOK)
	if len(products.Items) == 0 {
		t.Fatal("expected seeded products for miniapp flow")
	}

	goldPrices := decodeResponse[goldReferencePricesTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/gold-prices/reference", mini.Token, nil), http.StatusOK)
	if goldPrices.Source != "local_reference" || len(goldPrices.ReferencePrices) == 0 {
		t.Fatalf("unexpected fallback gold reference prices: %#v", goldPrices)
	}

	inventory := decodeResponse[inventorySummaryTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/inventory/summary", mini.Token, nil), http.StatusOK)
	if inventory.TotalSKU == 0 || len(inventory.Items) == 0 || inventory.EstimatedRetailValue <= 0 {
		t.Fatalf("unexpected inventory summary: %#v", inventory)
	}

	report := decodeResponse[dailyReportSummaryTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/reports/daily", mini.Token, nil), http.StatusOK)
	if report.Date == "" || report.VisibleStoreCount == 0 || len(report.StoreMetrics) == 0 {
		t.Fatalf("unexpected daily report: %#v", report)
	}

	quote := decodeResponse[recycleQuotePreviewTestResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/recycle/quote-preview", mini.Token, map[string]any{
		"storeId":             mini.StoreID,
		"category":            "金饰",
		"purity":              "足金999",
		"grossWeightGram":     10,
		"deductionWeightGram": 0.2,
	}), http.StatusOK)
	if quote.NetWeightGram != 9.8 || quote.UnitPrice <= 0 || quote.EstimatedAmount <= 0 || quote.PriceSource == "" {
		t.Fatalf("unexpected quote preview: %#v", quote)
	}

	cashier := decodeResponse[CashierOrder](t, performRequest(t, handler, http.MethodPost, "/api/v1/cashier/orders", mini.Token, map[string]any{
		"storeId":       mini.StoreID,
		"customerName":  "收银客户",
		"customerPhone": "13800138001",
		"remark":        "test cashier order",
		"items": []map[string]any{
			{
				"productId": "product-001",
				"sku":       "GJG-SZ-001",
				"name":      "足金手镯标准款",
				"quantity":  1,
				"unitPrice": 1288,
			},
		},
	}), http.StatusCreated)
	if cashier.ID == "" || cashier.OrderNo == "" {
		t.Fatalf("cashier order not created: %#v", cashier)
	}
	if cashier.CustomerName != "收银客户" || cashier.CustomerPhone != "13800138001" {
		t.Fatalf("cashier customer fields not preserved: %#v", cashier)
	}
	if len(cashier.Items) != 1 || cashier.Items[0].SKU != "GJG-SZ-001" || cashier.Items[0].ProductID != "product-001" {
		t.Fatalf("cashier sku fields not preserved: %#v", cashier.Items)
	}

	skuOnlyCashier := decodeResponse[CashierOrder](t, performRequest(t, handler, http.MethodPost, "/api/v1/cashier/orders", mini.Token, map[string]any{
		"storeId":       mini.StoreID,
		"customerName":  "编码客户",
		"customerPhone": "13800138002",
		"remark":        "sku lookup cashier order",
		"items": []map[string]any{
			{
				"sku":      "GJG-KG-018",
				"quantity": 2,
			},
		},
	}), http.StatusCreated)
	if len(skuOnlyCashier.Items) != 1 || skuOnlyCashier.Items[0].Name != "18K 项链" || skuOnlyCashier.Items[0].UnitPrice != 612 || skuOnlyCashier.Items[0].Amount != 1224 {
		t.Fatalf("cashier sku lookup did not populate product fields: %#v", skuOnlyCashier.Items)
	}

	recycleDraft := decodeResponse[RecycleOrder](t, performRequest(t, handler, http.MethodPost, "/api/v1/recycle/orders", mini.Token, map[string]any{
		"storeId":         mini.StoreID,
		"customerName":    "回收客户",
		"customerPhone":   "13800138000",
		"estimatedAmount": 3200,
		"items": []map[string]any{
			{
				"category":   "金饰",
				"purity":     "足金999",
				"weightGram": 5.2,
			},
		},
		"attachmentUrls": []string{
			"https://mock.example.com/test-1.jpg",
			"https://mock.example.com/test-2.jpg",
			"https://mock.example.com/test-3.jpg",
		},
		"remark": "draft",
	}), http.StatusCreated)
	if recycleDraft.ID == "" || recycleDraft.Status != "draft" {
		t.Fatalf("recycle draft not created: %#v", recycleDraft)
	}

	recycleConfirmed := decodeResponse[RecycleOrder](t, performRequest(t, handler, http.MethodPost, "/api/v1/recycle/orders/"+recycleDraft.ID+"/confirm", mini.Token, map[string]any{
		"confirmedAmount": 3180,
		"attachmentUrls":  recycleDraft.AttachmentURLs,
		"remark":          "confirmed",
	}), http.StatusOK)
	if recycleConfirmed.Status != "confirmed" {
		t.Fatalf("recycle order not confirmed: %#v", recycleConfirmed)
	}

	recycleDetail := decodeResponse[RecycleOrder](t, performRequest(t, handler, http.MethodGet, "/api/v1/recycle/orders/"+recycleDraft.ID, mini.Token, nil), http.StatusOK)
	if recycleDetail.ID != recycleDraft.ID || len(recycleDetail.AttachmentURLs) != 3 {
		t.Fatalf("unexpected recycle detail: %#v", recycleDetail)
	}
}

func TestInventoryAndMaterialLedgerRealAPIs(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	boss := passwordLogin(t, handler, defaultBossUsername, defaultBossPassword)
	manager := passwordLogin(t, handler, defaultManagerUsername, defaultManagerPassword)
	if boss.Token == "" || manager.Token == "" {
		t.Fatal("expected boss and manager tokens")
	}

	inventoryItem := decodeResponse[InventoryLedgerItem](t, performRequest(t, handler, http.MethodPost, "/api/v1/inventory/items", manager.Token, map[string]any{
		"styleNo":    "YS-TEST-001",
		"name":       "验收足金手链",
		"category":   "金饰",
		"purity":     "足金999",
		"pieceCount": 2,
		"weightGram": 18.66,
		"costAmount": 12888.5,
		"remark":     "库存真实接口验收",
	}), http.StatusCreated)
	if inventoryItem.ID == "" || inventoryItem.StoreID == "" || inventoryItem.StyleNo != "YS-TEST-001" {
		t.Fatalf("inventory item not created correctly: %#v", inventoryItem)
	}

	managerInventory := decodeResponse[InventoryLedgerSummary](t, performRequest(t, handler, http.MethodGet, "/api/v1/inventory/items", manager.Token, nil), http.StatusOK)
	if managerInventory.TotalPieceCount != 2 || managerInventory.TotalWeightGram != 18.66 || len(managerInventory.Items) != 1 {
		t.Fatalf("manager inventory summary not scoped or aggregated correctly: %#v", managerInventory)
	}

	bossInventory := decodeResponse[InventoryLedgerSummary](t, performRequest(t, handler, http.MethodGet, "/api/admin/inventory/items", boss.Token, nil), http.StatusOK)
	if bossInventory.VisibleStoreCount < 2 || bossInventory.TotalPieceCount != 2 {
		t.Fatalf("boss admin inventory summary should include all stores: %#v", bossInventory)
	}
	updatedInventory := decodeResponse[InventoryLedgerItem](t, performRequest(t, handler, http.MethodPut, "/api/admin/inventory/items/"+inventoryItem.ID, boss.Token, map[string]any{
		"storeId":    inventoryItem.StoreID,
		"styleNo":    inventoryItem.StyleNo,
		"name":       "验收足金手链已编辑",
		"category":   inventoryItem.Category,
		"purity":     inventoryItem.Purity,
		"pieceCount": 3,
		"weightGram": inventoryItem.WeightGram,
		"costAmount": inventoryItem.CostAmount,
		"status":     "in_stock",
		"remark":     "库存编辑验收",
	}), http.StatusOK)
	if updatedInventory.Name != "验收足金手链已编辑" || updatedInventory.PieceCount != 3 {
		t.Fatalf("inventory item not updated correctly: %#v", updatedInventory)
	}
	deletedInventory := decodeResponse[InventoryLedgerItem](t, performRequest(t, handler, http.MethodDelete, "/api/admin/inventory/items/"+inventoryItem.ID, boss.Token, nil), http.StatusOK)
	if deletedInventory.Status != "deleted" {
		t.Fatalf("inventory delete not applied: %#v", deletedInventory)
	}
	afterInventoryDelete := decodeResponse[InventoryLedgerSummary](t, performRequest(t, handler, http.MethodGet, "/api/admin/inventory/items", boss.Token, nil), http.StatusOK)
	if len(afterInventoryDelete.Items) != 0 {
		t.Fatalf("deleted inventory should be hidden from ledger: %#v", afterInventoryDelete.Items)
	}

	materialItem := decodeResponse[MaterialLedgerItem](t, performRequest(t, handler, http.MethodPost, "/api/v1/materials", manager.Token, map[string]any{
		"type":         "pledge",
		"orderNo":      "JL-TEST-001",
		"customerName": "验收客户",
		"category":     "金饰",
		"purity":       "足金999",
		"weightGram":   10.25,
		"amount":       6200,
		"dueDate":      "2026-08-01",
		"remark":       "抵押寄存真实接口验收",
	}), http.StatusCreated)
	if materialItem.ID == "" || materialItem.Type != "pledge" || materialItem.RemainingWeightGram != 10.25 {
		t.Fatalf("material item not created correctly: %#v", materialItem)
	}

	updatedMaterial := decodeResponse[MaterialLedgerItem](t, performRequest(t, handler, http.MethodPut, "/api/admin/materials/"+materialItem.ID, boss.Token, map[string]any{
		"type":                "leftover",
		"orderNo":             "JL-TEST-001-EDIT",
		"customerName":        "编辑客户",
		"category":            "黄金",
		"purity":              "足金9999",
		"weightGram":          11.5,
		"amount":              6600,
		"remainingWeightGram": 9.5,
		"status":              "in_stock",
		"dueDate":             "2026-08-08",
		"remark":              "后台编辑真实接口验收",
		"source":              "admin",
	}), http.StatusOK)
	if updatedMaterial.Type != "leftover" || updatedMaterial.OrderNo != "JL-TEST-001-EDIT" || updatedMaterial.RemainingWeightGram != 9.5 {
		t.Fatalf("material update not applied: %#v", updatedMaterial)
	}

	materials := decodeResponse[MaterialLedgerSummary](t, performRequest(t, handler, http.MethodGet, "/api/v1/materials", manager.Token, nil), http.StatusOK)
	if materials.TodayWeightGram != 11.5 || materials.MonthAmount != 6600 || materials.PledgeCount != 0 || len(materials.PurityStats) != 1 {
		t.Fatalf("material summary not aggregated correctly: %#v", materials)
	}

	outbound := decodeResponse[MaterialLedgerItem](t, performRequest(t, handler, http.MethodPost, "/api/v1/materials/"+materialItem.ID+"/outbound", manager.Token, map[string]any{
		"remark": "验收旧料出库",
	}), http.StatusOK)
	if outbound.Status != "outbound" || outbound.RemainingWeightGram != 0 || outbound.OutboundAt == nil {
		t.Fatalf("material outbound not applied: %#v", outbound)
	}

	afterOutbound := decodeResponse[MaterialLedgerSummary](t, performRequest(t, handler, http.MethodGet, "/api/admin/materials", boss.Token, nil), http.StatusOK)
	if afterOutbound.RemainingWeightGram != 0 || len(afterOutbound.Items) != 1 || afterOutbound.Items[0].Status != "outbound" {
		t.Fatalf("material summary after outbound incorrect: %#v", afterOutbound)
	}

	deleted := decodeResponse[MaterialLedgerItem](t, performRequest(t, handler, http.MethodDelete, "/api/admin/materials/"+materialItem.ID, boss.Token, nil), http.StatusOK)
	if deleted.Status != "deleted" {
		t.Fatalf("material delete not applied: %#v", deleted)
	}
	afterDelete := decodeResponse[MaterialLedgerSummary](t, performRequest(t, handler, http.MethodGet, "/api/admin/materials", boss.Token, nil), http.StatusOK)
	if len(afterDelete.Items) != 0 {
		t.Fatalf("deleted material should be hidden from ledger: %#v", afterDelete.Items)
	}
}

func TestCashierRefundExcludesOrderFromReports(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	mini := miniappLogin(t, handler, "manager")
	admin := passwordLogin(t, handler, defaultBossUsername, defaultBossPassword)
	today := time.Now().Format("2006-01-02")
	before := decodeResponse[dailyReportSummaryTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/reports/daily?date="+today, mini.Token, nil), http.StatusOK)

	cashier := decodeResponse[CashierOrder](t, performRequest(t, handler, http.MethodPost, "/api/v1/cashier/orders", mini.Token, map[string]any{
		"storeId":       mini.StoreID,
		"customerName":  "退单客户",
		"customerPhone": "13800138009",
		"items": []map[string]any{
			{"name": "退单测试商品", "quantity": 1, "unitPrice": 300},
		},
	}), http.StatusCreated)
	afterCreate := decodeResponse[dailyReportSummaryTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/reports/daily?date="+today, mini.Token, nil), http.StatusOK)
	if afterCreate.CashierOrderCount != before.CashierOrderCount+1 || afterCreate.CashierAmount != before.CashierAmount+300 {
		t.Fatalf("cashier order should be counted before refund: before=%#v after=%#v", before, afterCreate)
	}

	refunded := decodeResponse[CashierOrder](t, performRequest(t, handler, http.MethodPost, "/api/v1/cashier/orders/"+cashier.ID+"/refund", mini.Token, map[string]any{
		"reason": "客户退单验收",
	}), http.StatusOK)
	if refunded.Status != "refunded" || refunded.RefundReason != "客户退单验收" || refunded.RefundedBy == "" || refunded.RefundedAt == nil {
		t.Fatalf("cashier refund fields not populated: %#v", refunded)
	}
	afterRefund := decodeResponse[dailyReportSummaryTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/reports/daily?date="+today, mini.Token, nil), http.StatusOK)
	if afterRefund.CashierOrderCount != before.CashierOrderCount || afterRefund.CashierAmount != before.CashierAmount {
		t.Fatalf("refunded cashier order should be excluded from report: before=%#v after=%#v", before, afterRefund)
	}

	conflict := decodeEnvelope(t, performRequest(t, handler, http.MethodPost, "/api/admin/cashier-orders/"+cashier.ID+"/refund", admin.Token, map[string]any{
		"reason": "重复退单",
	}), http.StatusConflict)
	if conflict.Code != 40904 {
		t.Fatalf("unexpected duplicate refund conflict: %#v", conflict)
	}
}

func TestAllStoresAdminScopeCanSeeInventoryAndMaterials(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	manager := passwordLogin(t, handler, defaultManagerUsername, defaultManagerPassword)
	if manager.Token == "" {
		t.Fatal("expected manager token")
	}
	inventoryItem := decodeResponse[InventoryLedgerItem](t, performRequest(t, handler, http.MethodPost, "/api/v1/inventory/items", manager.Token, map[string]any{
		"styleNo":    "SCOPE-INV-001",
		"name":       "范围验收库存",
		"category":   "黄金",
		"purity":     "足金9999",
		"pieceCount": 1,
		"weightGram": 1.23,
	}), http.StatusCreated)
	materialItem := decodeResponse[MaterialLedgerItem](t, performRequest(t, handler, http.MethodPost, "/api/v1/materials", manager.Token, map[string]any{
		"type":       "leftover",
		"orderNo":    "SCOPE-MAT-001",
		"category":   "黄金",
		"purity":     "足金9999",
		"weightGram": 1.11,
		"amount":     888,
	}), http.StatusCreated)

	adminUser := app.store.usersByUsername[defaultBossUsername]
	adminUser.DataScope = "all_stores"
	adminUser.OrgID = "legacy-admin-org"
	adminSession, err := app.store.createSession(app.Config.TokenSecret, adminUser)
	if err != nil {
		t.Fatalf("create all_stores session: %v", err)
	}
	adminInventory := decodeResponse[InventoryLedgerSummary](t, performRequest(t, handler, http.MethodGet, "/api/admin/inventory/items", adminSession.Token, nil), http.StatusOK)
	if !containsInventoryItemForTest(adminInventory.Items, inventoryItem.ID) {
		t.Fatalf("all_stores admin should see manager inventory item: %#v", adminInventory.Items)
	}
	updatedMaterial := decodeResponse[MaterialLedgerItem](t, performRequest(t, handler, http.MethodPut, "/api/admin/materials/"+materialItem.ID, adminSession.Token, map[string]any{
		"type":                materialItem.Type,
		"orderNo":             materialItem.OrderNo,
		"category":            materialItem.Category,
		"purity":              materialItem.Purity,
		"weightGram":          materialItem.WeightGram,
		"amount":              889,
		"remainingWeightGram": materialItem.RemainingWeightGram,
		"status":              "in_stock",
	}), http.StatusOK)
	if updatedMaterial.Amount != 889 {
		t.Fatalf("all_stores admin should update manager material item: %#v", updatedMaterial)
	}
}

func containsInventoryItemForTest(items []InventoryLedgerItem, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func TestInventoryAndMaterialCreateIgnoreClientDisplayFields(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()
	manager := passwordLogin(t, handler, defaultManagerUsername, defaultManagerPassword)

	inventoryItem := decodeResponse[InventoryLedgerItem](t, performRequest(t, handler, http.MethodPost, "/api/v1/inventory/items", manager.Token, map[string]any{
		"id":         "local-inventory-id",
		"storeName":  "前端展示门店",
		"styleNo":    "CLIENT-DISPLAY-001",
		"name":       "前端展示字段验收",
		"category":   "黄金",
		"purity":     "足金9999",
		"pieceCount": 1,
		"weightGram": 11,
		"costAmount": 0,
		"status":     "in_stock",
		"source":     "miniapp",
		"createdAt":  "2026-07-11 22:40",
	}), http.StatusCreated)
	if inventoryItem.ID == "local-inventory-id" || inventoryItem.StyleNo != "CLIENT-DISPLAY-001" {
		t.Fatalf("inventory display-field payload not normalized by backend: %#v", inventoryItem)
	}

	materialItem := decodeResponse[MaterialLedgerItem](t, performRequest(t, handler, http.MethodPost, "/api/v1/materials", manager.Token, map[string]any{
		"id":                  "local-material-id",
		"storeName":           "前端展示门店",
		"type":                "pledge",
		"customerName":        "前端客户",
		"category":            "黄金",
		"purity":              "足金9999",
		"weightGram":          11,
		"amount":              10000,
		"remainingWeightGram": 10,
		"status":              "in_stock",
		"createdAt":           "2026-07-11 22:41",
	}), http.StatusCreated)
	if materialItem.ID == "local-material-id" || materialItem.RemainingWeightGram != 10 {
		t.Fatalf("material display-field payload not normalized by backend: %#v", materialItem)
	}
}

func TestMiniappCatalogWriteFlowsInProcess(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()
	mini := miniappLogin(t, handler, "manager")

	createdMember := decodeResponse[MemberProfile](t, performRequest(t, handler, http.MethodPost, "/api/v1/members", mini.Token, map[string]any{
		"name":            "联调会员",
		"phone":           "13800138111",
		"level":           "VIP",
		"status":          "follow_up",
		"preferredPurity": "足金999",
		"sourceChannel":   "小程序新增",
		"managerName":     "李店长",
		"tags":            []string{"联调", "新建"},
		"notes":           "用于验证小程序会员真实写接口",
	}), http.StatusCreated)
	if createdMember.ID == "" || createdMember.StoreID != mini.StoreID || createdMember.Status != "follow_up" {
		t.Fatalf("member create not applied: %#v", createdMember)
	}

	updatedMember := decodeResponse[MemberProfile](t, performRequest(t, handler, http.MethodPut, "/api/v1/members/"+createdMember.ID, mini.Token, map[string]any{
		"name":            "联调会员已更新",
		"phone":           "13800138112",
		"level":           "普通会员",
		"status":          "sleeping",
		"preferredPurity": "18K",
		"sourceChannel":   "回访更新",
		"managerName":     "李店长",
		"tags":            []string{"联调", "更新"},
		"notes":           "会员资料已更新",
	}), http.StatusOK)
	if updatedMember.Name != "联调会员已更新" || updatedMember.Phone != "13800138112" || updatedMember.Status != "sleeping" {
		t.Fatalf("member update not applied: %#v", updatedMember)
	}

	createdProduct := decodeResponse[CatalogProduct](t, performRequest(t, handler, http.MethodPost, "/api/v1/products", mini.Token, map[string]any{
		"name":             "联调商品",
		"sku":              "INT-SKU-001",
		"category":         "金饰",
		"categoryTab":      "黄金饰品",
		"purity":           "足金999",
		"benchPrice":       720,
		"retailPrice":      760,
		"gramWeight":       10.5,
		"status":           "active",
		"inventory":        3,
		"tags":             []string{"联调", "新建"},
		"recommendedScene": "用于验证小程序商品真实写接口",
		"quoteLeadTime":    "30 秒",
	}), http.StatusCreated)
	if createdProduct.ID == "" || createdProduct.SKU != "INT-SKU-001" || len(createdProduct.StoreIDs) != 1 || createdProduct.StoreIDs[0] != mini.StoreID {
		t.Fatalf("product create not applied: %#v", createdProduct)
	}

	updatedProduct := decodeResponse[CatalogProduct](t, performRequest(t, handler, http.MethodPut, "/api/v1/products/"+createdProduct.ID, mini.Token, map[string]any{
		"name":             "联调商品已更新",
		"sku":              "INT-SKU-001",
		"category":         "K金",
		"categoryTab":      "K金",
		"purity":           "18K",
		"benchPrice":       560,
		"retailPrice":      610,
		"gramWeight":       8.8,
		"status":           "draft",
		"inventory":        2,
		"tags":             []string{"联调", "更新"},
		"recommendedScene": "商品资料已更新",
		"quoteLeadTime":    "60 秒",
	}), http.StatusOK)
	if updatedProduct.Name != "联调商品已更新" || updatedProduct.Category != "K金" || updatedProduct.Status != "draft" || updatedProduct.Inventory != 2 {
		t.Fatalf("product update not applied: %#v", updatedProduct)
	}

	memberList := decodeResponse[listResponse[MemberProfile]](t, performRequest(t, handler, http.MethodGet, "/api/v1/members", mini.Token, nil), http.StatusOK)
	if !containsMember(memberList.Items, updatedMember.ID) {
		t.Fatalf("updated member missing from list: %#v", memberList.Items)
	}

	productList := decodeResponse[listResponse[CatalogProduct]](t, performRequest(t, handler, http.MethodGet, "/api/v1/products", mini.Token, nil), http.StatusOK)
	if !containsProduct(productList.Items, updatedProduct.ID) {
		t.Fatalf("updated product missing from list: %#v", productList.Items)
	}
}

func containsMember(items []MemberProfile, memberID string) bool {
	for _, item := range items {
		if item.ID == memberID {
			return true
		}
	}
	return false
}

func containsProduct(items []CatalogProduct, productID string) bool {
	for _, item := range items {
		if item.ID == productID {
			return true
		}
	}
	return false
}

func TestGoldReferencePricesCanDriveQuotePreview(t *testing.T) {
	app := newTestAppWithEnv(t, map[string]string{
		"GOLD_PRICE_CNY_PER_GRAM": "800",
		"GOLD_PRICE_UPDATED_AT":   "2026-06-06T00:00:00Z",
	})
	handler := app.Router()
	mini := miniappLogin(t, handler, "manager")

	goldPrices := decodeResponse[goldReferencePricesTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/gold-prices/reference", mini.Token, nil), http.StatusOK)
	if goldPrices.Source != "configured_cny_per_gram" || goldPrices.BaseCNYPerGram != 800 || len(goldPrices.ReferencePrices) < 5 {
		t.Fatalf("unexpected configured gold reference prices: %#v", goldPrices)
	}

	quote := decodeResponse[recycleQuotePreviewTestResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/recycle/quote-preview", mini.Token, map[string]any{
		"storeId":             mini.StoreID,
		"category":            "旧金料",
		"purity":              "18K",
		"grossWeightGram":     10,
		"deductionWeightGram": 0,
	}), http.StatusOK)
	if quote.UnitPrice != 600 || quote.PriceSource != "configured_cny_per_gram" || quote.EstimatedAmount != 5975 {
		t.Fatalf("quote preview did not use configured international gold price: %#v", quote)
	}
}

func TestLiveGoldReferencePricesPullSpotAndFX(t *testing.T) {
	spotServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/price/XAU" {
			t.Fatalf("unexpected spot path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"symbol":    "XAU",
			"name":      "Gold",
			"currency":  "USD",
			"price":     3110.34768,
			"updatedAt": "2026-06-06T04:30:24Z",
		})
	}))
	defer spotServer.Close()
	fxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"base": "USD",
			"date": "2026-06-05",
			"rates": map[string]float64{
				"CNY": 7,
			},
		})
	}))
	defer fxServer.Close()

	app := newTestAppWithEnv(t, map[string]string{
		"GOLD_PRICE_LIVE_ENABLED":         "true",
		"GOLD_PRICE_API_URL":              spotServer.URL + "/price/XAU",
		"GOLD_PRICE_FX_API_URL":           fxServer.URL + "/latest?base=USD&symbols=CNY",
		"GOLD_PRICE_CACHE_TTL_SECONDS":    "60",
		"GOLD_PRICE_HTTP_TIMEOUT_SECONDS": "2",
	})
	handler := app.Router()
	mini := miniappLogin(t, handler, "manager")

	goldPrices := decodeResponse[goldReferencePricesTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/gold-prices/reference", mini.Token, nil), http.StatusOK)
	if goldPrices.Source != "live_xau_usd_fx" || goldPrices.BaseCNYPerGram != 700 || goldPrices.XAUUSD != 3110.35 || goldPrices.USDCNY != 7 {
		t.Fatalf("unexpected live gold reference prices: %#v", goldPrices)
	}

	quote := decodeResponse[recycleQuotePreviewTestResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/recycle/quote-preview", mini.Token, map[string]any{
		"storeId":             mini.StoreID,
		"category":            "旧金料",
		"purity":              "18K",
		"grossWeightGram":     10,
		"deductionWeightGram": 0,
	}), http.StatusOK)
	if quote.UnitPrice != 525 || quote.PriceSource != "live_xau_usd_fx" || quote.EstimatedAmount != 5225 {
		t.Fatalf("quote preview did not use live international gold price: %#v", quote)
	}
}

func TestAdminWriteFlowsPersistInBootstrap(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	admin := adminLogin(t, handler, "boss", "Boss123!")
	before := decodeResponse[AdminBootstrap](t, performRequest(t, handler, http.MethodGet, "/api/admin/bootstrap", admin.Token, nil), http.StatusOK)

	newStore := decodeResponse[AdminStoreRecord](t, performRequest(t, handler, http.MethodPost, "/api/admin/stores", admin.Token, nil), http.StatusCreated)
	updatedStore := decodeResponse[AdminStoreRecord](t, performRequest(t, handler, http.MethodPut, "/api/admin/stores/"+newStore.ID, admin.Token, map[string]any{
		"id":          newStore.ID,
		"code":        "SH-PD-01",
		"name":        "浦东首店",
		"managerName": "赵店长",
		"city":        "上海",
		"address":     "浦东新区世纪大道 1 号",
		"status":      "active",
	}), http.StatusOK)
	if updatedStore.Name != "浦东首店" {
		t.Fatalf("store update not applied: %#v", updatedStore)
	}

	newUser := decodeResponse[AdminUserAccount](t, performRequest(t, handler, http.MethodPost, "/api/admin/users", admin.Token, nil), http.StatusCreated)
	if newUser.StoreIDs == nil || newUser.StoreNames == nil {
		t.Fatalf("new user store arrays must not be null: %#v", newUser)
	}
	updatedUser := decodeResponse[AdminUserAccount](t, performRequest(t, handler, http.MethodPut, "/api/admin/users/"+newUser.ID, admin.Token, map[string]any{
		"id":        newUser.ID,
		"name":      "赵店长",
		"account":   "manager.pd",
		"roleKey":   "shop_manager",
		"roleName":  "店长",
		"dataScope": "assigned_store",
		"storeIds":  []string{updatedStore.ID},
		"status":    "enabled",
	}), http.StatusOK)
	if updatedUser.Account != "manager.pd" || updatedUser.RoleKey != "shop_manager" || len(updatedUser.StoreIDs) != 1 {
		t.Fatalf("user update not applied: %#v", updatedUser)
	}

	newProduct := decodeResponse[AdminProductRecord](t, performRequest(t, handler, http.MethodPost, "/api/admin/products", admin.Token, nil), http.StatusCreated)
	if newProduct.StoreIDs == nil || newProduct.StoreNames == nil {
		t.Fatalf("new product store arrays must not be null: %#v", newProduct)
	}
	updatedProduct := decodeResponse[AdminProductRecord](t, performRequest(t, handler, http.MethodPut, "/api/admin/products/"+newProduct.ID, admin.Token, map[string]any{
		"id":         newProduct.ID,
		"name":       "浦东试营业礼包",
		"sku":        "PD-GIFT-001",
		"category":   "活动物料",
		"price":      99,
		"gramWeight": 0,
		"status":     "active",
		"storeIds":   []string{updatedStore.ID},
		"storeNames": []string{"浦东首店"},
		"tags":       []string{"试营业", "礼包"},
	}), http.StatusOK)
	if updatedProduct.Name != "浦东试营业礼包" || len(updatedProduct.StoreIDs) != 1 || updatedProduct.StoreIDs[0] != updatedStore.ID || len(updatedProduct.StoreNames) != 1 || updatedProduct.StoreNames[0] != "浦东首店" {
		t.Fatalf("product update not applied: %#v", updatedProduct)
	}

	systemProfile := before.SystemProfile
	systemProfile.BrandName = "黄金收银正式版"
	systemProfile.ReceiptTitle = "黄金门店收银"
	updatedSystem := decodeResponse[AdminSystemProfile](t, performRequest(t, handler, http.MethodPut, "/api/admin/system-profile", admin.Token, systemProfile), http.StatusOK)
	if updatedSystem.BrandName != "黄金收银正式版" {
		t.Fatalf("system profile not updated: %#v", updatedSystem)
	}

	printTemplate := before.PrintTemplate
	printTemplate.Receipt.HeaderTitle = "黄金门店收银"
	printTemplate.Recycle.SummaryText = "回收照片已留档，可按单号复查"
	updatedPrint := decodeResponse[AdminPrintTemplate](t, performRequest(t, handler, http.MethodPut, "/api/admin/print-template", admin.Token, printTemplate), http.StatusOK)
	if updatedPrint.Receipt.HeaderTitle != "黄金门店收银" {
		t.Fatalf("print template not updated: %#v", updatedPrint)
	}

	role := before.Roles[1]
	role.Description = "本店运营台，可管理本店资料、账号和商品。"
	updatedRole := decodeResponse[AdminRoleTemplate](t, performRequest(t, handler, http.MethodPut, "/api/admin/roles/2", admin.Token, role), http.StatusOK)
	if updatedRole.Description != role.Description {
		t.Fatalf("role template not updated: %#v", updatedRole)
	}

	after := decodeResponse[AdminBootstrap](t, performRequest(t, handler, http.MethodGet, "/api/admin/bootstrap", admin.Token, nil), http.StatusOK)
	if len(after.Stores) != len(before.Stores)+1 {
		t.Fatalf("expected store count to increase: before=%d after=%d", len(before.Stores), len(after.Stores))
	}
	if len(after.Users) != len(before.Users)+1 {
		t.Fatalf("expected user count to increase: before=%d after=%d", len(before.Users), len(after.Users))
	}
	if len(after.Products) != len(before.Products)+1 {
		t.Fatalf("expected product count to increase: before=%d after=%d", len(before.Products), len(after.Products))
	}
	if after.SystemProfile.BrandName != "黄金收银正式版" {
		t.Fatalf("bootstrap did not return updated system profile: %#v", after.SystemProfile)
	}
	if after.PrintTemplate.Recycle.SummaryText != "回收照片已留档，可按单号复查" {
		t.Fatalf("bootstrap did not return updated print template: %#v", after.PrintTemplate)
	}
	if len(after.AuditLogs) <= len(before.AuditLogs) {
		t.Fatalf("expected audit log growth: before=%d after=%d", len(before.AuditLogs), len(after.AuditLogs))
	}
}

func TestProductImportTemplateAndUploadIncludeCost(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()
	admin := adminLogin(t, handler, "boss", "Boss123!")
	storeID := "store-shenzhen-nanshan"

	templateRR := performRequest(t, handler, http.MethodGet, "/api/admin/products/import-template?storeId="+storeID, admin.Token, nil)
	if templateRR.Code != http.StatusOK {
		t.Fatalf("download template failed: status=%d body=%s", templateRR.Code, templateRR.Body.String())
	}
	templateFile, err := excelize.OpenReader(bytes.NewReader(templateRR.Body.Bytes()))
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	defer templateFile.Close()
	rows, err := templateFile.GetRows(templateFile.GetSheetName(0))
	if err != nil {
		t.Fatalf("read template rows: %v", err)
	}
	if len(rows) == 0 || len(rows[0]) < 8 || rows[0][4] != "成本" || rows[0][5] != "销售价格" {
		t.Fatalf("template headers missing cost column: %#v", rows)
	}

	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	headers := []string{"商品名称", "SKU", "分类", "克重", "成本", "销售价格", "库存数量", "门店ID或名称"}
	values := []any{"成本导入测试商品", "COST-IMPORT-001", "金饰", 8.88, 700.5, 899.9, 3, storeID}
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}
	for index, value := range values {
		cell, _ := excelize.CoordinatesToCellName(index+1, 2)
		_ = file.SetCellValue(sheet, cell, value)
	}
	var xlsx bytes.Buffer
	if _, err := file.WriteTo(&xlsx); err != nil {
		t.Fatalf("write import xlsx: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("storeId", storeID); err != nil {
		t.Fatalf("write store field: %v", err)
	}
	part, err := writer.CreateFormFile("file", "cost-import.xlsx")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(xlsx.Bytes()); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/admin/products/import", &body)
	req.Header.Set("Authorization", "Bearer "+admin.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	result := decodeResponse[AdminProductImportResult](t, rr, http.StatusOK)
	if result.SuccessCount != 1 || result.FailureCount != 0 {
		t.Fatalf("unexpected import result: %#v", result)
	}

	list := decodeResponse[AdminPagedItems[AdminProductRecord]](t, performRequest(t, handler, http.MethodGet, "/api/admin/products?keyword=COST-IMPORT-001", admin.Token, nil), http.StatusOK)
	if len(list.Items) != 1 {
		t.Fatalf("imported product not found: %#v", list)
	}
	product := list.Items[0]
	if product.CostPrice != 700.5 || product.Price != 899.9 || product.Inventory != 3 {
		t.Fatalf("imported product cost/price/inventory mismatch: %#v", product)
	}
}

func TestManagerScopeRestrictions(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	admin := adminLogin(t, handler, "manager.sz", "Manager123!")
	bootstrap := decodeResponse[AdminBootstrap](t, performRequest(t, handler, http.MethodGet, "/api/admin/bootstrap", admin.Token, nil), http.StatusOK)
	if bootstrap.CurrentUser.RoleKey != "shop_manager" {
		t.Fatalf("unexpected current user: %#v", bootstrap.CurrentUser)
	}
	if len(bootstrap.Stores) != 1 {
		t.Fatalf("manager should only see one store, got %d", len(bootstrap.Stores))
	}

	createStore := performRequest(t, handler, http.MethodPost, "/api/admin/stores", admin.Token, nil)
	if createStore.Code != http.StatusForbidden {
		t.Fatalf("manager store create should be forbidden: status=%d body=%s", createStore.Code, createStore.Body.String())
	}

	createUser := performRequest(t, handler, http.MethodPost, "/api/admin/users", admin.Token, nil)
	if createUser.Code != http.StatusForbidden {
		t.Fatalf("manager user create should be forbidden: status=%d body=%s", createUser.Code, createUser.Body.String())
	}

	stores := decodeResponse[listResponse[StoreInfo]](t, performRequest(t, handler, http.MethodGet, "/api/v1/stores", admin.Token, nil), http.StatusOK)
	if stores.DataScope != "assigned_stores" || len(stores.Items) != 1 {
		t.Fatalf("manager store scope not enforced: %#v", stores)
	}
}

func TestRecyclePhotoUploadPrepareAndComplete(t *testing.T) {
	t.Run("prepare and complete success", func(t *testing.T) {
		app := newTestApp(t)
		handler := app.Router()
		mini := miniappLogin(t, handler, "manager")

		draft := decodeResponse[RecycleOrder](t, performRequest(t, handler, http.MethodPost, "/api/v1/recycle/orders", mini.Token, map[string]any{
			"storeId":         mini.StoreID,
			"customerName":    "上传测试客户",
			"customerPhone":   "13800138001",
			"estimatedAmount": 2880,
			"items": []map[string]any{
				{
					"category":   "金饰",
					"purity":     "足金999",
					"weightGram": 4.8,
				},
			},
			"remark": "upload test",
		}), http.StatusCreated)

		preparation := decodeResponse[UploadPreparation](t, performRequest(t, handler, http.MethodPost, "/api/v1/uploads/recycle-photos/prepare", mini.Token, map[string]any{
			"storeId":     mini.StoreID,
			"fileName":    "现场图-1.jpg",
			"contentType": "image/jpeg",
			"sizeBytes":   2048,
		}), http.StatusOK)
		if preparation.UploadID == "" || preparation.StoreID != mini.StoreID || preparation.ObjectKey == "" || preparation.UploadMode == "" {
			t.Fatalf("unexpected upload preparation: %#v", preparation)
		}

		asset := decodeResponse[AttachmentAsset](t, performRequest(t, handler, http.MethodPost, "/api/v1/uploads/recycle-photos/complete", mini.Token, map[string]any{
			"uploadId":     preparation.UploadID,
			"orderId":      draft.ID,
			"publicUrl":    "https://cdn.example.com/recycle/upload-test-1.jpg",
			"thumbnailUrl": "https://cdn.example.com/recycle/upload-test-1-thumb.jpg",
		}), http.StatusOK)
		if asset.OrderID != draft.ID || asset.StoreID != mini.StoreID || asset.PublicURL != "https://cdn.example.com/recycle/upload-test-1.jpg" {
			t.Fatalf("unexpected upload completion asset: %#v", asset)
		}

		detail := decodeResponse[RecycleOrder](t, performRequest(t, handler, http.MethodGet, "/api/v1/recycle/orders/"+draft.ID, mini.Token, nil), http.StatusOK)
		if len(detail.Attachments) != 1 || len(detail.AttachmentURLs) != 1 || detail.AttachmentURLs[0] != asset.PublicURL {
			t.Fatalf("attachment not persisted onto recycle order: %#v", detail)
		}
	})

	t.Run("prepare uses runtime storage config", func(t *testing.T) {
		app := newTestAppWithEnv(t, map[string]string{
			"STORAGE_ENABLED":         "true",
			"STORAGE_PROVIDER":        "cos",
			"STORAGE_BUCKET":          "gold-recycle-prod",
			"STORAGE_REGION":          "ap-guangzhou",
			"STORAGE_PUBLIC_BASE_URL": "https://cdn.example.com/gold",
			"STORAGE_PATH_PREFIX":     "acceptance-evidence",
			"STORAGE_UPLOAD_STRATEGY": "direct_put",
		})
		handler := app.Router()
		mini := miniappLogin(t, handler, "manager")

		preparation := decodeResponse[UploadPreparation](t, performRequest(t, handler, http.MethodPost, "/api/v1/uploads/recycle-photos/prepare", mini.Token, map[string]any{
			"storeId":     mini.StoreID,
			"fileName":    "runtime-env.jpg",
			"contentType": "image/jpeg",
			"sizeBytes":   4096,
		}), http.StatusOK)
		if !preparation.StorageReady || preparation.Provider != "cos" || preparation.Bucket != "gold-recycle-prod" {
			t.Fatalf("runtime storage config not applied: %#v", preparation)
		}
		if preparation.UploadMode != "direct_put" || !strings.HasPrefix(preparation.PublicURL, "https://cdn.example.com/gold/acceptance-evidence/") {
			t.Fatalf("unexpected runtime upload target: %#v", preparation)
		}
	})

	t.Run("prepare generates oss signed put url", func(t *testing.T) {
		app := newTestAppWithEnv(t, map[string]string{
			"STORAGE_ENABLED":                "true",
			"STORAGE_PROVIDER":               "oss",
			"STORAGE_BUCKET":                 "gold-recycle-prod",
			"STORAGE_REGION":                 "cn-beijing",
			"STORAGE_ENDPOINT":               "oss-cn-beijing.aliyuncs.com",
			"STORAGE_PUBLIC_BASE_URL":        "https://img.example.com/recycle",
			"STORAGE_PATH_PREFIX":            "acceptance-evidence",
			"STORAGE_UPLOAD_STRATEGY":        "oss_signed_put",
			"STORAGE_ACCESS_KEY_ID":          "test-access-key",
			"STORAGE_ACCESS_KEY_SECRET":      "test-secret-value",
			"STORAGE_UPLOAD_URL_TTL_SECONDS": "60",
		})
		handler := app.Router()
		mini := miniappLogin(t, handler, "manager")

		preparation := decodeResponse[UploadPreparation](t, performRequest(t, handler, http.MethodPost, "/api/v1/uploads/recycle-photos/prepare", mini.Token, map[string]any{
			"storeId":     mini.StoreID,
			"fileName":    "oss-runtime.jpg",
			"contentType": "image/jpeg",
			"sizeBytes":   4096,
		}), http.StatusOK)
		if !preparation.StorageReady || preparation.UploadMode != "oss_signed_put" {
			t.Fatalf("oss signed upload not ready: %#v", preparation)
		}
		if !strings.HasPrefix(preparation.UploadURL, "https://gold-recycle-prod.oss-cn-beijing.aliyuncs.com/acceptance-evidence/") {
			t.Fatalf("unexpected oss upload url: %s", preparation.UploadURL)
		}
		if !strings.Contains(preparation.UploadURL, "OSSAccessKeyId=test-access-key") || !strings.Contains(preparation.UploadURL, "Signature=") || !strings.Contains(preparation.UploadURL, "Expires=") {
			t.Fatalf("oss signed url is missing query params: %s", preparation.UploadURL)
		}
		if strings.Contains(preparation.UploadURL, "test-secret-value") {
			t.Fatalf("oss signed url leaked secret: %s", preparation.UploadURL)
		}
		if !strings.HasPrefix(preparation.PublicURL, "https://img.example.com/recycle/acceptance-evidence/") {
			t.Fatalf("unexpected oss public url: %s", preparation.PublicURL)
		}
	})

	t.Run("prepare rejects inaccessible store", func(t *testing.T) {
		app := newTestApp(t)
		handler := app.Router()
		mini := miniappLogin(t, handler, "manager")
		if len(app.store.stores) < 2 {
			t.Fatal("expected multiple seeded stores")
		}

		var otherStoreID string
		for _, store := range app.store.stores {
			if store.ID != mini.StoreID {
				otherStoreID = store.ID
				break
			}
		}
		if otherStoreID == "" {
			t.Fatal("expected an inaccessible store for manager test")
		}

		envelope := decodeError(t, performRequest(t, handler, http.MethodPost, "/api/v1/uploads/recycle-photos/prepare", mini.Token, map[string]any{
			"storeId":     otherStoreID,
			"fileName":    "forbidden.jpg",
			"contentType": "image/jpeg",
			"sizeBytes":   1024,
		}), http.StatusForbidden, 40302)
		if !strings.Contains(envelope.Message, "store not accessible") {
			t.Fatalf("unexpected forbidden message: %#v", envelope)
		}
	})

	t.Run("complete returns not found for missing order", func(t *testing.T) {
		app := newTestApp(t)
		handler := app.Router()
		mini := miniappLogin(t, handler, "manager")

		preparation := decodeResponse[UploadPreparation](t, performRequest(t, handler, http.MethodPost, "/api/v1/uploads/recycle-photos/prepare", mini.Token, map[string]any{
			"storeId":     mini.StoreID,
			"fileName":    "missing-order.jpg",
			"contentType": "image/jpeg",
			"sizeBytes":   1024,
		}), http.StatusOK)

		envelope := decodeError(t, performRequest(t, handler, http.MethodPost, "/api/v1/uploads/recycle-photos/complete", mini.Token, map[string]any{
			"uploadId":  preparation.UploadID,
			"orderId":   "recycle-order-missing",
			"publicUrl": "https://cdn.example.com/recycle/missing-order.jpg",
		}), http.StatusNotFound, 40402)
		if !strings.Contains(envelope.Message, "recycle order not found") {
			t.Fatalf("unexpected complete not found message: %#v", envelope)
		}
	})
}

func TestMiniAppLoginWithWechatCodeExchange(t *testing.T) {
	mockWechat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sns/jscode2session" {
			t.Fatalf("unexpected wechat path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("appid"); got != "wx-test-appid" {
			t.Fatalf("unexpected appid: %s", got)
		}
		if got := r.URL.Query().Get("secret"); got != "wx-test-secret" {
			t.Fatalf("unexpected secret: %s", got)
		}
		if got := r.URL.Query().Get("js_code"); got != "real-wechat-code" {
			t.Fatalf("unexpected code: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"openid":      "wx-openid-manager-001",
			"session_key": "session-key-001",
		})
	}))
	defer mockWechat.Close()

	app := newTestAppWithEnv(t, map[string]string{
		"WECHAT_MINIAPP_APP_ID":     "wx-test-appid",
		"WECHAT_MINIAPP_APP_SECRET": "wx-test-secret",
		"WECHAT_API_BASE_URL":       mockWechat.URL,
		"MINIAPP_ALLOW_MOCK_LOGIN":  "false",
	})
	handler := app.Router()

	result := decodeResponse[miniappLoginResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", map[string]any{
		"code": "real-wechat-code",
		"profile": map[string]any{
			"name":      "李店长",
			"phone":     "13800002108",
			"roleKey":   "shop_manager",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-manager-001" || result.RoleKey != "shop_manager" || result.StoreID == "" {
		t.Fatalf("unexpected miniapp login result: %#v", result)
	}
}

func TestMiniAppLoginRejectsUnboundWechatOpenID(t *testing.T) {
	mockWechat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"openid":      "wx-openid-unbound-001",
			"session_key": "session-key-002",
		})
	}))
	defer mockWechat.Close()

	app := newTestAppWithEnv(t, map[string]string{
		"WECHAT_MINIAPP_APP_ID":     "wx-test-appid",
		"WECHAT_MINIAPP_APP_SECRET": "wx-test-secret",
		"WECHAT_API_BASE_URL":       mockWechat.URL,
		"MINIAPP_ALLOW_MOCK_LOGIN":  "false",
	})
	handler := app.Router()

	envelope := decodeError(t, performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", map[string]any{
		"code": "real-wechat-code",
		"profile": map[string]any{
			"name":      "未绑定店长",
			"roleKey":   "shop_manager",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusForbidden, 40304)

	if !strings.Contains(envelope.Message, "not bound") {
		t.Fatalf("unexpected unbound login message: %#v", envelope)
	}
}

func TestMiniAppLoginAutoBindsWechatOpenIDByPhone(t *testing.T) {
	mockWechat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"openid":      "wx-openid-owner-real-001",
			"session_key": "session-key-owner-001",
		})
	}))
	defer mockWechat.Close()

	app := newTestAppWithEnv(t, map[string]string{
		"WECHAT_MINIAPP_APP_ID":     "wx-test-appid",
		"WECHAT_MINIAPP_APP_SECRET": "wx-test-secret",
		"WECHAT_API_BASE_URL":       mockWechat.URL,
		"MINIAPP_ALLOW_MOCK_LOGIN":  "false",
	})
	app.store.mu.Lock()
	owner := app.store.usersByID["user-owner-001"]
	owner.WechatOpenID = ""
	app.store.usersByID[owner.ID] = owner
	app.store.usersByUsername[owner.Username] = owner
	app.store.mu.Unlock()
	handler := app.Router()

	result := decodeResponse[miniappLoginResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", map[string]any{
		"code": "real-wechat-code",
		"profile": map[string]any{
			"name":      "廖总",
			"phone":     "13800000001",
			"roleKey":   "boss",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
			"extra":     "ignored",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-001" || result.RoleKey != "boss" {
		t.Fatalf("unexpected auto-bound miniapp login result: %#v", result)
	}

	user, ok := app.store.findUserByWechatOpenID("wx-openid-owner-real-001")
	if !ok || user.ID != "user-owner-001" {
		t.Fatalf("expected owner openid to be bound, got ok=%v user=%#v", ok, user)
	}
}

func TestMiniAppLoginAutoBindsWechatOpenIDByPhoneCode(t *testing.T) {
	mockWechat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sns/jscode2session":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"openid":      "wx-openid-owner-phone-code-001",
				"session_key": "session-key-owner-phone-code-001",
			})
		case "/cgi-bin/token":
			if got := r.URL.Query().Get("grant_type"); got != "client_credential" {
				t.Fatalf("unexpected grant_type: %s", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "wechat-access-token-001",
				"expires_in":   7200,
			})
		case "/wxa/business/getuserphonenumber":
			if got := r.URL.Query().Get("access_token"); got != "wechat-access-token-001" {
				t.Fatalf("unexpected access_token: %s", got)
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode phone body: %v", err)
			}
			if got := body["code"]; got != "phone-code-001" {
				t.Fatalf("unexpected phone code: %s", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"phone_info": map[string]any{
					"phoneNumber":     "8613800000001",
					"purePhoneNumber": "13800000001",
					"countryCode":     "86",
				},
			})
		default:
			t.Fatalf("unexpected wechat path: %s", r.URL.Path)
		}
	}))
	defer mockWechat.Close()

	app := newTestAppWithEnv(t, map[string]string{
		"WECHAT_MINIAPP_APP_ID":     "wx-test-appid",
		"WECHAT_MINIAPP_APP_SECRET": "wx-test-secret",
		"WECHAT_API_BASE_URL":       mockWechat.URL,
		"MINIAPP_ALLOW_MOCK_LOGIN":  "false",
	})
	app.store.mu.Lock()
	owner := app.store.usersByID["user-owner-001"]
	owner.WechatOpenID = ""
	app.store.usersByID[owner.ID] = owner
	app.store.usersByUsername[owner.Username] = owner
	app.store.mu.Unlock()
	handler := app.Router()

	result := decodeResponse[miniappLoginResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", map[string]any{
		"code":      "real-wechat-code",
		"phoneCode": "phone-code-001",
		"profile": map[string]any{
			"name":      "廖总",
			"roleKey":   "boss",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-001" || result.RoleKey != "boss" {
		t.Fatalf("unexpected phone-code miniapp login result: %#v", result)
	}

	user, ok := app.store.findUserByWechatOpenID("wx-openid-owner-phone-code-001")
	if !ok || user.ID != "user-owner-001" {
		t.Fatalf("expected owner openid to be bound by phone code, got ok=%v user=%#v", ok, user)
	}
}

func TestMiniAppLoginAutoBindsWechatOpenIDByEncryptedPhoneData(t *testing.T) {
	sessionKey := "1234567890abcdef"
	iv := "abcdef1234567890"
	encryptedData, encryptedIV := encryptedWechatPhonePayloadForTest(t, sessionKey, iv, "13800000001")

	mockWechat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sns/jscode2session" {
			t.Fatalf("unexpected wechat path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"openid":      "wx-openid-owner-encrypted-phone-001",
			"session_key": base64.StdEncoding.EncodeToString([]byte(sessionKey)),
		})
	}))
	defer mockWechat.Close()

	app := newTestAppWithEnv(t, map[string]string{
		"WECHAT_MINIAPP_APP_ID":     "wx-test-appid",
		"WECHAT_MINIAPP_APP_SECRET": "wx-test-secret",
		"WECHAT_API_BASE_URL":       mockWechat.URL,
		"MINIAPP_ALLOW_MOCK_LOGIN":  "false",
	})
	app.store.mu.Lock()
	owner := app.store.usersByID["user-owner-001"]
	owner.WechatOpenID = ""
	app.store.usersByID[owner.ID] = owner
	app.store.usersByUsername[owner.Username] = owner
	app.store.mu.Unlock()
	handler := app.Router()

	result := decodeResponse[miniappLoginResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", map[string]any{
		"code":               "real-wechat-code",
		"phoneEncryptedData": encryptedData,
		"phoneIv":            encryptedIV,
		"profile": map[string]any{
			"name":      "廖总",
			"roleKey":   "boss",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-001" || result.RoleKey != "boss" {
		t.Fatalf("unexpected encrypted-phone miniapp login result: %#v", result)
	}

	user, ok := app.store.findUserByWechatOpenID("wx-openid-owner-encrypted-phone-001")
	if !ok || user.ID != "user-owner-001" {
		t.Fatalf("expected owner openid to be bound by encrypted phone data, got ok=%v user=%#v", ok, user)
	}
}

func TestMiniAppLoginAutoBindsCustomerOwnerByPhoneCode(t *testing.T) {
	mockWechat := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sns/jscode2session":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"openid":      "wx-openid-zhou-owner-001",
				"session_key": "session-key-zhou-owner-001",
			})
		case "/cgi-bin/token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "wechat-access-token-zhou",
				"expires_in":   7200,
			})
		case "/wxa/business/getuserphonenumber":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"phone_info": map[string]any{
					"phoneNumber":     "8613800000000",
					"purePhoneNumber": "13800000000",
					"countryCode":     "86",
				},
			})
		default:
			t.Fatalf("unexpected wechat path: %s", r.URL.Path)
		}
	}))
	defer mockWechat.Close()

	app := newTestAppWithEnv(t, map[string]string{
		"WECHAT_MINIAPP_APP_ID":     "wx-test-appid",
		"WECHAT_MINIAPP_APP_SECRET": "wx-test-secret",
		"WECHAT_API_BASE_URL":       mockWechat.URL,
		"MINIAPP_ALLOW_MOCK_LOGIN":  "false",
	})
	handler := app.Router()

	result := decodeResponse[miniappLoginResult](t, performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", map[string]any{
		"code":      "real-wechat-code",
		"phoneCode": "phone-code-zhou",
		"profile": map[string]any{
			"name":      "示例老板",
			"roleKey":   "boss",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-002" || result.RoleKey != "boss" {
		t.Fatalf("unexpected customer owner miniapp login result: %#v", result)
	}

	user, ok := app.store.findUserByWechatOpenID("wx-openid-zhou-owner-001")
	if !ok || user.ID != "user-owner-002" || user.DisplayName != "示例老板" {
		t.Fatalf("expected customer owner openid to be bound, got ok=%v user=%#v", ok, user)
	}
}

func TestMiniAppLoginRejectsWhenWechatAuthIsRequiredButNotConfigured(t *testing.T) {
	app := newTestAppWithEnv(t, map[string]string{
		"MINIAPP_ALLOW_MOCK_LOGIN": "false",
	})
	handler := app.Router()

	envelope := decodeError(t, performRequest(t, handler, http.MethodPost, "/api/v1/auth/wechat-login", "", map[string]any{
		"code": "real-wechat-code",
		"profile": map[string]any{
			"name":      "李店长",
			"roleKey":   "shop_manager",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusServiceUnavailable, 50301)

	if !strings.Contains(envelope.Message, "not configured") {
		t.Fatalf("unexpected missing-config message: %#v", envelope)
	}
}

func TestNewPersistentModeRequiresBackingServices(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name: "missing mysql dsn",
			env: map[string]string{
				"APP_MODE":   "persistent",
				"REDIS_ADDR": "127.0.0.1:6379",
			},
			wantErr: "persistent mode requires MYSQL_DSN",
		},
		{
			name: "missing redis addr",
			env: map[string]string{
				"APP_MODE":  "persistent",
				"MYSQL_DSN": "gold:secret@tcp(127.0.0.1:3306)/gold_recycle?parseTime=true",
			},
			wantErr: "persistent mode requires REDIS_ADDR",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{
				"APP_MODE",
				"PORT",
				"TOKEN_SECRET",
				"CORS_ORIGIN",
				"WECHAT_MINIAPP_APP_ID",
				"WECHAT_MINIAPP_APP_SECRET",
				"WECHAT_API_BASE_URL",
				"MINIAPP_ALLOW_MOCK_LOGIN",
				"MYSQL_DSN",
				"MYSQL_HOST",
				"MYSQL_PORT",
				"MYSQL_DATABASE",
				"MYSQL_USER",
				"MYSQL_PASSWORD",
				"MYSQL_CHARSET",
				"REDIS_ADDR",
				"REDIS_USER",
				"REDIS_PASSWORD",
				"REDIS_DB",
				"REDIS_KEY_PREFIX",
			} {
				t.Setenv(key, "")
			}
			t.Setenv("PORT", "8080")
			t.Setenv("TOKEN_SECRET", "test-secret")
			t.Setenv("CORS_ORIGIN", "*")
			for key, value := range tc.env {
				t.Setenv(key, value)
			}

			app, err := New()
			if app != nil {
				t.Cleanup(func() {
					_ = app.Close()
				})
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
