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
	"sync/atomic"
	"testing"
	"time"

	"github.com/pingchenchan/ad-placement-service/db"
	"github.com/pingchenchan/ad-placement-service/models"
	"github.com/stretchr/testify/assert"
)

type AdsResponse []models.Ad

func GenerateAds() []models.Ad {
	ads := make([]models.Ad, 3000)
	now := time.Now()

	genders := []string{"M", "F", ""}
	countries := []string{"TW", "JP", ""}
	platforms := []string{"android", "ios", "web", ""}
	ages := []int{-1, 0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90}

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

		if ages[i%len(ages)] != -1 {
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
			ad.StartAt = now.Add(-1 * time.Hour * 24)
			ad.EndAt = now.Add(1 * time.Hour * 24)
			ad.Title = "PastAd_Age-" + strconv.Itoa(ages[i%len(ages)]) + "~" + strconv.Itoa(ages[i%len(ages)]+10) + "_" + genders[i%len(genders)] + "_" + strings.Join(ad.Conditions[0].Country, ", ") + "_" + strings.Join(ad.Conditions[0].Platform, ", ") + "_" + string(randStr)
		} else if i < 2000 {
			// EndAt < NOW
			ad.StartAt = now.Add(-2 * time.Hour * 24)
			ad.EndAt = now.Add(-1 * time.Hour * 24)
			ad.Title = "CurrentAd_ Age-" + strconv.Itoa(ages[i%len(ages)]) + "~" + strconv.Itoa(ages[i%len(ages)]+10) + "_" + genders[i%len(genders)] + "_" + strings.Join(ad.Conditions[0].Country, ", ") + "_" + strings.Join(ad.Conditions[0].Platform, ", ") + "_" + string(randStr)
		} else {
			// NOW < StartAt
			ad.StartAt = now.Add(1 * time.Hour * 24)
			ad.EndAt = now.Add(2 * time.Hour * 24)
			ad.Title = "FutureAd_Age-" + strconv.Itoa(ages[i%len(ages)]) + "~" + strconv.Itoa(ages[i%len(ages)]+10) + "_" + genders[i%len(genders)] + "_" + strings.Join(ad.Conditions[0].Country, ", ") + "_" + strings.Join(ad.Conditions[0].Platform, ", ") + "_" + string(randStr)
		}

		ads[i] = ad
	}

	return ads
}

// createSampleGetQuery creates a sample GET query for retrieving ads
func createSampleGetQuery(p models.AdQueryParams) string {
	// Create a url.Values object and set the query parameters
	params := url.Values{}
	params.Add("offset", strconv.Itoa(p.Offset))
	params.Add("limit", strconv.Itoa(p.Limit))
	params.Add("age", strconv.Itoa(p.Age))
	params.Add("gender", p.Gender)
	params.Add("country", p.Country)
	params.Add("platform", p.Platform)

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
		t.Fatal(err)
	}
	_, err = db.EnsureCollectionAndIndexes("advertising", "ads")
	if err != nil {
		t.Fatal(err)
	}

	ads := GenerateAds()

	for _, ad := range ads {

		adData, err := json.Marshal(ad)
		if err != nil {
			t.Fatal(err)
		}
		_, err = CreateAdFunction("http://go-backend:8080/ads", adData)
		if err != nil {
			t.Fatal(err)
		}

	}
	log.Print("successfullly loaded 3000 ads")
}

func TestGet10000AdHttp(t *testing.T) {
	// TestInitialLoad(t )
	Get10000Ad(t, "http://go-backend:8080/ads", 1)

}
func TestGet10000AdRedixHttp(t *testing.T) {
	// TestInitialLoad(t )
	Get10000Ad(t, "http://go-backend:8080/adsRedix", 1000)
}

