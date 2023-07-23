package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
	fmt.Print(queryData) // check if you're receiving the data that you're supposed to get...
	c.IndentedJSON(http.StatusOK, books)
}

func findAllRoutes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, books)
}

func main() {
	router := gin.Default()
	// router.GET("/books", getBooks)
	router.POST("/route", findRoute)
	router.GET("all-routes", findAllRoutes)
	router.Run("localhost:8080")
}
