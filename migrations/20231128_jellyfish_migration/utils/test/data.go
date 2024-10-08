package test

import (
	"crypto/sha1"
	"encoding/base32"
	"io"
	"strings"
	"time"

	"github.com/tidepool-org/platform/test"
)

func datumBase(deviceID string, groupID string, userID string, t time.Time) map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"_id":         "17dbokav5t6pssjv72gm0nie3u25b54m",
		"deviceId":    deviceID,
		"deviceTime":  t.Format("2006-01-02T15:04:05"),
		"id":          "3f0075ad57ad603c83dc1e1a76aefcaf",
		"_userId":     userID,
		"_groupId":    groupID,
		"createdTime": now.Format("2006-01-02T15:04:05.999+07:00"),
		"_version":    0,
		"_active":     true,
		"uploadId":    "a21c82a5f5d2860add2539acded6b614",
		"time":        t.Format("2006-01-02T15:04:05.999+07:00"),
	}
}

// payload as a string rather than object or array
func dexG5MobDatumStringPayload(datum map[string]interface{}) map[string]interface{} {
	datum["payload"] = `{"systemTime":"2017-11-05T18:56:51Z","transmitterId":"410X6M","transmitterTicks":5796922,"trend":"flat","trendRate":0.6,"trendRateUnits":"mg/dL/min"}`
	datum["type"] = "cbg"
	datum["units"] = "mmol/L"
	datum["value"] = 8.1596
	return datum
}

func omnipodPumpSettingsDatumTargetSet(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "pumpSettings"
	datum["activeSchedule"] = "Mine-2016"
	datum["units"] = map[string]interface{}{"carb": "grams", "bg": "mg/dL"}
	datum["basalSchedules"] = map[string]interface{}{
		"Mine-2016": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
		"camp 2015": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
		"weekend b": []map[string]interface{}{},
	}
	datum["carbRatio"] = []map[string]interface{}{
		{"amount": 10, "start": 0},
		{"amount": 10, "start": 32400000},
	}
	datum["insulinSensitivity"] = []map[string]interface{}{
		{"amount": 2.7753739955227665, "start": 0},
		{"amount": 2.7753739955227665, "start": 46800000},
	}

	datum["bgTarget"] = []map[string]interface{}{
		{"target": 5.550747991045533, "start": 0, "high": 7.2159723883591935},
		{"target": 5.550747991045533, "start": 46800000, "high": 7.2159723883591935},
	}
	return datum
}

func tandemWizardDatum(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "wizard"

	datum["timezoneOffset"] = -300
	datum["clockDriftOffset"] = -221000
	datum["conversionOffset"] = 0
	datum["recommended"] = map[string]interface{}{
		"carb":         2,
		"deliveryType": "scheduled",
		"rate":         0.7,
	}

	datum["bgInput"] = 4.440598392836427

	datum["bgTarget"] = map[string]interface{}{
		"target": 4.440598392836427,
	}

	datum["units"] = "mmol/L"
	datum["duration"] = float64(300000)
	datum["rate"] = 0.335
	datum["percent"] = 0.47857142857142865
	datum["conversionOffset"] = 0
	datum["bolus"] = "g2h6nohp5sdndpvl2l8kdete00lle4gt"

	return datum
}

func alarmDeviceEventDatum(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "deviceEvent"
	datum["subType"] = "status"
	datum["status"] = "suspended"
	datum["reason"] = map[string]interface{}{
		"suspended": "automatic",
		"resumed":   "automatic",
	}
	return datum
}

func cgmSettingsDatum(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "cgmSettings"
	datum["units"] = "mmol/L"

	datum["lowAlerts"] = map[string]interface{}{
		"enabled": true,
		"level":   3.8855235937318735,
		"snooze":  900000,
	}

	datum["highAlerts"] = map[string]interface{}{
		"enabled": true,
		"level":   22.202991964182132,
		"snooze":  0,
	}

	datum["rateOfChangeAlerts"] = map[string]interface{}{
		"fallRate": map[string]interface{}{
			"enabled": false,
			"rate":    -0.16652243973136602,
		},
		"riseRate": map[string]interface{}{
			"enabled": false,
			"rate":    0.16652243973136602,
		},
	}

	datum["outOfRangeAlerts"] = map[string]interface{}{
		"enabled": true,
		"snooze":  1200000,
	}
	return datum
}

var makeJellyfishID = func(fields []string) string {
	h := sha1.New()
	hashFields := append(fields, "bootstrap")
	for _, field := range hashFields {
		io.WriteString(h, field)
		io.WriteString(h, "_")
	}
	sha1 := h.Sum(nil)
	id := strings.ToLower(base32.HexEncoding.WithPadding('-').EncodeToString(sha1))
	return id
}

func BulkJellyfishData(deviceID string, groupID string, userID string, requiredRecords int) []map[string]interface{} {
	data := []map[string]interface{}{}
	twoWeeksAgo := time.Now().AddDate(0, 0, 14)

	for count := 0; count < requiredRecords; count++ {
		typ := test.RandomChoice([]string{"cbg", "wizard", "deviceEvent"})
		dTime := twoWeeksAgo.Add(time.Duration(count) * time.Minute)
		base := datumBase(deviceID, groupID, userID, dTime)
		var datum map[string]interface{}

		switch typ {
		case "cbg":
			datum = dexG5MobDatumStringPayload(base)
		case "cgmSettings":
			datum = cgmSettingsDatum(base)
		case "pumpSettings":
			datum = omnipodPumpSettingsDatumTargetSet(base)
		case "wizard":
			datum = tandemWizardDatum(base)
		case "deviceEvent":
			datum = alarmDeviceEventDatum(base)
		}
		datum["_id"] = makeJellyfishID([]string{userID, deviceID, dTime.Format(time.RFC3339), typ})
		datum["id"] = datum["_id"]
		data = append(data, datum)
	}
	return data
}

func uploadDatum(datum map[string]interface{}, t time.Time) map[string]interface{} {
	datum["type"] = "upload"
	datum["computerTime"] = t.Format("2006-01-02T15:04:05")
	datum["deviceTags"] = []string{
		"cgm",
		"insulin-pump",
	}
	datum["deviceManufacturers"] = []string{
		"Medtronic",
	}
	datum["deviceModel"] = "MiniMed 530G 551"
	datum["timeProcessing"] = "utc-bootstrapping"
	return datum
}

func BulkJellyfishUploadData(deviceID string, groupID string, userID string, requiredRecords int) []map[string]interface{} {
	data := []map[string]interface{}{}
	twoMonthsAgo := time.Now().AddDate(0, 2, 00)
	for count := 0; count < requiredRecords; count++ {
		typ := test.RandomChoice([]string{"cbg", "wizard", "deviceEvent"})
		dTime := twoMonthsAgo.Add(time.Duration(count) * time.Hour)
		datum := uploadDatum(datumBase(deviceID, groupID, userID, dTime), dTime)
		datum["_id"] = makeJellyfishID([]string{userID, deviceID, dTime.Format(time.RFC3339), typ})
		datum["id"] = datum["_id"]
		data = append(data, datum)
	}
	return data
}
