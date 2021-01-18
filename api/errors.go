package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/nshumoogum/food-recipes/models"
	"github.com/pkg/errors"
)

// ErrorResponse sets the structured error message in the http response body
func ErrorResponse(ctx context.Context, w http.ResponseWriter, status int, errorResponse *models.ErrorResponse) {
	b, err := json.Marshal(errorResponse)
	if err != nil {
		http.Error(w, errs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(b); err != nil {
		log.ErrorCtx(ctx, errors.WithMessage(err, "failed to write error response body"), nil)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
