package api

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/nshumoogum/food-recipes/models"
)

// FoodRecipeAPI manages access to food recipes
type FoodRecipeAPI struct {
	DefaultMaxResults int
	RecipeData        map[string]models.Recipe
	Router            *mux.Router
}

// NewFoodRecipeAPI create a new Food Recipe API instance and register the API routes based on the application configuration.
func NewFoodRecipeAPI(ctx context.Context, data map[string]models.Recipe, defaultMaxResults int, router *mux.Router) *FoodRecipeAPI {
	api := &FoodRecipeAPI{
		DefaultMaxResults: defaultMaxResults,
		RecipeData:        data,
		Router:            router,
	}

	api.Router.HandleFunc("/recipes", api.createRecipe).Methods("POST")
	api.Router.HandleFunc("/recipes", api.getRecipes).Methods("GET")
	api.Router.HandleFunc("/recipes/{id}", api.getRecipe).Methods("GET")
	api.Router.HandleFunc("/recipe/{id}", api.updateRecipe).Methods("PATCH")

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
