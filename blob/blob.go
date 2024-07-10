package blob

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	SizeMaximum     = 100 * 1024 * 1024
	StatusAvailable = "available"
	StatusCreated   = "created"
)

var (
	ANY   = []string{"any"}
	NONE  = []string{"none"}
	TRUES = []string{"true", "yes", "y", "1"}
)

type parsedQueryPermissions []string

// parsePermissions replicates the functionality of the seagull node.js
// parsePermissions function which parses a query string value to a permissions
// slice
func parsePermissions(queryValuesOfKey string) parsedQueryPermissions {
	queryValuesOfKey = strings.TrimSpace(queryValuesOfKey)
	if queryValuesOfKey == "" {
		return nil
	}
	var vals []string
	for _, val := range strings.Split(queryValuesOfKey, ",") {
		val = strings.TrimSpace(val)
		// Remove falsey values that are strings
		if val == "0" || val == "false" || val == "" {
			continue
		}
		vals = append(vals, val)
	}
	if slices.Compare(vals, ANY) == 0 {
		return slices.Clone(ANY)
	}
	if slices.Compare(vals, NONE) == 0 {
		return slices.Clone(NONE)
	}
	for _, val := range vals {
		nonEmpty := val != ""
		if !nonEmpty {
			return vals
		}
	}
	return nil
}

func arePermissionsValid(perms parsedQueryPermissions) bool {
	if len(perms) > 1 {
		union := append(ANY, NONE...)
		for _, perm := range perms {
			// quadratic time complexity but very few elements so don't care
			if slices.Contains(union, perm) {
				return false
			}
		}
	}
	return true
}

func arePermissionsSatisfied(queryPermissions, userPermissions parsedQueryPermissions) bool {
	if slices.Compare(queryPermissions, ANY) == 0 {
		nonEmpty := len(userPermissions) > 0
		return nonEmpty
	}
	if slices.Compare(queryPermissions, NONE) == 0 {
		empty := len(userPermissions) == 0
		return empty
	}
	// Todo: really test this part
	for _, userPerm := range userPermissions {
		if !slices.Contains(queryPermissions, userPerm) {
			return false
		}
	}
	return true
}

func Statuses() []string {
	return []string{
		StatusAvailable,
		StatusCreated,
	}
}

// FUTURE: Add DeleteAll

