package schema

import (
	"time"

	"github.com/google/uuid"

	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/automated"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/errors"
)

const ErrEventTime = "unable to parse event time"

func (s *BasalSample) mapForBasal(event *basal.Basal) error {
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
	if event.GUID != nil {
		s.Guid = *event.GUID
	}
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (s *BasalSample) MapForAutomatedBasal(event *automated.Automated) error {
	return s.mapForBasal(&event.Basal)
}

func (s *BasalSample) MapForScheduledBasal(event *scheduled.Scheduled) error {
	return s.mapForBasal(&event.Basal)
}

func (s *BasalSample) MapForTempBasal(event *temporary.Temporary) error {
	return s.mapForBasal(&event.Basal)
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
	if event.GUID != nil {
		c.Guid = *event.GUID
	}

	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}
