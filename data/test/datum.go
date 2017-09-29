package test

import (
	"github.com/onsi/gomega"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/test"
)

type IdentityFieldsOutput struct {
	IdentityFields []string
	Error          error
}

type Datum struct {
	*test.Mock
	InitInvocations                      int
	MetaInvocations                      int
	MetaOutputs                          []interface{}
	ParseInvocations                     int
	ParseInputs                          []data.ObjectParser
	ParseOutputs                         []error
	ValidateInvocations                  int
	ValidateInputs                       []data.Validator
	ValidateOutputs                      []error
	NormalizeInvocations                 int
	NormalizeInputs                      []data.Normalizer
	NormalizeOutputs                     []error
	IdentityFieldsInvocations            int
	IdentityFieldsOutputs                []IdentityFieldsOutput
	PayloadInvocations                   int
	PayloadOutputs                       []*map[string]interface{}
	SetUserIDInvocations                 int
	SetUserIDInputs                      []string
	SetDatasetIDInvocations              int
	SetDatasetIDInputs                   []string
	SetActiveInvocations                 int
	SetActiveInputs                      []bool
	SetCreatedTimeInvocations            int
	SetCreatedTimeInputs                 []string
	SetCreatedUserIDInvocations          int
	SetCreatedUserIDInputs               []string
	SetModifiedTimeInvocations           int
	SetModifiedTimeInputs                []string
	SetModifiedUserIDInvocations         int
	SetModifiedUserIDInputs              []string
	SetDeletedTimeInvocations            int
	SetDeletedTimeInputs                 []string
	SetDeletedUserIDInvocations          int
	SetDeletedUserIDInputs               []string
	DeduplicatorDescriptorValue          *data.DeduplicatorDescriptor
	DeduplicatorDescriptorInvocations    int
	SetDeduplicatorDescriptorInvocations int
}

func NewDatum() *Datum {
	return &Datum{
		Mock: test.NewMock(),
	}
}

func (d *Datum) Init() {
	d.InitInvocations++
}

func (d *Datum) Meta() interface{} {
	d.MetaInvocations++

	gomega.Expect(d.MetaOutputs).ToNot(gomega.BeEmpty())

	output := d.MetaOutputs[0]
	d.MetaOutputs = d.MetaOutputs[1:]
	return output
}

func (d *Datum) Parse(parser data.ObjectParser) error {
	d.ParseInvocations++

	d.ParseInputs = append(d.ParseInputs, parser)

	gomega.Expect(d.ParseOutputs).ToNot(gomega.BeEmpty())

	output := d.ParseOutputs[0]
	d.ParseOutputs = d.ParseOutputs[1:]
	return output
}

func (d *Datum) Validate(validator data.Validator) error {
	d.ValidateInvocations++

	d.ValidateInputs = append(d.ValidateInputs, validator)

	gomega.Expect(d.ValidateOutputs).ToNot(gomega.BeEmpty())

	output := d.ValidateOutputs[0]
	d.ValidateOutputs = d.ValidateOutputs[1:]
	return output
}

func (d *Datum) Normalize(normalizer data.Normalizer) error {
	d.NormalizeInvocations++

	d.NormalizeInputs = append(d.NormalizeInputs, normalizer)

	gomega.Expect(d.NormalizeOutputs).ToNot(gomega.BeEmpty())

	output := d.NormalizeOutputs[0]
	d.NormalizeOutputs = d.NormalizeOutputs[1:]
	return output
}

func (d *Datum) IdentityFields() ([]string, error) {
	d.IdentityFieldsInvocations++

	gomega.Expect(d.IdentityFieldsOutputs).ToNot(gomega.BeEmpty())

	output := d.IdentityFieldsOutputs[0]
	d.IdentityFieldsOutputs = d.IdentityFieldsOutputs[1:]
	return output.IdentityFields, output.Error
}

func (d *Datum) GetPayload() *map[string]interface{} {
	d.PayloadInvocations++

	gomega.Expect(d.PayloadOutputs).ToNot(gomega.BeEmpty())

	output := d.PayloadOutputs[0]
	d.PayloadOutputs = d.PayloadOutputs[1:]
	return output
}

func (d *Datum) SetUserID(userID string) {
	d.SetUserIDInvocations++

	d.SetUserIDInputs = append(d.SetUserIDInputs, userID)
}

func (d *Datum) SetDatasetID(datasetID string) {
	d.SetDatasetIDInvocations++

	d.SetDatasetIDInputs = append(d.SetDatasetIDInputs, datasetID)
}

func (d *Datum) SetActive(active bool) {
	d.SetActiveInvocations++

	d.SetActiveInputs = append(d.SetActiveInputs, active)
}

func (d *Datum) SetCreatedTime(createdTime string) {
	d.SetCreatedTimeInvocations++

	d.SetCreatedTimeInputs = append(d.SetCreatedTimeInputs, createdTime)
}

func (d *Datum) SetCreatedUserID(createdUserID string) {
	d.SetCreatedUserIDInvocations++

	d.SetCreatedUserIDInputs = append(d.SetCreatedUserIDInputs, createdUserID)
}

func (d *Datum) SetModifiedTime(modifiedTime string) {
	d.SetModifiedTimeInvocations++

	d.SetModifiedTimeInputs = append(d.SetModifiedTimeInputs, modifiedTime)
}

func (d *Datum) SetModifiedUserID(modifiedUserID string) {
	d.SetModifiedUserIDInvocations++

	d.SetModifiedUserIDInputs = append(d.SetModifiedUserIDInputs, modifiedUserID)
}

func (d *Datum) SetDeletedTime(deletedTime string) {
	d.SetDeletedTimeInvocations++

	d.SetDeletedTimeInputs = append(d.SetDeletedTimeInputs, deletedTime)
}

func (d *Datum) SetDeletedUserID(deletedUserID string) {
	d.SetDeletedUserIDInvocations++

	d.SetDeletedUserIDInputs = append(d.SetDeletedUserIDInputs, deletedUserID)
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
	d.Mock.Expectations()
	gomega.Expect(d.MetaOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ParseOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.ValidateOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.NormalizeOutputs).To(gomega.BeEmpty())
	gomega.Expect(d.IdentityFieldsOutputs).To(gomega.BeEmpty())
}
