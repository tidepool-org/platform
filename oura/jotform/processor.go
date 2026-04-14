package jotform

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/store/structured/mongo"

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
	Enabled bool   `envconfig:"TIDEPOOL_OURA_JOTFORM_ENABLED"`
	BaseURL string `envconfig:"TIDEPOOL_OURA_JOTFORM_BASE_URL" default:"https://api.jotform.com"`
	APIKey  string `envconfig:"TIDEPOOL_OURA_JOTFORM_API_KEY"`
	FormID  string `envconfig:"TIDEPOOL_OURA_JOTFORM_FORM_ID"`
	TeamID  string `envconfig:"TIDEPOOL_OURA_JOTFORM_TEAM_ID"`
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

func (s *SubmissionProcessor) Reconcile(ctx context.Context, lastSubmissionID string) (ReconcileResult, error) {
	if !s.config.Enabled {
		s.logger.Debug("jotform reconcile was called, but jotform integration is not enabled")
		return ReconcileResult{
			LastProcessedID: lastSubmissionID,
		}, nil
	}
	return s.reconcile(ctx, s.config.FormID, lastSubmissionID)
}

// Reconcile fetches new submissions from Jotform and processes any that haven't been processed yet
// returns the last processed submission ID and an error if one occurred during processing
func (s *SubmissionProcessor) reconcile(ctx context.Context, formID string, lastSubmissionID string) (ReconcileResult, error) {
	logger := s.logger.WithField("formId", s.config.FormID)
	logger.Info("Starting Jotform submission reconciliation")

	result := ReconcileResult{
		LastProcessedID: lastSubmissionID,
	}

	for {
		if result.TotalProcessed >= MaxReconcileLimit {
			logger.WithField("limit", MaxReconcileLimit).Warn("Reached maximum reconciliation limit")
			return result, nil
		}

		filter := &SubmissionFilter{
			IDGreaterThan: result.LastProcessedID,
			Limit:         DefaultReconcileLimit,
		}
		submissions, err := s.jotformClient.ListFormSubmissions(ctx, formID, filter)
		if err != nil {
			return result, errors.Wrap(err, "failed to fetch submissions")
		}

		for _, content := range submissions.Content {
			submission := &SubmissionResponse{
				ResponseCode: submissions.ResponseCode,
				Content:      content,
			}

			err = s.processSubmission(ctx, submission)
			if err != nil {
				return result, errors.Wrapf(err, "failed to reconcile submission %s", submission.Content.ID)
			}
			result.LastProcessedID = submission.Content.ID
			result.TotalProcessed++
		}

		if len(submissions.Content) < filter.Limit {
			break
		}
	}

	return result, nil
}

func (s *SubmissionProcessor) ProcessSubmission(ctx context.Context, submissionID string) error {
	if !s.config.Enabled {
		s.logger.Debug("jotform process submission was called, but jotform integration is not enabled")
		return nil
	}

	submission, err := s.jotformClient.GetSubmission(ctx, submissionID)
	if err != nil {
		return errors.Wrap(err, "failed to get submission")
	}

	err = s.processSubmission(ctx, submission)
	if err != nil {
		s.logger.WithField("submissionId", submissionID).WithError(err).Warn("failed to process submission")
		return err
	}

	return nil
}

func (s *SubmissionProcessor) processSubmission(ctx context.Context, submission *SubmissionResponse) error {
	if submission == nil {
		return errors.New("submission is missing")
	}

	logger := s.logger.WithField("submissionId", submission.Content.ID)

	processed, err := s.store.GetProcessedSubmission(ctx, submission.Content.FormID, submission.Content.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get processed submission")
	}
	if processed != nil {
		logger.Info("submission is already processed")
		return nil
	}

	if submission.Content.Answers == nil {
		logger.Warn("submission has no answers")
		return nil
	}
	customer, err := s.validateUser(ctx, submission.Content.ID, submission.Content.Answers)
	if err != nil {
		return errors.Wrap(err, "unable to validate user")
	} else if customer == nil {
		logger.Warn("invalid user")
		return nil
	}

	if err := s.handleSurveyCompleted(ctx, *customer, submission); err != nil {
		return err
	}

	if err := s.saveProcessedSubmission(ctx, submission); err != nil {
		return errors.Wrap(err, "failed to save processed submission")
	}

	return nil
}

// validateUser validates the user id by comparing the participant id from the submission with the participant id from customer.io
// this is required because jotform webhooks are not signed or authenticated
func (s *SubmissionProcessor) validateUser(ctx context.Context, submissionID string, answers Answers) (*customerio.Customer, error) {
	logger := s.logger.WithField("submissionId", submissionID)

	userID := answers.GetAnswerTextByName(UserIDField)
	if userID == "" {
		logger.Info("submission has no user id")
		return nil, nil
	}
	logger = logger.WithField("userId", userID)

	participantID := answers.GetAnswerTextByName(ParticipantIDField)
	if participantID == "" {
		logger.Info("submission has no participant id")
		return nil, nil
	}
	logger = logger.WithField("submissionParticipantId", participantID)

	customer, err := s.customerIOClient.GetCustomer(ctx, userID, customerio.IDTypeUserID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get customer with id %s", userID)
	}

	if customer == nil {
		logger.Warnf("no matching customer found for user id")
		return nil, nil
	}
	if customer.OuraParticipantID != participantID {
		logger.
			WithField("customerParticipantId", customer.OuraParticipantID).
			Warnf("submission participant id does not match customer participant id")
		return nil, nil
	}

	usr, err := s.userClient.Get(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user")
	}
	if usr == nil {
		logger.Warnf("user not found")
		return nil, nil
	}

	return customer, nil
}

