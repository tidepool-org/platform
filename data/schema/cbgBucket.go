package schema

import "time"

type (
	CbgBucket struct {
		Id                string      `bson:"_id,omitempty"`
		CreationTimestamp time.Time   `bson:"creationTimestamp,omitempty"`
		UserId            string      `bson:"userId,omitempty"`
		Day               time.Time   `bson:"day,omitempty"` // ie: 2021-09-28
		Samples           []CbgSample `bson:"samples"`
	}

	CbgSample struct {
		Sample `bson:",inline"`
		Value  float64 `bson:"value,omitempty"`
		Units  string  `bson:"units,omitempty"`
	}
)

func (c CbgBucket) GetId() string {
	return c.Id
}

func (c CbgSample) GetTimestamp() time.Time {
	return c.Timestamp
}
