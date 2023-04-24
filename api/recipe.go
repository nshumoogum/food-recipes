package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/nshumoogum/food-recipes/helpers"
	"github.com/nshumoogum/food-recipes/models"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const defaultLimit = 20

var casing = cases.Title(language.English)

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

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Error(ctx, "get recipes: error returned attempting to count documents", err)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	if count > 0 {
		var cur *mongo.Cursor
		cur, err = collection.Find(ctx, bson.M{})
		if err != nil {
			log.Error(ctx, "get recipes: error returned retrieving a list of recipes", err)
			errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
			ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
			return
		}
		defer cur.Close(ctx)

		var offsetCounter, limitCounter int

		for cur.Next(ctx) {
			offsetCounter++
			if offsetCounter <= page.Offset {
				continue
			}

			if limitCounter >= page.Limit {
				break
			}
			limitCounter++

			item := &models.Recipe{}

			err = cur.Decode(item)
			if err != nil {
				log.Error(ctx, "get recipes: unable to decode recipe", err)
				errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
				ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
				return
			}

			list.Items = append(list.Items, *item)
		}
		if err = cur.Err(); err != nil {
			log.Error(ctx, "get recipes: mongo db cursor error", err)
			errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
			ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
			return
		}
	}

	list.Count = len(list.Items)
	list.Limit = page.Limit
	list.Offset = page.Offset
	list.TotalCount = count

	b, err := json.Marshal(list)
	if err != nil {
		log.Error(ctx, "get recipes: error returned from json marshal", err)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(b); err != nil {
		log.Error(ctx, "get recipes: failed to write response data", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info(ctx, "get recipes: request successful")
}

func (api *FoodRecipeAPI) getRecipe(w http.ResponseWriter, req *http.Request) {
	defer DrainBody(req)
	ctx := req.Context()

	vars := mux.Vars(req)
	id := vars["id"]
	logData := log.Data{"id": id}

	var recipe models.Recipe
	var errorObjects []*models.ErrorObject

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")
	if err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&recipe); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Warn(ctx, "get recipes: failed to find recipe", log.FormatErrors([]error{err}), logData)
			errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrRecipeNotFound.Error()})
			ErrorResponse(ctx, w, http.StatusNotFound, &models.ErrorResponse{Errors: errorObjects})
			return
		}

		log.Error(ctx, "get recipes: failed to find recipe, bad connection?", err)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	b, err := json.Marshal(recipe)
	if err != nil {
		log.Error(ctx, "error returned from json marshal", err, logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(b); err != nil {
		log.Error(ctx, "get recipe: failed to write response data", err, logData)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info(ctx, "get recipe: request successful", logData)
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

	// validate recipe fields
	if errorObjects = recipe.Validate(); len(errorObjects) != 0 {
		ErrorResponse(ctx, w, http.StatusBadRequest, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	recipe.Title = casing.String(recipe.Title)

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")

	if _, err = collection.InsertOne(ctx, recipe); err != nil {
		if strings.Contains(err.Error(), "E11000 duplicate key error collection") {
			log.Error(ctx, "add recipe: failed to insert recipe, recipe already exists", err, logData)
			errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrRecipeAlreadyExists.Error()})
			ErrorResponse(ctx, w, http.StatusConflict, &models.ErrorResponse{Errors: errorObjects})
			return
		}

		log.Error(ctx, "add recipe: failed to insert recipe", err, logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	b, err := json.Marshal(recipe)
	if err != nil {
		log.Error(ctx, "add recipe: failed to marshal instance to json", err, logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(b); err != nil {
		log.Warn(ctx, "add recipe: failed to write response data", log.FormatErrors([]error{err}), logData)
		return
	}

	log.Info(ctx, "add recipe: request successful", logData)
}

func (api *FoodRecipeAPI) updateRecipe(w http.ResponseWriter, req *http.Request) {
	defer DrainBody(req)
	ctx := req.Context()

	vars := mux.Vars(req)
	id := vars["id"]
	logData := log.Data{"id": id}

	var errorObjects []*models.ErrorObject

	recipe, err := unmarshalUpdateRecipe(ctx, req.Body)
	if err != nil {
		errorObjects = append(errorObjects, &models.ErrorObject{Error: err.Error()})
		ErrorResponse(ctx, w, http.StatusBadRequest, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	// validate recipe fields
	if errorObjects = recipe.Validate(); len(errorObjects) != 0 {
		ErrorResponse(ctx, w, http.StatusBadRequest, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	recipe.Title = casing.String(strings.ReplaceAll(id, "-", " "))

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")

	if _, err = collection.ReplaceOne(ctx, bson.M{"_id": id}, recipe); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error(ctx, "update recipe: failed to update recipe, recipe deos not exists", err, logData)
			errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrRecipeNotFound.Error()})
			ErrorResponse(ctx, w, http.StatusNotFound, &models.ErrorResponse{Errors: errorObjects})
			return
		}

		log.Error(ctx, "update recipe: failed to insert recipe", err, logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Info(ctx, "update recipe: request successful", logData)
}

func (api *FoodRecipeAPI) removeRecipe(w http.ResponseWriter, req *http.Request) {
	defer DrainBody(req)
	ctx := req.Context()

	vars := mux.Vars(req)
	id := vars["id"]
	logData := log.Data{"id": id}

	var errorObjects []*models.ErrorObject

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")

	res, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Error(ctx, "delete recipe: failed to remove recipe", err, logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	if res.DeletedCount == 0 {
		log.Warn(ctx, "delete recipe: failed to remove recipe as it does not exist", logData)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	log.Info(ctx, "delete recipe: request successful", logData)
}

func unmarshalRecipe(ctx context.Context, reader io.Reader) (*models.Recipe, error) {
	b, err := io.ReadAll(reader)
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

func unmarshalUpdateRecipe(ctx context.Context, reader io.Reader) (*models.UpdateRecipe, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.ErrUnableToReadMessage
	}

	var recipe models.UpdateRecipe
	err = json.Unmarshal(b, &recipe)
	if err != nil {
		return nil, errs.ErrUnableToParseJSON
	}

	return &recipe, nil
}
