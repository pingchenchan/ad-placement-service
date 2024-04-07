package main

import (
	"bytes"
	"encoding/json"
	"math"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingchenchan/ad-placement-service/handlers"
	"github.com/pingchenchan/ad-placement-service/models"
)


// createAdFunction sends a POST request to create an ad.
func createAdFunction(endpoint string, adData []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(adData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	return client.Do(req)
}

// createAdRequest creates a POST request to the specified router with the ad data.
func createAdRequest(t *testing.T, router *gin.Engine, ad models.Ad) int {
	adJSON, _ := json.Marshal(ad)
	req, _ := http.NewRequest(http.MethodPost, "/dummy", bytes.NewBuffer(adJSON))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr.Code
}


// testCreateAd tests the ad creation process in a concurrent environment.
func testCreateAd(t *testing.T, handlerFunc gin.HandlerFunc, endpoint string) {
	router := gin.Default()
    router.POST("/dummy", handlerFunc)
	var wg sync.WaitGroup
	wg.Add(10000)

	var successCount int64 = 0
	var failCount int64 = 0
	timeChannel := make(chan time.Duration, 10000)
    var firstRequestTime, lastRequestTime time.Time
    sampleAds :=  GenerateAds(10000)
	start := time.Now()


	for i := 0; i < 10000; i++ {
		go func(i int) {
			defer wg.Done()
            ad := sampleAds[i]
            startTime := time.Now()
            if i == 0 {
                firstRequestTime = startTime
            } else if i == 9999 {
                lastRequestTime = startTime
            }
            if status := createAdRequest(t, router, ad); status != http.StatusCreated {
                atomic.AddInt64(&failCount, 1)
            } else {
                atomic.AddInt64(&successCount, 1)
            }
			timeChannel <- time.Since(startTime)
		}(i)
	}

	wg.Wait()
	close(timeChannel)
	t.Logf("Numbers of successful requests: %v", atomic.LoadInt64(&successCount))
	t.Logf("Numbers of error requests: %v", atomic.LoadInt64(&failCount))
    t.Logf("Time from first to last request: %v", lastRequestTime.Sub(firstRequestTime))

	logRequestStats(t, timeChannel, endpoint, start)
}

// logRequestStats logs the statistics of the request times.
func logRequestStats(t *testing.T, timeChannel <-chan time.Duration, endpoint string, start time.Time) {
	var total, max time.Duration
	min := time.Duration(math.MaxInt64)
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
	t.Logf("Total duration of requests: %v", time.Since(start))
}

// TestCreate10000Ad tests the creation of 10000 individual ads.
func TestCreateAd_Concurrency(t *testing.T) {
    countryCodeValidator, err := models.LoadCountryCodes("./models/countryCode.json")
    if err != nil {
        log.Fatalf("Failed to load country codes: %v", err)
    }
    testCreateAd(t, func(c *gin.Context) {
        handlers.CreateAd(c, countryCodeValidator)
    }, "/ads")
}

// TestCreate10000BulkAd tests the bulk creation of 10000 ads.
func TestCreateAd_BulkWrite_Concurrency(t *testing.T) {
    countryCodeValidator, err := models.LoadCountryCodes("./models/countryCode.json")
    if err != nil {
        log.Fatalf("Failed to load country codes: %v", err)
    }
    testCreateAd(t, func(c *gin.Context) {
        handlers.CreateBulkAd(c, countryCodeValidator)
    }, "/bulkAds")
}

// TestCreate10000AdHttp tests the HTTP endpoint for creating 10000 individual ads.
func TestCreateAd_HTTP_Endpoint_Concurrency(t *testing.T) {
    create10000Ad(t, "http://go-backend:8080/ads")
}

// TestCreate10000BulkAdHttp tests the HTTP endpoint for bulk creating 10000 ads.
func TestCreateAd_BulkWrite_HTTP_Endpoint_Concurrency(t *testing.T) {
    create10000Ad(t, "http://go-backend:8080/bulkAds")
}

// create10000Ad is a helper function to create 10000 ads using the specified endpoint.
func create10000Ad(t *testing.T, endpoint string) {
    var wg sync.WaitGroup
    wg.Add(10000)

    errChannel := make(chan error, 10000)
    successChannel := make(chan struct{}, 10000)
    timeChannel := make(chan time.Duration, 10000)
    var firstRequestTime, lastRequestTime time.Time
    sampleAds :=  GenerateAds(10000)
    start := time.Now()

    for i := 0; i < 10000; i++ {
        go func(i int) {
            defer wg.Done()
            adData, err := json.Marshal(sampleAds[i]) 
            if err != nil {
                errChannel <- err
                return
            }

            startTime := time.Now()
            if i == 0 {
                firstRequestTime = startTime
            }else if i == 9999 {
                lastRequestTime = startTime
            }

            _, err = createAdFunction(endpoint, adData)
            timeChannel <- time.Since(startTime)

            if err != nil {
                errChannel <- err
            } else {
                successChannel <- struct{}{}
            }
        }(i)
    }

    wg.Wait()
    close(errChannel)
    close(successChannel)
    close(timeChannel)

    t.Logf("Numbers of error requests: %v", len(errChannel))
    t.Logf("Numbers of successful requests: %v", len(successChannel))
    t.Logf("Time from first to last request: %v", lastRequestTime.Sub(firstRequestTime))
    for err := range errChannel {
        if err != nil {
            t.Logf("Error: %v", err)
        }
    }

    logRequestStats(t, timeChannel, endpoint, start)
}