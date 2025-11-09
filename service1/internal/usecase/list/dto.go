package list

import "taskmaster2/service1/internal/domain"

type Input struct{}

type Output struct {
	Tasks []domain.Task `json:"tasks"`
}
