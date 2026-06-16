package model

import "time"

type Product struct {
	ID        string
	Type      string
	Name      string
	Description string
	Price     int64
	Stock     int
	Status    string
	WeightGrams int
	ImageURL  string
	CourseIDs []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
