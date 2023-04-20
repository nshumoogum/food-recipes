package models

import (
	"errors"
	"strconv"
)

// ErrorMaximumOffsetReached creates a unique error
func ErrorMaximumOffsetReached(m int) error {
	err := errors.New("the maximum offset has been reached, the offset cannot be more than " + strconv.Itoa(m))
	return err
}

// PageVariables are the necessary fields to determine paging
type PageVariables struct {
	DefaultMaxResults int
	Limit             int
	Offset            int
}

// ValidatePage represents a model for validating combination of offset and limit
func ValidatePage(page PageVariables) []*ErrorObject {
	var errorObjects []*ErrorObject

	if page.Offset >= page.DefaultMaxResults {
		pagingErrorValue := make(map[string]string)
		pagingErrorValue["offset"] = strconv.Itoa(page.Offset)
		errorObjects = append(errorObjects, &ErrorObject{Error: ErrorMaximumOffsetReached(page.DefaultMaxResults).Error(), ErrorValues: pagingErrorValue})
	}

	if errorObjects != nil {
		return errorObjects
	}

	if page.Offset+page.Limit > page.DefaultMaxResults {
		page.Limit = page.DefaultMaxResults - page.Offset
	}

	return nil
}
