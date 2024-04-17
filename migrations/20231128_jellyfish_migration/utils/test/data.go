package test

import (
	"crypto/sha1"
	"encoding/base32"
	"io"
	"strings"
	"time"

	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/test"
)

func base(deviceID string) map[string]interface{} {
	return map[string]interface{}{
		"_id":         "17dbokav5t6pssjv72gm0nie3u25b54m",
		"deviceId":    deviceID,
		"deviceTime":  "2017-11-05T12:56:51",
		"id":          "3f0075ad57ad603c83dc1e1a76aefcaf",
		"localTime":   "2017-11-05T12:56:51.000Z",
		"_userId":     "8da6e693b8",
		"_groupId":    "87df73fd41",
		"createdTime": "2022-06-21T22:40:07.732+00:00",
		"_version":    0,
		"_active":     true,
		"uploadId":    "a21c82a5f5d2860add2539acded6b614",
		"time":        "2022-06-21T22:40:07.732+00:00",
	}
}

func baseWithTime(deviceID string, groupID string, userID string, t time.Time) map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"_id":         "17dbokav5t6pssjv72gm0nie3u25b54m",
		"deviceId":    deviceID,
		"deviceTime":  t.Format("2006-01-02T15:04:05"),
		"id":          "3f0075ad57ad603c83dc1e1a76aefcaf",
		"localTime":   t.Format("2006-01-02T15:04:05.999Z"),
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

func dexG5MobDatumStringAnnotations(datum map[string]interface{}) map[string]interface{} {
	datum["annotations"] = `[{"code":"bg/out-of-range","threshold":40,"value":"low"}]`
	datum["type"] = "cbg"
	datum["units"] = "mmol/L"
	datum["value"] = 8.1596
	return datum
}

func tandemPumpSettingsDatum(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "pumpSettings"
	datum["activeSchedule"] = "Simple"
	datum["units"] = map[string]interface{}{"carb": "grams", "bg": "mg/dL"}
	datum["basalSchedules"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
		"Standard": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
	}
	datum["carbRatios"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"amount": 10, "start": 0},
			{"amount": 10, "start": 46800000},
		},
		"Standard": []map[string]interface{}{
			{"amount": 10, "start": 0},
			{"amount": 10, "start": 46800000},
		},
	}
	datum["insulinSensitivities"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"amount": 2.7753739955227665, "start": 0},
			{"amount": 2.7753739955227665, "start": 46800000},
		},
		"Standard": []map[string]interface{}{
			{"amount": 2.7753739955227665, "start": 0},
			{"amount": 2.7753739955227665, "start": 46800000},
		},
	}

	datum["bgTargets"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"target": 5.550747991045533, "start": 0},
			{"target": 5.550747991045533, "start": 46800000},
		},
		"Standard": []map[string]interface{}{
			{"target": 5.550747991045533, "start": 0},
			{"target": 5.550747991045533, "start": 46800000},
		},
	}

	datum["payload"] = map[string]interface{}{
		"logIndices": []interface{}{0},
	}

	return datum
}

func tandemPumpSettingsWithSleepScheduleDatum(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "pumpSettings"
	datum["activeSchedule"] = "Simple"
	datum["units"] = map[string]interface{}{"carb": "grams", "bg": "mg/dL"}
	datum["basalSchedules"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
		"Standard": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
	}
	datum["carbRatios"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"amount": 10, "start": 0},
			{"amount": 10, "start": 46800000},
		},
		"Standard": []map[string]interface{}{
			{"amount": 10, "start": 0},
			{"amount": 10, "start": 46800000},
		},
	}
	datum["insulinSensitivities"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"amount": 2.7753739955227665, "start": 0},
			{"amount": 2.7753739955227665, "start": 46800000},
		},
		"Standard": []map[string]interface{}{
			{"amount": 2.7753739955227665, "start": 0},
			{"amount": 2.7753739955227665, "start": 46800000},
		},
	}

	datum["bgTargets"] = map[string]interface{}{
		"Simple": []map[string]interface{}{
			{"target": 5.550747991045533, "start": 0},
			{"target": 5.550747991045533, "start": 46800000},
		},
		"Standard": []map[string]interface{}{
			{"target": 5.550747991045533, "start": 0},
			{"target": 5.550747991045533, "start": 46800000},
		},
	}

	datum["payload"] = map[string]interface{}{
		"logIndices": []interface{}{0},
	}

	//## TODO test for [pumpSettings] sleepSchedules []interface {}{map[string]interface {}{"days":[]interface {}{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}, "enabled":true, "end":25200, "start":82800}, map[string]interface {}{"days":[]interface {}{"Sunday"}, "enabled":false, "end":32400, "start":3600}}

	datum["sleepSchedules"] = []interface{}{
		map[string]interface{}{"days": []interface{}{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}, "enabled": true, "end": 25200, "start": 82800},
		map[string]interface{}{"days": []interface{}{"Sunday"}, "enabled": false, "end": 32400, "start": 3600},
	}

	return datum
}

