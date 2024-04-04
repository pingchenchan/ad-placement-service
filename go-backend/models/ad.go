package models

import "time"

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
    AgeStart *int      `json:"ageStart,omitempty" bson:"ageStart,omitempty"`
    AgeEnd   *int      `json:"ageEnd,omitempty" bson:"ageEnd,omitempty"`
    Gender   *string   `json:"gender,omitempty" bson:"gender,omitempty"`
    Country  []string  `json:"country,omitempty" bson:"country,omitempty"`
    Platform []string  `json:"platform,omitempty" bson:"platform,omitempty"`
}