package helpers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ONSdigital/log.go/v2/log"
	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/pkg/errors"
)

// CalculateLimit returns a valid limit for the number of items to be returned from query
func CalculateLimit(ctx context.Context, defaultLimit, maximumLimit int, requestedLimit string) (int, error) {
	if requestedLimit == "" {
		return defaultLimit, nil
	}

	errorValues := map[string]string{"limit": requestedLimit}

	requestedLimitNumber, err := strconv.Atoi(requestedLimit)
	if err != nil {
		log.Error(ctx, "invalid limit value", errors.WithMessage(err, errs.ErrLimitWrongType.Error()), log.Data{"requested_limit": requestedLimitNumber})
		return 0, errs.New(errs.ErrLimitWrongType, http.StatusBadRequest, errorValues)
	}

	if requestedLimitNumber < 0 {
		log.Error(ctx, "invalid limit value", errs.ErrNegativeLimit, log.Data{"requested_limit": requestedLimitNumber})
		return 0, errs.New(errs.ErrNegativeLimit, http.StatusBadRequest, errorValues)
	}

	if requestedLimitNumber > maximumLimit {
		err := fmt.Errorf("limit exceeded maximum value, limit cannot be greater than [%d]", maximumLimit)

		log.Error(ctx, "invalid limit value", err, log.Data{"requested_limit": requestedLimitNumber})
		return 0, errs.New(err, http.StatusBadRequest, errorValues)
	}

	return requestedLimitNumber, nil
}
