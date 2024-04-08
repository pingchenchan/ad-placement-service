package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/pingchenchan/ad-placement-service/db"
	"github.com/pingchenchan/ad-placement-service/handlers"
	"github.com/pingchenchan/ad-placement-service/models"
)

// init function initializes the MongoDB and Redis connections
func init() {

	// Connect to MongoDB
	err := db.ConnectMongoDB("mongodb://admin:admin@mongo:27017")
	if err != nil {
		log.Fatal(err)
	}
	
	// Ensure the collection and indexes exist in MongoDB
	_, err = db.EnsureCollectionAndIndexes("advertising", "ads")
	if err != nil {
		log.Fatal(err)
	}

	// Connect to Redis
	err = db.ConnectRedis("redis:6379")
	if err != nil {
		log.Fatal(err)
	}
}

// main function sets up the routes and starts the server
func main() {
    // Create a new gin router
	router := gin.Default()

	// Load the country codes for validation
	countryCodeValidator, err := models.LoadCountryCodes("./countryCode.json")
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	 // Start the ad handler
	handlers.StartAdHandler()

	// Set up the routes
	router.POST("/ads", func(c *gin.Context) {
		handlers.CreateAd(c, countryCodeValidator)
	})
	router.POST("/adsBulk", func(c *gin.Context) {
		handlers.CreateBulkAd(c, countryCodeValidator)
	})

	router.POST("/adsAsync", func(c *gin.Context) {
		handlers.CreateAsyncAd(c, countryCodeValidator)
	})
	router.GET("/ads", handlers.GetAds)
	router.GET("/adsRedisStringParams", handlers.GetadsRedisStringParams)
	router.GET("/adsRedisActiveDocs", handlers.GetAdsWRedisActiveDocs)

	// Start server
	router.Run(":8080")
}
