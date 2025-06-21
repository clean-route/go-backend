package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/viper"

	"github.com/clean-route/go-backend/internal/logger"
	waqi "github.com/clean-route/go-backend/internal/models/waqi"
)

func FetchAQIData(location []float64, delayCode uint8) (float64, error) {
	baseUrl := "https://api.waqi.info/feed/geo:" + fmt.Sprintf("%f;%f/?", location[1], location[0])

	var waqiAccessToken string
	var waqiAccessTokenError bool
	if os.Getenv("RAILWAY") == "true" {
		waqiAccessToken = os.Getenv("WAQI_API_KEY")
	} else {
		waqiAccessToken, waqiAccessTokenError = viper.Get("WAQI_API_KEY").(string)
		if !waqiAccessTokenError {
			logger.Error("Invalid WAQI API key configuration")
			log.Fatalf("Invalid type assertion")
		}
	}

	params := url.Values{}
	params.Add("token", waqiAccessToken)

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
		return 0, err
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
		return 0, err
	}

	var waqiResponse waqi.APIResponse

	err = json.Unmarshal([]byte(body), &waqiResponse)
	if err != nil {
		logger.Error("Failed to unmarshal WAQI API response",
			"error", err.Error(),
			"url", baseUrl,
			"location", location,
			"response_body", string(body),
		)
		log.Fatal("Error while unmarshling JSON: ", err)
	}

	// Currently not utilizing the delayCode - no forecasting till now.
	// Will have to to update from this part onwards to include:
	/*
		1. API call to fetch the weather data from darksky or similary apis
			- relative humidity
			- wind speed
			- wind direction
			- temperature
			- we need these current and forecasted both the values
	*/
	// 2. We need to make api call to AWS Sagemaker and get the forecasted aqi value.

	if waqiResponse.Status != "ok" {
		logger.Error("WAQI API returned non-OK status",
			"status", waqiResponse.Status,
			"url", baseUrl,
			"location", location,
			"response_body", string(body),
		)
		return 0, errors.New("WAQI response is not 'OK' but: " + waqiResponse.Status)
	} else {
		pm25Value := waqiResponse.Data.IAQI["pm25"].V
		// logger.Debug("Successfully fetched PM2.5 data from WAQI",
		// 	"pm25_value", pm25Value,
		// 	"location", location,
		// )
		return pm25Value, nil
	}
}

func checkErrNil(err error) {
	if err != nil {
		logger.Error("API error occurred", "error", err.Error())
		log.Fatal("Error: ", err)
	}
}
