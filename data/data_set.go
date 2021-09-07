package data

import (
	"context"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"
	"encoding/json"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	timeZone "github.com/tidepool-org/platform/time/zone"
)

// TODO: This is a migration in progress from upload.Upload to DataSet. Some structures
// have been duplicated.

type DataSetAccessor interface {
	ListUserDataSets(ctx context.Context, userID string, filter *DataSetFilter, pagination *page.Pagination) (DataSets, error)
	CreateUserDataSet(ctx context.Context, userID string, create *DataSetCreate) (*DataSet, error)
	GetDataSet(ctx context.Context, id string) (*DataSet, error)
	UpdateDataSet(ctx context.Context, id string, update *DataSetUpdate) (*DataSet, error)
	DeleteDataSet(ctx context.Context, id string) error
}

const (
	ComputerTimeFormat = "2006-01-02T15:04:05"
	TimeFormat         = time.RFC3339Nano
	DeviceTimeFormat   = "2006-01-02T15:04:05"

	DataSetTypeContinuous = "continuous"
	DataSetTypeNormal     = "normal" // TODO: Normal?

	DataSetStateClosed = "closed"
	DataSetStateOpen   = "open"

	DeviceTagBGM         = "bgm"
	DeviceTagCGM         = "cgm"
	DeviceTagInsulinPump = "insulin-pump"

	TimeProcessingAcrossTheBoardTimeZone = "across-the-board-timezone" // TODO: Rename to across-the-board-time-zone or alternative
	TimeProcessingNone                   = "none"
	TimeProcessingUTCBootstrapping       = "utc-bootstrapping" // TODO: Rename to utc-bootstrap or alternative
)

func DataSetTypes() []string {
	return []string{
		DataSetTypeContinuous,
		DataSetTypeNormal,
	}
}

func DataSetStates() []string {
	return []string{
		DataSetStateClosed,
		DataSetStateOpen,
	}
}

func DeviceTags() []string {
	return []string{
		DeviceTagBGM,
		DeviceTagCGM,
		DeviceTagInsulinPump,
	}
}

func TimeProcessings() []string {
	return []string{
		TimeProcessingAcrossTheBoardTimeZone,
		TimeProcessingNone,
		TimeProcessingUTCBootstrapping,
	}
}

// TODO: Add OAuth client id to DataSetClient when available
// TODO: Pull from OAuth rather than be dependent upon client to complete

type DataSetClient struct {
	Name    *string                 `json:"name,omitempty" bson:"name,omitempty"`
	Version *string                 `json:"version,omitempty" bson:"version,omitempty"`
	Private *map[string]interface{} `json:"private,omitempty" bson:"private,omitempty"`
}

func NewDataSetClient() *DataSetClient {
	return &DataSetClient{}
}

func (d *DataSetClient) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("name"); ptr != nil {
		d.Name = ptr
	}
	if ptr := parser.String("version"); ptr != nil {
		d.Version = ptr
	}
	if ptr := parser.Object("private"); ptr != nil {
		d.Private = ptr
	}
}

func (d *DataSetClient) Validate(validator structure.Validator) {
	validator.String("name", d.Name).NotEmpty()
	validator.String("version", d.Version).NotEmpty() // TODO: Semver validation
}

type DataSetFilter struct {
	ClientName *string
	Deleted    *bool
	DeviceID   *string
}

func NewDataSetFilter() *DataSetFilter {
	return &DataSetFilter{}
}

func (d *DataSetFilter) Parse(parser structure.ObjectParser) {
	d.ClientName = parser.String("client.name")
	d.Deleted = parser.Bool("deleted")
	d.DeviceID = parser.String("deviceId")
}

func (d *DataSetFilter) Validate(validator structure.Validator) {
	validator.String("client.name", d.ClientName).NotEmpty()
	validator.String("deviceId", d.DeviceID).NotEmpty()
}

func (d *DataSetFilter) MutateRequest(req *http.Request) error {
	parameters := map[string]string{}
	if d.ClientName != nil {
		parameters["client.name"] = *d.ClientName
	}
	if d.Deleted != nil {
		parameters["deleted"] = strconv.FormatBool(*d.Deleted)
	}
	if d.DeviceID != nil {
		parameters["deviceId"] = *d.DeviceID
	}
	return request.NewParametersMutator(parameters).MutateRequest(req)
}

