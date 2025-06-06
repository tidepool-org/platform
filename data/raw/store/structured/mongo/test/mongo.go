package test

import (
	"strings"
	"time"

	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"

	dataRawStoreStructuredMongo "github.com/tidepool-org/platform/data/raw/store/structured/mongo"
	"github.com/tidepool-org/platform/test"
)

func RandomDataRawID() string {
	return RandomDataRawIDFromTime(test.RandomTime())
}

func RandomDataRawIDFromTime(tm time.Time) string {
	return strings.Join([]string{bsonPrimitive.NewObjectID().Hex(), tm.Format(dataRawStoreStructuredMongo.IDDateFormat)}, dataRawStoreStructuredMongo.IDSeparator)
}