func automatedBolusRange(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "bolus"
	datum["subType"] = "automated"
	datum["normal"] = 2.51753
	datum["expectedNormal"] = 2.51752
	return datum
}

func carelinkPumpSettings(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "pumpSettings"
	datum["activeSchedule"] = "standard"
	datum["units"] = map[string]interface{}{"carb": "grams", "bg": "mg/dL"}
	datum["basalSchedules"] = map[string]interface{}{
		"standard": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
		"pattern a": []map[string]interface{}{
			{"rate": 0.5, "start": 0},
			{"rate": 1.35, "start": 55800000},
		},
		"pattern b": []map[string]interface{}{},
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
		{"target": 5.550747991045533, "start": 0},
		{"target": 5.550747991045533, "start": 46800000},
	}

	datum["payload"] = map[string]interface{}{
		"logIndices": []interface{}{5309},
	}
	return datum
}

func omnipodPumpSettingsDatum(datum map[string]interface{}) map[string]interface{} {
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

func tandemAutomatedBasalDatum(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "basal"
	datum["deliveryType"] = "automated"
	datum["timezoneOffset"] = -300
	datum["clockDriftOffset"] = -137000
	datum["conversionOffset"] = 0
	datum["duration"] = 300000
	datum["rate"] = 0.335
	datum["percent"] = 0.47857142857142865
	datum["conversionOffset"] = 0
	datum["suppressed"] = map[string]interface{}{
		"type":         "basal",
		"deliveryType": "scheduled",
		"rate":         0.7,
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

func reservoirChangeDeviceEventDatum(datum map[string]interface{}) map[string]interface{} {
	datum["type"] = "deviceEvent"
	datum["subType"] = "reservoirChange"
	datum["status"] = "cvv61jde62b6i28bgot57f18bor5au1n"
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

func emptyPayload(datum map[string]interface{}) map[string]interface{} {
	datum["payload"] = map[string]interface{}{}
	datum["type"] = "cbg"
	datum["units"] = "mmol/L"
	datum["value"] = 8.1596
	return datum
}

func pumpSettingsWithBolus(datum map[string]interface{}) map[string]interface{} {
	datum = tandemPumpSettingsDatum(datum)
	datum["bolus"] = &pump.BolusMap{
		"bolus-1": pumpTest.NewRandomBolus(),
		"bolus-2": pumpTest.NewRandomBolus(),
	}
	return datum
}

var CBGDexcomG5StringPayloadDatum = dexG5MobDatumStringPayload(base("DexG5Mob_iPhone"))
var CBGDexcomG5StringAnnotationsDatum = dexG5MobDatumStringAnnotations(base("DexG5Mob_iPhone"))
var PumpSettingsTandem = tandemPumpSettingsDatum(base("tandem99999999"))
var PumpSettingsWithSleepScheduleTandem = tandemPumpSettingsWithSleepScheduleDatum(base("tandem99999999"))
var PumpSettingsCarelink = carelinkPumpSettings(base("MiniMed 530G - 751-=-11111111"))
var PumpSettingsOmnipod = omnipodPumpSettingsDatum(base("InsOmn-837268"))
var PumpSettingsOmnipodBGTargetCorrect = omnipodPumpSettingsDatumTargetSet(base("InsOmn-837268"))
var AutomatedBasalTandem = tandemAutomatedBasalDatum(base("tandemCIQ1111111111111"))
var AutomatedBolus = automatedBolusRange(base("tandemCIQ1111111111111"))
var WizardTandem = tandemWizardDatum(base("tandemCIQ1111111111111"))
var ReservoirChangeWithStatus = reservoirChangeDeviceEventDatum(base("InsOmn-1111111111111"))
var AlarmDeviceEventDatum = alarmDeviceEventDatum(base("tandemCIQ100000000000"))
var CGMSetting = cgmSettingsDatum(base("DexG5MobRec-1111111111111"))
var EmptyPayloadDatum = emptyPayload(base("Dex-device"))
var PumpSettingsWithBolusDatum = pumpSettingsWithBolus(base("tandem99999999"))

func BulkJellyfishData(deviceID string, groupID string, userID string, requiredRecords int) []map[string]interface{} {
	data := []map[string]interface{}{}
	twoWeeksAgo := time.Now().AddDate(0, 0, 14)

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

	for count := 0; count < requiredRecords; count++ {
		typ := test.RandomChoice([]string{"cbg", "wizard", "deviceEvent"})
		dTime := twoWeeksAgo.Add(time.Duration(count) * time.Minute)
		base := baseWithTime(deviceID, groupID, userID, dTime)
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
