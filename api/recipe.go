package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/log.go/log"
	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/nshumoogum/food-recipes/helpers"
	"github.com/nshumoogum/food-recipes/models"
)

const defaultLimit = 20

func (api *FoodRecipeAPI) getRecipes(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	requestedOffset := req.FormValue("offset")
	requestedLimit := req.FormValue("limit")

	var errorObjects []*models.ErrorObject

	limit, err := helpers.CalculateLimit(ctx, defaultLimit, api.DefaultMaxResults, requestedLimit)
	if err != nil {
		errorObjects = append(errorObjects, &models.ErrorObject{Error: err.Error(), ErrorValues: err.(*errs.ErrorObject).Values()})
	}

	offset, err := helpers.CalculateOffset(ctx, requestedOffset)
	if err != nil {
		errorObjects = append(errorObjects, &models.ErrorObject{Error: err.Error(), ErrorValues: err.(*errs.ErrorObject).Values()})
	}

	page := helpers.PageVariables{
		DefaultMaxResults: api.DefaultMaxResults,
		Limit:             limit,
		Offset:            offset,
	}

	if errorObject := helpers.ValidatePage(page); errorObject != nil {
		errorObjects = append(errorObjects, errorObject...)
	}

	if errorObjects != nil {
		ErrorResponse(ctx, w, http.StatusBadRequest, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	var list models.Recipes
	list.Items = []models.Recipe{}

	recipes := api.RecipeData

	var offsetCounter, limitCounter int
	for _, item := range recipes {
		offsetCounter++
		if offsetCounter <= page.Offset {
			continue
		}

		if limitCounter >= page.Limit {
			break
		}
		limitCounter++

		// Add item
		list.Items = append(list.Items, item)
	}

	list.Count = len(list.Items)
	list.Limit = page.Limit
	list.Offset = page.Offset
	list.TotalCount = len(recipes)

	b, err := json.Marshal(list)
	if err != nil {
		log.Event(ctx, "error returned from json marshal", log.ERROR, log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (api *FoodRecipeAPI) getRecipe(w http.ResponseWriter, req *http.Request) {

}

func (api *FoodRecipeAPI) createRecipe(w http.ResponseWriter, req *http.Request) {

}

func (api *FoodRecipeAPI) updateRecipe(w http.ResponseWriter, req *http.Request) {

}
