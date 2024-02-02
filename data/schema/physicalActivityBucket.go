package schema

import (
	"time"

	"github.com/tidepool-org/platform/data/types/activity/physical"
	"github.com/tidepool-org/platform/errors"
)

type (
	PhysicalActivityBucket struct {
		Id                string             `bson:"_id,omitempty"`
		CreationTimestamp time.Time          `bson:"creationTimestamp,omitempty"`
		UserId            string             `bson:"userId,omitempty" `
		Day               time.Time          `bson:"day,omitempty"` // ie: 2021-09-28
		Samples           []PhysicalActivity `bson:"samples"`
	}

	PhysicalActivity struct {
		Sample            `bson:",inline"`
		Uuid              string   `bson:"uuid,omitempty"`
		Guid              string   `bson:"guid,omitempty"`
		DeviceId          string   `bson:"deviceId,omitempty"`
		Duration          Duration `bson:"duration,omitempty,omitempty"`
		ReportedIntensity string   `bson:"reportedIntensity,omitempty"`
		InputTimestamp    string   `bson:"inputTimestamp,omitempty"`
	}
)

func (p PhysicalActivityBucket) GetId() string {
	return p.Id
}

func (p PhysicalActivity) GetTimestamp() time.Time {
	return p.Timestamp
}
func (p *PhysicalActivity) MapForPhysical(event *physical.Physical) error {
	var err error

	if event.GUID != nil {
		p.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		p.DeviceId = *event.DeviceID
	}

	if event.Duration != nil {
		p.Duration = Duration{
			Units: *event.Duration.Units,
			Value: *event.Duration.Value,
		}
	}

	if event.ReportedIntensity != nil {
		p.ReportedIntensity = *event.ReportedIntensity
	}

	if event.InputTime != nil && event.InputTime.InputTime != nil {
		p.InputTimestamp = *event.InputTime.InputTime
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
