package v1_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

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

var _ = Describe("DataSetsDataCreate", func() {
	var controller *gomock.Controller
	var workClient *workTest.MockClient
	var dataRepository *dataStoreTest.DataRepository
	var deduplicator *dataDeduplicatorTest.Deduplicator
	var deduplicatorFactory *dataDeduplicatorTest.Factory
	var res *testRest.ResponseWriter
	var svcCtx *mockDataServiceContext
	var dataSet *data.DataSet

	newRestRequest := func(body string) *rest.Request {
		req := httptest.NewRequest(http.MethodPost, "/v1/datasets/"+*dataSet.UploadID+"/data", strings.NewReader(body))
		req = req.WithContext(log.NewContextWithLogger(req.Context(), logTest.NewLogger()))
		req = req.WithContext(request.NewContextWithAuthDetails(req.Context(), request.NewAuthDetails(request.MethodServiceSecret, "", "token")))
		return &rest.Request{Request: req, PathParams: map[string]string{"dataSetId": *dataSet.UploadID}, Env: map[string]interface{}{}}
	}

	glucoseDatum := func(typ string, tm time.Time) string {
		return fmt.Sprintf(`{"type":%q,"time":%q,"units":"mmol/L","value":5.5,"deviceId":"test-device-id"}`, typ, tm.Format(time.RFC3339Nano))
	}

	// The deduplicator of a continuous data set activates the data as it is added
	activateData := func() {
		deduplicator.AddDataStub = func(_ context.Context, _ *data.DataSet, dataSetData data.Data) error {
			dataSetData.SetActive(true)
			return nil
		}
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
		dataSet.State = pointer.FromString(data.DataSetStateOpen)
		dataSet.DataState = nil
		dataRepository = dataStoreTest.NewDataRepository()
		dataRepository.GetDataSetOutputs = []dataStoreTest.GetDataSetOutput{{DataSet: dataSet}}
		deduplicator = dataDeduplicatorTest.NewDeduplicator()
		deduplicator.AddDataOutput = pointer.FromAny[error](nil)
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
	})

	AfterEach(func() {
		controller.Finish()
	})

	It("reports the summaries the data added affects", func() {
		activateData()
		svcCtx.RestRequest = newRestRequest("[" + glucoseDatum("cbg", time.Now()) + "]")
		created := expectCreate()

		v1.DataSetsDataCreate(svcCtx)

		Expect(svcCtx.RespondedStatusCode).To(Equal(http.StatusOK))
		Expect(created().Type).To(Equal(dataWorkPostprocess.Type))
		Expect(created().Metadata).To(HaveKeyWithValue("userId", *dataSet.UserID))
		Expect(created().Metadata).To(HaveKeyWithValue("reasons", ConsistOf(dataWorkPostprocess.ReasonDataAdded)))
		Expect(created().Metadata).To(HaveKeyWithValue("summaryTypes", ConsistOf(summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeContinuous)))
	})

	It("reports the summaries of every glucose type added", func() {
		activateData()
		svcCtx.RestRequest = newRestRequest("[" + glucoseDatum("cbg", time.Now()) + "," + glucoseDatum("smbg", time.Now()) + "]")
		created := expectCreate()

		v1.DataSetsDataCreate(svcCtx)

		Expect(svcCtx.RespondedStatusCode).To(Equal(http.StatusOK))
		Expect(created().Metadata).To(HaveKeyWithValue("summaryTypes", ConsistOf(summaryTypes.AllSummaryTypes)))
	})

	// Nothing is postprocessed when the data added feeds no summary, so no work is created
	DescribeTable("reports nothing when the data added affects no summary",
		func(prepare func() string) {
			svcCtx.RestRequest = newRestRequest("[" + prepare() + "]")

			v1.DataSetsDataCreate(svcCtx)

			Expect(svcCtx.RespondedStatusCode).To(Equal(http.StatusOK))
			Expect(deduplicator.AddDataInvocations).To(Equal(1))
		},
		// The deduplicator of a device data set deactivates the data until the data set is closed
		Entry("as it is not yet active", func() string {
			return glucoseDatum("cbg", time.Now())
		}),
		Entry("as it is older than two years", func() string {
			activateData()
			return glucoseDatum("cbg", time.Now().AddDate(0, -24, -1))
		}),
		Entry("as it is more than a day in the future", func() string {
			activateData()
			return glucoseDatum("smbg", time.Now().Add(25*time.Hour))
		}),
	)

	It("responds successfully when the work cannot be created", func() {
		activateData()
		svcCtx.RestRequest = newRestRequest("[" + glucoseDatum("cbg", time.Now()) + "]")
		workClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errorsTest.RandomError())

		v1.DataSetsDataCreate(svcCtx)

		Expect(svcCtx.RespondedStatusCode).To(Equal(http.StatusOK))
	})

	It("responds with an error and reports nothing when the data is invalid", func() {
		activateData()
		svcCtx.RestRequest = newRestRequest(`[{"type":"cbg","time":"invalid"}]`)

		v1.DataSetsDataCreate(svcCtx)

		Expect(res.WriteHeaderInputs).To(ConsistOf(http.StatusBadRequest))
		Expect(deduplicator.AddDataInvocations).To(Equal(0))
	})
})
