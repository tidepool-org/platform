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
	"time"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
)

const SchemaVersionCurrent = 3

type Base struct {
	Active        bool   `json:"-" bson:"_active"`
	CreatedTime   string `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	GroupID       string `json:"-" bson:"_groupId,omitempty"`
	GUID          string `json:"guid,omitempty" bson:"guid,omitempty"`
	ID            string `json:"id,omitempty" bson:"id,omitempty"`
	SchemaVersion int    `json:"-" bson:"_schemaVersion,omitempty"`
	Type          string `json:"type,omitempty" bson:"type,omitempty"`
	UploadID      string `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID        string `json:"-" bson:"_userId,omitempty"`
	Version       int    `json:"-" bson:"_version,omitempty"`

	Annotations      *[]interface{} `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ClockDriftOffset *int           `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ConversionOffset *int           `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	DeviceID         *string        `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceTime       *string        `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	Payload          *interface{}   `json:"payload,omitempty" bson:"payload,omitempty"`
	Time             *string        `json:"time,omitempty" bson:"time,omitempty"`
	TimezoneOffset   *int           `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
}

type Meta struct {
	Type string `json:"type,omitempty"`
}

func New(baseType string) (*Base, error) {
	if baseType == "" {
		return nil, app.Error("base", "type is missing")
	}

	return &Base{
		Active:        false,
		CreatedTime:   time.Now().UTC().Format(time.RFC3339),
		GUID:          app.NewUUID(),
		ID:            app.NewID(), // TODO: Move calculation to Normalize to follow Jellyfish algorithm
		SchemaVersion: SchemaVersionCurrent,
		Type:          baseType,
		Version:       0,
	}, nil
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
	b.Time = parser.ParseString("time")
	b.TimezoneOffset = parser.ParseInteger("timezoneOffset")

	return nil
}

func (b *Base) Validate(validator data.Validator) error {
	// validator.ValidateInterfaceArray("annotations", b.Annotations)    // TODO: Any validations? Optional? Size?
	// validator.ValidateInteger("clockDriftOffset", b.ClockDriftOffset) // TODO: Any validations? Optional? Range?
	// validator.ValidateInteger("conversionOffset", b.ConversionOffset) // TODO: Any validations? Optional? Range?
	validator.ValidateString("deviceId", b.DeviceID).Exists().NotEmpty()
	validator.ValidateStringAsTime("deviceTime", b.DeviceTime, "2006-01-02T15:04:05") // TODO: Not in upload!  -> .Exists()
	// validator.ValidateInterface("payload", b.Payload) // TODO: Any validations? Optional? Size?
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
