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
	GetTimeZoneOffset() *int
	GetUploadID() *string

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
	SetProvenance(*Provenance)
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

func (d Data) Filter(predicate func(Datum) bool) Data {
	filtered := Data{}
	for _, datum := range d {
		if predicate(datum) {
			filtered = append(filtered, datum)
		}
	}
	return filtered
}

// Provenance of a document.
//
// Useful for determining additional actions to take. For example, if the
// document should be sent to Kafka for asynchronous processing.
type Provenance struct {
	// ClientID of the service making the request.
	//
	// Examples: "shoreline" or "tidepool-loop"
	ClientID string `json:"clientId" bson:"clientID"`
	// ByUserID the userId of the user submitting the data.
	//
	// This is a std Tidepool user id.
	ByUserID *string `json:"byUserId,omitempty" bson:"byUserID,omitempty"`
	// SourceIP address from the HTTP request submitting the data.
	SourceIP *string `json:"sourceIP,omitempty" bson:"sourceIP,omitempty"`
}

func ParseProvenance(parser structure.ObjectParser) *Provenance {
	if !parser.Exists() {
		return nil
	}
	datum := NewProvenance()
	parser.Parse(datum)
	return datum
}

func NewProvenance() *Provenance {
	return &Provenance{}
}

func (p *Provenance) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("clientId"); ptr != nil {
		p.ClientID = *ptr
	}
	p.ByUserID = parser.String("byUserId")
	p.SourceIP = parser.String("sourceIP")
}
