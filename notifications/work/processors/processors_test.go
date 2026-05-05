package processors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	mailerTest "github.com/tidepool-org/platform/mailer/test"
	notificationWorkClaimsTest "github.com/tidepool-org/platform/notifications/work/claims/test"
	notificationsWorkProcessors "github.com/tidepool-org/platform/notifications/work/processors"
	userTest "github.com/tidepool-org/platform/user/test"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processors", func() {
	Context("with dependencies", func() {
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockClinicClient *clinicsTest.MockClient
		var mockConfirmationClient *notificationWorkClaimsTest.MockConfirmationClient
		var mockDataSourceClient *dataSourceTest.MockClient
		var mockMailerClient *mailerTest.MockClient
		var mockUserClient *userTest.MockClient
		var dependencies notificationsWorkProcessors.Dependencies

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockClinicClient = clinicsTest.NewMockClient(mockController)
			mockConfirmationClient = notificationWorkClaimsTest.NewMockConfirmationClient(mockController)
			mockDataSourceClient = dataSourceTest.NewMockClient(mockController)
			mockMailerClient = mailerTest.NewMockClient(mockController)
			mockUserClient = userTest.NewMockClient(mockController)
			dependencies = notificationsWorkProcessors.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				ClinicClient:       mockClinicClient,
				ConfirmationClient: mockConfirmationClient,
				DataSourceClient:   mockDataSourceClient,
				MailerClient:       mockMailerClient,
				UserClient:         mockUserClient,
			}
		})

		Context("Dependencies", func() {
			Context("Validate", func() {
				It("returns an error if work client is missing", func() {
					dependencies.WorkClient = nil
					Expect(dependencies.Validate()).To(MatchError("work client is missing"))
				})

				It("returns an error if clinic client is missing", func() {
					dependencies.ClinicClient = nil
					Expect(dependencies.Validate()).To(MatchError("clinic client is missing"))
				})

				It("returns an error if confirmation client is missing", func() {
					dependencies.ConfirmationClient = nil
					Expect(dependencies.Validate()).To(MatchError("confirmation client is missing"))
				})

				It("returns an error if data source client is missing", func() {
					dependencies.DataSourceClient = nil
					Expect(dependencies.Validate()).To(MatchError("data source client is missing"))
				})

				It("returns an error if mailer client is missing", func() {
					dependencies.MailerClient = nil
					Expect(dependencies.Validate()).To(MatchError("mailer client is missing"))
				})

				It("returns an error if user client is missing", func() {
					dependencies.UserClient = nil
					Expect(dependencies.Validate()).To(MatchError("user client is missing"))
				})

				It("returns successfully", func() {
					Expect(dependencies.Validate()).To(Succeed())
				})
			})
		})

		Context("NewProcessorFactories", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processorFactories, err := notificationsWorkProcessors.NewProcessorFactories(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processorFactories).To(BeNil())
			})

			It("returns successfully", func() {
				processorFactories, err := notificationsWorkProcessors.NewProcessorFactories(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processorFactories).To(HaveLen(3))
			})
		})
	})
})
