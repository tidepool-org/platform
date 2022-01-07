package schema

import (
	"time"
)

type (
	BasalBucket struct {
		Id                string        `bson:"_id,omitempty"`
		CreationTimestamp time.Time     `bson:"creationTimestamp,omitempty"`
		UserId            string        `bson:"userId,omitempty" `
		Day               time.Time     `bson:"day,omitempty"` // ie: 2021-09-28
		Samples           []BasalSample `bson:"samples"`
	}

	BasalSample struct {
		Sample       `bson:",inline"`
		Guid         string  `bson:"guid,omitempty"`
		DeliveryType string  `bson:"deliveryType,omitempty"`
		Duration     int     `bson:"duration,omitempty"`
		Rate         float64 `bson:"rate"`
	}
)

func (b BasalBucket) GetId() string {
	return b.Id
}

func (b BasalSample) GetTimestamp() time.Time {
	return b.Timestamp
}
