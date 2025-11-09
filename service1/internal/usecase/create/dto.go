package create

import "taskmaster2/service1/internal/domain"

type Input struct {
	domain.Task
}

type Output struct {
	ID int
}
