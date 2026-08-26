package fetch

import (
	"context"
	"maps"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataSource "github.com/tidepool-org/platform/data/source"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/duration"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthToken "github.com/tidepool-org/platform/oauth/token"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
)

//go:generate mockgen -source=runner.go -destination=test/runner_mocks.go -package=test -typed

const (
	AvailableAfterDuration       = 120 * time.Minute
	AvailableAfterDurationJitter = 15 * time.Minute
	DataSetSize                  = 2000
	TaskDurationMaximum          = 15 * time.Minute
	TaskRetryCountMaximum        = 4                // Last retry after ~(AvailableAfterDuration * (2^TaskRetryCountMaximum - 1)) hours (discounting AvailableAfterDurationJitter)
	DataRangeDaysMaximum         = 30               // Per Dexcom documentation
	ResumeAfterDuration          = 1 * time.Minute  // Resume interrupted work after a short delay, so that a later run can pick up where the previous one left off
	ResumeExpirationDuration     = 10 * time.Minute // Several times ResumeAfterDuration, so ordinary queue lag is tolerated but a long wait is not
)

var initialDataTime = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)

type AuthClient interface {
	ServerSessionToken() (string, error)

	GetProviderSession(ctx context.Context, id string) (*auth.ProviderSession, error)
	UpdateProviderSession(ctx context.Context, id string, update *auth.ProviderSessionUpdate) (*auth.ProviderSession, error)
}

type DataClient interface {
	CreateUserDataSet(ctx context.Context, userID string, create *data.DataSetCreate) (*data.DataSet, error)
	GetDataSet(ctx context.Context, id string) (*data.DataSet, error)
	UpdateDataSet(ctx context.Context, id string, update *data.DataSetUpdate) (*data.DataSet, error)

	CreateDataSetsData(ctx context.Context, dataSetID string, datumArray []data.Datum) error
}

type DexcomClient interface {
	GetAlerts(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.AlertsResponse, error)
	GetCalibrations(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.CalibrationsResponse, error)
	GetDataRange(ctx context.Context, lastSyncTime *time.Time, tokenSource oauth.TokenSource) (*dexcom.DataRangesResponse, error)
	GetDevices(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.DevicesResponse, error)
	GetEGVs(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EGVsResponse, error)
	GetEvents(ctx context.Context, startTime time.Time, endTime time.Time, tokenSource oauth.TokenSource) (*dexcom.EventsResponse, error)
}

type Runner struct {
	authClient       AuthClient
	dataClient       DataClient
	dataSourceClient dataSource.Client
	dexcomClient     DexcomClient
}

func NewRunner(authClient AuthClient, dataClient DataClient, dataSourceClient dataSource.Client, dexcomClient DexcomClient) (*Runner, error) {
	if authClient == nil {
		return nil, errors.New("auth client is missing")
	}
	if dataClient == nil {
		return nil, errors.New("data client is missing")
	}
	if dataSourceClient == nil {
		return nil, errors.New("data source client is missing")
	}
	if dexcomClient == nil {
		return nil, errors.New("dexcom client is missing")
	}

	return &Runner{
		authClient:       authClient,
		dataClient:       dataClient,
		dataSourceClient: dataSourceClient,
		dexcomClient:     dexcomClient,
	}, nil
}

func (r *Runner) AuthClient() AuthClient {
	return r.authClient
}

func (r *Runner) DataClient() DataClient {
	return r.dataClient
}

func (r *Runner) DataSourceClient() dataSource.Client {
	return r.dataSourceClient
}

func (r *Runner) DexcomClient() DexcomClient {
	return r.dexcomClient
}

func (r *Runner) GetRunnerType() string {
	return Type
}

func (r *Runner) GetRunnerDeadline() time.Duration {
	return TaskDurationMaximum * 3
}

func (r *Runner) GetRunnerTimeout() time.Duration {
	return TaskDurationMaximum * 2
}

func (r *Runner) GetRunnerDurationMaximum() time.Duration {
	return TaskDurationMaximum
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) {
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, r.AuthClient())
	if taskRunner, err := NewTaskRunner(r, tsk); err != nil {
		log.LoggerFromContext(ctx).WithError(err).Warn("Unable to create task runner")
	} else {
		taskRunner.Run(ctx)
	}
}

