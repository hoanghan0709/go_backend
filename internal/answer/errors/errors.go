package errors

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateLabel = errors.New("answer label already exists in this question")
	ErrDuplicateOrder = errors.New("answer order already exists in this question")
)
