package deduplicator_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("DataSetDropHashReplaceID", func() {
	It("DataSetDropHashReplaceIDName is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DataSetDropHashReplaceIDName).To(Equal("org.tidepool.deduplicator.dataset.drop.hash.replace.id"))
	})

	It("DataSetDropHashReplaceIDVersion is expected", func() {
		Expect(dataDeduplicatorDeduplicator.DataSetDropHashReplaceIDVersion).To(Equal("1.0.0"))
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

		It("NewDataSetDropHashReplaceID returns successfully", func() {
			Expect(dataDeduplicatorDeduplicator.NewDataSetDropHashReplaceID(dependencies)).ToNot(BeNil())
		})

		Context("with new deduplicator", func() {
			var ctx context.Context
			var deduplicator *dataDeduplicatorDeduplicator.DataSetDropHashReplaceID
			var dataSet *data.DataSet

			BeforeEach(func() {
				var err error
				ctx = context.Background()
				deduplicator, err = dataDeduplicatorDeduplicator.NewDataSetDropHashReplaceID(dependencies)
				Expect(err).ToNot(HaveOccurred())
				dataSet = dataTest.RandomDataSet()
				dataSet.Deduplicator.Name = pointer.FromString(dataDeduplicatorDeduplicator.DataSetDropHashReplaceIDName)
			})

			It("Get returns true when the deduplicator name matches", func() {
				Expect(deduplicator.Get(ctx, dataSet)).To(BeTrue())
			})

			Context("AddData", func() {
				var newDatum *dataTypes.Base
				var droppedDatum *dataTypes.Base
				var replacingDatum *dataTypes.Base
				var droppedStoredID string
				var replacedSelectors *data.Selectors

				newBase := func() *dataTypes.Base {
					base := dataTypesTest.RandomBase()
					base.Deduplicator = nil
					return base
				}

				BeforeEach(func() {
					newDatum = newBase()
					droppedDatum = newBase()
					replacingDatum = newBase()
					droppedStoredID = dataTest.RandomDatumID()
					replacedSelectors = &data.Selectors{{ID: pointer.CloneString(replacingDatum.ID)}}

					// The hashes AddData assigns
					assignedData := data.Data{newDatum, droppedDatum, replacingDatum}
					assignedData.SetUserID(dataSet.UserID)
					assignedData.SetDataSetID(dataSet.ID)
					Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(assignedData, dataTypes.IdentityFieldsVersionDataSetID)).To(Succeed())

					// The data set has data with the hashes of the dropped and replacing datum, only the latter with its id
					dataRepository.ExistingDataSetDataStub = func(_ context.Context, _ *data.DataSet, _ *data.Selectors) (*data.Selectors, error) {
						return &data.Selectors{
							{ID: pointer.From(droppedStoredID), Deduplicator: &data.SelectorDeduplicator{Hash: pointer.CloneString(droppedDatum.Deduplicator.Hash)}},
							{ID: pointer.CloneString(replacingDatum.ID), Deduplicator: &data.SelectorDeduplicator{Hash: pointer.CloneString(replacingDatum.Deduplicator.Hash)}},
						}, nil
					}
				})

				It("returns an error when the context is missing", func() {
					Expect(deduplicator.AddData(nil, dataSet, data.Data{newDatum})).To(MatchError("context is missing"))
				})

				It("returns an error when the data set is missing", func() {
					Expect(deduplicator.AddData(ctx, nil, data.Data{newDatum})).To(MatchError("data set is missing"))
				})

				It("returns an error when the data set data is missing", func() {
					Expect(deduplicator.AddData(ctx, dataSet, nil)).To(MatchError("data set data is missing"))
				})

				It("returns an error when existing data set data returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataRepository.ExistingDataSetDataStub = nil
					dataRepository.ExistingDataSetDataOutputs = []dataStoreTest.ExistingDataSetDataOutput{{Error: responseErr}}
					Expect(deduplicator.AddData(ctx, dataSet, data.Data{newDatum})).To(Equal(responseErr))
				})

				It("drops data whose hash is in the data set without its id, and adds the rest", func() {
					dataRepository.CreateDataSetDataOutputs = []error{nil}
					Expect(deduplicator.AddData(ctx, dataSet, data.Data{newDatum, droppedDatum})).To(Succeed())
					Expect(dataRepository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: data.Data{newDatum}}}))
					Expect(dataRepository.DeleteDataSetDataInputs).To(BeEmpty())
					Expect(dataRepository.DestroyDeletedDataSetDataInputs).To(BeEmpty())
				})

				It("replaces the datum in the data set whose hash and id the datum has", func() {
					dataRepository.DeleteDataSetDataOutputs = []error{nil}
					dataRepository.CreateDataSetDataOutputs = []error{nil}
					dataRepository.DestroyDeletedDataSetDataOutputs = []error{nil}
					Expect(deduplicator.AddData(ctx, dataSet, data.Data{newDatum, droppedDatum, replacingDatum})).To(Succeed())
					Expect(dataRepository.DeleteDataSetDataInputs).To(Equal([]dataStoreTest.DeleteDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: replacedSelectors}}))
					Expect(dataRepository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: data.Data{newDatum, replacingDatum}}}))
					Expect(dataRepository.DestroyDeletedDataSetDataInputs).To(Equal([]dataStoreTest.DestroyDeletedDataSetDataInput{{Context: ctx, DataSet: dataSet, Selectors: replacedSelectors}}))
				})

				It("replaces with the datum that has the id rather than a later datum with its hash", func() {
					duplicateDatum := dataTypesTest.CloneBase(replacingDatum)
					duplicateDatum.ID = pointer.FromString(dataTest.RandomDatumID())
					dataRepository.DeleteDataSetDataOutputs = []error{nil}
					dataRepository.CreateDataSetDataOutputs = []error{nil}
					dataRepository.DestroyDeletedDataSetDataOutputs = []error{nil}
					Expect(deduplicator.AddData(ctx, dataSet, data.Data{replacingDatum, duplicateDatum})).To(Succeed())
					Expect(dataRepository.CreateDataSetDataInputs).To(Equal([]dataStoreTest.CreateDataSetDataInput{{Context: ctx, DataSet: dataSet, DataSetData: data.Data{replacingDatum}}}))
				})

				It("returns an error without adding data when deleting the replaced data returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataRepository.DeleteDataSetDataOutputs = []error{responseErr}
					Expect(deduplicator.AddData(ctx, dataSet, data.Data{replacingDatum})).To(Equal(responseErr))
					Expect(dataRepository.CreateDataSetDataInputs).To(BeEmpty())
				})

				It("returns an error without destroying the replaced data when adding data returns an error", func() {
					responseErr := errorsTest.RandomError()
					dataRepository.DeleteDataSetDataOutputs = []error{nil}
					dataRepository.CreateDataSetDataOutputs = []error{responseErr}
					Expect(deduplicator.AddData(ctx, dataSet, data.Data{replacingDatum})).To(Equal(responseErr))
					Expect(dataRepository.DestroyDeletedDataSetDataInputs).To(BeEmpty())
				})
			})
		})
	})
})
