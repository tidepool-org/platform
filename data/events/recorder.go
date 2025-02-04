package events

import (
	"context"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	lognull "github.com/tidepool-org/platform/log/null"
)

type Recorder struct {
	Repo alerts.RecordsRepository
}

func NewRecorder(repo alerts.RecordsRepository) *Recorder {
	return &Recorder{
		Repo: repo,
	}
}

func (r *Recorder) RecordReceivedDeviceData(ctx context.Context,
	lastComm alerts.LastCommunication) error {

	logger := r.log(ctx).WithFields(log.Fields{
		"userID":    lastComm.UserID,
		"dataSetID": lastComm.DataSetID,
	})
	logger.Info("recording received data")
	if err := r.Repo.RecordReceivedDeviceData(ctx, lastComm); err != nil {
		return errors.Wrap(err, "Unable to record metadata on reception of device data")
	}
	return nil
}

func (r *Recorder) log(ctx context.Context) log.Logger {
	if ctxLogger := log.LoggerFromContext(ctx); ctxLogger != nil {
		return ctxLogger
	}
	return lognull.NewLogger()
}