type Provider interface {
	AuthClient() AuthClient
	DataClient() DataClient
	DataSourceClient() dataSource.Client
	DexcomClient() DexcomClient
	GetRunnerDurationMaximum() time.Duration
}

type TaskRunner struct {
	Provider
	task                *task.Task
	context             context.Context
	logger              log.Logger
	providerSession     *auth.ProviderSession
	dataSource          *dataSource.Source
	tokenSource         *oauthToken.Source
	deviceHashes        map[string]string
	deviceHashesPending map[string]string
	dataSet             *data.DataSet
	dataSetPreloaded    bool
	runTime             time.Time
	availableAfter      *time.Duration
	completed           bool
	tokenUpdateError    error
}

func NewTaskRunner(provider Provider, tsk *task.Task) (*TaskRunner, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &TaskRunner{
		Provider: provider,
		task:     tsk,
	}, nil
}

func (t *TaskRunner) Run(ctx context.Context) {
	t.context = ctx
	t.logger = log.LoggerFromContext(t.context)
	t.runTime = time.Now()

	t.task.ClearError()

	if err := t.run(); err != nil {
		t.task.AppendError(err)
	}

	// If we didn't lose the claim, then update data source and repeat if not failed
	if !errors.Is(context.Cause(t.context), task.ErrClaimLost) {
		err := t.updateDataSourceWithTaskState()
		if err != nil {
			t.task.AppendError(err)
		}
		if err != nil || !t.task.IsFailed() {
			t.task.RepeatAvailableAfter(pointer.Default(t.availableAfter, availableAfterDuration()))
		}
	} else {
		t.logger.Warn("Skipped updating data source and task because the task claim was lost")
	}
}

func (t *TaskRunner) run() error {
	if len(t.task.Data) == 0 {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("data is missing"))
	}

	if err := t.getDataSource(); err != nil {
		return err
	}
	if err := t.getProviderSession(); err != nil {
		return err
	}
	if err := t.createTokenSource(); err != nil {
		return err
	}
	if err := t.getDeviceHashes(); err != nil {
		return err
	}
	if err := t.fetchSinceLatestDataTime(); err != nil {
		return err
	}

	return nil
}

func (t *TaskRunner) getProviderSession() error {
	providerSessionID, ok := t.task.Data[dexcom.DataKeyProviderSessionID].(string)
	if !ok || providerSessionID == "" {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("provider session id is missing"))
	}

	providerSession, err := t.AuthClient().GetProviderSession(t.context, providerSessionID)
	if err != nil {
		return ErrorResourceFailureError(errors.Wrap(err, "unable to get provider session"))
	} else if providerSession == nil {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("provider session is missing"))
	}
	t.providerSession = providerSession

	return nil
}

func (t *TaskRunner) updateProviderSession() error {
	token := t.tokenSource.Token()
	if token == t.providerSession.OAuthToken {
		return nil // Token not changed, do not update
	}

	// Do not interrupt normally, but do enforce a reasonable timeout
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.context), 10*time.Second)
	defer cancel()

	providerSessionUpdate := auth.NewProviderSessionUpdate()
	providerSessionUpdate.OAuthToken = token
	providerSessionUpdate.ExternalID = t.providerSession.ExternalID
	providerSession, err := t.AuthClient().UpdateProviderSession(ctx, t.providerSession.ID, providerSessionUpdate)
	if err != nil {
		return errors.Wrap(err, "unable to update provider session")
	} else if providerSession == nil {
		return errors.New("provider session is missing")
	}
	t.providerSession = providerSession

	return nil
}

func (t *TaskRunner) getDataSource() error {
	dataSourceID, ok := t.task.Data[dexcom.DataKeyDataSourceID].(string)
	if !ok || dataSourceID == "" {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("data source id is missing"))
	}

	source, err := t.DataSourceClient().Get(t.context, dataSourceID)
	if err != nil {
		return ErrorResourceFailureError(errors.Wrap(err, "unable to get data source"))
	} else if source == nil {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("data source is missing"))
	}
	t.dataSource = source

	return nil
}

func (t *TaskRunner) updateDataSourceWithDataSet(dataSet *data.DataSet) error {
	return t.updateDataSource(&dataSource.Update{DataSetID: dataSet.ID})
}

