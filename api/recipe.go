package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/nshumoogum/food-recipes/helpers"
	"github.com/nshumoogum/food-recipes/models"
)

const defaultLimit = 20

func (api *FoodRecipeAPI) getRecipes(w http.ResponseWriter, req *http.Request) {
	defer DrainBody(req)
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

	page := models.PageVariables{
		DefaultMaxResults: api.DefaultMaxResults,
		Limit:             limit,
		Offset:            offset,
	}

	if errorObject := models.ValidatePage(page); errorObject != nil {
		errorObjects = append(errorObjects, errorObject...)
	}

	if len(errorObjects) != 0 {
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
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

	log.Event(ctx, "get recipes: request successful", log.INFO)
}

func (api *FoodRecipeAPI) getRecipe(w http.ResponseWriter, req *http.Request) {
	defer DrainBody(req)
	ctx := req.Context()

	vars := mux.Vars(req)
	id := vars["id"]
	logData := log.Data{"id": id}

	recipe := api.RecipeData[id]

	var errorObjects []*models.ErrorObject

	if recipe.ID != id {
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrRecipeNotFound.Error(), ErrorValues: map[string]string{"id": id}})
		ErrorResponse(ctx, w, http.StatusNotFound, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	b, err := json.Marshal(recipe)
	if err != nil {
		log.Event(ctx, "error returned from json marshal", log.ERROR, log.Error(err), logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

	log.Event(ctx, "get recipe: request successful", log.INFO, logData)
}

func (api *FoodRecipeAPI) createRecipe(w http.ResponseWriter, req *http.Request) {
	defer DrainBody(req)
	ctx := req.Context()

	var errorObjects []*models.ErrorObject

	recipe, err := unmarshalRecipe(ctx, req.Body)
	if err != nil {
		errorObjects = append(errorObjects, &models.ErrorObject{Error: err.Error()})
		ErrorResponse(ctx, w, http.StatusBadRequest, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	recipe.ID = strings.ToLower(strings.ReplaceAll(recipe.Title, " ", "-"))
	logData := log.Data{"id": recipe.ID}

	r := api.RecipeData[recipe.ID]

	if r.ID == recipe.ID {
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrRecipeAlreadyExists.Error(), ErrorValues: map[string]string{"title": recipe.Title}})
		ErrorResponse(ctx, w, http.StatusConflict, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	// validate recipe fields
	if errorObjects = recipe.Validate(); len(errorObjects) != 0 {
		ErrorResponse(ctx, w, http.StatusBadRequest, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	api.RecipeData[recipe.ID] = *recipe

	b, err := json.Marshal(recipe)
	if err != nil {
		log.Event(ctx, "add recipe: failed to marshal instance to json", log.ERROR, log.Error(err), logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(b)

	log.Event(ctx, "add recipe: request successful", log.INFO, logData)
}

func (api *FoodRecipeAPI) updateRecipe(w http.ResponseWriter, req *http.Request) {
	// ctx := req.Context()
}

func unmarshalRecipe(ctx context.Context, reader io.Reader) (*models.Recipe, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var recipe models.Recipe
	err = json.Unmarshal(b, &recipe)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}

	return &recipe, nil
}
