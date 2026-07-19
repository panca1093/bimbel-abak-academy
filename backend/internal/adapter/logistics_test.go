package adapter

import (
	"context"
	"testing"

	"akademi-bimbel/internal/service"
)

func TestNoopLogisticsClient_GetRates(t *testing.T) {
	client := &NoopLogisticsClient{}
	req := service.ShippingQuoteRequest{
		DestinationPostalCode: "12345",
		WeightGrams:           1000,
	}

	rates, err := client.GetRates(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rates) != 2 {
		t.Errorf("expected 2 rates, got %d", len(rates))
	}

	if rates[0].Courier != "JNE" {
		t.Errorf("expected first courier JNE, got %s", rates[0].Courier)
	}

	if rates[0].Service != "REG" {
		t.Errorf("expected first service REG, got %s", rates[0].Service)
	}

	if rates[0].EstimatedDays != 3 {
		t.Errorf("expected first EstimatedDays=3, got %d", rates[0].EstimatedDays)
	}

	if rates[0].Price != 15000 {
		t.Errorf("expected first Price=15000, got %d", rates[0].Price)
	}
}
