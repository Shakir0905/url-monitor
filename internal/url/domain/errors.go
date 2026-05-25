package domain

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrInvalidURL       = errors.New("invalid url")
	ErrInvalidInterval  = errors.New("invalid check interval")
	ErrPermissionDenied = errors.New("permission denied")
)
