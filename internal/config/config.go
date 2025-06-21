package config

import (
	"os"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	MapboxAPIKey      string
	GraphhopperAPIKey string
	WAQIAPIKey        string
	OpenWeatherAPIKey string
	AWSModelEndpoint  string
	IsRailway         bool
}

var AppConfig *Config

// Init initializes the configuration
func Init() error {
	viper.SetConfigType("env")

	if os.Getenv("RAILWAY") == "true" {
		viper.SetConfigFile("ENV")
	} else {
		viper.SetConfigFile(".env")
	}

	viper.ReadInConfig()
	viper.AutomaticEnv()

	AppConfig = &Config{
		MapboxAPIKey:      getEnvVar("MAPBOX_API_KEY"),
		GraphhopperAPIKey: getEnvVar("GRAPHHOPPER_API_KEY"),
		WAQIAPIKey:        getEnvVar("WAQI_API_KEY"),
		OpenWeatherAPIKey: getEnvVar("OPEN_WEATHER_API_KEY"),
		AWSModelEndpoint:  getEnvVar("AWS_MODEL_ENDPOINT"),
		IsRailway:         os.Getenv("RAILWAY") == "true",
	}

	return nil
}

// getEnvVar gets environment variable with fallback to viper
func getEnvVar(key string) string {
	if os.Getenv("RAILWAY") == "true" {
		return os.Getenv(key)
	}

	if value, ok := viper.Get(key).(string); ok {
		return value
	}
	return ""
}
