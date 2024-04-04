package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingchenchan/ad-placement-service/db"
	"github.com/pingchenchan/ad-placement-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdWithCtx struct {
    Ad  models.Ad
    Ctx *gin.Context
    Done func()
}
// Create a channel to store Ad objects
var adChannel = make(chan AdWithCtx, 10000)
// var adChannel = make(chan models.Ad, 10000)

func init() {
    // Start a new goroutine
    go func() {
        // Create a ticker that fires every 0.3 seconds
        ticker := time.NewTicker(300 * time.Millisecond)
        defer ticker.Stop()

        // Create a slice to store Ad objects
        ads := make([]AdWithCtx, 0)


        for {
            select {
            case <-ticker.C:
                // Every 0.3 seconds, write all Ad objects in the slice to the database
                
                if len(ads) > 0 {

                    //print counter
                    fmt.Printf("bulk serverCounter: %v\n", serverCounter)



                    fmt.Printf("bulk len(ads): %v\n", len(ads))
                    writeBulkAds( ads)
                    // Clear the slice
                    ads = make([]AdWithCtx, 0)
                    //print the length of the slice
                    fmt.Printf("len(ads): %v\n", len(ads))
                    
                }
            case adWithCtx := <-adChannel:
                // When a new Ad object is received, add it to the slice
                ads = append(ads, adWithCtx)
                // fmt.Printf("***len(ads): %v\n", len(ads))
            }
        }
    }()
}
var serverCounter int32
func CreateAd(c *gin.Context) {
    var ad models.Ad
    if err := c.ShouldBindJSON(&ad); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var wg sync.WaitGroup
    wg.Add(1)
    adChannel <- AdWithCtx{
        Ad: ad,
        Ctx: c,
        Done:wg.Done,
    }
    atomic.AddInt32(&serverCounter, 1)


    wg.Wait()

}


func writeBulkAds(adsWithCtx []AdWithCtx) {
    start := time.Now()
    models := make([]mongo.WriteModel, len(adsWithCtx))
    for i, adWithCtx := range adsWithCtx {
        models[i] = mongo.NewInsertOneModel().SetDocument(adWithCtx.Ad)
    }

    collection := db.DB.Database("advertising").Collection("ads")
    result, err := collection.BulkWrite(context.Background(), models)
    duration := time.Since(start)
    fmt.Printf("Made %v BulkAds in %v\n", len(adsWithCtx), duration)

    if err != nil {
        fmt.Println("Error while inserting ads in bulk:", err)
        for _, adWithCtx := range adsWithCtx {
            adWithCtx.Ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            if adWithCtx.Done != nil {
                adWithCtx.Done()
            }
        }
        return
    }

    fmt.Println("Successful bulk insertion.", result)
    for _, adWithCtx := range adsWithCtx {
        adWithCtx.Ctx.Status(http.StatusCreated)
        if adWithCtx.Done != nil {
            adWithCtx.Done()
        }
    }

    duration2 := time.Since(start)
    fmt.Printf("Made %v BulkAds in %v After res\n", len(adsWithCtx), duration2,)
}

func GetAd(c *gin.Context) {
	//return hello world
 	c.String(http.StatusOK, "Hello World")
}
// GetAds retrieves ads based on the given conditions
func GetAds(c *gin.Context) {
    var condition models.Condition
    if err := c.ShouldBindQuery(&condition); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Here you'd add logic to construct a query based on the conditions received
    // For example, you might search for ads where StartAt is before now, and EndAt is after now
    // And then filter further based on other conditions such as age, country, etc.

    collection := db.DB.Database("advertising").Collection("ads")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    ads := []models.Ad{}
    cursor, err := collection.Find(ctx, bson.M{"startAt": bson.M{"$lte": time.Now()}, "endAt": bson.M{"$gte": time.Now()}})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while retrieving ads"})
        return
    }

    if err = cursor.All(ctx, &ads); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while reading ads"})
        return
    }

    c.JSON(http.StatusOK, ads)
}
