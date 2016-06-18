package upload

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type Deduplicator struct {
	Name string                 `bson:"name,omitempty"`
	Data map[string]interface{} `bson:"data,omitempty"`
}

type Upload struct {
	base.Base `bson:",inline"`

	DataState    string        `json:"-" bson:"_dataState,omitempty"`
	Deduplicator *Deduplicator `json:"-" bson:"_deduplicator,omitempty"`
	UploadUserID string        `json:"byUser,omitempty" bson:"byUser,omitempty"`

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
	u.Base.Type = Type()
	u.Base.UploadID = app.NewID()

	u.DataState = "open"
	u.Deduplicator = nil
	u.UploadUserID = ""

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

	validator.ValidateStringAsTime("computerTime", u.ComputerTime, "2006-01-02T15:04:05").Exists()
	validator.ValidateStringArray("deviceManufacturers", u.DeviceManufacturers).Exists().NotEmpty()
	validator.ValidateString("deviceModel", u.DeviceModel).Exists().LengthGreaterThan(1)
	validator.ValidateString("deviceSerialNumber", u.DeviceSerialNumber).Exists().LengthGreaterThan(1)
	validator.ValidateStringArray("deviceTags", u.DeviceTags).Exists().NotEmpty().EachOneOf([]string{"insulin-pump", "cgm", "bgm"})
	validator.ValidateString("timeProcessing", u.TimeProcessing).Exists().OneOf([]string{"across-the-board-timezone", "utc-bootstrapping", "none"})
	validator.ValidateString("timezone", u.TimeZone).Exists().LengthGreaterThan(1)
	validator.ValidateString("version", u.Version).Exists().LengthGreaterThan(5)

	return nil
}

func (u *Upload) Normalize(normalizer data.Normalizer) error {
	normalizer.SetMeta(u.Meta())

	return u.Base.Normalize(normalizer)
}

func (u *Upload) SetUploadUserID(uploadUserID string) {
	u.UploadUserID = uploadUserID
}

func (u *Upload) SetDataState(dataState string) {
	u.DataState = dataState
}

func (u *Upload) SetDeduplicator(deduplicator *Deduplicator) {
	u.Deduplicator = deduplicator
}
