package domain

import "errors"

// Business errors for the Auth domain.
// These are returned by the service layer and translated to gRPC errors in the server layer.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrWeakPassword      = errors.New("password is too weak")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
)
