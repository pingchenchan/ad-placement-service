package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
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

func StartAdHandler() {
	// Start a new goroutine
	go func() {
		// Create a ticker that fires every 0.3 seconds
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		// Create a slice to store Ad objects
		ads := make([]AdWithCtx, 0)
		for {
			select {
			case <-ticker.C:
				// Every 0.3 seconds, write all Ad objects in the slice to the database
				ctx := context.Background()
				lenAds, errAds := db.Redis.LLen(ctx, "ads").Result()
				if errAds != nil {
					log.Printf("Failed to get length of ads: %v", errAds)
					continue
			}
				if lenAds > 0 {
					go writeBulkAdsFromRedix()
				}
				if len(ads) > 0 {
					go writeBulkAds(ads)
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

    // Insert the ad into the database
    collection := db.DB.Database("advertising").Collection("ads")
    _, err := collection.InsertOne(context.Background(), ad)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.Status(http.StatusCreated)
}

func CreateAsyncAd(c *gin.Context, countryCode map[string]bool) {
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

	// Convert the ad to JSON
	adJson, err := json.Marshal(ad)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error converting ad to JSON"})
		return
	}

    // Push the ad to a Redis list
	ctx := context.Background()
	err = db.Redis.RPush(ctx, "ads", adJson).Err()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error pushing ad to Redis"})
        return
    }


    c.Status(http.StatusCreated)
}


func CreateBulkAd(c *gin.Context, countryCode map[string]bool) {
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
	fmt.Printf("Made %v Bulk Ads in %v with result%v\n", len(adsWithCtx), duration, result)
}
func redixPopAllAds() ([]string, error) {
    ctx := context.Background()

    // Start a new transaction
    pipe := db.Redis.TxPipeline()

    // Get all elements from the list
    lrange := pipe.LRange(ctx, "ads", 0, -1)

    // Remove all elements from the list
    pipe.LTrim(ctx, "ads", 1, 0)

    // Execute the transaction
    _, err := pipe.Exec(ctx)
    if err != nil {
        return nil, err
    }

    // Return the results
    return lrange.Val(), nil
}
func writeBulkAdsFromRedix(){
	// Read data from Redis
	start := time.Now()
	ctx := context.Background()
	adsJson, err := redixPopAllAds()
	if err != nil {
		log.Printf("Failed to read from Redis: %v", err)
		return
	}
	// Unmarshal the data
    var ads []models.Ad
    for _, adJson := range adsJson {
        var ad models.Ad
        err = json.Unmarshal([]byte(adJson), &ad)
        if err != nil {
            log.Printf("Failed to unmarshal ad: %v", err)
            return
        }
        ads = append(ads, ad)
    }
    // Try to write the data to MongoDB
	for i := 0; i < 3; i++ {
		models := make([]mongo.WriteModel, len(ads))
		for i, ad := range ads {
			models[i] = mongo.NewInsertOneModel().SetDocument(ad)
		}
	
		collection := db.DB.Database("advertising").Collection("ads")
		duration := time.Since(start)
		result, err := collection.BulkWrite(context.Background(), models)
		fmt.Printf("Made %v Bulk Ads in %v with result%v\n", len(ads), duration, result)
		if err == nil {
			return
		}
		
		// If the write operation failed, use exponential backoff
		time.Sleep(time.Second * time.Duration(math.Pow(2, float64(i))))
	}

    // If we've failed to write the data to MongoDB three times, push it back to failAds list in Redis
    err = db.Redis.RPush(ctx, "failAds", adsJson).Err()
    if err != nil {
        log.Printf("Failed to write back to Redis: %v", err)
    }
}

func GetAd(c *gin.Context) {
	c.String(http.StatusOK, "Hello World")
}

// GetAds retrieves ads based on the given conditions
func GetAdsWRedis(c *gin.Context) {
	ctx := context.Background()
    var params models.AdQueryParams
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	// Generate a unique key for this query (store the key in Redis)
    key := fmt.Sprintf("ads:%v", params)
	result, err := db.Redis.Get(context.Background(), key).Result()
	if err != nil {
		// The result is not in Redis, we need to query the database
		collection := db.DB.Database("advertising").Collection("ads")

		ads, err := queryAds(collection, params.Offset, params.Limit, params.Age, params.Gender, params.Country, params.Platform)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while retrieving ads"})
			return
		}
		// Store the result in Redis for future queries
		adsJson, _ := json.Marshal(ads)
		err = db.Redis.Set(ctx, key, adsJson, time.Minute).Err()
		if err != nil {
			log.Printf("Failed to cache the result in Redis: %v", err)
		}

		c.JSON(http.StatusOK, ads)
	}else {
        // The result was in Redis, we can return it directly
        var ads []*models.Ad
        json.Unmarshal([]byte(result), &ads)
        c.JSON(http.StatusOK, ads)
    }
}


func GetAds(c *gin.Context) {

    var params models.AdQueryParams
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

	// log.Print(params)
	// The result is not in Redis, we need to query the database
	collection := db.DB.Database("advertising").Collection("ads")
	ads, err := queryAds(collection, params.Offset, params.Limit, params.Age, params.Gender, params.Country, params.Platform)
	// log.Print(ads)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while retrieving ads"})
		return
	}

	c.JSON(http.StatusOK, ads)

}

func queryAds(collection *mongo.Collection, offset, limit int, age int, gender, country, platform string) ([]*models.Ad, error) {
	// Build the query
	query := bson.D{}
	now := time.Now()
    query = append(query, bson.E{Key: "startAt", Value: bson.D{{Key: "$lte", Value: now}}})
    query = append(query, bson.E{Key: "endAt", Value: bson.D{{Key: "$gte", Value: now}}})
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