func (t *TaskRunner) updateDataSourceWithDataTime(earliestDataTime *time.Time, latestDataTime *time.Time) error {
	// The comparisons below read the data source, so guard before them rather than relying on updateDataSource
	if t.dataSource == nil {
		return nil
	}

	update := dataSource.NewUpdate()

	if t.beforeEarliestDataTime(earliestDataTime) {
		update.EarliestDataTime = earliestDataTime
	}
	if t.afterLatestDataTime(latestDataTime) {
		update.LatestDataTime = latestDataTime
	}

	update.Error = errors.NewSerializable(nil)
	update.LastImportTime = pointer.FromTime(time.Now())
	if err := t.updateDataSource(update); err != nil {
		return err
	}

	if update.LatestDataTime != nil {
		return t.cancelPendingResume()
	}

	return nil
}

func (t *TaskRunner) updateDataSourceWithTaskState() error {
	update := dataSource.NewUpdate()
	if t.completed {
		update.LastImportTime = pointer.FromTime(time.Now())
	}
	if t.task.IsFailed() {
		update.State = pointer.FromString(dataSource.StateError)
	}
	update.Error = errors.NewSerializable(t.task.GetError())
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSource(update *dataSource.Update) error {
	if update.IsEmpty() || t.dataSource == nil {
		return nil
	}

	// Do not interrupt normally, but do enforce a reasonable timeout
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.context), 10*time.Second)
	defer cancel()

	dataSource, err := t.DataSourceClient().Update(ctx, t.dataSource.ID, nil, update)
	if err != nil {
		return ErrorResourceFailureError(errors.WithMeta(errors.Wrap(err, "unable to update data source"), update))
	} else if dataSource == nil {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("data source is missing"))
	}

	t.dataSource = dataSource
	return nil
}

func (t *TaskRunner) createTokenSource() error {
	tokenSource, err := oauthToken.NewSourceWithToken(t.providerSession.OAuthToken)
	if err != nil {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.Wrap(err, "unable to create token source"))
	}

	t.tokenSource = tokenSource
	return nil
}

func (t *TaskRunner) getDeviceHashes() error {
	raw, rawOK := t.task.Data[dexcom.DataKeyDeviceHashes]
	if !rawOK || raw == nil {
		return nil
	}
	rawMap, rawMapOK := raw.(map[string]any)
	if !rawMapOK || rawMap == nil {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("device hashes is invalid"))
	}
	deviceHashes := map[string]string{}
	for key, value := range rawMap {
		if valueString, valueStringOK := value.(string); valueStringOK {
			deviceHashes[key] = valueString
		} else {
			t.task.SetFailed()
			return ErrorInvalidStateError(errors.New("device hash is invalid"))
		}
	}

	t.deviceHashes = deviceHashes
	return nil
}

// updateDeviceHashPending reports whether the device hash changed, recording the new hash as pending until the resulting datum
// is stored. Recording it immediately would drop the datum for good if a later failure prevented the store, since the
// next run only translates a device whose hash changed.
func (t *TaskRunner) updateDeviceHashPending(device *dexcom.Device) bool {
	deviceID := device.ID()
	deviceHash, err := device.Hash()
	if err != nil {
		return false
	}

	// If the device hash has not changed, then no need to update
	if existingDeviceHash, ok := t.deviceHashesPending[deviceID]; ok && existingDeviceHash == deviceHash {
		return false
	} else if existingDeviceHash, ok = t.deviceHashes[deviceID]; ok && existingDeviceHash == deviceHash {
		return false
	}

	if t.deviceHashesPending == nil {
		t.deviceHashesPending = map[string]string{}
	}
	t.deviceHashesPending[deviceID] = deviceHash
	return true
}

// commitDeviceHashesPending promotes the pending device hashes now that their datums are stored, persisting them with the task
// so that the next run skips those devices.
func (t *TaskRunner) commitDeviceHashesPending() {
	if len(t.deviceHashesPending) == 0 {
		return
	}

	if t.deviceHashes == nil {
		t.deviceHashes = map[string]string{}
	}
	maps.Copy(t.deviceHashes, t.deviceHashesPending)
	clear(t.deviceHashesPending)

	t.task.Data[dexcom.DataKeyDeviceHashes] = maps.Clone(t.deviceHashes)
}