func genRandomQuery(numQueries int) []models.AdQueryParams {
    // Create slices for the enum parameters
    genders := []string{"M", "F", ""}
    countries := []string{"TW", "JP",""}
    platforms := []string{"android", "ios", "web",""}

    // Create a slice to hold the different queries
    queries := make([]models.AdQueryParams, 0, numQueries)

    // Generate the different queries
    for _, gender := range genders {
        for _, country := range countries {
            for _, platform := range platforms {
                for offset := 0; offset < 5; offset++ {
                    for limit := 1; limit <= 100; limit++ {
                        for age := 1; age <= 100; age++ {
                            if len(queries) < numQueries {
                                queries = append(queries, models.AdQueryParams{
                                    Offset:   offset,
                                    Limit:    limit,
                                    Age:      age,
                                    Gender:   gender,
                                    Country:  country,
                                    Platform: platform,
                                })
                            } else {
                                // If we have generated enough queries, return them
                                return queries
                            }
                        }
                    }
                }
            }
        }
    }

    return queries
}
func checkQueriesUnique(queries []models.AdQueryParams) bool {
    // Create a map to track the unique queries
    uniqueQueries := make(map[models.AdQueryParams]bool)
    isUnique := true

    // Check each query in the queries slice
    for _, query := range queries {
        // If the query is in the uniqueQueries map, set isUnique to false
        if _, exists := uniqueQueries[query]; exists {
            isUnique = false
        }

        // If the query is not in the uniqueQueries map, add it
        uniqueQueries[query] = true
    }

    // Print numbers of unique queries
    fmt.Println("Numbers of unique queries: ", len(uniqueQueries))

    // Return whether all queries are unique
    return isUnique
}
func Get10000Ad(t *testing.T, endpoint string, numQueries int) {
	//Get 10000 request with same query

	

	var TEST_NUMBER = 10000

	// Create a channel to handle errors
	errChannel := make(chan error, TEST_NUMBER)

	// Create a channel to handle success
	successChannel := make(chan struct{}, TEST_NUMBER)

	// Create a channel to collect request times
	timeChannel := make(chan time.Duration, TEST_NUMBER)

	// Create a slice to hold the different queries
	queries := genRandomQuery(numQueries)
	queriesChannel := make(chan  models.AdQueryParams, TEST_NUMBER)
	if checkQueriesUnique(queries) {
		fmt.Println("All queries are unique")
	} else {
		fmt.Println("There are duplicate queries")
	}
	var passValidation int32
	var failValidation int32

	var wg sync.WaitGroup
	wg.Add(TEST_NUMBER)
	start := time.Now()
	for i := 0; i < TEST_NUMBER; i++ {
		go func(i int) {
			defer wg.Done()

			// Record the start time of the request
			p := queries[i % numQueries]//random query
			//print i % numQueries and i and numQueries
			// fmt.Println(i % numQueries, i, numQueries)

			queriesChannel <- p
			// p := models.AdQueryParams{
			// 	Offset:   0,
			// 	Limit:    3,  // Limit between 1 and 100
			// 	Age:      15, // Age between 1 and 100
			// 	Gender:   "M",
			// 	Country:  "TW",
			// 	Platform: "android",
			// }

			// log.Printf("Query: %v", p)
			getData := createSampleGetQuery(p)

			startTime := time.Now()

			resp, err := getAdFunction(endpoint, getData)

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

				//Validate the response
				for _, ad := range adsResponse {

					if !validateAd(&ad, p) {
						atomic.AddInt32(&failValidation, 1)
						errChannel <- fmt.Errorf("ad does not match query parameters: %v", ad)
					} else {
						atomic.AddInt32(&passValidation, 1)
						// fmt.Println(ad)

					}
				}

				successChannel <- struct{}{}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errChannel)
	close(successChannel)
	close(timeChannel)
	close(queriesChannel)


	//trans queriesChannel to slice
	queriesChannelSlice := make([]models.AdQueryParams, 0)
	for query := range queriesChannel {
		queriesChannelSlice = append(queriesChannelSlice, query)
	}
	fmt.Println("Number of queries: ", len(queriesChannelSlice))
	if checkQueriesUnique(queriesChannelSlice) {
		fmt.Println("All queries are unique")
	} else {
		fmt.Println("There are duplicate queries")
	}


	t.Logf("")
	// Check how many errors
	t.Logf("Numbers of error requests: %v", len(errChannel))
	t.Logf("Numbers of fail validation: %v", failValidation)
	assert.Equal(t, len(successChannel), TEST_NUMBER)

	// Check how many successes
	t.Logf("Numbers of successful requests: %v", len(successChannel))
	t.Logf("Numbers of pass validation: %v", passValidation)
	// Check if there were any errors
	for err := range errChannel {
		if err != nil {
			t.Logf("error, %v", err)
		}
	}

	// Calculate the average, max, and min request time
	logRequestStats(t, timeChannel, endpoint, start)
}

func validateAd(ad *models.Ad, p models.AdQueryParams) bool {
	// Check if the ad's conditions match the query parameters
	for _, condition := range ad.Conditions {
		if p.Age != 0 && (*condition.AgeStart > p.Age || *condition.AgeEnd < p.Age) {
			return false
		}
		if p.Gender != "" && *condition.Gender != p.Gender {
			return false
		}
		if p.Country != "" && !contains(condition.Country, p.Country) {
			return false
		}
		if p.Platform != "" && !contains(condition.Platform, p.Platform) {
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
