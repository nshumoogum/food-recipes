package service

import (
	"context"
	"net/http"

	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/go-ns/server"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/nshumoogum/food-recipes/api"
	"github.com/nshumoogum/food-recipes/config"
	"github.com/nshumoogum/food-recipes/models"
	"github.com/pkg/errors"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . HTTPServer
//go:generate moq -out mock/closer.go -pkg mock . Closer

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer
}

// HTTPServer defines the required methods from the HTTP server
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// Closer defines the required methods for a closable resource
type Closer interface {
	Close(ctx context.Context) error
}

// Service contains all the configs, server and clients to run the Dataset API
type Service struct {
	api         *api.FoodRecipeAPI
	config      *config.Configuration
	mongoClient *mongo.Client
	server      HTTPServer
}

// New creates a new service
func New(cfg *config.Configuration, mongoClient *mongo.Client) *Service {
	svc := &Service{
		api:         &api.FoodRecipeAPI{},
		config:      cfg,
		mongoClient: mongoClient,
	}

	return svc
}

// getHTTPServer returns an http server
var getHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// Run the service
func (svc *Service) Run(ctx context.Context, recipeData map[string]models.Recipe, svcErrors chan error) (err error) {
	// Get HTTP router and server with middleware
	router := mux.NewRouter()
	svc.api = api.NewFoodRecipeAPI(ctx, svc.mongoClient, recipeData, svc.config.DefaultMaxResults, router)

	server := server.New(svc.config.BindAddr, router)

	// Disable this here to allow main to manage graceful shutdown of the entire app.
	server.HandleOSSignals = false

	svc.server = server

	// Run the http server in a new go-routine
	go func() {
		if err := svc.server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.config.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	shutdownContext, cancel := context.WithTimeout(ctx, timeout)
	hasShutdownError := false

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		// stop any incoming requests
		if err := svc.server.Shutdown(shutdownContext); err != nil {
			log.Event(shutdownContext, "failed to shutdown http server", log.Error(err), log.ERROR)
			hasShutdownError = true
		}

	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownContext.Done()

	// timeout expired
	if shutdownContext.Err() == context.DeadlineExceeded {
		log.Event(shutdownContext, "shutdown timed out", log.ERROR, log.Error(shutdownContext.Err()))
		return shutdownContext.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Event(shutdownContext, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(shutdownContext, "graceful shutdown was successful", log.INFO)
	return nil
}
