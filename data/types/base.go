package types

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
)

const (
	ArchivedTimeFormat      = time.RFC3339
	ClockDriftOffsetMaximum = 24 * 60 * 60 * 1000  // TODO: Fix! Limit to reasonable values
	ClockDriftOffsetMinimum = -24 * 60 * 60 * 1000 // TODO: Fix! Limit to reasonable values
	CreatedTimeFormat       = time.RFC3339
	DeletedTimeFormat       = time.RFC3339
	DeviceTimeFormat        = "2006-01-02T15:04:05"
	ModifiedTimeFormat      = time.RFC3339
	SchemaVersionCurrent    = SchemaVersionMaximum
	SchemaVersionMaximum    = 3
	SchemaVersionMinimum    = 1
	TimeFormat              = time.RFC3339
	TimezoneOffsetMaximum   = 7 * 24 * 60  // TODO: Fix! Limit to reasonable values
	TimezoneOffsetMinimum   = -7 * 24 * 60 // TODO: Fix! Limit to reasonable values
	VersionMinimum          = 0
)

type Base struct {
	Active            bool                         `json:"-" bson:"_active,omitempty"`
	Annotations       *data.BlobArray              `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ArchivedDataSetID *string                      `json:"archivedDatasetId,omitempty" bson:"archivedDatasetId,omitempty"`
	ArchivedTime      *string                      `json:"archivedTime,omitempty" bson:"archivedTime,omitempty"`
	ClockDriftOffset  *int                         `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ConversionOffset  *int                         `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	CreatedTime       *string                      `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID     *string                      `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	Deduplicator      *data.DeduplicatorDescriptor `json:"deduplicator,omitempty" bson:"_deduplicator,omitempty"`
	DeletedTime       *string                      `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID     *string                      `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	DeviceID          *string                      `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceTime        *string                      `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	GUID              *string                      `json:"guid,omitempty" bson:"guid,omitempty"`
	ID                *string                      `json:"id,omitempty" bson:"id,omitempty"`
	ModifiedTime      *string                      `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID    *string                      `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	Payload           *data.Blob                   `json:"payload,omitempty" bson:"payload,omitempty"`
	SchemaVersion     int                          `json:"-" bson:"_schemaVersion,omitempty"`
	Source            *string                      `json:"source,omitempty" bson:"source,omitempty"`
	Time              *string                      `json:"time,omitempty" bson:"time,omitempty"`
	TimezoneOffset    *int                         `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
	Type              string                       `json:"type,omitempty" bson:"type,omitempty"`
	UploadID          *string                      `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID            *string                      `json:"-" bson:"_userId,omitempty"`
	Version           int                          `json:"-" bson:"_version,omitempty"`
}

type Meta struct {
	Type string `json:"type,omitempty"`
}

func (b *Base) Init() {
	b.Active = false
	b.Annotations = nil
	b.ArchivedDataSetID = nil
	b.ArchivedTime = nil
	b.ClockDriftOffset = nil
	b.ConversionOffset = nil
	b.CreatedTime = nil
	b.CreatedUserID = nil
	b.Deduplicator = nil
	b.DeletedTime = nil
	b.DeletedUserID = nil
	b.DeviceID = nil
	b.DeviceTime = nil
	b.GUID = nil
	b.ID = nil
	b.ModifiedTime = nil
	b.ModifiedUserID = nil
	b.Payload = nil
	b.SchemaVersion = 0
	b.Source = nil
	b.Time = nil
	b.TimezoneOffset = nil
	b.Type = ""
	b.UploadID = nil
	b.UserID = nil
	b.Version = 0
}

func (b *Base) Meta() interface{} {
	return &Meta{
		Type: b.Type,
	}
}

func (b *Base) Parse(parser data.ObjectParser) error {
	b.Annotations = data.ParseBlobArray(parser.NewChildArrayParser("annotations"))
	b.ClockDriftOffset = parser.ParseInteger("clockDriftOffset")
	b.ConversionOffset = parser.ParseInteger("conversionOffset")
	b.DeviceID = parser.ParseString("deviceId")
	b.DeviceTime = parser.ParseString("deviceTime")
	b.ID = parser.ParseString("id")
	b.Payload = data.ParseBlob(parser.NewChildObjectParser("payload"))
	b.Source = parser.ParseString("source")
	b.Time = parser.ParseString("time")
	b.TimezoneOffset = parser.ParseInteger("timezoneOffset")

	return nil
}

func (b *Base) Validate(validator structure.Validator) {
	var archivedTime time.Time
	var createdTime time.Time
	var modifiedTime time.Time

	if b.ArchivedTime != nil {
		archivedTime, _ = time.Parse(ArchivedTimeFormat, *b.ArchivedTime)
	}
	if b.CreatedTime != nil {
		createdTime, _ = time.Parse(CreatedTimeFormat, *b.CreatedTime)
	}
	if b.ModifiedTime != nil {
		modifiedTime, _ = time.Parse(ModifiedTimeFormat, *b.ModifiedTime)
	}

	if b.Annotations != nil {
		b.Annotations.Validate(validator.WithReference("annotations"))
	}

	if validator.Origin() <= structure.OriginInternal {
		if b.ArchivedTime != nil {
			validator.String("archivedDatasetId", b.ArchivedDataSetID).Exists().Using(data.ValidateDataSetID)
			validator.String("archivedTime", b.ArchivedTime).AsTime(ArchivedTimeFormat).After(createdTime).BeforeNow(time.Second)
		} else {
			validator.String("archivedDatasetId", b.ArchivedDataSetID).NotExists()
		}
	}

	validator.Int("clockDriftOffset", b.ClockDriftOffset).InRange(ClockDriftOffsetMinimum, ClockDriftOffsetMaximum)

	if validator.Origin() <= structure.OriginInternal {
		if b.CreatedTime != nil {
			validator.String("createdTime", b.CreatedTime).AsTime(CreatedTimeFormat).BeforeNow(time.Second)
			validator.String("createdUserId", b.CreatedUserID).Using(data.ValidateUserID)
		} else {
			validator.String("createdTime", b.CreatedTime).Exists()
			validator.String("createdUserId", b.CreatedUserID).NotExists()
		}

		if b.DeletedTime != nil {
			validator.String("deletedTime", b.DeletedTime).AsTime(DeletedTimeFormat).After(latestTime(archivedTime, createdTime, modifiedTime)).BeforeNow(time.Second)
			validator.String("deletedUserId", b.DeletedUserID).Using(data.ValidateUserID)
		} else {
			validator.String("deletedUserId", b.DeletedUserID).NotExists()
		}

		if b.Deduplicator != nil {
			b.Deduplicator.Validate(validator.WithReference("_deduplicator"))
		}
	}

	validator.String("deviceId", b.DeviceID).Exists().NotEmpty()
	validator.String("deviceTime", b.DeviceTime).AsTime(DeviceTimeFormat)

	validator.String("id", b.ID).Using(id.Validate)
	if validator.Origin() <= structure.OriginInternal {
		validator.String("id", b.ID).Exists()
	}

	if validator.Origin() <= structure.OriginInternal {
		if b.ModifiedTime != nil {
			validator.String("modifiedTime", b.ModifiedTime).AsTime(ModifiedTimeFormat).After(latestTime(archivedTime, createdTime)).BeforeNow(time.Second)
			validator.String("modifiedUserId", b.ModifiedUserID).Using(data.ValidateUserID)
		} else {
			if b.ArchivedTime != nil {
				validator.String("modifiedTime", b.ModifiedTime).Exists()
			}
			validator.String("modifiedUserId", b.ModifiedUserID).NotExists()
		}
	}

	if b.Payload != nil {
		b.Payload.Validate(validator.WithReference("payload"))
	}

	if validator.Origin() <= structure.OriginStore {
		validator.Int("_schemaVersion", &b.SchemaVersion).Exists().InRange(SchemaVersionMinimum, SchemaVersionMaximum)
	}

	validator.String("source", b.Source).EqualTo("carelink")
	validator.String("time", b.Time).Exists().AsTime(TimeFormat)
	validator.Int("timezoneOffset", b.TimezoneOffset).InRange(TimezoneOffsetMinimum, TimezoneOffsetMaximum)
	validator.String("type", &b.Type).Exists().NotEmpty()

	if validator.Origin() <= structure.OriginInternal {
		validator.String("uploadId", b.UploadID).Exists().Using(data.ValidateDataSetID)
	}
	if validator.Origin() <= structure.OriginStore {
		validator.String("_userId", b.UserID).Exists().Using(data.ValidateUserID)
		validator.Int("_version", &b.Version).Exists().GreaterThanOrEqualTo(VersionMinimum)
	}
}

func (b *Base) Normalize(normalizer data.Normalizer) {
	if b.Annotations != nil {
		b.Annotations.Normalize(normalizer.WithReference("annotations"))
	}
	if b.Deduplicator != nil {
		b.Deduplicator.Normalize(normalizer.WithReference("_deduplicator"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if b.GUID == nil {
			b.GUID = pointer.String(id.New())
		}
		if b.ID == nil {
			b.ID = pointer.String(id.New())
		}
	}

	if b.Payload != nil {
		b.Payload.Normalize(normalizer.WithReference("payload"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if b.SchemaVersion == 0 {
			b.SchemaVersion = SchemaVersionCurrent
		}
	}
}

func (b *Base) IdentityFields() ([]string, error) {
	if b.UserID == nil {
		return nil, errors.New("user id is missing")
	}
	if *b.UserID == "" {
		return nil, errors.New("user id is empty")
	}
	if b.DeviceID == nil {
		return nil, errors.New("device id is missing")
	}
	if *b.DeviceID == "" {
		return nil, errors.New("device id is empty")
	}
	if b.Time == nil {
		return nil, errors.New("time is missing")
	}
	if *b.Time == "" {
		return nil, errors.New("time is empty")
	}
	if b.Type == "" {
		return nil, errors.New("type is empty")
	}

	return []string{*b.UserID, *b.DeviceID, *b.Time, b.Type}, nil
}

func (b *Base) GetPayload() *data.Blob {
	return b.Payload
}

func (b *Base) SetUserID(userID *string) {
	b.UserID = userID
}

func (b *Base) SetDatasetID(datasetID *string) {
	b.UploadID = datasetID
}

func (b *Base) SetActive(active bool) {
	b.Active = active
}

func (b *Base) SetDeviceID(deviceID *string) {
	b.DeviceID = deviceID
}

func (b *Base) SetCreatedTime(createdTime *string) {
	b.CreatedTime = createdTime
}

func (b *Base) SetCreatedUserID(createdUserID *string) {
	b.CreatedUserID = createdUserID
}

func (b *Base) SetModifiedTime(modifiedTime *string) {
	b.ModifiedTime = modifiedTime
}

func (b *Base) SetModifiedUserID(modifiedUserID *string) {
	b.ModifiedUserID = modifiedUserID
}

func (b *Base) SetDeletedTime(deletedTime *string) {
	b.DeletedTime = deletedTime
}

func (b *Base) SetDeletedUserID(deletedUserID *string) {
	b.DeletedUserID = deletedUserID
}

func (b *Base) DeduplicatorDescriptor() *data.DeduplicatorDescriptor {
	return b.Deduplicator
}

func (b *Base) SetDeduplicatorDescriptor(deduplicatorDescriptor *data.DeduplicatorDescriptor) {
	b.Deduplicator = deduplicatorDescriptor
}

func latestTime(tms ...time.Time) time.Time {
	var latestTime time.Time
	for _, tm := range tms {
		if tm.After(latestTime) {
			latestTime = tm
		}
	}
	return latestTime
}
