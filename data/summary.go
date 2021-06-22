package data

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Summary struct {
    ID                 primitive.ObjectID  `json:"_id" bson:"_id"`
	UserID             string              `json:"_userId" bson:"_userId"`

	LastUpdated        time.Time   `json:"lastUpdated" bson:"lastUpdated"`
	LastUpload         time.Time   `json:"lastUpload" bson:"lastUpload"`
	AverageGlucose     float64     `json:"avgGlucose" bson:"avgGlucose"`
	TimeInRange        uint64      `json:"timeInRange" bson:"timeInRange"`
}
