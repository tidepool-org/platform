package raw

import (
	"io"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

// TODO: How to make common with blob?

const (
	SizeMaximum     = 100 * 1024 * 1024 // TODO: Is this appropriate?
	StatusAvailable = "available"
	StatusCreated   = "created"
)

func Statuses() []string {
	return []string{
		StatusAvailable,
		StatusCreated,
	}
}

type Content struct {
	Body      io.ReadCloser
	DigestMD5 *string
}

func NewContent() *Content {
	return &Content{}
}

func (c *Content) Validate(validator structure.Validator) {
	if c.Body == nil {
		validator.WithReference("body").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("digestMD5", c.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
}

type Raw struct {
	ID           *string    `json:"id,omitempty" bson:"id,omitempty"`
	UserID       *string    `json:"userId,omitempty" bson:"userId,omitempty"`
	DataSetID    *string    `json:"dataSetId,omitempty" bson:"dataSetId,omitempty"`
	DigestMD5    *string    `json:"digestMD5,omitempty" bson:"digestMD5,omitempty"`
	Size         *int       `json:"size,omitempty" bson:"size,omitempty"`
	Status       *string    `json:"status,omitempty" bson:"status,omitempty"`
	CreatedTime  *time.Time `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime *time.Time `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	DeletedTime  *time.Time `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	Revision     *int       `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (r *Raw) Parse(parser structure.ObjectParser) {
	r.ID = parser.String("id")
	r.UserID = parser.String("userId")
	r.DataSetID = parser.String("dataSetId")
	r.DigestMD5 = parser.String("digestMD5")
	r.Size = parser.Int("size")
	r.Status = parser.String("status")
	r.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	r.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	r.DeletedTime = parser.Time("deletedTime", time.RFC3339Nano)
	r.Revision = parser.Int("revision")
}

func (r *Raw) Validate(validator structure.Validator) {
	validator.String("id", r.ID).Exists().Using(IDValidator)
	validator.String("userId", r.UserID).Exists().Using(user.IDValidator)
	validator.String("dataSetId", r.DataSetID).Exists().Using(data.SetIDValidator)
	validator.String("digestMD5", r.DigestMD5).Exists().Using(crypto.Base64EncodedMD5HashValidator)
	validator.Int("size", r.Size).Exists().GreaterThanOrEqualTo(0)
	validator.String("status", r.Status).Exists().OneOf(Statuses()...)
	validator.Time("createdTime", r.CreatedTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", r.ModifiedTime).NotZero().After(pointer.ToTime(r.CreatedTime)).BeforeNow(time.Second)
	validator.Time("deletedTime", r.DeletedTime).NotZero().After(pointer.ToTime(r.CreatedTime)).BeforeNow(time.Second)
	validator.Int("revision", r.Revision).Exists().GreaterThanOrEqualTo(0)
}

type RawArray []*Raw

// TODO: Move ID to common area

func NewID() string {
	return id.Must(id.New(16))
}

func IsValidID(value string) bool {
	return ValidateID(value) == nil
}

func IDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateID(value))
}

func ValidateID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !idExpression.MatchString(value) {
		return ErrorValueStringAsIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as raw id", value)
}

var idExpression = regexp.MustCompile("^[0-9a-f]{32}$")
