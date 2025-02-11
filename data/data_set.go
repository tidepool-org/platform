package data

import (
	"context"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	timeZone "github.com/tidepool-org/platform/time/zone"
	"github.com/tidepool-org/platform/user"
)

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

	ClockDriftOffsetMaximum = 24 * 60 * 60 * 1000  // TODO: Fix! Limit to reasonable values
	ClockDriftOffsetMinimum = -24 * 60 * 60 * 1000 // TODO: Fix! Limit to reasonable values

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

	TimeZoneOffsetMaximum  = 7 * 24 * 60  // TODO: Fix! Limit to reasonable values
	TimeZoneOffsetMinimum  = -7 * 24 * 60 // TODO: Fix! Limit to reasonable values
	VersionInternalMinimum = 0

	VersionLengthMinimum = 5
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
	Name    *string            `json:"name,omitempty" bson:"name,omitempty"`
	Version *string            `json:"version,omitempty" bson:"version,omitempty"`
	Private *metadata.Metadata `json:"private,omitempty" bson:"private,omitempty"`
}

func ParseDataSetClient(parser structure.ObjectParser) *DataSetClient {
	if !parser.Exists() {
		return nil
	}
	datum := NewDataSetClient()
	datum.Parse(parser)
	return datum
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
	if ptr := metadata.ParseMetadata(parser.WithReferenceObjectParser("private")); ptr != nil {
		d.Private = ptr
	}
}

func (d *DataSetClient) Validate(validator structure.Validator) {
	validator.String("name", d.Name).NotEmpty()
	validator.String("version", d.Version).NotEmpty() // TODO: Semver validation
	if d.Private != nil {
		d.Private.Validate(validator.WithReference("private"))
	}
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
	TimeProcessing      *string                 `json:"timeProcessing,omitempty"`
	TimeZoneName        *string                 `json:"timezone,omitempty"`
	TimeZoneOffset      *int                    `json:"timezoneOffset,omitempty"`
}

func NewDataSetCreate() *DataSetCreate {
	return &DataSetCreate{}
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
	ByUser              *string                 `json:"byUser,omitempty" bson:"byUser,omitempty"`
	Client              *DataSetClient          `json:"client,omitempty" bson:"client,omitempty"`
	ClockDriftOffset    *int                    `json:"clockDriftOffset,omitempty" bson:"clockDriftOffset,omitempty"`
	ComputerTime        *string                 `json:"computerTime,omitempty" bson:"computerTime,omitempty"` // TODO: Do we really need this? CreatedTime should suffice.
	ConversionOffset    *int                    `json:"conversionOffset,omitempty" bson:"conversionOffset,omitempty"`
	CreatedTime         *time.Time              `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	CreatedUserID       *string                 `json:"createdUserId,omitempty" bson:"createdUserId,omitempty"`
	DataSetType         *string                 `json:"dataSetType,omitempty" bson:"dataSetType,omitempty"` // TODO: Migrate to "type" after migration to DataSet (not based on Base)
	DataState           *string                 `json:"-" bson:"_dataState,omitempty"`                      // TODO: Deprecated DataState (after data migration)
	Deduplicator        *DeduplicatorDescriptor `json:"deduplicator,omitempty" bson:"_deduplicator,omitempty"`
	DeletedTime         *time.Time              `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	DeletedUserID       *string                 `json:"deletedUserId,omitempty" bson:"deletedUserId,omitempty"`
	DeviceID            *string                 `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceManufacturers *[]string               `json:"deviceManufacturers,omitempty" bson:"deviceManufacturers,omitempty"`
	DeviceModel         *string                 `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	DeviceSerialNumber  *string                 `json:"deviceSerialNumber,omitempty" bson:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string               `json:"deviceTags,omitempty" bson:"deviceTags,omitempty"`
	DeviceTime          *string                 `json:"deviceTime,omitempty" bson:"deviceTime,omitempty"`
	ID                  *string                 `json:"id,omitempty" bson:"id,omitempty"`
	ModifiedTime        *time.Time              `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	ModifiedUserID      *string                 `json:"modifiedUserId,omitempty" bson:"modifiedUserId,omitempty"`
	Payload             *metadata.Metadata      `json:"payload,omitempty" bson:"payload,omitempty"`
	Provenance          *Provenance             `json:"-" bson:"provenance,omitempty"`
	State               *string                 `json:"-" bson:"_state,omitempty"` // TODO: Should this be returned in JSON? I think so.
	Time                *time.Time              `json:"time,omitempty" bson:"time,omitempty"`
	TimeProcessing      *string                 `json:"timeProcessing,omitempty" bson:"timeProcessing,omitempty"`
	TimeZoneName        *string                 `json:"timezone,omitempty" bson:"timezone,omitempty"`
	TimeZoneOffset      *int                    `json:"timezoneOffset,omitempty" bson:"timezoneOffset,omitempty"`
	Type                string                  `json:"type,omitempty" bson:"type,omitempty"`
	UploadID            *string                 `json:"uploadId,omitempty" bson:"uploadId,omitempty"`
	UserID              *string                 `json:"-" bson:"_userId,omitempty"`
	Version             *string                 `json:"version,omitempty" bson:"version,omitempty"` // TODO: Deprecate in favor of Client.Version
	VersionInternal     int                     `json:"-" bson:"_version,omitempty"`
}

