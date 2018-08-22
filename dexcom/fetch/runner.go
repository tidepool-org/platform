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
	"github.com/tidepool-org/platform/data/types/upload"
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
	AvailableAfterDurationMaximum = 75 * time.Minute
	AvailableAfterDurationMinimum = 45 * time.Minute
	DataSetSize                   = 2000
	TaskDurationMaximum           = 5 * time.Minute
)

var initialDataTime = time.Unix(1420070400, 0) // 2015-01-01T00:00:00Z

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

func (r *Runner) CanRunTask(tsk *task.Task) bool {
	return tsk != nil && tsk.Type == Type
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) {
	taskStartTime := time.Now()

	ctx = log.NewContextWithLogger(ctx, r.Logger())

	// HACK: Dexcom - skip 2:45am - 3:45am PST to avoid intermittent refresh token failure due to Dexcom backups (per Dexcom)
	var skipToAvoidDexcomBackup bool
	if location, err := time.LoadLocation("America/Los_Angeles"); err != nil {
		r.Logger().WithError(err).Warn("Unable to load location to detect Dexcom backup")
	} else {
		tm := time.Now().In(location).Format("15:04:05")
		skipToAvoidDexcomBackup = (tm >= "02:45:00") && (tm < "03:45:00")
	}

	if !skipToAvoidDexcomBackup {
		tsk.ClearError()

		if serverSessionToken, sErr := r.AuthClient().ServerSessionToken(); sErr != nil {
			tsk.AppendError(errors.Wrap(sErr, "unable to get server session token"))
		} else {
			ctx = auth.NewContextWithServerSessionToken(ctx, serverSessionToken)

			if taskRunner, tErr := NewTaskRunner(r, tsk); tErr != nil {
				tsk.AppendError(errors.Wrap(tErr, "unable to create task runner"))
			} else if tErr = taskRunner.Run(ctx); tErr != nil {
				tsk.AppendError(errors.Wrap(tErr, "unable to run task runner"))
			}
		}
	}

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(AvailableAfterDurationMinimum + time.Duration(rand.Int63n(int64(AvailableAfterDurationMaximum-AvailableAfterDurationMinimum+1))))
	}

	if taskDuration := time.Since(taskStartTime); taskDuration > TaskDurationMaximum {
		r.Logger().WithField("taskDuration", taskDuration.Truncate(time.Millisecond).Seconds()).Warn("Task duration exceeds maximum")
	}
}

