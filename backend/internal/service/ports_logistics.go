package service

import "context"

type ShippingQuoteRequest struct {
	DestinationZip string
	WeightGrams    int
}

type CourierRate struct {
	Courier       string
	Service       string
	EstimatedDays int
	Price         int64
}

type LogisticsClient interface {
	GetRates(ctx context.Context, req ShippingQuoteRequest) ([]CourierRate, error)
}

type NoopLogisticsClient struct{}

func (n *NoopLogisticsClient) GetRates(ctx context.Context, req ShippingQuoteRequest) ([]CourierRate, error) {
	return []CourierRate{
		{Courier: "JNE", Service: "REG", EstimatedDays: 3, Price: 15000},
		{Courier: "TIKI", Service: "ONS", EstimatedDays: 1, Price: 25000},
	}, nil
}
