package jotform

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/customerio"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura"
	"github.com/tidepool-org/platform/oura/jotform/store"
	"github.com/tidepool-org/platform/oura/shopify"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

const (
	EligibleField    = "eligible"
	NameField        = "name"
	DateOfBirthField = "dateOfBirth"

	UserIDField        = "userId"
	ParticipantIDField = "participantId"

	DefaultReconcileLimit = 100
	MaxReconcileLimit     = 1000
)

type SubmissionProcessor struct {
	config Config

	logger log.Logger

	jotformClient    Client
	store            store.Store
	consentService   consent.Service
	customerIOClient *customerio.Client
	shopifyClient    shopify.Client
	userClient       user.Client
}

type Config struct {
	BaseURL string `envconfig:"TIDEPOOL_OURA_JOTFORM_BASE_URL"`
	APIKey  string `envconfig:"TIDEPOOL_OURA_JOTFORM_API_KEY"`
	FormID  string `envconfig:"TIDEPOOL_OURA_JOTFORM_FORM_ID"`
}

func NewSubmissionProcessor(config Config, logger log.Logger, consentService consent.Service, customerIOClient *customerio.Client, userClient user.Client, shopifyClient shopify.Client, submissionStore store.Store) (*SubmissionProcessor, error) {
	jotformClient, err := NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create jotform defaultClient")
	}

	return &SubmissionProcessor{
		config: config,

		logger: logger,

		jotformClient:    jotformClient,
		store:            submissionStore,
		consentService:   consentService,
		customerIOClient: customerIOClient,
		shopifyClient:    shopifyClient,
		userClient:       userClient,
	}, nil
}

// Reconcile fetches new submissions from Jotform and processes any that haven't been processed yet
// returns the last processed submission ID and an error if one occurred during processing
func (s *SubmissionProcessor) Reconcile(ctx context.Context, formID string, lastSubmissionID string) (string, error) {
	logger := s.logger.WithField("formId", s.config.FormID)
	logger.Info("Starting Jotform submission reconciliation")

	filter := &SubmissionFilter{
		IDGreaterThan: lastSubmissionID,
		Limit:         DefaultReconcileLimit,
	}

	var result ReconcileResult
	var err error

	for {
		if result.TotalProcessed >= MaxReconcileLimit {
			logger.WithField("limit", MaxReconcileLimit).Warn("Reached maximum reconciliation limit")
			break
		}

		var submissions *FormSubmissionsResponse
		submissions, err = s.jotformClient.ListFormSubmissions(ctx, formID, filter)
		if err != nil {
			err = errors.Wrap(err, "failed to fetch submissions")
			break
		}

		if len(submissions.Content) == 0 {
			break
		}

		for _, content := range submissions.Content {
			submission := &SubmissionResponse{
				ResponseCode: submissions.ResponseCode,
				Content:      content,
			}

			err = s.processSubmission(ctx, submission)
			if err != nil {
				err = errors.Wrapf(err, "failed to reconcile submission %s", submission.Content.ID)
				break
			}
			filter.IDGreaterThan = submission.Content.ID
			result.TotalProcessed++
		}

		if len(submissions.Content) < filter.Limit {
			break
		}
	}

	logger = logger.WithFields(log.Fields{
		"processed": result.TotalProcessed,
		"errors":    result.TotalErrors,
	})
	if err != nil {
		logger = logger.WithError(err)
	}

	logger.Info("Completed Jotform submission reconciliation")

	return filter.IDGreaterThan, err
}

func (s *SubmissionProcessor) ProcessSubmission(ctx context.Context, submissionID string) error {
	submission, err := s.jotformClient.GetSubmission(ctx, submissionID)
	if err != nil {
		return errors.Wrap(err, "failed to get submission")
	}

	return s.processSubmission(ctx, submission)
}

func (s *SubmissionProcessor) processSubmission(ctx context.Context, submission *SubmissionResponse) error {
	if submission == nil {
		return errors.New("submission is missing")
	}
	logger := s.logger.WithField("submission", submission.Content.ID)

	processed, err := s.store.GetProcessedSubmission(ctx, submission.Content.FormID, submission.Content.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get processed submission")
	}
	if processed != nil {
		logger.Debug("submission is already processed")
		return nil
	}

	if submission.Content.Answers == nil {
		logger.Warn("submission has no answers")
		return nil
	}
	identifiers, err := s.validateUser(ctx, submission.Content.ID, submission.Content.Answers)
	if err != nil {
		logger.WithError(err).Warn("user is invalid")
		return nil
	} else if identifiers == nil {
		logger.Warn("invalid user")
		return nil
	}

	if err := s.handleSurveyCompleted(ctx, *identifiers, submission); err != nil {
		return err
	}

	processedSubmission := &store.ProcessedSubmission{
		SubmissionID: submission.Content.ID,
		FormID:       submission.Content.FormID,
		CreatedTime:  time.Now(),
	}
	if err := s.store.SaveProcessedSubmission(ctx, processedSubmission); err != nil {
		logger.WithError(err).Warn("failed to save processed submission")
	}

	return nil
}

