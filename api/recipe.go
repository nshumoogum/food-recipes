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
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
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

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Event(ctx, "get recipes: error returned attempting to count documents", log.ERROR, log.Error(err))
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	if count > 0 {
		cur, err := collection.Find(ctx, bson.M{})
		if err != nil {
			log.Event(ctx, "get recipes: error returned retrieving a list of recipes", log.ERROR, log.Error(err))
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

			err := cur.Decode(item)
			if err != nil {
				log.Event(ctx, "get recipes: unable to decode recipe", log.ERROR, log.Error(err))
				errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
				ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
				return
			}

			list.Items = append(list.Items, *item)
		}
		if err := cur.Err(); err != nil {
			log.Event(ctx, "get recipes: mongo db cursor error", log.ERROR, log.Error(err))
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
		log.Event(ctx, "get recipes: error returned from json marshal", log.ERROR, log.Error(err))
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

	var recipe models.Recipe
	var errorObjects []*models.ErrorObject

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")
	if err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&recipe); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Event(ctx, "get recipes: failed to find recipe", log.WARN, log.Error(err), logData)
			errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrRecipeNotFound.Error()})
			ErrorResponse(ctx, w, http.StatusNotFound, &models.ErrorResponse{Errors: errorObjects})
			return
		}

		log.Event(ctx, "get recipes: failed to find recipe, bad connection?", log.ERROR, log.Error(err))
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
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

	// validate recipe fields
	if errorObjects = recipe.Validate(); len(errorObjects) != 0 {
		ErrorResponse(ctx, w, http.StatusBadRequest, &models.ErrorResponse{Errors: errorObjects})
		return
	}

	collection := api.MongoClient.Database("food-recipes").Collection("recipes")

	if _, err = collection.InsertOne(ctx, recipe); err != nil {
		if strings.Contains(err.Error(), "E11000 duplicate key error collection") {
			log.Event(ctx, "add recipe: failed to insert recipe, recipe already exists", log.ERROR, log.Error(err), logData)
			errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrRecipeAlreadyExists.Error()})
			ErrorResponse(ctx, w, http.StatusConflict, &models.ErrorResponse{Errors: errorObjects})
			return
		}

		log.Event(ctx, "add recipe: failed to insert recipe", log.ERROR, log.Error(err), logData)
		errorObjects = append(errorObjects, &models.ErrorObject{Error: errs.ErrInternalServer.Error()})
		ErrorResponse(ctx, w, http.StatusInternalServerError, &models.ErrorResponse{Errors: errorObjects})
		return
	}

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