type Client interface {
	List(ctx context.Context, userID string, filter *Filter, pagination *page.Pagination) (BlobArray, error)
	Create(ctx context.Context, userID string, content *Content) (*Blob, error)
	CreateDeviceLogs(ctx context.Context, userID string, content *DeviceLogsContent) (*DeviceLogsBlob, error)
	DeleteAll(ctx context.Context, userID string) error
	Get(ctx context.Context, id string) (*Blob, error)
	GetContent(ctx context.Context, id string) (*Content, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (bool, error)
}

type Filter struct {
	MediaType *[]string
	Status    *[]string
}

func NewFilter() *Filter {
	return &Filter{}
}

func (f *Filter) Parse(parser structure.ObjectParser) {
	f.MediaType = parser.StringArray("mediaType")
	f.Status = parser.StringArray("status")
}

func (f *Filter) Validate(validator structure.Validator) {
	validator.StringArray("mediaType", f.MediaType).NotEmpty().EachUsing(net.MediaTypeValidator).EachUnique()
	validator.StringArray("status", f.Status).NotEmpty().EachOneOf(Statuses()...).EachUnique()
}

func (f *Filter) MutateRequest(req *http.Request) error {
	parameters := map[string][]string{}
	if f.MediaType != nil {
		parameters["mediaType"] = *f.MediaType
	}
	if f.Status != nil {
		parameters["status"] = *f.Status
	}
	return request.NewArrayParametersMutator(parameters).MutateRequest(req)
}

type Content struct {
	Body      io.ReadCloser
	DigestMD5 *string
	MediaType *string
}

func NewContent() *Content {
	return &Content{}
}

func (c *Content) Validate(validator structure.Validator) {
	if c.Body == nil {
		validator.WithReference("body").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("digestMD5", c.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", c.MediaType).Exists().Using(net.MediaTypeValidator)
}

type Blob struct {
	ID           *string    `json:"id,omitempty" bson:"id,omitempty"`
	UserID       *string    `json:"userId,omitempty" bson:"userId,omitempty"`
	DigestMD5    *string    `json:"digestMD5,omitempty" bson:"digestMD5,omitempty"`
	MediaType    *string    `json:"mediaType,omitempty" bson:"mediaType,omitempty"`
	Size         *int       `json:"size,omitempty" bson:"size,omitempty"`
	Status       *string    `json:"status,omitempty" bson:"status,omitempty"`
	CreatedTime  *time.Time `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime *time.Time `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	DeletedTime  *time.Time `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	Revision     *int       `json:"revision,omitempty" bson:"revision,omitempty"`
}

func (b *Blob) Parse(parser structure.ObjectParser) {
	b.ID = parser.String("id")
	b.UserID = parser.String("userId")
	b.DigestMD5 = parser.String("digestMD5")
	b.MediaType = parser.String("mediaType")
	b.Size = parser.Int("size")
	b.Status = parser.String("status")
	b.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	b.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	b.DeletedTime = parser.Time("deletedTime", time.RFC3339Nano)
	b.Revision = parser.Int("revision")
}

func (b *Blob) Validate(validator structure.Validator) {
	validator.String("id", b.ID).Exists().Using(IDValidator)
	validator.String("userId", b.UserID).Exists().Using(user.IDValidator)
	validator.String("digestMD5", b.DigestMD5).Exists().Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", b.MediaType).Exists().Using(net.MediaTypeValidator)
	validator.Int("size", b.Size).Exists().GreaterThanOrEqualTo(0)
	validator.String("status", b.Status).Exists().OneOf(Statuses()...)
	validator.Time("createdTime", b.CreatedTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", b.ModifiedTime).NotZero().After(pointer.ToTime(b.CreatedTime)).BeforeNow(time.Second)
	validator.Time("deletedTime", b.DeletedTime).NotZero().After(pointer.ToTime(b.CreatedTime)).BeforeNow(time.Second)
	validator.Int("revision", b.Revision).Exists().GreaterThanOrEqualTo(0)
}

type BlobArray []*Blob

type DeviceLogsContent struct {
	Body      io.ReadCloser
	DigestMD5 *string
	MediaType *string
	StartAt   *time.Time
	EndAt     *time.Time
}

func NewDeviceLogsContent() *DeviceLogsContent {
	return &DeviceLogsContent{}
}

func (c *DeviceLogsContent) Validate(validator structure.Validator) {
	if c.Body == nil {
		validator.WithReference("body").ReportError(structureValidator.ErrorValueNotExists())
	}
	validator.String("digestMD5", c.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", c.MediaType).Exists().Using(net.MediaTypeValidator)
	validator.Time("startAt", c.StartAt).Exists().NotZero()
	validator.Time("endAt", c.EndAt).Exists().NotZero()
}

type DeviceLogsBlob struct {
	ID          *string    `json:"id,omitempty" bson:"id,omitempty"`
	UserID      *string    `json:"userId,omitempty" bson:"userId,omitempty"`
	DigestMD5   *string    `json:"digestMD5,omitempty" bson:"digestMD5,omitempty"`
	MediaType   *string    `json:"mediaType,omitempty" bson:"mediaType,omitempty"`
	Size        *int       `json:"size,omitempty" bson:"size,omitempty"`
	CreatedTime *time.Time `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	StartAtTime *time.Time `json:"startAtTime,omitempty" bson:"startAtTime,omitempty"`
	EndAtTime   *time.Time `json:"endAtTime,omitempty" bson:"endAtTime,omitempty"`
	Revision    *int       `json:"revision,omitempty" bson:"revision,omitempty"`
}

type DeviceLogsBlobArray []*DeviceLogsBlob

func (b *DeviceLogsBlob) Parse(parser structure.ObjectParser) {
	b.ID = parser.String("id")
	b.UserID = parser.String("userId")
	b.DigestMD5 = parser.String("digestMD5")
	b.MediaType = parser.String("mediaType")
	b.Size = parser.Int("size")
	b.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	b.StartAtTime = parser.Time("startAtTime", time.RFC3339Nano)
	b.EndAtTime = parser.Time("endAtTime", time.RFC3339Nano)
	b.Revision = parser.Int("revision")
}

func (b *DeviceLogsBlob) Validate(validator structure.Validator) {
	validator.String("id", b.ID).Exists().Using(IDValidator)
	validator.String("userId", b.UserID).Exists().Using(user.IDValidator)
	validator.String("digestMD5", b.DigestMD5).Exists().Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", b.MediaType).Exists().Using(net.MediaTypeValidator)
	validator.Int("size", b.Size).Exists().GreaterThanOrEqualTo(0)
	validator.Time("createdTime", b.CreatedTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("startAtTime", b.StartAtTime).Exists().NotZero().BeforeNow(time.Second)
	validator.Time("endAtTime", b.EndAtTime).Exists().NotZero().After(pointer.ToTime(b.CreatedTime)).BeforeNow(time.Second)
	validator.Int("revision", b.Revision).Exists().GreaterThanOrEqualTo(0)
}

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

var idExpression = regexp.MustCompile("^[0-9a-z]{32}$")
