package create

import "errors"

var (
	ErrEmptyTitle       = errors.New("handler: client sent a request with an empty task title")
	ErrUnmarshalingBody = errors.New("handler: failed to unmarshal body")
	ErrReadingBody      = errors.New("handler: failed to read a body")
)

var (
	ErrAlreadyExists   = errors.New("usecase: task already exists in the database")
	ErrDatabaseFailure = errors.New("usecase: database failed")
)
