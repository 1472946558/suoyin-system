package app

import (
	"net/http"
	"testing"
)

func TestCustomerAppointmentOwnerIsolationAndActionFlags(t *testing.T) {
	app := newTestApp(t)
	handler := app.Router()
	storeID := testActiveStoreID(t, app)
	date, appointmentTime := testAppointmentSlot()

	owner, ownerToken := testCustomerSession(t, app, "13800000333")
	_, otherToken := testCustomerSession(t, app, "13800000444")
	appointment, err := app.store.createCustomerAppointment(owner, CreateAppointmentRequest{
		StoreID:         storeID,
		ServiceType:     ServiceTypeConsult,
		AppointmentDate: date,
		AppointmentTime: appointmentTime,
		ContactName:     "预约所有者",
		ContactPhone:    owner.Phone,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	ownerDetail := performRequest(t, handler, http.MethodGet, "/api/v1/customer/appointments/"+appointment.ID, ownerToken, nil)
	detail := decodeResponse[CustomerAppointmentDTO](t, ownerDetail, http.StatusOK)
	if !detail.CanCancel || !detail.CanEditNotes {
		t.Fatalf("owner action flags = cancel:%v notes:%v, want both true", detail.CanCancel, detail.CanEditNotes)
	}

	otherDetail := performRequest(t, handler, http.MethodGet, "/api/v1/customer/appointments/"+appointment.ID, otherToken, nil)
	decodeError(t, otherDetail, http.StatusNotFound, 40401)

	otherCancel := performRequest(t, handler, http.MethodPost, "/api/v1/customer/appointments/"+appointment.ID+"/cancel", otherToken, map[string]any{
		"reason": "越权测试",
	})
	decodeError(t, otherCancel, http.StatusNotFound, 40401)

	otherNotes := performRequest(t, handler, http.MethodPut, "/api/v1/customer/appointments/"+appointment.ID+"/notes", otherToken, map[string]any{
		"remark": "越权修改",
	})
	decodeError(t, otherNotes, http.StatusNotFound, 40401)

	unchanged, ok := app.store.getCustomerAppointment(owner.ID, appointment.ID)
	if !ok {
		t.Fatal("owner appointment not found after isolation checks")
	}
	if unchanged.Remark != "" || unchanged.Status != AppointmentStatusPending {
		t.Fatalf("owner appointment mutated by other customer: status=%q remark=%q", unchanged.Status, unchanged.Remark)
	}
}
