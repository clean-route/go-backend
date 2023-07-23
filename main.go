package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	mapbox "github.com/sadityakumar9211/clean-route-backend/models/mapbox"
	graphhopper "github.com/sadityakumar9211/clean-route-backend/models/graphhopper"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type formData struct {
	Source          [2]float64 `json:"source"`
	Destination     [2]float64 `json:"destination"`
	DelayCode       uint8      `json:"delay_code"`
	Mode            string     `json:"mode"`
	RoutePreference string     `json:"route_preference,omitempty"`
}

type book struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var books = []book{
	{ID: "1", Name: "Hello"},
	{ID: "2", Name: "BOLO"},
}

// func getBooks(c *gin.Context) {
// 	c.IndentedJSON(http.StatusOK, books)
// }

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

	// if mode is driving-traffic & route preference is shortest, fastest, balanced, then only we will call Mapbox API otherwise,
	// we will call Graphhopper API

	// Calling Mapbox API
	if mode == "driving-traffic" && (routePref == "shortest" || routePref == "fastest" || routePref == "balanced") {
		// constructing the URL
		baseUrl := "https://api.mapbox.com/directions/v5/mapbox/driving-traffic/" + fmt.Sprintf("%f,%f;%f,%f", source[0], source[1], destination[0], destination[1])

		mapboxAccessToken, ok := viper.Get("MAPBOX_API_KEY").(string)
		if !ok {
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

		fmt.Println(params)

		url := baseUrl + "?" + params.Encode()

		// Making the API call
		resp, err := http.Get(url)
		checkErr(err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		checkErr(err)

		var routes mapbox.Routes

		err = json.Unmarshal([]byte(body), &routes)
		if err != nil {
			fmt.Println("Error while unmarshaling JSON:", err)
			return
		}

		// Access and work with the unmarshaled data
		fmt.Println("Distance", routes.Routes[0].Distance)
		fmt.Println("Code:", routes.Code)
		fmt.Println("UUID:", routes.UUID)
		c.IndentedJSON(http.StatusOK, routes.Routes)
		return
	} else {
		baseUrl := "https://graphhopper.com/api/1/route?"

		graphhopperApikey := viper.Get("GRAPHHOPPER_API_KEY").(string)

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

		// Making the API call
		resp, err := http.Get(url)
		checkErr(err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		checkErr(err)

		var routes graphhopper.RouteData

		err = json.Unmarshal([]byte(body), &routes)
		if err != nil {
			fmt.Println("Error while unmarshaling JSON:", err)
			return
		}

		// Access and work with the unmarshaled data
		fmt.Println("Total Paths: ", len(routes.Paths))
		fmt.Println("Distance:", routes.Paths[0].Distance)
		fmt.Println("UUID:", routes.Paths[0].Time)
		c.IndentedJSON(http.StatusOK, routes.Paths)
		return		
	}
	// c.IndentedJSON(http.StatusOK, queryData)
}

func findAllRoutes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, books)
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

func checkErr(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