// validateUser validates the user id by comparing the participant id from the submission with the participant id from customer.io
// this is required because jotform webhooks are not signed or authenticated
func (s *SubmissionProcessor) validateUser(ctx context.Context, submissionID string, answers Answers) (*customerio.Identifiers, error) {
	logger := s.logger.WithField("submission", submissionID)

	userID := answers.GetAnswerTextByName(UserIDField)
	if userID == "" {
		logger.Debugf("submission has no user id")
		return nil, nil
	}

	participantID := answers.GetAnswerTextByName(ParticipantIDField)
	if participantID == "" {
		logger.Debugf("submission has no participant id")
		return nil, nil
	}

	customer, err := s.customerIOClient.GetCustomer(ctx, userID, customerio.IDTypeUserID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get customer with id %s", userID)
	}

	if customer == nil {
		return nil, errors.Newf("customer with id %s not found", userID)
	}
	if customer.OuraParticipantID != participantID {
		return nil, errors.Newf("participant id mismatch for user with id %s", userID)
	}

	usr, err := s.userClient.Get(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user")
	}
	if usr == nil {
		return nil, errors.New("user not found")
	}

	return &customer.Identifiers, nil
}

func (s *SubmissionProcessor) handleSurveyCompleted(ctx context.Context, identifiers customerio.Identifiers, submission *SubmissionResponse) error {
	surveyCompletedData := oura.OuraEligibilitySurveyCompletedData{
		OuraEligibilitySurveyID:       submission.Content.ID,
		OuraEligibilitySurveyEligible: submission.Content.Answers.GetAnswerTextByName(EligibleField) == "true",
	}

	if surveyCompletedData.OuraEligibilitySurveyEligible {
		if err := s.ensureConsentRecordExists(ctx, identifiers.ID, submission); err != nil {
			s.logger.WithField("submission", submission.Content.ID).WithError(err).Warn("unable to ensure consent record exists")
			return err
		}

		surveyCompletedData.OuraSizingKitDiscountCode = shopify.RandomDiscountCode()
		err := s.shopifyClient.CreateDiscountCode(ctx, shopify.DiscountCodeInput{
			Title:     shopify.OuraSizingKitDiscountCodeTitle,
			Code:      surveyCompletedData.OuraSizingKitDiscountCode,
			ProductID: shopify.OuraSizingKitProductID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create oura ring discount code")
		}
	}

	surveyCompleted := customerio.Event{
		Name: oura.OuraEligibilitySurveyCompletedEventType,
		ID:   surveyCompletedData.OuraEligibilitySurveyID,
		Data: surveyCompletedData,
	}

	err := s.customerIOClient.SendEvent(ctx, identifiers.ID, surveyCompleted)
	if err != nil {
		return errors.Wrap(err, "unable to send sizing kit delivered event")
	}

	return nil
}

func (s *SubmissionProcessor) ensureConsentRecordExists(ctx context.Context, userID string, submission *SubmissionResponse) error {
	logger := s.logger.WithField("submission", submission.Content.ID)

	survey := OuraEligibilitySurvey{
		DateOfBirth: submission.Content.Answers.GetAnswerTextByName(DateOfBirthField),
		Name:        submission.Content.Answers.GetAnswerTextByName(NameField),
	}

	v := validator.New(s.logger)
	survey.Validate(v)
	if err := v.Error(); err != nil {
		logger.WithError(err).Warn("consent survey is invalid")
		return nil
	}

	filter := consent.NewRecordFilter()
	filter.Latest = pointer.FromAny(true)
	filter.Status = pointer.FromAny(consent.RecordStatusActive)
	filter.Type = pointer.FromAny(consent.TypeBigDataDonationProject)
	filter.Version = pointer.FromAny(1)

	pagination := page.NewPagination()

	records, err := s.consentService.ListConsentRecords(ctx, userID, filter, pagination)
	if err != nil {
		return errors.Wrap(err, "unable to list consent records")
	}

	if records.Count > 0 {
		logger.WithField("userId", userID).Info("consent record already exists")
		return nil
	}

	create := consent.NewRecordCreate()
	create.AgeGroup = consent.AgeGroupEighteenOrOver
	create.GrantorType = consent.GrantorTypeOwner
	create.OwnerName = survey.Name
	create.Type = consent.TypeBigDataDonationProject
	create.Version = 1

	_, err = s.consentService.CreateConsentRecord(ctx, userID, create)
	if err != nil {
		return errors.Wrap(err, "unable to create consent record")
	}

	return nil
}

type ReconcileResult struct {
	TotalErrors     int
	TotalProcessed  int
	LastProcessedID string
}
