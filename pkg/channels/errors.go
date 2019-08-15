package channels

import "errors"

var (
	// ErrorNotFound is returned when a record doesn't exist
	ErrorNotFound = errors.New("record not found")
	// ErrorNotUnique is returned when attempting to create a record that violates a uniqueness constraint
	ErrorNotUnique = errors.New("not unique")
	// ErrorInvalidHandle is returned when a handle does not match the expected expression
	ErrorInvalidHandle = errors.New("invalid handle")
	// ErrorInvalidEmail is returned when an email address does not match the expected expression
	ErrorInvalidEmail = errors.New("invalid email")
	// ErrorHandleAlreadySet is returned when a user tries to update their handle but they already have a handle set (TODO: remove this once logic exists to let users update handles appropriately)
	ErrorHandleAlreadySet = errors.New("handle already set")
	// ErrorUnauthorized is returned when attempting to perform an action the user is not authorized to do
	ErrorUnauthorized = errors.New("unauthorized")
	// ErrorsInvalidInput is returned when the input is invalid
	ErrorsInvalidInput = errors.New("invalid input")
	// ErrorStripeIssue is returned when a stripe service returns an error
	ErrorStripeIssue = errors.New("error with stripe request")
)
