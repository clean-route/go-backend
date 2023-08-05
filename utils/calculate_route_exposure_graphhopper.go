package utils

import (
	"fmt"

	api "github.com/sadityakumar9211/clean-route-backend/api"
	graphhopper "github.com/sadityakumar9211/clean-route-backend/models/graphhopper"
)

func CalculateRouteExposureGraphhopper(route graphhopper.Path, delayCode uint8) graphhopper.Path {
	var routePoints [][]float64
	var routePointTime []float64

	var skippedDistance float64
	var skippedTime float64

	routeCoordinates := route.Points.Coordinates
	steps := route.Instructions
	for j := 0; j < len(steps); j++ {
		if steps[j].Distance < 1000 {
			// if the distance is less than 1 KM, we skip the distance
			if skippedDistance >= 2 {
				index := steps[j].Interval[1]
				routePoints = append(routePoints, routeCoordinates[index][:])
				routePointTime = append(routePointTime, (float64(steps[j].Time)*0.001)+skippedTime)
			} else {
				skippedDistance += steps[j].Distance * 0.001
				skippedTime += float64(steps[j].Time) * 0.001
				continue
			}
		} else if steps[j].Distance < 2000 {
			// for distance between 1km and 2km
			skippedDistance = 0
			skippedTime = 0

			// taking the middle coordinate of the step
			index := int((steps[j].Interval[0]+steps[j].Interval[1])/2) + 1
			routePoints = append(routePoints, routeCoordinates[index][:])
			routePointTime = append(routePointTime, float64(steps[j].Time)*0.001)
		} else if steps[j].Distance >= 2000 {
			skippedDistance = 0
			skippedTime = 0

			chunks := int(steps[j].Distance / 2000)                              // number of chunks
			timeChunk := int((float64(steps[j].Time) * 0.001) / float64(chunks)) // time for each chunk

			chunkLength := (steps[j].Interval[1] - steps[j].Interval[0]) / chunks // number of coordinates in each chunk

			for k := 0; k < chunks; k++ {
				startChunkIndex := steps[j].Interval[0] + k*chunkLength
				index := (int(startChunkIndex+startChunkIndex+chunkLength) / 2) + 1
				routePoints = append(routePoints, routeCoordinates[index][:])
				routePointTime = append(routePointTime, float64(timeChunk))
			}
		}
	}

	// for each route the points are adding in this array

	// fmt.Println(routePoints)
	fmt.Println("The total points taken in the route is: ", len(routePoints))
	// fmt.Println(routePointTime)

	if delayCode == 0 {
		var totalRouteExposure float64

		for j := 0; j < len(routePoints); j++ {
			// fetch the aqi values for the points in the routes
			if routePoints[j] == nil /* || routePointTime[j] == nil */ {
				continue
			}
			pm25, err := api.FetchAQIData(routePoints[j], delayCode)
			checkErrNil(err)
			fmt.Println("The PM 2.5 concentration: ", pm25)
			totalRouteExposure += pm25 * routePointTime[j] / 3600 // converting time to hours
		}
		route.TotalExposure = totalRouteExposure
		// fmt.Println("************The total exposure for the route is**********: ", totalRouteExposure)
		return route
	}

	// Fetch the weather data for source and destination and we will use the
	// average of the both for any point in route to get the weather measurement
	sourceWeatherData := api.FetchWeatherData(routePoints[0])
	sourceWeatherData.Current.RelativeHumidity = GetRelativeHumidity(sourceWeatherData.Current.DewPoint, sourceWeatherData.Current.Temp)
	destinationWeatherData := api.FetchWeatherData(routePoints[len(routePoints)-1])
	destinationWeatherData.Current.RelativeHumidity = GetRelativeHumidity(destinationWeatherData.Current.DewPoint, destinationWeatherData.Current.Temp)

	inputFeatures := GetInputFeatures(sourceWeatherData, destinationWeatherData, delayCode) // except IPM

	// fetching the aqi values for the points in the route
	var totalRouteExposure float64
	for j := 0; j < len(routePoints); j++ {
		// fetch the aqi values for the points in the routes
		if routePoints[j] == nil /* || routePointTime[j] == nil */ {
			continue
		}
		pm25, err := api.FetchAQIData(routePoints[j], delayCode)
		checkErrNil(err)
		inputFeatures.IPM = pm25

		// call the aws sagemaker and get the predicted value of pm25 concentration.
		fpm, err := api.GetPredictedPm25(inputFeatures, delayCode)
		checkErrNil(err)

		// calculate the total exposure based on that.
		print("\n\nPartial Exposure: ", totalRouteExposure, " # ", fpm, " # ", routePointTime[j])
		totalRouteExposure += fpm * routePointTime[j] / 3600 // converting time to hours
	}

	route.TotalExposure = totalRouteExposure
	fmt.Println("&&&&&&&&&&&&&&&&The total exposure for the route is&&&&&&&&&&&&&&&: ", totalRouteExposure)
	return route
}
