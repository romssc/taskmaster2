package domain

import "net/http"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GENERAL ERRORS
var (
	ErrMethodNotAllowed   = Error{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}
	ErrMalformedBody      = Error{Code: http.StatusBadRequest, Message: "malformed body"}
	ErrInternal           = Error{Code: http.StatusInternalServerError, Message: "internal error"}
	ErrMalformedPathValue = Error{Code: http.StatusBadRequest, Message: "malformed path value"}
)

// SITUATIONAL ERRORS
var (
	ErrEmptyTitle    = Error{Code: http.StatusBadRequest, Message: "task's title can't be empty"}
	ErrAlreadyExists = Error{Code: http.StatusConflict, Message: "task already exists"}
	ErrNotFound      = Error{Code: http.StatusNotFound, Message: "no tasks found"}
)
