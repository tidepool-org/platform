package service

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"

	"github.com/tidepool-org/mailer/mailer"

	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/page"

	prescriptionStore "github.com/tidepool-org/platform/prescription/store"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
)

type PrescriptionService struct {
	emailTemplate     *mailer.EmailTemplate
	config            *PrescriptionServiceConfig
	mailer            mailer.Mailer
	prescriptionStore prescriptionStore.Store
}

type PrescriptionServiceParams struct {
	fx.In

	Config *PrescriptionServiceConfig
	Store  prescriptionStore.Store
	Mailer mailer.Mailer
}

type PrescriptionServiceConfig struct {
	WebAppURL string `envconfig:"TIDEPOOL_WEBAPP_URL" required:"true"`
	AssetURL  string `envconfig:"TIDEPOOL_ASSETS_URL" required:"true"`
}

func NewPrescriptionServiceConfig(lifecycle fx.Lifecycle) *PrescriptionServiceConfig {
	cfg := &PrescriptionServiceConfig{}
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return cfg.Load()
		},
	})

	return cfg
}

func (p *PrescriptionServiceConfig) Load() error {
	return envconfig.Process("", p)
}

func NewService(params PrescriptionServiceParams) (prescription.Service, error) {
	if params.Store == nil {
		return nil, errors.New("prescription store is missing")
	}
	if params.Mailer == nil {
		return nil, errors.New("mailer is missing")
	}
	if params.Config == nil {
		return nil, errors.New("config is missing")
	}
	emailTemplate, err := NewPrescriptionEmailTemplate()
	if err != nil {
		return nil, err
	}

	return &PrescriptionService{
		config:            params.Config,
		emailTemplate:     emailTemplate,
		mailer:            params.Mailer,
		prescriptionStore: params.Store,
	}, nil
}

func (p *PrescriptionService) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	prescr, err := repo.CreatePrescription(ctx, userID, create)
	if err != nil || prescr == nil {
		return prescr, err
	}
	if p.shouldSendAccessCodeEmailAfterSuccessfulCreation(create) {
		return prescr, p.sendAccessCodeEmail(ctx, prescr)
	}
	return prescr, err
}

func (p *PrescriptionService) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.ListPrescriptions(ctx, filter, pagination)
}

func (p *PrescriptionService) DeletePrescription(ctx context.Context, clinicianID string, id string) (bool, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.DeletePrescription(ctx, clinicianID, id)
}

func (p *PrescriptionService) AddRevision(ctx context.Context, usr *user.User, id string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.AddRevision(ctx, usr, id, create)
}

func (p *PrescriptionService) ClaimPrescription(ctx context.Context, usr *user.User, claim *prescription.Claim) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.ClaimPrescription(ctx, usr, claim)
}

func (p *PrescriptionService) UpdatePrescriptionState(ctx context.Context, usr *user.User, id string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	prescr, err := repo.UpdatePrescriptionState(ctx, usr, id, update)
	if err != nil || prescr == nil {
		return prescr, err
	}
	if p.shouldSendAccessCodeEmailAfterSuccessfulUpdate(update) {
		return prescr, p.sendAccessCodeEmail(ctx, prescr)
	}
	return prescr, err
}

func (p *PrescriptionService) shouldSendAccessCodeEmailAfterSuccessfulCreation(create *prescription.RevisionCreate) bool {
	return create.State == prescription.StateSubmitted
}

func (p *PrescriptionService) shouldSendAccessCodeEmailAfterSuccessfulUpdate(update *prescription.StateUpdate) bool {
	return update.State == prescription.StateSubmitted
}

func (p *PrescriptionService) sendAccessCodeEmail(ctx context.Context, prescr *prescription.Prescription) error {
	email, err := p.createEmail(prescr)
	if err != nil {
		return err
	}

	return p.mailer.Send(ctx, email)
}

func (p *PrescriptionService) createEmail(prescr *prescription.Prescription) (*mailer.Email, error) {
	email := &mailer.Email{
		Recipients: []string{prescr.LatestRevision.Attributes.Email},
	}
	emailParams := map[string]string{
		"AccessCode": prescr.AccessCode,
		"AssetURL":   p.config.AssetURL,
		"WebURL":     p.config.WebAppURL,
	}
	if err := p.emailTemplate.RenderToEmail(emailParams, email); err != nil {
		return nil, err
	}
	return email, nil
}
