package listid

import "errors"

var (
	ErrMalformedID = errors.New("handler: client sent a malformed id")
)

var (
	ErrNoRows          = errors.New("usecase: no records found")
	ErrDatabaseFailure = errors.New("usecase: database failed")
)
