package test

import (
	"time"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/origin"
	"github.com/tidepool-org/platform/structure"
)

type IdentityFieldsOutput struct {
	IdentityFields []string
	Error          error
}

type Datum struct {
	MetaInvocations                      int
	MetaOutputs                          []interface{}
	ParseInvocations                     int
	ParseInputs                          []structure.ObjectParser
	ValidateInvocations                  int
	ValidateInputs                       []structure.Validator
	NormalizeInvocations                 int
	NormalizeInputs                      []data.Normalizer
	IdentityFieldsInvocations            int
	IdentityFieldsOutputs                []IdentityFieldsOutput
	GetPayloadInvocations                int
	GetPayloadOutputs                    []*metadata.Metadata
	GetOriginInvocations                 int
	GetOriginOutputs                     []*origin.Origin
	GetTimeInvocations                   int
	GetTimeOutputs                       []*time.Time
	GetTimeZoneOffsetInvocations         int
	GetTimeZoneOffsetOutputs             []*int
	GetTypeInvocations                   int
	GetTypeOutputs                       []string
	SetTypeInvocations                   int
	SetTypeInputs                        []string
	SetUserIDInvocations                 int
	SetUserIDInputs                      []*string
	SetDataSetIDInvocations              int
	SetDataSetIDInputs                   []*string
	SetActiveInvocations                 int
	SetActiveInputs                      []bool
	SetDeviceIDInvocations               int
	SetDeviceIDInputs                    []*string
	SetCreatedTimeInvocations            int
	SetCreatedTimeInputs                 []*time.Time
	SetCreatedUserIDInvocations          int
	SetCreatedUserIDInputs               []*string
	SetModifiedTimeInvocations           int
	SetModifiedTimeInputs                []*time.Time
	SetModifiedUserIDInvocations         int
	SetModifiedUserIDInputs              []*string
	SetDeletedTimeInvocations            int
	SetDeletedTimeInputs                 []*time.Time
	SetDeletedUserIDInvocations          int
	SetDeletedUserIDInputs               []*string
	DeduplicatorDescriptorValue          *data.DeduplicatorDescriptor
	DeduplicatorDescriptorInvocations    int
	SetDeduplicatorDescriptorInvocations int
	IsActiveInvocations                  int
	IsActiveOutputs                      []bool
	SetProvenanceInvocations             int
	SetProvenanceInputs                  []*data.Provenance
	GetUploadIDInvocations               int
	GetUploadIDOutputs                   []*string
}

func NewDatum() *Datum {
	return &Datum{}
}

func (d *Datum) Meta() interface{} {
	d.MetaInvocations++

	gomega.Expect(d.MetaOutputs).ToNot(gomega.BeEmpty())

	output := d.MetaOutputs[0]
	d.MetaOutputs = d.MetaOutputs[1:]
	return output
}

func (d *Datum) Parse(parser structure.ObjectParser) {
	d.ParseInvocations++

	d.ParseInputs = append(d.ParseInputs, parser)
}

func (d *Datum) Validate(validator structure.Validator) {
	d.ValidateInvocations++

	d.ValidateInputs = append(d.ValidateInputs, validator)
}

func (d *Datum) Normalize(normalizer data.Normalizer) {
	d.NormalizeInvocations++

	d.NormalizeInputs = append(d.NormalizeInputs, normalizer)
}

func (d *Datum) IdentityFields() ([]string, error) {
	d.IdentityFieldsInvocations++

	gomega.Expect(d.IdentityFieldsOutputs).ToNot(gomega.BeEmpty())

	output := d.IdentityFieldsOutputs[0]
	d.IdentityFieldsOutputs = d.IdentityFieldsOutputs[1:]
	return output.IdentityFields, output.Error
}

func (d *Datum) GetPayload() *metadata.Metadata {
	d.GetPayloadInvocations++

	gomega.Expect(d.GetPayloadOutputs).ToNot(gomega.BeEmpty())

	output := d.GetPayloadOutputs[0]
	d.GetPayloadOutputs = d.GetPayloadOutputs[1:]
	return output
}

func (d *Datum) GetOrigin() *origin.Origin {
	d.GetOriginInvocations++

	gomega.Expect(d.GetOriginOutputs).ToNot(gomega.BeEmpty())

	output := d.GetOriginOutputs[0]
	d.GetOriginOutputs = d.GetOriginOutputs[1:]
	return output
}

func (d *Datum) GetType() string {
	d.GetTypeInvocations++

	gomega.Expect(d.GetTypeOutputs).ToNot(gomega.BeEmpty())

	output := d.GetTypeOutputs[0]
	d.GetTypeOutputs = d.GetTypeOutputs[1:]
	return output
}

