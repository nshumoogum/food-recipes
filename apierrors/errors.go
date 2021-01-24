package apierrors

import (
	"errors"
)

// New returns an error that formats as the given text.
func New(err error, status int, values map[string]string) error {
	return &ErrorObject{
		Code:    status,
		Keys:    values,
		Message: err.Error(),
	}
}

// ErrorObject is a trivial implementation of error.
type ErrorObject struct {
	Code    int
	Keys    map[string]string
	Message string
}

func (e *ErrorObject) Error() string {
	return e.Message
}

// Status represents the status code to return from error
func (e *ErrorObject) Status() int {
	return e.Code
}

// Values represents a list of key value pairs to return from error
func (e *ErrorObject) Values() map[string]string {
	return e.Keys
}

// A list of error messages for Dataset API
var (
	ErrLimitWrongType  = errors.New("limit value needs to be a number")
	ErrNegativeLimit   = errors.New("limit needs to be a positive number, limit cannot be lower than 0")
	ErrOffsetWrongType = errors.New("offset value needs to be a number")
	ErrNegativeOffset  = errors.New("offset needs to be a positive number, offset cannot be lower than 0")

	ErrRecipeNotFound      = errors.New("recipe not found")
	ErrRecipeAlreadyExists = errors.New("recipe already exists, use different title")

	ErrMissingFields       = errors.New("missing mandatory fields")
	ErrInvalidUnits        = errors.New("invalid units for ingredient")
	ErrInvalidPortionSize  = errors.New("invalid portion size, cannot be less than 1")
	ErrUnableToChangeTitle = errors.New("not allowed to change the existing title for recipe")

	ErrUnableToParseJSON   = errors.New("failed to parse json body")
	ErrUnableToReadMessage = errors.New("failed to read message body")
	ErrInternalServer      = errors.New("internal server error")
)
