package jotform_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/oura/shopify"
	ouraTest "github.com/tidepool-org/platform/oura/test"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/user"

	consentTest "github.com/tidepool-org/platform/consent/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura/customerio"
	"github.com/tidepool-org/platform/oura/jotform"
	jotformTest "github.com/tidepool-org/platform/oura/jotform/test"
	shopfiyTest "github.com/tidepool-org/platform/oura/shopify/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("WebhookProcessor", func() {
	var (
		ctx       context.Context
		processor *jotform.WebhookProcessor
		logger    log.Logger

		consentCtrl    *gomock.Controller
		consentService *consentTest.MockService

		shopifyCtrl *gomock.Controller
		shopifyClnt *shopfiyTest.MockClient

		userCtrl   *gomock.Controller
		userClient *userTest.MockClient

		appAPIServer    *httptest.Server
		appAPIResponses *ouraTest.StubResponses

		trackAPIServer    *httptest.Server
		trackAPIResponses *ouraTest.StubResponses

		jotformServer    *httptest.Server
		jotformResponses *ouraTest.StubResponses
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = logTest.NewLogger()

		consentCtrl = gomock.NewController(GinkgoT())
		consentService = consentTest.NewMockService(consentCtrl)

		userCtrl = gomock.NewController(GinkgoT())
		userClient = userTest.NewMockClient(userCtrl)

		jotformResponses = ouraTest.NewStubResponses()
		jotformServer = ouraTest.NewStubServer(jotformResponses)
		jotformConfig := jotform.Config{
			BaseURL: jotformServer.URL,
		}

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

		shopifyCtrl = gomock.NewController(GinkgoT())
		shopifyClnt = shopfiyTest.NewMockClient(shopifyCtrl)

		processor, err = jotform.NewWebhookProcessor(jotformConfig, logger, consentService, customerIOClient, userClient, shopifyClnt)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		jotformServer.Close()
		appAPIServer.Close()
		trackAPIServer.Close()
		consentCtrl.Finish()
		userCtrl.Finish()
		shopifyCtrl.Finish()
	})

	Context("ProcessSubmission", func() {
		It("should successfully process an eligible submission and create consent record", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := jotformTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			customer, err := jotformTest.LoadFixture("./test/fixtures/customer.json")
			Expect(err).ToNot(HaveOccurred())

			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/customers/"+userID+"/attributes")},
				ouraTest.Response{StatusCode: http.StatusOK, Body: customer},
			)
			trackAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/api/v1/customers/"+userID+"/events")},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			usr := &user.User{UserID: &userID}
			userClient.EXPECT().Get(gomock.Any(), userID).Return(usr, nil)

			consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("big_data_donation_project")))
					Expect(filter.Version).To(PointTo(Equal(1)))
					Expect(filter.Latest).To(PointTo(Equal(true)))
				}).
				Return(&storeStructuredMongo.ListResult[consent.Record]{
					Count: 0,
				}, nil)

			consentService.EXPECT().CreateConsentRecord(gomock.Any(), userID, gomock.Any()).
				Do(func(ctx context.Context, userID string, create *consent.RecordCreate) {
					Expect(create).ToNot(BeNil())
					Expect(create.Type).To(Equal("big_data_donation_project"))
					Expect(create.Version).To(Equal(1))
					Expect(create.OwnerName).To(Equal("James Jellyfish"))
					Expect(create.AgeGroup).To(Equal(consent.AgeGroupEighteenOrOver))
					Expect(create.GrantorType).To(Equal(consent.GrantorTypeOwner))
				}).
				Return(&consent.Record{
					ID:          "1234567890",
					UserID:      userID,
					Status:      consent.RecordStatusActive,
					AgeGroup:    consent.AgeGroupEighteenOrOver,
					OwnerName:   "James Jellyfish",
					GrantorType: "owner",
					Type:        "big_data_donation_project",
					Version:     1,
				}, nil)

			shopifyClnt.EXPECT().
				CreateDiscountCode(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, input shopify.DiscountCodeInput) error {
					Expect(input.Title).To(Equal("Oura Sizing Kit Discount Code"))
					Expect(len(input.Code)).To(BeNumerically(">=", 12))
					//Expect(input.ProductID).To(Equal("9122899853526"))
					return nil
				})

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process an eligible submission and not attempt to create consent record if one already exists", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := jotformTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			customer, err := jotformTest.LoadFixture("./test/fixtures/customer.json")
			Expect(err).ToNot(HaveOccurred())

			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/customers/"+userID+"/attributes")},
				ouraTest.Response{StatusCode: http.StatusOK, Body: customer},
			)
			trackAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodPost, "/api/v1/customers/"+userID+"/events")},
				ouraTest.Response{StatusCode: http.StatusOK, Body: "{}"},
			)

			usr := &user.User{UserID: &userID}
			userClient.EXPECT().Get(gomock.Any(), userID).Return(usr, nil)

			consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("big_data_donation_project")))
					Expect(filter.Version).To(PointTo(Equal(1)))
					Expect(filter.Latest).To(PointTo(Equal(true)))
				}).
				Return(&storeStructuredMongo.ListResult[consent.Record]{
					Count: 1,
					Data: []consent.Record{{
						ID:          "1234567890",
						UserID:      userID,
						Status:      consent.RecordStatusActive,
						AgeGroup:    consent.AgeGroupEighteenOrOver,
						OwnerName:   "James Jellyfish",
						GrantorType: "owner",
						Type:        "big_data_donation_project",
						Version:     1,
					}},
				}, nil)

			shopifyClnt.EXPECT().
				CreateDiscountCode(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, input shopify.DiscountCodeInput) error {
					Expect(input.Title).To(Equal("Oura Sizing Kit Discount Code"))
					Expect(len(input.Code)).To(BeNumerically(">=", 12))
					//Expect(input.ProductID).To(Equal("9122899853526"))
					return nil
				})

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process an eligible submission and not return error if the customer doesn't exist", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := jotformTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/customers/"+userID+"/attributes")},
				ouraTest.Response{StatusCode: http.StatusNotFound},
			)

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process an eligible submission and not return error if the customer doesn't have the correct participant id", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := jotformTest.LoadFixture("./test/fixtures/submission_participant_mismatch.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/customers/"+userID+"/attributes")},
				ouraTest.Response{StatusCode: http.StatusNotFound},
			)

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process an eligible submission and not return error if the user doesn't exist", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := jotformTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			customer, err := jotformTest.LoadFixture("./test/fixtures/customer.json")
			Expect(err).ToNot(HaveOccurred())

			appAPIResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/customers/"+userID+"/attributes")},
				ouraTest.Response{StatusCode: http.StatusOK, Body: customer},
			)

			userClient.EXPECT().Get(gomock.Any(), userID).Return(nil, nil)

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
