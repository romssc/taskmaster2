package domain

import "time"

type Event struct {
	ID        int       `json:"id"`
	Action    Action    `json:"action"`
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
