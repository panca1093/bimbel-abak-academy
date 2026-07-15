package model

import "time"

type Product struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
	Status      string `json:"status"`
	WeightGrams int    `json:"weight_grams"`
	ImageURL    string `json:"image_url"`
	CourseIDs   []string `json:"course_ids"`
	ExamIDs     []string `json:"exam_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
