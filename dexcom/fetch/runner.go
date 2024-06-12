package fetch

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	dataClient "github.com/tidepool-org/platform/data/client"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/oauth"
	oauthToken "github.com/tidepool-org/platform/oauth/token"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/version"
)

const (
	AvailableAfterDurationMaximum = 135 * time.Minute
	AvailableAfterDurationMinimum = 105 * time.Minute
	DataSetSize                   = 2000
	TaskDurationMaximum           = 10 * time.Minute
)

var initialDataTime = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)

type Runner struct {
	logger           log.Logger
	versionReporter  version.Reporter
	authClient       auth.Client
	dataClient       dataClient.Client
	dataSourceClient dataSource.Client
	dexcomClient     dexcom.Client
}

func NewRunner(logger log.Logger, versionReporter version.Reporter, authClient auth.Client, dataClient dataClient.Client, dataSourceClient dataSource.Client, dexcomClient dexcom.Client) (*Runner, error) {
	if logger == nil {
		return nil, errors.New("logger is missing")
	}
	if versionReporter == nil {
		return nil, errors.New("version reporter is missing")
	}
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
		logger:           logger,
		versionReporter:  versionReporter,
		authClient:       authClient,
		dataClient:       dataClient,
		dataSourceClient: dataSourceClient,
		dexcomClient:     dexcomClient,
	}, nil
}

func (r *Runner) Logger() log.Logger {
	return r.logger
}

func (r *Runner) VersionReporter() version.Reporter {
	return r.versionReporter
}

func (r *Runner) AuthClient() auth.Client {
	return r.authClient
}

func (r *Runner) DataClient() dataClient.Client {
	return r.dataClient
}

func (r *Runner) DataSourceClient() dataSource.Client {
	return r.dataSourceClient
}

func (r *Runner) DexcomClient() dexcom.Client {
	return r.dexcomClient
}

func (r *Runner) GetRunnerType() string {
	return Type
}

func (r *Runner) GetRunnerDeadline() time.Time {
	return time.Now().Add(TaskDurationMaximum * 3)
}

func (r *Runner) GetRunnerMaximumDuration() time.Duration {
	return TaskDurationMaximum
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) bool {
	now := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.Logger())

	// HACK: Dexcom - skip 2:45am - 3:45am PST to avoid intermittent refresh token failure due to Dexcom backups (per Dexcom)
	var skipToAvoidDexcomBackup bool
	if location, err := time.LoadLocation("America/Los_Angeles"); err != nil {
		r.Logger().WithError(err).Warn("Unable to load location to detect Dexcom backup")
	} else {
		tm := now.In(location).Format("15:04:05")
		skipToAvoidDexcomBackup = (tm >= "02:45:00") && (tm < "03:45:00")
	}

	if !skipToAvoidDexcomBackup {
		tsk.ClearError()

		if serverSessionToken, sErr := r.AuthClient().ServerSessionToken(); sErr != nil {
			r.ignoreAndLogTaskError(tsk, errors.Wrap(sErr, "unable to get server session token"))
		} else {
			ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)
			if taskRunner, tErr := NewTaskRunner(r, tsk); tErr != nil {
				r.ignoreAndLogTaskError(tsk, errors.Wrap(sErr, "unable to create task runner"))
			} else if tErr = taskRunner.Run(ctx); tErr != nil {
				ErrorOrRetryTask(tsk, errors.Wrap(tErr, "unable to run task runner"))
			}
		}
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(AvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(AvailableAfterDurationMaximum-AvailableAfterDurationMinimum+1))))
	}

	if taskDuration := time.Since(now); taskDuration > TaskDurationMaximum {
		r.Logger().WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}
	return true
}

func (r *Runner) ignoreAndLogTaskError(tsk *task.Task, err error) {
	r.logger.Warnf("dexcom task %s error, task will be retried: %s", tsk.ID, err)
}

type TaskRunner struct {
	*Runner
	task             *task.Task
	context          context.Context
	validator        structure.Validator
	providerSession  *auth.ProviderSession
	dataSource       *dataSource.Source
	tokenSource      oauth.TokenSource
	deviceHashes     map[string]string
	dataSet          *data.DataSet
	dataSetPreloaded bool
}

