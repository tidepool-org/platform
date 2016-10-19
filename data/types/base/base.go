package base

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
)

const SchemaVersionCurrent = 3

type Base struct {
	Active         bool   `json:"-" bson:"_active"`
	CreatedTime    string `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID  string `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	DeletedTime    string `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID  string `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	GroupID        string `json:"-" bson:"_groupId,omitempty"`
	GUID           string `json:"guid,omitempty" bson:"guid,omitempty"`
	ID             string `json:"id,omitempty" bson:"id,omitempty"`
	ModifiedTime   string `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID string `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	SchemaVersion  int    `json:"-" bson:"_schemaVersion,omitempty"`
	Type           string `json:"type,omitempty" bson:"type,omitempty"`
	UploadID       string `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID         string `json:"-" bson:"_userId,omitempty"`
	Version        int    `json:"-" bson:"_version,omitempty"`

	Annotations      *[]interface{} `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ClockDriftOffset *int           `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ConversionOffset *int           `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	DeviceID         *string        `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceTime       *string        `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	Payload          *interface{}   `json:"payload,omitempty" bson:"payload,omitempty"`
	Source           *string        `json:"source,omitempty" bson:"source,omitempty"`
	Time             *string        `json:"time,omitempty" bson:"time,omitempty"`
	TimezoneOffset   *int           `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
}

type Meta struct {
	Type string `json:"type,omitempty"`
}

func (b *Base) Init() {
	b.Active = false
	b.CreatedTime = ""
	b.CreatedUserID = ""
	b.DeletedTime = ""
	b.DeletedUserID = ""
	b.GroupID = ""
	b.GUID = app.NewUUID()
	b.ID = app.NewID() // TODO: Move calculation to Normalize to follow Jellyfish algorithm
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
	b.Annotations = parser.ParseInterfaceArray("annotations")
	b.ClockDriftOffset = parser.ParseInteger("clockDriftOffset")
	b.ConversionOffset = parser.ParseInteger("conversionOffset")
	b.DeviceID = parser.ParseString("deviceId")
	b.DeviceTime = parser.ParseString("deviceTime")
	b.Payload = parser.ParseInterface("payload")
	b.Source = parser.ParseString("source")
	b.Time = parser.ParseString("time")
	b.TimezoneOffset = parser.ParseInteger("timezoneOffset")

	return nil
}

func (b *Base) Validate(validator data.Validator) error {
	validator.ValidateString("type", &b.Type).Exists().NotEmpty()

	// validator.ValidateInterfaceArray("annotations", b.Annotations)    // TODO: Any validations? Optional? Size?
	// validator.ValidateInteger("clockDriftOffset", b.ClockDriftOffset) // TODO: Any validations? Optional? Range?
	// validator.ValidateInteger("conversionOffset", b.ConversionOffset) // TODO: Any validations? Optional? Range?
	validator.ValidateString("deviceId", b.DeviceID).Exists().NotEmpty()
	validator.ValidateStringAsTime("deviceTime", b.DeviceTime, "2006-01-02T15:04:05") // TODO: Not in upload!  -> .Exists()
	// validator.ValidateInterface("payload", b.Payload) // TODO: Any validations? Optional? Size?
	validator.ValidateString("source", b.Source).NotEmpty()
	validator.ValidateStringAsTime("time", b.Time, "2006-01-02T15:04:05Z07:00").Exists()
	// validator.ValidateInteger("timezoneOffset", b.TimezoneOffset) // TODO: Any validations? Optional? Range?

	// TODO: NOT IN UPLOAD: annotations, clockDriftOffset, deviceTime, payload

	return nil
}

func (b *Base) Normalize(normalizer data.Normalizer) error {
	return nil
}

func (b *Base) SetUserID(userID string) {
	b.UserID = userID
}

func (b *Base) SetGroupID(groupID string) {
	b.GroupID = groupID
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
