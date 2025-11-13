package emailnotificationsprocessor

import (
	"context"
	"fmt"
	"time"

	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/work"
)

const (
	claimAccountProcessorType = "org.tidepool.processors.claim.account"
)

// ClaimAccountReminderData is the metadata added to a [work.Work] item for reminding users to claim their account.
type ClaimAccountReminderData struct {
	UserId     string
	Email      string
	ClinicId   string
	ClinicName string
}

type claimAccountProcessor struct {
	dependencies Dependencies
}

func newClaimAccountProcessor(dependencies Dependencies) *claimAccountProcessor {
	return &claimAccountProcessor{
		dependencies: dependencies,
	}
}

func (p *claimAccountProcessor) Type() string {
	return claimAccountProcessorType
}

func (p *claimAccountProcessor) Quantity() int {
	return Quantity
}

func (p *claimAccountProcessor) Frequency() time.Duration {
	return Frequency
}

func (p *claimAccountProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	data, err := toClaimAccountData(wrk)
	if err != nil {
		return NewFailedResult(err, wrk)
	}

	patient, err := p.dependencies.Clinics.GetPatient(ctx, data.ClinicId, data.UserId)
	if err != nil {
		return NewFailedResult(err, wrk)
	}
	if patient != nil && (patient.Permissions.Custodian == nil || len(*patient.Permissions.Custodian) == 0) {
		// User has already claimed account, no need to send reminder.
		return *work.NewProcessResultDelete()
	}

	if _, err := p.dependencies.Confirmations.ResendAccountSignupWithResponse(ctx, data.Email); err != nil {
		return NewFailedResult(err, wrk)
	}
	return *work.NewProcessResultDelete()
}

func (d *ClaimAccountReminderData) Parse(parser structure.ObjectParser) {
	d.UserId = *parser.String("userId")
	d.Email = *parser.String("email")
	d.ClinicId = *parser.String("clinicId")
	d.ClinicName = *parser.String("clinicName")
}

func (d *ClaimAccountReminderData) Validate(validator structure.Validator) {
	validator.String("userId", &d.UserId).NotEmpty()
	validator.String("email", &d.Email).NotEmpty()
	validator.String("clinicId", &d.ClinicId).NotEmpty()
	validator.String("clinicName", &d.ClinicName).NotEmpty()
}

func NewClaimAccountWorkCreate(notBefore time.Time, metadata ClaimAccountReminderData) *work.Create {
	return &work.Create{
		Type:                    claimAccountProcessorType,
		SerialID:                pointer.FromString(fmt.Sprintf("claim.%s.%s", metadata.UserId, metadata.Email)),
		GroupID:                 pointer.FromString(fmt.Sprintf("claim.%s", metadata.UserId)), // grouping related claim account by userId in case we need to bulk delete them
		ProcessingTimeout:       ProcessingTimeoutSeconds,
		ProcessingAvailableTime: notBefore,
		Metadata:                fromClaimAccountData(metadata),
	}
}

func toClaimAccountData(wrk *work.Work) (*ClaimAccountReminderData, error) {
	wrk.EnsureMetadata()
	var data ClaimAccountReminderData
	if userId, ok := wrk.Metadata["userId"].(string); ok {
		data.UserId = userId
	} else {
		return nil, fmt.Errorf(`expected field "userId" to exist and be a string, received %T`, wrk.Metadata["userId"])
	}
	if email, ok := wrk.Metadata["email"].(string); ok {
		data.Email = email
	} else {
		return nil, fmt.Errorf(`expected field "email" to exist and be a string, received %T`, wrk.Metadata["email"])
	}
	if clinicId, ok := wrk.Metadata["clinicId"].(string); ok {
		data.ClinicId = clinicId
	} else {
		return nil, fmt.Errorf(`expected field "clinicId" to exist and be a string, received %T`, wrk.Metadata["clinicId"])
	}
	if clinicName, ok := wrk.Metadata["clinicName"].(string); ok {
		data.ClinicName = clinicName
	} else {
		return nil, fmt.Errorf(`expected field "clinicName" to exist and be a string, received %T`, wrk.Metadata["clinicName"])
	}
	return &data, nil
}

func fromClaimAccountData(data ClaimAccountReminderData) map[string]any {
	return map[string]any{
		"userId":     data.UserId,
		"clinicId":   data.ClinicId,
		"clinicName": data.ClinicName,
		"email":      data.Email,
	}
}
