package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/clean-route/go-backend/internal/config"
	"github.com/clean-route/go-backend/internal/logger"
	"github.com/clean-route/go-backend/internal/models"
	openweather "github.com/clean-route/go-backend/internal/models/openweather"
	waqimodels "github.com/clean-route/go-backend/internal/models/waqi"
)

// FetchWeatherData fetches weather data from OpenWeather API
func FetchWeatherData(location []float64) openweather.WeatherData {
	baseUrl := "https://api.openweathermap.org/data/3.0/onecall?"

	weatherParams := url.Values{}
	weatherParams.Add("lat", fmt.Sprintf("%f", location[1]))
	weatherParams.Add("lon", fmt.Sprintf("%f", location[0]))
	weatherParams.Add("exclude", "alerts,daily")
	weatherParams.Add("units", "metric")
	weatherParams.Add("appid", config.AppConfig.OpenWeatherAPIKey)

	weatherUrl := baseUrl + weatherParams.Encode()

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
		log.Printf("Error fetching weather data: %v", err)
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
		log.Printf("Error reading weather response: %v", err)
		return openweather.WeatherData{}
	}

	var weatherResponse openweather.WeatherData
	if err := json.Unmarshal(body, &weatherResponse); err != nil {
		logger.Error("Failed to unmarshal OpenWeather API response",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
			"response_body", string(body),
		)
		log.Printf("Error unmarshaling weather JSON: %v", err)
		return openweather.WeatherData{}
	}

	logger.Debug("Successfully fetched weather data from OpenWeather",
		"temperature", weatherResponse.Current.Temp,
		"humidity", weatherResponse.Current.Humidity,
		"location", location,
	)

	return weatherResponse
}

// FetchAQIData fetches air quality data from WAQI API
func FetchAQIData(location []float64, delayCode uint8) (float64, error) {
	baseUrl := "https://api.waqi.info/feed/geo:" + fmt.Sprintf("%f;%f/?", location[1], location[0])

	params := url.Values{}
	params.Add("token", config.AppConfig.WAQIAPIKey)

	waqiUrl := baseUrl + params.Encode()

	// logger.Debug("Calling WAQI API",
	// 	"url", baseUrl,
	// 	"location", location,
	// 	"delay_code", delayCode,
	// )

	resp, err := http.Get(waqiUrl)
	if err != nil {
		logger.Error("Failed to call WAQI API",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
		)
		return 0, fmt.Errorf("error calling WAQI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("WAQI API returned error status",
			"status_code", resp.StatusCode,
			"url", baseUrl,
			"location", location,
		)
		return 0, fmt.Errorf("WAQI API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read WAQI API response",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
		)
		return 0, fmt.Errorf("error reading WAQI response: %w", err)
	}

	var waqiResponse waqimodels.APIResponse
	if err := json.Unmarshal(body, &waqiResponse); err != nil {
		logger.Error("Failed to unmarshal WAQI API response",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
			"response_body", string(body),
		)
		return 0, fmt.Errorf("error unmarshaling WAQI JSON: %w", err)
	}

	if waqiResponse.Status != "ok" {
		logger.Error("WAQI API returned non-OK status",
			"status", waqiResponse.Status,
			"url", baseUrl,
			"location", location,
			"response_body", string(body),
		)
		return 0, errors.New("WAQI response is not 'OK' but: " + waqiResponse.Status)
	}

	pm25Value := waqiResponse.Data.IAQI["pm25"].V
	// logger.Debug("Successfully fetched PM2.5 data from WAQI",
	// 	"pm25_value", pm25Value,
	// 	"location", location,
	// )

	return pm25Value, nil
}

// GetPredictedPm25 gets PM2.5 predictions from AWS model
func GetPredictedPm25(features []models.FeatureVector) ([]float64, error) {
	logger.Debug("Calling AWS PM2.5 prediction API",
		"endpoint", config.AppConfig.AWSModelEndpoint,
		"features_count", len(features),
	)

	jsonData, err := json.Marshal(features)
	if err != nil {
		logger.Error("Failed to marshal feature vector data",
			"error", err.Error(),
			"features_count", len(features),
		)
		return nil, fmt.Errorf("error marshaling features: %w", err)
	}

	req, err := http.NewRequest("POST", config.AppConfig.AWSModelEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create HTTP request for AWS API",
			"error", err.Error(),
			"endpoint", config.AppConfig.AWSModelEndpoint,
		)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to call AWS PM2.5 prediction API",
			"error", err.Error(),
			"endpoint", config.AppConfig.AWSModelEndpoint,
		)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("AWS PM2.5 prediction API returned error status",
			"status_code", resp.StatusCode,
			"endpoint", config.AppConfig.AWSModelEndpoint,
		)
		return nil, fmt.Errorf("AWS model returned status code: %d", resp.StatusCode)
	}

	var response struct {
		FPMVec []float64 `json:"fpm_vec"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error("Failed to decode AWS PM2.5 prediction response",
			"error", err.Error(),
			"endpoint", config.AppConfig.AWSModelEndpoint,
		)
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	logger.Debug("Successfully received PM2.5 predictions from AWS",
		"predictions_count", len(response.FPMVec),
		"endpoint", config.AppConfig.AWSModelEndpoint,
	)

	return response.FPMVec, nil
}
