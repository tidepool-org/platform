package processors

import (
	confirmationClient "github.com/tidepool-org/hydrophone/client"

	"github.com/tidepool-org/platform/clinics"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/mailer"
	notificationsWorkClaims "github.com/tidepool-org/platform/notifications/work/claims"
	notificationsWorkConnectionsIssues "github.com/tidepool-org/platform/notifications/work/connections/issues"
	notificationsWorkConnectionsRequests "github.com/tidepool-org/platform/notifications/work/connections/requests"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
)

type (
	ClinicClient       = clinics.Client
	ConfirmationClient = confirmationClient.ClientWithResponsesInterface
	DataSourceClient   = dataSource.Client
	MailerClient       = mailer.Client
	UserClient         = user.Client
)

type Dependencies struct {
	workBase.Dependencies
	ClinicClient
	ConfirmationClient
	DataSourceClient
	MailerClient
	UserClient
}

func (d Dependencies) Validate() error {
	if err := d.Dependencies.Validate(); err != nil {
		return err
	}
	if d.ClinicClient == nil {
		return errors.New("clinic client is missing")
	}
	if d.ConfirmationClient == nil {
		return errors.New("confirmation client is missing")
	}
	if d.DataSourceClient == nil {
		return errors.New("data source client is missing")
	}
	if d.MailerClient == nil {
		return errors.New("mailer client is missing")
	}
	if d.UserClient == nil {
		return errors.New("user client is missing")
	}
	return nil
}

func NewProcessorFactories(dependencies Dependencies) ([]work.ProcessorFactory, error) {
	if err := dependencies.Validate(); err != nil {
		return nil, errors.Wrap(err, "dependencies is invalid")
	}

	var processorFactories []work.ProcessorFactory

	if processorFactory, err := notificationsWorkClaims.NewProcessorFactory(notificationsWorkClaims.Dependencies{
		Dependencies:       dependencies.Dependencies,
		ClinicClient:       dependencies.ClinicClient,
		ConfirmationClient: dependencies.ConfirmationClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := notificationsWorkConnectionsIssues.NewProcessorFactory(notificationsWorkConnectionsIssues.Dependencies{
		Dependencies: dependencies.Dependencies,
		MailerClient: dependencies.MailerClient,
		UserClient:   dependencies.UserClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	if processorFactory, err := notificationsWorkConnectionsRequests.NewProcessorFactory(notificationsWorkConnectionsRequests.Dependencies{
		Dependencies:     dependencies.Dependencies,
		ClinicClient:     dependencies.ClinicClient,
		DataSourceClient: dependencies.DataSourceClient,
		MailerClient:     dependencies.MailerClient,
		UserClient:       dependencies.UserClient,
	}); err != nil {
		return nil, err
	} else {
		processorFactories = append(processorFactories, processorFactory)
	}

	return processorFactories, nil
}
