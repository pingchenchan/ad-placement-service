package main

import (
	"github.com/pingchenchan/ad-placement-service/db"
	"github.com/pingchenchan/ad-placement-service/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
    db.ConnectDB()

    router := gin.Default()

    // Routes
    router.POST("/ads", handlers.CreateAd)
    router.GET("/ads", handlers.GetAds)

	router.GET("/ad", handlers.GetAd)
    // Start server
    router.Run(":8080")
}
