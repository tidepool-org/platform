package schema

import (
	"time"

	"github.com/tidepool-org/platform/data/types/settings/basalsecurity"
	"github.com/tidepool-org/platform/errors"
)

type (

	// CurrentSettings definition of all elements that at part of the patient close loop setup
	// this definition does not contain new fields required for the new products (dbg2)
	// it is dedicated to dblg1
	CurrentSettings struct {
		UserId         string          `bson:"_id,omitempty" binding:"omitempty,max=128"`
		SecurityBasals *SecurityBasals `bson:"securityBasals" binding:"omitempty,dive"`
	}

	SecurityBasals struct {
		Timestamp      time.Time            `bson:"timestamp" binding:"required" `
		Timezone       string               `bson:"timezone" binding:"required,max=128" `
		TimezoneOffset int                  `bson:"timezoneOffset,omitempty" binding:"omitempty" `
		Rates          []SecurityBasalRates `bson:"rates" binding:"required"`
	}

	SecurityBasalRates struct {
		Rate *float32 `bson:"rate"`

		// Start of the security basal (in minutes) starting at midnight
		// start=10 means that the security basal start at 00:00 + 10 minutes so 00:10
		Start *int `bson:"start"`
	}
)

func (c CurrentSettings) GetId() string {
	return c.UserId
}

func (s SecurityBasals) GetTimestamp() time.Time {
	return s.Timestamp
}

func (s *SecurityBasals) MapForBasalSecurity(event *basalsecurity.BasalSecurity) error {
	var err error
	basalRateScheduleArray := event.BasalRateSchedule
	if basalRateScheduleArray != nil {
		for _, brs := range *basalRateScheduleArray {
			// convert start in ms to minutes
			item := SecurityBasalRates{}
			if brs.Start != nil {
				startInMin := *brs.Start / 60000 // 60*1000
				item.Start = &startInMin
			}
			if brs.Rate != nil {
				value := float32(*brs.Rate)
				item.Rate = &value
			}
			s.Rates = append(s.Rates, item)
		}
	}

	// time infos mapping
	s.Timezone = *event.TimeZoneName
	s.TimezoneOffset = *event.TimeZoneOffset
	// what is this mess ???
	strTime := *event.Time
	s.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}
