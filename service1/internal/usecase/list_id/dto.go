package listid

import "taskmaster2/service1/internal/domain"

type Input struct {
	ID int
}

type Output struct {
	Task domain.Task `json:"task"`
}
