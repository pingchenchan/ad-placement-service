package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pingchenchan/ad-placement-service/db"
	"github.com/pingchenchan/ad-placement-service/handlers"
    "github.com/pingchenchan/ad-placement-service/models"
)


func init() {
    // Set client options
    err := db.ConnectMongoDB("mongodb://admin:admin@mongo:27017")
    if err != nil {
        log.Fatal(err)
    }

    _, err = db.EnsureCollectionAndIndexes("advertising", "ads")
    if err != nil {
        log.Fatal(err)
    }

    err = db.ConnectRedis("redis:6379")
    if err != nil {
        log.Fatal(err)
    }
}
func main() {
   
    router := gin.Default()
    countryCodeValidator, err := models.LoadCountryCodes("./countryCode.json")
    if err != nil {
        log.Fatalf("Failed to create validator: %v", err)
    }
    
    // Routes

    router.POST("/ads", func(c *gin.Context) {
        handlers.CreateAd(c, countryCodeValidator)
    })
    router.POST("/bulkAds", func(c *gin.Context) {
        handlers.CreateBulkAd(c, countryCodeValidator)
    })
    router.GET("/ads", handlers.GetAds)
    router.GET("/adsRedix", handlers.GetAdsWRedis)

	router.GET("/ad", handlers.GetAd)
    // Start server
    router.Run(":8080")
}
