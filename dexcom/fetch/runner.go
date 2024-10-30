package fetch

import (
	"context"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
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
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
)

//go:generate mockgen -destination=./test/mock.go -package test . AuthClient,DataClient,DataSourceClient,DexcomClient,Provider

const (
	AvailableAfterDuration       = 120 * time.Minute
	AvailableAfterDurationJitter = 15 * time.Minute
	DataSetSize                  = 2000
	TaskDurationMaximum          = 15 * time.Minute
	TaskRetryCountMaximum        = 4  // Last retry after ~(AvailableAfterDuration * (2^TaskRetryCountMaximum - 1)) hours (discounting AvailableAfterDurationJitter)
	DataRangeDaysMaximum         = 30 // Per Dexcom documentation
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

type DataSourceClient interface {
	Get(ctx context.Context, id string) (*dataSource.Source, error)
	Update(ctx context.Context, id string, condition *request.Condition, create *dataSource.Update) (*dataSource.Source, error)
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
	dataSourceClient DataSourceClient
	dexcomClient     DexcomClient
}

func NewRunner(authClient AuthClient, dataClient DataClient, dataSourceClient DataSourceClient, dexcomClient DexcomClient) (*Runner, error) {
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

func (r *Runner) DataSourceClient() DataSourceClient {
	return r.dataSourceClient
}

func (r *Runner) DexcomClient() DexcomClient {
	return r.dexcomClient
}

func (r *Runner) GetRunnerType() string {
	return Type
}

func (r *Runner) GetRunnerDeadline() time.Time {
	return time.Now().Add(TaskDurationMaximum * 3)
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
	DataSourceClient() DataSourceClient
	DexcomClient() DexcomClient
	GetRunnerDurationMaximum() time.Duration
}

type TaskRunner struct {
	Provider
	task             *task.Task
	context          context.Context
	logger           log.Logger
	providerSession  *auth.ProviderSession
	dataSource       *dataSource.Source
	tokenSource      oauth.TokenSource
	deviceHashes     map[string]string
	dataSet          *data.DataSet
	dataSetPreloaded bool
	deadline         time.Time
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
	t.deadline = time.Now().Add(t.GetRunnerDurationMaximum())

	t.task.ClearError()
	if err := t.run(); err == nil {
		t.rescheduleTask()
	} else if !t.task.HasError() {
		t.rescheduleTaskWithResourceError(err)
	}
}

func (t *TaskRunner) run() error {
	defer t.updateDataSourceWithTaskState()

	if len(t.task.Data) == 0 {
		return t.failTaskWithInvalidStateError(errors.New("data is missing"))
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
	if err := t.updateDataSourceWithLastImportTime(); err != nil {
		return err
	}

	return nil
}

func (t *TaskRunner) getProviderSession() error {
	providerSessionID, ok := t.task.Data[dexcom.DataKeyProviderSessionID].(string)
	if !ok || providerSessionID == "" {
		return t.failTaskWithInvalidStateError(errors.New("provider session id is missing"))
	}

	providerSession, err := t.AuthClient().GetProviderSession(t.context, providerSessionID)
	if err != nil {
		return t.rescheduleTaskWithResourceError(errors.Wrap(err, "unable to get provider session"))
	} else if providerSession == nil {
		return t.failTaskWithInvalidStateError(errors.New("provider session is missing"))
	}
	t.providerSession = providerSession

	return nil
}

func (t *TaskRunner) updateProviderSession() error {
	refreshedToken, err := t.tokenSource.RefreshedToken()
	if err != nil {
		return t.retryOrRescheduleTaskWithDexcomClientError(errors.Wrap(err, "unable to get refreshed token"))
	} else if refreshedToken == nil {
		return nil // Token still valid, but has not changed, so no need to update
	}

	// Without cancel to ensure provider session is updated in the database
	updateProviderSession := auth.NewProviderSessionUpdate()
	updateProviderSession.OAuthToken = refreshedToken
	providerSession, err := t.AuthClient().UpdateProviderSession(context.WithoutCancel(t.context), t.providerSession.ID, updateProviderSession)
	if err != nil {
		return t.rescheduleTaskWithResourceError(errors.WithMeta(errors.Wrap(err, "unable to update provider session"), updateProviderSession))
	} else if providerSession == nil {
		return t.failTaskWithInvalidStateError(errors.New("provider session is missing"))
	}
	t.providerSession = providerSession

	return nil
}

func (t *TaskRunner) getDataSource() error {
	dataSourceID, ok := t.task.Data[dexcom.DataKeyDataSourceID].(string)
	if !ok || dataSourceID == "" {
		return t.failTaskWithInvalidStateError(errors.New("data source id is missing"))
	}

	source, err := t.DataSourceClient().Get(t.context, dataSourceID)
	if err != nil {
		return t.rescheduleTaskWithResourceError(errors.Wrap(err, "unable to get data source"))
	} else if source == nil {
		return t.failTaskWithInvalidStateError(errors.New("data source is missing"))
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

	update.LastImportTime = pointer.FromTime(time.Now())
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithLastImportTime() error {
	update := dataSource.NewUpdate()
	update.LastImportTime = pointer.FromTime(time.Now())
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithTaskState() error {
	update := dataSource.NewUpdate()
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

	// Without cancel to ensure data source is updated in the database
	dataSource, err := t.DataSourceClient().Update(context.WithoutCancel(t.context), *t.dataSource.ID, nil, update)
	if err != nil {
		return t.rescheduleTaskWithResourceError(errors.WithMeta(errors.Wrap(err, "unable to update data source"), update))
	} else if dataSource == nil {
		return t.failTaskWithInvalidStateError(errors.New("data source is missing"))
	}

	t.dataSource = dataSource
	return nil
}

func (t *TaskRunner) createTokenSource() error {
	tokenSource, err := oauthToken.NewSourceWithToken(t.providerSession.OAuthToken)
	if err != nil {
		return t.failTaskWithInvalidStateError(errors.Wrap(err, "unable to create token source"))
	}

	t.tokenSource = tokenSource
	return nil
}

func (t *TaskRunner) getDeviceHashes() error {
	raw, rawOK := t.task.Data[dexcom.DataKeyDeviceHashes]
	if !rawOK || raw == nil {
		return nil
	}
	rawMap, rawMapOK := raw.(map[string]interface{})
	if !rawMapOK || rawMap == nil {
		return t.failTaskWithInvalidStateError(errors.New("device hashes is invalid"))
	}
	deviceHashes := map[string]string{}
	for key, value := range rawMap {
		if valueString, valueStringOK := value.(string); valueStringOK {
			deviceHashes[key] = valueString
		} else {
			return t.failTaskWithInvalidStateError(errors.New("device hash is invalid"))
		}
	}

	t.deviceHashes = deviceHashes
	return nil
}

func (t *TaskRunner) updateDeviceHash(device *dexcom.Device) bool {
	deviceID := device.ID()
	deviceHash, err := device.Hash()
	if err != nil {
		return false
	}

	if t.deviceHashes == nil {
		t.deviceHashes = map[string]string{}
	}

	// If the device hash has not changed, then no need to update
	if existingDeviceHash, ok := t.deviceHashes[deviceID]; ok && existingDeviceHash == deviceHash {
		return false
	}

	t.deviceHashes[deviceID] = deviceHash
	return true
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

	// Without cancel to ensure data set is updated in the database
	dataSet, err := t.DataClient().UpdateDataSet(context.WithoutCancel(t.context), *t.dataSet.UploadID, update)
	if err != nil {
		return t.rescheduleTaskWithResourceError(errors.WithMeta(errors.Wrap(err, "unable to update data set"), update))
	} else if dataSet == nil {
		return t.failTaskWithInvalidStateError(errors.New("data set is missing"))
	}

	t.dataSet = dataSet
	return nil
}

func (t *TaskRunner) fetchSinceLatestDataTime() error {
	dataRange, err := t.fetchDataRange()
	if err != nil {
		return err
	} else if dataRange == nil {
		return nil // Nothing to fetch
	}

	startTime := dataRange.StartTime
	for startTime.Before(dataRange.EndTime) {
		endTime := startTime.AddDate(0, 0, DataRangeDaysMaximum)
		if endTime.After(dataRange.EndTime) {
			endTime = dataRange.EndTime
		}

		if err := t.fetch(startTime, endTime); err != nil {
			return err
		}

		// If past deadline (based upon runner maximum duration), then bail
		if time.Now().After(t.deadline) {
			return t.rescheduleTaskWithResourceError(context.DeadlineExceeded)
		}

		startTime = startTime.AddDate(0, 0, DataRangeDaysMaximum)
	}

	return t.updateDataSourceWithLastImportTime()
}

func (t *TaskRunner) fetchDataRange() (*DataRange, error) {

	// HACK: Dexcom V3 (2024-05-30) - Can only use latest data time as last sync time if not
	// older than 100 days, otherwise will return erroneous results. Use 30 days to be on
	// the safe side.
	var lastSyncTime *time.Time
	if t.dataSource.LatestDataTime != nil && time.Now().Before(t.dataSource.LatestDataTime.AddDate(0, 0, DataRangeDaysMaximum)) {
		lastSyncTime = t.dataSource.LatestDataTime
	}

	response, err := t.DexcomClient().GetDataRange(t.context, lastSyncTime, t.tokenSource)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
	}

	// Get data range, if none valid, then indicate nothing to fetch
	dataRange := response.DataRange()
	if dataRange == nil {
		return nil, nil
	}

	// Clamp data range, if none valid, then indicate nothing to fetch
	latestDataTime := *pointer.DefaultTime(t.dataSource.LatestDataTime, initialDataTime)
	now := time.Now()
	startTime := ClampTime(*dataRange.Start.SystemTimeRaw(), latestDataTime, now)
	endTime := ClampTime(*dataRange.End.SystemTimeRaw(), latestDataTime, now)
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

	return t.storeDatumArray(datumArray)
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

	fetchDatumArray, err = t.fetchDevices(startTime, endTime)
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

	sort.Stable(DataByTime(datumArray))

	return datumArray, nil
}

func (t *TaskRunner) fetchAlerts(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetAlerts(t.context, startTime, endTime, t.tokenSource)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
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
	response, err := t.DexcomClient().GetCalibrations(t.context, startTime, endTime, t.tokenSource)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
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
	response, err := t.DexcomClient().GetDevices(t.context, startTime, endTime, t.tokenSource)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
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
		if time := device.LastUploadDate.Raw(); time != nil && InTimeRange(*time, startTime, endTime) && t.updateDeviceHash(device) {
			datumArray = append(datumArray, translateDeviceToDatum(t.context, device))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEGVs(startTime time.Time, endTime time.Time) (data.Data, error) {
	response, err := t.DexcomClient().GetEGVs(t.context, startTime, endTime, t.tokenSource)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
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
	response, err := t.DexcomClient().GetEvents(t.context, startTime, endTime, t.tokenSource)
	if err = t.handleDexcomClientError(err); err != nil {
		return nil, err
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
				return nil, t.rescheduleTaskWithResourceError(errors.Wrap(err, "unable to get data set"))
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
	dataSetCreate.Deduplicator.Version = pointer.FromString(dataDeduplicatorDeduplicator.NoneVersion)
	dataSetCreate.DeviceManufacturers = pointer.FromStringArray([]string{"Dexcom"})
	dataSetCreate.DeviceTags = pointer.FromStringArray([]string{data.DeviceTagCGM})
	dataSetCreate.Time = pointer.FromTime(time.Now())
	dataSetCreate.TimeProcessing = pointer.FromString(dataTypesUpload.TimeProcessingNone)

	dataSet, err := t.DataClient().CreateUserDataSet(t.context, t.providerSession.UserID, dataSetCreate)
	if err != nil {
		return nil, t.rescheduleTaskWithResourceError(errors.WithMeta(errors.Wrap(err, "unable to create data set"), dataSetCreate))
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
			return t.rescheduleTaskWithResourceError(errors.Wrap(err, "unable to create data set data"))
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

	t.task.Data[dexcom.DataKeyDeviceHashes] = t.deviceHashes

	return nil
}

func (t *TaskRunner) beforeEarliestDataTime(earliestDataTime *time.Time) bool {
	return earliestDataTime != nil && (t.dataSource.EarliestDataTime == nil || earliestDataTime.Before(*t.dataSource.EarliestDataTime))
}

func (t *TaskRunner) afterLatestDataTime(latestDataTime *time.Time) bool {
	return latestDataTime != nil && (t.dataSource.LatestDataTime == nil || latestDataTime.After(*t.dataSource.LatestDataTime))
}

// Handle potential dexcom client error. Update provider session with latest token.
// If error, then retry or reschedule. Otherwise, reset retry count.
func (t *TaskRunner) handleDexcomClientError(err error) error {
	if updateErr := t.updateProviderSession(); updateErr != nil {
		return updateErr
	}
	if err != nil {
		return t.retryOrRescheduleTaskWithDexcomClientError(err)
	} else {
		return t.resetTaskRetryCount()
	}
}

// Retry task if Dexcom authentication failure. Otherwise, reschedule task.
func (t *TaskRunner) retryOrRescheduleTaskWithDexcomClientError(err error) error {
	if request.IsErrorUnauthenticated(errors.Cause(err)) {
		return t.retryTaskWithError(ErrorAuthenticationFailureError(err))
	} else {
		return t.rescheduleTaskWithResourceError(err)
	}
}

// Increment task retry count. If task retry count exceeds maximum, then fail task.
// Otherwise, reschedule task. Typically used for Dexcom authentication failures that
// may or may not be transient.
func (t *TaskRunner) retryTaskWithError(err error) error {
	retryCount := t.incrementTaskRetryCount()
	if retryCount > TaskRetryCountMaximum {
		return t.failTaskWithError(err)
	}

	t.task.AppendError(err)
	t.task.RepeatAvailableAfter(availableAfterDurationWithFallbackFactor(fallbackFactorWithRetryCount(retryCount)))
	return err
}

func (t *TaskRunner) rescheduleTaskWithResourceError(err error) error {
	return t.rescheduleTaskWithError(ErrorResourceFailureError(err))
}

// Reschedule task for next run. Append error to task.
func (t *TaskRunner) rescheduleTaskWithError(err error) error {
	t.task.AppendError(err)
	t.rescheduleTask()
	return err
}

func (t *TaskRunner) rescheduleTask() {
	t.task.RepeatAvailableAfter(availableAfterDuration())
}

func (t *TaskRunner) failTaskWithInvalidStateError(err error) error {
	return t.failTaskWithError(ErrorInvalidStateError(err))
}

// Fail task immediately and permanently. Do not reschedule. For situations where any future attempt is
// also guaranteed to fail. For example, when the task data is missing information. Should not normally happen.
func (t *TaskRunner) failTaskWithError(err error) error {
	t.task.AppendError(err)
	t.task.SetFailed()
	return err
}

func (t *TaskRunner) incrementTaskRetryCount() int {
	retryCount := 1
	if valueRaw, ok := t.task.Data[dexcom.DataKeyRetryCount]; ok && valueRaw != nil {
		if value, ok := valueRaw.(int32); ok {
			retryCount = int(value) + 1
		}
	}
	t.task.Data[dexcom.DataKeyRetryCount] = retryCount
	return retryCount
}

func (t *TaskRunner) resetTaskRetryCount() error {
	delete(t.task.Data, dexcom.DataKeyRetryCount)
	return nil
}

type DataByTime data.Data

func (d DataByTime) Len() int {
	return len(d)
}

func (d DataByTime) Less(left int, right int) bool {
	if leftTime := d[left].GetTime(); leftTime == nil {
		return true
	} else if rightTime := d[right].GetTime(); rightTime == nil {
		return false
	} else {
		return leftTime.Before(*rightTime)
	}
}

func (d DataByTime) Swap(left int, right int) {
	d[left], d[right] = d[right], d[left]
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

type DataRange struct {
	StartTime time.Time
	EndTime   time.Time
}

func availableAfterDuration() time.Duration {
	return availableAfterDurationWithFallbackFactor(1)
}

func availableAfterDurationWithFallbackFactor(fallbackFactor float64) time.Duration {
	return time.Duration(float64(AvailableAfterDuration)*fallbackFactor) + time.Duration(rand.Int63n(int64(2*AvailableAfterDurationJitter))) - AvailableAfterDurationJitter
}

func fallbackFactorWithRetryCount(retryCount int) float64 {
	return float64(int(1) << (retryCount - 1))
}
