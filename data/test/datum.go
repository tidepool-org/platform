package test

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/id"
)

type IdentityFieldsOutput struct {
	IdentityFields []string
	Error          error
}

type Datum struct {
	ID                                   string
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
		ID: id.New(),
	}
}

func (d *Datum) Init() {
	d.InitInvocations++
}

func (d *Datum) Meta() interface{} {
	d.MetaInvocations++

	if len(d.MetaOutputs) == 0 {
		panic("Unexpected invocation of Meta on Datum")
	}

	output := d.MetaOutputs[0]
	d.MetaOutputs = d.MetaOutputs[1:]
	return output
}

func (d *Datum) Parse(parser data.ObjectParser) error {
	d.ParseInvocations++

	d.ParseInputs = append(d.ParseInputs, parser)

	if len(d.ParseOutputs) == 0 {
		panic("Unexpected invocation of Parse on Datum")
	}

	output := d.ParseOutputs[0]
	d.ParseOutputs = d.ParseOutputs[1:]
	return output
}

func (d *Datum) Validate(validator data.Validator) error {
	d.ValidateInvocations++

	d.ValidateInputs = append(d.ValidateInputs, validator)

	if len(d.ValidateOutputs) == 0 {
		panic("Unexpected invocation of Validate on Datum")
	}

	output := d.ValidateOutputs[0]
	d.ValidateOutputs = d.ValidateOutputs[1:]
	return output
}

func (d *Datum) Normalize(normalizer data.Normalizer) error {
	d.NormalizeInvocations++

	d.NormalizeInputs = append(d.NormalizeInputs, normalizer)

	if len(d.NormalizeOutputs) == 0 {
		panic("Unexpected invocation of Normalize on Datum")
	}

	output := d.NormalizeOutputs[0]
	d.NormalizeOutputs = d.NormalizeOutputs[1:]
	return output
}

func (d *Datum) IdentityFields() ([]string, error) {
	d.IdentityFieldsInvocations++

	if len(d.IdentityFieldsOutputs) == 0 {
		panic("Unexpected invocation of IdentityFields on Datum")
	}

	output := d.IdentityFieldsOutputs[0]
	d.IdentityFieldsOutputs = d.IdentityFieldsOutputs[1:]
	return output.IdentityFields, output.Error
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

func (d *Datum) UnusedOutputsCount() int {
	return len(d.MetaOutputs) +
		len(d.ParseOutputs) +
		len(d.ValidateOutputs) +
		len(d.NormalizeOutputs) +
		len(d.IdentityFieldsOutputs)
}
