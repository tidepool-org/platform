package deduplicator_test

import (
	"context"
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/data/revision"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

// setOf returns the fields an update sets, the revisions aside; unsetOf the keys it removes, sorted, the revisions
// aside; revisionsOf the revisions it sets, nil when it unsets them.
func setOf(update bson.M) map[string]any {
	result := map[string]any{}
	if set, ok := update["$set"].(bson.M); ok {
		for key, value := range set {
			if key != "revisions" {
				result[key] = value
			}
		}
	}
	return result
}

func unsetOf(update bson.M) []string {
	result := []string{}
	if unset, ok := update["$unset"].(bson.M); ok {
		for key := range unset {
			if key != "revisions" {
				result = append(result, key)
			}
		}
	}
	sort.Strings(result)
	return result
}

func revisionsOf(update bson.M) []data.Revision {
	if set, ok := update["$set"].(bson.M); ok {
		if revisions, ok := set["revisions"].([]data.Revision); ok {
			return revisions
		}
	}
	return nil
}

var _ = Describe("DataSetOriginRevise", func() {
	It("DataSetOriginReviseName is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DataSetOriginReviseName).To(Equal("org.tidepool.deduplicator.dataset.origin.revise"))
	})

	It("DataSetOriginReviseVersion is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DataSetOriginReviseVersion).To(Equal("1.0.0"))
	})

	It("RevisionsMaximum is expected", func() {
		Expect(dataDeduplicatorDeduplicator.RevisionsMaximum).To(Equal(50))
	})

	Context("with dependencies", func() {
		var dataSetRepository *dataStoreTest.DataRepository
		var dataRepository *dataStoreTest.DataRepository
		var dependencies dataDeduplicatorDeduplicator.Dependencies

		BeforeEach(func() {
			dataSetRepository = dataStoreTest.NewDataRepository()
			dataRepository = dataStoreTest.NewDataRepository()
			dependencies = dataDeduplicatorDeduplicator.Dependencies{
				DataSetStore: dataSetRepository,
				DataStore:    dataRepository,
			}
		})

		AfterEach(func() {
			dataRepository.AssertOutputsEmpty()
			dataSetRepository.AssertOutputsEmpty()
		})

		Context("NewDataSetOriginRevise", func() {
			It("returns an error when the dependencies are invalid", func() {
				dependencies.DataStore = nil
				deduplicator, err := dataDeduplicatorDeduplicator.NewDataSetOriginRevise(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; data store is missing"))
				Expect(deduplicator).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(dataDeduplicatorDeduplicator.NewDataSetOriginRevise(dependencies)).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator", func() {
			var deduplicator *dataDeduplicatorDeduplicator.DataSetOriginRevise
			var ctx context.Context
			var dataSet *data.DataSet
			var dataRawID string

			// newDatum is an incoming datum with an origin id and the raw's provenance, as the plugin sends it
			newDatum := func(originID string) *types.Base {
				base := dataTypesTest.RandomBase()
				base.Active = false
				base.CreatedTime = nil
				base.CreatedUserID = nil
				base.ModifiedTime = nil
				base.ModifiedUserID = nil
				base.DeletedTime = nil
				base.DeletedUserID = nil
				base.ArchivedTime = nil
				base.ArchivedDataSetID = nil
				base.UploadID = nil
				base.UserID = nil
				base.VersionInternal = 0
				base.Revisions = nil
				base.Origin.ID = pointer.FromString(originID)
				base.Provenance = &data.Provenance{ClientID: "tandem", DataRaw: &data.DataRawReference{ID: dataRawID}}
				return base
			}
			entry := func(rawID string, undo map[string]any) data.Revision {
				return data.Revision{
					Provenance: &data.Provenance{ClientID: "tandem", DataRaw: &data.DataRawReference{ID: rawID}},
					Time:       test.RandomTime(),
					Undo:       undo,
				}
			}
			// storedDocument is the datum as its stored document decodes, after an earlier insert at the version
			storedDocument := func(datum *types.Base, version int, revisions []data.Revision) map[string]any {
				stored := dataTypesTest.CloneBase(datum)
				stored.Active = true
				stored.ID = pointer.FromString(dataTest.RandomDatumID())
				stored.UserID = dataSet.UserID
				stored.UploadID = dataSet.UploadID
				stored.CreatedTime = pointer.FromTime(test.RandomTime())
				stored.ModifiedTime = pointer.FromTime(test.RandomTime())
				stored.VersionInternal = version
				stored.Revisions = &revisions
				document, err := revision.ToMap(stored)
				Expect(err).ToNot(HaveOccurred())
				document["_id"] = primitive.NewObjectID()
				return document
			}

			BeforeEach(func() {
				var err error
				deduplicator, err = dataDeduplicatorDeduplicator.NewDataSetOriginRevise(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(deduplicator).ToNot(BeNil())
				ctx = context.Background()
				dataSet = dataTest.RandomDataSet()
				dataSet.Deduplicator.Name = pointer.FromString(dataDeduplicatorDeduplicator.DataSetOriginReviseName)
				dataRawID = dataTest.RandomDataRawID()
			})

			Context("New and Get", func() {
				It("return true when the data set deduplicator name matches", func() {
					Expect(deduplicator.New(ctx, dataSet)).To(BeTrue())
					Expect(deduplicator.Get(ctx, dataSet)).To(BeTrue())
				})

				It("return false when the data set deduplicator name does not match", func() {
					dataSet.Deduplicator.Name = pointer.FromString("org.tidepool.deduplicator.other")
					Expect(deduplicator.New(ctx, dataSet)).To(BeFalse())
					Expect(deduplicator.Get(ctx, dataSet)).To(BeFalse())
				})
			})

			originSelectors := func(originIDs ...string) *data.Selectors {
				selectors := data.Selectors{}
				for _, originID := range originIDs {
					selectors = append(selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: pointer.FromString(originID)}})
				}
				return &selectors
			}
			dataRawSelectors := func(dataRawIDs ...string) *data.Selectors {
				selectors := data.Selectors{}
				for _, dataRawID := range dataRawIDs {
					selectors = append(selectors, &data.Selector{DataRaw: &data.SelectorDataRaw{ID: pointer.FromString(dataRawID)}})
				}
				return &selectors
			}

			Context("AddData", func() {
				var dataSetData data.Data
				var originIDs []string

				BeforeEach(func() {
					dataSetData = data.Data{}
					originIDs = nil
					for index := range test.RandomIntFromRange(1, 3) {
						originID := dataRawID + ":" + string(rune('a'+index)) + ":kind"
						dataSetData = append(dataSetData, newDatum(originID))
						originIDs = append(originIDs, originID)
					}
				})

				It("returns an error when the context is missing", func() {
					Expect(deduplicator.AddData(nil, dataSet, dataSetData)).To(MatchError("context is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.AddData(ctx, nil, dataSetData)).To(MatchError("data set is missing"))
				})

				It("returns an error when the data set data is missing", func() {
					Expect(deduplicator.AddData(ctx, dataSet, nil)).To(MatchError("data set data is missing"))
				})

				It("returns successfully without writing when the data set data is empty", func() {
					Expect(deduplicator.AddData(ctx, dataSet, data.Data{})).To(Succeed())
					Expect(dataRepository.GetDataSetDocumentsInvocations).To(Equal(0))
					Expect(dataRepository.CreateDataSetDataInvocations).To(Equal(0))
				})

				It("returns an error when a datum has no origin", func() {
					dataSetData[0].(*types.Base).Origin = nil
					Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(MatchError("origin id is missing"))
				})

				It("returns an error when a datum has no origin id", func() {
					dataSetData[0].(*types.Base).Origin.ID = nil
					Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(MatchError("origin id is missing"))
				})

				It("returns an error when an origin id is duplicated within the batch", func() {
					dataSetData = append(dataSetData, newDatum(originIDs[0]))
					Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(MatchError(`origin id "` + originIDs[0] + `" is duplicated`))
				})

				It("returns the error when fetching the stored documents fails", func() {
					responseErr := errorsTest.RandomError()
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Error: responseErr}}
					Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Equal(responseErr))
					Expect(dataRepository.GetDataSetDocumentsInputs).To(Equal([]dataStoreTest.GetDataSetDocumentsInput{{Context: ctx, DataSet: dataSet, Selectors: originSelectors(originIDs...)}}))
				})

				It("returns a corrupt error when an origin id is stored more than once", func() {
					datum := dataSetData[0].(*types.Base)
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{
						storedDocument(datum, 1, []data.Revision{entry(dataTest.RandomDataRawID(), nil)}),
						storedDocument(datum, 1, []data.Revision{entry(dataTest.RandomDataRawID(), nil)}),
					}}}
					err := deduplicator.AddData(ctx, dataSet, dataSetData)
					Expect(dataDeduplicatorDeduplicator.IsErrorRevisionsCorrupt(err)).To(BeTrue())
					Expect(err).To(MatchError(`origin id "` + originIDs[0] + `" is stored more than once`))
				})

				When("nothing is stored", func() {
					BeforeEach(func() {
						dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{}}}
					})

					It("returns the error when creating the data fails", func() {
						responseErr := errorsTest.RandomError()
						dataRepository.CreateDataSetDataOutputs = []error{responseErr}
						Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Equal(responseErr))
					})

					It("inserts every datum active, at version 1, with a creation revision and no separate activation", func() {
						dataRepository.CreateDataSetDataOutputs = []error{nil}
						before := time.Now()
						Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Succeed())
						Expect(dataRepository.CreateDataSetDataInputs).To(HaveLen(1))
						Expect(dataRepository.CreateDataSetDataInputs[0].DataSet).To(Equal(dataSet))
						Expect(dataRepository.CreateDataSetDataInputs[0].DataSetData).To(Equal([]data.Datum(dataSetData)))
						Expect(dataRepository.ActivateDataSetDataInvocations).To(Equal(0))
						Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(0))
						for _, datum := range dataSetData {
							base := datum.(*types.Base)
							Expect(base.Active).To(BeTrue())
							Expect(base.VersionInternal).To(Equal(1))
							Expect(base.Revisions).To(PointTo(HaveLen(1)))
							creation := (*base.Revisions)[0]
							Expect(creation.Provenance).To(Equal(base.Provenance))
							Expect(creation.Time).To(BeTemporally(">=", before.Truncate(time.Millisecond)))
							Expect(creation.Time).To(BeTemporally("<=", time.Now()))
							Expect(creation.Undo).To(BeNil())
						}
					})
				})

				When("the datum is stored", func() {
					var datum *types.Base
					var original *types.Base
					var stored map[string]any
					var storedRevisions []data.Revision
					var storedID string

					BeforeEach(func() {
						dataSetData = dataSetData[:1]
						originIDs = originIDs[:1]
						datum = dataSetData[0].(*types.Base)
						original = dataTypesTest.CloneBase(datum)
						storedRevisions = []data.Revision{entry(dataTest.RandomDataRawID(), nil), entry(dataTest.RandomDataRawID(), map[string]any{"deviceTime": "2026-01-01T00:00:00"})}
						stored = storedDocument(datum, 3, storedRevisions)
						storedID = stored["id"].(string)
						dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
					})

					It("writes nothing when the datum is identical", func() {
						Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Succeed())
						Expect(dataRepository.CreateDataSetDataInvocations).To(Equal(0))
						Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(0))
						Expect(datum.ID).To(Equal(pointer.FromString(storedID)))
					})

					When("the datum changed", func() {
						var previousDeviceTime string
						var previousNotes []string
						var previousTags []string

						BeforeEach(func() {
							previousDeviceTime = *datum.DeviceTime
							previousNotes = *datum.Notes
							previousTags = *datum.Tags
							datum.DeviceTime = pointer.FromString("2026-10-20T08:00:00")
							datum.Notes = nil
							datum.Tags = pointer.FromStringArray([]string{"changed"})
						})

						It("updates the stored datum in place with the forward change and a revision that undoes it", func() {
							dataRepository.UpdateDataSetDatumOutputs = []error{nil}
							before := time.Now()
							Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Succeed())
							Expect(dataRepository.CreateDataSetDataInvocations).To(Equal(0))
							Expect(dataRepository.UpdateDataSetDatumInputs).To(HaveLen(1))
							input := dataRepository.UpdateDataSetDatumInputs[0]
							Expect(input.Context).To(Equal(ctx))
							Expect(input.DataSet).To(Equal(dataSet))
							Expect(input.Selectors).To(Equal(&data.Selectors{{Origin: &data.SelectorOrigin{ID: pointer.FromString(originIDs[0])}, Version: pointer.FromInt(3)}}))
							Expect(setOf(input.Update)).To(Equal(map[string]any{"deviceTime": "2026-10-20T08:00:00", "tags": []any{"changed"}}))
							Expect(unsetOf(input.Update)).To(Equal([]string{"notes"}))
							Expect(revisionsOf(input.Update)).To(HaveLen(3))
							Expect(revisionsOf(input.Update)[:2]).To(Equal(storedRevisions))
							added := revisionsOf(input.Update)[2]
							Expect(added.Provenance).To(Equal(datum.Provenance))
							Expect(added.Time).To(BeTemporally(">=", before.Truncate(time.Millisecond)))
							Expect(added.Undo).To(Equal(map[string]any{
								"deviceTime": previousDeviceTime,
								"notes":      []any{previousNotes[0], previousNotes[1]},
								"tags":       []any{previousTags[0]},
							}))
							Expect(datum.ID).To(Equal(pointer.FromString(storedID)))
							Expect(datum.Active).To(BeTrue())
							Expect(datum.VersionInternal).To(Equal(4))
							Expect(datum.Revisions).To(PointTo(Equal(revisionsOf(input.Update))))
						})

						It("expects no version when the stored document has none", func() {
							stored = storedDocument(original, 0, storedRevisions)
							Expect(stored).ToNot(HaveKey("_version"))
							dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
							dataRepository.UpdateDataSetDatumOutputs = []error{nil}
							Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Succeed())
							Expect(dataRepository.UpdateDataSetDatumInputs[0].Selectors).To(Equal(&data.Selectors{{Origin: &data.SelectorOrigin{ID: pointer.FromString(originIDs[0])}, Version: pointer.FromInt(0)}}))
							Expect(datum.VersionInternal).To(Equal(1))
						})

						It("appends to a datum without revisions, written by another deduplicator", func() {
							stored = storedDocument(original, 0, nil)
							dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
							dataRepository.UpdateDataSetDatumOutputs = []error{nil}
							Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Succeed())
							Expect(revisionsOf(dataRepository.UpdateDataSetDatumInputs[0].Update)).To(HaveLen(1))
							Expect(revisionsOf(dataRepository.UpdateDataSetDatumInputs[0].Update)[0].Undo).To(HaveKey("deviceTime"))
						})

						It("returns a corrupt error when the stored datum already has a revision by the raw", func() {
							stored = storedDocument(original, 3, []data.Revision{entry(dataRawID, nil), entry(dataTest.RandomDataRawID(), map[string]any{"deviceTime": "x"})})
							dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
							err := deduplicator.AddData(ctx, dataSet, dataSetData)
							Expect(dataDeduplicatorDeduplicator.IsErrorRevisionsCorrupt(err)).To(BeTrue())
							Expect(err).To(MatchError(`origin id "` + originIDs[0] + `" already has a revision by data raw id "` + dataRawID + `"`))
							Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(0))
						})

						It("returns an exceeded error without writing when the stored datum is at the revisions maximum", func() {
							revisions := []data.Revision{}
							for range dataDeduplicatorDeduplicator.RevisionsMaximum {
								revisions = append(revisions, entry(dataTest.RandomDataRawID(), map[string]any{"deviceTime": "x"}))
							}
							stored = storedDocument(original, 3, revisions)
							dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
							err := deduplicator.AddData(ctx, dataSet, dataSetData)
							Expect(dataDeduplicatorDeduplicator.IsErrorRevisionsExceeded(err)).To(BeTrue())
							Expect(err).To(MatchError(`origin id "` + originIDs[0] + `" has 50 revisions`))
							Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(0))
						})

						It("returns a conflict error when the stored datum changed since it was read", func() {
							dataRepository.UpdateDataSetDatumOutputs = []error{dataStore.ErrorDatumUpdateConflict()}
							err := deduplicator.AddData(ctx, dataSet, dataSetData)
							Expect(dataDeduplicatorDeduplicator.IsErrorRevisionConflict(err)).To(BeTrue())
							Expect(err).To(MatchError(`origin id "` + originIDs[0] + `" at version 3`))
						})

						It("returns any other update error", func() {
							responseErr := errorsTest.RandomError()
							dataRepository.UpdateDataSetDatumOutputs = []error{responseErr}
							Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Equal(responseErr))
						})
					})
				})

				It("updates the changed datum, skips the identical one and inserts the new one in a mixed batch", func() {
					changed := newDatum(dataRawID + ":1:kind")
					identical := newDatum(dataRawID + ":2:kind")
					added := newDatum(dataRawID + ":3:kind")
					dataSetData = data.Data{changed, identical, added}
					storedChanged := storedDocument(changed, 1, []data.Revision{entry(dataTest.RandomDataRawID(), nil)})
					storedIdentical := storedDocument(identical, 1, []data.Revision{entry(dataTest.RandomDataRawID(), nil)})
					changed.DeviceTime = pointer.FromString("2026-10-20T08:00:00")
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{storedIdentical, storedChanged}}}
					dataRepository.UpdateDataSetDatumOutputs = []error{nil}
					dataRepository.CreateDataSetDataOutputs = []error{nil}
					Expect(deduplicator.AddData(ctx, dataSet, dataSetData)).To(Succeed())
					Expect(dataRepository.GetDataSetDocumentsInputs[0].Selectors).To(Equal(originSelectors(dataRawID+":1:kind", dataRawID+":2:kind", dataRawID+":3:kind")))
					Expect(dataRepository.UpdateDataSetDatumInputs).To(HaveLen(1))
					Expect(dataRepository.UpdateDataSetDatumInputs[0].Selectors).To(Equal(&data.Selectors{{Origin: &data.SelectorOrigin{ID: pointer.FromString(dataRawID + ":1:kind")}, Version: pointer.FromInt(1)}}))
					Expect(setOf(dataRepository.UpdateDataSetDatumInputs[0].Update)).To(Equal(map[string]any{"deviceTime": "2026-10-20T08:00:00"}))
					Expect(dataRepository.CreateDataSetDataInputs).To(HaveLen(1))
					Expect(dataRepository.CreateDataSetDataInputs[0].DataSetData).To(Equal([]data.Datum{added}))
					Expect(identical.ID).To(Equal(pointer.FromString(storedIdentical["id"].(string))))
					Expect(identical.Revisions).To(BeNil())
				})
			})

			Context("DeleteData", func() {
				It("returns an error when the context is missing", func() {
					Expect(deduplicator.DeleteData(nil, dataSet, dataRawSelectors(dataRawID))).To(MatchError("context is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.DeleteData(ctx, nil, dataRawSelectors(dataRawID))).To(MatchError("data set is missing"))
				})

				It("returns an error when the selectors are missing", func() {
					Expect(deduplicator.DeleteData(ctx, dataSet, nil)).To(MatchError("selectors is missing"))
				})

				It("destroys the data as the base deduplicator does when the selectors are not data raw selectors", func() {
					selectors := originSelectors(dataRawID + ":1:kind")
					dataRepository.DestroyDataSetDataOutputs = []error{nil}
					Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Succeed())
					Expect(dataRepository.GetDataSetDocumentsInvocations).To(Equal(0))
					Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
				})

				It("destroys as the base deduplicator does when data raw selectors are mixed with another kind, which the store refuses", func() {
					selectors := dataRawSelectors(dataRawID)
					*selectors = append(*selectors, &data.Selector{ID: pointer.FromString(dataTest.RandomDatumID())})
					dataRepository.DestroyDataSetDataOutputs = []error{nil}
					Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Succeed())
					Expect(dataRepository.GetDataSetDocumentsInvocations).To(Equal(0))
					Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
				})

				It("skips a nil selector among data raw selectors", func() {
					selectors := dataRawSelectors(dataRawID)
					*selectors = append(*selectors, nil)
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{}}}
					dataRepository.DestroyDataSetDataOutputs = []error{nil}
					Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Succeed())
					Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
				})

				It("returns the error when fetching the documents fails", func() {
					responseErr := errorsTest.RandomError()
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Error: responseErr}}
					Expect(deduplicator.DeleteData(ctx, dataSet, dataRawSelectors(dataRawID))).To(Equal(responseErr))
					Expect(dataRepository.GetDataSetDocumentsInputs).To(Equal([]dataStoreTest.GetDataSetDocumentsInput{{Context: ctx, DataSet: dataSet, Selectors: dataRawSelectors(dataRawID)}}))
					Expect(dataRepository.DestroyDataSetDataInvocations).To(Equal(0))
				})

				It("destroys by the data raw selectors without any undo when the raw revised nothing", func() {
					selectors := dataRawSelectors(dataRawID)
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{}}}
					dataRepository.DestroyDataSetDataOutputs = []error{nil}
					Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Succeed())
					Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(0))
					Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
				})

				It("returns the error when destroying fails", func() {
					responseErr := errorsTest.RandomError()
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{}}}
					dataRepository.DestroyDataSetDataOutputs = []error{responseErr}
					Expect(deduplicator.DeleteData(ctx, dataSet, dataRawSelectors(dataRawID))).To(Equal(responseErr))
				})

				When("the raw created the datum", func() {
					BeforeEach(func() {
						stored := storedDocument(newDatum(dataRawID+":1:kind"), 1, []data.Revision{entry(dataRawID, nil)})
						dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
					})

					It("leaves the datum to the destroy without updating it", func() {
						dataRepository.DestroyDataSetDataOutputs = []error{nil}
						Expect(deduplicator.DeleteData(ctx, dataSet, dataRawSelectors(dataRawID))).To(Succeed())
						Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(0))
						Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: dataRawSelectors(dataRawID)}}))
					})
				})

				When("the raw revised the datum", func() {
					var stored map[string]any
					var revisions []data.Revision

					BeforeEach(func() {
						datum := newDatum(dataRawID + ":1:kind")
						datum.Notes = nil
						revisions = []data.Revision{
							entry(dataTest.RandomDataRawID(), nil),
							entry(dataRawID, map[string]any{"deviceTime": "2026-01-01T00:00:00", "notes": []any{"restored"}, "tags": nil}),
						}
						stored = storedDocument(datum, 2, revisions)
						dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
					})

					It("applies the undo and pops the revision under the version guard, then destroys by the selectors", func() {
						dataRepository.UpdateDataSetDatumOutputs = []error{nil}
						dataRepository.DestroyDataSetDataOutputs = []error{nil}
						Expect(deduplicator.DeleteData(ctx, dataSet, dataRawSelectors(dataRawID))).To(Succeed())
						Expect(dataRepository.UpdateDataSetDatumInputs).To(HaveLen(1))
						input := dataRepository.UpdateDataSetDatumInputs[0]
						Expect(input.Selectors).To(Equal(&data.Selectors{{Origin: &data.SelectorOrigin{ID: pointer.FromString(dataRawID + ":1:kind")}, Version: pointer.FromInt(2)}}))
						Expect(setOf(input.Update)).To(Equal(map[string]any{"deviceTime": "2026-01-01T00:00:00", "notes": []any{"restored"}}))
						Expect(unsetOf(input.Update)).To(Equal([]string{"tags"}))
						Expect(revisionsOf(input.Update)).To(Equal(revisions[:1]))
						Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: dataRawSelectors(dataRawID)}}))
					})

					It("returns a conflict error without destroying when the stored datum changed since it was read", func() {
						dataRepository.UpdateDataSetDatumOutputs = []error{dataStore.ErrorDatumUpdateConflict()}
						err := deduplicator.DeleteData(ctx, dataSet, dataRawSelectors(dataRawID))
						Expect(dataDeduplicatorDeduplicator.IsErrorRevisionConflict(err)).To(BeTrue())
						Expect(dataRepository.DestroyDataSetDataInvocations).To(Equal(0))
					})
				})

				It("returns a corrupt error without writing when the raw's revision sits below one by another raw", func() {
					stored := storedDocument(newDatum(dataRawID+":1:kind"), 2, []data.Revision{entry(dataRawID, nil), entry(dataTest.RandomDataRawID(), map[string]any{"deviceTime": "x"})})
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
					err := deduplicator.DeleteData(ctx, dataSet, dataRawSelectors(dataRawID))
					Expect(dataDeduplicatorDeduplicator.IsErrorRevisionsCorrupt(err)).To(BeTrue())
					Expect(err).To(MatchError(`origin id "` + dataRawID + `:1:kind" has revision 0 by data raw id "` + dataRawID + `" below one by another raw`))
					Expect(dataRepository.DestroyDataSetDataInvocations).To(Equal(0))
					Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(0))
				})

				It("undoes a datum another deduplicator wrote, whose only revision is by the raw, instead of destroying it", func() {
					datum := newDatum(dataRawID + ":1:kind")
					stored := storedDocument(datum, 1, []data.Revision{entry(dataRawID, map[string]any{"deviceTime": "2026-01-01T00:00:00"})})
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
					dataRepository.UpdateDataSetDatumOutputs = []error{nil}
					dataRepository.DestroyDataSetDataOutputs = []error{nil}
					Expect(deduplicator.DeleteData(ctx, dataSet, dataRawSelectors(dataRawID))).To(Succeed())
					input := dataRepository.UpdateDataSetDatumInputs[0]
					Expect(setOf(input.Update)).To(Equal(map[string]any{"deviceTime": "2026-01-01T00:00:00"}))
					Expect(revisionsOf(input.Update)).To(BeEmpty())
					Expect(input.Update["$unset"]).To(Equal(bson.M{"revisions": 1}))
				})

				It("undoes the revisions of several raws of one datum newest first, whatever the selector order", func() {
					otherDataRawID := dataTest.RandomDataRawID()
					datum := newDatum(dataRawID + ":1:kind")
					datum.DeviceTime = pointer.FromString("2026-03-03T00:00:00")
					stored := storedDocument(datum, 3, []data.Revision{
						entry(dataTest.RandomDataRawID(), nil),
						entry(dataRawID, map[string]any{"deviceTime": "2026-01-01T00:00:00"}),
						entry(otherDataRawID, map[string]any{"deviceTime": "2026-02-02T00:00:00"}),
					})
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{stored}}}
					dataRepository.UpdateDataSetDatumOutputs = []error{nil}
					dataRepository.DestroyDataSetDataOutputs = []error{nil}
					selectors := dataRawSelectors(dataRawID, otherDataRawID)
					Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Succeed())
					input := dataRepository.UpdateDataSetDatumInputs[0]
					Expect(setOf(input.Update)).To(Equal(map[string]any{"deviceTime": "2026-01-01T00:00:00"}))
					Expect(revisionsOf(input.Update)).To(HaveLen(1))
					Expect(revisionsOf(input.Update)[0].Undo).To(BeNil())
				})

				It("undoes every revised datum of every raw with one fetch, then destroys once by all the selectors", func() {
					otherDataRawID := dataTest.RandomDataRawID()
					created := storedDocument(newDatum(dataRawID+":1:kind"), 1, []data.Revision{entry(dataRawID, nil)})
					revised := storedDocument(newDatum(dataRawID+":2:kind"), 2, []data.Revision{entry(dataTest.RandomDataRawID(), nil), entry(dataRawID, map[string]any{"deviceTime": "x"})})
					revisedByOther := storedDocument(newDatum(otherDataRawID+":1:kind"), 2, []data.Revision{entry(dataTest.RandomDataRawID(), nil), entry(otherDataRawID, map[string]any{"deviceTime": "y"})})
					selectors := dataRawSelectors(dataRawID, otherDataRawID)
					dataRepository.GetDataSetDocumentsOutputs = []dataStoreTest.GetDataSetDocumentsOutput{{Documents: []map[string]any{created, revised, revisedByOther}}}
					dataRepository.UpdateDataSetDatumOutputs = []error{nil, nil}
					dataRepository.DestroyDataSetDataOutputs = []error{nil}
					Expect(deduplicator.DeleteData(ctx, dataSet, selectors)).To(Succeed())
					Expect(dataRepository.GetDataSetDocumentsInputs).To(Equal([]dataStoreTest.GetDataSetDocumentsInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
					Expect(dataRepository.UpdateDataSetDatumInvocations).To(Equal(2))
					Expect(setOf(dataRepository.UpdateDataSetDatumInputs[0].Update)).To(Equal(map[string]any{"deviceTime": "x"}))
					Expect(setOf(dataRepository.UpdateDataSetDatumInputs[1].Update)).To(Equal(map[string]any{"deviceTime": "y"}))
					Expect(dataRepository.DestroyDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: selectors}}))
				})
			})
		})
	})
})