func ParseDataSet(parser structure.ObjectParser) *DataSet {
	if !parser.Exists() {
		return nil
	}

	datum := NewDataSet()
	datum.Parse(parser)
	return datum
}

func NewDataSet() *DataSet {
	return &DataSet{
		Type: "upload",
	}
}

func (d *DataSet) Parse(parser structure.ObjectParser) {
	d.Annotations = metadata.ParseMetadataArray(parser.WithReferenceArrayParser("annotations"))
	d.ByUser = parser.String("byUser")
	d.Client = ParseDataSetClient(parser.WithReferenceObjectParser("client"))
	d.ClockDriftOffset = parser.Int("clockDriftOffset")
	d.ComputerTime = parser.String("computerTime")
	d.ConversionOffset = parser.Int("conversionOffset")
	d.CreatedTime = parser.Time("createdTime", time.RFC3339Nano)
	d.CreatedUserID = parser.String("createdUserId")
	d.DataSetType = parser.String("dataSetType")
	d.Deduplicator = ParseDeduplicatorDescriptor(parser.WithReferenceObjectParser("deduplicator"))
	d.DeletedTime = parser.Time("deletedTime", time.RFC3339Nano)
	d.DeletedUserID = parser.String("deletedUserId")
	d.DeviceID = parser.String("deviceId")
	d.DeviceManufacturers = parser.StringArray("deviceManufacturers")
	d.DeviceModel = parser.String("deviceModel")
	d.DeviceSerialNumber = parser.String("deviceSerialNumber")
	d.DeviceTags = parser.StringArray("deviceTags")
	d.DeviceTime = parser.String("deviceTime")
	d.ID = parser.String("id")
	d.ModifiedTime = parser.Time("modifiedTime", time.RFC3339Nano)
	d.ModifiedUserID = parser.String("modifiedUserId")
	d.Payload = metadata.ParseMetadata(parser.WithReferenceObjectParser("payload"))
	d.Time = parser.Time("time", TimeFormat)
	d.TimeProcessing = parser.String("timeProcessing")
	d.TimeZoneName = parser.String("timezone")
	d.TimeZoneOffset = parser.Int("timezoneOffset")
	_ = parser.String("type")
	d.UploadID = parser.String("uploadId")
	d.Version = parser.String("version")
}

