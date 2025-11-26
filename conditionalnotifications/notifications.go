package conditionalnotifications

import (
	"math/rand"
	"time"

	confirmationClient "github.com/tidepool-org/hydrophone/client"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/clinics"
	dataSourceStore "github.com/tidepool-org/platform/data/source/store/structured"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/mailer"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/work"
)

const (
	baseRetryDuration   = 1 * time.Minute
	retryDurationJitter = 5 * time.Second
)

type Dependencies struct {
	Auth          auth.RestrictedTokenAccessor
	Clinics       clinics.Client
	Confirmations confirmationClient.ClientWithResponsesInterface
	DataSources   dataSourceStore.DataSourcesRepository
	Mailer        mailer.Mailer
	Users         user.Client
	Worker        work.Client
}

func NewFailingResult(err error, wrk *work.Work) work.ProcessResult {
	failingRetryCount := pointer.DefaultInt(wrk.FailingRetryCount, 0) + 1
	return *work.NewProcessResultFailing(work.FailingUpdate{
		FailingError:      errors.Serializable{Error: err},
		FailingRetryCount: pointer.DefaultInt(wrk.FailingRetryCount, 0) + 1,
		FailingRetryTime:  time.Now().Add(retryDuration(failingRetryCount)),
	})
}

func retryDuration(retryCount int) time.Duration {
	fallbackFactor := time.Duration(1 << (retryCount - 1))
	retryDurationJitter := int64(retryDurationJitter * fallbackFactor)
	return baseRetryDuration*fallbackFactor + time.Duration(rand.Int63n(2*retryDurationJitter)-retryDurationJitter)
}
