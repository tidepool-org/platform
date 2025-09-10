package util

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/log"
)

const (
	DefaultName     = "Unknown"
	DefaultBirthday = "1900-01-01"
)

type SeagullUserDocument struct {
	Profile Profile `json:"profile"`
}

type Profile struct {
	FullName *string        `json:"fullName"`
	Patient  PatientProfile `json:"patient"`
}

type PatientProfile struct {
	Birthday      *string `json:"birthday"`
	IsOtherPerson bool    `json:"isOtherPerson"`
	FullName      *string `json:"fullName"`
}

func PopulateCreateFromSeagullDocumentValue(value string, create *consent.RecordCreate, logger log.Logger) {
	userDocument := SeagullUserDocument{}
	if value != "" {
		if err := json.NewDecoder(strings.NewReader(value)).Decode(&userDocument); err != nil {
			logger.Warnf("error decoding user document")
		}
	}

	// Provide defaults if profile data is invalid or missing
	ownerName := DefaultName
	parentGuardianName := DefaultName
	birthday := DefaultBirthday

	profile := userDocument.Profile
	if profile.Patient.Birthday == nil {
		logger.Warnf("birthday is missing")
	} else {
		birthday = *profile.Patient.Birthday
	}

	if profile.Patient.IsOtherPerson {
		// This is a "fake child" account that was created on behalf of somebody else
		if profile.Patient.FullName != nil {
			ownerName = *profile.Patient.FullName
		} else {
			logger.Warnf("full name of fake child account is missing")
		}
		if profile.FullName != nil {
			parentGuardianName = *profile.FullName
		} else {
			logger.Warnf("full name of parent is missing")
		}
	} else if profile.FullName != nil {
		ownerName = *profile.FullName
	} else {
		logger.Warnf("owner full name is missing")
	}

	// Determine age at grant time
	parsed, err := time.Parse(time.DateOnly, birthday)
	if err != nil {
		parsed = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
		logger.WithError(err).Warnf("unable to parse birthday")
	}
	if age := yearsDifference(create.GrantTime, parsed); age >= 18 {
		create.AgeGroup = consent.AgeGroupEighteenOrOver
		create.GrantorType = consent.GrantorTypeOwner
		create.OwnerName = ownerName
	} else {
		if age >= 13 {
			create.AgeGroup = consent.AgeGroupThirteenSeventeen
		} else {
			create.AgeGroup = consent.AgeGroupUnderThirteen
		}
		create.GrantorType = consent.GrantorTypeParentGuardian
		create.ParentGuardianName = &parentGuardianName
		create.OwnerName = ownerName
	}
}

func yearsDifference(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}

	years := end.Year() - start.Year()
	if end.YearDay() < start.YearDay() {
		years--
	}

	return years
}
