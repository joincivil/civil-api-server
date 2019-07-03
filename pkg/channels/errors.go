package channels

import "errors"

var (
	// ErrorNotFound is returned when a record doesn't exist
	ErrorNotFound = errors.New("record not found")
	// ErrorNotUnique is returned when attempting to create a record that violates a uniqueness constraint
	ErrorNotUnique = errors.New("not unique")
	// ErrorInvalidHandle is returned when a handle does not match the expected expression
	ErrorInvalidHandle = errors.New("invalid handle")
)
