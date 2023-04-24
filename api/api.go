package api

import (
	"context"
	"io"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"
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
func NewFoodRecipeAPI(ctx context.Context, connectionString string, mongoClient *mongo.Client, data map[string]models.Recipe, defaultMaxResults int, router *mux.Router) *FoodRecipeAPI {
	api := &FoodRecipeAPI{
		DefaultMaxResults: defaultMaxResults,
		MongoClient:       mongoClient,
		Router:            router,
	}

	api.Router.HandleFunc("/recipes", authorise(connectionString, api.createRecipe)).Methods("POST")
	api.Router.HandleFunc("/recipes", api.getRecipes).Methods("GET")
	api.Router.HandleFunc("/recipes/{id}", api.getRecipe).Methods("GET")
	api.Router.HandleFunc("/recipes/{id}", authorise(connectionString, api.updateRecipe)).Methods("PUT")
	api.Router.HandleFunc("/recipes/{id}", authorise(connectionString, api.removeRecipe)).Methods("DELETE")

	return api
}

func authorise(connectionString string, handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		logData := log.Data{"requested_uri": req.URL.RequestURI()}

		authValue := req.Header.Get("Authorization")

		// Check connection string
		if authValue != connectionString || authValue == "" {
			log.Warn(ctx, "caller unauthorised to perform requested action", logData)

			w.WriteHeader(401)
			return
		}

		log.Info(ctx, "caller authorised to perform requested action", logData)
		handler(w, req)
	})
}

// DrainBody drains the body of the given HTTP request
func DrainBody(r *http.Request) {
	if r.Body == nil {
		return
	}

	_, err := io.Copy(io.Discard, r.Body)
	if err != nil {
		log.Error(r.Context(), "error draining request body", err)
	}

	err = r.Body.Close()
	if err != nil {
		log.Error(r.Context(), "error closing request body", err)
	}
}
