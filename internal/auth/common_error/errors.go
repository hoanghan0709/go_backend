package commonError

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidPassword    = errors.New("not found")
	ErrUserNotFound       = errors.New("not found")
	ErrNotFound           = errors.New("not found")
)
