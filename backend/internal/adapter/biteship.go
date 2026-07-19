package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/service"
)

// configReader provides access to system configuration.
type configReader interface {
	ListSystemConfig(context.Context) ([]repository.SystemConfigRow, error)
}

type BiteshipClient struct {
	repo       configReader
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewBiteshipClient creates a real BiteshipClient.
func NewBiteshipClient(repo configReader, apiKey, baseURL string, httpClient *http.Client) *BiteshipClient {
	return &BiteshipClient{
		repo:       repo,
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// GetRates calls Biteship's Rates API with the given request.
// It reads the origin postal code from system_config and returns the parsed rates
// or an error if any step fails.
func (c *BiteshipClient) GetRates(ctx context.Context, req service.ShippingQuoteRequest) ([]service.CourierRate, error) {
	// Read origin postal code from system_config
	originPostalCode, err := c.getOriginPostalCode(ctx)
	if err != nil {
		return nil, err
	}

	// Build request to Biteship
	biteshipReq := c.buildBiteshipRequest(originPostalCode, req)

	// Make HTTP request
	body, err := json.Marshal(biteshipReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Biteship request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/rates/couriers", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("authorization", c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Biteship API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Biteship API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var biteshipResp biteshipRatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&biteshipResp); err != nil {
		return nil, fmt.Errorf("failed to parse Biteship response: %w", err)
	}

	// Convert to service.CourierRate
	rates := c.parsePricing(biteshipResp.Pricing)
	return rates, nil
}

// getOriginPostalCode reads app_kode_pos from system_config.
func (c *BiteshipClient) getOriginPostalCode(ctx context.Context) (string, error) {
	rows, err := c.repo.ListSystemConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read system_config: %w", err)
	}

	for _, row := range rows {
		if row.Key == "app_kode_pos" && row.Value != "" {
			return row.Value, nil
		}
	}

	return "", fmt.Errorf("app_kode_pos not configured in system_config")
}

// buildBiteshipRequest constructs the request body for Biteship's Rates API.
func (c *BiteshipClient) buildBiteshipRequest(originPostalCode string, req service.ShippingQuoteRequest) map[string]interface{} {
	return map[string]interface{}{
		"origin_postal_code":      originPostalCode,
		"destination_postal_code": req.DestinationPostalCode,
		"couriers":                "anteraja,jne,sicepat,tiki",
		"items": []map[string]interface{}{
			{
				"name":     "items",
				"value":    1, // Default value
				"quantity": 1,
				"weight":   req.WeightGrams,
			},
		},
	}
}

// parsePricing converts Biteship pricing array to service.CourierRate slice.
func (c *BiteshipClient) parsePricing(pricing []biteshipPricingItem) []service.CourierRate {
	var rates []service.CourierRate
	for _, item := range pricing {
		estimatedDays := 0
		if item.Duration != "" {
			// Try to parse duration as integer
			if days, err := strconv.Atoi(item.Duration); err == nil {
				estimatedDays = days
			}
		}

		rate := service.CourierRate{
			Courier:       item.CourierName,
			Service:       item.CourierServiceName,
			EstimatedDays: estimatedDays,
			Price:         int64(item.Price),
		}
		rates = append(rates, rate)
	}
	return rates
}

// biteshipRatesResponse represents the response from Biteship Rates API.
type biteshipRatesResponse struct {
	Pricing []biteshipPricingItem `json:"pricing"`
}

// biteshipPricingItem represents a single rate option from Biteship.
type biteshipPricingItem struct {
	CourierName        string `json:"courier_name"`
	CourierServiceName string `json:"courier_service_name"`
	Price              int64  `json:"price"`
	Duration           string `json:"duration"`
}
