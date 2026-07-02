package models

import "time"

type Item struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	SerialNumber string    `json:"serial_number"`
	ImageURL     string    `json:"image_url"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
