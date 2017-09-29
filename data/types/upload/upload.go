package upload

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
)

const (
	ComputerTimeFormat = "2006-01-02T15:04:05"

	DataSetTypeNormal     = "normal"
	DataSetTypeContinuous = "continuous"

	DeviceTagInsulinPump = "insulin-pump"
	DeviceTagCGM         = "cgm"
	DeviceTagBGM         = "bgm"

	TimeProcessingAcrossTheBoardTimezone = "across-the-board-timezone"
	TimeProcessingUTCBootstrapping       = "utc-bootstrapping"
	TimeProcessingNone                   = "none"
)

type Upload struct {
	types.Base `bson:",inline"`

	State     string `json:"-" bson:"_state,omitempty"`
	DataState string `json:"-" bson:"_dataState,omitempty"` // TODO: Deprecated DataState (after data migration)
	ByUser    string `json:"byUser,omitempty" bson:"byUser,omitempty"`

	Client              *Client   `json:"client,omitempty" bson:"client,omitempty"`
	ComputerTime        *string   `json:"computerTime,omitempty" bson:"computerTime,omitempty"`
	DataSetType         *string   `json:"dataSetType,omitempty" bson:"dataSetType,omitempty"`
	DeviceManufacturers *[]string `json:"deviceManufacturers,omitempty" bson:"deviceManufacturers,omitempty"`
	DeviceModel         *string   `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	DeviceSerialNumber  *string   `json:"deviceSerialNumber,omitempty" bson:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string `json:"deviceTags,omitempty" bson:"deviceTags,omitempty"`
	TimeProcessing      *string   `json:"timeProcessing,omitempty" bson:"timeProcessing,omitempty"`
	Timezone            *string   `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Version             *string   `json:"version,omitempty" bson:"version,omitempty"`
}

func Type() string {
	return "upload"
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
	u.Type = Type()
	u.UploadID = id.New()

	u.State = "open"
	u.DataState = "open" // TODO: Deprecated DataState (after data migration)
	u.ByUser = ""

	u.Client = nil
	u.ComputerTime = nil
	u.DataSetType = nil
	u.DeviceManufacturers = nil
	u.DeviceModel = nil
	u.DeviceSerialNumber = nil
	u.DeviceTags = nil
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

func (u *Upload) Validate(validator data.Validator) error {
	validator.SetMeta(u.Meta())

	if err := u.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &u.Type).EqualTo(Type())

	if u.Client != nil {
		u.Client.Validate(validator.NewChildValidator("client"))
	}
	validator.ValidateStringAsTime("computerTime", u.ComputerTime, ComputerTimeFormat)
	validator.ValidateString("dataSetType", u.DataSetType).OneOf([]string{DataSetTypeNormal, DataSetTypeContinuous})
	validator.ValidateStringArray("deviceManufacturers", u.DeviceManufacturers).Exists().NotEmpty()
	validator.ValidateString("deviceModel", u.DeviceModel).Exists().LengthGreaterThan(1)
	validator.ValidateString("deviceSerialNumber", u.DeviceSerialNumber).Exists().LengthGreaterThan(1)
	validator.ValidateStringArray("deviceTags", u.DeviceTags).Exists().NotEmpty().EachOneOf([]string{DeviceTagInsulinPump, DeviceTagCGM, DeviceTagBGM})
	validator.ValidateString("timeProcessing", u.TimeProcessing).Exists().OneOf([]string{TimeProcessingAcrossTheBoardTimezone, TimeProcessingUTCBootstrapping, TimeProcessingNone})
	validator.ValidateString("timezone", u.Timezone).LengthGreaterThan(1)        // .Exists()
	validator.ValidateString("version", u.Version).LengthGreaterThanOrEqualTo(5) // .Exists()

	return nil
}

func (u *Upload) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(u.Meta())

	if err := u.Base.Normalize(normalizer); err != nil {
		return err
	}

	if u.Client != nil {
		u.Client.Normalize(normalizer.NewChildNormalizer("client"))
	}

	if u.DataSetType == nil {
		u.DataSetType = pointer.String(DataSetTypeNormal)
	}

	return nil
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
