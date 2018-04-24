package data

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/structure"
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
	SchemaVersionCurrent = 3 // DEPRECATED

	ComputerTimeFormat = "2006-01-02T15:04:05"
	TimeFormat         = time.RFC3339

	DataSetTypeContinuous = "continuous"
	DataSetTypeNormal     = "normal" // TODO: Normal?

	DataSetStateClosed = "closed"
	DataSetStateOpen   = "open"

	DeviceTagBGM         = "bgm"
	DeviceTagCGM         = "cgm"
	DeviceTagInsulinPump = "insulin-pump"

	TimeProcessingAcrossTheBoardTimezone = "across-the-board-timezone"
	TimeProcessingNone                   = "none"
	TimeProcessingUTCBootstrapping       = "utc-bootstrapping"
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
		TimeProcessingAcrossTheBoardTimezone,
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
	Deleted  *bool   `json:"deleted,omitempty" bson:"deleted,omitempty"`
	DeviceID *string `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
}

func NewDataSetFilter() *DataSetFilter {
	return &DataSetFilter{}
}

func (d *DataSetFilter) Parse(parser structure.ObjectParser) {
	d.Deleted = parser.Bool("deleted")
	d.DeviceID = parser.String("deviceId")
}

func (d *DataSetFilter) Validate(validator structure.Validator) {
	validator.String("deviceId", d.DeviceID).NotEmpty()
}

func (d *DataSetFilter) Mutate(req *http.Request) error {
	parameters := map[string]string{}
	if d.Deleted != nil {
		parameters["deleted"] = strconv.FormatBool(*d.Deleted)
	}
	if d.DeviceID != nil {
		parameters["deviceId"] = *d.DeviceID
	}
	return client.NewParametersMutator(parameters).Mutate(req)
}

type DataSetCreate struct {
	Client              *DataSetClient `json:"client,omitempty"`
	DataSetType         string         `json:"dataSetType,omitempty"`
	DeviceID            string         `json:"deviceId,omitempty"`
	DeviceManufacturers []string       `json:"deviceManufacturers,omitempty"`
	DeviceModel         string         `json:"deviceModel,omitempty"`
	DeviceSerialNumber  string         `json:"deviceSerialNumber,omitempty"`
	DeviceTags          []string       `json:"deviceTags,omitempty"`
	Time                time.Time      `json:"time,omitempty"`
	Type                string         `json:"type,omitempty"`
	TimeProcessing      string         `json:"timeProcessing,omitempty"`
	Timezone            string         `json:"timezone,omitempty"`
	TimezoneOffset      int            `json:"timezoneOffset,omitempty"`
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
	if ptr := parser.String("dataSetType"); ptr != nil {
		d.DataSetType = *ptr
	}
	if ptr := parser.String("deviceId"); ptr != nil {
		d.DeviceID = *ptr
	}
	if ptr := parser.StringArray("deviceManufacturers"); ptr != nil {
		d.DeviceManufacturers = *ptr
	}
	if ptr := parser.String("deviceModel"); ptr != nil {
		d.DeviceModel = *ptr
	}
	if ptr := parser.String("deviceSerialNumber"); ptr != nil {
		d.DeviceSerialNumber = *ptr
	}
	if ptr := parser.StringArray("deviceTags"); ptr != nil {
		d.DeviceTags = *ptr
	}
	if ptr := parser.Time("time", TimeFormat); ptr != nil {
		d.Time = *ptr
	}
	if ptr := parser.String("timeProcessing"); ptr != nil {
		d.TimeProcessing = *ptr
	}
	if ptr := parser.String("timezone"); ptr != nil {
		d.Timezone = *ptr
	}
	if ptr := parser.Int("timezoneOffset"); ptr != nil {
		d.TimezoneOffset = *ptr
	}
}

func (d *DataSetCreate) Validate(validator structure.Validator) {
	if d.Client != nil {
		d.Client.Validate(validator.WithReference("client"))
	}
	validator.String("dataSetType", &d.DataSetType).OneOf(DataSetTypes()...)
	validator.String("deviceId", &d.DeviceID).NotEmpty()
	validator.StringArray("deviceManufacturers", &d.DeviceManufacturers).NotEmpty()
	validator.String("deviceModel", &d.DeviceModel).NotEmpty()
	validator.String("deviceSerialNumber", &d.DeviceSerialNumber).NotEmpty()
	validator.StringArray("deviceTags", &d.DeviceTags).NotEmpty().EachOneOf(DeviceTags()...)
	validator.Time("time", &d.Time).NotZero()
	validator.String("timeProcessing", &d.TimeProcessing).OneOf(TimeProcessings()...)
	validator.String("timezone", &d.Timezone) // TODO: Timezone
	validator.Int("timezoneOffset", &d.TimezoneOffset).InRange(-12*60, 14*60)
}

func (d *DataSetCreate) Normalize(normalizer structure.Normalizer) {
	sort.Strings(d.DeviceManufacturers)
	sort.Strings(d.DeviceTags)
}

type DataSetUpdate struct {
	Active             *bool                   `json:"-"`
	DeviceID           *string                 `json:"deviceId,omitempty"`
	DeviceModel        *string                 `json:"deviceModel,omitempty"`
	DeviceSerialNumber *string                 `json:"deviceSerialNumber,omitempty"`
	Deduplicator       *DeduplicatorDescriptor `json:"-"`
	State              *string                 `json:"state,omitempty"`
	Time               *time.Time              `json:"time,omitempty"`
	Timezone           *string                 `json:"timezone,omitempty"`
	TimezoneOffset     *int                    `json:"timezoneOffset,omitempty"`
}

func NewDataSetUpdate() *DataSetUpdate {
	return &DataSetUpdate{}
}

func (d *DataSetUpdate) HasUpdates() bool {
	return d.Active != nil || d.DeviceID != nil || d.DeviceModel != nil || d.DeviceSerialNumber != nil ||
		d.Deduplicator != nil || d.State != nil || d.Time != nil || d.Timezone != nil || d.TimezoneOffset != nil
}

func (d *DataSetUpdate) Parse(parser structure.ObjectParser) {
	d.DeviceID = parser.String("deviceId")
	d.DeviceModel = parser.String("deviceModel")
	d.DeviceSerialNumber = parser.String("deviceSerialNumber")
	d.State = parser.String("state")
	d.Time = parser.Time("time", TimeFormat)
	d.Timezone = parser.String("timezone")
	d.TimezoneOffset = parser.Int("timezoneOffset")
}

func (d *DataSetUpdate) Validate(validator structure.Validator) {
	validator.String("deviceId", d.DeviceID).NotEmpty()
	validator.String("deviceModel", d.DeviceModel).LengthGreaterThan(1)
	validator.String("deviceSerialNumber", d.DeviceSerialNumber).LengthGreaterThan(1)
	validator.String("state", d.State).OneOf(DataSetStates()...)
	validator.Time("time", d.Time).NotZero()
	validator.String("timezone", d.Timezone) // TODO: Timezone
	validator.Int("timezoneOffset", d.TimezoneOffset).InRange(-12*60, 14*60)
}

type DataSet struct {
	Active              bool                    `json:"-" bson:"_active"`
	Annotations         *BlobArray              `json:"annotations,omitempty" bson:"annotations,omitempty"`
	ArchivedDatasetID   string                  `json:"-" bson:"archivedDatasetId,omitempty"`
	ArchivedTime        string                  `json:"-" bson:"archivedTime,omitempty"`
	ByUser              *string                 `json:"byUser,omitempty" bson:"byUser,omitempty"`
	Client              *DataSetClient          `json:"client,omitempty" bson:"client,omitempty"`
	ClockDriftOffset    *int                    `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ComputerTime        *string                 `json:"computerTime,omitempty" bson:"computerTime,omitempty"`
	ConversionOffset    *int                    `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	CreatedTime         string                  `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID       string                  `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	DataSetType         *string                 `json:"dataSetType,omitempty" bson:"dataSetType,omitempty"`
	DataState           string                  `json:"-" bson:"_dataState,omitempty"` // TODO: Deprecated DataState (after data migration)
	Deduplicator        *DeduplicatorDescriptor `json:"-" bson:"_deduplicator,omitempty"`
	DeletedTime         string                  `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID       string                  `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	DeviceID            *string                 `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceManufacturers *[]string               `json:"deviceManufacturers,omitempty" bson:"deviceManufacturers,omitempty"`
	DeviceModel         *string                 `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	DeviceSerialNumber  *string                 `json:"deviceSerialNumber,omitempty" bson:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string               `json:"deviceTags,omitempty" bson:"deviceTags,omitempty"`
	DeviceTime          *string                 `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	GUID                string                  `json:"guid,omitempty" bson:"guid,omitempty"`
	ID                  string                  `json:"id,omitempty" bson:"id,omitempty"`
	ModifiedTime        string                  `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID      string                  `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	Payload             *Blob                   `json:"payload,omitempty" bson:"payload,omitempty"`
	SchemaVersion       int                     `json:"-" bson:"_schemaVersion,omitempty"`
	Source              *string                 `json:"source,omitempty" bson:"source,omitempty"`
	State               string                  `json:"-" bson:"_state,omitempty"`
	Time                *string                 `json:"time,omitempty" bson:"time,omitempty"`
	TimeProcessing      *string                 `json:"timeProcessing,omitempty" bson:"timeProcessing,omitempty"`
	Timezone            *string                 `json:"timezone,omitempty" bson:"timezone,omitempty"`
	TimezoneOffset      *int                    `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
	Type                string                  `json:"type,omitempty" bson:"type,omitempty"`
	UploadID            string                  `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID              string                  `json:"-" bson:"_userId,omitempty"`
	Version             *string                 `json:"version,omitempty" bson:"version,omitempty"`
	VersionInternal     int                     `json:"-" bson:"_version,omitempty"`
}

type DataSets []*DataSet
