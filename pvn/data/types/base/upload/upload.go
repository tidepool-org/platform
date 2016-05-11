package upload

import "github.com/tidepool-org/platform/pvn/data"
import "github.com/tidepool-org/platform/pvn/data/types/base"

type Upload struct {
	base.Base `bson:",inline"`

	UploadUserID        *string      `json:"byUser" bson:"byUser"`
	Version             *string      `json:"version" bson:"version"`
	ComputerTime        *string      `json:"computerTime" bson:"computerTime"`
	DeviceTags          *[]string    `json:"deviceTags" bson:"deviceTags"`
	DeviceManufacturers *[]string    `json:"deviceManufacturers" bson:"deviceManufacturers"`
	DeviceModel         *string      `json:"deviceModel" bson:"deviceModel"`
	DeviceSerialNumber  *string      `json:"deviceSerialNumber" bson:"deviceSerialNumber"`
	TimeProcessing      *string      `json:"timeProcessing" bson:"timeProcessing"`
	DataState           *string      `json:"dataState" bson:"dataState"`
	Deduplicator        *interface{} `json:"deduplicator" bson:"deduplicator"`
}

func Type() string {
	return "upload"
}

func New() *Upload {
	uploadType := Type()

	upload := &Upload{}
	upload.Type = &uploadType
	return upload
}

func (u *Upload) Parse(parser data.ObjectParser) {
	u.Base.Parse(parser)

	u.UploadUserID = parser.ParseString("byUser")
	u.Version = parser.ParseString("version")
	u.ComputerTime = parser.ParseString("computerTime")
	u.DeviceTags = parser.ParseStringArray("deviceTags")
	u.DeviceManufacturers = parser.ParseStringArray("deviceManufacturers")
	u.DeviceModel = parser.ParseString("deviceModel")
	u.DeviceSerialNumber = parser.ParseString("deviceSerialNumber")
	u.TimeProcessing = parser.ParseString("timeProcessing")
	u.DataState = parser.ParseString("dataState")
	u.Deduplicator = parser.ParseInterface("deduplicator")
}

func (u *Upload) Validate(validator data.Validator) {
	u.Base.Validate(validator)

	validator.ValidateString("type", u.Type).Exists()
	validator.ValidateString("byUser", u.UploadUserID).Exists().LengthGreaterThanOrEqualTo(10)
	validator.ValidateString("version", u.Version).Exists().LengthGreaterThan(5)
	validator.ValidateString("computerTime", u.ComputerTime).Exists()
	validator.ValidateStringArray("deviceTags", u.DeviceTags).Exists().LengthGreaterThanOrEqualTo(1).EachOneOf([]string{"insulin-pump", "cgm", "bgm"})
	validator.ValidateStringArray("deviceManufacturers", u.DeviceManufacturers).Exists().LengthGreaterThanOrEqualTo(1)
	validator.ValidateString("deviceModel", u.DeviceModel).Exists().LengthGreaterThan(1)
	validator.ValidateString("deviceSerialNumber", u.DeviceSerialNumber).Exists().LengthGreaterThan(1)
	validator.ValidateString("timeProcessing", u.TimeProcessing).Exists().OneOf([]string{"across-the-board-timezone", "utc-bootstrapping", "none"})
	validator.ValidateString("dataState", u.DataState).Exists()
	validator.ValidateInterface("deduplicator", u.Deduplicator).Exists()
}

func (u *Upload) Normalize(normalizer data.Normalizer) {
	u.Base.Normalize(normalizer)
}
