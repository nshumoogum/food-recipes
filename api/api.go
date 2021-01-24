package api

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/nshumoogum/food-recipes/models"
	"go.mongodb.org/mongo-driver/mongo"
)

// FoodRecipeAPI manages access to food recipes
type FoodRecipeAPI struct {
	DefaultMaxResults int
	MongoClient       *mongo.Client
	Router            *mux.Router
}

// NewFoodRecipeAPI create a new Food Recipe API instance and register the API routes based on the application configuration.
func NewFoodRecipeAPI(ctx context.Context, mongoClient *mongo.Client, data map[string]models.Recipe, defaultMaxResults int, router *mux.Router) *FoodRecipeAPI {
	api := &FoodRecipeAPI{
		DefaultMaxResults: defaultMaxResults,
		MongoClient:       mongoClient,
		Router:            router,
	}

	api.Router.HandleFunc("/recipes", api.createRecipe).Methods("POST")
	api.Router.HandleFunc("/recipes", api.getRecipes).Methods("GET")
	api.Router.HandleFunc("/recipes/{id}", api.getRecipe).Methods("GET")
	api.Router.HandleFunc("/recipes/{id}", api.updateRecipe).Methods("PUT")
	api.Router.HandleFunc("/recipes/{id}", api.removeRecipe).Methods("DELETE")

	return api
}

// DrainBody drains the body of the given HTTP request
func DrainBody(r *http.Request) {

	if r.Body == nil {
		return
	}

	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		log.Event(r.Context(), "error draining request body", log.Error(err))
	}

	err = r.Body.Close()
	if err != nil {
		log.Event(r.Context(), "error closing request body", log.Error(err))
	}
}