// getStartTimeMinimum returns the minimum start time for fetching data, considering the latest data time from the data
// source, the absolute initial data time, and any unexpired resume data time, if applicable.
func (t *TaskRunner) getStartTimeMinimum() time.Time {
	// Default to latest data time from data source or use the initial data time if none
	startTimeMinimum := pointer.DefaultTime(t.dataSource.LatestDataTime, initialDataTime)

	// If there is a resume data time and it is after the start time minimum then see if we can use it
	resumeDataTime := t.getDataTimeField(dexcom.DataKeyResumeDataTime)
	if resumeDataTime != nil && resumeDataTime.After(startTimeMinimum) {
		resumeExpirationTime := t.getDataTimeField(dexcom.DataKeyResumeExpirationTime)

		lgr := t.logger.WithFields(log.Fields{
			dexcom.DataKeyResumeDataTime:       *resumeDataTime,
			dexcom.DataKeyResumeExpirationTime: resumeExpirationTime,
		})

		// Ensure the resume data time isn't expired (too long after last run)
		if resumeExpirationTime == nil || resumeExpirationTime.Before(t.runTime) {
			lgr.Warn("Ignoring expired resume data time")
		} else {
			lgr.Debug("Resuming fetch from resume data time")
			startTimeMinimum = *resumeDataTime
		}
	}

	return startTimeMinimum
}

func (t *TaskRunner) resumeWithDataTime(dataTime time.Time) error {
	t.setDataTimeField(dexcom.DataKeyResumeDataTime, dataTime)
	t.setDataTimeField(dexcom.DataKeyResumeExpirationTime, time.Now().Add(ResumeExpirationDuration))
	t.availableAfter = pointer.From(duration.WithJitter(ResumeAfterDuration, 0.2))
	return nil
}

func (t *TaskRunner) cancelPendingResume() error {
	delete(t.task.Data, dexcom.DataKeyResumeDataTime)
	delete(t.task.Data, dexcom.DataKeyResumeExpirationTime)
	return nil
}

func (t *TaskRunner) complete() error {
	t.completed = true
	return t.cancelPendingResume()
}

func (t *TaskRunner) updateDataSetWithTimezoneOffset(timezoneOffset *int) error {
	if timezoneOffset == nil {
		return nil
	}
	return t.updateDataSet(&data.DataSetUpdate{TimeZoneOffset: timezoneOffset})
}

func (t *TaskRunner) updateDataSet(update *data.DataSetUpdate) error {
	if update.IsEmpty() || t.dataSet == nil {
		return nil
	}

	// Do not interrupt normally, but do enforce a reasonable timeout
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.context), 10*time.Second)
	defer cancel()

	dataSet, err := t.DataClient().UpdateDataSet(ctx, *t.dataSet.UploadID, update)
	if err != nil {
		return ErrorResourceFailureError(errors.WithMeta(errors.Wrap(err, "unable to update data set"), update))
	} else if dataSet == nil {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("data set is missing"))
	}

	t.dataSet = dataSet
	return nil
}

func (t *TaskRunner) fetchSinceLatestDataTime() error {
	dataRange, err := t.fetchDataRange()
	if err != nil {
		return err
	} else if dataRange == nil {
		return t.complete() // No data, but still successful import
	}

	startTime := dataRange.StartTime
	for {
		endTime := startTime.AddDate(0, 0, DataRangeDaysMaximum)
		if endTime.After(dataRange.EndTime) {
			endTime = dataRange.EndTime
		}

		if err := t.fetch(startTime, endTime); err != nil {
			return err
		}

		// Next fetch starts at the end of the last fetch, since Dexcom data ranges are inclusive of the start and
		// end times, so that no data is missed.
		startTime = endTime

		// If no more remains to fetch, then indicate import completed.
		if !startTime.Before(dataRange.EndTime) {
			return t.complete()
		}

		// If the task has been running for longer than the maximum duration, then stop fetching and repeat quickly,
		// capturing how far this run got so that the next one resumes instead of walking it all again.
		if time.Since(t.runTime) > t.GetRunnerDurationMaximum() {
			return t.resumeWithDataTime(startTime)
		}
	}
}

