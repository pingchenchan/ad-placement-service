package main

import (
	"bytes"
	"encoding/json"
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
	// "github.com/stretchr/testify/assert"
)

func CreateAdFunction() (*http.Response, error) {
    uniqueID := time.Now().UnixNano() // Generate a unique ID based on the current time
    ad := models.Ad{
        Title: "Sample Ad"+ strconv.FormatInt(uniqueID, 10),
        StartAt: time.Now(),
        EndAt: time.Now().Add(24 * time.Hour),
        Conditions: []models.Condition{
            {
                AgeStart: new(int),
                AgeEnd:   new(int),
                Gender:   new(string),
                Country:  []string{"TW", "JP"},
                Platform: []string{"ios", "wed"},
            },
        },
    }

    *ad.Conditions[0].AgeStart = 25
    *ad.Conditions[0].AgeEnd = 35
    *ad.Conditions[0].Gender = "M"

    adData, err := json.Marshal(ad)
    if err != nil {
        return nil, err
    }
    atomic.AddInt32(&counter, 1)

    req, err := http.NewRequest("POST", "http://127.0.0.1:8080/ads", bytes.NewBuffer(adData))
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


func TestCreate10000Ad(t *testing.T) {
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
        handlers.CreateAd(c, countryCodeValidator)
    })

    var wg sync.WaitGroup
    wg.Add(10000)

    var successCount int64 = 0
    var failCount int64 = 0

    // Record the start time
    start := time.Now()

    for i := 0; i < 10000; i++ {
        go func() {
            defer wg.Done()

            // Create a sample ad
            ad := createSampleAd()
            // Check the status code
            if status := createAdRequest(t, router, ad); status != http.StatusCreated {
                atomic.AddInt64(&failCount, 1)
            } else {
                atomic.AddInt64(&successCount, 1)
            }
        }()
    }

    // Wait for all goroutines to finish
    wg.Wait()

    // Record the end time
    end := time.Now()

    // Calculate the total response time
    totalResponseTime := end.Sub(start)

    t.Logf("Numbers of successful requests: %v", atomic.LoadInt64(&successCount))
    t.Logf("Numbers of failed requests: %v", atomic.LoadInt64(&failCount))
    t.Logf("Total response time: %v", totalResponseTime)
}

func TestCreate10000BulkAd(t *testing.T) {
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

    var wg sync.WaitGroup
    wg.Add(10000)

    var successCount int64 = 0
    var failCount int64 = 0

    // Record the start time
    start := time.Now()

    for i := 0; i < 10000; i++ {
        go func() {
            defer wg.Done()

            // Create a sample ad
            ad := createSampleAd()
            // Check the status code
            if status := createAdRequest(t, router, ad); status != http.StatusCreated {
                atomic.AddInt64(&failCount, 1)
            } else {
                atomic.AddInt64(&successCount, 1)
            }
        }()
    }

    // Wait for all goroutines to finish
    wg.Wait()

    // Record the end time
    end := time.Now()

    // Calculate the total response time
    totalResponseTime := end.Sub(start)

    t.Logf("Numbers of successful requests: %v", atomic.LoadInt64(&successCount))
    t.Logf("Numbers of failed requests: %v", atomic.LoadInt64(&failCount))
    t.Logf("Total response time: %v", totalResponseTime)
}

var counter int32

// func TestCreate500AdPerformance(t *testing.T) {
//     start := time.Now()

//     // Create a channel to handle errors
//     errChannel := make(chan error, 500)

//     // Create a channel to control the number of concurrent goroutines
//     pool := make(chan struct{}, 500) // adjust the size of the pool as needed

//     var wg sync.WaitGroup
//     wg.Add(500)

//     for i := 0; i < 500; i++ {
//         pool <- struct{}{} // block if the pool is full

//         go func() {
//             defer wg.Done()
//             defer func() { <-pool }() // release the pool slot

//             _, err := CreateAdFunction()
//             if err != nil {
//                 errChannel <- err
//                 return
//             }
//         }()
//     }

//     // Wait for all goroutines to finish
//     wg.Wait()
//     close(errChannel)

//     // Check how many errors
//     t.Logf("Numbers of error requests: %v", len(errChannel))
//     // Check if there were any errors
//     for err := range errChannel {
//         if err != nil {
//             t.Logf("error, %v", err)
//             duration := time.Since(start)
//             t.Logf("Made 10000 POST requests in %v", duration)
//             return 
//         }
//     }
// }

// func TestCreate10000AdPerformance2(t *testing.T) {
//     start := time.Now()

//     // Create a channel to handle errors
//     errChannel := make(chan error, 10000)

//     // Create a channel to handle success
//     successChannel := make(chan struct{}, 10000)

//     var wg sync.WaitGroup
//     wg.Add(10000)

//     for i := 0; i < 10000; i++ {
//         go func() {
//             defer wg.Done()

//             _, err := CreateAdFunction()
//             if err != nil {
//                 errChannel <- err
//             } else {
//                 successChannel <- struct{}{}
//             }
//         }()
//     }

//     // Wait for all goroutines to finish
//     wg.Wait()
//     close(errChannel)
//     close(successChannel)

//     // Check how many errors
//     t.Logf("Numbers of error requests: %v", len(errChannel))
//     // Check how many successes
//     t.Logf("Numbers of successful requests: %v", len(successChannel))

//     // Check if there were any errors
//     for err := range errChannel {
//         if err != nil {
//             t.Logf("error, %v", err)
//         }
//     }

//     duration := time.Since(start)
//     t.Logf("Made 10000 POST requests in %v", duration)
// }