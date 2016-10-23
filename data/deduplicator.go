package data

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

type Deduplicator interface {
	InitializeDataset() error
	AddDataToDataset(datasetData []Datum) error
	FinalizeDataset() error
}

type DeduplicatorDescriptor struct {
	Name string `bson:"name,omitempty"`
	Hash string `bson:"hash,omitempty"`
}
