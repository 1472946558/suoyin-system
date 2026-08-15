/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: product_import.go
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-06-17
 */

package app

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var productImportHeaders = []string{"商品名称", "SKU", "分类", "克重", "销售价格", "库存数量", "门店ID或名称"}

func (s *MockStore) buildProductImportTemplate(user UserAccount, storeID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cleanStoreID := strings.TrimSpace(storeID)
	if cleanStoreID != "" && !s.canAccessStore(user, cleanStoreID) {
		return nil, errUnauthorizedStore
	}
	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	for index, header := range productImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}
	sampleStoreID := ""
	if cleanStoreID != "" {
		sampleStoreID = cleanStoreID
	} else {
		visibleStores := s.adminStoreIDsForUser(user)
		if len(visibleStores) > 0 {
			sampleStoreID = visibleStores[0]
		}
	}
	sample := []interface{}{"足金手镯标准款", "GJG-SZ-001", "金饰", 12.68, 786, 9, sampleStoreID}
	for index, value := range sample {
		cell, _ := excelize.CoordinatesToCellName(index+1, 2)
		_ = file.SetCellValue(sheet, cell, value)
	}
	_ = file.SetColWidth(sheet, "A", "A", 22)
	_ = file.SetColWidth(sheet, "B", "B", 18)
	_ = file.SetColWidth(sheet, "C", "G", 14)
	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *MockStore) importAdminProducts(user UserAccount, storeID string, fileName string, reader io.Reader) (AdminProductImportResult, error) {
	cleanStoreID := strings.TrimSpace(storeID)
	if cleanStoreID == "" || !s.canAccessStore(user, cleanStoreID) {
		return AdminProductImportResult{}, errUnauthorizedStore
	}

	file, err := excelize.OpenReader(reader)
	if err != nil {
		return AdminProductImportResult{}, fmt.Errorf("open xlsx: %w", err)
	}
	defer func() { _ = file.Close() }()

	sheet := file.GetSheetName(0)
	rows, err := file.GetRows(sheet)
	if err != nil {
		return AdminProductImportResult{}, fmt.Errorf("read xlsx: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	store, ok := s.getStoreByIDLocked(cleanStoreID)
	if !ok {
		return AdminProductImportResult{}, errUnauthorizedStore
	}

	result := AdminProductImportResult{
		ImportID:   fmt.Sprintf("import-%d", time.Now().UnixNano()),
		StoreID:    store.ID,
		StoreName:  store.Name,
		Failures:   []AdminProductImportFailure{},
		ImportedAt: time.Now().Format("2006-01-02 15:04"),
	}

	for rowIndex, row := range rows {
		if rowIndex == 0 {
			continue
		}
		lineNo := rowIndex + 1
		if rowIsBlank(row) {
			continue
		}
		product, failure := s.parseProductImportRow(row, lineNo, store)
		if failure != nil {
			result.Failures = append(result.Failures, *failure)
			continue
		}
		s.upsertImportedProductLocked(product)
		result.SuccessCount++
	}
	result.FailureCount = len(result.Failures)

	log := AdminProductImportLog{
		ID:           result.ImportID,
		StoreID:      result.StoreID,
		StoreName:    result.StoreName,
		FileName:     strings.TrimSpace(fileName),
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
		Failures:     append([]AdminProductImportFailure(nil), result.Failures...),
		ImportedBy:   user.DisplayName,
		ImportedAt:   result.ImportedAt,
	}
	s.productImportLogs = append([]AdminProductImportLog{log}, s.productImportLogs...)
	if len(s.productImportLogs) > 50 {
		s.productImportLogs = s.productImportLogs[:50]
	}
	if s.persistence != nil {
		_ = s.persistProductsLocked()
		_ = s.persistProductImportLogsLocked()
	}
	s.appendAuditLogLocked("商品导入", "Excel 导入商品", user.DisplayName, "success", "medium", fmt.Sprintf("%s 导入成功 %d 条，失败 %d 条", store.Name, result.SuccessCount, result.FailureCount))
	return result, nil
}

func rowIsBlank(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func (s *MockStore) parseProductImportRow(row []string, lineNo int, store StoreInfo) (CatalogProduct, *AdminProductImportFailure) {
	value := func(index int) string {
		if index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}
	name := value(0)
	sku := value(1)
	category := value(2)
	gramText := value(3)
	priceText := value(4)
	inventoryText := value(5)
	rowStore := value(6)
	if name == "" {
		return CatalogProduct{}, &AdminProductImportFailure{Row: lineNo, Reason: "商品名称不能为空"}
	}
	if sku == "" {
		return CatalogProduct{}, &AdminProductImportFailure{Row: lineNo, Reason: "SKU不能为空"}
	}
	if category == "" {
		return CatalogProduct{}, &AdminProductImportFailure{Row: lineNo, Reason: "分类不能为空"}
	}
	if rowStore != "" && !matchesImportStore(rowStore, store) {
		return CatalogProduct{}, &AdminProductImportFailure{Row: lineNo, Reason: fmt.Sprintf("门店应为 %s（也可填 %s）", store.Name, store.ID)}
	}
	gramWeight, err := strconv.ParseFloat(gramText, 64)
	if err != nil || gramWeight < 0 {
		return CatalogProduct{}, &AdminProductImportFailure{Row: lineNo, Reason: "克重必须是非负数字"}
	}
	price, err := strconv.ParseFloat(priceText, 64)
	if err != nil || price < 0 {
		return CatalogProduct{}, &AdminProductImportFailure{Row: lineNo, Reason: "销售价格必须是非负数字"}
	}
	inventory, err := strconv.Atoi(inventoryText)
	if err != nil || inventory < 0 {
		return CatalogProduct{}, &AdminProductImportFailure{Row: lineNo, Reason: "库存数量必须是非负整数"}
	}
	return CatalogProduct{
		ID:               fmt.Sprintf("product-import-%d-%s", time.Now().UnixNano(), sanitizeObjectSegment(sku)),
		OrgID:            store.OrgID,
		Name:             name,
		SKU:              sku,
		Category:         category,
		CategoryTab:      category,
		ImageURL:         "/assets/ui/document.png",
		Purity:           "足金999",
		RetailPrice:      price,
		GramWeight:       gramWeight,
		Status:           "active",
		Inventory:        inventory,
		StockStatus:      "normal",
		StoreIDs:         []string{store.ID},
		Stores:           []string{store.Name},
		Tags:             []string{"导入"},
		RecommendedScene: "门店导入商品",
		QuoteLeadTime:    "按门店库存销售",
	}, nil
}

func matchesImportStore(value string, store StoreInfo) bool {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return true
	}
	return clean == store.ID || clean == store.Name || strings.EqualFold(clean, store.Code)
}

func (s *MockStore) upsertImportedProductLocked(product CatalogProduct) {
	for index, item := range s.catalogProducts {
		if strings.EqualFold(strings.TrimSpace(item.SKU), strings.TrimSpace(product.SKU)) && containsID(item.StoreIDs, product.StoreIDs[0]) {
			product.ID = item.ID
			product.Tags = cleanUniqueStrings(append(item.Tags, product.Tags...))
			s.catalogProducts[index] = product
			return
		}
	}
	s.catalogProducts = append([]CatalogProduct{product}, s.catalogProducts...)
}
