package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataDeduplicator "github.com/tidepool-org/platform/data/deduplicator/test"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	testDataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED/test"
	"github.com/tidepool-org/platform/metric"
	testMetric "github.com/tidepool-org/platform/metric/test"
	"github.com/tidepool-org/platform/service"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
	testSyncTaskStore "github.com/tidepool-org/platform/synctask/store/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
	testUser "github.com/tidepool-org/platform/user/test"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/service/api/v1")
}

type RespondWithInternalServerFailureInput struct {
	message string
	failure []interface{}
}

type RespondWithStatusAndErrorsInput struct {
	statusCode int
	errors     []*service.Error
}

type RespondWithStatusAndDataInput struct {
	statusCode int
	data       interface{}
}

type TestContext struct {
	*test.Mock
	RespondWithErrorInputs                 []*service.Error
	RespondWithInternalServerFailureInputs []RespondWithInternalServerFailureInput
	RespondWithStatusAndErrorsInputs       []RespondWithStatusAndErrorsInput
	RespondWithStatusAndDataInputs         []RespondWithStatusAndDataInput
	MetricClientImpl                       *testMetric.Client
	UserClientImpl                         *testUser.Client
	DataDeduplicatorFactoryImpl            *testDataDeduplicator.Factory
	DataSessionImpl                        *testDataStoreDEPRECATED.DataSession
	SyncTaskSessionImpl                    *testSyncTaskStore.SyncTaskSession
}

func NewTestContext() *TestContext {
	return &TestContext{
		MetricClientImpl:            testMetric.NewClient(),
		UserClientImpl:              testUser.NewClient(),
		DataDeduplicatorFactoryImpl: testDataDeduplicator.NewFactory(),
		DataSessionImpl:             testDataStoreDEPRECATED.NewDataSession(),
		SyncTaskSessionImpl:         testSyncTaskStore.NewSyncTaskSession(),
	}
}

func (t *TestContext) Response() rest.ResponseWriter {
	panic("Unexpected invocation of Response on TestContext")
}

func (t *TestContext) RespondWithError(err *service.Error) {
	t.RespondWithErrorInputs = append(t.RespondWithErrorInputs, err)
}

func (t *TestContext) RespondWithInternalServerFailure(message string, failure ...interface{}) {
	t.RespondWithInternalServerFailureInputs = append(t.RespondWithInternalServerFailureInputs, RespondWithInternalServerFailureInput{message, failure})
}

func (t *TestContext) RespondWithStatusAndErrors(statusCode int, errors []*service.Error) {
	t.RespondWithStatusAndErrorsInputs = append(t.RespondWithStatusAndErrorsInputs, RespondWithStatusAndErrorsInput{statusCode, errors})
}

func (t *TestContext) RespondWithStatusAndData(statusCode int, data interface{}) {
	t.RespondWithStatusAndDataInputs = append(t.RespondWithStatusAndDataInputs, RespondWithStatusAndDataInput{statusCode, data})
}

func (t *TestContext) MetricClient() metric.Client {
	return t.MetricClientImpl
}

func (t *TestContext) UserClient() user.Client {
	return t.UserClientImpl
}

func (t *TestContext) DataFactory() data.Factory {
	panic("Unexpected invocation of DataFactory on TestContext")
}

func (t *TestContext) DataDeduplicatorFactory() deduplicator.Factory {
	return t.DataDeduplicatorFactoryImpl
}

func (t *TestContext) DataSession() dataStoreDEPRECATED.DataSession {
	return t.DataSessionImpl
}

func (t *TestContext) SyncTaskSession() syncTaskStore.SyncTaskSession {
	return t.SyncTaskSessionImpl
}

func (t *TestContext) Expectations() {
	t.Mock.Expectations()
	t.MetricClientImpl.Expectations()
	t.UserClientImpl.Expectations()
	t.DataDeduplicatorFactoryImpl.Expectations()
	t.DataSessionImpl.Expectations()
	t.SyncTaskSessionImpl.Expectations()
}
