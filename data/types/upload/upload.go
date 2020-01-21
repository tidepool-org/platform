package upload

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	Type = "upload"

	ComputerTimeFormat                   = "2006-01-02T15:04:05"
	DataSetTypeContinuous                = "continuous"
	DataSetTypeNormal                    = "normal"
	DeviceTagBGM                         = "bgm"
	DeviceTagCGM                         = "cgm"
	DeviceTagInsulinPump                 = "insulin-pump"
	StateClosed                          = "closed"
	StateOpen                            = "open"
	TimeProcessingAcrossTheBoardTimeZone = "across-the-board-timezone" // TODO: Rename to across-the-board-time-zone or alternative
	TimeProcessingNone                   = "none"
	TimeProcessingUTCBootstrapping       = "utc-bootstrapping" // TODO: Rename to utc-bootstrap or alternative
	VersionLengthMinimum                 = 5
)

func DataSetTypes() []string {
	return []string{
		DataSetTypeContinuous,
		DataSetTypeNormal,
	}
}

func DeviceTags() []string {
	return []string{
		DeviceTagBGM,
		DeviceTagCGM,
		DeviceTagInsulinPump,
	}
}

func States() []string {
	return []string{
		StateClosed,
		StateOpen,
	}
}

func TimeProcessings() []string {
	return []string{
		TimeProcessingAcrossTheBoardTimeZone,
		TimeProcessingNone,
		TimeProcessingUTCBootstrapping,
	}
}

// TODO: Upload does not use at least the following fields from Base: annotations, clockDriftOffset, deviceTime, payload, others?
// TODO: Upload (DataSet) should be separate from Base and eliminate all unnecessary fields

type Upload struct {
	types.Base `bson:",inline"`

	ByUser              *string   `json:"byUser,omitempty" bson:"byUser,omitempty"` // TODO: Deprecate in favor of CreatedUserID
	Client              *Client   `json:"client,omitempty" bson:"client,omitempty"`
	ComputerTime        *string   `json:"computerTime,omitempty" bson:"computerTime,omitempty"` // TODO: Do we really need this? CreatedTime should suffice.
	DataSetType         *string   `json:"dataSetType,omitempty" bson:"dataSetType,omitempty"`   // TODO: Migrate to "type" after migration to DataSet (not based on Base)
	DataState           *string   `json:"-" bson:"_dataState,omitempty"`                        // TODO: Deprecated! (remove after data migration)
	DeviceManufacturers *[]string `json:"deviceManufacturers,omitempty" bson:"deviceManufacturers,omitempty"`
	DeviceModel         *string   `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	DeviceSerialNumber  *string   `json:"deviceSerialNumber,omitempty" bson:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string `json:"deviceTags,omitempty" bson:"deviceTags,omitempty"`
	State               *string   `json:"-" bson:"_state,omitempty"` // TODO: Should this be returned in JSON? I think so.
	TimeProcessing      *string   `json:"timeProcessing,omitempty" bson:"timeProcessing,omitempty"`
	Version             *string   `json:"version,omitempty" bson:"version,omitempty"` // TODO: Deprecate in favor of Client.Version
}

func NewUpload(parser structure.ObjectParser) *Upload {
	if !parser.Exists() {
		return nil
	}

	if value := parser.String("type"); value == nil {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueNotExists())
		return nil
	} else if *value != Type {
		parser.WithReferenceErrorReporter("type").ReportError(structureValidator.ErrorValueStringNotOneOf(*value, []string{Type}))
		return nil
	}

	return New()
}

func ParseUpload(parser structure.ObjectParser) *Upload {
	if !parser.Exists() {
		return nil
	}

	_ = parser.String("type")

	datum := New()
	datum.Parse(parser)
	return datum
}

func NewDatum() data.Datum {
	return New()
}

func New() *Upload {
	return &Upload{
		Base: types.New(Type),
	}
}

func (u *Upload) Parse(parser structure.ObjectParser) {
	if !parser.HasMeta() {
		parser = parser.WithMeta(u.Meta())
	}

	u.Base.Parse(parser)

	u.Deduplicator = data.ParseDeduplicatorDescriptor(parser.WithReferenceObjectParser("deduplicator"))

	u.Client = ParseClient(parser.WithReferenceObjectParser("client"))
	u.ComputerTime = parser.String("computerTime")
	u.DataSetType = parser.String("dataSetType")
	u.DeviceManufacturers = parser.StringArray("deviceManufacturers")
	u.DeviceModel = parser.String("deviceModel")
	u.DeviceSerialNumber = parser.String("deviceSerialNumber")
	u.DeviceTags = parser.StringArray("deviceTags")
	u.TimeProcessing = parser.String("timeProcessing")
	u.Version = parser.String("version")
}

func (u *Upload) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(u.Meta())
	}

	u.Base.Validate(validator)

	if u.Type != "" {
		validator.String("type", &u.Type).EqualTo(Type)
	}

	if u.Client != nil {
		u.Client.Validate(validator.WithReference("client"))
	}

	validator.String("computerTime", u.ComputerTime).AsTime(ComputerTimeFormat)
	validator.String("dataSetType", u.DataSetType).OneOf(DataSetTypes()...) // TODO: New field; add .Exists(); requires fix & DB migration

	if validator.Origin() <= structure.OriginInternal {
		validator.String("dataState", u.DataState).OneOf(States()...)
	}

	validator.StringArray("deviceManufacturers", u.DeviceManufacturers).NotEmpty().EachNotEmpty().EachUnique()
	validator.String("deviceModel", u.DeviceModel).NotEmpty()               // TODO: Some clients USED to send ""; requires DB migration
	validator.String("deviceSerialNumber", u.DeviceSerialNumber).NotEmpty() // TODO: Some clients STILL send "" via Jellyfish; requires fix & DB migration
	validator.StringArray("deviceTags", u.DeviceTags).NotEmpty().EachOneOf(DeviceTags()...).EachUnique()

	if validator.Origin() <= structure.OriginInternal {
		validator.String("state", u.State).OneOf(States()...)
	}

	validator.String("timeProcessing", u.TimeProcessing).OneOf(TimeProcessings()...) // TODO: Some clients USED to send ""; requires DB migration
	validator.String("version", u.Version).LengthGreaterThanOrEqualTo(VersionLengthMinimum)
}

// IsValid returns true if there is no error in the validator
func (u *Upload) IsValid(validator structure.Validator) bool {
	return !(validator.HasError())
}

func (u *Upload) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(u.Meta())
	}

	u.Base.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		if u.UploadID == nil {
			u.UploadID = u.ID
		}
	}

	if u.Client != nil {
		u.Client.Normalize(normalizer.WithReference("client"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if u.DataSetType == nil {
			u.DataSetType = pointer.FromString(DataSetTypeNormal)
		}
		if u.DeviceManufacturers != nil {
			sort.Strings(*u.DeviceManufacturers)
		}
		if u.DeviceTags != nil {
			sort.Strings(*u.DeviceTags)
		}
	}
}

func (u *Upload) HasDataSetTypeContinuous() bool {
	return u.DataSetType != nil && *u.DataSetType == DataSetTypeContinuous
}

func (u *Upload) HasDataSetTypeNormal() bool {
	return u.DataSetType == nil || *u.DataSetType == DataSetTypeNormal
}

func (u *Upload) HasDeduplicatorName() bool {
	return u.Deduplicator != nil && u.Deduplicator.HasName()
}

func (u *Upload) HasDeduplicatorNameMatch(name string) bool {
	return u.Deduplicator != nil && u.Deduplicator.HasNameMatch(name)
}
