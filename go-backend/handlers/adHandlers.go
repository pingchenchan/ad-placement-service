package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/pingchenchan/ad-placement-service/db"
	"github.com/pingchenchan/ad-placement-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdWithCtx struct {
	Ad   models.Ad
	Ctx  *gin.Context
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

					writeBulkAds(ads)
					// Clear the slice
					ads = make([]AdWithCtx, 0)

				}
			case adWithCtx := <-adChannel:
				// When a new Ad object is received, add it to the slice
				ads = append(ads, adWithCtx)
			}
		}
	}()
}

var serverCounter int32

func CreateAd(c *gin.Context, countryCode map[string]bool) {
	var ad models.Ad

	// Bind the POST data to the 'ad' variable
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, errMsg := models.ValidateAd(ad, countryCode)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	adChannel <- AdWithCtx{
		Ad:   ad,
		Ctx:  c,
		Done: wg.Done,
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

	for _, adWithCtx := range adsWithCtx {
		adWithCtx.Ctx.Status(http.StatusCreated)
		if adWithCtx.Done != nil {
			adWithCtx.Done()
		}
	}
	fmt.Printf("Made %v BulkAds in %v with result%v\n", len(adsWithCtx), duration, result)
}

func GetAd(c *gin.Context) {
	c.String(http.StatusOK, "Hello World")
}

// GetAds retrieves ads based on the given conditions
func GetAds(c *gin.Context) {
    var params models.AdQueryParams
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	log.Print(params)

    collection := db.DB.Database("advertising").Collection("ads")

    ads, err := queryAds(collection, params.Offset, params.Limit, params.Age, params.Gender, params.Country, params.Platform)
    log.Print(ads)
	if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while retrieving ads"})
        return
    }

    c.JSON(http.StatusOK, ads)
}

func queryAds(collection *mongo.Collection, offset, limit int, age int, gender, country, platform string) ([]*models.Ad, error) {
	// Build the query
	query := bson.D{}
	if age != 0 {
		query = append(query, bson.E{Key: "conditions", Value: bson.D{{Key: "$elemMatch", Value: bson.D{
			{Key: "ageStart", Value: bson.D{{Key: "$lte", Value: age}}},
			{Key: "ageEnd", Value: bson.D{{Key: "$gte", Value: age}}},
		}}}})
	}
	if gender != "" {
		query = append(query, bson.E{Key: "conditions", Value: bson.D{{Key: "$elemMatch", Value: bson.D{
			{Key: "gender", Value: gender},
		}}}})
	}
	if country != "" {
		query = append(query, bson.E{Key: "conditions", Value: bson.D{{Key: "$elemMatch", Value: bson.D{
			{Key: "country", Value: bson.D{{Key: "$in", Value: []string{country}}}},
		}}}})
	}
	if platform != "" {
		query = append(query, bson.E{Key: "conditions", Value: bson.D{{Key: "$elemMatch", Value: bson.D{
			{Key: "platform", Value: bson.D{{Key: "$in", Value: []string{platform}}}},
		}}}})
	}

	// Execute the query
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "endAt", Value: 1}})
	cursor, err := collection.Find(context.Background(), query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Decode the results
	var results []*models.Ad
	if err := cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
