package schema

import (
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/automated"
	"github.com/tidepool-org/platform/data/types/basal/scheduled"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/bolus/biphasic"
	"github.com/tidepool-org/platform/data/types/bolus/normal"
	"github.com/tidepool-org/platform/data/types/bolus/pen"
	"github.com/tidepool-org/platform/errors"
)

const ErrEventTime = "unable to parse event time"

func (s *BolusSample) MapForNormalBolus(event *normal.Normal) error {
	var err error
	// assign an uuid for keeping a link between the two collection
	s.Uuid = *event.ID

	if event.GUID != nil {
		s.Guid = *event.GUID
	}

	if event.DeviceID != nil {
		s.DeviceId = *event.DeviceID
	}
	// bolus struct field
	s.BolusType = "normal"
	if event.InsulinOnBoard != nil && event.InsulinOnBoard.InsulinOnBoard != nil {
		s.InsulinOnBoard = *event.InsulinOnBoard.InsulinOnBoard
	}
	if event.Prescriptor != nil && event.Prescriptor.Prescriptor != nil {
		s.Prescriptor = *event.Prescriptor.Prescriptor
	}

	// normal field
	s.Normal = *event.Normal
	if event.NormalExpected != nil {
		s.ExpectedNormal = *event.NormalExpected
	}

	// map
	s.Timezone = *event.TimeZoneName
	s.TimezoneOffset = *event.TimeZoneOffset

	strTime := *event.Time
	s.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (s *BolusSample) MapForPenBolus(event *pen.Pen) error {
	var err error
	// assign an uuid for keeping a link between the two collection
	s.Uuid = *event.ID

	if event.GUID != nil {
		s.Guid = *event.GUID
	}

	if event.DeviceID != nil {
		s.DeviceId = *event.DeviceID
	}
	// bolus struct field
	s.BolusType = "pen"
	if event.InsulinOnBoard != nil && event.InsulinOnBoard.InsulinOnBoard != nil {
		s.InsulinOnBoard = *event.InsulinOnBoard.InsulinOnBoard
	}
	if event.Prescriptor != nil && event.Prescriptor.Prescriptor != nil {
		s.Prescriptor = *event.Prescriptor.Prescriptor
	}

	// normal field
	if event.Normal != nil {
		s.Normal = *event.Normal
	}

	// map
	s.Timezone = *event.TimeZoneName
	s.TimezoneOffset = *event.TimeZoneOffset

	strTime := *event.Time
	s.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (s *BolusSample) MapForBiphasicBolus(event *biphasic.Biphasic) error {
	var err error

	s.Uuid = *event.ID

	if event.GUID != nil {
		s.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		s.DeviceId = *event.DeviceID
	}
	// bolus struct field
	s.BolusType = "biphasic"
	if event.InsulinOnBoard != nil && event.InsulinOnBoard.InsulinOnBoard != nil {
		s.InsulinOnBoard = *event.InsulinOnBoard.InsulinOnBoard
	}
	if event.Prescriptor != nil && event.Prescriptor.Prescriptor != nil {
		s.Prescriptor = *event.Prescriptor.Prescriptor
	}

	s.Normal = *event.Normal.Normal
	if event.NormalExpected != nil {
		s.ExpectedNormal = *event.NormalExpected
	}

	// biphasic field
	if event.Part != nil {
		s.Part, _ = strconv.ParseInt(*event.Part, 10, 64)
	}

	// is a guid in fact
	s.BiphasicId = *event.BiphasicID

	// map
	s.Timezone = *event.TimeZoneName
	s.TimezoneOffset = *event.TimeZoneOffset

	strTime := *event.Time
	s.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

// Basal
func (b *BasalSample) mapForBasal(event *basal.Basal) error {
	var err error
	// assign an uuid for keeping a link between the two collection
	event.InternalID = uuid.New().String()

	// map
	b.DeliveryType = event.DeliveryType
	b.Duration = *event.Duration
	b.Rate = *event.Rate
	b.Timezone = *event.TimeZoneName
	b.TimezoneOffset = *event.TimeZoneOffset
	b.InternalID = event.InternalID
	strTime := *event.Time
	b.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if event.GUID != nil {
		b.Guid = *event.GUID
	}
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (b *BasalSample) MapForAutomatedBasal(event *automated.Automated) error {
	return b.mapForBasal(&event.Basal)
}

func (b *BasalSample) MapForScheduledBasal(event *scheduled.Scheduled) error {
	return b.mapForBasal(&event.Basal)
}

func (b *BasalSample) MapForTempBasal(event *temporary.Temporary) error {
	return b.mapForBasal(&event.Basal)
}

// CBG
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
