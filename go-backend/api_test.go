package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
                Platform: []string{"iOS", "Android"},
            },
        },
    }

    *ad.Conditions[0].AgeStart = 25
    *ad.Conditions[0].AgeEnd = 35
    *ad.Conditions[0].Gender = "Male"

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
// func TestCreateAd(t *testing.T) {
//     resp, err := CreateAdFunction()
//     if err != nil {
//         t.Fatal("创建广告时发生错误:", err)
//     }

//     assert.Equal(t, http.StatusCreated, resp.StatusCode)
// }

// func TestGetAd(t *testing.T) {
// 	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/ad", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.Equal(t, http.StatusOK, resp.StatusCode)
// }



// func GetAdsFunction() (*http.Response, error) {
//     req, err := http.NewRequest("GET", "http://127.0.0.1:8080/ads", nil)
//     if err != nil {
//         return nil, err
//     }

//     client := &http.Client{}
//     return client.Do(req)
// }


// func TestGetAds(t *testing.T) {
//     // 同前面的TestGetAds函数
//     // 在这里调用 GetAdsFunction()
//     resp, err := GetAdsFunction()
//     if err != nil {
//         t.Fatal("获取广告列表时发生错误:", err)
//     }

//     assert.Equal(t, http.StatusOK, resp.StatusCode)
// }
// func TestGetAdsPerformance(t *testing.T) {
//     start := time.Now()
//     count := 0
//     for time.Since(start) < time.Second {
//         _, err := GetAdsFunction()
//         if err != nil {
//             t.Fatal("获取广告列表时发生错误:", err)
//         }
//         count++
//     }
//     t.Logf("Made %d GET requests in 1 second", count)
// }

// func TestCreateAdPerformance(t *testing.T) {
//     start := time.Now()
//     count := 0
//     for time.Since(start) < time.Second {
//         _, err := CreateAdFunction()
//         if err != nil {
//             t.Fatal("创建广告时发生错误:", err)
//         }
//         count++
//     }
//     t.Logf("Made %d POST requests in 1 second", count)
// }

// func TestGet1000AdsPerformance(t *testing.T) {
//     start := time.Now()
//     for i := 0; i < 1000; i++ {
//         _, err := GetAdsFunction()
//         if err != nil {
//             t.Fatal("获取广告列表时发生错误:", err)
//         }
//     }
//     duration := time.Since(start)
//     t.Logf("Made 1000 GET requests in %v", duration)
// }

// func TestCreate1000AdPerformance(t *testing.T) {
//     start := time.Now()
//     for i := 0; i < 3000; i++ {
//         _, err := CreateAdFunction()
//         if err != nil {
//             t.Fatal("创建广告时发生错误:", err)
//         }
//     }
//     duration := time.Since(start)
//     t.Logf("Made 1000 POST requests in %v", duration)
// }
var counter int32

func TestCreate10000AdPerformance(t *testing.T) {
    start := time.Now()

    // Create a channel to handle errors
    errChannel := make(chan error, 1000)

    // Create a channel to control the number of concurrent goroutines
    pool := make(chan struct{}, 500) // adjust the size of the pool as needed

    var wg sync.WaitGroup
    wg.Add(500)

    for i := 0; i < 500; i++ {
        pool <- struct{}{} // block if the pool is full

        go func() {
            defer wg.Done()
            defer func() { <-pool }() // release the pool slot

            _, err := CreateAdFunction()
            if err != nil {
                errChannel <- err
                return
            }
        }()
    }

    // Wait for all goroutines to finish
    wg.Wait()
    close(errChannel)

    // Check how many errors
    t.Logf("Numbers of error requests: %v", len(errChannel))
    // Check if there were any errors
    for err := range errChannel {
        if err != nil {
            t.Logf("error, %v", err)
            duration := time.Since(start)
            t.Logf("Made 10000 POST requests in %v", duration)
            return 
        }
    }
}