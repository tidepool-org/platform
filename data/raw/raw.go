package raw

import (
	"io"
	"regexp"
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	FilterCreatedDateFormat = time.DateOnly
	DataSizeMaximum         = 8 * 1024 * 1024 // Until create directly to S3 is supported
	MetadataSizeMaximum     = 4 * 1024
	MediaTypeDefault        = "application/octet-stream"
)

type Filter struct {
	CreatedDate *string `json:"createdDate,omitempty"`
	DataSetID   *string `json:"dataSetId,omitempty"`
	Processed   *bool   `json:"processed,omitempty"`
	Archivable  *bool   `json:"archivable,omitempty"`
	Archived    *bool   `json:"archived,omitempty"`
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	f.CreatedDate = parser.String("createdDate")
	f.DataSetID = parser.String("dataSetId")
	f.Processed = parser.Bool("processed")
	f.Archivable = parser.Bool("archivable")
	f.Archived = parser.Bool("archived")
}

func (f *Filter) Validate(validator structure.Validator) {
	validator.String("createdDate", f.CreatedDate).AsTime(FilterCreatedDateFormat).NotZero()
	validator.String("dataSetId", f.DataSetID).Using(data.SetIDValidator)
}

func (f *Filter) CreatedTime() *time.Time {
	if f.CreatedDate == nil {
		return nil
	}
	if createdDate, err := time.Parse(FilterCreatedDateFormat, *f.CreatedDate); err != nil {
		return nil
	} else {
		return pointer.FromTime(createdDate)
	}
}

type Create struct {
	Metadata       map[string]any `json:"metadata,omitempty"`
	DigestMD5      *string        `json:"digestMD5,omitempty"`
	MediaType      *string        `json:"mediaType,omitempty"`
	ArchivableTime *time.Time     `json:"archivableTime,omitempty"`
}

func (c *Create) Parse(parser structure.ObjectParser) {
	if ptr := parser.Object("metadata"); ptr != nil {
		c.Metadata = *ptr
	}
	c.DigestMD5 = parser.String("digestMD5")
	c.MediaType = parser.String("mediaType")
	c.ArchivableTime = parser.Time("archivableTime", time.RFC3339Nano)
}

func (c *Create) Validate(validator structure.Validator) {
	validator.Object("metadata", &c.Metadata).SizeLessThanOrEqualTo(MetadataSizeMaximum)
	validator.String("digestMD5", c.DigestMD5).Using(net.DigestMD5Validator)
	validator.String("mediaType", c.MediaType).Using(net.MediaTypeValidator)
	validator.Time("archivableTime", c.ArchivableTime).NotZero()
}

type Content struct {
	DigestMD5  string        `json:"digestMD5"`
	MediaType  string        `json:"mediaType"`
	ReadCloser io.ReadCloser `json:"-"`
}

func (c *Content) Validate(validator structure.Validator) {
	validator.String("digestMD5", &c.DigestMD5).Using(net.DigestMD5Validator)
	validator.String("mediaType", &c.MediaType).Using(net.MediaTypeValidator)
	if c.ReadCloser == nil {
		validator.WithReference("readCloser").ReportError(structureValidator.ErrorValueNotExists())
	}
}

type Update struct {
	ProcessedTime  *time.Time      `json:"processedTime,omitempty"`
	ArchivableTime *time.Time      `json:"archivableTime,omitempty"`
	ArchivedTime   *time.Time      `json:"archivedTime,omitempty"`
	Metadata       *map[string]any `json:"metadata,omitempty"`
}

func (u *Update) Parse(parser structure.ObjectParser) {
	u.ProcessedTime = parser.Time("processedTime", time.RFC3339Nano)
	u.ArchivableTime = parser.Time("archivableTime", time.RFC3339Nano)
	u.ArchivedTime = parser.Time("archivedTime", time.RFC3339Nano)
	u.Metadata = parser.Object("metadata")
}

