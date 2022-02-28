package schema

import (
	"time"

	"github.com/google/uuid"

	"github.com/tidepool-org/platform/data/types/basal/automated"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/errors"
)

const ErrEventTime = "unable to parse event time"

func (s *BasalSample) MapForAutomatedBasal(event *automated.Automated) error {
	var err error
	// assign an uuid for keeping a link between the two collection
	event.InternalID = uuid.New().String()

	// map
	s.DeliveryType = event.DeliveryType
	s.Duration = *event.Duration
	s.Rate = *event.Rate
	s.Timezone = *event.TimeZoneName
	s.TimezoneOffset = *event.TimeZoneOffset
	s.InternalID = event.InternalID
	strTime := *event.Time
	s.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)

	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (s *BasalSample) MapForScheduledBasal(event *scheduled.Scheduled) error {
	var err error
	// assign an uuid for keeping a link between the two collection
	event.InternalID = uuid.New().String()

	// map
	s.DeliveryType = event.DeliveryType
	s.Duration = *event.Duration
	s.Rate = *event.Rate
	s.Timezone = *event.TimeZoneName
	s.TimezoneOffset = *event.TimeZoneOffset
	s.InternalID = event.InternalID
	strTime := *event.Time
	s.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)

	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (c *CbgSample) Map(event *continuous.Continuous) error {
	var err error
	c.Value = *event.Value
	c.Units = *event.Units
	// extract string value (dereference)
	c.Timezone = *event.TimeZoneName
	c.TimezoneOffset = *event.TimeZoneOffset
	// what is this mess ???
	strTime := *event.Time
	c.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)

	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}
