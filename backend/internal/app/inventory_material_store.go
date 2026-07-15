package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (s *MockStore) listInventoryLedger(user UserAccount) InventoryLedgerSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visibleStores, storeSet := visibleStoresForUserLocked(s.stores, user)
	summary := InventoryLedgerSummary{
		OrgID:             user.OrgID,
		VisibleStoreCount: len(visibleStores),
		Items:             make([]InventoryLedgerItem, 0, len(s.inventoryLedger)),
	}
	for _, item := range s.inventoryLedger {
		if !ledgerOrgVisibleToUserLocked(item.OrgID, user) || !storeVisibleToUserLocked(item.StoreID, user, storeSet) || item.Status == "deleted" {
			continue
		}
		summary.Items = append(summary.Items, item)
		if item.Status != "outbound" {
			summary.TotalStyleCount++
			summary.TotalPieceCount += item.PieceCount
			summary.TotalWeightGram += item.WeightGram
			summary.TotalCostAmount += item.CostAmount
		}
	}
	sort.Slice(summary.Items, func(i, j int) bool {
		return summary.Items[i].CreatedAt.After(summary.Items[j].CreatedAt)
	})
	summary.TotalWeightGram = round2(summary.TotalWeightGram)
	summary.TotalCostAmount = round2(summary.TotalCostAmount)
	return summary
}

func (s *MockStore) createInventoryLedgerItem(user UserAccount, req InventoryLedgerItem) (InventoryLedgerItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	store, ok := resolveStoreForWriteLocked(s.stores, user, req.StoreID)
	if !ok {
		return InventoryLedgerItem{}, errUnauthorizedStore
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.TrimSpace(req.Category)
	}
	if name == "" {
		name = "库存商品"
	}
	styleNo := strings.TrimSpace(req.StyleNo)
	if styleNo == "" {
		styleNo = fmt.Sprintf("STYLE-%s-%03d", store.Code, s.inventorySeq+1)
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "in_stock"
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "manual"
	}
	s.inventorySeq++
	item := InventoryLedgerItem{
		ID:         fmt.Sprintf("inv-%06d", s.inventorySeq),
		OrgID:      user.OrgID,
		StoreID:    store.ID,
		StoreName:  store.Name,
		StyleNo:    styleNo,
		Name:       name,
		Category:   strings.TrimSpace(req.Category),
		Purity:     strings.TrimSpace(req.Purity),
		PieceCount: maxInt(req.PieceCount, 1),
		WeightGram: round2(req.WeightGram),
		CostAmount: round2(req.CostAmount),
		Status:     status,
		Source:     source,
		Remark:     strings.TrimSpace(req.Remark),
		CreatedBy:  user.DisplayName,
		CreatedAt:  time.Now(),
	}
	s.inventoryLedger = append(s.inventoryLedger, item)
	s.persistInventoryLedgerLocked()
	return item, nil
}

func (s *MockStore) updateInventoryLedgerItem(user UserAccount, itemID string, req InventoryLedgerItem) (InventoryLedgerItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, storeSet := visibleStoresForUserLocked(s.stores, user)
	for idx, item := range s.inventoryLedger {
		if item.ID != itemID {
			continue
		}
		if !ledgerOrgVisibleToUserLocked(item.OrgID, user) || !storeVisibleToUserLocked(item.StoreID, user, storeSet) {
			return InventoryLedgerItem{}, errUnauthorizedStore
		}
		store, ok := resolveStoreForWriteLocked(s.stores, user, firstNonEmpty(strings.TrimSpace(req.StoreID), item.StoreID))
		if !ok {
			return InventoryLedgerItem{}, errUnauthorizedStore
		}
		item.StoreID = store.ID
		item.StoreName = store.Name
		item.StyleNo = strings.TrimSpace(firstNonEmpty(req.StyleNo, item.StyleNo))
		item.Name = strings.TrimSpace(firstNonEmpty(req.Name, item.Name))
		item.Category = strings.TrimSpace(req.Category)
		item.Purity = strings.TrimSpace(req.Purity)
		item.PieceCount = maxInt(req.PieceCount, 1)
		item.WeightGram = round2(req.WeightGram)
		item.CostAmount = round2(req.CostAmount)
		if value := strings.TrimSpace(req.Status); value != "" {
			item.Status = value
		}
		if value := strings.TrimSpace(req.Source); value != "" {
			item.Source = value
		}
		item.Remark = strings.TrimSpace(req.Remark)
		s.inventoryLedger[idx] = item
		s.persistInventoryLedgerLocked()
		return item, nil
	}
	return InventoryLedgerItem{}, errInventoryNotFound
}

