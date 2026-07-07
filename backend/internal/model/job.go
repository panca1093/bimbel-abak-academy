package model

import "time"

type Job struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	InputURL  *string   `json:"input_url"`
	ResultURL *string   `json:"result_url"`
	Error     *string   `json:"error"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
