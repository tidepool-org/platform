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

	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/pvn/data"
)

type Base struct {
	_ID      bson.ObjectId `bson:"_id"`
	ID       string        `json:"id" bson:"id"`
	UserID   *string       `json:"userId" bson:"userId"`
	DeviceID *string       `json:"deviceId" bson:"deviceId"`
	Time     *string       `json:"time" bson:"time"`
	Type     *string       `json:"type" bson:"type"`
	UploadID *string       `json:"uploadId" bson:"uploadId"`
	GroupID  *string       `json:"-" bson:"_groupId"`
	//optional data
	DeviceTime       *string        `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	TimezoneOffset   *int           `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
	ConversionOffset *int           `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	ClockDriftOffset *int           `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	Payload          *interface{}   `json:"payload,omitempty" bson:"payload,omitempty"`
	Annotations      *[]interface{} `json:"annotations,omitempty" bson:"annotations,omitempty"`
	//existing fields used for versioning and de-deping
	CreatedTime   string `json:"createdTime" bson:"createdTime"`
	ActiveFlag    bool   `json:"-" bson:"_active"`
	SchemaVersion int    `json:"-" bson:"_schemaVersion"`
	Version       int    `json:"-" bson:"_version,omitempty"`
}

func (b *Base) Parse(parser data.ObjectParser) {
	b.Type = parser.ParseString("type")
	b.UserID = parser.ParseString("userId")
	b.DeviceID = parser.ParseString("deviceId")
	b.Time = parser.ParseString("time")
	b.UploadID = parser.ParseString("uploadId")
	b.DeviceTime = parser.ParseString("deviceTime")
	b.TimezoneOffset = parser.ParseInteger("timezoneOffset")
	b.ConversionOffset = parser.ParseInteger("conversionOffset")
	b.ClockDriftOffset = parser.ParseInteger("clockDriftOffset")
	b.Payload = parser.ParseInterface("payload")
	b.Annotations = parser.ParseInterfaceArray("annotations")

	b.GroupID = parser.ParseString("_groupId")
}

func (b *Base) Validate(validator data.Validator) {
	validator.ValidateString("type", b.Type).Exists()
	validator.ValidateString("userId", b.UserID).Exists()
	validator.ValidateString("deviceId", b.DeviceID).Exists()
	validator.ValidateString("time", b.Time).Exists()
	validator.ValidateString("uploadId", b.UploadID).Exists()
	validator.ValidateString("_groupId", b.GroupID).Exists()

	validator.ValidateString("deviceTime", b.DeviceTime)
	validator.ValidateInteger("timezoneOffset", b.TimezoneOffset)
	validator.ValidateInteger("conversionOffset", b.ConversionOffset)
	validator.ValidateInteger("clockDriftOffset", b.ClockDriftOffset)
	validator.ValidateInterface("payload", b.Payload)
	validator.ValidateInterfaceArray("annotations", b.Annotations)
}

func (b *Base) Normalize(normalizer data.Normalizer) {

	b._ID = bson.NewObjectId()
	b.ID = bson.NewObjectId().Hex()
	b.ActiveFlag = false
	b.SchemaVersion = 10
	b.CreatedTime = time.Now().Format(time.RFC3339)

}
