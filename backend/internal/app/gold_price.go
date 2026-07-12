/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: gold_price.go
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-06-06
 */

package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const troyOunceGram = 31.1034768

type goldSpotAPIResponse struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Currency  string  `json:"currency"`
	Price     float64 `json:"price"`
	UpdatedAt string  `json:"updatedAt"`
}

type goldFXAPIResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

func (a *App) goldReferencePriceSnapshot(ctx context.Context) GoldReferencePriceSnapshot {
	if snapshot, ok := a.cachedLiveGoldReferencePriceSnapshot(ctx); ok {
		return snapshot
	}
	return a.configuredGoldReferencePriceSnapshot()
}

func (a *App) cachedLiveGoldReferencePriceSnapshot(ctx context.Context) (GoldReferencePriceSnapshot, bool) {
	if !a.Config.GoldPriceLiveEnabled {
		return GoldReferencePriceSnapshot{}, false
	}

	a.goldReferenceMu.Lock()
	defer a.goldReferenceMu.Unlock()

	ttl := time.Duration(a.Config.GoldPriceCacheTTL) * time.Second
	if ttl <= 0 {
		ttl = time.Minute
	}
	if a.goldReferenceSnapshot.Source != "" && time.Since(a.goldReferenceFetchedAt) < ttl {
		return a.goldReferenceSnapshot, true
	}

	snapshot, err := a.fetchLiveGoldReferencePriceSnapshot(ctx)
	if err != nil {
		if a.goldReferenceSnapshot.Source != "" {
			return a.goldReferenceSnapshot, true
		}
		return GoldReferencePriceSnapshot{}, false
	}
	a.goldReferenceSnapshot = snapshot
	a.goldReferenceFetchedAt = time.Now()
	return snapshot, true
}