func (d *DataSet) Validate(validator structure.Validator) {
	// NOTE we copy these to default them if nil without writing to the originals
	// the logic below does not like null pointers
	var createdTime time.Time
	var modifiedTime time.Time

	if d.CreatedTime != nil {
		createdTime = *d.CreatedTime
	}
	if d.ModifiedTime != nil {
		modifiedTime = *d.ModifiedTime
	}

	if d.Annotations != nil {
		d.Annotations.Validate(validator.WithReference("annotations"))
	}

	if d.Client != nil {
		d.Client.Validate(validator.WithReference("client"))
	}

	validator.Int("clockDriftOffset", d.ClockDriftOffset).InRange(ClockDriftOffsetMinimum, ClockDriftOffsetMaximum)

	if validator.Origin() <= structure.OriginInternal {
		if d.CreatedTime != nil {
			validator.Time("createdTime", d.CreatedTime).BeforeNow(time.Second)
			validator.String("createdUserId", d.CreatedUserID).Using(user.IDValidator)
		} else {
			validator.Time("createdTime", d.CreatedTime).Exists()
			validator.String("createdUserId", d.CreatedUserID).NotExists()
		}

		if d.DeletedTime != nil {
			validator.Time("deletedTime", d.DeletedTime).After(latestTime(createdTime, modifiedTime)).BeforeNow(time.Second)
			validator.String("deletedUserId", d.DeletedUserID).Using(user.IDValidator)
		} else {
			validator.String("deletedUserId", d.DeletedUserID).NotExists()
		}
	}

	validator.String("computerTime", d.ComputerTime).AsTime(ComputerTimeFormat)
	validator.String("dataSetType", d.DataSetType).OneOf(DataSetTypes()...) // TODO: New field; add .Exists(); requires fix & DB migration

	if validator.Origin() <= structure.OriginInternal {
		validator.String("dataState", d.DataState).OneOf(DataSetStates()...)
	}

	if d.Deduplicator != nil {
		d.Deduplicator.Validate(validator.WithReference("deduplicator"))
	}

	validator.String("deviceId", d.DeviceID).NotEmpty()
	validator.StringArray("deviceManufacturers", d.DeviceManufacturers).NotEmpty().EachNotEmpty().EachUnique()
	validator.String("deviceModel", d.DeviceModel).NotEmpty()               // TODO: Some clients USED to send ""; requires DB migration
	validator.String("deviceSerialNumber", d.DeviceSerialNumber).NotEmpty() // TODO: Some clients STILL send "" via Jellyfish; requires fix & DB migration
	validator.StringArray("deviceTags", d.DeviceTags).NotEmpty().EachOneOf(DeviceTags()...).EachUnique()
	validator.String("deviceTime", d.DeviceTime).AsTime(DeviceTimeFormat)

	validator.String("id", d.ID).Using(IDValidator)
	if validator.Origin() <= structure.OriginInternal {
		validator.String("id", d.ID).Exists()
	}

	if validator.Origin() <= structure.OriginInternal {
		if d.ModifiedTime != nil {
			validator.Time("modifiedTime", d.ModifiedTime).After(createdTime).BeforeNow(time.Second)
			validator.String("modifiedUserId", d.ModifiedUserID).Using(user.IDValidator)
		} else {
			validator.String("modifiedUserId", d.ModifiedUserID).NotExists()
		}
	}

	if d.Payload != nil {
		d.Payload.Validate(validator.WithReference("payload"))
	}

	if validator.Origin() <= structure.OriginInternal {
		validator.String("state", d.State).OneOf(DataSetStates()...)
	}

	validator.String("timeProcessing", d.TimeProcessing).OneOf(TimeProcessings()...) // TODO: Some clients USED to send ""; requires DB migration
	validator.String("timezone", d.TimeZoneName).Using(timeZone.NameValidator)
	validator.Int("timezoneOffset", d.TimeZoneOffset).InRange(TimeZoneOffsetMinimum, TimeZoneOffsetMaximum)

	if validator.Origin() <= structure.OriginInternal {
		validator.String("uploadId", d.UploadID).Exists().Using(SetIDValidator)
	}
	if validator.Origin() <= structure.OriginStore {
		validator.String("_userId", d.UserID).Exists().Using(user.IDValidator)
		validator.Int("_version", &d.VersionInternal).Exists().GreaterThanOrEqualTo(VersionInternalMinimum)
	}

	validator.String("version", d.Version).LengthGreaterThanOrEqualTo(VersionLengthMinimum)
}

func (d *DataSet) Normalize(normalizer structure.Normalizer) {
	if normalizer.Origin() == structure.OriginExternal {
		if d.DataSetType == nil {
			d.DataSetType = pointer.FromString(DataSetTypeNormal)
		}
		if d.DeviceManufacturers != nil {
			sort.Strings(*d.DeviceManufacturers)
		}
		if d.DeviceTags != nil {
			sort.Strings(*d.DeviceTags)
		}
		if d.ID == nil {
			d.ID = pointer.FromString(NewID())
		}
		if d.UploadID == nil {
			d.UploadID = d.ID
		}
	}
}

func (d *DataSet) HasDataSetTypeContinuous() bool {
	return d.DataSetType != nil && *d.DataSetType == DataSetTypeContinuous
}

func (d *DataSet) HasDataSetTypeNormal() bool {
	return d.DataSetType == nil || *d.DataSetType == DataSetTypeNormal
}

func (d *DataSet) HasDeduplicatorName() bool {
	return d.Deduplicator != nil && d.Deduplicator.HasName()
}

func (d *DataSet) HasDeduplicatorNameMatch(name string) bool {
	return d.Deduplicator != nil && d.Deduplicator.HasNameMatch(name)
}

func (d *DataSet) IsOpen() bool {
	return d.State == nil || *d.State == DataSetStateOpen
}

type DataSets []*DataSet

func latestTime(tms ...time.Time) time.Time {
	var latestTime time.Time
	for _, tm := range tms {
		if tm.After(latestTime) {
			latestTime = tm
		}
	}
	return latestTime
}
