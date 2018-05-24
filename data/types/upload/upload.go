package upload

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/structure"
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

func NewUpload(parser data.ObjectParser) *Upload {
	if parser.Object() == nil {
		return nil
	}

	if value := parser.ParseString("type"); value == nil {
		parser.AppendError("type", service.ErrorValueNotExists())
		return nil
	} else if *value != Type {
		parser.AppendError("type", service.ErrorValueStringNotOneOf(*value, []string{Type}))
		return nil
	}

	return New()
}

func ParseUpload(parser data.ObjectParser) *Upload {
	datum := NewUpload(parser)
	if datum == nil {
		return nil
	}

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

func (u *Upload) Parse(parser data.ObjectParser) error {
	parser.SetMeta(u.Meta())

	if err := u.Base.Parse(parser); err != nil {
		return err
	}

	u.Client = ParseClient(parser.NewChildObjectParser("client"))
	u.ComputerTime = parser.ParseString("computerTime")
	u.DataSetType = parser.ParseString("dataSetType")
	u.DeviceManufacturers = parser.ParseStringArray("deviceManufacturers")
	u.DeviceModel = parser.ParseString("deviceModel")
	u.DeviceSerialNumber = parser.ParseString("deviceSerialNumber")
	u.DeviceTags = parser.ParseStringArray("deviceTags")
	u.TimeProcessing = parser.ParseString("timeProcessing")
	u.Version = parser.ParseString("version")

	return nil
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

	validator.StringArray("deviceManufacturers", u.DeviceManufacturers).Exists().NotEmpty().EachNotEmpty().EachUnique()
	validator.String("deviceModel", u.DeviceModel).Exists().NotEmpty()               // TODO: Some clients USED to send ""; requires DB migration
	validator.String("deviceSerialNumber", u.DeviceSerialNumber).Exists().NotEmpty() // TODO: Some clients STILL send "" via Jellyfish; requires fix & DB migration
	validator.StringArray("deviceTags", u.DeviceTags).Exists().NotEmpty().EachOneOf(DeviceTags()...).EachUnique()

	if validator.Origin() <= structure.OriginInternal {
		validator.String("state", u.State).OneOf(States()...)
	}

	validator.String("timeProcessing", u.TimeProcessing).Exists().OneOf(TimeProcessings()...) // TODO: Some clients USED to send ""; requires DB migration
	validator.String("version", u.Version).LengthGreaterThanOrEqualTo(VersionLengthMinimum)
}

func (u *Upload) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(u.Meta())
	}

	u.Base.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		if u.UploadID == nil {
			u.UploadID = pointer.FromString(id.New())
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

func (u *Upload) HasDeviceManufacturerOneOf(deviceManufacturers []string) bool {
	if u.DeviceManufacturers == nil {
		return false
	}

	for _, uploadDeviceManufacturer := range *u.DeviceManufacturers {
		for _, deviceManufacturer := range deviceManufacturers {
			if deviceManufacturer == uploadDeviceManufacturer {
				return true
			}
		}
	}

	return false
}
