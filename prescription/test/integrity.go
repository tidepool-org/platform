package test

import (
	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/prescription"
)

var IntegrityAttributes = prescription.IntegrityAttributes{
	DataAttributes: prescription.DataAttributes{
		AccountType:        "caregiver",
		CaregiverFirstName: "Aliya",
		CaregiverLastName:  "Morissette",
		FirstName:          "Estella",
		LastName:           "Abbott",
		Birthday:           "1988-09-11",
		MRN:                "43730068-2",
		Email:              "alberta_weber@jacobikling.name",
		Sex:                "undisclosed",
		Weight: &prescription.Weight{
			Value: pointer.FromFloat64(93.5),
			Units: "kg",
		},
		YearOfDiagnosis: 1990,
		PhoneNumber: &prescription.PhoneNumber{
			CountryCode: 1,
			Number:      "888-555-4444",
		},
		InitialSettings: &prescription.InitialSettings{
			BloodGlucoseUnits: "mg/dL",
			BasalRateSchedule: &pump.BasalRateStartArray{
				&pump.BasalRateStart{
					Rate:  pointer.FromFloat64(0.5),
					Start: pointer.FromInt(1234567890),
				},
				&pump.BasalRateStart{
					Rate:  pointer.FromFloat64(0.7),
					Start: pointer.FromInt(34567890),
				},
			},
			BloodGlucoseTargetPhysicalActivity: &glucose.Target{
				High: pointer.FromFloat64(5.4),
				Low:  pointer.FromFloat64(3.0),
			},
			BloodGlucoseTargetPreprandial: &glucose.Target{
				High: pointer.FromFloat64(6.4),
				Low:  pointer.FromFloat64(3.5),
			},
			BloodGlucoseTargetSchedule: &pump.BloodGlucoseTargetStartArray{
				&pump.BloodGlucoseTargetStart{
					Target: glucose.Target{
						High: pointer.FromFloat64(5.3),
						Low:  pointer.FromFloat64(3.8),
					},
					Start: pointer.FromInt(1234567),
				},
				&pump.BloodGlucoseTargetStart{
					Target: glucose.Target{
						High: pointer.FromFloat64(6.7),
						Low:  pointer.FromFloat64(4.5),
					},
					Start: pointer.FromInt(2345678),
				},
			},
			CarbohydrateRatioSchedule: &pump.CarbohydrateRatioStartArray{
				&pump.CarbohydrateRatioStart{
					Amount: pointer.FromFloat64(4.50),
					Start:  pointer.FromInt(76543),
				},
				&pump.CarbohydrateRatioStart{
					Amount: pointer.FromFloat64(4.70),
					Start:  pointer.FromInt(12345),
				},
			},
			GlucoseSafetyLimit: pointer.FromFloat64(0.7),
			InsulinModel:       pointer.FromString("rapidAdult"),
			InsulinSensitivitySchedule: &pump.InsulinSensitivityStartArray{
				&pump.InsulinSensitivityStart{
					Amount: pointer.FromFloat64(0.8),
					Start:  pointer.FromInt(456789),
				},
				&pump.InsulinSensitivityStart{
					Amount: pointer.FromFloat64(0.9),
					Start:  pointer.FromInt(9876),
				},
			},
			BasalRateMaximum: &pump.BasalRateMaximum{
				Units: pointer.FromString("Units/hour"),
				Value: pointer.FromFloat64(0.2),
			},
			BolusAmountMaximum: &pump.BolusAmountMaximum{
				Units: pointer.FromString("Units"),
				Value: pointer.FromFloat64(0.2),
			},
			PumpID: "1234567890",
			CgmID:  "1234567890",
		},
		Calculator: &prescription.Calculator{
			Method:                        "weight",
			RecommendedBasalRate:          pointer.FromFloat64(0.1),
			RecommendedCarbohydrateRatio:  pointer.FromFloat64(1.0234),
			RecommendedInsulinSensitivity: pointer.FromFloat64(0.89),
			TotalDailyDose:                pointer.FromFloat64(2.33),
			TotalDailyDoseScaleFactor:     pointer.FromFloat64(1.02),
			Weight:                        pointer.FromFloat64(81.33),
			WeightUnits:                   pointer.FromString("kg"),
		},
		Training:                "inPerson",
		TherapySettings:         "transfer",
		PrescriberTermsAccepted: true,
		State:                   "submitted",
	},
	CreatedUserID: "1234567890",
}

// ExpectedHash is the expected integrity hash for the values above
const ExpectedHash = "994c84b67bbace6721fc490e2ede422a88c75b21b20d8d9480d3d85b726ce226dc1f035e9961e604e4c7b230459bddb1950df1e397b3ff518b9ce07c32bc1327"
