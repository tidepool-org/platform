package deduplicator

import (
	"context"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/revision"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	DataSetOriginReviseName    = "org.tidepool.deduplicator.dataset.origin.revise"
	DataSetOriginReviseVersion = "1.0.0"

	// RevisionsMaximum caps a datum's revisions, the creation entry included. A basal is extended once, so
	// reaching the cap means a bug.
	RevisionsMaximum = 50
)

const (
	ErrorCodeRevisionsExceeded = "revisions-exceeded"
	ErrorCodeRevisionsCorrupt  = "revisions-corrupt"
	ErrorCodeRevisionConflict  = "revision-conflict"
)

// ErrorRevisionsExceeded is returned when a change would take a datum past RevisionsMaximum.
func ErrorRevisionsExceeded(originID string, count int) error {
	return errors.Preparedf(ErrorCodeRevisionsExceeded, "revisions exceeded", "origin id %q has %d revisions", originID, count)
}

// ErrorRevisionsCorrupt is returned when a datum's revisions are not what the deduplicator's writes can produce.
func ErrorRevisionsCorrupt(originID string, format string, a ...any) error {
	return errors.Preparedf(ErrorCodeRevisionsCorrupt, "revisions corrupt", "origin id %q "+format, append([]any{originID}, a...)...)
}

// ErrorRevisionConflict is returned when a datum changed between being read and being written.
func ErrorRevisionConflict(originID string, version int) error {
	return errors.Preparedf(ErrorCodeRevisionConflict, "revision conflict", "origin id %q at version %d", originID, version)
}

func IsErrorRevisionsExceeded(err error) bool {
	return errors.Code(err) == ErrorCodeRevisionsExceeded
}

func IsErrorRevisionsCorrupt(err error) bool {
	return errors.Code(err) == ErrorCodeRevisionsCorrupt
}

func IsErrorRevisionConflict(err error) bool {
	return errors.Code(err) == ErrorCodeRevisionConflict
}

// DataSetOriginRevise resolves each datum of a batch by its origin id, inserts the new ones, writes nothing for the
// identical ones and updates the changed ones in place, recording every write in the datum's revisions. Deleting
// data by a data raw selector undoes everything the raw data record wrote.
type DataSetOriginRevise struct {
	*Base
}

func NewDataSetOriginRevise(dependencies Dependencies) (*DataSetOriginRevise, error) {
	base, err := NewBase(dependencies, DataSetOriginReviseName, DataSetOriginReviseVersion)
	if err != nil {
		return nil, err
	}

	return &DataSetOriginRevise{
		Base: base,
	}, nil
}

func (d *DataSetOriginRevise) AddData(ctx context.Context, dataSet *data.DataSet, dataSetData data.Data) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if dataSetData == nil {
		return errors.New("data set data is missing")
	}
	if len(dataSetData) == 0 {
		return nil
	}

	originIDs := make([]string, 0, len(dataSetData))
	seen := make(map[string]bool, len(dataSetData))
	for _, datum := range dataSetData {
		originID := originIDOf(datum)
		if originID == "" {
			return errors.New("origin id is missing")
		} else if seen[originID] {
			return errors.Newf("origin id %q is duplicated", originID)
		}
		seen[originID] = true
		originIDs = append(originIDs, originID)
	}

	selectors := data.Selectors{}
	for _, originID := range originIDs {
		selectors = append(selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: pointer.FromString(originID)}})
	}
	documents, err := d.DataStore.GetDataSetDocuments(ctx, dataSet, &selectors)
	if err != nil {
		return err
	}
	stored := make(map[string]*storedDatum, len(documents))
	for _, document := range documents {
		datum, err := decodeStoredDatum(document)
		if err != nil {
			return err
		} else if _, exists := stored[datum.originID()]; exists {
			return ErrorRevisionsCorrupt(datum.originID(), "is stored more than once")
		}
		stored[datum.originID()] = datum
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	var inserts data.Data
	for _, datum := range dataSetData {
		entry := data.Revision{Provenance: datum.GetProvenance(), Time: now}
		if existing, exists := stored[originIDOf(datum)]; exists {
			if err = d.revise(ctx, dataSet, datum, existing, entry); err != nil {
				return err
			}
		} else {
			datum.SetActive(true)
			datum.SetVersionInternal(1)
			datum.SetRevisions(&[]data.Revision{entry})
			inserts = append(inserts, datum)
		}
	}
	if len(inserts) == 0 {
		return nil
	}
	return d.DataStore.CreateDataSetData(ctx, dataSet, inserts)
}

