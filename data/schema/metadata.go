package schema

import "time"

type (
	Metadata struct {
		Id                  string    `bson:"_id,omitempty"`
		CreationTimestamp   time.Time `bson:"creationTimestamp,omitempty"`
		UserId              string    `bson:"userId,omitempty"`
		OldestDataTimestamp time.Time `bson:"oldestDataTimestamp,omitempty"`
		NewestDataTimestamp time.Time `bson:"newestDataTimestamp,omitempty"`
	}
)
