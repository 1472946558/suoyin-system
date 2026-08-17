package app

import (
	"net/http"
	"testing"
	"time"
)

func testCustomerSession(t *testing.T, app *App, phone string) (CustomerProfile, string) {
	t.Helper()

	profile, _ := app.store.findOrCreateCustomerByOpenID(
		"org-gold-v1",
		"test-customer-"+generateID(),
		"",
		"预约测试顾客",
		"",
	)
	if phone != "" {
		if err := app.store.updateCustomerPhone(profile.ID, phone); err != nil {
			t.Fatalf("update customer phone: %v", err)
		}
		profile, _ = app.store.findCustomerByOpenID(profile.OpenID)
	}

	session, err := app.store.createCustomerSession(app.Config.TokenSecret, profile)
	if err != nil {
		t.Fatalf("create customer session: %v", err)
	}
	return profile, session.Token
}

func testAppointmentSlot() (string, string) {
	return time.Now().AddDate(0, 0, 2).Format("2006-01-02"), "20:00"
}

func testActiveStoreID(t *testing.T, app *App) string {
	t.Helper()

	app.store.mu.RLock()
	defer app.store.mu.RUnlock()
	for _, store := range app.store.stores {
		if store.Status == "active" {
			return store.ID
		}
	}
	t.Fatal("no active store in test fixtures")
	return ""
}

func TestCustomerAppointmentCancelDoesNotDeadlock(t *testing.T) {
	app := newTestApp(t)
	customer, _ := testCustomerSession(t, app, "13800000111")
	storeID := testActiveStoreID(t, app)
	date, appointmentTime := testAppointmentSlot()

	appointment, err := app.store.createCustomerAppointment(customer, CreateAppointmentRequest{
		StoreID:         storeID,
		ServiceType:     ServiceTypeConsult,
		AppointmentDate: date,
		AppointmentTime: appointmentTime,
		ContactName:     "预约测试顾客",
		ContactPhone:    customer.Phone,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- app.store.cancelCustomerAppointment(customer.ID, appointment.ID, "测试取消")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("cancel appointment: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel appointment timed out; possible mutex re-entry deadlock")
	}

	cancelled, ok := app.store.getCustomerAppointment(customer.ID, appointment.ID)
	if !ok {
		t.Fatal("cancelled appointment not found")
	}
	if cancelled.Status != AppointmentStatusCancelled {
		t.Fatalf("appointment status = %q, want %q", cancelled.Status, AppointmentStatusCancelled)
	}
}

func TestCustomerAppointmentRequiresVerifiedPhoneAndMapsCapacity(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()
	storeID := testActiveStoreID(t, app)
	date, appointmentTime := testAppointmentSlot()

	_, unverifiedToken := testCustomerSession(t, app, "")
	unverified := performRequest(t, handler, http.MethodPost, "/api/v1/customer/appointments", unverifiedToken, map[string]any{
		"storeId":         storeID,
		"serviceType":     ServiceTypeConsult,
		"appointmentDate": date,
		"appointmentTime": appointmentTime,
		"contactName":     "未授权顾客",
		"contactPhone":    "13800000999",
	})
	decodeError(t, unverified, http.StatusForbidden, 40310)

	app.store.mu.Lock()
	rules := defaultAppointmentRules()
	rules.SlotCapacity = 1
	app.store.appointmentRules = rules
	app.store.mu.Unlock()

	firstCustomer, firstToken := testCustomerSession(t, app, "13800000111")
	_, secondToken := testCustomerSession(t, app, "13800000222")
	request := map[string]any{
		"storeId":         storeID,
		"serviceType":     ServiceTypeConsult,
		"appointmentDate": date,
		"appointmentTime": appointmentTime,
		"contactName":     "已授权顾客",
	}

	first := performRequest(t, handler, http.MethodPost, "/api/v1/customer/appointments", firstToken, request)
	decodeResponse[CustomerAppointmentDTO](t, first, http.StatusCreated)

	second := performRequest(t, handler, http.MethodPost, "/api/v1/customer/appointments", secondToken, request)
	decodeError(t, second, http.StatusConflict, 40904)

	app.store.mu.RLock()
	defer app.store.mu.RUnlock()
	if len(app.store.customerAppointments) == 0 {
		t.Fatal("expected created appointment")
	}
	last := app.store.customerAppointments[len(app.store.customerAppointments)-1]
	if last.CustomerPhone != firstCustomer.Phone {
		t.Fatalf("appointment phone = %q, want server-side verified phone %q", last.CustomerPhone, firstCustomer.Phone)
	}
}

func TestCustomerProfileDTOExposesVerificationStateWithoutRawPhone(t *testing.T) {
	unverified := toCustomerProfileDTO(CustomerProfile{ID: "customer-1"})
	if unverified.PhoneVerified {
		t.Fatal("unverified customer should not report phoneVerified")
	}

	verified := toCustomerProfileDTO(CustomerProfile{ID: "customer-2", Phone: "13800000111"})
	if !verified.PhoneVerified {
		t.Fatal("verified customer should report phoneVerified")
	}
	if verified.Phone == "13800000111" {
		t.Fatal("customer DTO must not expose raw phone")
	}
}