func (u *Update) Validate(validator structure.Validator) {
	validator.Time("processedTime", u.ProcessedTime).NotZero().BeforeNow(time.Second)
	validator.Time("archivableTime", u.ArchivableTime).NotZero()
	validator.Time("archivedTime", u.ArchivedTime).NotZero().BeforeNow(time.Second)
	validator.Object("metadata", u.Metadata).SizeLessThanOrEqualTo(MetadataSizeMaximum)

	if u.ProcessedTime == nil && u.ArchivableTime == nil && u.ArchivedTime == nil && u.Metadata == nil {
		validator.ReportError(structureValidator.ErrorValuesNotExistForAny("archivableTime", "archivedTime", "metadata", "processedTime"))
	}
}

type Raw struct {
	ID             string         `json:"id"`
	UserID         string         `json:"userId"`
	DataSetID      string         `json:"dataSetId"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	DigestMD5      string         `json:"digestMD5"`
	MediaType      string         `json:"mediaType"`
	Size           int            `json:"size"`
	ProcessedTime  *time.Time     `json:"processedTime,omitempty"`
	ArchivableTime *time.Time     `json:"archivableTime,omitempty"`
	ArchivedTime   *time.Time     `json:"archivedTime,omitempty"`
	CreatedTime    time.Time      `json:"createdTime"`
	ModifiedTime   *time.Time     `json:"modifiedTime,omitempty"`
	Revision       int            `json:"revision"`
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
	if ptr := parser.Object("metadata"); ptr != nil {
		r.Metadata = *ptr
	}
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
	r.ArchivableTime = parser.Time("archivableTime", time.RFC3339Nano)
	r.ArchivedTime = parser.Time("archivedTime", time.RFC3339Nano)
	if ptr := parser.Time("createdTime", time.RFC3339Nano); ptr != nil {
		r.CreatedTime = *ptr
	}
	r.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	if ptr := parser.Int("revision"); ptr != nil {
		r.Revision = *ptr
	}
}

func (r *Raw) Validate(validator structure.Validator) {
	validator.String("id", &r.ID).Using(DataRawIDValidator)
	validator.String("userId", &r.UserID).Using(user.IDValidator)
	validator.String("dataSetId", &r.DataSetID).Using(data.SetIDValidator)
	validator.Object("metadata", &r.Metadata).SizeLessThanOrEqualTo(MetadataSizeMaximum)
	validator.String("digestMD5", &r.DigestMD5).Using(net.DigestMD5Validator)
	validator.String("mediaType", &r.MediaType).Using(net.MediaTypeValidator)
	validator.Int("size", &r.Size).GreaterThanOrEqualTo(0)
	validator.Time("processedTime", r.ProcessedTime).After(r.CreatedTime).BeforeNow(time.Second)
	validator.Time("archivableTime", r.ArchivableTime).After(r.CreatedTime)
	validator.Time("archivedTime", r.ArchivedTime).After(r.CreatedTime).BeforeNow(time.Second)
	validator.Time("createdTime", &r.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", r.ModifiedTime).After(r.CreatedTime).BeforeNow(time.Second)
	validator.Int("revision", &r.Revision).GreaterThanOrEqualTo(0)
}

func (r *Raw) IsProcessed() bool {
	return r.ProcessedTime != nil
}

func (r *Raw) IsArchivable() bool {
	return r.ArchivableTime != nil && r.ArchivableTime.Before(time.Now())
}

func (r *Raw) IsArchived() bool {
	return r.ArchivedTime != nil
}

func IsValidDataRawID(value string) bool {
	return ValidateDataRawID(value) == nil
}

func DataRawIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateDataRawID(value))
}

func ValidateDataRawID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if matches := idExpression.FindStringSubmatch(value); len(matches) != 3 {
		return ErrorValueStringAsDataRawIDNotValid(value)
	} else if _, err := time.Parse(FilterCreatedDateFormat, matches[2]); err != nil {
		return ErrorValueStringAsDataRawIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsDataRawIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as data raw id", value)
}

var idExpression = regexp.MustCompile("^([0-9a-f]{24}):([0-9]{4}-[0-9]{2}-[0-9]{2})$")
