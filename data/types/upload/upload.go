package upload

import (
	"sort"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/device"
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
	TimeProcessingAcrossTheBoardTimezone = "across-the-board-timezone"
	TimeProcessingNone                   = "none"
	TimeProcessingUTCBootstrapping       = "utc-bootstrapping"
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
		TimeProcessingAcrossTheBoardTimezone,
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
	Timezone            *string   `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Version             *string   `json:"version,omitempty" bson:"version,omitempty"` // TODO: Deprecate in favor of Client.Version
}

func NewUpload(parser data.ObjectParser) *Upload {
	if parser.Object() == nil {
		return nil
	}

	if value := parser.ParseString("type"); value == nil {
		parser.AppendError("type", service.ErrorValueNotExists())
		return nil
	} else if *value != device.Type {
		parser.AppendError("type", service.ErrorValueStringNotOneOf(*value, []string{device.Type}))
		return nil
	}

	return Init()
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
	return &Upload{}
}

func Init() *Upload {
	upload := New()
	upload.Init()
	return upload
}

func (u *Upload) Init() {
	u.Base.Init()
	u.Type = Type

	u.ByUser = nil
	u.Client = nil
	u.ComputerTime = nil
	u.DataSetType = nil
	u.DataState = nil
	u.DeviceManufacturers = nil
	u.DeviceModel = nil
	u.DeviceSerialNumber = nil
	u.DeviceTags = nil
	u.State = nil
	u.TimeProcessing = nil
	u.Timezone = nil
	u.Version = nil
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
	u.Timezone = parser.ParseString("timezone")
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

	validator.StringArray("deviceManufacturers", u.DeviceManufacturers).Exists().NotEmpty().EachNotEmpty()
	validator.String("deviceModel", u.DeviceModel).Exists().NotEmpty()               // TODO: Some clients USED to send ""; requires DB migration
	validator.String("deviceSerialNumber", u.DeviceSerialNumber).Exists().NotEmpty() // TODO: Some clients STILL send "" via Jellyfish; requires fix & DB migration
	validator.StringArray("deviceTags", u.DeviceTags).Exists().NotEmpty().EachOneOf(DeviceTags()...)

	if validator.Origin() <= structure.OriginInternal {
		validator.String("state", u.State).OneOf(States()...)
	}

	validator.String("timeProcessing", u.TimeProcessing).Exists().OneOf(TimeProcessings()...) // TODO: Some clients USED to send ""; requires DB migration
	validator.String("timezone", u.Timezone).NotEmpty()
	validator.String("version", u.Version).LengthGreaterThanOrEqualTo(VersionLengthMinimum)
}

func (u *Upload) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(u.Meta())
	}

	u.Base.Normalize(normalizer)

	if normalizer.Origin() == structure.OriginExternal {
		if u.UploadID == nil {
			u.UploadID = pointer.String(id.New())
		}
	}

	if u.Client != nil {
		u.Client.Normalize(normalizer.WithReference("client"))
	}

	if normalizer.Origin() == structure.OriginExternal {
		if u.DataSetType == nil {
			u.DataSetType = pointer.String(DataSetTypeNormal)
		}
		SortAndDeduplicateStringArray(u.DeviceManufacturers)
		SortAndDeduplicateStringArray(u.DeviceTags)
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

func SortAndDeduplicateStringArray(strs *[]string) {
	if strs != nil {
		if length := len(*strs); length > 1 {
			sort.Strings(*strs)

			var lastIndex int
			for index := 1; index < length; index++ {
				if (*strs)[lastIndex] != (*strs)[index] {
					lastIndex++
					(*strs)[lastIndex] = (*strs)[index]
				}
			}
			*strs = (*strs)[:lastIndex+1]
		}
	}
}
