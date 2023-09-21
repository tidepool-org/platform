package main

import (
	"strconv"

	"github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"go.mongodb.org/mongo-driver/bson"
)

// used for BACK-37 to set hash for jellyfish datum for migrating to platform API
func CreateHash(bsonData bson.M) string {
	// represents data from base id fields
	var baseData = func(bsonData bson.M) []string {
		return []string{
			bsonData["_userId"].(string),
			bsonData["deviceId"].(string),
			bsonData["time"].(string),
			bsonData["type"].(string),
		}
	}
	var theHash string
	switch bsonData["type"] {
	case "basal":
		theHash = makeJellyfishHash(
			append(baseData(bsonData), bsonData["deliveryType"].(string))...,
		)
	case "bolus", "deviceEvent":
		theHash = makeJellyfishHash(
			append(baseData(bsonData), bsonData["subType"].(string))...,
		)
	case "smbg", "bloodKetone", "cbg":
		theHash = makeJellyfishHash(
			append(
				baseData(bsonData),
				bsonData["units"].(string),
				strconv.FormatFloat(bsonData["value"].(float64), 'f', -1, 64),
			)...,
		)
	default:
		theHash = makeJellyfishHash(baseData(bsonData)...)
	}
	return theHash
}

func JellyfishObjectIDHash(bsonData bson.M) string {
	return makeJellyfishHash(
		CreateHash(bsonData),
	)
}

func makeJellyfishHash(fields ...string) string {
	hash, err := deduplicator.GenerateIdentityHash(fields)
	if err != nil {
		panic(err)
	}
	return hash
}
