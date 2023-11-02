package schema

import (
	"time"

	"github.com/tidepool-org/platform/data/types/device/calibration"
	"github.com/tidepool-org/platform/data/types/device/flush"
	"github.com/tidepool-org/platform/data/types/device/mode"
	"github.com/tidepool-org/platform/data/types/device/prime"
	"github.com/tidepool-org/platform/data/types/device/reservoirchange"
	"github.com/tidepool-org/platform/errors"
)

type (
	DeviceEventBucket struct {
		Id                string    `bson:"_id,omitempty"`
		CreationTimestamp time.Time `bson:"creationTimestamp,omitempty"`
		UserId            string    `bson:"userId,omitempty" `
		Day               time.Time `bson:"day,omitempty"` // ie: 2021-09-28
		Modes             []Mode    `bson:"modes"`
	}

	Duration struct {
		Units string  `bson:"units,omitempty"`
		Value float64 `bson:"value,omitempty"`
	}

	Mode struct {
		Sample         `bson:",inline"`
		SubType        string   `bson:"subType,omitempty"`
		DeviceId       string   `bson:"deviceId,omitempty"`
		Guid           string   `bson:"guid,omitempty"`
		Duration       Duration `bson:"duration,omitempty"`
		InputTimestamp string   `bson:"inputTimestamp,omitempty"`
	}

	Calibration struct {
		Sample   `bson:",inline"`
		SubType  string  `bson:"subType,omitempty"`
		DeviceId string  `bson:"deviceId,omitempty"`
		Guid     string  `bson:"guid,omitempty"`
		Units    string  `bson:"units,omitempty"`
		Value    float64 `bson:"value,omitempty"`
	}

	Flush struct {
		Sample     `bson:",inline"`
		DeviceId   string  `bson:"deviceId,omitempty"`
		Guid       string  `bson:"guid,omitempty"`
		Status     string  `bson:"status,omitempty"`
		StatusCode *int    `bson:"statusCode,omitempt"`
		Volume     float64 `bson:"volume,omitempty"`
	}

	Prime struct {
		Sample   `bson:",inline"`
		DeviceId string  `bson:"deviceId,omitempty"`
		Guid     string  `bson:"guid,omitempty"`
		Target   string  `bson:"primeTarget,omitempty"`
		Volume   float64 `bson:"volume,omitempty"`
	}

	ReservoirChange struct {
		Sample   `bson:",inline"`
		DeviceId string `bson:"deviceId,omitempty"`
		Guid     string `bson:"guid,omitempty"`
		Status   string `bson:"status,omitempty"`
	}
)

func (b DeviceEventBucket) GetId() string {
	return b.Id
}

func (s Mode) GetTimestamp() time.Time {
	return s.Timestamp
}

func (s *Mode) MapForMode(event *mode.Mode) error {
	var err error

	if event.GUID != nil {
		s.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		s.DeviceId = *event.DeviceID
	}

	//s.Duration= event.D
	s.SubType = event.SubType
	if event.Duration != nil {
		s.Duration = Duration{
			Units: *event.Duration.Units,
			Value: *event.Duration.Value,
		}
	}

	if event.InputTime != nil {
		s.InputTimestamp = *event.InputTime.InputTime
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

func (c Calibration) GetTimestamp() time.Time {
	return c.Timestamp
}

func (c *Calibration) MapForCalibration(event *calibration.Calibration) error {
	var err error
	if event.GUID != nil {
		c.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		c.DeviceId = *event.DeviceID
	}

	c.SubType = event.SubType
	if event.Units != nil {
		c.Units = *event.Units
	}
	if event.Value != nil {
		c.Value = *event.Value
	}

	// time infos mapping
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

func (f Flush) GetTimestamp() time.Time {
	return f.Timestamp
}

func (f *Flush) MapForFlush(event *flush.Flush) error {
	var err error
	if event.GUID != nil {
		f.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		f.DeviceId = *event.DeviceID
	}
	if event.Status != nil {
		f.Status = *event.Status
	}
	if event.StatusCode != nil {
		f.StatusCode = event.StatusCode
	}
	if event.Volume != nil {
		f.Volume = *event.Volume
	}

	// time infos mapping
	f.Timezone = *event.TimeZoneName
	f.TimezoneOffset = *event.TimeZoneOffset
	// what is this mess ???
	strTime := *event.Time
	f.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (p Prime) GetTimestamp() time.Time {
	return p.Timestamp
}

func (p *Prime) MapForPrime(event *prime.Prime) error {
	var err error
	if event.GUID != nil {
		p.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		p.DeviceId = *event.DeviceID
	}
	if event.Target != nil {
		p.Target = *event.Target
	}
	if event.Volume != nil {
		p.Volume = *event.Volume
	}

	// time infos mapping
	p.Timezone = *event.TimeZoneName
	p.TimezoneOffset = *event.TimeZoneOffset
	// what is this mess ???
	strTime := *event.Time
	p.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}

	return nil
}

func (r ReservoirChange) GetTimestamp() time.Time {
	return r.Timestamp
}

func (r *ReservoirChange) MapForReservoirChange(event *reservoirchange.ReservoirChange) error {
	var err error
	if event.GUID != nil {
		r.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		r.DeviceId = *event.DeviceID
	}

	if event.StatusID != nil {
		r.Status = *event.StatusID
	}
	// time infos mapping
	r.Timezone = *event.TimeZoneName
	r.TimezoneOffset = *event.TimeZoneOffset
	// what is this mess ???
	strTime := *event.Time
	r.Timestamp, err = time.Parse(time.RFC3339Nano, strTime)
	if err != nil {
		return errors.Wrap(err, ErrEventTime)
	}
	return nil
}
