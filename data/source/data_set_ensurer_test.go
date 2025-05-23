package source_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("DataSetEnsurer", func() {
	Context("Ensure", func() {
		var (
			controller     *gomock.Controller
			client         *dataSourceTest.MockDataSetEnsurerClient
			factory        *dataSourceTest.MockDataSetEnsurerFactory
			dataSetEnsurer source.DataSetEnsurer
			ctx            context.Context
			userID         string
			dataSrc        source.Source
		)

		BeforeEach(func() {
			controller = gomock.NewController(GinkgoT())
			client = dataSourceTest.NewMockDataSetEnsurerClient(controller)
			factory = dataSourceTest.NewMockDataSetEnsurerFactory(controller)
			dataSetEnsurer = source.DataSetEnsurer{Client: client, Factory: factory}
			ctx = context.Background()
			userID = test.RandomString()
			dataSrc = source.Source{UserID: &userID}
		})

		AfterEach(func() {
			controller.Finish()
		})

		It("returns error if context is missing", func() {
			dataSet, err := dataSetEnsurer.Ensure(nil, dataSrc)
			Expect(dataSet).To(BeNil())
			Expect(err).To(MatchError("context is missing"))
		})

		It("returns error if client is missing", func() {
			dataSetEnsurer.Client = nil
			dataSet, err := dataSetEnsurer.Ensure(ctx, dataSrc)
			Expect(dataSet).To(BeNil())
			Expect(err).To(MatchError("client is missing"))
		})

		It("returns error if factory is missing", func() {
			dataSetEnsurer.Factory = nil
			dataSet, err := dataSetEnsurer.Ensure(ctx, dataSrc)
			Expect(dataSet).To(BeNil())
			Expect(err).To(MatchError("factory is missing"))
		})

		Context("with DataSetIDs", func() {
			BeforeEach(func() {
				dataSrc.DataSetIDs = pointer.FromStringArray([]string{"test_data_set_1", "test_data_set_2", "test_data_set_3", "test_data_set_4"})
			})

			It("returns error if GetDataSet returns error", func() {
				testErr := errorsTest.RandomError()
				client.EXPECT().GetDataSet(ctx, "test_data_set_1").Return(nil, testErr)
				dataSet, err := dataSetEnsurer.Ensure(ctx, dataSrc)
				Expect(dataSet).To(BeNil())
				Expect(err).To(MatchError(fmt.Sprintf("unable to get data set; %s", testErr)))
			})

			It("returns the first open DataSet", func() {
				closedDataSet := &data.DataSet{State: pointer.FromString(data.DataSetStateClosed)}
				openDataSet := &data.DataSet{State: pointer.FromString(data.DataSetStateOpen)}
				client.EXPECT().GetDataSet(ctx, "test_data_set_1").Return(nil, nil)
				client.EXPECT().GetDataSet(ctx, "test_data_set_2").Return(closedDataSet, nil)
				client.EXPECT().GetDataSet(ctx, "test_data_set_3").Return(openDataSet, nil)
				dataSet, err := dataSetEnsurer.Ensure(ctx, dataSrc)
				Expect(err).ToNot(HaveOccurred())
				Expect(dataSet).To(Equal(openDataSet))
			})

			Context("with no open DataSet", func() {
				var (
					dataSetCreate data.DataSetCreate
				)

				BeforeEach(func() {
					dataSetCreate = data.DataSetCreate{}
					client.EXPECT().GetDataSet(ctx, gomock.Any()).Return(nil, nil).Times(4)
					factory.EXPECT().NewDataSetCreate(gomock.Any()).Return(dataSetCreate)
				})

				It("returns error if CreateUserDataSet returns error", func() {
					testErr := errorsTest.RandomError()
					client.EXPECT().CreateUserDataSet(ctx, userID, &dataSetCreate).Return(nil, testErr)
					dataSet, err := dataSetEnsurer.Ensure(ctx, dataSrc)
					Expect(dataSet).To(BeNil())
					Expect(err).To(MatchError(fmt.Sprintf("unable to create data set; %s", testErr)))
				})

				It("creates a new DataSet if none open", func() {
					createdDataSet := &data.DataSet{}
					client.EXPECT().CreateUserDataSet(ctx, userID, &dataSetCreate).Return(createdDataSet, nil)
					dataSet, err := dataSetEnsurer.Ensure(ctx, dataSrc)
					Expect(err).ToNot(HaveOccurred())
					Expect(dataSet).To(Equal(createdDataSet))
				})
			})
		})
	})
})
