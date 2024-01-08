package v1_test

import (
	"io"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	. "github.com/onsi/ginkgo/v2"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	deduplicatortest "github.com/tidepool-org/platform/data/deduplicator/test"
	v1 "github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/data/service/api/v1/mocks"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataStore "github.com/tidepool-org/platform/data/store"
	datatest "github.com/tidepool-org/platform/data/store/test"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/types/upload"
	uploadtest "github.com/tidepool-org/platform/data/types/upload/test"
	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metric"
	metrictest "github.com/tidepool-org/platform/metric/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/service"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
)

var _ = Describe("UsersDataSetsCreate", func() {
	It("sets the CreatedUserID", func() {
		dataServiceContext := newMockDataServiceContext(GinkgoT())
		dataServiceContext.UploadTester = func(t testingT, up *upload.Upload) {
			exp := "testuser001"
			if up.CreatedUserID == nil {
				t.Fatalf("expected %q, got %v", exp, up.CreatedUserID)
			}
			if *up.CreatedUserID != exp {
				t.Errorf("expected %q, got %q", exp, *up.CreatedUserID)
			}
		}
		v1.UsersDataSetsCreate(dataServiceContext)
	})
})

type testingT interface {
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

type mockDataServiceContext struct {
	t testingT
	// UploadTester tests the resulting upload.
	UploadTester func(testingT, *upload.Upload)
}

func newMockDataServiceContext(t testingT) *mockDataServiceContext {
	return &mockDataServiceContext{
		t: t,
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
	authDetails := request.NewAuthDetails("method", "test", "token")
	r = r.WithContext(request.NewContextWithAuthDetails(r.Context(), authDetails))

	r.Body = io.NopCloser(strings.NewReader(`{}`))

	rr := &rest.Request{
		Request: r,
		PathParams: map[string]string{
			"userId": "test",
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

func (c *mockDataServiceContext) RespondWithStatusAndData(statusCode int, data interface{}) {
	up, ok := data.(*upload.Upload)
	if !ok {
		c.t.Errorf("expected upload.Upload response, got %v", data)
	}

	if c.UploadTester != nil {
		c.UploadTester(c.t, up)
	}
}

func (c *mockDataServiceContext) AuthClient() auth.Client {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) MetricClient() metric.Client {
	mc := metrictest.NewClient()
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
	d := deduplicatortest.NewDeduplicator()
	up := uploadtest.RandomUpload()
	up.CreatedUserID = pointer.FromString("testuser001")
	d.OpenOutputs = []deduplicatortest.OpenOutput{
		{
			DataSet: up,
			Error:   nil,
		},
	}
	f := deduplicatortest.NewFactory()
	f.NewOutputs = []deduplicatortest.NewOutput{
		{
			Deduplicator: d,
			Error:        nil,
		},
	}
	return f
}

func (c *mockDataServiceContext) DataRepository() dataStore.DataRepository {
	r := datatest.NewDataRepository()
	r.CreateDataSetOutputs = []error{nil}
	return r
}

func (c *mockDataServiceContext) SummaryRepository() dataStore.SummaryRepository {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) SyncTaskRepository() syncTaskStore.SyncTaskRepository {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) AlertsRepository() alerts.Repository {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) SummarizerRegistry() *summary.SummarizerRegistry {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) DataClient() dataClient.Client {
	panic("not implemented") // TODO: Implement
}

func (c *mockDataServiceContext) DataSourceClient() dataSource.Client {
	panic("not implemented") // TODO: Implement
}
