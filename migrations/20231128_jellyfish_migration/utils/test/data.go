package test

import (
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
)

func base(deviceID string) map[string]interface{} {
	return map[string]interface{}{
		"_id":         "17dbokav5t6pssjv72gm0nie3u25b54m",
		"deviceId":    deviceID,
		"deviceTime":  "2017-11-05T12:56:51",
		"id":          "3f0075ad57ad603c83dc1e1a76aefcaf",
		"localTime":   "2017-11-05T12:56:51.000Z",
		"_userId":     "87df73fd41",
		"_groupId":    "8da6e693b8",
		"createdTime": "2022-06-21T22:40:07.732+00:00",
		"_version":    0,
		"_active":     true,
		"uploadId":    "a21c82a5f5d2860add2539acded6b614",
		"time":        "2022-06-21T22:40:07.732+00:00",
	}
}

// annotations and payload as a string rather than object or array
func dexG5MobDatum() map[string]interface{} {
	datum := base("DexG5Mob_iPhone")
	datum["annotations"] = `[{"code":"bg/out-of-range","threshold":40,"value":"low"}]`
	datum["payload"] = `{"systemTime":"2017-11-05T18:56:51Z","transmitterId":"410X6M","transmitterTicks":5796922,"trend":"flat","trendRate":0.6,"trendRateUnits":"mg/dL/min"}`
	datum["type"] = "cbg"
	datum["units"] = "mmol/L"
	datum["value"] = 8.1596
	return datum
}

func tandemPumpSettingsDatum() map[string]interface{} {
	datum := base("tandem99999999")

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

func carelinkPumpSettings() map[string]interface{} {
	datum := base("MiniMed 530G - 751-=-11111111")

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

func omnipodPumpSettingsDatum() map[string]interface{} {

	datum := base("InsOmn-837268")
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

func omnipodPumpSettingsDatumTargetSet() map[string]interface{} {

	datum := base("InsOmn-837268")
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

func tandemAutomatedBasalDatum() map[string]interface{} {
	datum := base("tandemCIQ1111111111111")
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

func tandemWizardDatum() map[string]interface{} {
	datum := base("tandemCIQ1111111111111")
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
	datum["duration"] = 300000
	datum["rate"] = 0.335
	datum["percent"] = 0.47857142857142865
	datum["conversionOffset"] = 0
	datum["bolus"] = "g2h6nohp5sdndpvl2l8kdete00lle4gt"

	return datum
}

func reservoirChangeDeviceEventDatum() map[string]interface{} {
	datum := base("InsOmn-1111111111111")
	datum["type"] = "deviceEvent"
	datum["subType"] = "reservoirChange"
	datum["status"] = "cvv61jde62b6i28bgot57f18bor5au1n"
	return datum
}

func cgmSettingsDatum() map[string]interface{} {
	datum := base("DexG5MobRec-1111111111111")
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

func emptyPayload() map[string]interface{} {
	datum := base("Dex-device")
	datum["payload"] = map[string]interface{}{}
	datum["type"] = "cbg"
	datum["units"] = "mmol/L"
	datum["value"] = 8.1596
	return datum
}

func pumpSettingsWithBolus() map[string]interface{} {
	datum := tandemPumpSettingsDatum()

	datum["bolus"] = &pump.BolusMap{
		"bolus-1": pumpTest.NewRandomBolus(),
		"bolus-2": pumpTest.NewRandomBolus(),
	}

	return datum
}

var CBGDexcomG5MobDatum = dexG5MobDatum()
var PumpSettingsTandem = tandemPumpSettingsDatum()
var PumpSettingsCarelink = carelinkPumpSettings()
var PumpSettingsOmnipod = omnipodPumpSettingsDatum()
var PumpSettingsOmnipodBGTargetCorrect = omnipodPumpSettingsDatumTargetSet()
var AutomatedBasalTandem = tandemAutomatedBasalDatum()
var WizardTandem = tandemWizardDatum()
var ReservoirChange = reservoirChangeDeviceEventDatum()
var CGMSetting = cgmSettingsDatum()
var EmptyPayloadDatum = emptyPayload()
var PumpSettingsWithBolusDatum = pumpSettingsWithBolus()
