package domain

import "time"

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

var (
	StatusCreated = "created"
)
