package service_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorTest "github.com/tidepool-org/platform/data/deduplicator/test"
	dataServiceService "github.com/tidepool-org/platform/data/service/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/request"
)

// dataStoreStub hands out one repository double so a test can inspect every call the client made.
type dataStoreStub struct {
	dataStore.Store
	dataRepository *dataStoreTest.DataRepository
}

func (d *dataStoreStub) NewDataRepository() dataStore.DataRepository {
	return d.dataRepository
}

var _ = Describe("Client", func() {
	var (
		mockController    *gomock.Controller
		dataRepository    *dataStoreTest.DataRepository
		str               *dataStoreStub
		deduplicatorMock  *dataDeduplicatorTest.MockDeduplicator
		deduplicatorFctry *dataDeduplicatorTest.MockFactory
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		dataRepository = dataStoreTest.NewDataRepository()
		str = &dataStoreStub{dataRepository: dataRepository}
		deduplicatorMock = dataDeduplicatorTest.NewMockDeduplicator(mockController)
		deduplicatorFctry = dataDeduplicatorTest.NewMockFactory(mockController)
	})

	AfterEach(func() {
		dataRepository.Expectations()
	})

	Context("NewClient", func() {
		It("returns an error when the data store is missing", func() {
			clnt, err := dataServiceService.NewClient(nil, deduplicatorFctry)
			Expect(err).To(MatchError("data store deprecated is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns an error when the data deduplicator factory is missing", func() {
			clnt, err := dataServiceService.NewClient(str, nil)
			Expect(err).To(MatchError("data deduplicator factory is missing"))
			Expect(clnt).To(BeNil())
		})

		It("returns a client", func() {
			Expect(dataServiceService.NewClient(str, deduplicatorFctry)).ToNot(BeNil())
		})
	})

	Context("with a client", func() {
		var (
			ctx  context.Context
			clnt *dataServiceService.Client
		)

		BeforeEach(func() {
			ctx = context.Background()
			var err error
			clnt, err = dataServiceService.NewClient(str, deduplicatorFctry)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("DeleteDataSet", func() {
			var dataSet *data.DataSet

			BeforeEach(func() {
				dataSet = dataTest.RandomDataSet()
			})

			It("returns an error when the data set cannot be retrieved", func() {
				responseErr := errorsTest.RandomError()
				dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{Error: responseErr}}
				Expect(clnt.DeleteDataSet(ctx, *dataSet.ID)).To(Equal(responseErr))
				Expect(dataRepository.GetDataSetInputs).To(Equal([]dataStoreTest.GetDataSetInput{{Context: ctx, ID: *dataSet.ID}}))
			})

			It("returns a resource not found error when the data set does not exist", func() {
				dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{}}
				err := clnt.DeleteDataSet(ctx, *dataSet.ID)
				Expect(err).To(HaveOccurred())
				Expect(request.IsErrorResourceNotFound(err)).To(BeTrue())
				Expect(dataRepository.DeleteDataSetInvocations).To(Equal(0))
			})

			It("returns an error when the deduplicator cannot be retrieved", func() {
				responseErr := errorsTest.RandomError()
				dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{DataSet: dataSet}}
				deduplicatorFctry.EXPECT().Get(ctx, dataSet).Return(nil, responseErr)
				Expect(clnt.DeleteDataSet(ctx, *dataSet.ID)).To(MatchError("unable to get deduplicator; " + responseErr.Error()))
				Expect(dataRepository.DeleteDataSetInvocations).To(Equal(0))
			})

			It("deletes with the repository when the data set has no deduplicator", func() {
				dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{DataSet: dataSet}}
				dataRepository.DeleteDataSetOutputs = []error{nil}
				deduplicatorFctry.EXPECT().Get(ctx, dataSet).Return(nil, nil)
				Expect(clnt.DeleteDataSet(ctx, *dataSet.ID)).To(Succeed())
				Expect(dataRepository.DeleteDataSetInputs).To(Equal([]dataStoreTest.DeleteDataSetInput{{Context: ctx, DataSet: dataSet}}))
			})

			It("returns an error when the repository cannot delete the data set", func() {
				responseErr := errorsTest.RandomError()
				dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{DataSet: dataSet}}
				dataRepository.DeleteDataSetOutputs = []error{responseErr}
				deduplicatorFctry.EXPECT().Get(ctx, dataSet).Return(nil, nil)
				Expect(clnt.DeleteDataSet(ctx, *dataSet.ID)).To(Equal(responseErr))
			})

			It("deletes with the deduplicator when the data set has one", func() {
				dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{DataSet: dataSet}}
				deduplicatorFctry.EXPECT().Get(ctx, dataSet).Return(deduplicatorMock, nil)
				deduplicatorMock.EXPECT().Delete(ctx, dataSet).Return(nil)
				Expect(clnt.DeleteDataSet(ctx, *dataSet.ID)).To(Succeed())
				Expect(dataRepository.DeleteDataSetInvocations).To(Equal(0))
			})

			It("returns an error when the deduplicator cannot delete the data set", func() {
				responseErr := errorsTest.RandomError()
				dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{DataSet: dataSet}}
				deduplicatorFctry.EXPECT().Get(ctx, dataSet).Return(deduplicatorMock, nil)
				deduplicatorMock.EXPECT().Delete(ctx, dataSet).Return(responseErr)
				Expect(clnt.DeleteDataSet(ctx, *dataSet.ID)).To(Equal(responseErr))
			})
		})
	})
})
