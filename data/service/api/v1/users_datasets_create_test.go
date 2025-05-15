package v1_test

import (
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/data"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataDeduplicatorTest "github.com/tidepool-org/platform/data/deduplicator/test"
	v1 "github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/data/service/api/v1/mocks"
	dataSourceService "github.com/tidepool-org/platform/data/source/service"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreTest "github.com/tidepool-org/platform/data/store/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metric"
	metricTest "github.com/tidepool-org/platform/metric/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/summary"
	"github.com/tidepool-org/platform/summary/reporters"
	synctaskStore "github.com/tidepool-org/platform/synctask/store"
	"github.com/tidepool-org/platform/twiist"
)

var _ = Describe("UsersDataSetsCreate", func() {
	Context("CreatedUserID", func() {
		It("does set the CreatedUserID if the auth details are for a user", func() {
			dataServiceContext := newMockDataServiceContext(GinkgoT())
			dataServiceContext.AuthDetails = request.NewAuthDetails(request.MethodAccessToken, "test-auth-details-user-id", "token")
			dataServiceContext.DataSetTester = func(t testingT, dataSet *data.DataSet) {
				Expect(dataSet.CreatedUserID).ToNot(BeNil())
				Expect(*dataSet.CreatedUserID).To(Equal("test-deduplicator-created-user-id"))
			}

			v1.UsersDataSetsCreate(dataServiceContext)

			Expect(dataServiceContext.dataDeduplicatorFactory.NewInputs).To(HaveLen(1))
			Expect(dataServiceContext.dataDeduplicatorFactory.NewInputs[0]).ToNot(BeNil())
			Expect(dataServiceContext.dataDeduplicatorFactory.NewInputs[0].DataSet.CreatedUserID).ToNot(BeNil())
			Expect(*dataServiceContext.dataDeduplicatorFactory.NewInputs[0].DataSet.CreatedUserID).To(Equal("test-auth-details-user-id"))

			dataServiceContext.dataDeduplicator.AssertOutputsEmpty()
			dataServiceContext.dataDeduplicatorFactory.AssertOutputsEmpty()
		})

		It("does not set the CreatedUserID if the auth details are not for a user", func() {
			dataServiceContext := newMockDataServiceContext(GinkgoT())
			dataServiceContext.AuthDetails = request.NewAuthDetails(request.MethodServiceSecret, "", "token")
			dataServiceContext.DataSetTester = func(t testingT, dataSet *data.DataSet) {
				Expect(dataSet.CreatedUserID).ToNot(BeNil())
				Expect(*dataSet.CreatedUserID).To(Equal("test-deduplicator-created-user-id"))
			}

			v1.UsersDataSetsCreate(dataServiceContext)

			Expect(dataServiceContext.dataDeduplicatorFactory.NewInputs).To(HaveLen(1))
			Expect(dataServiceContext.dataDeduplicatorFactory.NewInputs[0]).ToNot(BeNil())
			Expect(dataServiceContext.dataDeduplicatorFactory.NewInputs[0].DataSet.CreatedUserID).To(BeNil())

			dataServiceContext.dataDeduplicator.AssertOutputsEmpty()
			dataServiceContext.dataDeduplicatorFactory.AssertOutputsEmpty()
		})
	})
})

