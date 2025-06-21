package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/spf13/viper"

	"github.com/clean-route/go-backend/internal/logger"
	"github.com/clean-route/go-backend/internal/models"
)

type Post struct {
	FPMVec []float64 `json:"fpm_vec"`
}

func GetPredictedPm25(df []models.FeatureVector) ([]float64, error) {

	var awsModelEndpoint string
	var awsModelEndpointError bool

	if os.Getenv("RAILWAY") == "true" {
		awsModelEndpoint = os.Getenv("AWS_MODEL_ENDPOINT")
	} else {
		awsModelEndpoint, awsModelEndpointError = viper.Get("AWS_MODEL_ENDPOINT").(string)
		if !awsModelEndpointError {
			logger.Error("Invalid AWS model endpoint configuration")
			log.Fatalf("Invalid type assertion")
		}
	}

	logger.Debug("Calling AWS PM2.5 prediction API",
		"endpoint", awsModelEndpoint,
		"features_count", len(df),
	)

	jsonData, err := json.Marshal(df)
	if err != nil {
		logger.Error("Failed to marshal feature vector data",
			"error", err.Error(),
			"features_count", len(df),
		)
		return nil, err
	}

	r, err := http.NewRequest("POST", awsModelEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create HTTP request for AWS API",
			"error", err.Error(),
			"endpoint", awsModelEndpoint,
		)
		return nil, err
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Error("Failed to call AWS PM2.5 prediction API",
			"error", err.Error(),
			"endpoint", awsModelEndpoint,
		)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("AWS PM2.5 prediction API returned error status",
			"status_code", resp.StatusCode,
			"endpoint", awsModelEndpoint,
		)
		log.Fatal("The Amazon EC2 instance is not responding...Got status code: ", resp.StatusCode)
	}

	defer resp.Body.Close()

	post := &Post{}

	err = json.NewDecoder(resp.Body).Decode(post)
	if err != nil {
		logger.Error("Failed to decode AWS PM2.5 prediction response",
			"error", err.Error(),
			"endpoint", awsModelEndpoint,
		)
		return nil, err
	}

	logger.Debug("Successfully received PM2.5 predictions from AWS",
		"predictions_count", len(post.FPMVec),
		"endpoint", awsModelEndpoint,
	)

	return post.FPMVec, nil
}
