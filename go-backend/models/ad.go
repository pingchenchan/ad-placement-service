package models

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Ad represents the structure of our resource
type Ad struct {
    ID         string      `json:"id" bson:"_id,omitempty"`
    Title      string      `json:"title" bson:"title"`
    StartAt    time.Time   `json:"startAt" bson:"startAt"`
    EndAt      time.Time   `json:"endAt" bson:"endAt"`
    Conditions []Condition `json:"conditions" bson:"conditions"`
}

// Condition represents the targeting criteria for an ad
type Condition struct {
    AgeStart *int      `json:"ageStart,omitempty" bson:"ageStart,omitempty" binding:"omitempty,min=1,max=100"`
    AgeEnd   *int      `json:"ageEnd,omitempty" bson:"ageEnd,omitempty" binding:"omitempty,min=1,max=100"`
    Gender   *string   `json:"gender,omitempty" bson:"gender,omitempty" binding:"omitempty,oneof=M F"`
    Country  []string  `json:"country,omitempty" bson:"country,omitempty" binding:"omitempty,dive,iso3166_1_alpha2"`
    Platform []string  `json:"platform,omitempty" bson:"platform,omitempty" binding:"omitempty,dive,oneof=android ios web"`
}

type AdQueryParams struct {
    Offset   int    `form:"offset"`
    Limit    int    `form:"limit"`
    Age      int    `form:"age"`
    Gender   string `form:"gender"`
    Country  string `form:"country"`
    Platform string `form:"platform"`
}

// CountryCode represents the structure of a country code from the JSON file
type CountryCode struct {
    Alpha2      string `json:"alpha-2"`
}

// loadCountryCodes loads the country codes from the JSON file
func LoadCountryCodes(path string) (map[string]bool, error) {
    v := make(map[string]bool)
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var codes []CountryCode
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&codes)
    if err != nil {
        return nil, err
    }

    for _, cc := range codes {
        v[cc.Alpha2] = true
    }

    //log success message
    log.Printf("Successfullly Loaded %d country codes", len(v)) 

    return v, nil
}

func ValidateCountry(ad Ad, countryCode map[string]bool) bool {
    for _, condition := range ad.Conditions {
        if condition.Country != nil {
            countries := condition.Country
            for _, country := range countries {
                _, ok := countryCode[country]
                if !ok {
                    return false
                }
            }
        }
    }
    return true
}

func ValidateAge(ad Ad) bool { //Age, min=1,max=100
    for _, condition := range ad.Conditions {
        if condition.AgeStart != nil && condition.AgeEnd != nil {
            if *condition.AgeStart > *condition.AgeEnd {
                return false
            }
            if *condition.AgeStart < 1 || *condition.AgeEnd > 100 {
                return false
            }
        }
    }
    return true
}

func ValidateGender(ad Ad) bool { //Gender must be M or F
    for _, condition := range ad.Conditions {
        if condition.Gender != nil {
            gender := *condition.Gender
            if gender != "M" && gender != "F" {
                return false
            }
        }
    }
    return true
}


func ValidatePlatform(ad Ad) bool { //Platform must be android, ios, or web
    
    validPlatforms := map[string]bool{
        "android": true,
        "ios":     true,
        "web":     true,
    }
    for _, condition := range ad.Conditions {
        for _, platform := range condition.Platform {
            if _, ok := validPlatforms[platform]; !ok {
                return false
            }
        }
    }
    return true
}

func ValidateDates(ad Ad) bool {
    if ad.StartAt.IsZero() || ad.EndAt.IsZero() {
        log.Printf("Invalid date range 1, startAt: %v, endAt: %v", ad.StartAt, ad.EndAt)
        return false
    }
    if !ad.EndAt.After(ad.StartAt) {
        log.Printf("Invalid date range 2, startAt: %v, endAt: %v", ad.StartAt, ad.EndAt)
        return false
    }
    return true
}

func ValidateAd(ad Ad, countryCode map[string]bool) (bool, string) {
    if !ValidateDates(ad) {
        return false, "Invalid date range"
    }
    if !ValidateCountry(ad, countryCode) {
        return false, "Invalid country code"
    }
    if !ValidateAge(ad) {
        return false, "Invalid age range"
    }
    if !ValidateGender(ad) {
        return false, "Invalid gender"
    }
    if !ValidatePlatform(ad) {
        return false, "Invalid platform"
    }
    return true, ""
}