func (s *MockStore) deleteInventoryLedgerItem(user UserAccount, itemID string) (InventoryLedgerItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, storeSet := visibleStoresForUserLocked(s.stores, user)
	for idx, item := range s.inventoryLedger {
		if item.ID != itemID {
			continue
		}
		if !ledgerOrgVisibleToUserLocked(item.OrgID, user) || !storeVisibleToUserLocked(item.StoreID, user, storeSet) {
			return InventoryLedgerItem{}, errUnauthorizedStore
		}
		item.Status = "deleted"
		item.Remark = strings.TrimSpace(firstNonEmpty(item.Remark, "后台删除"))
		s.inventoryLedger[idx] = item
		s.persistInventoryLedgerLocked()
		return item, nil
	}
	return InventoryLedgerItem{}, errInventoryNotFound
}

func (s *MockStore) listMaterialLedger(user UserAccount) MaterialLedgerSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	visibleStores, storeSet := visibleStoresForUserLocked(s.stores, user)
	now := time.Now()
	today := now.Format("2006-01-02")
	month := now.Format("2006-01")
	stats := make(map[string]MaterialPurityStat)
	summary := MaterialLedgerSummary{
		OrgID:             user.OrgID,
		VisibleStoreCount: len(visibleStores),
		Items:             make([]MaterialLedgerItem, 0, len(s.materialLedger)),
	}
	for _, item := range s.materialLedger {
		if !ledgerOrgVisibleToUserLocked(item.OrgID, user) || !storeVisibleToUserLocked(item.StoreID, user, storeSet) || item.Status == "deleted" {
			continue
		}
		summary.Items = append(summary.Items, item)
		if item.CreatedAt.Format("2006-01-02") == today {
			summary.TodayWeightGram += item.WeightGram
			summary.TodayAmount += item.Amount
		}
		if item.CreatedAt.Format("2006-01") == month {
			summary.MonthWeightGram += item.WeightGram
			summary.MonthAmount += item.Amount
		}
		if item.Status != "outbound" {
			summary.RemainingWeightGram += item.RemainingWeightGram
		}
		if item.Type == "pledge" || item.Type == "consignment" {
			summary.PledgeCount++
			summary.PledgeAmount += item.Amount
		}
		purity := item.Purity
		if purity == "" {
			purity = "未填写"
		}
		stat := stats[purity]
		stat.Purity = purity
		stat.Count++
		stat.WeightGram += item.WeightGram
		stat.Amount += item.Amount
		stats[purity] = stat
	}
	sort.Slice(summary.Items, func(i, j int) bool {
		return summary.Items[i].CreatedAt.After(summary.Items[j].CreatedAt)
	})
	for _, stat := range stats {
		stat.WeightGram = round2(stat.WeightGram)
		stat.Amount = round2(stat.Amount)
		summary.PurityStats = append(summary.PurityStats, stat)
	}
	sort.Slice(summary.PurityStats, func(i, j int) bool {
		return summary.PurityStats[i].WeightGram > summary.PurityStats[j].WeightGram
	})
	summary.TodayWeightGram = round2(summary.TodayWeightGram)
	summary.TodayAmount = round2(summary.TodayAmount)
	summary.MonthWeightGram = round2(summary.MonthWeightGram)
	summary.MonthAmount = round2(summary.MonthAmount)
	summary.RemainingWeightGram = round2(summary.RemainingWeightGram)
	summary.PledgeAmount = round2(summary.PledgeAmount)
	return summary
}

