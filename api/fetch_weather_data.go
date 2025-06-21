package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/viper"

	"github.com/clean-route/go-backend/internal/logger"
	openweather "github.com/clean-route/go-backend/internal/models/openweather"
)

func FetchWeatherData(location []float64) openweather.WeatherData {
	baseUrl := "https://api.openweathermap.org/data/2.5/weather?"

	var openweatherAccessToken string
	var openweatherAccessTokenError bool
	if os.Getenv("RAILWAY") == "true" {
		openweatherAccessToken = os.Getenv("OPENWEATHER_API_KEY")
	} else {
		openweatherAccessToken, openweatherAccessTokenError = viper.Get("OPENWEATHER_API_KEY").(string)
		if !openweatherAccessTokenError {
			logger.Error("Invalid OpenWeather API key configuration")
			log.Fatalf("Invalid type assertion")
		}
	}

	params := url.Values{}
	params.Add("lat", fmt.Sprintf("%f", location[1]))
	params.Add("lon", fmt.Sprintf("%f", location[0]))
	params.Add("appid", openweatherAccessToken)
	params.Add("units", "metric")

	weatherUrl := baseUrl + params.Encode()

	logger.Debug("Calling OpenWeather API",
		"url", baseUrl,
		"location", location,
	)

	resp, err := http.Get(weatherUrl)
	if err != nil {
		logger.Error("Failed to call OpenWeather API",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
		)
		return openweather.WeatherData{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("OpenWeather API returned error status",
			"status_code", resp.StatusCode,
			"url", baseUrl,
			"location", location,
		)
		return openweather.WeatherData{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read OpenWeather API response",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
		)
		return openweather.WeatherData{}
	}

	var weatherResponse openweather.WeatherData

	err = json.Unmarshal([]byte(body), &weatherResponse)
	if err != nil {
		logger.Error("Failed to unmarshal OpenWeather API response",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
			"response_body", string(body),
		)
		log.Fatal("Error while unmarshling JSON: ", err)
	}

	logger.Debug("Successfully fetched weather data from OpenWeather",
		"temperature", weatherResponse.Current.Temp,
		"humidity", weatherResponse.Current.Humidity,
		"location", location,
	)

	return weatherResponse
}
