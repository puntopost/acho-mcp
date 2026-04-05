package persistence

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}