func (a *App) fetchLiveGoldReferencePriceSnapshot(ctx context.Context) (GoldReferencePriceSnapshot, error) {
	timeout := time.Duration(a.Config.GoldPriceHTTPTimeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	spot, err := fetchGoldSpot(ctx, client, a.Config.GoldPriceAPIURL)
	if err != nil {
		return GoldReferencePriceSnapshot{}, err
	}
	fx, fxDate, err := fetchUSDToCNY(ctx, client, a.Config.GoldFXAPIURL)
	if err != nil {
		return GoldReferencePriceSnapshot{}, err
	}

	basePrice := round2(spot.Price * fx / troyOunceGram)
	updatedAt := strings.TrimSpace(spot.UpdatedAt)
	if updatedAt == "" && fxDate != "" {
		updatedAt = fxDate
	}
	if updatedAt == "" {
		updatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	return GoldReferencePriceSnapshot{
		Source:          "live_xau_usd_fx",
		SourceText:      "实时国际金价",
		BaseCNYPerGram:  basePrice,
		XAUUSD:          round2(spot.Price),
		USDCNY:          round2(fx),
		UpdatedAt:       updatedAt,
		ReferenceNote:   "按实时 XAU/USD 国际金价和 USD/CNY 汇率换算人民币/克后，再按成色折算；最终回收金额仍以门店复秤和确认单为准。",
		ReferencePrices: buildGoldReferencePrices(basePrice),
	}, nil
}

func fetchGoldSpot(ctx context.Context, client *http.Client, endpoint string) (goldSpotAPIResponse, error) {
	var spot goldSpotAPIResponse
	if strings.TrimSpace(endpoint) == "" {
		return spot, errors.New("gold price API URL is empty")
	}
	if err := getJSON(ctx, client, endpoint, &spot); err != nil {
		return spot, err
	}
	if spot.Price <= 0 {
		return spot, fmt.Errorf("gold spot response missing positive price")
	}
	return spot, nil
}

func fetchUSDToCNY(ctx context.Context, client *http.Client, endpoint string) (float64, string, error) {
	var payload goldFXAPIResponse
	if strings.TrimSpace(endpoint) == "" {
		return 0, "", errors.New("gold FX API URL is empty")
	}
	if err := getJSON(ctx, client, endpoint, &payload); err != nil {
		return 0, "", err
	}
	rate := payload.Rates["CNY"]
	if rate <= 0 {
		return 0, payload.Date, fmt.Errorf("FX response missing positive CNY rate")
	}
	return rate, payload.Date, nil
}

func getJSON(ctx context.Context, client *http.Client, endpoint string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GET %s returned HTTP %d", endpoint, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (a *App) configuredGoldReferencePriceSnapshot() GoldReferencePriceSnapshot {
	basePrice, source := a.configuredGoldBaseCNYPerGram()
	sourceText := goldPriceSourceText(source)
	updatedAt := strings.TrimSpace(a.Config.GoldPriceUpdatedAt)
	if updatedAt == "" {
		updatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	if basePrice <= 0 {
		return GoldReferencePriceSnapshot{
			Source:          "local_reference",
			SourceText:      sourceText,
			BaseCNYPerGram:  0,
			UpdatedAt:       updatedAt,
			ReferenceNote:   "实时国际金价不可用且未配置手动金价，当前返回系统本地兜底参考价。",
			ReferencePrices: fallbackGoldReferencePrices(),
		}
	}

	return GoldReferencePriceSnapshot{
		Source:          source,
		SourceText:      sourceText,
		BaseCNYPerGram:  basePrice,
		XAUUSD:          round2(a.Config.GoldPriceXAUUSD),
		USDCNY:          round2(a.Config.GoldPriceUSDCNY),
		UpdatedAt:       updatedAt,
		ReferenceNote:   "按配置的国际金价换算人民币/克后，再按成色折算；最终回收金额仍以门店复秤和确认单为准。",
		ReferencePrices: buildGoldReferencePrices(basePrice),
	}
}

func (a *App) configuredGoldBaseCNYPerGram() (float64, string) {
	if a.Config.GoldPriceCNYPerGram > 0 {
		return round2(a.Config.GoldPriceCNYPerGram), "configured_cny_per_gram"
	}
	if a.Config.GoldPriceXAUUSD > 0 && a.Config.GoldPriceUSDCNY > 0 {
		return round2(a.Config.GoldPriceXAUUSD * a.Config.GoldPriceUSDCNY / troyOunceGram), "configured_xau_usd_fx"
	}
	return 0, "local_reference"
}

func goldPriceSourceText(source string) string {
	switch source {
	case "live_xau_usd_fx":
		return "实时国际金价"
	case "configured_cny_per_gram":
		return "国际金价手动基准"
	case "configured_xau_usd_fx":
		return "手动 XAU/USD 汇率换算"
	default:
		return "本地兜底参考价"
	}
}

func buildGoldReferencePrices(basePrice float64) []GoldReferencePriceItem {
	items := []struct {
		purity string
		label  string
	}{
		{purity: "足金9999", label: "国际现货折算"},
		{purity: "足金999", label: "国际现货折算"},
		{purity: "22K", label: "按 22/24 折算"},
		{purity: "18K", label: "按 18/24 折算"},
		{purity: "14K", label: "按 14/24 折算"},
	}
	prices := make([]GoldReferencePriceItem, 0, len(items))
	for _, item := range items {
		prices = append(prices, GoldReferencePriceItem{
			Purity: item.purity,
			Price:  round2(basePrice * purityGoldRatio(item.purity)),
			Trend:  "实时",
			Label:  item.label,
		})
	}
	return prices
}

func fallbackGoldReferencePrices() []GoldReferencePriceItem {
	return []GoldReferencePriceItem{
		{Purity: "足金9999", Price: 748, Trend: "+6", Label: "大盘回收参考"},
		{Purity: "足金999", Price: 742, Trend: "+5", Label: "门店常用价"},
		{Purity: "22K", Price: 680, Trend: "+4", Label: "高成色 K 金"},
		{Purity: "18K", Price: 558, Trend: "+3", Label: "常规 K 金"},
		{Purity: "14K", Price: 436, Trend: "+2", Label: "低成色 K 金"},
	}
}

func purityGoldRatio(purity string) float64 {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(purity), " ", ""))
	switch normalized {
	case "足金9999", "AU9999", "9999":
		return 0.9999
	case "22K":
		return 22.0 / 24.0
	case "18K":
		return 18.0 / 24.0
	case "14K":
		return 14.0 / 24.0
	default:
		return 0.999
	}
}
