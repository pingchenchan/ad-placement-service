package main

import (
	"bytes"
	"encoding/json"
	"math"

	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingchenchan/ad-placement-service/handlers"
	"github.com/pingchenchan/ad-placement-service/models"

)

func createSampleAdJSON() ([]byte, error){
    ad := createSampleAd()
    adData, err := json.Marshal(ad)
    if err != nil {
        return nil, err
    }
    return adData, nil
}

func CreateAdFunction(endpoint string, adData []byte) (*http.Response, error) {
    req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(adData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    return client.Do(req)
}
func createAdRequest(t *testing.T, router *gin.Engine, ad models.Ad) int {
    // Convert the ad to JSON
    adJson, _ := json.Marshal(ad)

    // Create a request to pass to our handler
    req, _ := http.NewRequest(http.MethodPost, "/ads", bytes.NewBuffer(adJson))

    // Create a ResponseRecorder to record the response
    rr := httptest.NewRecorder()

    // Perform the request
    router.ServeHTTP(rr, req)

    // Check the status code
    return rr.Code
}

func createSampleAd() models.Ad {
    // Create a sample ad
    uniqueID := time.Now().UnixNano() // Generate a unique ID based on the current time
    ad := models.Ad{
        Title: "Sample Ad" + strconv.FormatInt(uniqueID, 10),
        StartAt: time.Now(),
        EndAt: time.Now().Add(24 * time.Hour),
        Conditions: []models.Condition{
            {
                AgeStart: new(int),
                AgeEnd:   new(int),
                Gender:   new(string),
                Country:  []string{"TW", "JP"},
                Platform: []string{"ios", "web"},
            },
        },
    }

    *ad.Conditions[0].AgeStart = 25
    *ad.Conditions[0].AgeEnd = 35
    *ad.Conditions[0].Gender = "M"

    return ad
}

func TestCreateBulkAd(t *testing.T) {
    // Create a Gin router
    // gin.SetMode(gin.TestMode)
    gin.SetMode(gin.ReleaseMode)
    router := gin.Default()
    countryCodeValidator, err := models.LoadCountryCodes("./models/countryCode.json")

    if err != nil {
        log.Fatalf("Failed to create validator: %v", err)
    }

    // Define the route similar to your actual application
    router.POST("/ads", func(c *gin.Context) {
        handlers.CreateBulkAd(c, countryCodeValidator)
    })

    ad := createSampleAd()
    if status := createAdRequest(t, router, ad); status != http.StatusCreated {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
    }
}

func testCreateAd(t *testing.T, handlerFunc gin.HandlerFunc, endpoint string) {
    // Create a Gin router
    gin.SetMode(gin.ReleaseMode)
    router := gin.Default()


    // Define the route similar to your actual application
    router.POST("/ads", handlerFunc)

    var wg sync.WaitGroup
    wg.Add(10000)

    var successCount int64 = 0
    var failCount int64 = 0
    timeChannel := make(chan time.Duration, 10000)
    // Record the start time
    start := time.Now()

    for i := 0; i < 10000; i++ {
        go func() {
            defer wg.Done()

            // Create a sample ad
            ad := createSampleAd()
            startTime := time.Now()
            // Check the status code
            if status := createAdRequest(t, router, ad); status != http.StatusCreated {
                atomic.AddInt64(&failCount, 1)
            } else {
                atomic.AddInt64(&successCount, 1)
            }
            timeChannel <- time.Since(startTime)
        }()
    }

    // Wait for all goroutines to finish
    wg.Wait()
    close(timeChannel)
    // Record the end time
    end := time.Now()

    // Calculate the total response time
    totalResponseTime := end.Sub(start)

    t.Logf("Numbers of successful requests: %v", atomic.LoadInt64(&successCount))
    t.Logf("Numbers of error requests: %v", atomic.LoadInt64(&failCount))
    t.Logf("Total response time: %v", totalResponseTime)

    // Calculate the average, max, and min request time
    logRequestStats(t, timeChannel, endpoint, start)
}

func TestCreate10000Ad(t *testing.T) {
    countryCodeValidator, err := models.LoadCountryCodes("./models/countryCode.json")
    if err != nil {
        log.Fatalf("Failed to create validator: %v", err)
    }
    testCreateAd(t, func(c *gin.Context) {
        handlers.CreateAd(c, countryCodeValidator)
    }, "create Ad")
}

func TestCreate10000BulkAd(t *testing.T) {
    countryCodeValidator, err := models.LoadCountryCodes("./models/countryCode.json")
    if err != nil {
        log.Fatalf("Failed to create validator: %v", err)
    }
    testCreateAd(t, func(c *gin.Context) {
        handlers.CreateBulkAd(c, countryCodeValidator)
    }, "create Bulk Ad")
}



func TestCreate10000AdHttp(t *testing.T) {
     create10000Ad(t, "http://go-backend:8080/ads")

}
func TestCreate10000bulkAdHttp(t *testing.T) {
    create10000Ad(t, "http://go-backend:8080/bulkAds")
}

func create10000Ad(t *testing.T, endpoint string) {
    start := time.Now()

    // Create a channel to handle errors
    errChannel := make(chan error, 10000)

    // Create a channel to handle success
    successChannel := make(chan struct{}, 10000)

    // Create a channel to collect request times
    timeChannel := make(chan time.Duration, 10000)

    var wg sync.WaitGroup
    wg.Add(10000)

    for i := 0; i < 10000; i++ {
        go func() {
            defer wg.Done()

            // Record the start time of the request
            adData, err := createSampleAdJSON()
            
            startTime := time.Now()

            _, err = CreateAdFunction(endpoint,adData)

            // Record the end time of the request and send the duration to the time channel
            timeChannel <- time.Since(startTime)

            if err != nil {
                errChannel <- err
            } else {
                successChannel <- struct{}{}
            }
        }()
    }

    // Wait for all goroutines to finish
    wg.Wait()
    close(errChannel)
    close(successChannel)
    close(timeChannel)



    t.Logf("")
    // Check how many errors
    t.Logf("Numbers of error requests: %v", len(errChannel))
    // Check how many successes
    t.Logf("Numbers of successful requests: %v", len(successChannel))

    // Check if there were any errors
    for err := range errChannel {
        if err != nil {
            t.Logf("error, %v", err)
        }
    }

    // Calculate the average, max, and min request time
    logRequestStats(t, timeChannel, endpoint, start)
}

func logRequestStats(t *testing.T, timeChannel <-chan time.Duration, endpoint string, start time.Time) {
    // Calculate the average, max, and min request time
    var total time.Duration
    var max time.Duration
    var min time.Duration = time.Duration(math.MaxInt64)
    var count int
    for requestTime := range timeChannel {
        total += requestTime
        if requestTime > max {
            max = requestTime
        }
        if requestTime < min {
            min = requestTime
        }
        count++
    }
    average := total / time.Duration(count)

    t.Logf("Endpoint: %v", endpoint)

    t.Logf("Average request time: %v", average)
    t.Logf("Max request time: %v", max)
    t.Logf("Min request time: %v", min)

    duration := time.Since(start)
    t.Logf("Made 10000 POST requests in %v", duration)
}