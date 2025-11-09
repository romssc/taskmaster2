package list

import "taskmaster2/internal/domain"

type Input struct{}

type Output struct {
	Tasks []domain.Task `json:"tasks"`
}
