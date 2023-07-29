package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	waqi "github.com/sadityakumar9211/clean-route-backend/models/waqi"
	"github.com/spf13/viper"
)

func FetchAQIData(location []float64, deplayCode uint8) (float64, error) {
	baseUrl := "https://api.waqi.info/feed/geo:" + fmt.Sprintf("%f;%f/?", location[1], location[2])

	waqiAccessToken, waqiAccessTokenError := viper.Get("WAQI_API_KEY").(string)
	if !waqiAccessTokenError {
		log.Fatalf("Invalid type assertion")
	}

	params := url.Values{}
	params.Add("token", waqiAccessToken)

	url := baseUrl + params.Encode()

	resp, err := http.Get(url)
	checkErrNil(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	checkErrNil(err)

	var waqiResponse waqi.APIResponse

	err = json.Unmarshal([]byte(body), &waqiResponse)
	if err != nil {
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
	2. We need to make api call to AWS Sagemaker and get the forecasted aqi value. 
	*/


	
	if waqiResponse.Status != "ok" {
		return 0, errors.New("WAQI response is not 'OK' but: " +  waqiResponse.Status)
	} else{
		return waqiResponse.Data.IAQI["pm25"].V, nil 
	}
}

func checkErrNil(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
