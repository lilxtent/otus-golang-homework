package storage

import "errors"

var (
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventNotFound      = errors.New("event not found")
	ErrDateBusy           = errors.New("date is busy")
)