type testingT interface {
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

type mockDataServiceContext struct {
	t testingT

	dataDeduplicator        *dataDeduplicatorTest.Deduplicator
	dataDeduplicatorFactory *dataDeduplicatorTest.Factory

	AuthDetails request.AuthDetails

	// DataSetTester tests the resulting data set.
	DataSetTester func(testingT, *data.DataSet)
}

func newMockDataServiceContext(t testingT) *mockDataServiceContext {
	dataSet := dataTest.RandomDataSet()
	dataSet.CreatedUserID = pointer.FromString("test-deduplicator-created-user-id")

	dataDeduplicator := dataDeduplicatorTest.NewDeduplicator()
	dataDeduplicator.OpenOutputs = []dataDeduplicatorTest.OpenOutput{{DataSet: dataSet, Error: nil}}

	dataDeduplicatorFactory := dataDeduplicatorTest.NewFactory()
	dataDeduplicatorFactory.NewOutput = &dataDeduplicatorTest.NewOutput{Deduplicator: dataDeduplicator, Error: nil}

	return &mockDataServiceContext{
		t:                       t,
		dataDeduplicator:        dataDeduplicator,
		dataDeduplicatorFactory: dataDeduplicatorFactory,
	}
}

func (c *mockDataServiceContext) Response() rest.ResponseWriter {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) Request() *rest.Request {
	r, err := http.NewRequest(http.MethodGet, "", nil)
	if err != nil {
		c.t.Fatalf("creating test request: %s", err)
	}

	testLogger := logtest.NewLogger()
	r = r.WithContext(log.NewContextWithLogger(r.Context(), testLogger))
	r = r.WithContext(request.NewContextWithAuthDetails(r.Context(), c.AuthDetails))

	r.Body = io.NopCloser(strings.NewReader(`{}`))

	rr := &rest.Request{
		Request: r,
		PathParams: map[string]string{
			"userId": "test-path-params-user-id",
		},
	}

	return rr
}

func (c *mockDataServiceContext) RespondWithError(err *service.Error) {
	c.t.Errorf("got error: %s", err)
}

func (c *mockDataServiceContext) RespondWithInternalServerFailure(message string, failure ...interface{}) {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) RespondWithStatusAndErrors(statusCode int, errors []*service.Error) {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) RespondWithStatusAndData(statusCode int, object interface{}) {
	dataSet, ok := object.(*data.DataSet)
	if !ok {
		c.t.Errorf("expected data.DataSet response, got %v", object)
	}

	if c.DataSetTester != nil {
		c.DataSetTester(c.t, dataSet)
	}
}

func (c *mockDataServiceContext) AuthClient() auth.Client {
	return authTest.NewClient()
}

func (c *mockDataServiceContext) MetricClient() metric.Client {
	mc := metricTest.NewClient()
	mc.RecordMetricOutputs = []error{nil}
	return mc
}

func (c *mockDataServiceContext) PermissionClient() permission.Client {
	fullPerms := permission.Permissions{
		permission.Custodian: map[string]interface{}{},
		permission.Follow:    map[string]interface{}{},
		permission.Read:      map[string]interface{}{},
		permission.Owner:     map[string]interface{}{},
		permission.Write:     map[string]interface{}{},
	}
	return mocks.NewPermission(nil, fullPerms, nil)
}

func (c *mockDataServiceContext) DataDeduplicatorFactory() deduplicator.Factory {
	return c.dataDeduplicatorFactory
}

func (c *mockDataServiceContext) DataRepository() dataStore.DataRepository {
	r := dataStoreTest.NewDataRepository()
	r.CreateDataSetOutputs = []error{nil}
	return r
}

func (c *mockDataServiceContext) SummaryRepository() dataStore.SummaryRepository {
	panic("not implemented")
}

func (c *mockDataServiceContext) SyncTaskRepository() synctaskStore.SyncTaskRepository {
	panic("not implemented")
}

func (c *mockDataServiceContext) AlertsRepository() alerts.Repository {
	panic("not implemented")
}

func (c *mockDataServiceContext) SummarizerRegistry() *summary.SummarizerRegistry {
	panic("not implemented")
}

func (c *mockDataServiceContext) DataClient() dataClient.Client {
	panic("not implemented")
}

func (c *mockDataServiceContext) ClinicsClient() clinics.Client {
	panic("not implemented")
}

func (c *mockDataServiceContext) DataSourceClient() dataSourceService.Client {
	panic("not implemented")
}

func (c *mockDataServiceContext) SummaryReporter() *reporters.PatientRealtimeDaysReporter {
	panic("not implemented")
}

func (c *mockDataServiceContext) TwiistServiceAccountAuthorizer() twiist.ServiceAccountAuthorizer {
	panic("not implemented")
}
