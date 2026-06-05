package app

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

type inventorySummaryTestResult struct {
	TotalSKU             int             `json:"totalSku"`
	LowStockSKU          int             `json:"lowStockSku"`
	EstimatedRetailValue float64         `json:"estimatedRetailValue"`
	Items                []InventoryItem `json:"items"`
}

type dailyReportSummaryTestResult struct {
	Date              string                   `json:"date"`
	VisibleStoreCount int                      `json:"visibleStoreCount"`
	StoreMetrics      []DailyReportStoreMetric `json:"storeMetrics"`
}

func newTestApp(t *testing.T) *App {
	return newTestAppWithEnv(t, nil)
}

func newTestAppWithEnv(t *testing.T, overrides map[string]string) *App {
	t.Helper()
	defaults := map[string]string{
		"APP_MODE":                  "memory",
		"PORT":                      "8080",
		"TOKEN_SECRET":              "test-secret",
		"CORS_ORIGIN":               "*",
		"WECHAT_MINIAPP_APP_ID":     "",
		"WECHAT_MINIAPP_APP_SECRET": "",
		"WECHAT_API_BASE_URL":       "",
		"MINIAPP_ALLOW_MOCK_LOGIN":  "true",
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
	updatedUser := decodeResponse[AdminUserAccount](t, performRequest(t, handler, http.MethodPut, "/api/admin/users/"+newUser.ID, admin.Token, map[string]any{
		"id":        newUser.ID,
		"name":      "赵收银",
		"account":   "cashier.pd",
		"roleKey":   "manager",
		"roleName":  "店长",
		"dataScope": "assigned_store",
		"storeIds":  []string{updatedStore.ID},
		"status":    "enabled",
	}), http.StatusOK)
	if updatedUser.Account != "cashier.pd" || len(updatedUser.StoreIDs) != 1 {
		t.Fatalf("user update not applied: %#v", updatedUser)
	}

	newProduct := decodeResponse[AdminProductRecord](t, performRequest(t, handler, http.MethodPost, "/api/admin/products", admin.Token, nil), http.StatusCreated)
	updatedProduct := decodeResponse[AdminProductRecord](t, performRequest(t, handler, http.MethodPut, "/api/admin/products/"+newProduct.ID, admin.Token, map[string]any{
		"id":         newProduct.ID,
		"name":       "浦东试营业礼包",
		"sku":        "PD-GIFT-001",
		"category":   "活动物料",
		"price":      99,
		"gramWeight": 0,
		"status":     "active",
		"storeNames": []string{"浦东首店"},
		"tags":       []string{"试营业", "礼包"},
	}), http.StatusOK)
	if updatedProduct.Name != "浦东试营业礼包" {
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

func TestManagerScopeRestrictions(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	admin := adminLogin(t, handler, "manager.sz", "Manager123!")
	bootstrap := decodeResponse[AdminBootstrap](t, performRequest(t, handler, http.MethodGet, "/api/admin/bootstrap", admin.Token, nil), http.StatusOK)
	if bootstrap.CurrentUser.RoleKey != "manager" {
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
			"roleKey":   "manager",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-manager-001" || result.RoleKey != "manager" || result.StoreID == "" {
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
			"name":      "未绑定店员",
			"roleKey":   "manager",
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
			"roleKey":   "owner",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
			"extra":     "ignored",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-001" || result.RoleKey != "owner" {
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
			"roleKey":   "owner",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-001" || result.RoleKey != "owner" {
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
			"roleKey":   "owner",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-001" || result.RoleKey != "owner" {
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
			"roleKey":   "owner",
			"storeName": "南山旗舰店",
			"storeCode": "SZ-NS",
		},
	}), http.StatusOK)

	if result.UserID != "user-owner-002" || result.RoleKey != "owner" {
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
			"roleKey":   "manager",
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
