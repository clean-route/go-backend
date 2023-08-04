package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sadityakumar9211/clean-route-backend/models"
	"github.com/spf13/viper"
)

type Post struct {
	FPM float64 `json:"fpm"`
}

func GetPredictedPm25(inputFeatures models.FeatureVector, delayCode uint8) (float64, error) {

	awsModelEndpoint, awsModelEndpointError := viper.Get("AWS_MODEL_ENDPOINT").(string)
	if !awsModelEndpointError {
		log.Fatalf("Invalid type assertion")
	}

	fmt.Println("The Query url is: ", awsModelEndpoint)

	inputFeatures.DelayCode = delayCode
	jsonData, err := json.Marshal(inputFeatures)
	checkErrNil(err)

	r, err := http.NewRequest("POST", awsModelEndpoint, bytes.NewBuffer(jsonData))
	checkErrNil(err)

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	checkErrNil(err)

	defer resp.Body.Close()

	post := &Post{}

	err = json.NewDecoder(resp.Body).Decode(post)
	checkErrNil(err)

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Error while making calling API endpoint", err)
		return 0, err
	}

	return post.FPM, nil
}
