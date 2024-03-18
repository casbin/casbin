package errors

import "errors"

var (
	ErrInvalidFilePath = errors.New("invalid file path, file path cannot be empty")
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}
