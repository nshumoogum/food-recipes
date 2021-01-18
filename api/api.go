package api

import (
	"context"

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

	api.Router.HandleFunc("/recipes", api.getRecipes).Methods("GET")
	api.Router.HandleFunc("/recipe/{id}", api.getRecipe).Methods("GET")
	api.Router.HandleFunc("/recipe/{id}", api.createRecipe).Methods("POST")
	api.Router.HandleFunc("/recipe/{id}", api.updateRecipe).Methods("PATCH")

	return api
}
