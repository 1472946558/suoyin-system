package app

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestCustomerRecycleInfoV1Boundary(t *testing.T) {
	safe := defaultCustomerRecycleInfo()
	if err := validateCustomerRecycleInfoV1(safe); err != nil {
		t.Fatalf("default recycle info violates V1 boundary: %v", err)
	}

	unsafe := CustomerRecycleInfo{
		Title: "黄金回收服务",
		Intro: "根据实时金价在线估价，现场结算并即时到账",
		Process: []RecycleProcessStep{
			{Step: 1, Title: "上门回收", Desc: "提供邮寄回收"},
		},
	}
	if !errors.Is(validateCustomerRecycleInfoV1(unsafe), errRecycleInfoV1Boundary) {
		t.Fatal("unsafe recycle content should be rejected by V1 boundary")
	}

	app := newTestApp(t)
	handler := app.Router()
	admin := adminLogin(t, handler, "boss", "Boss123!")
	response := performRequest(t, handler, http.MethodPut, "/api/admin/customer/recycle-info", admin.Token, unsafe)
	decodeError(t, response, http.StatusBadRequest, 40006)

	app.store.mu.Lock()
	app.store.customerRecycleInfo = unsafe
	app.store.mu.Unlock()
	publicResponse := performRequest(t, handler, http.MethodGet, "/api/v1/customer/recycle-info", "", nil)
	publicInfo := decodeResponse[CustomerRecycleInfo](t, publicResponse, http.StatusOK)
	if err := validateCustomerRecycleInfoV1(publicInfo); err != nil {
		t.Fatalf("public recycle info violates V1 boundary: %v; body=%s", err, publicResponse.Body.String())
	}
	if strings.Contains(publicResponse.Body.String(), "实时金价") || strings.Contains(publicResponse.Body.String(), "即时到账") {
		t.Fatalf("unsafe persisted recycle info leaked to customer endpoint: %s", publicResponse.Body.String())
	}
}
