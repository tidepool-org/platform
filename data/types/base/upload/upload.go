package upload

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base"
)

type Upload struct {
	base.Base `bson:",inline"`

	DataState    string      `json:"-" bson:"_dataState,omitempty"`
	Deduplicator interface{} `json:"-" bson:"_deduplicator,omitempty"`
	UploadUserID string      `json:"byUser,omitempty" bson:"byUser,omitempty"`

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

func New() (*Upload, error) {
	uploadBase, err := base.New(Type())
	if err != nil {
		return nil, err
	}

	uploadBase.UploadID = app.NewID()

	return &Upload{
		Base:      *uploadBase,
		DataState: "open",
	}, nil
}

func (u *Upload) Parse(parser data.ObjectParser) error {
	parser.SetMeta(u.Meta())

	if err := u.Base.Parse(parser); err != nil {
		return err
	}

	u.Version = parser.ParseString("version")
	u.ComputerTime = parser.ParseString("computerTime")
	u.DeviceTags = parser.ParseStringArray("deviceTags")
	u.DeviceManufacturers = parser.ParseStringArray("deviceManufacturers")
	u.DeviceModel = parser.ParseString("deviceModel")
	u.DeviceSerialNumber = parser.ParseString("deviceSerialNumber")
	u.TimeProcessing = parser.ParseString("timeProcessing")
	u.TimeZone = parser.ParseString("timezone")

	return nil
}

func (u *Upload) Validate(validator data.Validator) error {
	validator.SetMeta(u.Meta())

	if err := u.Base.Validate(validator); err != nil {
		return err
	}

	validator.ValidateString("version", u.Version).Exists().LengthGreaterThan(5)
	validator.ValidateStringAsTime("computerTime", u.ComputerTime, "2006-01-02T15:04:05").Exists()
	validator.ValidateStringArray("deviceTags", u.DeviceTags).Exists().NotEmpty().EachOneOf([]string{"insulin-pump", "cgm", "bgm"})
	validator.ValidateStringArray("deviceManufacturers", u.DeviceManufacturers).Exists().NotEmpty()
	validator.ValidateString("deviceModel", u.DeviceModel).Exists().LengthGreaterThan(1)
	validator.ValidateString("deviceSerialNumber", u.DeviceSerialNumber).Exists().LengthGreaterThan(1)
	validator.ValidateString("timeProcessing", u.TimeProcessing).Exists().OneOf([]string{"across-the-board-timezone", "utc-bootstrapping", "none"})
	validator.ValidateString("timezone", u.TimeZone).Exists().LengthGreaterThan(1)

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

func (u *Upload) SetDeduplicator(deduplicator interface{}) {
	u.Deduplicator = deduplicator
}
