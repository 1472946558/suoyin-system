package app

import (
	"errors"
	"net/http"
	"strings"
)

type materialOutboundRequest struct {
	Remark string `json:"remark"`
}

type inventoryItemCreateRequest struct {
	StoreID    string  `json:"storeId"`
	StyleNo    string  `json:"styleNo"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Purity     string  `json:"purity"`
	PieceCount int     `json:"pieceCount"`
	WeightGram float64 `json:"weightGram"`
	CostAmount float64 `json:"costAmount"`
	Status     string  `json:"status"`
	Source     string  `json:"source"`
	Remark     string  `json:"remark"`
}

type materialItemCreateRequest struct {
	StoreID             string  `json:"storeId"`
	Type                string  `json:"type"`
	OrderNo             string  `json:"orderNo"`
	CustomerName        string  `json:"customerName"`
	Category            string  `json:"category"`
	Purity              string  `json:"purity"`
	WeightGram          float64 `json:"weightGram"`
	Amount              float64 `json:"amount"`
	RemainingWeightGram float64 `json:"remainingWeightGram"`
	Status              string  `json:"status"`
	DueDate             string  `json:"dueDate"`
	Remark              string  `json:"remark"`
	Source              string  `json:"source"`
}

func (req inventoryItemCreateRequest) toLedgerItem() InventoryLedgerItem {
	return InventoryLedgerItem{
		StoreID:    req.StoreID,
		StyleNo:    req.StyleNo,
		Name:       req.Name,
		Category:   req.Category,
		Purity:     req.Purity,
		PieceCount: req.PieceCount,
		WeightGram: req.WeightGram,
		CostAmount: req.CostAmount,
		Status:     req.Status,
		Source:     req.Source,
		Remark:     req.Remark,
	}
}

func (req materialItemCreateRequest) toLedgerItem() MaterialLedgerItem {
	return MaterialLedgerItem{
		StoreID:             req.StoreID,
		Type:                req.Type,
		OrderNo:             req.OrderNo,
		CustomerName:        req.CustomerName,
		Category:            req.Category,
		Purity:              req.Purity,
		WeightGram:          req.WeightGram,
		Amount:              req.Amount,
		RemainingWeightGram: req.RemainingWeightGram,
		Status:              req.Status,
		DueDate:             req.DueDate,
		Remark:              req.Remark,
		Source:              req.Source,
	}
}

func (a *App) handleInventoryItems(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())
	switch r.Method {
	case http.MethodGet:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.listInventoryLedger(user))
	case http.MethodPost:
		var req inventoryItemCreateRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		item, err := a.store.createInventoryLedgerItem(user, req.toLedgerItem())
		switch {
		case errors.Is(err, errUnauthorizedStore):
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to create inventory item")
		default:
			a.writeJSON(w, r, http.StatusCreated, 0, "ok", item)
		}
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleAdminInventoryItems(w http.ResponseWriter, r *http.Request) {
	a.handleInventoryItems(w, r)
}

func (a *App) handleMaterialItems(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r.Context())
	switch r.Method {
	case http.MethodGet:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", a.store.listMaterialLedger(user))
	case http.MethodPost:
		var req materialItemCreateRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		item, err := a.store.createMaterialLedgerItem(user, req.toLedgerItem())
		switch {
		case errors.Is(err, errUnauthorizedStore):
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to create material item")
		default:
			a.writeJSON(w, r, http.StatusCreated, 0, "ok", item)
		}
	default:
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
	}
}

func (a *App) handleAdminMaterialItems(w http.ResponseWriter, r *http.Request) {
	a.handleMaterialItems(w, r)
}

func (a *App) handleMaterialActions(w http.ResponseWriter, r *http.Request) {
	a.handleMaterialActionPath(w, r, "/api/v1/materials/")
}

func (a *App) handleAdminMaterialActions(w http.ResponseWriter, r *http.Request) {
	a.handleMaterialActionPath(w, r, "/api/admin/materials/")
}

func (a *App) handleMaterialActionPath(w http.ResponseWriter, r *http.Request, prefix string) {
	if r.Method == http.MethodPut {
		itemID := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
		if itemID == "" || strings.Contains(itemID, "/") {
			a.writeError(w, r, http.StatusBadRequest, 40002, "invalid material id")
			return
		}
		var req materialItemCreateRequest
		if err := decodeJSONLenient(r, &req); err != nil {
			a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
			return
		}
		item, err := a.store.updateMaterialLedgerItem(currentUser(r.Context()), itemID, req.toLedgerItem())
		switch {
		case errors.Is(err, errUnauthorizedStore):
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		case errors.Is(err, errMaterialNotFound):
			a.writeError(w, r, http.StatusNotFound, 40402, "material item not found")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to update material item")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", item)
		}
		return
	}
	if r.Method == http.MethodDelete {
		itemID := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
		if itemID == "" || strings.Contains(itemID, "/") {
			a.writeError(w, r, http.StatusBadRequest, 40002, "invalid material id")
			return
		}
		item, err := a.store.deleteMaterialLedgerItem(currentUser(r.Context()), itemID)
		switch {
		case errors.Is(err, errUnauthorizedStore):
			a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
		case errors.Is(err, errMaterialNotFound):
			a.writeError(w, r, http.StatusNotFound, 40402, "material item not found")
		case err != nil:
			a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to delete material item")
		default:
			a.writeJSON(w, r, http.StatusOK, 0, "ok", item)
		}
		return
	}
	if r.Method != http.MethodPost {
		a.writeError(w, r, http.StatusMethodNotAllowed, 40005, "method not allowed")
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if !strings.HasSuffix(path, "/outbound") {
		a.writeError(w, r, http.StatusNotFound, 40404, "material action not found")
		return
	}
	itemID := strings.Trim(strings.TrimSuffix(path, "/outbound"), "/")
	if itemID == "" {
		a.writeError(w, r, http.StatusBadRequest, 40002, "invalid material id")
		return
	}
	var req materialOutboundRequest
	if err := decodeJSON(r, &req); err != nil {
		a.writeError(w, r, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	item, err := a.store.outboundMaterialLedgerItem(currentUser(r.Context()), itemID, req.Remark)
	switch {
	case errors.Is(err, errUnauthorizedStore):
		a.writeError(w, r, http.StatusForbidden, 40302, "store not accessible")
	case errors.Is(err, errMaterialNotFound):
		a.writeError(w, r, http.StatusNotFound, 40402, "material item not found")
	case err != nil:
		a.writeError(w, r, http.StatusInternalServerError, 50005, "failed to outbound material item")
	default:
		a.writeJSON(w, r, http.StatusOK, 0, "ok", item)
	}
}
