package app

import (
	"net/http"
	"testing"
	"time"
)

// TestAdminCustomerContentCoversCustomerEndpoints 验证后台「顾客端内容管理」接口
// 与顾客端公开接口的联动：首页文案/Banner、款式扩展字段、门店扩展配置、预约规则。
func TestAdminCustomerContentCoversCustomerEndpoints(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()
	admin := adminLogin(t, handler, "boss", "Boss123!")

	// --- 1. 首页文案配置 ---
	updatedHome := decodeResponse[AdminHomeConfigRequest](t, performRequest(t, handler, http.MethodPut, "/api/admin/customer-home/config", admin.Token, map[string]any{
		"brandName":    "金匠馆测试品牌",
		"brandSlogan1": "专业黄金服务",
		"brandSlogan2": "值得信赖",
		"serviceCopy":  "以旧换新 · 维修保养 · 专业咨询",
		"entryFeeText": "工费透明 · 明码标价",
	}), http.StatusOK)
	if updatedHome.BrandName != "金匠馆测试品牌" {
		t.Fatalf("home config brand not applied: %#v", updatedHome)
	}

	customerHome := decodeResponse[CustomerHomeResponse](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/home", "", nil), http.StatusOK)
	if customerHome.BrandName != "金匠馆测试品牌" || customerHome.ServiceCopy != "以旧换新 · 维修保养 · 专业咨询" {
		t.Fatalf("customer home did not reflect admin config: %#v", customerHome)
	}

	// --- 2. Banner：新增两条（一条停用），顾客端只看到启用的 ---
	bannerA := decodeResponse[CustomerBanner](t, performRequest(t, handler, http.MethodPost, "/api/admin/customer-home/banners", admin.Token, map[string]any{
		"title":    "主推活动",
		"imageUrl": "/assets/banner/a.jpg",
		"linkType": "styles",
		"sortOrder": 5,
	}), http.StatusOK)
	disabled := false
	bannerB := decodeResponse[CustomerBanner](t, performRequest(t, handler, http.MethodPost, "/api/admin/customer-home/banners", admin.Token, map[string]any{
		"title":    "停用Banner",
		"imageUrl": "/assets/banner/b.jpg",
		"linkType": "none",
		"sortOrder": 1,
		"enabled":  &disabled,
	}), http.StatusOK)

	customerHome = decodeResponse[CustomerHomeResponse](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/home", "", nil), http.StatusOK)
	foundA, foundB := false, false
	for _, b := range customerHome.Banners {
		if b.ID == bannerA.ID {
			foundA = true
		}
		if b.ID == bannerB.ID {
			foundB = true
		}
	}
	if !foundA {
		t.Fatalf("enabled banner missing on customer home: %#v", customerHome.Banners)
	}
	if foundB {
		t.Fatalf("disabled banner leaked to customer home: %#v", customerHome.Banners)
	}

	// Banner 上限 8 条启用：补齐到 8 后再新增应 400
	existing := decodeResponse[CustomerHomeResponse](t, performRequest(t, handler, http.MethodGet, "/api/admin/customer-home/config", admin.Token, nil), http.StatusOK)
	enabledCount := 0
	for _, b := range existing.Banners {
		if bannerEnabled(b.Enabled) {
			enabledCount++
		}
	}
	for i := enabledCount; i < 8; i++ {
		performRequest(t, handler, http.MethodPost, "/api/admin/customer-home/banners", admin.Token, map[string]any{
			"title":    "填充Banner",
			"imageUrl": "/assets/banner/f.jpg",
			"linkType": "none",
		})
	}
	exceeded := performRequest(t, handler, http.MethodPost, "/api/admin/customer-home/banners", admin.Token, map[string]any{
		"title":    "超限Banner",
		"imageUrl": "/assets/banner/over.jpg",
		"linkType": "none",
	})
	if exceeded.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when exceeding 8 enabled banners, got %d body=%s", exceeded.Code, exceeded.Body.String())
	}

	// --- 3. 款式/工费管理：新建款式带扩展字段，顾客端可见 ---
	styleRecord := decodeResponse[AdminCustomerProductRecord](t, performRequest(t, handler, http.MethodPost, "/api/admin/customer-products", admin.Token, map[string]any{
		"name":       "足金传承手镯",
		"imageUrl":   "/assets/style/bracelet.jpg",
		"category":   "手镯",
		"purity":     "足金999",
		"retailPrice": 38000,
		"gramWeight": 30.5,
		"laborFeeRef": "35元/克 起",
		"description": "古法工艺，经典传承",
		"images":     []string{"/assets/style/bracelet.jpg", "/assets/style/bracelet-2.jpg"},
		"isHot":      true,
	}), http.StatusOK)
	if styleRecord.LaborFeeRef != "35元/克 起" {
		t.Fatalf("style laborFeeRef not saved: %#v", styleRecord)
	}

	customerProduct := decodeResponse[CustomerProductDTO](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/products/"+styleRecord.ID, "", nil), http.StatusOK)
	if customerProduct.LaborFeeRef != "35元/克 起" || len(customerProduct.Images) != 2 || !customerProduct.IsHot {
		t.Fatalf("customer product missing admin-configured extension fields: %#v", customerProduct)
	}

	// --- 4. 门店扩展配置：关闭预约开关后时段为空，顾客端可见开关状态 ---
	stores := decodeResponse[[]CustomerStoreDTO](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/stores", "", nil), http.StatusOK)
	if len(stores) == 0 {
		t.Fatalf("no active stores in fixtures")
	}
	storeID := stores[0].ID
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	off := false
	decodeResponse[AdminCustomerStoreConfig](t, performRequest(t, handler, http.MethodPut, "/api/admin/stores/"+storeID+"/customer-config", admin.Token, map[string]any{
		"appointmentEnabled": &off,
		"serviceTags":        []string{"以旧换新", "维修"},
	}), http.StatusOK)

	slotsResp := decodeResponse[customerSlotsTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/stores/"+storeID+"/slots?date="+tomorrow, "", nil), http.StatusOK)
	if len(slotsResp.Slots) != 0 {
		t.Fatalf("expected no slots when store appointment disabled, got %d", len(slotsResp.Slots))
	}

	on := true
	decodeResponse[AdminCustomerStoreConfig](t, performRequest(t, handler, http.MethodPut, "/api/admin/stores/"+storeID+"/customer-config", admin.Token, map[string]any{
		"appointmentEnabled": &on,
	}), http.StatusOK)

	// --- 5. 预约规则：非法值 400；合法值改变时段粒度 ---
	invalidRules := performRequest(t, handler, http.MethodPut, "/api/admin/appointment-rules", admin.Token, map[string]any{
		"bookableDays":       7,
		"slotMinutes":        45, // 非法：只允许 15/30/60
		"openTime":           "09:30",
		"closeTime":          "21:30",
		"sameDayLeadMinutes": 60,
		"cancelLeadMinutes":  120,
		"slotCapacity":       2,
	})
	if invalidRules.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid slotMinutes, got %d body=%s", invalidRules.Code, invalidRules.Body.String())
	}

	decodeResponse[AppointmentRules](t, performRequest(t, handler, http.MethodPut, "/api/admin/appointment-rules", admin.Token, map[string]any{
		"bookableServiceTypes": []string{"OLD_FOR_NEW", "REPAIR", "CONSULT", "RECYCLE"},
		"bookableDays":         7,
		"slotMinutes":          60,
		"openTime":             "09:00",
		"closeTime":            "12:00",
		"sameDayLeadMinutes":   60,
		"cancelLeadMinutes":    120,
		"slotCapacity":         2,
	}), http.StatusOK)

	slotsResp = decodeResponse[customerSlotsTestResult](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/stores/"+storeID+"/slots?date="+tomorrow, "", nil), http.StatusOK)
	if len(slotsResp.Slots) != 3 { // 09:00 / 10:00 / 11:00
		t.Fatalf("expected 3 hourly slots for 09:00-12:00 with 60min granularity, got %d: %#v", len(slotsResp.Slots), slotsResp.Slots)
	}
	if len(slotsResp.Slots) > 0 && slotsResp.Slots[0].Time != "09:00" {
		t.Fatalf("expected first slot 09:00, got %s", slotsResp.Slots[0].Time)
	}
}

type customerSlotsTestResult struct {
	StoreID string `json:"storeId"`
	Date    string `json:"date"`
	Slots   []struct {
		Time      string `json:"time"`
		Available bool   `json:"available"`
		State     string `json:"state"`
	} `json:"slots"`
}
