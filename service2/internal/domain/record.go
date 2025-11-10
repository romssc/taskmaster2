package domain

import "time"

type Record struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Status    Status    `json:"status"`
}
