package claims_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	clinic "github.com/tidepool-org/clinic/client"
	confirmationClient "github.com/tidepool-org/hydrophone/client"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	notificationsHistory "github.com/tidepool-org/platform/notifications/history"
	historyTest "github.com/tidepool-org/platform/notifications/history/test"
	"github.com/tidepool-org/platform/notifications/work/claims"
	claimsTest "github.com/tidepool-org/platform/notifications/work/claims/test"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("Processor", func() {
	var ctx context.Context
	var mockController *gomock.Controller
	var mockWorkClient *workTest.MockClient
	var mockClinicClient *clinicsTest.MockClient
	var mockConfirmationClient *claimsTest.MockConfirmationClient
	var mockHistoryRecorder *historyTest.MockRecorder
	var dependencies claims.Dependencies

	// history collects every entry recorded during a spec, in order.
	var history []notificationsHistory.Entry

	eventTypes := func() []string {
		types := make([]string, 0, len(history))
		for _, entry := range history {
			types = append(types, entry.EventType)
		}
		return types
	}

	BeforeEach(func() {
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		mockController, ctx = gomock.WithContext(ctx, GinkgoT())
		mockWorkClient = workTest.NewMockClient(mockController)
		mockClinicClient = clinicsTest.NewMockClient(mockController)
		mockConfirmationClient = claimsTest.NewMockConfirmationClient(mockController)
		mockHistoryRecorder = historyTest.NewMockRecorder(mockController)
		dependencies = claims.Dependencies{
			Dependencies:       workBase.Dependencies{WorkClient: mockWorkClient},
			ClinicClient:       mockClinicClient,
			ConfirmationClient: mockConfirmationClient,
			HistoryRecorder:    mockHistoryRecorder,
		}

		history = nil
		mockHistoryRecorder.EXPECT().Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, entry notificationsHistory.Entry) error {
				history = append(history, entry)
				return nil
			}).AnyTimes()
	})

	Context("NewProcessor", func() {
		It("returns an error if dependencies is invalid", func() {
			dependencies.ClinicClient = nil
			processor, err := claims.NewProcessor(dependencies)
			Expect(err).To(MatchError("dependencies is invalid; clinic client is missing"))
			Expect(processor).To(BeNil())
		})

		It("returns successfully", func() {
			processor, err := claims.NewProcessor(dependencies)
			Expect(err).ToNot(HaveOccurred())
			Expect(processor).ToNot(BeNil())
		})
	})

	Context("Process", func() {
		var clinicID, userID, email string
		var patient *clinic.PatientV1
		var wrk *work.Work
		var mockProcessingUpdater *workTest.MockProcessingUpdater
		var processor *claims.Processor

		BeforeEach(func() {
			clinicID = "6aaaa3eec8bd4a4d15327f83"
			userID = userTest.RandomUserID()
			email = "custodial@tidepool.org"
			// A custodial patient who hasn't claimed the account yet.
			custodian := map[string]interface{}{}
			patient = &clinic.PatientV1{
				Id:          pointer.FromString(userID),
				Email:       pointer.FromString(email),
				Permissions: &clinic.PatientPermissionsV1{Custodian: &custodian},
			}

			metadata := claims.Metadata{ClinicID: clinicID, UserID: userID}
			wrkCreate, err := claims.NewWorkCreate(metadata)
			Expect(err).ToNot(HaveOccurred())
			wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
			mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)

			processor, err = claims.NewProcessor(dependencies)
			Expect(err).ToNot(HaveOccurred())
		})

		process := func() *work.ProcessResult {
			return processor.Process(ctx, wrk, mockProcessingUpdater)
		}

		expectPatient := func(patient *clinic.PatientV1, err error) {
			mockClinicClient.EXPECT().GetPatient(gomock.Any(), clinicID, userID).
				Return(patient, err)
		}

		expectMarkResent := func(err error) {
			mockClinicClient.EXPECT().
				RecordInvitationResent(gomock.Any(), clinicID, userID).
				Return(err)
		}

		expectResend := func() {
			mockConfirmationClient.EXPECT().
				ResendAccountSignupWithResponse(gomock.Any(), email).
				Return(&confirmationClient.ResendAccountSignupResponse{}, nil)
		}

		It("returns a failing result if the patient can't be fetched", func() {
			testErr := errorsTest.RandomError()
			expectPatient(nil, testErr)

			Expect(process()).To(workTest.MatchFailingProcessResultError(
				MatchError(errors.Wrap(testErr, "unable to get patient").Error())))
			Expect(history).To(BeEmpty())
		})

		It("does nothing more once the patient has claimed the account", func() {
			patient.Permissions = nil
			expectPatient(patient, nil)

			Expect(process()).To(workTest.MatchDeleteProcessResult())
			Expect(eventTypes()).To(Equal([]string{
				notificationsHistory.NotificationConditionsExpired,
			}))
		})

		It("returns a failing result if the reminder can't be sent", func() {
			testErr := errorsTest.RandomError()
			expectPatient(patient, nil)
			mockConfirmationClient.EXPECT().
				ResendAccountSignupWithResponse(gomock.Any(), email).
				Return(nil, testErr)

			expectedErr := errors.Wrap(testErr, "unable to send email for account claim")
			Expect(process()).To(workTest.MatchFailingProcessResultError(
				MatchError(expectedErr.Error())))
			Expect(eventTypes()).To(Equal([]string{
				notificationsHistory.NotificationAttempted,
				notificationsHistory.NotificationEmailError,
			}))
		})

		It("records the invitation re-sent with the clinic service", func() {
			expectPatient(patient, nil)
			expectResend()
			expectMarkResent(nil)

			Expect(process()).To(workTest.MatchDeleteProcessResult())
			Expect(eventTypes()).To(Equal([]string{
				notificationsHistory.NotificationAttempted,
				notificationsHistory.NotificationEmailSent,
			}))
		})

		It("records but doesn't fail on an error from the clinic service", func() {
			testErr := errorsTest.RandomError()
			expectPatient(patient, nil)
			expectResend()
			expectMarkResent(testErr)

			// The email has been sent; retrying would send it again.
			Expect(process()).To(workTest.MatchDeleteProcessResult())
			Expect(eventTypes()).To(Equal([]string{
				notificationsHistory.NotificationAttempted,
				notificationsHistory.NotificationEmailSent,
				notificationsHistory.NotificationGeneralError,
			}))
			Expect(history[2].Error).To(MatchError(
				errors.Wrap(testErr, "unable to record patient invitation re-sent").Error()))
		})
	})
})