func (s *MockStore) createMaterialLedgerItem(user UserAccount, req MaterialLedgerItem) (MaterialLedgerItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	store, ok := resolveStoreForWriteLocked(s.stores, user, req.StoreID)
	if !ok {
		return MaterialLedgerItem{}, errUnauthorizedStore
	}
	itemType := strings.TrimSpace(req.Type)
	if itemType == "" {
		itemType = "recycle"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "in_stock"
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "manual"
	}
	remaining := req.RemainingWeightGram
	if remaining <= 0 {
		remaining = req.WeightGram
	}
	s.materialSeq++
	item := MaterialLedgerItem{
		ID:                  fmt.Sprintf("mat-%06d", s.materialSeq),
		OrgID:               user.OrgID,
		StoreID:             store.ID,
		StoreName:           store.Name,
		Type:                itemType,
		OrderNo:             strings.TrimSpace(req.OrderNo),
		CustomerName:        strings.TrimSpace(req.CustomerName),
		Category:            strings.TrimSpace(req.Category),
		Purity:              strings.TrimSpace(req.Purity),
		WeightGram:          round2(req.WeightGram),
		Amount:              round2(req.Amount),
		RemainingWeightGram: round2(remaining),
		Status:              status,
		DueDate:             strings.TrimSpace(req.DueDate),
		Remark:              strings.TrimSpace(req.Remark),
		Source:              source,
		CreatedBy:           user.DisplayName,
		CreatedAt:           time.Now(),
	}
	s.materialLedger = append(s.materialLedger, item)
	s.persistMaterialLedgerLocked()
	return item, nil
}

func (s *MockStore) outboundMaterialLedgerItem(user UserAccount, itemID, remark string) (MaterialLedgerItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, storeSet := visibleStoresForUserLocked(s.stores, user)
	for idx, item := range s.materialLedger {
		if item.ID != itemID {
			continue
		}
		if !ledgerOrgVisibleToUserLocked(item.OrgID, user) || !storeVisibleToUserLocked(item.StoreID, user, storeSet) {
			return MaterialLedgerItem{}, errUnauthorizedStore
		}
		now := time.Now()
		item.Status = "outbound"
		item.RemainingWeightGram = 0
		item.OutboundAt = &now
		item.OutboundBy = user.DisplayName
		item.OutboundRemark = strings.TrimSpace(remark)
		s.materialLedger[idx] = item
		s.persistMaterialLedgerLocked()
		return item, nil
	}
	return MaterialLedgerItem{}, errMaterialNotFound
}

func (s *MockStore) updateMaterialLedgerItem(user UserAccount, itemID string, req MaterialLedgerItem) (MaterialLedgerItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, storeSet := visibleStoresForUserLocked(s.stores, user)
	for idx, item := range s.materialLedger {
		if item.ID != itemID {
			continue
		}
		if !ledgerOrgVisibleToUserLocked(item.OrgID, user) || !storeVisibleToUserLocked(item.StoreID, user, storeSet) {
			return MaterialLedgerItem{}, errUnauthorizedStore
		}
		store, ok := resolveStoreForWriteLocked(s.stores, user, firstNonEmpty(strings.TrimSpace(req.StoreID), item.StoreID))
		if !ok {
			return MaterialLedgerItem{}, errUnauthorizedStore
		}
		item.StoreID = store.ID
		item.StoreName = store.Name
		if value := strings.TrimSpace(req.Type); value != "" {
			item.Type = value
		}
		item.OrderNo = strings.TrimSpace(firstNonEmpty(req.OrderNo, item.OrderNo))
		item.CustomerName = strings.TrimSpace(req.CustomerName)
		item.Category = strings.TrimSpace(req.Category)
		item.Purity = strings.TrimSpace(req.Purity)
		item.WeightGram = round2(req.WeightGram)
		item.Amount = round2(req.Amount)
		if value := strings.TrimSpace(req.Status); value != "" {
			item.Status = value
		}
		remaining := req.RemainingWeightGram
		if remaining <= 0 && item.Status != "outbound" && item.Status != "已出库" {
			remaining = req.WeightGram
		}
		item.RemainingWeightGram = round2(remaining)
		item.DueDate = strings.TrimSpace(req.DueDate)
		item.Remark = strings.TrimSpace(req.Remark)
		if value := strings.TrimSpace(req.Source); value != "" {
			item.Source = value
		}
		s.materialLedger[idx] = item
		s.persistMaterialLedgerLocked()
		return item, nil
	}
	return MaterialLedgerItem{}, errMaterialNotFound
}

