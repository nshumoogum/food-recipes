package config

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Configuration structure which hold information for configuring the import API
type Configuration struct {
	BindAddr                string        `envconfig:"BIND_ADDR"`
	ConnectionString        string        `envconfig:"CONNECTION_STRING"          json:"-"`
	DefaultMaxResults       int           `envconfig:"DEFAULT_MAX_RESULTS"`
	DownloadData            bool          `envconfig:"DOWNLOAD_DATA"`
	DownloadTimeout         time.Duration `envconfig:"DOWNLOAD_TIMEOUT"`
	GSURL                   string        `envconfig:"GOOGLE_SHEET_URL"`
	GracefulShutdownTimeout time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	MongoConfig             MongoConfig
}

// MongoConfig contains the config required to connect to MongoDB.
type MongoConfig struct {
	BindAddr   string `envconfig:"MONGODB_BIND_ADDR"   json:"-"`
	Collection string `envconfig:"MONGODB_COLLECTION"`
	Database   string `envconfig:"MONGODB_DATABASE"`
}

var cfg *Configuration

// Get the application and returns the configuration structure
func Get() (*Configuration, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Configuration{
		BindAddr:                ":30000",
		ConnectionString:        "",
		DefaultMaxResults:       50,
		DownloadData:            false,
		DownloadTimeout:         5 * time.Second,
		GSURL:                   "",
		GracefulShutdownTimeout: 5 * time.Second,
		MongoConfig: MongoConfig{
			BindAddr:   "mongodb://localhost:27017",
			Collection: "recipes",
			Database:   "food-recipes",
		},
	}

	return cfg, envconfig.Process("", cfg)
}

// String is implemented to prevent sensitive fields being logged.
// The config is returned as JSON with sensitive fields omitted.
func (config Configuration) String() string {
	json, _ := json.Marshal(config)
	return string(json)
}
