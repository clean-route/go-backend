package utils

import (
	"fmt"
	"log"

	"github.com/clean-route/go-backend/api"
	"github.com/clean-route/go-backend/internal/models"
	mapbox "github.com/clean-route/go-backend/internal/models/mapbox"
)

func CalculateRouteExposureMapbox(route mapbox.Route, delayCode uint8) mapbox.Route {
	var routePoints [][]float64
	var routePointTime []float64

	var skippedDistance float64
	var skippedTime float64

	steps := route.Legs[0].Steps

	for j := 0; j < len(steps); j++ {
		if steps[j].Distance < 1000 {
			// if the distance is less than 1 KM, we skip the distance
			if skippedDistance >= 2 {
				index := len(steps[j].Geometry.Coordinates) / 2
				routePoints = append(routePoints, steps[j].Geometry.Coordinates[index])
				routePointTime = append(routePointTime, steps[j].Duration+skippedTime)
			} else {
				skippedDistance += steps[j].Distance * 0.001
				skippedTime += steps[j].Duration
				continue
			}
		} else if steps[j].Distance < 2000 {
			// for distance between 1km and 2km
			skippedDistance = 0
			skippedTime = 0

			// taking the middle coordinate of the step
			index := len(steps[j].Geometry.Coordinates) / 2
			routePoints = append(routePoints, steps[j].Geometry.Coordinates[index])
			routePointTime = append(routePointTime, steps[j].Duration)
		} else if steps[j].Distance >= 2000 {
			skippedDistance = 0
			skippedTime = 0

			chunks := int(steps[j].Distance / 2000)          // number of chunks
			timeChunk := steps[j].Duration / float64(chunks) // time for each chunk

			chunkLength := len(steps[j].Geometry.Coordinates) / chunks // number of coordinates in each chunk

			for k := 0; k < chunks; k++ {
				startChunkIndex := k * chunkLength
				endChunkIndex := (k + 1) * chunkLength
				index := (startChunkIndex + endChunkIndex) / 2
				routePoints = append(routePoints, steps[j].Geometry.Coordinates[index])
				routePointTime = append(routePointTime, timeChunk)
			}
		}
	}

	// for each route the points are adding in this array

	// fmt.Println(routePoints)
	fmt.Println("The total points taken in the route is: ", len(routePoints))
	// fmt.Println(routePointTime)

	// fetching the aqi values for the points in the route

	if delayCode == 0 {
		var totalRouteExposure float64

		for j := 0; j < len(routePoints); j++ {
			// fetch the aqi values for the points in the routes
			if routePoints[j] == nil {
				continue
			}
			pm25, err := api.FetchAQIData(routePoints[j], delayCode)
			checkErrNil(err)
			fmt.Println("The PM 2.5 concentration: ", pm25)
			totalRouteExposure += pm25 * routePointTime[j] / 3600 // converting time to hours
		}
		route.TotalExposure = totalRouteExposure
		return route
	}

	// // Fetch the weather data for source and destination and we will use the average of the both for any point in route to get the weather measurement

	// sourceWeatherData := api.FetchWeatherData(routePoints[0])
	// sourceWeatherData.Current.RelativeHumidity = GetRelativeHumidity(sourceWeatherData.Current.DewPoint, sourceWeatherData.Current.Temp)
	// destinationWeatherData := api.FetchWeatherData(routePoints[len(routePoints)-1])
	// destinationWeatherData.Current.RelativeHumidity = GetRelativeHumidity(destinationWeatherData.Current.DewPoint, destinationWeatherData.Current.Temp)

	// inputFeatures := GetInputFeatures(sourceWeatherData, destinationWeatherData, delayCode) // except IPM

	// // fetching the aqi values for the points in the route

	// var totalRouteExposure float64

	// for j := 0; j < len(routePoints); j++ {
	// 	if routePoints[j] == nil {
	// 		continue
	// 	}
	// 	pm25, err := api.FetchAQIData(routePoints[j], delayCode)
	// 	checkErrNil(err)
	// 	inputFeatures.IPM = pm25

	// 	// call the aws ec2 instance and get the predicted value of pm25 concentration.
	// 	fpm, err := api.GetPredictedPm25(inputFeatures, delayCode)
	// 	checkErrNil(err)

	// 	// Calculate the total exposure based on that.
	// 	fmt.Println("\n\nPartial Exposure: ", totalRouteExposure, " # ", fpm, " # ", routePointTime[j])
	// 	totalRouteExposure += fpm * routePointTime[j] / 3600  // converting time to hours
	// }

	var totalRouteExposure float64 = GetRouteExposureFromRoutePoints(routePoints, routePointTime, delayCode)
	route.TotalExposure = totalRouteExposure
	fmt.Println("&&&&&&&&&&&&&&&&The total exposure for the route is&&&&&&&&&&&&&&&: ", totalRouteExposure)
	return route
}

func GetRouteExposureFromRoutePoints(routePoints [][]float64, routePointTime []float64, delayCode uint8) float64 {
	// Fetch the weather data for source and destination and we will use the average of the both for any point in route to get the weather measurement
	sourceWeatherData := api.FetchWeatherData(routePoints[0])
	sourceWeatherData.Current.RelativeHumidity = GetRelativeHumidity(sourceWeatherData.Current.DewPoint, sourceWeatherData.Current.Temp)
	destinationWeatherData := api.FetchWeatherData(routePoints[len(routePoints)-1])
	destinationWeatherData.Current.RelativeHumidity = GetRelativeHumidity(destinationWeatherData.Current.DewPoint, destinationWeatherData.Current.Temp)

	inputFeatures := GetInputFeatures(sourceWeatherData, destinationWeatherData, delayCode) // except IPM
	inputFeatures.DelayCode = delayCode

	// fetching the aqi values for the points in the route

	var totalRouteExposure float64
	var df []models.FeatureVector

	// constructing the dataframe (input features along the entire route)
	for j := 0; j < len(routePoints); j++ {
		if routePoints[j] == nil {
			continue
		}
		pm25, err := api.FetchAQIData(routePoints[j], delayCode)
		checkErrNil(err)
		inputFeatures.IPM = pm25

		// call the aws ec2 instance and get the predicted value of pm25 concentration.
		df = append(df, inputFeatures)
		// fpm, err := api.GetPredictedPm25(inputFeatures, delayCode)
		// checkErrNil(err)

		// // Calculate the total exposure based on that.
		// fmt.Println("\n\nPartial Exposure: ", totalRouteExposure, " # ", fpm, " # ", routePointTime[j])
		// totalRouteExposure += fpm * routePointTime[j] / 3600  // converting time to hours
	}

	// making the api call to get the entire
	fpmVec, err := api.GetPredictedPm25(df)
	checkErrNil(err)

	// calculating the total exposure
	for j := 0; j < len(routePoints); j++ {
		// calculate the total exposure using the predicted fpm
		fmt.Println("\n\nPartial Exposure: ", totalRouteExposure, " # ", fpmVec[j], " # ", routePointTime[j])
		totalRouteExposure += fpmVec[j] * routePointTime[j] / 3600 // converting time to hours
	}

	return totalRouteExposure
}

func checkErrNil(err error) {
	if err != nil {
		log.Fatalf("Error encountered: %s", err)
	}
}
