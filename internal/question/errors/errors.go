package errors

import "errors"

var ErrNotFound = errors.New("not found")

var (
	ErrInvalidAnswers         = errors.New("question must have at least two answers")
	ErrDuplicateAnswerLabel   = errors.New("answer labels must be unique")
	ErrDuplicateAnswerContent = errors.New("answer contents must be unique")
	ErrDuplicateAnswerOrder   = errors.New("answer orders must be unique")
	ErrInvalidCorrectAnswer   = errors.New("question must have exactly one correct answer")
)
