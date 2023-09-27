package main

import (
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/data/types"
)

// used for BACK-37 to set hash for jellyfish datum for migrating to platform API
func CreateHash(bsonData bson.M) string {
	// represents data from base id fields
	var baseData = func(bsonData bson.M) []string {
		return []string{
			bsonData["_userId"].(string),
			bsonData["deviceId"].(string),
			bsonData["time"].(time.Time).Format(types.TimeFormat),
			bsonData["type"].(string),
		}
	}
	var theHash string
	switch bsonData["type"] {
	case "basal":
		theHash = makeHash(
			append(baseData(bsonData), bsonData["deliveryType"].(string))...,
		)
	case "bolus", "deviceEvent":
		theHash = makeHash(
			append(baseData(bsonData), bsonData["subType"].(string))...,
		)
	case "smbg", "bloodKetone", "cbg":
		theHash = makeHash(
			append(
				baseData(bsonData),
				bsonData["units"].(string),
				strconv.FormatFloat(bsonData["value"].(float64), 'f', -1, 64),
			)...,
		)
	default:
		theHash = makeHash(baseData(bsonData)...)
	}
	return theHash
}

func makeHash(fields ...string) string {
	hash, err := deduplicator.GenerateIdentityHash(fields)
	if err != nil {
		panic(err)
	}
	return hash
}