func (s *MockStore) deleteMaterialLedgerItem(user UserAccount, itemID string) (MaterialLedgerItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, storeSet := visibleStoresForUserLocked(s.stores, user)
	for idx, item := range s.materialLedger {
		if item.ID != itemID {
			continue
		}
		if !ledgerOrgVisibleToUserLocked(item.OrgID, user) || !storeVisibleToUserLocked(item.StoreID, user, storeSet) {
			return MaterialLedgerItem{}, errUnauthorizedStore
		}
		item.Status = "deleted"
		item.RemainingWeightGram = 0
		item.OutboundRemark = strings.TrimSpace(firstNonEmpty(item.OutboundRemark, "后台删除"))
		s.materialLedger[idx] = item
		s.persistMaterialLedgerLocked()
		return item, nil
	}
	return MaterialLedgerItem{}, errMaterialNotFound
}

func (s *MockStore) bootstrapInventoryMaterialConfigs(ctx context.Context) error {
	if s.persistence == nil {
		s.inventorySeq = nextSequenceFromInventory(s.inventoryLedger)
		s.materialSeq = nextSequenceFromMaterials(s.materialLedger)
		return nil
	}

	var inventory []InventoryLedgerItem
	found, err := s.persistence.loadConfig(ctx, configKeyInventory, &inventory)
	if err != nil {
		return err
	}
	if found {
		s.inventoryLedger = inventory
	} else if err := s.persistence.saveConfig(ctx, configKeyInventory, s.inventoryLedger); err != nil {
		return err
	}

	var materials []MaterialLedgerItem
	found, err = s.persistence.loadConfig(ctx, configKeyMaterials, &materials)
	if err != nil {
		return err
	}
	if found {
		s.materialLedger = materials
	} else if err := s.persistence.saveConfig(ctx, configKeyMaterials, s.materialLedger); err != nil {
		return err
	}
	s.inventorySeq = nextSequenceFromInventory(s.inventoryLedger)
	s.materialSeq = nextSequenceFromMaterials(s.materialLedger)
	return nil
}

func (s *MockStore) persistInventoryLedgerLocked() {
	if s.persistence != nil {
		_ = s.persistence.saveConfig(context.Background(), configKeyInventory, s.inventoryLedger)
	}
}

func (s *MockStore) persistMaterialLedgerLocked() {
	if s.persistence != nil {
		_ = s.persistence.saveConfig(context.Background(), configKeyMaterials, s.materialLedger)
	}
}

func storeVisibleToUserLocked(storeID string, user UserAccount, storeSet map[string]struct{}) bool {
	if hasAllStoresScope(user.DataScope) {
		return true
	}
	_, ok := storeSet[storeID]
	return ok
}

func ledgerOrgVisibleToUserLocked(itemOrgID string, user UserAccount) bool {
	if hasAllStoresScope(user.DataScope) {
		return true
	}
	itemOrgID = strings.TrimSpace(itemOrgID)
	if itemOrgID == "" {
		return true
	}
	return itemOrgID == strings.TrimSpace(user.OrgID)
}

func resolveStoreForWriteLocked(stores []StoreInfo, user UserAccount, storeID string) (StoreInfo, bool) {
	visible, storeSet := visibleStoresForUserLocked(stores, user)
	if storeID != "" {
		if !storeVisibleToUserLocked(storeID, user, storeSet) {
			return StoreInfo{}, false
		}
		for _, store := range stores {
			if store.ID == storeID {
				return store, true
			}
		}
		return StoreInfo{}, false
	}
	for _, store := range visible {
		if store.IsDefault {
			return store, true
		}
	}
	if len(visible) > 0 {
		return visible[0], true
	}
	return StoreInfo{}, false
}

func nextSequenceFromInventory(items []InventoryLedgerItem) int {
	maxSeq := 0
	for _, item := range items {
		var seq int
		if _, err := fmt.Sscanf(item.ID, "inv-%06d", &seq); err == nil && seq > maxSeq {
			maxSeq = seq
		}
	}
	return maxSeq
}

func nextSequenceFromMaterials(items []MaterialLedgerItem) int {
	maxSeq := 0
	for _, item := range items {
		var seq int
		if _, err := fmt.Sscanf(item.ID, "mat-%06d", &seq); err == nil && seq > maxSeq {
			maxSeq = seq
		}
	}
	return maxSeq
}

func maxInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