// revise updates the stored datum in place when the incoming one differs, appending the revision that undoes it.
func (d *DataSetOriginRevise) revise(ctx context.Context, dataSet *data.DataSet, datum data.Datum, stored *storedDatum, entry data.Revision) error {
	datum.SetID(pointer.CloneString(stored.ID))
	datum.SetActive(true)

	document, err := revision.ToMap(datum)
	if err != nil {
		return err
	}
	incoming := revision.Canonicalize(revision.Project(document))
	current := revision.Canonicalize(revision.Project(stored.Document))

	forward := revision.CreatePatch(current, incoming)
	if len(forward) == 0 {
		return nil
	}

	// Clean removes everything a raw wrote before the raw is converted again, so a stored revision by the raw being
	// stored means the history is inconsistent
	if dataRawID := dataRawIDOf(entry.Provenance); dataRawID != "" {
		for _, existing := range stored.revisions() {
			if dataRawIDOf(existing.Provenance) == dataRawID {
				return ErrorRevisionsCorrupt(stored.originID(), "already has a revision by data raw id %q", dataRawID)
			}
		}
	}
	if len(stored.revisions()) >= RevisionsMaximum {
		return ErrorRevisionsExceeded(stored.originID(), len(stored.revisions()))
	}

	entry.Undo = revision.CreatePatch(incoming, current)
	update, err := revision.NewUpdate(forward, incoming)
	if err != nil {
		return err
	}
	revisions := append(slices.Clone(stored.revisions()), entry)
	if err = d.updateDatum(ctx, dataSet, stored, update, revisions); err != nil {
		return err
	}

	datum.SetVersionInternal(stored.VersionInternal + 1)
	datum.SetRevisions(&revisions)
	return nil
}

// DeleteData with data raw selectors undoes everything the raw data records wrote: a datum a raw only changed gets
// its revisions undone and popped, and a datum a raw created is destroyed. Any other selector destroys as the base
// deduplicator does.
func (d *DataSetOriginRevise) DeleteData(ctx context.Context, dataSet *data.DataSet, selectors *data.Selectors) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if dataSet == nil {
		return errors.New("data set is missing")
	}
	if selectors == nil {
		return errors.New("selectors is missing")
	}
	dataRawIDs, ok := dataRawIDsOf(selectors)
	if !ok {
		return d.Base.DeleteData(ctx, dataSet, selectors)
	}

	documents, err := d.DataStore.GetDataSetDocuments(ctx, dataSet, selectors)
	if err != nil {
		return err
	}
	for _, document := range documents {
		stored, err := decodeStoredDatum(document)
		if err != nil {
			return err
		}
		if err = d.undo(ctx, dataSet, stored, dataRawIDs); err != nil {
			return err
		}
	}

	// Undone datums no longer carry a revision by the raws, so only the datums the raws created still match
	return d.DataStore.DestroyDataSetData(ctx, dataSet, selectors)
}

// undo pops the stored datum's revisions by the raws from the newest down, as merge patches do not commute, unless
// the raws created the datum, which the destroy that follows removes. A revision by one of the raws below one by
// another raw cannot be undone.
func (d *DataSetOriginRevise) undo(ctx context.Context, dataSet *data.DataSet, stored *storedDatum, dataRawIDs map[string]bool) error {
	revisions := stored.revisions()
	count := len(revisions)
	for count > 0 && dataRawIDs[dataRawIDOf(revisions[count-1].Provenance)] {
		if isCreation(revisions, count-1) {
			return nil
		}
		count--
	}
	for index, existing := range revisions[:count] {
		if dataRawID := dataRawIDOf(existing.Provenance); dataRawIDs[dataRawID] {
			return ErrorRevisionsCorrupt(stored.originID(), "has revision %d by data raw id %q below one by another raw", index, dataRawID)
		}
	}
	if count == len(revisions) {
		return ErrorRevisionsCorrupt(stored.originID(), "has no revision by the data raw ids")
	}

	restored := revision.Project(stored.Document)
	for index := len(revisions) - 1; index >= count; index-- {
		if undo := revisions[index].Undo; undo != nil {
			restored = revision.Apply(restored, undo)
		}
	}
	undo := revision.CreatePatch(revision.Project(stored.Document), restored)
	update, err := revision.NewUpdate(undo, restored)
	if err != nil {
		return err
	}
	return d.updateDatum(ctx, dataSet, stored, update, revisions[:count])
}

