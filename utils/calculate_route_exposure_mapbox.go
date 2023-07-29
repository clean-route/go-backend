package utils

import (
	"log"
	"fmt"
	mapbox "github.com/sadityakumar9211/clean-route-backend/models/mapbox"
	api "github.com/sadityakumar9211/clean-route-backend/api"
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

			chunks := int(steps[j].Distance / 2000)   // number of chunks
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

	fmt.Println(routePoints)
	fmt.Println(routePointTime)

	// fetching the aqi values for the points in the route

	var totalRouteExposure float64

	for j := 0; j < len(routePoints); j++ {
		// fetch the aqi values for the points in the routes
		if routePoints[j] == nil /* || routePointTime[j] == nil */ {
			continue
		}
		aqiData, err := api.FetchAQIData(routePoints[j], delayCode)
		checkErrNil(err)

		fmt.Println("The PM 2.5 concentration: ", aqiData)
		// this will cause error.
		totalRouteExposure += (aqiData * routePointTime[j]) / 3600 // converting time to hours

	}

	route.TotalExposure = totalRouteExposure
	fmt.Println("The total exposure for the route is: ", totalRouteExposure)
	return route
}

func checkErrNil(err error) {
	if err != nil {
		log.Fatalf("Error encountered: %s", err)
	}
}
