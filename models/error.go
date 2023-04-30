package models

import (
	"strings"

	errs "github.com/nshumoogum/food-recipes/apierrors"
)

// ErrorResponse builds a list of errors for an unsuccessful request
type ErrorResponse struct {
	Errors []*ErrorObject `json:"errors"`
}

// ErrorObject contains an error message and error values
type ErrorObject struct {
	Error       string            `json:"error"`
	ErrorValues map[string]string `json:"error_values,omitempty"`
}

// CreateErrorObject formulates an error object from an error
func CreateErrorObject(err error) *ErrorObject {
	return &ErrorObject{Error: err.Error(), ErrorValues: err.(*errs.ErrorObject).Values()}
}

// HandleValidationErrors works out a human friendly error for all Validation Errors
// Extend function for additional validation tags
func HandleValidationErrors(patchIndex, tag, field, value, param string) *ErrorObject {
	var err error

	switch tag {
	case "required":
		values := map[string]string{"[" + patchIndex + "]." + strings.ToLower(field): ""}
		err = errs.New(errs.ErrMissingFields, 0, values)
	case "requirevalueifopis":
		values := map[string]string{"[" + patchIndex + "]." + "value": ""}
		err = errs.New(errs.ErrMissingFields, 0, values)
	case "requirefromifopis":
		values := map[string]string{"[" + patchIndex + "]." + "from": ""}
		err = errs.New(errs.ErrMissingFields, 0, values)
	case "supportedops":
		values := map[string]string{"[" + patchIndex + "]." + strings.ToLower(field): value}
		err = errs.New(errs.ErrUnsupportedOperation, 0, values)
	case "oneof":
		values := map[string]string{"[" + patchIndex + "]." + strings.ToLower(field): value}
		err = errs.New(errs.ErrInvalidOperation, 0, values)
	case "nefield":
		values := map[string]string{"[" + patchIndex + "]." + strings.ToLower(field): value}

		params := strings.SplitAfter(param, " ")
		for i := range params {
			values["["+patchIndex+"]."+strings.ToLower(params[i])] = value
		}

		err = errs.New(errs.ErrPathAndFromFieldsCannotMatch, 0, values)
	}

	return CreateErrorObject(err)
}
