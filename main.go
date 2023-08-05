package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	graphhopper "github.com/sadityakumar9211/clean-route-backend/models/graphhopper"
	mapbox "github.com/sadityakumar9211/clean-route-backend/models/mapbox"
	"github.com/sadityakumar9211/clean-route-backend/utils"
	"github.com/spf13/viper"
)

type formData struct {
	Source          [2]float64 `json:"source"`
	Destination     [2]float64 `json:"destination"`
	DelayCode       uint8      `json:"delay_code"`
	Mode            string     `json:"mode"`
	RoutePreference string     `json:"route_preference,omitempty"`
}

func findMapboxRoute(source [2]float64, destination [2]float64, delayCode uint8) mapbox.RouteData {
	baseUrl := "https://api.mapbox.com/directions/v5/mapbox/driving-traffic/" + fmt.Sprintf("%f,%f;%f,%f", source[0], source[1], destination[0], destination[1])

	mapboxAccessToken, mapboxAccessTokenError := viper.Get("MAPBOX_API_KEY").(string)
	if !mapboxAccessTokenError {
		log.Fatalf("Invalid type assertion")
	}

	localTime := time.Now()
	departureTime := localTime.Add(30 * time.Duration(delayCode) * time.Minute).Format("2006-01-02T15:04")

	params := url.Values{}
	params.Add("steps", "true")
	params.Add("geometries", "geojson")
	params.Add("alternatives", "true")
	params.Add("waypoints_per_route", "true")
	params.Add("access_token", mapboxAccessToken)
	params.Add("depart_at", departureTime)

	url := baseUrl + "?" + params.Encode()

	resp, err := http.Get(url)
	checkErrNil(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	checkErrNil(err)

	var routes mapbox.RouteData

	err = json.Unmarshal([]byte(body), &routes)
	if err != nil {
		log.Fatal("Error while unmarshling JSON: ", err)
	}

	fmt.Println("Distance", routes.Routes[0].Distance)
	fmt.Println("Total Exposure: ", routes.Routes[0].TotalExposure)
	fmt.Println("Total Energy: ", routes.Routes[0].TotalEnergy)
	fmt.Println("Code:", routes.Code)
	fmt.Println("UUID:", routes.UUID)
	return routes
}

func findGraphhopperRoute(source [2]float64, destination [2]float64, mode string) graphhopper.RouteData {
	baseUrl := "https://graphhopper.com/api/1/route?"

	graphhopperApikey, graphhopperApikeyError := viper.Get("GRAPHHOPPER_API_KEY").(string)
	if !graphhopperApikeyError {
		log.Fatal("Found GraphHopper API key: ", graphhopperApikeyError)
	}

	params := url.Values{}
	params.Add("point", fmt.Sprintf("%f,%f", source[1], source[0]))
	params.Add("point", fmt.Sprintf("%f,%f", destination[1], destination[0]))
	params.Add("vehicle", mode)
	params.Add("debug", "true")
	params.Add("key", graphhopperApikey)
	params.Add("type", "json")
	params.Add("points_encoded", "false")
	params.Add("algorithm", "alternative_route")
	params.Add("alternative_route.max_paths", "4")
	params.Add("alternative_route.max_weight_factor", "1.4")
	params.Add("alternative_route.max_share_factor", "0.6")
	params.Add("elevation", "true")

	url := baseUrl + params.Encode()

	resp, err := http.Get(url)
	checkErrNil(err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	checkErrNil(err)

	var routes graphhopper.RouteData

	err = json.Unmarshal([]byte(body), &routes)
	if err != nil {
		log.Fatalf("Error while unmarshling JSON: %s", err)
	}

	if len(routes.Paths) == 0 {
		return routes
	}

	// Access and work with the unmarshaled data
	fmt.Println("Total Paths: ", len(routes.Paths))
	fmt.Println("Distance:", routes.Paths[0].Distance)
	fmt.Println("UUID:", routes.Paths[0].Time)
	return routes
}

func findRoute(c *gin.Context) {
	var queryData formData
	if err := c.BindJSON(&queryData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(queryData)

	source := queryData.Source
	destination := queryData.Destination
	delayCode := queryData.DelayCode
	mode := queryData.Mode
	routePref := queryData.RoutePreference

	if mode == "driving-traffic" && (routePref == "fastest" || routePref == "balanced") {
		// Calling Mapbox API
		var routes mapbox.RouteData = findMapboxRoute(source, destination, delayCode)
		// Finding the energy of each routes and exposure
		if mode == "driving-traffic" {
			mode = "car"
		}
		var energy_route graphhopper.RouteData = findGraphhopperRoute(source, destination, mode)

		sort.SliceStable(energy_route.Paths, func(i, j int) bool {
			return energy_route.Paths[i].Time < energy_route.Paths[j].Time
		})

		sort.SliceStable(routes.Routes, func(i, j int) bool {
			return routes.Routes[i].Duration < routes.Routes[j].Duration
		})

		// Exposure and Energy
		for i := 0; i < len(routes.Routes); i++ {
			routes.Routes[i] = utils.CalculateRouteExposureMapbox(routes.Routes[i], delayCode)
			routes.Routes[i].Duration *= 1000
			routes.Routes[i].TotalEnergy = utils.CalculateRouteEnergy(energy_route.Paths[i], mode)
			fmt.Println("Total Energy: ", routes.Routes[i].TotalEnergy)
			fmt.Println("Distance: ", routes.Routes[i].Distance)
			fmt.Println("Duration: ", routes.Routes[i].Duration)
		}
		// fmt.Println("I was here...")
		// Perform calculations and return the best path
		if routePref == "fastest" {
			c.IndentedJSON(http.StatusOK, routes.Routes[0]) // returning the fastest route
			return
		} else if routePref == "balanced" {
			// If we have only one path in the result of API call
			if len(routes.Routes) == 1 {
				c.IndentedJSON(http.StatusOK, routes.Routes[0])
				return
			}

			// If we have two paths (max paths in case of mapbox)
			if len(routes.Routes) == 2 {
				if routes.Routes[0].Duration-routes.Routes[1].Duration < 5*60*1000 && routes.Routes[0].Distance-routes.Routes[1].Distance < 500 {
					// return path with least exposure
					if routes.Routes[0].TotalExposure < routes.Routes[1].TotalExposure {
						c.IndentedJSON(http.StatusOK, routes.Routes[0])
					} else {
						c.IndentedJSON(http.StatusOK, routes.Routes[1])
					}
				} else {
					if routes.Routes[0].Duration < routes.Routes[1].Duration {
						c.IndentedJSON(http.StatusOK, routes.Routes[0])
					} else {
						c.IndentedJSON(http.StatusOK, routes.Routes[1])
					}
				}
				return
			}
		}
		c.IndentedJSON(http.StatusBadRequest, "Error: Incorrect route preference suspected.")
	} else {
		// Calling GraphHopper API
		var routes graphhopper.RouteData = findGraphhopperRoute(source, destination, mode)
		for i := 0; i < len(routes.Paths); i++ {
			routes.Paths[i] = utils.CalculateRouteExposureGraphhopper(routes.Paths[i], delayCode)
			routes.Paths[i].TotalEnergy = utils.CalculateRouteEnergy(routes.Paths[i], mode)
		}

		// Perform calculation and return the best path
		if routePref == "shortest" {
			// sort the routes with distance and return the shortest path.
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].Distance < routes.Paths[j].Distance
			})
			// fmt.Println(routes)
			c.IndentedJSON(http.StatusOK, routes.Paths[0])
			return
		} else if routePref == "fastest" {
			// sort the routes with time and return the fastest path
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].Time < routes.Paths[j].Time
			})
			c.IndentedJSON(http.StatusOK, routes.Paths[0])
			return
		} else if routePref == "leap" {
			// sort the routes with exposure and return the leap path
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].TotalExposure < routes.Paths[j].TotalExposure
			})
			// fmt.Println(routes)
			c.IndentedJSON(http.StatusOK, routes.Paths[0])
			return
		} else if routePref == "emission" {
			// sort the routes with total energy and return the lco2 path
			sort.SliceStable(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].TotalEnergy < routes.Paths[j].TotalEnergy
			})
			c.IndentedJSON(http.StatusOK, routes.Paths[0])
			return
		} else if routePref == "balanced" {
			// check it out how it is implemented...
			// First find the two fastest paths and then find the path with the smallest exposure

			// If we have only one path in the result of API call
			if len(routes.Paths) == 1 {
				c.IndentedJSON(http.StatusOK, routes.Paths[0])
				return
			}

			// If we have two paths
			if len(routes.Paths) == 2 {
				if routes.Paths[0].Time-routes.Paths[1].Time < 5*60*1000 && routes.Paths[0].Distance-routes.Paths[1].Distance < 500 {
					// return path with least exposure
					if routes.Paths[0].TotalExposure < routes.Paths[1].TotalExposure {
						c.IndentedJSON(http.StatusOK, routes.Paths[0])
					} else {
						c.IndentedJSON(http.StatusOK, routes.Paths[1])
					}
				} else {
					if routes.Paths[0].Time < routes.Paths[1].Time {
						c.IndentedJSON(http.StatusOK, routes.Paths[0])
					} else {
						c.IndentedJSON(http.StatusOK, routes.Paths[1])
					}
				}

				return
			}

			// sorting the top three routes based on exposure
			sort.Slice(routes.Paths, func(i, j int) bool {
				return routes.Paths[i].TotalExposure > routes.Paths[i].TotalExposure
			})

			// sorting all the routes based on time
			sort.SliceStable(routes.Paths[:3], func(i, j int) bool {
				return routes.Paths[i].Time > routes.Paths[j].Time
			})

			// sorting the top two balanced(time, exposure) routes with energy
			sort.Slice(routes.Paths[:2], func(i, j int) bool {
				return routes.Paths[i].TotalEnergy > routes.Paths[i].TotalEnergy
			})

			c.IndentedJSON(http.StatusOK, routes.Paths[0])
			return
		}
		c.IndentedJSON(http.StatusBadRequest, "Error: Incorrect route preference suspected.")
	}
}

//

func findAllRoutes(c *gin.Context) {
	// c.IndentedJSON(http.StatusOK, books)
}

func main() {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}

	router := gin.Default()
	// router.GET("/books", getBooks)
	router.POST("/route", findRoute)
	router.GET("all-routes", findAllRoutes)
	router.Run("localhost:8080")
}

func checkErrNil(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
