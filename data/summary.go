package data

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Summary struct {
    ID                 primitive.ObjectID  `json:"-" bson:"_id,omitempty"`
	UserID             string              `json:"UserID" bson:"_userId"`

	LastUpdated        time.Time   `json:"lastUpdated" bson:"lastUpdated"`
	LastUpload         time.Time   `json:"lastUpload" bson:"lastUpload"`
	AverageGlucose     float64     `json:"avgGlucose" bson:"avgGlucose"`
	TimeInRange        float64     `json:"timeInRange" bson:"timeInRange"`
}
