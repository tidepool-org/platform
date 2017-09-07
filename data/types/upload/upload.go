package upload

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
)

type Upload struct {
	types.Base `bson:",inline"`

	State     string `json:"-" bson:"_state,omitempty"`
	DataState string `json:"-" bson:"_dataState,omitempty"` // TODO: Deprecated DataState (after data migration)
	ByUser    string `json:"byUser,omitempty" bson:"byUser,omitempty"`

	ComputerTime        *string   `json:"computerTime,omitempty" bson:"computerTime,omitempty"`
	DeviceManufacturers *[]string `json:"deviceManufacturers,omitempty" bson:"deviceManufacturers,omitempty"`
	DeviceModel         *string   `json:"deviceModel,omitempty" bson:"deviceModel,omitempty"`
	DeviceSerialNumber  *string   `json:"deviceSerialNumber,omitempty" bson:"deviceSerialNumber,omitempty"`
	DeviceTags          *[]string `json:"deviceTags,omitempty" bson:"deviceTags,omitempty"`
	TimeProcessing      *string   `json:"timeProcessing,omitempty" bson:"timeProcessing,omitempty"`
	TimeZone            *string   `json:"timezone,omitempty" bson:"timezone,omitempty"`
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
	u.UploadID = app.NewID()

	u.State = "open"
	u.DataState = "open" // TODO: Deprecated DataState (after data migration)
	u.ByUser = ""

	u.ComputerTime = nil
	u.DeviceManufacturers = nil
	u.DeviceModel = nil
	u.DeviceSerialNumber = nil
	u.DeviceTags = nil
	u.TimeProcessing = nil
	u.TimeZone = nil
	u.Version = nil
}

func (u *Upload) Parse(parser data.ObjectParser) error {
	parser.SetMeta(u.Meta())

	if err := u.Base.Parse(parser); err != nil {
		return err
	}

	u.ComputerTime = parser.ParseString("computerTime")
	u.DeviceManufacturers = parser.ParseStringArray("deviceManufacturers")
	u.DeviceModel = parser.ParseString("deviceModel")
	u.DeviceSerialNumber = parser.ParseString("deviceSerialNumber")
	u.DeviceTags = parser.ParseStringArray("deviceTags")
	u.TimeProcessing = parser.ParseString("timeProcessing")
	u.TimeZone = parser.ParseString("timezone")
	u.Version = parser.ParseString("version")

	return nil
}

func (u *Upload) Validate(validator data.Validator) error {
	validator.SetMeta(u.Meta())

	if err := u.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("type", &u.Type).EqualTo(Type())

	validator.ValidateStringAsTime("computerTime", u.ComputerTime, "2006-01-02T15:04:05").Exists()
	validator.ValidateStringArray("deviceManufacturers", u.DeviceManufacturers).Exists().NotEmpty()
	validator.ValidateString("deviceModel", u.DeviceModel).Exists().LengthGreaterThan(1)
	validator.ValidateString("deviceSerialNumber", u.DeviceSerialNumber).Exists().LengthGreaterThan(1)
	validator.ValidateStringArray("deviceTags", u.DeviceTags).Exists().NotEmpty().EachOneOf([]string{"insulin-pump", "cgm", "bgm"})
	validator.ValidateString("timeProcessing", u.TimeProcessing).Exists().OneOf([]string{"across-the-board-timezone", "utc-bootstrapping", "none"})
	validator.ValidateString("timezone", u.TimeZone).Exists().LengthGreaterThan(1)
	validator.ValidateString("version", u.Version).Exists().LengthGreaterThanOrEqualTo(5)

	return nil
}

func (u *Upload) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(u.Meta())

	return u.Base.Normalize(normalizer)
}
