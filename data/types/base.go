package types

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
)

const SchemaVersionCurrent = 3
const DeviceTimeFormat = "2006-01-02T15:04:05"
const TimeFormat = "2006-01-02T15:04:05Z07:00"

type Base struct {
	Active            bool                         `json:"-" bson:"_active"`
	ArchivedDatasetID string                       `json:"-" bson:"archivedDatasetId,omitempty"`
	ArchivedTime      string                       `json:"-" bson:"archivedTime,omitempty"`
	CreatedTime       string                       `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID     string                       `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	Deduplicator      *data.DeduplicatorDescriptor `json:"-" bson:"_deduplicator,omitempty"`
	DeletedTime       string                       `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID     string                       `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	GUID              string                       `json:"guid,omitempty" bson:"guid,omitempty"`
	ID                string                       `json:"id,omitempty" bson:"id,omitempty"`
	ModifiedTime      string                       `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID    string                       `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	SchemaVersion     int                          `json:"-" bson:"_schemaVersion,omitempty"`
	Type              string                       `json:"type,omitempty" bson:"type,omitempty"`
	UploadID          string                       `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID            string                       `json:"-" bson:"_userId,omitempty"`
	Version           int                          `json:"-" bson:"_version,omitempty"`

	Annotations      *[]map[string]interface{} `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ClockDriftOffset *int                      `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ConversionOffset *int                      `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	DeviceID         *string                   `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceTime       *string                   `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	Payload          *map[string]interface{}   `json:"payload,omitempty" bson:"payload,omitempty"`
	Source           *string                   `json:"source,omitempty" bson:"source,omitempty"`
	Time             *string                   `json:"time,omitempty" bson:"time,omitempty"`
	TimezoneOffset   *int                      `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
}

type Meta struct {
	Type string `json:"type,omitempty"`
}

func (b *Base) Init() {
	b.Active = false
	b.ArchivedDatasetID = ""
	b.ArchivedTime = ""
	b.CreatedTime = ""
	b.CreatedUserID = ""
	b.Deduplicator = nil
	b.DeletedTime = ""
	b.DeletedUserID = ""
	b.GUID = id.New()
	b.ID = id.New() // TODO: Move calculation to Normalize to follow Jellyfish algorithm
	b.ModifiedTime = ""
	b.ModifiedUserID = ""
	b.SchemaVersion = SchemaVersionCurrent
	b.Type = ""
	b.UploadID = ""
	b.UserID = ""
	b.Version = 0

	b.Annotations = nil
	b.ClockDriftOffset = nil
	b.ConversionOffset = nil
	b.DeviceID = nil
	b.DeviceTime = nil
	b.Payload = nil
	b.Source = nil
	b.Time = nil
	b.TimezoneOffset = nil
}

func (b *Base) Meta() interface{} {
	return &Meta{
		Type: b.Type,
	}
}

func (b *Base) Parse(parser data.ObjectParser) error {
	b.Annotations = parser.ParseObjectArray("annotations")
	b.ClockDriftOffset = parser.ParseInteger("clockDriftOffset")
	b.ConversionOffset = parser.ParseInteger("conversionOffset")
	b.DeviceID = parser.ParseString("deviceId")
	b.DeviceTime = parser.ParseString("deviceTime")
	b.Payload = parser.ParseObject("payload")
	b.Source = parser.ParseString("source")
	b.Time = parser.ParseString("time")
	b.TimezoneOffset = parser.ParseInteger("timezoneOffset")

	return nil
}

func (b *Base) Validate(validator data.Validator) error {
	validator.ValidateString("type", &b.Type).NotEmpty()

	// validator.ValidateInterfaceArray("annotations", b.Annotations)    // TODO: Any validations? Optional? Size?
	// validator.ValidateInteger("clockDriftOffset", b.ClockDriftOffset) // TODO: Any validations? Optional? Range?
	// validator.ValidateInteger("conversionOffset", b.ConversionOffset) // TODO: Any validations? Optional? Range?
	validator.ValidateString("deviceId", b.DeviceID).Exists().NotEmpty()
	validator.ValidateStringAsTime("deviceTime", b.DeviceTime, DeviceTimeFormat) // TODO: Not in upload!  -> .Exists()
	// validator.ValidateInterface("payload", b.Payload) // TODO: Any validations? Optional? Size?
	validator.ValidateString("source", b.Source).NotEmpty()
	validator.ValidateStringAsTime("time", b.Time, TimeFormat).Exists()
	// validator.ValidateInteger("timezoneOffset", b.TimezoneOffset) // TODO: Any validations? Optional? Range?

	// TODO: NOT IN UPLOAD: annotations, clockDriftOffset, deviceTime, payload

	return nil
}

func (b *Base) Normalize(normalizer data.Normalizer) error {
	return nil
}

func (b *Base) IdentityFields() ([]string, error) {
	if b.UserID == "" {
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

	return []string{b.UserID, *b.DeviceID, *b.Time, b.Type}, nil
}

func (b *Base) GetPayload() *map[string]interface{} {
	return b.Payload
}

func (b *Base) SetUserID(userID string) {
	b.UserID = userID
}

func (b *Base) SetDatasetID(datasetID string) {
	b.UploadID = datasetID
}

func (b *Base) SetActive(active bool) {
	b.Active = active
}

func (b *Base) SetCreatedTime(createdTime string) {
	b.CreatedTime = createdTime
}

func (b *Base) SetCreatedUserID(createdUserID string) {
	b.CreatedUserID = createdUserID
}

func (b *Base) SetModifiedTime(modifiedTime string) {
	b.ModifiedTime = modifiedTime
}

func (b *Base) SetModifiedUserID(modifiedUserID string) {
	b.ModifiedUserID = modifiedUserID
}

func (b *Base) SetDeletedTime(deletedTime string) {
	b.DeletedTime = deletedTime
}

func (b *Base) SetDeletedUserID(deletedUserID string) {
	b.DeletedUserID = deletedUserID
}

func (b *Base) DeduplicatorDescriptor() *data.DeduplicatorDescriptor {
	return b.Deduplicator
}

func (b *Base) SetDeduplicatorDescriptor(deduplicatorDescriptor *data.DeduplicatorDescriptor) {
	b.Deduplicator = deduplicatorDescriptor
}
