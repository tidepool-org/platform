package jotform_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/oura/jotform/store"

	"github.com/tidepool-org/platform/customerio"

	"github.com/tidepool-org/platform/oura/shopify"
	ouraTest "github.com/tidepool-org/platform/oura/test"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/page"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/user"

	consentTest "github.com/tidepool-org/platform/consent/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura/jotform"
	shopfiyTest "github.com/tidepool-org/platform/oura/shopify/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("SubmissionProcessor", func() {
	var (
		ctx       context.Context
		processor *jotform.SubmissionProcessor
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

		formID string
	)

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)

		consentCtrl = gomock.NewController(GinkgoT())
		consentService = consentTest.NewMockService(consentCtrl)

		userCtrl = gomock.NewController(GinkgoT())
		userClient = userTest.NewMockClient(userCtrl)

		jotformResponses = ouraTest.NewStubResponses()
		jotformServer = ouraTest.NewStubServer(jotformResponses)

		formID = strconv.Itoa(rand.Intn(1000000000))
		jotformConfig := jotform.Config{
			BaseURL: jotformServer.URL,
			FormID:  formID,
			Enabled: true,
			TeamID:  "test-team",
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

		processor, err = jotform.NewSubmissionProcessor(jotformConfig, logger, consentService, customerIOClient, userClient, shopifyClnt, GetSuiteStore())
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_, err := mongoStore.GetCollection(store.CollectionName).DeleteMany(context.Background(), bson.M{})
		Expect(err).ToNot(HaveOccurred())

		jotformServer.Close()
		appAPIServer.Close()
		trackAPIServer.Close()
		consentCtrl.Finish()
		userCtrl.Finish()
		shopifyCtrl.Finish()
	})

	Context("ProcessSubmission", func() {
		It("should successfully process an eligible submission and create consent records", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := ouraTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			customer, err := ouraTest.LoadFixture("./test/fixtures/customer.json")
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

			// First call checks for RIPPLE
			consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("ripple")))
					Expect(filter.Latest).To(PointTo(Equal(true)))
				}).
				Return(&storeStructuredMongo.ListResult[consent.Record]{
					Count: 0,
				}, nil)

			// Second call checks for the latest active BDDP
			consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("big_data_donation_project")))
					Expect(filter.Latest).To(PointTo(Equal(true)))
					Expect(filter.Status).To(PointTo(Equal(consent.RecordStatusActive)))
				}).
				Return(&storeStructuredMongo.ListResult[consent.Record]{
					Count: 0,
				}, nil)

			consentService.EXPECT().ListConsents(ctx, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, filter *consent.Filter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("big_data_donation_project")))
					Expect(filter.Latest).To(PointTo(Equal(true)))
				}).
				Return(&storeStructuredMongo.ListResult[consent.Consent]{
					Count: 1,
					Data: []consent.Consent{{
						Type:    "big_data_donation_project",
						Version: 2,
					}},
				}, nil)

			consentService.EXPECT().CreateConsentRecords(gomock.Any(), userID, gomock.Any()).
				Do(func(ctx context.Context, userID string, creates []*consent.RecordCreate) {
					Expect(creates).To(HaveLen(2))
					Expect(creates[0].Type).To(Equal("big_data_donation_project"))
					Expect(creates[0].Version).To(Equal(2))
					Expect(creates[0].OwnerName).To(Equal("James Jellyfish"))
					Expect(creates[0].AgeGroup).To(Equal(consent.AgeGroupEighteenOrOver))
					Expect(creates[0].GrantorType).To(Equal(consent.GrantorTypeOwner))
					Expect(creates[1].Type).To(Equal("ripple"))
					Expect(creates[1].Version).To(Equal(1))
					Expect(creates[1].OwnerName).To(Equal("James Jellyfish"))
					Expect(creates[1].AgeGroup).To(Equal(consent.AgeGroupEighteenOrOver))
					Expect(creates[1].GrantorType).To(Equal(consent.GrantorTypeOwner))
				}).
				Return([]*consent.Record{
					{
						ID:          "1234567890",
						UserID:      userID,
						Status:      consent.RecordStatusActive,
						AgeGroup:    consent.AgeGroupEighteenOrOver,
						OwnerName:   "James Jellyfish",
						GrantorType: "owner",
						Type:        "big_data_donation_project",
						Version:     2,
					},
					{
						ID:          "1234567891",
						UserID:      userID,
						Status:      consent.RecordStatusActive,
						AgeGroup:    consent.AgeGroupEighteenOrOver,
						OwnerName:   "James Jellyfish",
						GrantorType: "owner",
						Type:        "ripple",
						Version:     1,
					},
				}, nil)

			shopifyClnt.EXPECT().
				CreateDiscountCode(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, input shopify.DiscountCodeInput) error {
					Expect(input.Title).To(Equal("Oura Sizing Kit Discount Code"))
					Expect(len(input.Code)).To(BeNumerically(">=", 12))
					Expect(input.ProductID).To(Equal("9122899853526"))
					return nil
				})

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not create consent records if RIPPLE already exists", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := ouraTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			customer, err := ouraTest.LoadFixture("./test/fixtures/customer.json")
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

			// RIPPLE already exists - no further consent calls expected
			consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("ripple")))
					Expect(filter.Latest).To(PointTo(Equal(true)))
				}).
				Return(&storeStructuredMongo.ListResult[consent.Record]{
					Count: 1,
					Data: []consent.Record{{
						ID:          "1234567891",
						UserID:      userID,
						Status:      consent.RecordStatusActive,
						AgeGroup:    consent.AgeGroupEighteenOrOver,
						OwnerName:   "James Jellyfish",
						GrantorType: "owner",
						Type:        "ripple",
						Version:     1,
					}},
				}, nil)

			shopifyClnt.EXPECT().
				CreateDiscountCode(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, input shopify.DiscountCodeInput) error {
					Expect(input.Title).To(Equal("Oura Sizing Kit Discount Code"))
					Expect(len(input.Code)).To(BeNumerically(">=", 12))
					return nil
				})

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create only RIPPLE when BDDP v2 already exists", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := ouraTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			customer, err := ouraTest.LoadFixture("./test/fixtures/customer.json")
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

			// First call: no RIPPLE
			consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("ripple")))
				}).
				Return(&storeStructuredMongo.ListResult[consent.Record]{Count: 0}, nil)

			// Second call checks for the latest active BDDP
			consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
					Expect(filter.Type).To(PointTo(Equal("big_data_donation_project")))
					Expect(filter.Latest).To(PointTo(Equal(true)))
					Expect(filter.Status).To(PointTo(Equal(consent.RecordStatusActive)))
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
						Version:     2,
					}},
				}, nil)

			// Should create only RIPPLE
			consentService.EXPECT().CreateConsentRecords(gomock.Any(), userID, gomock.Any()).
				Do(func(ctx context.Context, userID string, creates []*consent.RecordCreate) {
					Expect(creates).To(HaveLen(1))
					Expect(creates[0].Type).To(Equal("ripple"))
					Expect(creates[0].Version).To(Equal(1))
					Expect(creates[0].OwnerName).To(Equal("James Jellyfish"))
					Expect(creates[0].AgeGroup).To(Equal(consent.AgeGroupEighteenOrOver))
					Expect(creates[0].GrantorType).To(Equal(consent.GrantorTypeOwner))
				}).
				Return([]*consent.Record{
					{
						ID:          "1234567891",
						UserID:      userID,
						Status:      consent.RecordStatusActive,
						AgeGroup:    consent.AgeGroupEighteenOrOver,
						OwnerName:   "James Jellyfish",
						GrantorType: "owner",
						Type:        "ripple",
						Version:     1,
					},
				}, nil)

			shopifyClnt.EXPECT().
				CreateDiscountCode(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, input shopify.DiscountCodeInput) error {
					Expect(input.Title).To(Equal("Oura Sizing Kit Discount Code"))
					Expect(len(input.Code)).To(BeNumerically(">=", 12))
					return nil
				})

			err = processor.ProcessSubmission(ctx, submissionID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should successfully process an eligible submission and not return error if the customer doesn't exist", func() {
			submissionID := "6410095903544943563"
			userID := "1aacb960-430c-4081-8b3b-a32688807dc5"

			submission, err := ouraTest.LoadFixture("./test/fixtures/submission.json")
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

			submission, err := ouraTest.LoadFixture("./test/fixtures/submission_participant_mismatch.json")
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

			submission, err := ouraTest.LoadFixture("./test/fixtures/submission.json")
			Expect(err).ToNot(HaveOccurred())

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, "/v1/submission/"+submissionID)},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submission},
			)

			customer, err := ouraTest.LoadFixture("./test/fixtures/customer.json")
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

	Context("Reconcile", func() {
		type expected struct {
			SubmissionID    string
			UserID          string
			Name            string
			CustomerFixture string
		}

		var expectedSubmissions []expected
		var lastSubmissionID string

		BeforeEach(func() {
			expectedSubmissions = []expected{
				{SubmissionID: "6410095903544943563", UserID: "1aacb960-430c-4081-8b3b-a32688807dc5", Name: "James Jellyfish", CustomerFixture: "./test/fixtures/customer.json"},
				{SubmissionID: "6441197313548135134", UserID: "db3fcb48-1da5-4a4a-aa92-14564a5b8fea", Name: "Jill Jellyfish", CustomerFixture: "./test/fixtures/customer_jill.json"},
			}

			lastSubmissionID = "6410095903544943561"

			submissions, err := ouraTest.LoadFixture("./test/fixtures/submissions.json")
			Expect(err).ToNot(HaveOccurred())

			limit := 100
			filter := url.QueryEscape(fmt.Sprintf(`{"id:gt":"%s"}`, lastSubmissionID))
			orderBy := "-id"

			jotformResponses.AddResponse(
				[]ouraTest.RequestMatcher{
					ouraTest.NewRequestMethodAndPathMatcher(http.MethodGet, fmt.Sprintf("/v1/form/%s/submissions", formID)),
					ouraTest.NewRequestQueryMatcher(fmt.Sprintf("filter=%s&limit=%d&orderby=%s", filter, limit, orderBy)),
				},
				ouraTest.Response{StatusCode: http.StatusOK, Body: submissions},
			)

			for _, expected := range expectedSubmissions {
				userID := expected.UserID
				name := expected.Name
				customer, err := ouraTest.LoadFixture(expected.CustomerFixture)
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

				// First call: check for RIPPLE
				consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
						Expect(filter.Type).To(PointTo(Equal("ripple")))
						Expect(filter.Latest).To(PointTo(Equal(true)))
					}).
					Return(&storeStructuredMongo.ListResult[consent.Record]{
						Count: 0,
					}, nil)

				// Second call checks for the latest active BDDP
				consentService.EXPECT().ListConsentRecords(gomock.Any(), userID, gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, userID string, filter *consent.RecordFilter, pagination *page.Pagination) {
						Expect(filter.Type).To(PointTo(Equal("big_data_donation_project")))
						Expect(filter.Latest).To(PointTo(Equal(true)))
						Expect(filter.Status).To(PointTo(Equal(consent.RecordStatusActive)))
					}).
					Return(&storeStructuredMongo.ListResult[consent.Record]{
						Count: 0,
					}, nil)

				consentService.EXPECT().ListConsents(ctx, gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, filter *consent.Filter, pagination *page.Pagination) {
						Expect(filter.Type).To(PointTo(Equal("big_data_donation_project")))
						Expect(filter.Latest).To(PointTo(Equal(true)))
					}).
					Return(&storeStructuredMongo.ListResult[consent.Consent]{
						Count: 1,
						Data: []consent.Consent{{
							Type:    "big_data_donation_project",
							Version: 2,
						}},
					}, nil)

				consentService.EXPECT().CreateConsentRecords(gomock.Any(), userID, gomock.Any()).
					Do(func(ctx context.Context, userID string, creates []*consent.RecordCreate) {
						Expect(creates).To(HaveLen(2))
						Expect(creates[0].Type).To(Equal("big_data_donation_project"))
						Expect(creates[0].Version).To(Equal(2))
						Expect(creates[0].OwnerName).To(Equal(name))
						Expect(creates[0].AgeGroup).To(Equal(consent.AgeGroupEighteenOrOver))
						Expect(creates[0].GrantorType).To(Equal(consent.GrantorTypeOwner))
						Expect(creates[1].Type).To(Equal("ripple"))
						Expect(creates[1].Version).To(Equal(1))
						Expect(creates[1].OwnerName).To(Equal(name))
					}).
					Return([]*consent.Record{
						{
							ID:          "1234567890",
							UserID:      userID,
							Status:      consent.RecordStatusActive,
							AgeGroup:    consent.AgeGroupEighteenOrOver,
							OwnerName:   name,
							GrantorType: "owner",
							Type:        "big_data_donation_project",
							Version:     2,
						},
						{
							ID:          "1234567891",
							UserID:      userID,
							Status:      consent.RecordStatusActive,
							AgeGroup:    consent.AgeGroupEighteenOrOver,
							OwnerName:   name,
							GrantorType: "owner",
							Type:        "ripple",
							Version:     1,
						},
					}, nil)

				shopifyClnt.EXPECT().
					CreateDiscountCode(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, input shopify.DiscountCodeInput) error {
						Expect(input.Title).To(Equal("Oura Sizing Kit Discount Code"))
						Expect(len(input.Code)).To(BeNumerically(">=", 12))
						//Expect(input.ProductID).To(Equal("9122899853526"))
						return nil
					})

			}
		})

		It("should process all returned submissions", func() {
			result, err := processor.Reconcile(ctx, lastSubmissionID)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.TotalProcessed).To(Equal(len(expectedSubmissions)))
			Expect(result.LastProcessedID).To(Equal(expectedSubmissions[len(expectedSubmissions)-1].SubmissionID))
		})

		It("should be idempotent", func() {
			count := 3

			// The user service mock will fail if it's invoked more than once. We are using this fact to verify that the processor is idempotent.
			for range count {
				result, err := processor.Reconcile(ctx, lastSubmissionID)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.TotalProcessed).To(Equal(len(expectedSubmissions)))
				Expect(result.LastProcessedID).To(Equal(expectedSubmissions[len(expectedSubmissions)-1].SubmissionID))
			}
		})
	})
})
