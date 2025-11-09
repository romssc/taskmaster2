package storage

import "errors"

var (
	ErrOpeningConnection  = errors.New("storage: failed to open connection")
	ErrPinging            = errors.New("storage: failed to ping")
	ErrClosingConnection  = errors.New("storage: failed to close connection")
	ErrPreparingStatement = errors.New("storage: failed to prepare a statement")
	ErrExecutingStatement = errors.New("storage: failed to execute a statement")
	ErrAlreadyExists      = errors.New("storage: failed due to record already existing")
	ErrScanningRows       = errors.New("storage: failed to scan rows")
	ErrIteratingRows      = errors.New("storage: failed while iterating rows")
	ErrNoRows             = errors.New("storage: failed due to no records available")
)
