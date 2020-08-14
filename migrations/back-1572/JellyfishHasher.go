package main

import (
	"crypto/sha1"
	"encoding/base32"
	"io"
	"strconv"
	"strings"

	"github.com/globalsign/mgo/bson"
)

func JellyfishIDHash(bsonData bson.M) string {
	var theHash string
	switch bsonData["type"] {
	case "basal":
		theHash = makeJellyfishHash(
			bsonData["type"].(string),
			bsonData["deliveryType"].(string),
			bsonData["deviceId"].(string),
			bsonData["time"].(string),
		)
	case "bolus":
		theHash = makeJellyfishHash(
			bsonData["type"].(string),
			bsonData["subType"].(string),
			bsonData["deviceId"].(string),
			bsonData["time"].(string),
		)
	case "deviceEvent":
		theHash = makeJellyfishHash(
			bsonData["type"].(string),
			bsonData["subType"].(string),
			bsonData["time"].(string),
			bsonData["deviceId"].(string),
		)
	case "smbg":
		theHash = makeJellyfishHash(
			bsonData["type"].(string),
			bsonData["deviceId"].(string),
			bsonData["time"].(string),
			strconv.FormatFloat(bsonData["value"].(float64), 'f', -1, 64),
		)
	default:
		theHash = makeJellyfishHash(
			bsonData["type"].(string),
			bsonData["deviceId"].(string),
			bsonData["time"].(string),
		)
	}
	return theHash
}

func JellyfishObjectIDHash(bsonData bson.M) string {
	return makeJellyfishHash(
		bsonData["id"].(string),
		bsonData["_groupId"].(string),
	)
}

func makeJellyfishHash(fields ...string) string {
	h := sha1.New()
	hashFields := append(fields, "bootstrap")
	for _, field := range hashFields {
		io.WriteString(h, field)
		io.WriteString(h, "_")
	}
	sha1 := h.Sum(nil)
	return strings.ToLower(base32.HexEncoding.WithPadding('-').EncodeToString(sha1))
}
