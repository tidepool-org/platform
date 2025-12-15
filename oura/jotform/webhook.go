package jotform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oura/customerio"
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
)

type WebhookProcessor struct {
	baseURL string
	apiKey  string

	logger log.Logger

	consentService   consent.Service
	customerIOClient *customerio.Client
	shopifyClient    shopify.Client
	userClient       user.Client
}

type Config struct {
	BaseURL string `envconfig:"TIDEPOOL_JOTFORM_BASE_URL"`
	APIKey  string `envconfig:"TIDEPOOL_JOTFORM_API_KEY"`
}

func NewWebhookProcessor(config Config, logger log.Logger, consentService consent.Service, customerIOClient *customerio.Client, userClient user.Client, shopifyClient shopify.Client) (*WebhookProcessor, error) {
	return &WebhookProcessor{
		apiKey:  config.APIKey,
		baseURL: config.BaseURL,

		logger: logger,

		consentService:   consentService,
		customerIOClient: customerIOClient,
		shopifyClient:    shopifyClient,
		userClient:       userClient,
	}, nil
}

func (w *WebhookProcessor) ProcessSubmission(ctx context.Context, submissionID string) error {
	logger := w.logger.WithField("submission", submissionID)
	submission, err := w.getSubmission(ctx, submissionID)
	if err != nil {
		return errors.Wrap(err, "failed to get submission")
	}
	if submission.Content.Answers == nil {
		logger.Warn("submission has no answers")
		return nil
	}
	identifiers, err := w.validateUser(ctx, submissionID, submission.Content.Answers)
	if err != nil {
		logger.WithError(err).Warn("user is invalid")
		return nil
	} else if identifiers == nil {
		logger.Warn("invalid user")
		return nil
	}

	return w.handleSurveyCompleted(ctx, *identifiers, submission)
}

// validateUser validates the user id by comparing the participant id from the submission with the participant id from customer.io
// this is required because jotform webhooks are not signed or authenticated
func (w *WebhookProcessor) validateUser(ctx context.Context, submissionID string, answers Answers) (*customerio.Identifiers, error) {
	logger := w.logger.WithField("submission", submissionID)

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

	customer, err := w.customerIOClient.GetCustomer(ctx, userID, customerio.IDTypeUserID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get customer with id %s", userID)
	}

	if customer == nil {
		return nil, errors.Newf("customer with id %s not found", userID)
	}
	if customer.OuraParticipantID != participantID {
		return nil, errors.Newf("participant id mismatch for user with id %s", userID)
	}

	usr, err := w.userClient.Get(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user")
	}
	if usr == nil {
		return nil, errors.New("user not found")
	}

	return &customer.Identifiers, nil
}

func (w *WebhookProcessor) getSubmission(ctx context.Context, submissionID string) (*SubmissionResponse, error) {
	url := fmt.Sprintf("%s/v1/submission/%s", w.baseURL, submissionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request: %w")
	}

	// Add authorization header
	req.Header.Set("APIKEY", w.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request: %w")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("unexpected status code: %d", resp.StatusCode)
	}

	var response SubmissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	// Sometimes the jotform webhook returns a 200 http response with a non-200 response code in the body
	if response.ResponseCode != http.StatusOK {
		return nil, errors.Newf("unexpected response code: %d", response.ResponseCode)
	}

	return &response, nil
}

func (w *WebhookProcessor) handleSurveyCompleted(ctx context.Context, identifiers customerio.Identifiers, submission *SubmissionResponse) error {
	surveyCompletedData := customerio.OuraEligibilitySurveyCompletedData{
		OuraEligibilitySurveyID:       submission.Content.ID,
		OuraEligibilitySurveyEligible: submission.Content.Answers.GetAnswerTextByName(EligibleField) == "true",
	}

	if surveyCompletedData.OuraEligibilitySurveyEligible {
		if err := w.ensureConsentRecordExists(ctx, identifiers.ID, submission); err != nil {
			w.logger.WithField("submission", submission.Content.ID).WithError(err).Warn("unable to ensure consent record exists")
			return err
		}

		surveyCompletedData.OuraSizingKitDiscountCode = shopify.RandomDiscountCode()
		err := w.shopifyClient.CreateDiscountCode(ctx, shopify.DiscountCodeInput{
			Title:     shopify.OuraSizingKitDiscountCodeTitle,
			Code:      surveyCompletedData.OuraSizingKitDiscountCode,
			ProductID: shopify.OuraSizingKitProductID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create oura ring discount code")
		}
	}

	surveyCompleted := customerio.Event{
		Name: customerio.OuraEligibilitySurveyCompletedEventType,
		ID:   surveyCompletedData.OuraEligibilitySurveyID,
		Data: surveyCompletedData,
	}

	err := w.customerIOClient.SendEvent(ctx, identifiers.CID, surveyCompleted)
	if err != nil {
		return errors.Wrap(err, "unable to send sizing kit delivered event")
	}

	return nil
}

func (w *WebhookProcessor) ensureConsentRecordExists(ctx context.Context, userID string, submission *SubmissionResponse) error {
	logger := w.logger.WithField("submission", submission.Content.ID)

	survey := OuraEligibilitySurvey{
		DateOfBirth: submission.Content.Answers.GetAnswerTextByName(DateOfBirthField),
		Name:        submission.Content.Answers.GetAnswerTextByName(NameField),
	}

	v := validator.New(w.logger)
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

	records, err := w.consentService.ListConsentRecords(ctx, userID, filter, pagination)
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

	_, err = w.consentService.CreateConsentRecord(ctx, userID, create)
	if err != nil {
		return errors.Wrap(err, "unable to create consent record")
	}

	return nil
}
