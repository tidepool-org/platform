package schema

import (
	"time"

	"github.com/tidepool-org/platform/data/types/device/alarm"
	"github.com/tidepool-org/platform/errors"
)

type (
	AlarmBucket struct {
		Id                string        `bson:"_id,omitempty"`
		CreationTimestamp time.Time     `bson:"creationTimestamp,omitempty"`
		UserId            string        `bson:"userId,omitempty" `
		Day               time.Time     `bson:"day,omitempty"` // ie: 2021-09-28
		Alarms            []AlarmSample `bson:"alarms"`
	}

	AlarmSample struct {
		Sample          `bson:",inline"`
		DeviceId        string `bson:"deviceId,omitempty"`
		Guid            string `bson:"guid,omitempty"`
		Code            string `bson:"code,omitempty"`
		Level           string `bson:"level,omitempty"`
		Type            string `bson:"type,omitempty"`
		AckStatus       string `bson:"ackStatus,omitempty"`
		UpdateTimestamp string `bson:"updateTimestamp,omitempty"`
	}
)

func (b AlarmBucket) GetId() string {
	return b.Id
}

func (s AlarmSample) GetTimestamp() time.Time {
	return s.Timestamp
}

func (s *AlarmSample) MapForAlarm(event *alarm.Alarm) error {
	var err error

	if event.GUID != nil {
		s.Guid = *event.GUID
	}
	if event.DeviceID != nil {
		s.DeviceId = *event.DeviceID
	}

	s.Code = *event.AlarmCode
	s.Level = *event.AlarmLevel
	s.Type = *event.AlarmType
	s.AckStatus = *event.AckStatus
	s.UpdateTimestamp = *event.UpdateTime

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
