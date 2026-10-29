package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorTest "github.com/tidepool-org/platform/data/deduplicator/test"
	v1 "github.com/tidepool-org/platform/data/service/api/v1"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
	testRest "github.com/tidepool-org/platform/test/rest"
	"github.com/tidepool-org/platform/work"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("DataSetsUpdate", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var dataRepository *dataStoreTest.DataRepository
	var deduplicator *dataDeduplicatorTest.Deduplicator
	var deduplicatorFactory *dataDeduplicatorTest.Factory
	var res *testRest.ResponseWriter
	var svcCtx *mockDataServiceContext
	var dataSet *data.DataSet
	var closedDataSet *data.DataSet

	newRestRequest := func(body string) *rest.Request {
		req := httptest.NewRequest(http.MethodPut, "/v1/datasets/"+*dataSet.UploadID, strings.NewReader(body))
		req = req.WithContext(log.NewContextWithLogger(req.Context(), logTest.NewLogger()))
		req = req.WithContext(request.NewContextWithAuthDetails(req.Context(), request.NewAuthDetails(request.MethodServiceSecret, "", "token")))
		return &rest.Request{Request: req, PathParams: map[string]string{"dataSetId": *dataSet.UploadID}, Env: map[string]interface{}{}}
	}

	expectCreate := func() func() *work.Create {
		var created *work.Create
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, create *work.Create) (*work.Work, error) {
				created = create
				return &work.Work{ID: create.Type}, nil
			})
		return func() *work.Create { return created }
	}

	BeforeEach(func() {
		controller = gomock.NewController(GinkgoT())
		workClient = workTest.NewMockClient(controller)
		dataSet = dataTest.RandomDataSet()
		dataSet.DataSetType = pointer.FromString(data.DataSetTypeNormal)
		dataSet.State = pointer.FromString(data.DataSetStateOpen)
		dataSet.DataState = nil
		closedDataSet = dataTest.CloneDataSet(dataSet)
		closedDataSet.State = pointer.FromString(data.DataSetStateClosed)
		dataRepository = dataStoreTest.NewDataRepository()
		dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{DataSet: dataSet}}
		dataRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: closedDataSet}}
		deduplicator = dataDeduplicatorTest.NewDeduplicator()
		deduplicator.CloseOutput = pointer.FromAny[error](nil)
		deduplicatorFactory = dataDeduplicatorTest.NewFactory()
		deduplicatorFactory.GetOutput = &dataDeduplicatorTest.GetOutput{Deduplicator: deduplicator}
		res = testRest.NewResponseWriter()
		res.HeaderOutput = &http.Header{}
		res.WriteStub = func(bites []byte) (int, error) { return len(bites), nil }
		svcCtx = &mockDataServiceContext{
			t:                       GinkgoT(),
			dataDeduplicator:        deduplicator,
			dataDeduplicatorFactory: deduplicatorFactory,
			RestResponse:            res,
			TestWorkClient:          workClient,
			TestDataRepository:      dataRepository,
		}
		svcCtx.RestRequest = newRestRequest("")
	})

	AfterEach(func() {
		controller.Finish()
	})

	// Closing activates the data of the set, whatever its types, so every summary is reported
	It("closes the data set and reports the upload completed for every summary", func() {
		created := expectCreate()

		v1.DataSetsUpdate(svcCtx)

		Expect(svcCtx.RespondedStatusCode).To(Equal(http.StatusOK))
		Expect(svcCtx.RespondedObject).To(Equal(closedDataSet))
		Expect(deduplicator.CloseInvocations).To(Equal(1))
		Expect(created().Type).To(Equal(dataWorkPostprocess.Type))
		Expect(created().Metadata).To(HaveKeyWithValue("userId", *dataSet.UserID))
		Expect(created().Metadata).To(HaveKeyWithValue("reasons", ConsistOf(dataWorkPostprocess.ReasonUploadCompleted)))
		Expect(created().Metadata).To(HaveKeyWithValue("summaryTypes", ConsistOf(summaryTypes.AllSummaryTypes)))
	})

	It("responds successfully when the work cannot be created", func() {
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())

		v1.DataSetsUpdate(svcCtx)

		Expect(svcCtx.RespondedStatusCode).To(Equal(http.StatusOK))
	})

	It("reports nothing when a continuous data set is updated without being closed", func() {
		dataSet.DataSetType = pointer.FromString(data.DataSetTypeContinuous)
		dataRepository.UpdateDataSetOutputs = []dataStoreTest.UpdateDataSetOutput{{DataSet: dataSet}}
		svcCtx.RestRequest = newRestRequest(`{"deviceId":"test-device-id"}`)

		v1.DataSetsUpdate(svcCtx)

		Expect(svcCtx.RespondedStatusCode).To(Equal(http.StatusOK))
		Expect(deduplicator.CloseInvocations).To(Equal(0))
	})
})
