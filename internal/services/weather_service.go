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

	resp, err := http.Get(weatherUrl)
	if err != nil {
		log.Printf("Error fetching weather data: %v", err)
		return openweather.WeatherData{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading weather response: %v", err)
		return openweather.WeatherData{}
	}

	var weatherResponse openweather.WeatherData
	if err := json.Unmarshal(body, &weatherResponse); err != nil {
		log.Printf("Error unmarshaling weather JSON: %v", err)
		return openweather.WeatherData{}
	}

	return weatherResponse
}

// FetchAQIData fetches air quality data from WAQI API
func FetchAQIData(location []float64, delayCode uint8) (float64, error) {
	baseUrl := "https://api.waqi.info/feed/geo:" + fmt.Sprintf("%f;%f/?", location[1], location[0])

	params := url.Values{}
	params.Add("token", config.AppConfig.WAQIAPIKey)

	waqiUrl := baseUrl + params.Encode()

	resp, err := http.Get(waqiUrl)
	if err != nil {
		return 0, fmt.Errorf("error calling WAQI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading WAQI response: %w", err)
	}

	var waqiResponse waqimodels.APIResponse
	if err := json.Unmarshal(body, &waqiResponse); err != nil {
		return 0, fmt.Errorf("error unmarshaling WAQI JSON: %w", err)
	}

	if waqiResponse.Status != "ok" {
		return 0, errors.New("WAQI response is not 'OK' but: " + waqiResponse.Status)
	}

	return waqiResponse.Data.IAQI["pm25"].V, nil
}

// GetPredictedPm25 gets PM2.5 predictions from AWS model
func GetPredictedPm25(features []models.FeatureVector) ([]float64, error) {
	jsonData, err := json.Marshal(features)
	if err != nil {
		return nil, fmt.Errorf("error marshaling features: %w", err)
	}

	req, err := http.NewRequest("POST", config.AppConfig.AWSModelEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AWS model returned status code: %d", resp.StatusCode)
	}

	var response struct {
		FPMVec []float64 `json:"fpm_vec"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return response.FPMVec, nil
}
