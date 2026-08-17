package app

import (
	"net/http"
	"strings"
	"testing"
)

type customerProductListTestResponse struct {
	Items      []CustomerProductDTO `json:"items"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	Categories []string             `json:"categories"`
}

func TestCustomerBrowseEndpointsExposePublicSafeData(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()

	app.store.mu.Lock()
	activeIDs := make(map[string]struct {
		longitude float64
		latitude  float64
	})
	activeIndex := 0
	for i := range app.store.stores {
		if app.store.stores[i].Status != "active" {
			continue
		}
		longitude := 113.2644 + float64(activeIndex)
		latitude := 23.1291 + float64(activeIndex)
		app.store.stores[i].Longitude = longitude
		app.store.stores[i].Latitude = latitude
		activeIDs[app.store.stores[i].ID] = struct {
			longitude float64
			latitude  float64
		}{longitude: longitude, latitude: latitude}
		activeIndex++
	}
	app.store.mu.Unlock()

	home := decodeResponse[CustomerHomeResponse](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/home", "", nil), http.StatusOK)
	if home.BrandName == "" {
		t.Fatal("customer home should expose a brand name")
	}

	storeResponse := performRequest(t, handler, http.MethodGet, "/api/v1/customer/stores", "", nil)
	stores := decodeResponse[[]CustomerStoreDTO](t, storeResponse, http.StatusOK)
	if len(stores) != len(activeIDs) {
		t.Fatalf("customer store count = %d, want %d active stores", len(stores), len(activeIDs))
	}
	if strings.Contains(storeResponse.Body.String(), `"manager"`) {
		t.Fatalf("customer store response leaked staff-only fields: %s", storeResponse.Body.String())
	}
	for _, store := range stores {
		want, ok := activeIDs[store.ID]
		if !ok {
			t.Fatalf("inactive or unknown store leaked to customer list: %#v", store)
		}
		if store.Longitude != want.longitude || store.Latitude != want.latitude {
			t.Fatalf("store %s coordinates = (%v, %v), want (%v, %v)", store.ID, store.Longitude, store.Latitude, want.longitude, want.latitude)
		}
	}

	firstStoreID := stores[0].ID
	detail := decodeResponse[CustomerStoreDetailDTO](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/stores/"+firstStoreID, "", nil), http.StatusOK)
	if detail.ID != firstStoreID || detail.Longitude == 0 || detail.Latitude == 0 {
		t.Fatalf("customer store detail missing coordinates: %#v", detail)
	}

	products := decodeResponse[customerProductListTestResponse](t, performRequest(t, handler, http.MethodGet, "/api/v1/customer/products?page=1&pageSize=1", "", nil), http.StatusOK)
	if products.Page != 1 || products.PageSize != 1 {
		t.Fatalf("customer product pagination = page %d/pageSize %d", products.Page, products.PageSize)
	}
	if len(products.Items) > 1 {
		t.Fatalf("customer product page returned %d items, want at most 1", len(products.Items))
	}
	if len(products.Items) > 0 {
		productResponse := performRequest(t, handler, http.MethodGet, "/api/v1/customer/products/"+products.Items[0].ID, "", nil)
		product := decodeResponse[CustomerProductDTO](t, productResponse, http.StatusOK)
		if product.ID != products.Items[0].ID {
			t.Fatalf("customer product detail id = %q, want %q", product.ID, products.Items[0].ID)
		}
		if strings.Contains(productResponse.Body.String(), `"costPrice"`) || strings.Contains(productResponse.Body.String(), `"benchPrice"`) {
			t.Fatalf("customer product response leaked internal price fields: %s", productResponse.Body.String())
		}
	}
}