// updateDatum writes the update and the revisions to the stored datum, guarded by the origin id and the version it
// was read at. A stale update that matches a datum a rewind destroyed and recreated at the same version is undone
// by the clean that precedes the next conversion, or refused as corrupt by it.
func (d *DataSetOriginRevise) updateDatum(ctx context.Context, dataSet *data.DataSet, stored *storedDatum, update bson.M, revisions []data.Revision) error {
	selectors := &data.Selectors{{
		Origin:  &data.SelectorOrigin{ID: pointer.FromString(stored.originID())},
		Version: pointer.FromInt(stored.VersionInternal),
	}}
	update = withRevisions(update, revisions)

	err := d.DataStore.UpdateDataSetDatum(ctx, dataSet, selectors, update)
	if dataStore.IsErrorDatumUpdateConflict(err) {
		return ErrorRevisionConflict(stored.originID(), stored.VersionInternal)
	}
	return err
}

// withRevisions returns the update with the revisions set, or unset when there are none left.
func withRevisions(update bson.M, revisions []data.Revision) bson.M {
	operator, value := "$set", any(revisions)
	if len(revisions) == 0 {
		operator, value = "$unset", 1
	}
	result := bson.M{}
	for key, fields := range update {
		result[key] = fields
	}
	fields := bson.M{}
	if existing, ok := result[operator].(bson.M); ok {
		for key, field := range existing {
			fields[key] = field
		}
	}
	fields["revisions"] = value
	result[operator] = fields
	return result
}

// storedDatum is a stored document decoded as a base datum, with the document itself for the diff.
type storedDatum struct {
	types.Base
	Document map[string]any
}

func decodeStoredDatum(document map[string]any) (*storedDatum, error) {
	bytes, err := bson.Marshal(document)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal stored datum")
	}
	stored := &storedDatum{Document: document}
	if err = storeStructuredMongo.UnmarshalDocument(bytes, &stored.Base); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal stored datum")
	}
	if stored.ID == nil || *stored.ID == "" || originIDOf(&stored.Base) == "" {
		return nil, errors.New("stored datum is missing an id or origin id")
	}
	return stored, nil
}

func (s *storedDatum) originID() string {
	return originIDOf(&s.Base)
}

func (s *storedDatum) revisions() []data.Revision {
	if s.Revisions == nil {
		return nil
	}
	return *s.Revisions
}

func originIDOf(datum data.Datum) string {
	if origin := datum.GetOrigin(); origin != nil && origin.ID != nil {
		return *origin.ID
	}
	return ""
}

func dataRawIDOf(provenance *data.Provenance) string {
	if provenance != nil && provenance.DataRaw != nil {
		return provenance.DataRaw.ID
	}
	return ""
}

// isCreation reports whether the revision at the index is the one that created the datum: the first, with no undo.
// A datum written by another deduplicator has no creation revision.
func isCreation(revisions []data.Revision, index int) bool {
	return index == 0 && revisions[0].Undo == nil
}

// dataRawIDsOf returns the data raw ids of the selectors when they are data raw selectors, skipping nil ones.
func dataRawIDsOf(selectors *data.Selectors) (map[string]bool, bool) {
	dataRawIDs := map[string]bool{}
	for _, selector := range *selectors {
		if selector == nil {
			continue
		} else if selector.DataRaw == nil || selector.DataRaw.ID == nil {
			return nil, false
		}
		dataRawIDs[*selector.DataRaw.ID] = true
	}
	return dataRawIDs, len(dataRawIDs) > 0
}
