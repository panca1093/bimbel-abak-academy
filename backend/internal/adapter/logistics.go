package adapter

import (
	"context"

	"akademi-bimbel/internal/service"
)

type NoopLogisticsClient struct{}

func (n *NoopLogisticsClient) GetRates(ctx context.Context, req service.ShippingQuoteRequest) ([]service.CourierRate, error) {
	return []service.CourierRate{
		{Courier: "JNE", Service: "REG", EstimatedDays: 3, Price: 15000},
		{Courier: "TIKI", Service: "ONS", EstimatedDays: 1, Price: 25000},
	}, nil
}
