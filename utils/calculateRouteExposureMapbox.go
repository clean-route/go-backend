package utils

import (
	mapbox "github.com/sadityakumar9211/clean-route-backend/models/mapbox"
)

func CalculateRouteExposureMapbox(route mapbox.Routes) {
	// routePoints := [][]float64
	// aqiValues := []float64
	// routePointTime := []float64

	// skippedDistance := 0
	// skippedTime := 0

	// const steps = route.Routes.Legs[0].Steps

	// for j := 0; j < len(steps); j++ {
	//     if steps[j].Distance < 1000 {
	//         // if the distance is less than 1 KM, we skip the distance
	//         if skippedDistance >= 2 {
	//             index := len(steps[j].Geometry.Coordinates) / 2
	//             append(routePoints, steps[j].Geometry.Coordinates[index])
	//             append(routePointTime, steps[j].Duration + skippedTime)
	//         } else {
	//             skippedDistance += steps[j].Distance * 0.001
	//             skippedTime += steps[j].Duration
	//             continue
	//         }
	//     } else if steps[j].Distance < 2000 {
	//         // for distance between 1km and 2km
	//         skippedDistance = 0
	//         skippedTime = 0

	//         // taking the middle coordinate of the step
	//         index := len(steps[j].Geometry.Coordinates) / 2
	//         append(routePoints, steps[j].Geometry.Coordinates[index])
	//         append(routePointTime, steps[j].Duration)
	//     } else if steps[j].Distance >= 2000 {
	//         skippedDistance = 0
	//         skippedTime = 0

	//         chunks := steps[j].Distance / 2000  // number of chunks
	//         timeChunk := steps[j].Duration / chunks // time for each chunk

	//         chunkLength := len(steps[j].Geometry.Coordinates) / chunks // number of coordinates in each chunk

	//         for k := 0; k < chunks; k++ {
	//             startChunkIndex := k * chunkLength
	//             endChunkIndex := (k+1) * chunkLength
	//             index := (startChunkIndex + endChunkIndex) / 2
	//             append(routePoints, steps[j].Geometry.Coordinates[index])
	//             append(routePointTime, timeChunk)
	//         }
	//     }
	// }

	// // for each route the points are adding in this array

	// fmt.Println(routePoints)
	// fmt.Pritnln(routePointTime)

	// // fetching the aqi values for the points in the route

	// aqiValues := []float64

	// totalRouteExposure := 0

	// for j := 0; j < len(routePoints); j++ {
	//     // fetch the aqi values for the points in the routes
	//     if routePoints[j] == nil || routePoints[j] == nil {
	//         continue
	//     }

	//     aqiData, err := fetchAqiData([routePoints[j][0], routePoints[j][1]])
	//     checkErrNil(err)

	//     fmt.Println(aqiData)
	//     // this will cause error.
	//     totalRouteExposure += (aqiData.data.iaqi.pm25.v) / 36000 // converting time to hours

	//     fmt.Printf("The total Exposure: %v", totalRouteExposure)

	//     append(aqiValues, aqiData)
	// }

	// route.totalExposure = totalRouteExposure
	// route.aqiValues = aqiValues
	// fmt.Println("The total exposure for the route is: ", totalRouteExposure)
	// return route
}

// func checkErrNil(err error) {
// 	if err != nil {
// 		log.Fatalf("Error encountered: %s", err)
// 	}
// }