func NewTaskRunner(rnnr *Runner, tsk *task.Task) (*TaskRunner, error) {
	if rnnr == nil {
		return nil, errors.New("runner is missing")
	}
	if tsk == nil {
		return nil, errors.New("task is missing")
	}

	return &TaskRunner{
		Runner: rnnr,
		task:   tsk,
	}, nil
}

func (t *TaskRunner) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is missing")
	}

	if len(t.task.Data) == 0 {
		return FailTask(t.logger, t.task, errors.New("data is missing"))
	}

	t.context = ctx
	t.validator = structureValidator.New()

	if err := t.getProviderSession(); err != nil {
		return err
	}
	if err := t.getDataSource(); err != nil {
		return err
	}
	if err := t.createTokenSource(); err != nil {
		return err
	}
	if err := t.getDeviceHashes(); err != nil {
		return err
	}
	if err := t.fetchSinceLatestDataTime(); err != nil {
		if request.IsErrorUnauthenticated(errors.Cause(err)) {
			FailTask(t.logger, t.task, err)
			if updateErr := t.updateDataSourceWithError(err); updateErr != nil {
				t.Logger().WithError(updateErr).Error("unable to update data source with error")
			}
		}
		return err
	}
	return t.updateDataSourceWithLastImportTime()
}

func (t *TaskRunner) getProviderSession() error {
	providerSessionID, ok := t.task.Data["providerSessionId"].(string)
	if !ok || providerSessionID == "" {
		return FailTask(t.logger, t.task, errors.New("provider session id is missing"))
	}

	providerSession, err := t.AuthClient().GetProviderSession(t.context, providerSessionID)
	if err != nil {
		return errors.Wrap(err, "unable to get provider session")
	} else if providerSession == nil {
		return FailTask(t.logger, t.task, errors.Wrap(err, "provider session is missing"))
	}
	t.providerSession = providerSession

	return nil
}

func (t *TaskRunner) updateProviderSession() error {
	refreshedToken, err := t.tokenSource.RefreshedToken()
	if err != nil {
		return errors.Wrap(err, "unable to get refreshed token")
	} else if refreshedToken == nil {
		return nil
	}

	updateProviderSession := auth.NewProviderSessionUpdate()
	updateProviderSession.OAuthToken = refreshedToken
	providerSession, err := t.AuthClient().UpdateProviderSession(t.context, t.providerSession.ID, updateProviderSession)
	if err != nil {
		return errors.Wrap(err, "unable to update provider session")
	} else if providerSession == nil {
		return FailTask(t.logger, t.task, errors.Wrap(err, "provider session is missing"))
	}
	t.providerSession = providerSession

	return nil
}

func (t *TaskRunner) getDataSource() error {
	dataSourceID, ok := t.task.Data["dataSourceId"].(string)
	if !ok || dataSourceID == "" {
		return FailTask(t.logger, t.task, errors.New("data source id is missing"))
	}

	source, err := t.DataSourceClient().Get(t.context, dataSourceID)
	if err != nil {
		return errors.Wrap(err, "unable to get data source")
	} else if source == nil {
		return FailTask(t.logger, t.task, errors.Wrap(err, "data source is missing"))
	}
	t.dataSource = source

	return nil
}