func (d *Datum) IsActive() bool {
	d.IsActiveInvocations++

	gomega.Expect(d.IsActiveOutputs).ToNot(gomega.BeEmpty())

	output := d.IsActiveOutputs[0]
	d.IsActiveOutputs = d.IsActiveOutputs[1:]
	return output
}

func (d *Datum) SetType(typ string) {
	d.SetTypeInvocations++

	d.SetTypeInputs = append(d.SetTypeInputs, typ)
}

func (d *Datum) GetTime() *time.Time {
	d.GetTimeInvocations++

	gomega.Expect(d.GetTimeOutputs).ToNot(gomega.BeEmpty())

	output := d.GetTimeOutputs[0]
	d.GetTimeOutputs = d.GetTimeOutputs[1:]
	return output
}

func (d *Datum) GetTimeZoneOffset() *int {
	d.GetTimeZoneOffsetInvocations++

	gomega.Expect(d.GetTimeZoneOffsetOutputs).ToNot(gomega.BeEmpty())

	output := d.GetTimeZoneOffsetOutputs[0]
	d.GetTimeZoneOffsetOutputs = d.GetTimeZoneOffsetOutputs[1:]
	return output
}

func (d *Datum) SetUserID(userID *string) {
	d.SetUserIDInvocations++

	d.SetUserIDInputs = append(d.SetUserIDInputs, userID)
}

func (d *Datum) SetDataSetID(dataSetID *string) {
	d.SetDataSetIDInvocations++

	d.SetDataSetIDInputs = append(d.SetDataSetIDInputs, dataSetID)
}

func (d *Datum) SetActive(active bool) {
	d.SetActiveInvocations++

	d.SetActiveInputs = append(d.SetActiveInputs, active)
}

func (d *Datum) SetDeviceID(deviceID *string) {
	d.SetDeviceIDInvocations++

	d.SetDeviceIDInputs = append(d.SetDeviceIDInputs, deviceID)
}

func (d *Datum) SetCreatedTime(createdTime *time.Time) {
	d.SetCreatedTimeInvocations++

	d.SetCreatedTimeInputs = append(d.SetCreatedTimeInputs, createdTime)
}

func (d *Datum) SetCreatedUserID(createdUserID *string) {
	d.SetCreatedUserIDInvocations++

	d.SetCreatedUserIDInputs = append(d.SetCreatedUserIDInputs, createdUserID)
}

func (d *Datum) SetModifiedTime(modifiedTime *time.Time) {
	d.SetModifiedTimeInvocations++

	d.SetModifiedTimeInputs = append(d.SetModifiedTimeInputs, modifiedTime)
}

func (d *Datum) SetModifiedUserID(modifiedUserID *string) {
	d.SetModifiedUserIDInvocations++

	d.SetModifiedUserIDInputs = append(d.SetModifiedUserIDInputs, modifiedUserID)
}

func (d *Datum) SetDeletedTime(deletedTime *time.Time) {
	d.SetDeletedTimeInvocations++

	d.SetDeletedTimeInputs = append(d.SetDeletedTimeInputs, deletedTime)
}

func (d *Datum) SetDeletedUserID(deletedUserID *string) {
	d.SetDeletedUserIDInvocations++

	d.SetDeletedUserIDInputs = append(d.SetDeletedUserIDInputs, deletedUserID)
}

func (d *Datum) SetProvenance(p *data.Provenance) {
	d.SetProvenanceInvocations++
	d.SetProvenanceInputs = append(d.SetProvenanceInputs, p)
}

func (d *Datum) DeduplicatorDescriptor() *data.DeduplicatorDescriptor {
	d.DeduplicatorDescriptorInvocations++

	return d.DeduplicatorDescriptorValue
}

func (d *Datum) SetDeduplicatorDescriptor(deduplicatorDescriptor *data.DeduplicatorDescriptor) {
	d.SetDeduplicatorDescriptorInvocations++

	d.DeduplicatorDescriptorValue = deduplicatorDescriptor
}

func (d *Datum) Expectations() {
	gomega.Expect(d.MetaOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.IdentityFieldsOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetPayloadOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetOriginOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetTimeOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetTimeZoneOffsetOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.GetTypeOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.IsActiveOutputs).To(gomega.BeEmpty())
}

func (d *Datum) GetUploadID() *string {
	d.GetUploadIDInvocations++

	output := d.GetUploadIDOutputs[0]
	d.GetUploadIDOutputs = d.GetUploadIDOutputs[1:]

	return output
}
