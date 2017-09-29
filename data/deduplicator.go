package data

import (
	"context"

	"github.com/tidepool-org/platform/errors"
)

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
