package v1_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura"
	ouraServiceApiV1 "github.com/tidepool-org/platform/oura/service/api/v1"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("v1", func() {
	var (
		ctx            context.Context
		mockController *gomock.Controller
		mockAuthClient *authTest.MockClient
		mockOuraClient *ouraTest.MockClient
		mockWorkClient *workTest.MockClient
		dependencies   ouraServiceApiV1.Dependencies
	)

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		mockController, ctx = gomock.WithContext(ctx, GinkgoT())
		mockAuthClient = authTest.NewMockClient(mockController)
		mockOuraClient = ouraTest.NewMockClient(mockController)
		mockWorkClient = workTest.NewMockClient(mockController)
		dependencies = ouraServiceApiV1.Dependencies{
			AuthClient: mockAuthClient,
			OuraClient: mockOuraClient,
			WorkClient: mockWorkClient,
		}
	})

	Context("Dependencies", func() {
		Context("Validate", func() {
			It("returns an error if auth client is missing", func() {
				dependencies.AuthClient = nil
				Expect(dependencies.Validate()).To(MatchError("auth client is missing"))
			})

			It("returns an error if oura client is missing", func() {
				dependencies.OuraClient = nil
				Expect(dependencies.Validate()).To(MatchError("oura client is missing"))
			})

			It("returns an error if work client is missing", func() {
				dependencies.WorkClient = nil
				Expect(dependencies.Validate()).To(MatchError("work client is missing"))
			})

			It("returns successfully", func() {
				Expect(dependencies.Validate()).To(Succeed())
			})
		})
	})

	Context("NewRouter", func() {
		It("returns an error if dependencies is invalid", func() {
			dependencies.AuthClient = nil
			router, err := ouraServiceApiV1.NewRouter(dependencies)
			Expect(err).To(MatchError("dependencies is invalid; auth client is missing"))
			Expect(router).To(BeNil())
		})

		It("returns successfully", func() {
			router, err := ouraServiceApiV1.NewRouter(dependencies)
			Expect(err).ToNot(HaveOccurred())
			Expect(router).ToNot(BeNil())
		})

		Context("with router", func() {
			var router *ouraServiceApiV1.Router

			BeforeEach(func() {
				var err error
				router, err = ouraServiceApiV1.NewRouter(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(router).ToNot(BeNil())
			})

			Context("Routes", func() {
				It("returns the expected routes", func() {
					Expect(router.Routes()).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"HttpMethod": Equal("GET"),
							"PathExp":    Equal(oura.PartnerPathPrefix + ouraWebhook.EventPath),
							"Func":       Not(BeNil()),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"HttpMethod": Equal("POST"),
							"PathExp":    Equal(oura.PartnerPathPrefix + ouraWebhook.EventPath),
							"Func":       Not(BeNil()),
						})),
					))
				})
			})
		})
	})

	Context("Routes", func() {
		It("returns the expected routes", func() {
			Expect(ouraServiceApiV1.Routes()).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Method":  Equal("GET"),
					"Path":    Equal(oura.PartnerPathPrefix + ouraWebhook.EventPath),
					"Handler": Not(BeNil()),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Method":  Equal("POST"),
					"Path":    Equal(oura.PartnerPathPrefix + ouraWebhook.EventPath),
					"Handler": Not(BeNil()),
				}),
			))
		})
	})
})