func (t *TaskRunner) fetchDataRange() (*DataRange, error) {
	// NOTE: Per Dexcom support, the lastSyncTime parameter does not work as
	// expected in all situations (e.g. signal loss, last data more than 100
	// days ago). Dexcom support recommends to not specify the lastSyncTime
	// parameter for any request. Since the code below clamps any date range
	// to the data source latestDateTime and the current time, this will work
	// as expected FOR NOW. If/when we add deduplication and support
	// update and delete of OLDER data, we will need to revisit this logic.
	response, err := t.DexcomClient().GetDataRange(t.context, nil, t)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
	} else if response == nil {
		return nil, ErrorResourceFailureError(errors.New("data ranges response is missing"))
	}

	// Get data range, if none valid, then indicate nothing to fetch
	dataRange := response.DataRange()
	if dataRange == nil {
		return nil, nil
	}

	// Clamp data range, if none valid, then indicate nothing to fetch
	startTimeMinimum := t.getStartTimeMinimum()
	endTimeMaximum := time.Now()
	startTime := ClampTime(*dataRange.Start.SystemTimeRaw(), startTimeMinimum, endTimeMaximum)
	endTime := ClampTime(*dataRange.End.SystemTimeRaw(), startTimeMinimum, endTimeMaximum)
	if !startTime.Before(endTime) {
		return nil, nil
	}

	return &DataRange{
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

func (t *TaskRunner) fetch(startTime time.Time, endTime time.Time) error {
	datumArray, err := t.fetchData(startTime, endTime)
	if err != nil {
		return err
	} else if len(datumArray) == 0 {
		return nil
	}

	if err = t.prepareDataSet(); err != nil {
		return err
	}

	if err = t.storeDatumArray(datumArray); err != nil {
		return err
	}

	t.commitDeviceHashesPending()
	return nil
}

func (t *TaskRunner) preloadDataSet() error {
	if t.dataSet != nil || t.dataSetPreloaded {
		return nil
	}

	dataSet, err := t.findDataSet()
	if err != nil {
		return err
	}

	t.dataSet = dataSet
	t.dataSetPreloaded = true
	return nil
}

func (t *TaskRunner) fetchData(startTime time.Time, endTime time.Time) (data.Data, error) {
	datumArray := data.Data{}

	if fetchDatumArray, err := t.fetchAlerts(startTime, endTime); err != nil {
		return nil, err
	} else {
		datumArray = append(datumArray, fetchDatumArray...)
	}

	if fetchDatumArray, err := t.fetchCalibrations(startTime, endTime); err != nil {
		return nil, err
	} else {
		datumArray = append(datumArray, fetchDatumArray...)
	}

	if fetchDatumArray, err := t.fetchDevices(startTime, endTime); err != nil {
		return nil, err
	} else {
		datumArray = append(datumArray, fetchDatumArray...)
	}

	if fetchDatumArray, err := t.fetchEGVs(startTime, endTime); err != nil {
		return nil, err
	} else {
		datumArray = append(datumArray, fetchDatumArray...)
	}

	if fetchDatumArray, err := t.fetchEvents(startTime, endTime); err != nil {
		return nil, err
	} else {
		datumArray = append(datumArray, fetchDatumArray...)
	}

	sort.Stable(data.DataByTime(datumArray))

	return datumArray, nil
}

func (t *TaskRunner) fetchAlerts(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetAlerts(t.context, startTime, endTime, t)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
	} else if response == nil || response.Records == nil {
		return nil, ErrorResourceFailureError(errors.New("alerts response is missing"))
	}

	var alerts dexcom.Alerts
	for index, record := range *response.Records {
		if err := structureValidator.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Validate(record); err != nil {
			t.logger.WithError(err).Error("Failure validating Dexcom Alert")
		} else if err := structureNormalizer.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Normalize(record); err != nil {
			t.logger.WithError(err).Error("Failure normalizing Dexcom Alert")
		} else {
			alerts = append(alerts, record)
		}
	}

	datumArray := data.Data{}
	for _, alert := range alerts {
		if time := alert.SystemTime.Raw(); time != nil && InTimeRange(*time, startTime, endTime) {
			datumArray = append(datumArray, translateAlertToDatum(t.context, alert, response.RecordVersion))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchCalibrations(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetCalibrations(t.context, startTime, endTime, t)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
	} else if response == nil || response.Records == nil {
		return nil, ErrorResourceFailureError(errors.New("calibrations response is missing"))
	}

	var calibrations dexcom.Calibrations
	for index, record := range *response.Records {
		if err := structureValidator.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Validate(record); err != nil {
			t.logger.WithError(err).Error("Failure validating Dexcom Calibration")
		} else if err := structureNormalizer.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Normalize(record); err != nil {
			t.logger.WithError(err).Error("Failure normalizing Dexcom Calibration")
		} else {
			calibrations = append(calibrations, record)
		}
	}

	datumArray := data.Data{}
	for _, calibration := range calibrations {
		if time := calibration.SystemTime.Raw(); time != nil && InTimeRange(*time, startTime, endTime) {
			datumArray = append(datumArray, translateCalibrationToDatum(t.context, calibration))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchDevices(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetDevices(t.context, startTime, endTime, t)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
	} else if response == nil || response.Records == nil {
		return nil, ErrorResourceFailureError(errors.New("devices response is missing"))
	}

	var devices dexcom.Devices
	for index, record := range *response.Records {
		if err := structureValidator.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Validate(record); err != nil {
			t.logger.WithError(err).Error("Failure validating Dexcom Device")
		} else if err := structureNormalizer.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Normalize(record); err != nil {
			t.logger.WithError(err).Error("Failure normalizing Dexcom Device")
		} else {
			devices = append(devices, record)
		}
	}

	datumArray := data.Data{}
	for _, device := range devices {
		if time := device.LastUploadDate.Raw(); time != nil && InTimeRange(*time, startTime, endTime) && t.updateDeviceHashPending(device) {
			datumArray = append(datumArray, translateDeviceToDatum(t.context, device))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEGVs(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetEGVs(t.context, startTime, endTime, t)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
	} else if response == nil || response.Records == nil {
		return nil, ErrorResourceFailureError(errors.New("egvs response is missing"))
	}

	var egvs dexcom.EGVs
	for index, record := range *response.Records {
		if err := structureValidator.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Validate(record); err != nil {
			t.logger.WithError(err).Error("Failure validating Dexcom EGV")
		} else if err := structureNormalizer.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Normalize(record); err != nil {
			t.logger.WithError(err).Error("Failure normalizing Dexcom EGV")
		} else {
			egvs = append(egvs, record)
		}
	}

	datumArray := data.Data{}
	for _, egv := range egvs {
		if time := egv.SystemTime.Raw(); time != nil && InTimeRange(*time, startTime, endTime) {
			datumArray = append(datumArray, translateEGVToDatum(t.context, egv))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEvents(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetEvents(t.context, startTime, endTime, t)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
	} else if response == nil || response.Records == nil {
		return nil, ErrorResourceFailureError(errors.New("events response is missing"))
	}

	var events dexcom.Events
	for index, record := range *response.Records {
		if err := structureValidator.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Validate(record); err != nil {
			t.logger.WithError(err).Error("Failure validating Dexcom Event")
		} else if err := structureNormalizer.New(t.logger).WithReference("records").WithReference(strconv.Itoa(index)).Normalize(record); err != nil {
			t.logger.WithError(err).Error("Failure normalizing Dexcom Event")
		} else {
			events = append(events, record)
		}
	}

	datumArray := data.Data{}
	for _, event := range events {
		switch *event.EventStatus {
		case dexcom.EventStatusCreated:
			if time := event.SystemTime.Raw(); time != nil && InTimeRange(*time, startTime, endTime) {
				switch *event.EventType {
				case dexcom.EventTypeCarbs:
					datumArray = append(datumArray, translateEventCarbsToDatum(t.context, event))
				case dexcom.EventTypeExercise:
					datumArray = append(datumArray, translateEventExerciseToDatum(t.context, event))
				case dexcom.EventTypeHealth:
					datumArray = append(datumArray, translateEventHealthToDatum(t.context, event))
				case dexcom.EventTypeInsulin:
					datumArray = append(datumArray, translateEventInsulinToDatum(t.context, event))
				case dexcom.EventTypeBloodGlucose:
					datumArray = append(datumArray, translateEventBloodGlucoseToDatum(t.context, event))
				case dexcom.EventTypeNotes:
					datumArray = append(datumArray, translateEventNotesToDatum(t.context, event))
				}
			}
		case dexcom.EventStatusUpdated, dexcom.EventStatusDeleted:
			// FUTURE: Handle updated events
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) prepareDataSet() error {
	if err := t.preloadDataSet(); err != nil {
		return err
	}

	if t.dataSet == nil {
		dataSet, err := t.createDataSet()
		if err != nil {
			return err
		}
		t.dataSet = dataSet
	}

	// Everything downstream addresses the data set by upload id
	if t.dataSet.UploadID == nil {
		t.task.SetFailed()
		return ErrorInvalidStateError(errors.New("data set upload id is missing"))
	}

	return nil
}

func (t *TaskRunner) findDataSet() (*data.DataSet, error) {
	if t.dataSource.DataSetID == nil {
		return nil, nil
	}
	dataSet, err := t.DataClient().GetDataSet(t.context, *t.dataSource.DataSetID)
	if err != nil {
		return nil, ErrorResourceFailureError(errors.Wrap(err, "unable to get data set"))
	}
	return dataSet, nil
}

func (t *TaskRunner) createDataSet() (*data.DataSet, error) {
	dataSetCreate := data.NewDataSetCreate()
	dataSetCreate.Client = &data.DataSetClient{
		Name:    pointer.FromString(DataSetClientName),
		Version: pointer.FromString(DataSetClientVersion),
	}
	dataSetCreate.DataSetType = pointer.FromString(data.DataSetTypeContinuous)
	dataSetCreate.Deduplicator = data.NewDeduplicatorDescriptor()
	dataSetCreate.Deduplicator.Name = pointer.FromString(dataDeduplicatorDeduplicator.NoneName)
	dataSetCreate.Deduplicator.Version = pointer.FromString(dataDeduplicatorDeduplicator.NoneVersion)
	dataSetCreate.DeviceManufacturers = pointer.FromStringArray([]string{"Dexcom"})
	dataSetCreate.DeviceTags = pointer.FromStringArray([]string{data.DeviceTagCGM})
	dataSetCreate.Time = pointer.FromTime(time.Now())
	dataSetCreate.TimeProcessing = pointer.FromString(data.TimeProcessingNone)

	dataSet, err := t.DataClient().CreateUserDataSet(t.context, t.providerSession.UserID, dataSetCreate)
	if err != nil {
		return nil, ErrorResourceFailureError(errors.WithMeta(errors.Wrap(err, "unable to create data set"), dataSetCreate))
	}
	if err = t.updateDataSourceWithDataSet(dataSet); err != nil {
		return nil, err
	}

	return dataSet, nil
}

func (t *TaskRunner) storeDatumArray(datumArray data.Data) error {
	length := len(datumArray)
	for startIndex := 0; startIndex < length; startIndex += DataSetSize {
		endIndex := startIndex + DataSetSize
		if endIndex > length {
			endIndex = length
		}

		partialDatumArray := datumArray[startIndex:endIndex]

		if err := t.DataClient().CreateDataSetsData(t.context, *t.dataSet.UploadID, partialDatumArray); err != nil {
			return ErrorResourceFailureError(errors.Wrap(err, "unable to create data set data"))
		}

		earliestDataTime := partialDatumArray[0].GetTime()
		latestDataTime := partialDatumArray[len(partialDatumArray)-1].GetTime()
		if err := t.updateDataSourceWithDataTime(earliestDataTime, latestDataTime); err != nil {
			return err
		}

		// Determine last known timezone offset and persist with the data set
		var timezoneOffset *int
		for index := len(partialDatumArray) - 1; index >= 0; index-- {
			if timezoneOffset = partialDatumArray[index].GetTimeZoneOffset(); timezoneOffset != nil {
				break
			}
		}
		if err := t.updateDataSetWithTimezoneOffset(timezoneOffset); err != nil {
			return err
		}
	}

	return nil
}

func (t *TaskRunner) beforeEarliestDataTime(earliestDataTime *time.Time) bool {
	return earliestDataTime != nil && (t.dataSource.EarliestDataTime == nil || earliestDataTime.Before(*t.dataSource.EarliestDataTime))
}

func (t *TaskRunner) afterLatestDataTime(latestDataTime *time.Time) bool {
	return latestDataTime != nil && (t.dataSource.LatestDataTime == nil || latestDataTime.After(*t.dataSource.LatestDataTime))
}

func (t *TaskRunner) handleDexcomClientError(err error) error {
	// The stored refresh token is spent and its replacement was lost, so every later call authenticates against a token
	// the database does not have. Stop here rather than continuing, and leave the retry count alone.
	if t.tokenUpdateError != nil {
		return ErrorResourceFailureError(t.tokenUpdateError)
	}

	// If success, then reset retry count and return no error
	if err == nil {
		t.resetTaskRetryCount()
		return nil
	}

	// If not an authentication error, then just treat as a resource failure
	if !request.IsErrorUnauthenticated(errors.LastCause(err)) {
		return ErrorResourceFailureError(err)
	}

	// It is an authentication error, attempt retry, if possible
	err = ErrorAuthenticationFailureError(err)
	if retryCount := t.incrementTaskRetryCount(); retryCount <= TaskRetryCountMaximum {
		t.availableAfter = pointer.From(availableAfterDurationWithRetryCount(retryCount))
		return err
	}

	// Otherwise, we are failed
	t.task.SetFailed()
	return err
}

func (t *TaskRunner) incrementTaskRetryCount() int {
	retryCount := taskRetryCount(t.task.Data[dexcom.DataKeyRetryCount]) + 1
	t.task.Data[dexcom.DataKeyRetryCount] = int32(retryCount)
	return retryCount
}

func (t *TaskRunner) resetTaskRetryCount() {
	delete(t.task.Data, dexcom.DataKeyRetryCount)
}

func (t *TaskRunner) setDataTimeField(key string, value time.Time) {
	t.task.Data[key] = value.UTC().Format(time.RFC3339Nano)
}

func (t *TaskRunner) getDataTimeField(key string) *time.Time {
	if value, ok := t.task.Data[key].(string); !ok {
		return nil
	} else if tm, err := time.Parse(time.RFC3339Nano, value); err != nil {
		t.logger.WithError(err).WithField(key, value).Warn("Unable to parse data time")
		return nil
	} else {
		return pointer.From(tm)
	}
}

func (t *TaskRunner) HTTPClient(ctx context.Context, tokenSourceSource oauth.TokenSourceSource) (*http.Client, error) {
	return t.tokenSource.HTTPClient(ctx, tokenSourceSource)
}

func (t *TaskRunner) UpdateToken(ctx context.Context) (bool, error) {
	if updated, err := t.tokenSource.UpdateToken(ctx); err != nil || !updated {
		return updated, err
	}

	// Dexcom rotates the refresh token and accepts each only once, so the refresh just consumed the stored token and
	// losing its replacement permanently breaks the connection. The OAuth client only logs what this returns, so record
	// the failure for handleDexcomClientError to surface.
	err := request.RetryError(ctx, updateTokenRetrier, func(ctx context.Context) error {
		return t.updateProviderSession()
	})
	if err != nil {
		t.tokenUpdateError = err
	}
	return true, err
}

func (t *TaskRunner) ExpireToken(ctx context.Context) (bool, error) {
	return t.tokenSource.ExpireToken(ctx)
}

func InTimeRange(time time.Time, lower time.Time, upper time.Time) bool {
	if time.Before(lower) {
		return false
	} else if time.After(upper) {
		return false
	} else {
		return true
	}
}

func ClampTime(time time.Time, lower time.Time, upper time.Time) time.Time {
	if time.Before(lower) {
		return lower
	} else if time.After(upper) {
		return upper
	} else {
		return time
	}
}

// taskRetryCount accepts any numeric representation, since the count round trips as int32 through BSON, but as float64
// through JSON. Narrowing to one would silently restart the count and defeat TaskRetryCountMaximum. This will be
// unnecessary once Dexcom is moved to the work system (which has much better metadata support).
func taskRetryCount(valueRaw any) int {
	switch value := valueRaw.(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

type DataRange struct {
	StartTime time.Time
	EndTime   time.Time
}

func availableAfterDuration() time.Duration {
	return availableAfterDurationWithFallbackFactor(1)
}

func availableAfterDurationWithRetryCount(retryCount int) time.Duration {
	return availableAfterDurationWithFallbackFactor(fallbackFactorWithRetryCount(retryCount))
}

func availableAfterDurationWithFallbackFactor(fallbackFactor float64) time.Duration {
	return time.Duration(float64(AvailableAfterDuration)*fallbackFactor) + time.Duration(crypto.RandomInt64N(int64(2*AvailableAfterDurationJitter))) - AvailableAfterDurationJitter
}

func fallbackFactorWithRetryCount(retryCount int) float64 {
	return float64(int(1) << (retryCount - 1))
}

var updateTokenRetrier = request.NewRetrier(3, time.Second, 0.2)
