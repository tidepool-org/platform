package data

import (
	"time"

	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/origin"
	"github.com/tidepool-org/platform/structure"
)

type Datum interface {
	Meta() interface{}

	Parse(parser structure.ObjectParser)
	Validate(validator structure.Validator)
	Normalize(normalizer Normalizer)

	IdentityFields() ([]string, error)

	GetOrigin() *origin.Origin
	GetPayload() *metadata.Metadata

	GetType() string
	IsActive() bool
	GetTime() *time.Time

	SetUserID(userID *string)
	SetDataSetID(dataSetID *string)
	SetActive(active bool)
	SetType(typ string)
	SetDeviceID(deviceID *string)
	SetCreatedTime(createdTime *time.Time)
	SetCreatedUserID(createdUserID *string)
	SetModifiedTime(modifiedTime *time.Time)
	SetModifiedUserID(modifiedUserID *string)
	SetDeletedTime(deletedTime *time.Time)
	SetDeletedUserID(deletedUserID *string)
	DeduplicatorDescriptor() *DeduplicatorDescriptor
	SetDeduplicatorDescriptor(deduplicatorDescriptor *DeduplicatorDescriptor)
}

func DatumAsPointer(datum Datum) *Datum {
	return &datum
}

type Data []Datum

func (d Data) SetActive(active bool) {
	for _, datum := range d {
		datum.SetActive(active)
	}
}

func (d Data) SetModifiedTime(modifiedTime *time.Time) {
	for _, datum := range d {
		datum.SetModifiedTime(modifiedTime)
	}
}
