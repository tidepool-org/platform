package event_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/customerio"
	"github.com/tidepool-org/platform/customerio/work/event"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/work"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

const (
	timeout = 5 * time.Second
)

var _ = Describe("Processor", func() {
	var (
		mockController        *gomock.Controller
		mockWorkClient        *workTest.MockClient
		mockProcessingUpdater *workTest.MockProcessingUpdater

		logger log.Logger

		appAPIServer    *httptest.Server
		appAPIResponses *ouraTest.StubResponses

		trackAPIServer    *httptest.Server
		trackAPIResponses *ouraTest.StubResponses

		processor *event.Processor
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockWorkClient = workTest.NewMockClient(mockController)
		mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)

		logger = logTest.NewLogger()

		appAPIResponses = ouraTest.NewStubResponses()
		appAPIServer = ouraTest.NewStubServer(appAPIResponses)

		trackAPIResponses = ouraTest.NewStubResponses()
		trackAPIServer = ouraTest.NewStubServer(trackAPIResponses)

		customerIOConfig := customerio.Config{
			AppAPIBaseURL:   appAPIServer.URL,
			TrackAPIBaseURL: trackAPIServer.URL,
		}
		customerIOClient, err := customerio.NewClient(customerIOConfig, logger)
		Expect(err).ToNot(HaveOccurred())

		dependencies := event.Dependencies{
			Dependencies: workBase.Dependencies{
				WorkClient: mockWorkClient,
			},
			CustomerIOClient: customerIOClient,
		}
		processor, err = event.NewProcessor(dependencies)
		Expect(err).ToNot(HaveOccurred())
		Expect(processor).ToNot(BeNil())
	})

	AfterEach(func() {
		Expect(trackAPIResponses.UnmatchedResponses()).To(Equal(0))
		Expect(appAPIResponses.UnmatchedResponses()).To(Equal(0))

		appAPIServer.Close()
		trackAPIServer.Close()
	})

	Context("", func() {
		It("", func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			ctx = log.NewContextWithLogger(ctx, logger)

			dataSource := dataSourceTest.RandomSource()
			deduplicationTime := event.DeduplicationTimeFromDataSource(*dataSource)
			deduplicationID := event.WorkDeduplicationIDFromDataSource("data_source_state_changed", *dataSource)

			wrk := &work.Work{
				ID: workTest.RandomID(),
				Metadata: storeStructuredMongo.BSONToMap(bson.M{
					"userId":    dataSource.UserID,
					"eventType": "data_source_state_changed",
					"eventData": bson.M{
						"provider_name": dataSource.ProviderName,
						"state":         dataSource.State,
					},
					"eventDeduplicationTime": deduplicationTime,
					"eventDeduplicationId":   deduplicationID,
				}),
			}

			id, err := customerio.CreateUlid(deduplicationTime, deduplicationID)
			Expect(err).ToNot(HaveOccurred())

			trackAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/api/v1/customers/"+dataSource.UserID+"/events"),
					ouraTest.NewRequestJSONBodyMatcher(`{
					  	        "name": "data_source_state_changed",
						        "id": "` + id.String() + `",
						        "data": {
                                    "provider_name": "` + dataSource.ProviderName + `",
                                    "state": "` + dataSource.State + `"
                                }
					        }`),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			result := processor.Process(ctx, wrk, mockProcessingUpdater)
			Expect(result.Result).To(Equal(work.ResultDelete))
		})
	})
})
