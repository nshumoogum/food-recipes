package helpers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ONSdigital/log.go/v2/log"
	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/pkg/errors"
)

// CalculateOffset returns a valid offset value to skip a list of items returned from query
func CalculateOffset(ctx context.Context, requestedOffset string) (offset int, err error) {
	errorValues := map[string]string{"offset": requestedOffset}

	if requestedOffset != "" {
		offset, err = strconv.Atoi(requestedOffset)
		if err != nil {
			log.Error(ctx, "invalid offset parameter", errors.WithMessage(err, errs.ErrOffsetWrongType.Error()), log.Data{"requested_offset": requestedOffset})
			return 0, errs.New(errs.ErrOffsetWrongType, http.StatusBadRequest, errorValues)
		}

		if offset < 0 {
			log.Error(ctx, "invalid offset parameter", errors.WithMessage(err, errs.ErrNegativeLimit.Error()), log.Data{"requested_offset": requestedOffset})
			return 0, errs.New(errs.ErrNegativeOffset, http.StatusBadRequest, errorValues)
		}
	}

	return
}
