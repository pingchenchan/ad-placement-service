package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pingchenchan/ad-placement-service/db"
	"github.com/pingchenchan/ad-placement-service/models"
)
type AdsResponse struct {
    Ads []models.Ad `json:"ads"`
}



func GenerateAds() []models.Ad {
    ads := make([]models.Ad, 3000)
    now := time.Now()

    genders := []string{"M", "F", ""}
    countries := []string{"TW", "JP", ""}
    platforms := []string{"android", "ios", "web", ""}
    ages := []int{0, 10, 15, 20, 25, 30, 35, 40, 45, 50}

    rand.Seed(time.Now().UnixNano())
    charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    for i := 0; i < 3000; i++ {
        ad := models.Ad{
            Conditions: []models.Condition{
                {
					Country:  make([]string, 0),
					Platform: make([]string, 0),
				},
            },
        }
		if genders[i%len(genders)] != "" {
			ad.Conditions[0].Gender = new(string)
			*ad.Conditions[0].Gender = genders[i%len(genders)]
		}

		if ages[i%len(ages)] != 0 {
			ad.Conditions[0].AgeStart = new(int)
			*ad.Conditions[0].AgeStart = ages[i%len(ages)]
			ad.Conditions[0].AgeEnd = new(int)
			*ad.Conditions[0].AgeEnd = *ad.Conditions[0].AgeStart + 10
		}
	
		if genders[i%len(genders)] != "" {
			ad.Conditions[0].Gender = new(string)
			*ad.Conditions[0].Gender = genders[i%len(genders)]
		}

		if countries[i%len(countries)] != "" {
			ad.Conditions[0].Country = append(ad.Conditions[0].Country, countries[i%len(countries)])
		}

        if platforms[i%len(platforms)] != "" {
            ad.Conditions[0].Platform = append(ad.Conditions[0].Platform, platforms[i%len(platforms)])
        }


        randStr := make([]byte, 5)
        for i := range randStr {
            randStr[i] = charset[rand.Intn(len(charset))]
        }

        if i < 1000 {
            // StartAt < NOW < EndAt
            ad.StartAt = now.Add(-1 * time.Hour)
            ad.EndAt = now.Add(1 * time.Hour)
            ad.Title = "Past Ad: " + genders[i%len(genders)] + " " + strings.Join(ad.Conditions[0].Country, ", ") + " " + strings.Join(ad.Conditions[0].Platform, ", ") + " " + string(randStr)
        } else if i < 2000 {
            // EndAt < NOW
            ad.StartAt = now.Add(-2 * time.Hour)
            ad.EndAt = now.Add(-1 * time.Hour)
            ad.Title = "Current Ad: " + genders[i%len(genders)] + " " + strings.Join(ad.Conditions[0].Country, ", ") + " " + strings.Join(ad.Conditions[0].Platform, ", ") + " " + string(randStr)
        } else {
            // NOW < StartAt
            ad.StartAt = now.Add(1 * time.Hour)
            ad.EndAt = now.Add(2 * time.Hour)
            ad.Title = "Future Ad: " + genders[i%len(genders)] + " " + strings.Join(ad.Conditions[0].Country, ", ") + " " + strings.Join(ad.Conditions[0].Platform, ", ") + " " + string(randStr)
        }

        ads[i] = ad
    }

    return ads
}

// createSampleGetQuery creates a sample GET query for retrieving ads
func createSampleGetQuery(age int, gender, country, platform string) string {
	// Create a url.Values object and set the query parameters
	params := url.Values{}
	params.Add("offset", "10")
	params.Add("limit", "3")
	params.Add("age", strconv.Itoa(age))
	params.Add("gender", gender)
	params.Add("country", country)
	params.Add("platform", platform)

	// Return the encoded query parameters
	return params.Encode()
}

// getAdFunction sends a GET request to the specified endpoint and returns the response
func getAdFunction(endpoint string, query string) (*http.Response, error) {
    // Create a new GET request
    req, err := http.NewRequest("GET", endpoint+"?"+query, nil)
    if err != nil {
        return nil, err
    }

    // Send the request and return the response
    client := &http.Client{}
    return client.Do(req)
}


func TestInitialLoad(t *testing.T) {
	err := db.DropDatabaseAndCollection("advertising", "ads")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.EnsureCollectionAndIndexes("advertising", "ads")
	if err != nil {
		log.Fatal(err)
	}

	ads := GenerateAds()

	for idx, ad := range ads {
		
		adData, err := json.Marshal(ad)
		if err != nil {
			t.Fatal(err)
		}
		_, err = CreateAdFunction("http://go-backend:8080/ads",adData)
		if err != nil {
			t.Fatal(err)
		}
		log.Print(idx,ad)
	}
}


func TestGet10000AdHttp(t *testing.T) {
	Get10000Ad(t, "http://go-backend:8080/ads")

}
func TestGet10000AdRedixHttp(t *testing.T) {
	Get10000Ad(t, "http://go-backend/adsRedix")
}

func Get10000Ad(t *testing.T, endpoint string) {

//clean the db
	err := db.DropDatabaseAndCollection("advertising", "ads")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.EnsureCollectionAndIndexes("advertising", "ads")
	if err != nil {
		log.Fatal(err)
	}


    start := time.Now()

    // Create a channel to handle errors
    errChannel := make(chan error, 10000)

    // Create a channel to handle success
    successChannel := make(chan struct{}, 10000)

    // Create a channel to collect request times
    timeChannel := make(chan time.Duration, 10000)

    var wg sync.WaitGroup
    wg.Add(1)

    for i := 0; i < 1; i++ {
        go func() {
            defer wg.Done()

            // Record the start time of the request
            getData := createSampleGetQuery(25, "M", "TW", "android")
            
            startTime := time.Now()

            resp, err := getAdFunction(endpoint,getData)

            // Record the end time of the request and send the duration to the time channel
            timeChannel <- time.Since(startTime)

			if err != nil {
				errChannel <- err
			} else if resp.StatusCode != http.StatusOK {
				errChannel <- fmt.Errorf("unexpected status code: got %v want %v", resp.StatusCode, http.StatusOK)
			} else {
				var adsResponse AdsResponse
				if err := json.NewDecoder(resp.Body).Decode(&adsResponse); err != nil {
                    errChannel <- fmt.Errorf("failed to decode response: %v", err)
                    return
                }

				//print the ads
				for _, ad := range adsResponse.Ads {
					if !validateAd(&ad, 25, "M", "TW", "android") {
						errChannel <- fmt.Errorf("ad does not match query parameters: %v", ad)
					} else {
						fmt.Println(ad)
					}
				}



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
    logRequestStats(t, timeChannel, endpoint , start)
}


func validateAd(ad *models.Ad, age int, gender, country, platform string) bool {
    // Check if the ad's conditions match the query parameters
	for _, condition := range ad.Conditions {
		if age != 0 && (*condition.AgeStart > age || *condition.AgeEnd < age) {
			return false
		}
		if gender != "" && *condition.Gender != gender {
			return false
		}
		if country != "" && !contains(condition.Country, country) {
			return false
		}
		if platform != "" && !contains(condition.Platform, platform) {
			return false
		}
	}

	// Check if the ad's start and end times are within the current time
	now := time.Now()
		if ad.StartAt.After(now) || ad.EndAt.Before(now) {
			return false
		}

		return true
}

func contains(slice []string, item string) bool {
    for _, a := range slice {
        if a == item {
            return true
        }
    }
    return false
}