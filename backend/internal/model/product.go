package model

import "time"

type Product struct {
	ID            string
	Type          string
	Title         string
	Description   string
	Price         int64
	Stock         int
	Status        string
	IsVisible     bool
	WeightGrams   int
	CoverImageURL string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
