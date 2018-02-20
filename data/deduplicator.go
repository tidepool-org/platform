package data

import (
	"context"
	"regexp"

	"github.com/blang/semver"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

// TODO: Need to update all deduplicator descriptors in the database to have a name and version

type Deduplicator interface {
	Name() string
	Version() string

	RegisterDataset(ctx context.Context) error

	AddDatasetData(ctx context.Context, datasetData []Datum) error
	DeduplicateDataset(ctx context.Context) error

	DeleteDataset(ctx context.Context) error
}

type DeduplicatorDescriptor struct {
	Name    string `bson:"name,omitempty"`
	Version string `bson:"version,omitempty"`
	Hash    string `bson:"hash,omitempty"`
}

func NewDeduplicatorDescriptor() *DeduplicatorDescriptor {
	return &DeduplicatorDescriptor{}
}

func (d *DeduplicatorDescriptor) Validate(validator structure.Validator) {
	if d.Name != "" { // TODO: Remove once all deduplicator descriptions have a name and version
		validator.String("name", &d.Name).Exists().Using(ValidateReverseDomain)
	}
	if d.Version != "" { // TODO: Remove once all deduplicator descriptions have a name and version
		validator.String("version", &d.Version).Exists().Using(ValidateSemanticVersion)
	}
}

func (d *DeduplicatorDescriptor) Normalize(normalizer Normalizer) {}

func (d *DeduplicatorDescriptor) IsRegisteredWithAnyDeduplicator() bool {
	return d.Name != ""
}

func (d *DeduplicatorDescriptor) IsRegisteredWithNamedDeduplicator(name string) bool {
	return d.Name == name
}

func (d *DeduplicatorDescriptor) RegisterWithDeduplicator(deduplicator Deduplicator) error {
	if d.Name != "" {
		return errors.Newf("deduplicator descriptor already registered with %q", d.Name)
	}
	if d.Version != "" {
		return errors.New("deduplicator descriptor already registered with unknown deduplicator")
	}

	d.Name = deduplicator.Name()
	d.Version = deduplicator.Version()
	return nil
}

var reverseDomainExpression = regexp.MustCompile(`^[A-Za-z]{2,6}(\.([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]{0,61}[A-Za-z0-9]))+$`)

func ValidateReverseDomain(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if !reverseDomainExpression.MatchString(value) {
		errorReporter.ReportError(ErrorValueStringAsReverseDomainNotValid(value))
	}
}

func ValidateSemanticVersion(value string, errorReporter structure.ErrorReporter) {
	if value == "" {
		errorReporter.ReportError(structureValidator.ErrorValueEmpty())
	} else if _, err := semver.Parse(value); err != nil {
		errorReporter.ReportError(ErrorValueStringAsSemanticVersionNotValid(value))
	}
}

func ErrorValueStringAsReverseDomainNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as reverse domain", value)
}

func ErrorValueStringAsSemanticVersionNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as semantic version", value)
}