func (s *SubmissionProcessor) handleSurveyCompleted(ctx context.Context, customer customerio.Customer, submission *SubmissionResponse) error {
	logger := s.logger.WithField("submission", submission.Content.ID)

	surveyCompletedData := oura.OuraEligibilitySurveyCompletedData{
		OuraEligibilitySurveyID:       submission.Content.ID,
		OuraEligibilitySurveyEligible: submission.Content.Answers.GetAnswerTextByName(EligibleField) == "true",
	}

	v := validator.New(s.logger)
	survey := OuraEligibilitySurvey{
		DateOfBirth: submission.Content.Answers.GetAnswerTextByName(DateOfBirthField),
		Name:        submission.Content.Answers.GetAnswerTextByName(NameField),
	}

	survey.Validate(v)
	if err := v.Error(); err != nil {
		surveyCompletedData.OuraEligibilitySurveyEligible = false
		logger.WithError(err).Warn("consent survey is invalid")
	}

	if surveyCompletedData.OuraEligibilitySurveyEligible {
		if err := s.ensureConsentRecordExists(ctx, customer.Identifiers.ID, submission, survey); err != nil {
			s.logger.WithError(err).Warn("unable to ensure consent record exists")
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

	surveyCompleted := &customerio.Event{
		Name: oura.OuraEligibilitySurveyCompletedEventType,
		Data: surveyCompletedData,
	}

	// Use oura participant id as deduplication attribute to prevent multiple submissions
	err := surveyCompleted.SetDeduplicationID(nil, customer.OuraParticipantID)
	if err != nil {
		return errors.Wrap(err, "unable to set event id")
	}

	err = s.customerIOClient.SendEvent(ctx, customer.Identifiers.ID, surveyCompleted)
	if err != nil {
		return errors.Wrap(err, "unable to send sizing kit delivered event")
	}

	return nil
}

func (s *SubmissionProcessor) ensureConsentRecordExists(ctx context.Context, userID string, submission *SubmissionResponse, survey OuraEligibilitySurvey) error {
	logger := s.logger.WithField("submission", submission.Content.ID)

	creates := make([]*consent.RecordCreate, 0)

	rippleRecord, err := s.consentService.GetActiveConsentRecord(ctx, userID, consent.TypeRipple)
	if err != nil {
		return errors.Wrap(err, "unable to get active ripple consent record")
	}
	if rippleRecord != nil && rippleRecord.Version >= consent.RippleVersionForJotform {
		// If RIPPLE already exists, BDDP must also exist - nothing to do
		logger.WithField("userId", userID).Info("ripple consent record already exists")
		return nil
	}

	bddpRecord, err := s.consentService.GetActiveConsentRecord(ctx, userID, consent.TypeBigDataDonationProject)
	if err != nil {
		return errors.Wrap(err, "unable to get active bddp consent record")
	}
	if bddpRecord == nil || bddpRecord.Version < consent.MinimumBDDPVersionForRipple {
		latestBDDP, err := s.consentService.ListConsents(ctx, &consent.Filter{
			Type:   pointer.FromAny(consent.TypeBigDataDonationProject),
			Latest: pointer.FromAny(true),
		}, page.NewPaginationMinimum())
		if err != nil || latestBDDP.Count == 0 {
			return errors.Wrap(err, "unable to find latest BDDP version")
		}

		bddpCreate := consent.NewRecordCreate()
		bddpCreate.AgeGroup = consent.AgeGroupEighteenOrOver
		bddpCreate.GrantorType = consent.GrantorTypeOwner
		bddpCreate.OwnerName = survey.Name
		bddpCreate.Type = consent.TypeBigDataDonationProject
		bddpCreate.Version = latestBDDP.Data[0].Version
		creates = append(creates, bddpCreate)
	}

	rippleCreate := consent.NewRecordCreate()
	rippleCreate.AgeGroup = consent.AgeGroupEighteenOrOver
	rippleCreate.GrantorType = consent.GrantorTypeOwner
	rippleCreate.OwnerName = survey.Name
	rippleCreate.Type = consent.TypeRipple
	rippleCreate.Version = consent.RippleVersionForJotform
	creates = append(creates, rippleCreate)

	_, err = s.consentService.CreateConsentRecords(ctx, userID, creates)
	if err != nil {
		return errors.Wrap(err, "unable to create consent records")
	}

	return nil
}

func (s *SubmissionProcessor) saveProcessedSubmission(ctx context.Context, submission *SubmissionResponse) error {
	var rawContent bson.Raw
	if submission.Content.RawContent != nil {
		if err := bson.UnmarshalExtJSON(submission.Content.RawContent, true, &rawContent); err != nil {
			return errors.Wrap(err, "failed to unmarshal raw content to bson")
		}
	}

	processedSubmission := &store.ProcessedSubmission{
		SubmissionID: submission.Content.ID,
		FormID:       submission.Content.FormID,
		RawContent:   rawContent,
		CreatedTime:  time.Now(),
	}

	if err := s.store.CreateProcessedSubmission(ctx, processedSubmission); err != nil && !mongo.IsDup(err) {
		return err
	}

	return nil
}

type ReconcileResult struct {
	TotalProcessed  int
	LastProcessedID string
}
