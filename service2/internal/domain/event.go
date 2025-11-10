package domain

type Event struct {
	ID     int    `json:"id"`
	Action Action `json:"action"`
}
