package emailnotificationsprocessor

import (
	"fmt"
	"time"

	confirmationClient "github.com/tidepool-org/hydrophone/client"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/clinics"
	dataSourceStore "github.com/tidepool-org/platform/data/source/store/structured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/mailer"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
)

const (
	Quantity                       = 4
	Frequency                      = time.Minute
	ProcessingTimeoutSeconds       = 60
	ProcessingAvailableRepeatDelay = time.Hour
)

type Dependencies struct {
	Mailer        mailer.Mailer
	Auth          auth.RestrictedTokenAccessor
	DataSources   dataSourceStore.DataSourcesRepository
	Confirmations confirmationClient.ClientWithResponsesInterface
	Users         user.Client
	Clinics       clinics.Client
}

func NewFailedResult(err error, wrk *work.Work) work.ProcessResult {
	return *work.NewProcessResultFailed(work.FailedUpdate{
		FailedError: errors.Serializable{Error: err},
		Metadata:    wrk.Metadata,
	})
}

func NewProcessors(deps Dependencies) ([]work.Processor, error) {
	if deps.Mailer == nil {
		return nil, fmt.Errorf(`dependency "Mailer" is nil`)
	}
	if deps.Auth == nil {
		return nil, fmt.Errorf(`dependency "Auth" is nil`)
	}
	if deps.DataSources == nil {
		return nil, fmt.Errorf(`dependency "DataSources" is nil`)
	}
	if deps.Users == nil {
		return nil, fmt.Errorf(`dependency "Users" is nil`)
	}
	if deps.Clinics == nil {
		return nil, fmt.Errorf(`dependency "Clinics" is nil`)
	}
	if deps.Confirmations == nil {
		return nil, fmt.Errorf(`dependency "Confirmations" is nil`)
	}

	return []work.Processor{
		newDeviceConnectionIssuesProcessor(deps),
		newClaimAccountProcessor(deps),
		newConnectAccountProcessor(deps),
	}, nil
}