func (t *TaskRunner) updateDataSourceWithDataSet(dataSet *data.DataSet) error {
	update := dataSource.NewUpdate()
	update.DataSetIDs = pointer.FromStringArray(append(pointer.ToStringArray(t.dataSource.DataSetIDs), *dataSet.UploadID))
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithDataTime(earliestDataTime *time.Time, latestDataTime *time.Time) error {
	update := dataSource.NewUpdate()

	if t.beforeEarliestDataTime(earliestDataTime) {
		update.EarliestDataTime = earliestDataTime
	}
	if t.afterLatestDataTime(latestDataTime) {
		update.LatestDataTime = latestDataTime
	}

	if update.EarliestDataTime == nil && update.LatestDataTime == nil {
		return nil
	}

	update.LastImportTime = pointer.FromTime(time.Now())
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithLastImportTime() error {
	update := dataSource.NewUpdate()
	update.LastImportTime = pointer.FromTime(time.Now())
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithError(err error) error {
	update := dataSource.NewUpdate()
	update.State = pointer.FromString(dataSource.StateError)
	update.Error = errors.NewSerializable(err)
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSource(update *dataSource.Update) error {
	if update.IsEmpty() {
		return nil
	}

	source, err := t.DataSourceClient().Update(t.context, *t.dataSource.ID, nil, update)
	if err != nil {
		return errors.Wrap(err, "unable to update data source")
	} else if source == nil {
		return FailTask(t.logger, t.task, errors.Wrap(err, "data source is missing"))
	}

	t.dataSource = source
	return nil
}

func (t *TaskRunner) createTokenSource() error {
	tokenSource, err := oauthToken.NewSourceWithToken(t.providerSession.OAuthToken)
	if err != nil {
		return FailTask(t.logger, t.task, errors.Wrap(err, "unable to create token source"))
	}

	t.tokenSource = tokenSource
	return nil
}

func (t *TaskRunner) getDeviceHashes() error {
	raw, rawOK := t.task.Data["deviceHashes"]
	if !rawOK || raw == nil {
		return nil
	}
	rawMap, rawMapOK := raw.(map[string]interface{})
	if !rawMapOK || rawMap == nil {
		return FailTask(t.logger, t.task, errors.New("device hashes is invalid"))
	}
	deviceHashes := map[string]string{}
	for key, value := range rawMap {
		if valueString, valueStringOK := value.(string); valueStringOK {
			deviceHashes[key] = valueString
		} else {
			return FailTask(t.logger, t.task, errors.New("device hash is invalid"))
		}
	}

	t.deviceHashes = deviceHashes
	return nil
}

func (t *TaskRunner) updateDeviceHash(device *dexcom.Device) bool {
	deviceHash, err := device.Hash()
	if err != nil {
		return false
	}

	if t.deviceHashes == nil {
		t.deviceHashes = map[string]string{}
	}

	if device.TransmitterID != nil && t.deviceHashes[*device.TransmitterID] != deviceHash {
		t.deviceHashes[*device.TransmitterID] = deviceHash
		return true
	}

	return false
}

func (t *TaskRunner) updateDataSetWithTimezoneOffset(timezoneOffset *int) error {
	if timezoneOffset == nil {
		return nil
	}
	return t.updateDataSet(&data.DataSetUpdate{TimeZoneOffset: timezoneOffset})
}

func (t *TaskRunner) updateDataSet(dataSetUpdate *data.DataSetUpdate) error {
	if dataSetUpdate.IsEmpty() {
		return nil
	}

	dataSet, err := t.DataClient().UpdateDataSet(t.context, *t.dataSet.UploadID, dataSetUpdate)
	if err != nil {
		return errors.Wrap(err, "unable to update data set")
	} else if dataSet == nil {
		return FailTask(t.logger, t.task, errors.Wrap(err, "data set is missing"))
	}

	t.dataSet = dataSet
	return nil
}

func (t *TaskRunner) fetchSinceLatestDataTime() error {
	startTime := initialDataTime
	if t.dataSource.LatestDataTime != nil && startTime.Before(*t.dataSource.LatestDataTime) {
		startTime = *t.dataSource.LatestDataTime
	}

	almostNow := time.Now().Add(-time.Minute)
	for startTime.Before(almostNow) {
		endTime := startTime.AddDate(0, 0, 30)
		if endTime.After(almostNow) {
			endTime = almostNow
		}

		if err := t.fetch(startTime, endTime); err != nil {
			return err
		}

		startTime = startTime.AddDate(0, 0, 30)
		almostNow = time.Now().Add(-time.Minute)
	}
	return nil
}

func (t *TaskRunner) fetch(startTime time.Time, endTime time.Time) error {
	devices, devicesDatumArray, err := t.fetchDevices(startTime, endTime)
	if err != nil {
		return err
	}

	// HACK: Dexcom - does not guarantee to return a device for G5 Mobile if time range < 24 hours (per Dexcom)
	if endTime.Sub(startTime) > 24*time.Hour {
		if len(*devices) == 0 {
			return nil
		}
	} else {
		if err = t.preloadDataSet(); err != nil {
			return err
		} else if t.dataSet == nil {
			return nil
		}
	}

	datumArray, err := t.fetchData(startTime, endTime)
	if err != nil {
		return err
	}

	if len(datumArray) == 0 && len(devicesDatumArray) == 0 {
		return nil
	}

	if err = t.prepareDataSet(); err != nil {
		return err
	}

	if err = t.storeDatumArray(datumArray); err != nil {
		return err
	}

	if err = t.storeDevicesDatumArray(devicesDatumArray); err != nil {
		return err
	}

	return nil
}

func (t *TaskRunner) fetchDevices(startTime time.Time, endTime time.Time) (*dexcom.Devices, data.Data, error) {
	response, err := t.DexcomClient().GetDevices(t.context, startTime, endTime, t.tokenSource)
	if updateErr := t.updateProviderSession(); updateErr != nil {
		return nil, nil, updateErr
	}
	if err != nil {
		return nil, nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, nil, err
	}

	devices := response.Devices

	var devicesDatumArray data.Data
	for _, device := range *devices {
		if t.updateDeviceHash(device) {
			devicesDatumArray = append(devicesDatumArray, translateDeviceToDatum(device))
		}
	}

	return devices, devicesDatumArray, nil
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

	fetchDatumArray, err := t.fetchAlerts(startTime, endTime)
	if err != nil {
		return nil, err
	}
	datumArray = append(datumArray, fetchDatumArray...)

	fetchDatumArray, err = t.fetchCalibrations(startTime, endTime)
	if err != nil {
		return nil, err
	}
	datumArray = append(datumArray, fetchDatumArray...)

	fetchDatumArray, err = t.fetchEGVs(startTime, endTime)
	if err != nil {
		return nil, err
	}

	datumArray = append(datumArray, fetchDatumArray...)

	fetchDatumArray, err = t.fetchEvents(startTime, endTime)
	if err != nil {
		return nil, err
	}
	datumArray = append(datumArray, fetchDatumArray...)

	sort.Sort(ByTime(datumArray))

	return datumArray, nil
}

func (t *TaskRunner) fetchCalibrations(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetCalibrations(t.context, startTime, endTime, t.tokenSource)
	if updateErr := t.updateProviderSession(); updateErr != nil {
		return nil, updateErr
	}
	if err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}

	datumArray := data.Data{}
	for _, c := range *response.Calibrations {
		if t.afterLatestDataTime(c.SystemTime.Raw()) {
			datumArray = append(datumArray, translateCalibrationToDatum(c))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchAlerts(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetAlerts(t.context, startTime, endTime, t.tokenSource)
	if updateErr := t.updateProviderSession(); updateErr != nil {
		return nil, updateErr
	}
	if err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}
	datumArray := data.Data{}
	for _, c := range *response.Alerts {
		if t.afterLatestDataTime(c.SystemTime.Raw()) {
			datumArray = append(datumArray, translateAlertToDatum(c, response.RecordVersion))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEGVs(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetEGVs(t.context, startTime, endTime, t.tokenSource)

	if updateErr := t.updateProviderSession(); updateErr != nil {
		return nil, updateErr
	}
	if err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}

	datumArray := data.Data{}
	for _, e := range *response.EGVs {
		if t.afterLatestDataTime(e.SystemTime.Raw()) {
			datumArray = append(datumArray, translateEGVToDatum(e))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEvents(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetEvents(t.context, startTime, endTime, t.tokenSource)
	if updateErr := t.updateProviderSession(); updateErr != nil {
		return nil, updateErr
	}
	if err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}

	datumArray := data.Data{}
	for _, e := range *response.Events {
		switch *e.Status {
		case dexcom.EventStatusCreated:
			if t.afterLatestDataTime(e.SystemTime.Raw()) {
				switch *e.Type {
				case dexcom.EventTypeCarbs:
					datumArray = append(datumArray, translateEventCarbsToDatum(e))
				case dexcom.EventTypeExercise:
					datumArray = append(datumArray, translateEventExerciseToDatum(e))
				case dexcom.EventTypeHealth:
					datumArray = append(datumArray, translateEventHealthToDatum(e))
				case dexcom.EventTypeInsulin:
					datumArray = append(datumArray, translateEventInsulinToDatum(e))
				case dexcom.EventTypeBG:
					datumArray = append(datumArray, translateEventBGToDatum(e))
				case dexcom.EventTypeNote, dexcom.EventTypeNotes:
					datumArray = append(datumArray, translateEventNoteToDatum(e))
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

	if t.dataSet != nil {
		return nil
	}

	dataSet, err := t.createDataSet()
	if err != nil {
		return err
	}
	t.dataSet = dataSet
	return nil
}

func (t *TaskRunner) findDataSet() (*data.DataSet, error) {
	if t.dataSource.DataSetIDs != nil {
		for index := len(*t.dataSource.DataSetIDs) - 1; index >= 0; index-- {
			if dataSet, err := t.DataClient().GetDataSet(t.context, (*t.dataSource.DataSetIDs)[index]); err != nil {
				return nil, errors.Wrap(err, "unable to get data set")
			} else if dataSet != nil {
				return dataSet, nil
			}
		}
	}
	return nil, nil
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
	dataSetCreate.DeviceManufacturers = pointer.FromStringArray([]string{"Dexcom"})
	dataSetCreate.DeviceTags = pointer.FromStringArray([]string{data.DeviceTagCGM})
	dataSetCreate.Time = pointer.FromTime(time.Now())
	dataSetCreate.TimeProcessing = pointer.FromString(dataTypesUpload.TimeProcessingNone)

	dataSet, err := t.DataClient().CreateUserDataSet(t.context, t.providerSession.UserID, dataSetCreate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data set")
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

		if err := t.DataClient().CreateDataSetsData(t.context, *t.dataSet.UploadID, datumArray[startIndex:endIndex]); err != nil {
			return errors.Wrap(err, "unable to create data set data")
		}

		earliestDataTime := datumArray[0].GetTime()
		latestDataTime := datumArray[endIndex-1].GetTime()
		if err := t.updateDataSourceWithDataTime(earliestDataTime, latestDataTime); err != nil {
			return err
		}

		// Determine last known timezone offset and persist with the data set
		var timezoneOffset *int
		for index := endIndex - 1; index >= 0; index-- {
			if timezoneOffset = datumArray[index].GetTimeZoneOffset(); timezoneOffset != nil {
				break
			}
		}
		if err := t.updateDataSetWithTimezoneOffset(timezoneOffset); err != nil {
			return err
		}
	}

	return nil
}

func (t *TaskRunner) storeDevicesDatumArray(devicesDatumArray data.Data) error {
	if len(devicesDatumArray) > 0 {
		if err := t.DataClient().CreateDataSetsData(t.context, *t.dataSet.UploadID, devicesDatumArray); err != nil {
			return errors.Wrap(err, "unable to create data set data")
		}
		t.task.Data["deviceHashes"] = t.deviceHashes
	}
	return nil
}

func (t *TaskRunner) beforeEarliestDataTime(earliestDataTime *time.Time) bool {
	return earliestDataTime != nil && (t.dataSource.EarliestDataTime == nil || earliestDataTime.Before(*t.dataSource.EarliestDataTime))
}

func (t *TaskRunner) afterLatestDataTime(latestDataTime *time.Time) bool {
	return latestDataTime != nil && (t.dataSource.LatestDataTime == nil || latestDataTime.After(*t.dataSource.LatestDataTime))
}

type ByTime data.Data

func (b ByTime) Len() int {
	return len(b)
}

func (b ByTime) Less(left int, right int) bool {
	if leftTime := b[left].GetTime(); leftTime == nil {
		return true
	} else if rightTime := b[right].GetTime(); rightTime == nil {
		return false
	} else {
		return leftTime.Before(*rightTime)
	}
}

func (b ByTime) Swap(left int, right int) {
	b[left], b[right] = b[right], b[left]
}
