package domain

type Event struct {
	Action Action `json:"action"`
	Record Record `json:"record"`
}
