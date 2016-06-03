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

type Base struct {
	Active        bool   `json:"-" bson:"_active"`                                   // SET
	CreatedTime   string `json:"createdTime,omitempty" bson:"createdTime,omitempty"` // SET
	GroupID       string `json:"-" bson:"_groupId,omitempty"`                        // SET
	ID            string `json:"id,omitempty" bson:"id,omitempty"`                   // SET - old deduplication id???
	SchemaVersion int    `json:"-" bson:"_schemaVersion,omitempty"`                  // SET
	Type          string `json:"type,omitempty" bson:"type,omitempty"`               // SET (AFTER PARSE)
	UploadID      string `json:"uploadId,omitempty" bson:"uploadId,omitempty"`       // SET
	UserID        string `json:"userId,omitempty" bson:"userId,omitempty"`           // SET		// TODO: Should this be _userId in bson?, Should it be returned in JSON?

	Annotations      *[]interface{} `json:"annotations,omitempty" bson:"annotations,omitempty"`           // PARSE
	ClockDriftOffset *int           `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"` // PARSE
	ConversionOffset *int           `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"` // PARSE
	DeviceID         *string        `json:"deviceId,omitempty" bson:"deviceId,omitempty"`                 // PARSE
	DeviceTime       *string        `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`             // PARSE
	Payload          *interface{}   `json:"payload,omitempty" bson:"payload,omitempty"`                   // PARSE
	Time             *string        `json:"time,omitempty" bson:"time,omitempty"`                         // PARSE
	TimezoneOffset   *int           `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`     // PARSE
	Version          *int           `json:"-" bson:"_version,omitempty"`                                  // SET
}

type Meta struct {
	Type string `json:"type,omitempty"`
}

func New(Type string) (*Base, error) {
	if Type == "" {
		return nil, app.Error("base", "type is missing")
	}

	return &Base{
		Active:        false,
		CreatedTime:   time.Now().Format(time.RFC3339),
		ID:            app.NewID(), // TODO: Move calculation to Normalize to follow Jellyfish algorithm
		SchemaVersion: 3,           // TODO: Use constant here
		Type:          Type,
	}, nil
}

func (b *Base) Meta() interface{} {
	return &Meta{
		Type: b.Type,
	}
}

func (b *Base) Parse(parser data.ObjectParser) error {
	// b.Type = parser.ParseString("type")	// TODO_DATA: We do not parse out the type (this is set in the New function)
	// b.UserID = parser.ParseString("userId")	// TODO_DATA: We do not parse UserID, we set this when we receive the data for the target user
	b.DeviceID = parser.ParseString("deviceId")
	b.Time = parser.ParseString("time")
	// b.UploadID = parser.ParseString("uploadId") // TODO_DATA: We do not parse UploadID, we create and set this
	b.DeviceTime = parser.ParseString("deviceTime")
	b.TimezoneOffset = parser.ParseInteger("timezoneOffset")
	b.ConversionOffset = parser.ParseInteger("conversionOffset")
	b.ClockDriftOffset = parser.ParseInteger("clockDriftOffset")
	b.Payload = parser.ParseInterface("payload")
	b.Annotations = parser.ParseInterfaceArray("annotations")

	// b.GroupID = parser.ParseString("_groupId")	// TODO_DATA: We do not parse GroupID, we set this when we receive the data for the target user

	return nil
}

func (b *Base) Validate(validator data.Validator) error {
	// validator.ValidateString("createdTime", &b.CreatedTime).LengthGreaterThanOrEqualTo(1)
	// validator.ValidateString("type", &b.Type).LengthGreaterThanOrEqualTo(1)

	// validator.ValidateString("userId", b.UserID).Exists().LengthGreaterThanOrEqualTo(10)	// TODO_DATA: Validation is for parsed data only
	validator.ValidateString("deviceId", b.DeviceID).Exists().LengthGreaterThanOrEqualTo(1)
	validator.ValidateStringAsTime("time", b.Time, "2006-01-02T15:04:05Z07:00").Exists()
	// validator.ValidateString("uploadId", b.UploadID).Exists().LengthGreaterThanOrEqualTo(1) // TODO_DATA: Validation is for parsed data only
	// validator.ValidateString("_groupId", b.GroupID).Exists().LengthGreaterThanOrEqualTo(10) // TODO_DATA: Validation is for parsed data only

	validator.ValidateString("deviceTime", b.DeviceTime)

	//validator.ValidateStringAsTime("deviceTime", b.DeviceTime, "2006-01-02T15:04:05")
	validator.ValidateInteger("timezoneOffset", b.TimezoneOffset)
	// validator.ValidateInteger("conversionOffset", b.ConversionOffset).GreaterThanOrEqualTo(0) // TODO_DATA: Real data can have negative values
	// validator.ValidateInteger("clockDriftOffset", b.ClockDriftOffset).GreaterThanOrEqualTo(0) // TODO_DATA: Real data can have negative values
	validator.ValidateInterface("payload", b.Payload)
	validator.ValidateInterfaceArray("annotations", b.Annotations)

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