type DataSetCreate struct {
	Client              *DataSetClient          `json:"client,omitempty"`
	DataSetType         *string                 `json:"dataSetType,omitempty"`
	Deduplicator        *DeduplicatorDescriptor `json:"deduplicator,omitempty"`
	DeviceID            *string                 `json:"deviceId,omitempty"`
	DeviceManufacturers *[]string               `json:"deviceManufacturers,omitempty"`
	DeviceModel         *string                 `json:"deviceModel,omitempty"`
	DeviceSerialNumber  *string                 `json:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string               `json:"deviceTags,omitempty"`
	Time                *time.Time              `json:"time,omitempty"`
	Type                string                  `json:"type,omitempty"`
	TimeProcessing      *string                 `json:"timeProcessing,omitempty"`
	TimeZoneName        *string                 `json:"timezone,omitempty"`
	TimeZoneOffset      *int                    `json:"timezoneOffset,omitempty"`
}

func NewDataSetCreate() *DataSetCreate {
	return &DataSetCreate{
		Type: "upload",
	}
}

func (d *DataSetCreate) Parse(parser structure.ObjectParser) {
	if clientParser := parser.WithReferenceObjectParser("client"); clientParser.Exists() {
		d.Client = NewDataSetClient()
		d.Client.Parse(clientParser)
		clientParser.NotParsed()
	}
	d.DataSetType = parser.String("dataSetType")
	d.Deduplicator = ParseDeduplicatorDescriptor(parser.WithReferenceObjectParser("deduplicator"))
	d.DeviceID = parser.String("deviceId")
	d.DeviceManufacturers = parser.StringArray("deviceManufacturers")
	d.DeviceModel = parser.String("deviceModel")
	d.DeviceSerialNumber = parser.String("deviceSerialNumber")
	d.DeviceTags = parser.StringArray("deviceTags")
	d.Time = parser.Time("time", TimeFormat)
	d.TimeProcessing = parser.String("timeProcessing")
	d.TimeZoneName = parser.String("timezone")
	d.TimeZoneOffset = parser.Int("timezoneOffset")
}

func (d *DataSetCreate) Validate(validator structure.Validator) {
	if d.Client != nil {
		d.Client.Validate(validator.WithReference("client"))
	}
	validator.String("dataSetType", d.DataSetType).OneOf(DataSetTypes()...)
	if d.Deduplicator != nil {
		d.Deduplicator.Validate(validator.WithReference("deduplicator"))
	}
	validator.String("deviceId", d.DeviceID).NotEmpty()
	validator.StringArray("deviceManufacturers", d.DeviceManufacturers).NotEmpty()
	validator.String("deviceModel", d.DeviceModel).NotEmpty()
	validator.String("deviceSerialNumber", d.DeviceSerialNumber).NotEmpty()
	validator.StringArray("deviceTags", d.DeviceTags).NotEmpty().EachOneOf(DeviceTags()...)
	validator.Time("time", d.Time).NotZero()
	validator.String("timeProcessing", d.TimeProcessing).OneOf(TimeProcessings()...)
	validator.String("timezone", d.TimeZoneName).Using(timeZone.NameValidator)
	validator.Int("timezoneOffset", d.TimeZoneOffset).InRange(-12*60, 14*60)
}

func (d *DataSetCreate) Normalize(normalizer structure.Normalizer) {
	if d.Deduplicator != nil {
		d.Deduplicator.Normalize(normalizer.WithReference("deduplicator"))
	}
	if d.DeviceManufacturers != nil {
		sort.Strings(*d.DeviceManufacturers)
	}
	if d.DeviceTags != nil {
		sort.Strings(*d.DeviceTags)
	}
}

type DataSetUpdate struct {
	Active             *bool                   `json:"-"`
	DeviceID           *string                 `json:"deviceId,omitempty"`
	DeviceModel        *string                 `json:"deviceModel,omitempty"`
	DeviceSerialNumber *string                 `json:"deviceSerialNumber,omitempty"`
	Deduplicator       *DeduplicatorDescriptor `json:"-"`
	State              *string                 `json:"state,omitempty"`
	Time               *time.Time              `json:"time,omitempty"`
	TimeZoneName       *string                 `json:"timezone,omitempty"`
	TimeZoneOffset     *int                    `json:"timezoneOffset,omitempty"`
}

func NewDataSetUpdate() *DataSetUpdate {
	return &DataSetUpdate{}
}

func (d *DataSetUpdate) Parse(parser structure.ObjectParser) {
	d.DeviceID = parser.String("deviceId")
	d.DeviceModel = parser.String("deviceModel")
	d.DeviceSerialNumber = parser.String("deviceSerialNumber")
	d.State = parser.String("state")
	d.Time = parser.Time("time", TimeFormat)
	d.TimeZoneName = parser.String("timezone")
	d.TimeZoneOffset = parser.Int("timezoneOffset")
}

func (d *DataSetUpdate) Validate(validator structure.Validator) {
	validator.String("deviceId", d.DeviceID).NotEmpty()
	validator.String("deviceModel", d.DeviceModel).LengthGreaterThan(1)
	validator.String("deviceSerialNumber", d.DeviceSerialNumber).LengthGreaterThan(1)
	validator.String("state", d.State).OneOf(DataSetStates()...)
	validator.Time("time", d.Time).NotZero()
	validator.String("timezone", d.TimeZoneName).Using(timeZone.NameValidator)
	validator.Int("timezoneOffset", d.TimeZoneOffset).InRange(-12*60, 14*60)
}

func (d *DataSetUpdate) IsEmpty() bool {
	return d.Active == nil && d.DeviceID == nil && d.DeviceModel == nil && d.DeviceSerialNumber == nil &&
		d.Deduplicator == nil && d.State == nil && d.Time == nil && d.TimeZoneName == nil && d.TimeZoneOffset == nil
}

func NewSetID() string {
	return id.Must(id.New(16))
}

func IsValidSetID(value string) bool {
	return ValidateSetID(value) == nil
}

func SetIDValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateSetID(value))
}

func ValidateSetID(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if !setIDExpression.MatchString(value) {
		return ErrorValueStringAsSetIDNotValid(value)
	}
	return nil
}

func ErrorValueStringAsSetIDNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as data set id", value)
}

var setIDExpression = regexp.MustCompile("^(upid_[0-9a-f]{12}|upid_[0-9a-f]{32}|[0-9a-f]{32})$") // TODO: Want just "[0-9a-f]{32}"

type DataSet struct {
	Active              bool                    `json:"-" bson:"_active"`
	Annotations         *metadata.MetadataArray `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ArchivedDataSetID   *string                 `json:"-" bson:"archivedDatasetId,omitempty"`
	ArchivedTime        *time.Time                 `json:"-" bson:"archivedTime,omitempty"`
	ByUser              *string                 `json:"byUser,omitempty" bson:"byUser,omitempty"`
	Client              *DataSetClient          `json:"client,omitempty" bson:"client,omitempty"`
	ClockDriftOffset    *int                    `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ComputerTime        *time.Time              `json:"computerTime,omitempty" bson:"computerTime,omitempty"`
	ConversionOffset    *int                    `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	CreatedTime         *time.Time              `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID       *string                 `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	DataSetType         *string                 `json:"dataSetType,omitempty" bson:"dataSetType,omitempty"`
	DataState           *string                 `json:"-" bson:"_dataState,omitempty"` // TODO: Deprecated DataState (after data migration)
	Deduplicator        *DeduplicatorDescriptor `json:"deduplicator,omitempty" bson:"_deduplicator,omitempty"`
	DeletedTime         *time.Time              `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID       *string                 `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	DeviceID            *string                 `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceManufacturers *[]string               `json:"deviceManufacturers,omitempty" bson:"deviceManufacturers,omitempty"`
	DeviceModel         *string                 `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	DeviceSerialNumber  *string                 `json:"deviceSerialNumber,omitempty" bson:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string               `json:"deviceTags,omitempty" bson:"deviceTags,omitempty"`
	DeviceTime          *time.Time              `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	GUID                *string                 `json:"guid,omitempty" bson:"guid,omitempty"`
	ID                  *string                 `json:"id,omitempty" bson:"id,omitempty"`
	ModifiedTime        *time.Time              `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID      *string                 `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	Payload             *metadata.Metadata      `json:"payload,omitempty" bson:"payload,omitempty"`
	Source              *string                 `json:"source,omitempty" bson:"source,omitempty"`
	State               *string                 `json:"-" bson:"_state,omitempty"`
	Time                *time.Time              `json:"time,omitempty" bson:"time,omitempty"`
	TimeProcessing      *string                 `json:"timeProcessing,omitempty" bson:"timeProcessing,omitempty"`
	TimeZoneName        *string                 `json:"timezone,omitempty" bson:"timezone,omitempty"`
	TimeZoneOffset      *int                    `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
	Type                string                  `json:"type,omitempty" bson:"type,omitempty"`
	UploadID            *string                 `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID              *string                 `json:"-" bson:"_userId,omitempty"`
	Version             *string                 `json:"version,omitempty" bson:"version,omitempty"`
	VersionInternal     int                     `json:"-" bson:"_version,omitempty"`
}

// custom marshalling for DeviceTime
type DeviceTime time.Time

func (t DeviceTime) MarshalJSON() ([]byte, error) {
    b := make([]byte, 0, len(DeviceTimeFormat)+2)
    b = append(b, '"')
    b = time.Time(t).AppendFormat(b, DeviceTimeFormat)
    b = append(b, '"')
    return b, nil
}
func (d *DataSet) MarshalJSON() ([]byte, error) {
    type Alias DataSet
    dataSet := &struct {
        DeviceTime DeviceTime `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
        *Alias
    }{DeviceTime(*d.DeviceTime), (*Alias)(d)}

    return json.Marshal(dataSet)
}

type DataSets []*DataSet