type TaskRunner struct {
	*Runner
	task             *task.Task
	context          context.Context
	validator        structure.Validator
	providerSession  *auth.ProviderSession
	dataSource       *dataSource.Source
	tokenSource      oauth.TokenSource
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
		t.task.SetFailed()
		return errors.New("data is missing")
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
	if err := t.fetchSinceLatestDataTime(); err != nil {
		if request.IsErrorUnauthenticated(errors.Cause(err)) {
			t.task.SetFailed()
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
		t.task.SetFailed()
		return errors.New("provider session id is missing")
	}

	providerSession, err := t.AuthClient().GetProviderSession(t.context, providerSessionID)
	if err != nil {
		return errors.Wrap(err, "unable to get provider session")
	} else if providerSession == nil {
		t.task.SetFailed()
		return errors.Wrap(err, "provider session is missing")
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
		t.task.SetFailed()
		return errors.Wrap(err, "provider session is missing")
	}
	t.providerSession = providerSession

	return nil
}

func (t *TaskRunner) getDataSource() error {
	dataSourceID, ok := t.task.Data["dataSourceId"].(string)
	if !ok || dataSourceID == "" {
		t.task.SetFailed()
		return errors.New("data source id is missing")
	}

	source, err := t.DataSourceClient().Get(t.context, dataSourceID)
	if err != nil {
		return errors.Wrap(err, "unable to get data source")
	} else if source == nil {
		t.task.SetFailed()
		return errors.Wrap(err, "data source is missing")
	}
	t.dataSource = source

	return nil
}

func (t *TaskRunner) updateDataSourceWithDataSet(dataSet *data.DataSet) error {
	update := dataSource.NewUpdate()
	update.DataSetIDs = pointer.FromStringArray(append(pointer.ToStringArray(t.dataSource.DataSetIDs), *dataSet.UploadID))
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithDataTime(earliestDataTime time.Time, latestDataTime time.Time) error {
	update := dataSource.NewUpdate()

	if t.beforeEarliestDataTime(earliestDataTime) {
		update.EarliestDataTime = pointer.FromTime(earliestDataTime.Truncate(time.Second))
	}
	if t.afterLatestDataTime(latestDataTime) {
		update.LatestDataTime = pointer.FromTime(latestDataTime.Truncate(time.Second))
	}

	if update.EarliestDataTime == nil && update.LatestDataTime == nil {
		return nil
	}

	update.LastImportTime = pointer.FromTime(time.Now().Truncate(time.Second))
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithLastImportTime() error {
	update := dataSource.NewUpdate()
	update.LastImportTime = pointer.FromTime(time.Now().Truncate(time.Second))
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSourceWithError(err error) error {
	update := dataSource.NewUpdate()
	update.State = pointer.FromString(dataSource.StateError)
	update.Error = errors.NewSerializable(err)
	return t.updateDataSource(update)
}

func (t *TaskRunner) updateDataSource(update *dataSource.Update) error {
	if !update.HasUpdates() {
		return nil
	}

	source, err := t.DataSourceClient().Update(t.context, *t.dataSource.ID, nil, update)
	if err != nil {
		return errors.Wrap(err, "unable to update data source")
	} else if source == nil {
		t.task.SetFailed()
		return errors.Wrap(err, "data source is missing")
	}

	t.dataSource = source
	return nil
}

func (t *TaskRunner) createTokenSource() error {
	tokenSource, err := oauthToken.NewSourceWithToken(t.providerSession.OAuthToken)
	if err != nil {
		t.task.SetFailed()
		return errors.Wrap(err, "unable to create token source")
	}

	t.tokenSource = tokenSource
	return nil
}

func (t *TaskRunner) updateDataSetWithDeviceInfo(deviceInfo *DeviceInfo) error {
	dataSetDeviceInfo, err := NewDeviceInfoFromDataSet(t.dataSet)
	if err != nil {
		return err
	}
	dataSetDeviceInfo, err = dataSetDeviceInfo.Merge(deviceInfo)
	if err != nil {
		return err
	}

	dataSetUpdate := data.NewDataSetUpdate()
	if t.dataSet.DeviceID == nil || *t.dataSet.DeviceID != dataSetDeviceInfo.DeviceID {
		dataSetUpdate.DeviceID = pointer.FromString(dataSetDeviceInfo.DeviceID)
	}
	if t.dataSet.DeviceModel == nil || *t.dataSet.DeviceModel != dataSetDeviceInfo.DeviceModel {
		dataSetUpdate.DeviceModel = pointer.FromString(dataSetDeviceInfo.DeviceModel)
	}
	if t.dataSet.DeviceSerialNumber == nil || *t.dataSet.DeviceSerialNumber != dataSetDeviceInfo.DeviceSerialNumber {
		dataSetUpdate.DeviceSerialNumber = pointer.FromString(dataSetDeviceInfo.DeviceSerialNumber)
	}
	return t.updateDataSet(dataSetUpdate)
}

func (t *TaskRunner) updateDataSet(dataSetUpdate *data.DataSetUpdate) error {
	if !dataSetUpdate.HasUpdates() {
		return nil
	}

	dataSet, err := t.DataClient().UpdateDataSet(t.context, *t.dataSet.UploadID, dataSetUpdate)
	if err != nil {
		return errors.Wrap(err, "unable to update data set")
	} else if dataSet == nil {
		t.task.SetFailed()
		return errors.Wrap(err, "data set is missing")
	}

	t.dataSet = dataSet
	return nil
}

func (t *TaskRunner) fetchSinceLatestDataTime() error {
	startTime := initialDataTime
	if t.dataSource.LatestDataTime != nil && startTime.Before(*t.dataSource.LatestDataTime) {
		startTime = *t.dataSource.LatestDataTime
	}

	now := time.Now().Add(-time.Minute).Truncate(time.Second)
	for startTime.Before(now) {
		endTime := startTime.AddDate(0, 0, 90)
		if endTime.After(now) {
			endTime = now
		}

		if err := t.fetch(startTime, endTime); err != nil {
			return err
		}

		startTime = startTime.AddDate(0, 0, 90)
		now = time.Now().Add(-time.Minute).Truncate(time.Second)
	}
	return nil
}

func (t *TaskRunner) fetch(startTime time.Time, endTime time.Time) error {
	devices, err := t.fetchDevices(startTime, endTime)
	if err != nil {
		return err
	}

	// HACK: Dexcom - does not guarantee to return a device for G5 Mobile if time range < 24 hours (per Dexcom)
	var deviceInfo *DeviceInfo
	if endTime.Sub(startTime) > 24*time.Hour {
		if len(*devices) == 0 {
			return nil
		} else if deviceInfo, err = t.calculateDeviceInfo(devices); err != nil {
			return err
		}
	} else {
		if err = t.preloadDataSet(); err != nil {
			return err
		} else if t.dataSet == nil {
			return nil
		} else if deviceInfo, err = NewDeviceInfoFromDataSet(t.dataSet); err != nil {
			return err
		} else if !deviceInfo.IsDeviceModelG5Mobile() && !deviceInfo.IsDeviceModelUnknown() {
			deviceInfo = NewDeviceInfoFromMultiple()
		}
	}

	datumArray, err := t.fetchData(startTime, endTime)
	if err != nil {
		return err
	} else if len(datumArray) == 0 {
		return nil
	}

	if err = t.prepareDatumArray(datumArray, deviceInfo); err != nil {
		return err
	}

	if err = t.prepareDataSet(deviceInfo); err != nil {
		return err
	}

	return t.storeDatumArray(datumArray)
}

func (t *TaskRunner) fetchDevices(startTime time.Time, endTime time.Time) (*dexcom.Devices, error) {
	response, err := t.DexcomClient().GetDevices(t.context, startTime, endTime, t.tokenSource)
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

	return response.Devices, nil
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

func (t *TaskRunner) calculateDeviceInfo(devices *dexcom.Devices) (*DeviceInfo, error) {
	deviceInfo := NewDeviceInfo()
	for _, device := range *devices {
		if deviceDeviceInfo, err := NewDeviceInfoFromDevice(device); err != nil {
			return nil, err
		} else if deviceInfo, err = deviceInfo.Merge(deviceDeviceInfo); err != nil {
			return nil, err
		}
	}
	return deviceInfo, nil
}

func (t *TaskRunner) fetchData(startTime time.Time, endTime time.Time) ([]data.Datum, error) {
	datumArray := []data.Datum{}

	fetchDatumArray, err := t.fetchCalibrations(startTime, endTime)
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

	return datumArray, nil
}

func (t *TaskRunner) fetchCalibrations(startTime time.Time, endTime time.Time) ([]data.Datum, error) {
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

	datumArray := []data.Datum{}
	for _, c := range *response.Calibrations {
		if t.afterLatestDataTime(*c.SystemTime) {
			datumArray = append(datumArray, translateCalibrationToDatum(c))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEGVs(startTime time.Time, endTime time.Time) ([]data.Datum, error) {
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

	datumArray := []data.Datum{}
	for _, e := range *response.EGVs {
		if t.afterLatestDataTime(*e.SystemTime) {
			datumArray = append(datumArray, translateEGVToDatum(e, response.Unit, response.RateUnit))
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEvents(startTime time.Time, endTime time.Time) ([]data.Datum, error) {
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

	datumArray := []data.Datum{}
	for _, e := range *response.Events {
		if t.afterLatestDataTime(*e.SystemTime) {
			switch *e.EventType {
			case dexcom.EventCarbs:
				datumArray = append(datumArray, translateEventCarbsToDatum(e))
			case dexcom.EventExercise:
				datumArray = append(datumArray, translateEventExerciseToDatum(e))
			case dexcom.EventHealth:
				datumArray = append(datumArray, translateEventHealthToDatum(e))
			case dexcom.EventInsulin:
				datumArray = append(datumArray, translateEventInsulinToDatum(e))
			}
		}
	}

	return datumArray, nil
}

func (t *TaskRunner) prepareDatumArray(datumArray []data.Datum, deviceInfo *DeviceInfo) error {
	var datumDeviceID *string
	if deviceInfo.DeviceID != dexcom.DeviceIDMultiple {
		datumDeviceID = pointer.FromString(deviceInfo.DeviceID)
	} else {
		datumDeviceID = pointer.FromString(dexcom.DeviceIDUnknown)
	}

	for _, datum := range datumArray {
		datum.SetDeviceID(datumDeviceID)
	}

	sort.Sort(BySystemTime(datumArray))

	return nil
}

func (t *TaskRunner) prepareDataSet(deviceInfo *DeviceInfo) error {
	if err := t.preloadDataSet(); err != nil {
		return err
	}

	if t.dataSet != nil {
		return t.updateDataSetWithDeviceInfo(deviceInfo)
	}

	dataSet, err := t.createDataSet(deviceInfo)
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

func (t *TaskRunner) createDataSet(deviceInfo *DeviceInfo) (*data.DataSet, error) {
	dataSetCreate := data.NewDataSetCreate()
	dataSetCreate.Client = &data.DataSetClient{
		Name:    pointer.FromString(DataSetClientName),
		Version: pointer.FromString(DataSetClientVersion),
	}
	dataSetCreate.DataSetType = pointer.FromString(data.DataSetTypeContinuous)
	dataSetCreate.Deduplicator = data.NewDeduplicatorDescriptor()
	dataSetCreate.Deduplicator.Name = pointer.FromString(dataDeduplicatorDeduplicator.NoneName)
	dataSetCreate.DeviceID = pointer.FromString(deviceInfo.DeviceID)
	dataSetCreate.DeviceManufacturers = pointer.FromStringArray([]string{"Dexcom"})
	dataSetCreate.DeviceModel = pointer.FromString(deviceInfo.DeviceModel)
	dataSetCreate.DeviceSerialNumber = pointer.FromString(deviceInfo.DeviceSerialNumber)
	dataSetCreate.DeviceTags = pointer.FromStringArray([]string{data.DeviceTagCGM})
	dataSetCreate.Time = pointer.FromTime(time.Now().Truncate(time.Second))
	dataSetCreate.TimeProcessing = pointer.FromString(upload.TimeProcessingNone)

	dataSet, err := t.DataClient().CreateUserDataSet(t.context, t.providerSession.UserID, dataSetCreate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create data set")
	}
	if err = t.updateDataSourceWithDataSet(dataSet); err != nil {
		return nil, err
	}

	return dataSet, nil
}

func (t *TaskRunner) storeDatumArray(datumArray []data.Datum) error {
	length := len(datumArray)
	for startIndex := 0; startIndex < length; startIndex += DataSetSize {
		endIndex := startIndex + DataSetSize
		if endIndex > length {
			endIndex = length
		}

		if err := t.DataClient().CreateDataSetsData(t.context, *t.dataSet.UploadID, datumArray[startIndex:endIndex]); err != nil {
			return errors.Wrap(err, "unable to create data set data")
		}

		earliestDataTime := payloadSystemTime(datumArray[0])
		latestDataTime := payloadSystemTime(datumArray[endIndex-1])
		if err := t.updateDataSourceWithDataTime(earliestDataTime, latestDataTime); err != nil {
			return err
		}
	}

	return nil
}

func (t *TaskRunner) beforeEarliestDataTime(earliestDataTime time.Time) bool {
	return t.dataSource.EarliestDataTime == nil || earliestDataTime.Before(*t.dataSource.EarliestDataTime)
}

func (t *TaskRunner) afterLatestDataTime(latestDataTime time.Time) bool {
	return t.dataSource.LatestDataTime == nil || latestDataTime.After(*t.dataSource.LatestDataTime)
}

func payloadSystemTime(datum data.Datum) time.Time {
	payload := datum.GetPayload()
	if payload == nil {
		return time.Time{}
	}
	value := payload.Get("systemTime")
	if value == nil {
		return time.Time{}
	}
	systemTime, ok := value.(time.Time)
	if !ok {
		return time.Time{}
	}
	return systemTime
}

type BySystemTime []data.Datum

func (b BySystemTime) Len() int {
	return len(b)
}

func (b BySystemTime) Less(left int, right int) bool {
	return payloadSystemTime(b[left]).Before(payloadSystemTime(b[right]))
}

func (b BySystemTime) Swap(left int, right int) {
	b[left], b[right] = b[right], b[left]
}

type DeviceInfo struct {
	DeviceID           string
	DeviceModel        string
	DeviceSerialNumber string
}

func NewDeviceInfo() *DeviceInfo {
	return &DeviceInfo{}
}

func NewDeviceInfoFromMultiple() *DeviceInfo {
	return &DeviceInfo{
		DeviceID:           dexcom.DeviceIDMultiple,
		DeviceModel:        dexcom.DeviceModelMultiple,
		DeviceSerialNumber: dexcom.DeviceSerialNumberMultiple,
	}
}

func NewDeviceInfoFromDataSet(dataSet *data.DataSet) (*DeviceInfo, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}

	deviceInfo := &DeviceInfo{}
	if dataSet.DeviceID != nil {
		deviceInfo.DeviceID = *dataSet.DeviceID
	}
	if dataSet.DeviceModel != nil {
		deviceInfo.DeviceModel = *dataSet.DeviceModel
	}
	if dataSet.DeviceSerialNumber != nil {
		deviceInfo.DeviceSerialNumber = *dataSet.DeviceSerialNumber
	}
	return deviceInfo, nil
}

func NewDeviceInfoFromDevice(device *dexcom.Device) (*DeviceInfo, error) {
	if device == nil {
		return nil, errors.New("device is missing")
	}

	var deviceID string
	var deviceIDPrefix string
	var deviceModel string
	var deviceSerialNumber string

	switch *device.Model {
	case dexcom.ModelG5MobileApp:
		deviceModel = "G5Mobile"
		deviceIDPrefix = "DexG5Mob_"
	case dexcom.ModelG5Receiver:
		deviceModel = "G5MobileReceiver"
		deviceIDPrefix = "DexG5MobRec_"
	case dexcom.ModelG4WithShareReceiver:
		deviceModel = "G4ShareReceiver"
		deviceIDPrefix = "DexG4RecwitSha_"
	case dexcom.ModelG4Receiver:
		deviceModel = "G4Receiver"
		deviceIDPrefix = "DexG4Rec_"
	case dexcom.ModelUnknown:
		deviceModel = "Unknown"
		deviceIDPrefix = "DexUnknown_"
	default:
		return nil, errors.New("unknown device model")
	}

	if device.SerialNumber != nil {
		deviceSerialNumber = *device.SerialNumber
		deviceID = deviceIDPrefix + deviceSerialNumber
	}

	return &DeviceInfo{
		DeviceID:           deviceID,
		DeviceModel:        deviceModel,
		DeviceSerialNumber: deviceSerialNumber,
	}, nil
}

func (d *DeviceInfo) IsEmpty() bool {
	return d.DeviceID == "" && d.DeviceModel == "" && d.DeviceSerialNumber == ""
}

func (d *DeviceInfo) IsDeviceModelG5Mobile() bool {
	return d.DeviceModel == "G5Mobile"
}

func (d *DeviceInfo) IsDeviceModelUnknown() bool {
	return d.DeviceModel == "Unknown"
}

func (d *DeviceInfo) Merge(deviceInfo *DeviceInfo) (*DeviceInfo, error) {
	if deviceInfo == nil {
		return nil, errors.New("device info is missing")
	} else if deviceInfo.IsEmpty() {
		return d, nil
	}

	if d.IsEmpty() {
		return deviceInfo, nil
	}

	mergedDeviceInfo := &DeviceInfo{
		DeviceID:           d.DeviceID,
		DeviceModel:        d.DeviceModel,
		DeviceSerialNumber: d.DeviceSerialNumber,
	}

	if d.DeviceID != deviceInfo.DeviceID {
		mergedDeviceInfo.DeviceID = dexcom.DeviceIDMultiple
	}
	if d.DeviceModel != deviceInfo.DeviceModel {
		mergedDeviceInfo.DeviceModel = dexcom.DeviceModelMultiple
	}
	if d.DeviceSerialNumber != deviceInfo.DeviceSerialNumber {
		mergedDeviceInfo.DeviceSerialNumber = dexcom.DeviceSerialNumberMultiple
	}

	return mergedDeviceInfo, nil
}
