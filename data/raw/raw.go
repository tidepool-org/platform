package raw

import (
	"io"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	FilterCreatedDateFormat       = time.DateOnly
	FilterDataSetIDsLengthMaximum = 100
	DataSizeMaximum               = 8 * 1024 * 1024 // Until create directly to S3 is supported
	MediaTypeDefault              = "application/octet-stream"
)

type Filter struct {
	CreatedDate *time.Time `json:"createdDate,omitempty"`
	DataSetIDs  *[]string  `json:"dataSetIds,omitempty"`
	Processed   *bool      `json:"processed,omitempty"`
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	f.CreatedDate = parser.Time("createdDate", FilterCreatedDateFormat)
	f.DataSetIDs = parser.StringArray("dataSetIds")
	f.Processed = parser.Bool("processed")
}

func (f *Filter) Validate(validator structure.Validator) {
	validator.Time("createdDate", f.CreatedDate).NotZero()
	validator.StringArray("dataSetIds", f.DataSetIDs).NotEmpty().LengthLessThanOrEqualTo(FilterDataSetIDsLengthMaximum).EachUsing(data.SetIDValidator).EachUnique()
}

func (f *Filter) CreatedTime() *time.Time {
	if f.CreatedDate == nil {
		return nil
	}
	year, month, day := f.CreatedDate.UTC().Date()
	return pointer.FromTime(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

type Create struct {
	Metadata  *metadata.Metadata `json:"metadata,omitempty"`
	DigestMD5 *string            `json:"digestMD5,omitempty"`
	MediaType *string            `json:"mediaType,omitempty"`
}

func NewCreate() *Create {
	return &Create{}
}

func (c *Create) Parse(parser structure.ObjectParser) {
	c.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
	c.DigestMD5 = parser.String("digestMD5")
	c.MediaType = parser.String("mediaType")
}

func (c *Create) Validate(validator structure.Validator) {
	if c.Metadata != nil {
		c.Metadata.Validate(validator.WithReference("metadata"))
	}
	validator.String("digestMD5", c.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", c.MediaType).Using(net.MediaTypeValidator)
}

type Content struct {
	DigestMD5  string        `json:"digestMD5,omitempty"`
	MediaType  string        `json:"mediaType,omitempty"`
	ReadCloser io.ReadCloser `json:"-"`
}

func NewContent() *Content {
	return &Content{}
}

func (c *Content) Validate(validator structure.Validator) {
	validator.String("digestMD5", &c.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", &c.MediaType).Using(net.MediaTypeValidator)
	if c.ReadCloser == nil {
		validator.WithReference("readCloser").ReportError(structureValidator.ErrorValueNotExists())
	}
}

type Update struct {
	ProcessedTime time.Time `json:"processedTime,omitempty"`
}

func NewUpdate() *Update {
	return &Update{}
}

func (u *Update) Parse(parser structure.ObjectParser) {
	if ptr := parser.Time("processedTime", time.RFC3339Nano); ptr != nil {
		u.ProcessedTime = *ptr
	}
}

func (u *Update) Validate(validator structure.Validator) {
	validator.Time("processedTime", &u.ProcessedTime).NotZero().BeforeNow(time.Second)
}

type Raw struct {
	ID            string             `json:"id,omitempty"`
	UserID        string             `json:"userId,omitempty"`
	DataSetID     string             `json:"dataSetId,omitempty"`
	Metadata      *metadata.Metadata `json:"metadata,omitempty"`
	DigestMD5     string             `json:"digestMD5,omitempty"`
	MediaType     string             `json:"mediaType,omitempty"`
	Size          int                `json:"size,omitempty"`
	ProcessedTime *time.Time         `json:"processedTime,omitempty"`
	CreatedTime   time.Time          `json:"createdTime,omitempty"`
	ModifiedTime  *time.Time         `json:"modifiedTime,omitempty"`
	Revision      int                `json:"revision,omitempty"`
}

func (r *Raw) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("id"); ptr != nil {
		r.ID = *ptr
	}
	if ptr := parser.String("userId"); ptr != nil {
		r.UserID = *ptr
	}
	if ptr := parser.String("dataSetId"); ptr != nil {
		r.DataSetID = *ptr
	}
	r.Metadata = metadata.ParseMetadata(parser.WithReferenceObjectParser("metadata"))
	if ptr := parser.String("digestMD5"); ptr != nil {
		r.DigestMD5 = *ptr
	}
	if ptr := parser.String("mediaType"); ptr != nil {
		r.MediaType = *ptr
	}
	if ptr := parser.Int("size"); ptr != nil {
		r.Size = *ptr
	}
	r.ProcessedTime = parser.Time("processedTime", time.RFC3339Nano)
	if ptr := parser.Time("createdTime", time.RFC3339Nano); ptr != nil {
		r.CreatedTime = *ptr
	}
	r.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	if ptr := parser.Int("revision"); ptr != nil {
		r.Revision = *ptr
	}
}

func (r *Raw) Validate(validator structure.Validator) {
	validator.String("id", &r.ID).Using(IDValidator)
	validator.String("userId", &r.UserID).Using(user.IDValidator)
	validator.String("dataSetId", &r.DataSetID).Using(data.SetIDValidator)
	if r.Metadata != nil {
		r.Metadata.Validate(validator.WithReference("metadata"))
	}
	validator.String("digestMD5", &r.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", &r.MediaType).Using(net.MediaTypeValidator)
	validator.Int("size", &r.Size).GreaterThanOrEqualTo(0)
	validator.Time("processedTime", r.ProcessedTime).NotZero().After(r.CreatedTime).BeforeNow(time.Second)
	validator.Time("createdTime", &r.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", r.ModifiedTime).NotZero().After(r.CreatedTime).BeforeNow(time.Second)
	validator.Int("revision", &r.Revision).GreaterThanOrEqualTo(0)
}

func (r *Raw) Processed() bool {
	return r.ProcessedTime != nil
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

var idExpression = regexp.MustCompile("^[0-9a-fA-F]{24}:[0-9]{8}$")
