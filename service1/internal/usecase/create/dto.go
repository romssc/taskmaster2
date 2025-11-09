package create

import "taskmaster2/internal/domain"

type Input struct {
	domain.Task
}

type Output struct {
	ID int
}
