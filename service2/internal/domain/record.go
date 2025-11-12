package domain

type Record struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	Status    Status `json:"status"`
}
