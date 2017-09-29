package fetch

import (
	"context"
	"sort"
	"time"

	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data"
	dataClient "github.com/tidepool-org/platform/data/client"
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

const AvailableAfterDuration = time.Hour
const DataSetSize = 2000

var InitialDataTime = time.Unix(1420070400, 0) // 2015-01-01T00:00:00Z

type Runner struct {
	logger          log.Logger
	versionReporter version.Reporter
	authClient      auth.Client
	dataClient      dataClient.Client
	dexcomClient    dexcom.Client
}

func NewRunner(logger log.Logger, versionReporter version.Reporter, authClient auth.Client, dataClient dataClient.Client, dexcomClient dexcom.Client) (*Runner, error) {
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
	if dexcomClient == nil {
		return nil, errors.New("dexcom client is missing")
	}

	return &Runner{
		logger:          logger,
		versionReporter: versionReporter,
		authClient:      authClient,
		dataClient:      dataClient,
		dexcomClient:    dexcomClient,
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

func (r *Runner) DexcomClient() dexcom.Client {
	return r.dexcomClient
}

func (r *Runner) CanRunTask(tsk *task.Task) bool {
	return tsk != nil && tsk.Type == Type
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) {
	ctx = log.NewContextWithLogger(ctx, r.Logger())

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

	if !tsk.IsFailed() {
		tsk.RepeatAvailableAfter(AvailableAfterDuration)
	}
}

type TaskRunner struct {
	*Runner
	task            *task.Task
	context         context.Context
	validator       *structureValidator.Validator
	providerSession *auth.ProviderSession
	dataSource      *data.DataSource
	tokenSource     oauth.TokenSource
	dataSetID       string
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
	t.validator = structureValidator.New().WithSource(structure.NewPointerSource())

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
		if errors.Code(errors.Cause(err)) == request.ErrorCodeUnauthorized {
			t.task.SetFailed()
			if err = t.updateDataSourceWithError(err); err != nil {
				t.Logger().WithError(err).Error("unable to update data source with error")
			}
		}
		return err
	}

	return nil
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

	dataSource, err := t.DataClient().GetDataSource(t.context, dataSourceID)
	if err != nil {
		return errors.Wrap(err, "unable to get data source")
	} else if dataSource == nil {
		t.task.SetFailed()
		return errors.Wrap(err, "data source is missing")
	}

	t.dataSource = dataSource
	return nil
}

func (t *TaskRunner) updateDataSourceWithDataSetID(dataSetID string) error {
	dataSourceUpdate := data.NewDataSourceUpdate()
	dataSourceUpdate.DataSetIDs = pointer.StringArray(append(t.dataSource.DataSetIDs, dataSetID))
	return t.updateDataSource(dataSourceUpdate)
}

func (t *TaskRunner) updateDataSourceWithLatestDataTime(latestDataTime time.Time) error {
	if t.dataSource.LatestDataTime != nil && latestDataTime.Before(*t.dataSource.LatestDataTime) {
		return nil
	}

	dataSourceUpdate := data.NewDataSourceUpdate()
	dataSourceUpdate.LastImportTime = pointer.Time(time.Now())
	dataSourceUpdate.LatestDataTime = pointer.Time(latestDataTime)
	return t.updateDataSource(dataSourceUpdate)
}

func (t *TaskRunner) updateDataSourceWithError(err error) error {
	dataSourceUpdate := data.NewDataSourceUpdate()
	dataSourceUpdate.State = pointer.String(data.DataSourceStateError)
	dataSourceUpdate.Error = &errors.Serializable{Error: err}
	return t.updateDataSource(dataSourceUpdate)
}

func (t *TaskRunner) updateDataSource(dataSourceUpdate *data.DataSourceUpdate) error {
	dataSource, err := t.DataClient().UpdateDataSource(t.context, t.dataSource.ID, dataSourceUpdate)
	if err != nil {
		return errors.Wrap(err, "unable to update data source")
	} else if dataSource == nil {
		t.task.SetFailed()
		return errors.Wrap(err, "data source is missing")
	}
	t.dataSource = dataSource

	return nil
}

func (t *TaskRunner) createTokenSource() error {
	tokenSource, err := oauthToken.NewSource(t.providerSession.OAuthToken)
	if err != nil {
		t.task.SetFailed()
		return errors.Wrap(err, "unable to create token source")
	}

	t.tokenSource = tokenSource
	return nil
}

func (t *TaskRunner) updateDataSet(dataSetUpdate *data.DataSetUpdate) error {
	dataSet, err := t.DataClient().UpdateDataSet(t.context, t.dataSetID, dataSetUpdate)
	if err != nil {
		return errors.Wrap(err, "unable to update data set")
	} else if dataSet == nil {
		t.task.SetFailed()
		return errors.Wrap(err, "data set is missing")
	}

	return nil
}

func (t *TaskRunner) fetchSinceLatestDataTime() error {
	startTime := InitialDataTime
	if t.dataSource.LatestDataTime != nil && startTime.Before(*t.dataSource.LatestDataTime) {
		startTime = *t.dataSource.LatestDataTime
	}

	now := time.Now()
	for startTime.Before(now) {
		endTime := startTime.AddDate(0, 0, 90)
		if endTime.After(now) {
			endTime = now
		}

		if err := t.fetch(startTime, endTime); err != nil {
			return err
		}

		startTime = startTime.AddDate(0, 0, 90)
		now = time.Now()
	}
	return nil
}

func (t *TaskRunner) fetch(startTime time.Time, endTime time.Time) error {
	devices, err := t.fetchDevices(startTime, endTime)
	if err != nil {
		return err
	} else if len(devices) == 0 {
		return nil
	}

	datumArray := []data.Datum{}

	fetchDatumArray, err := t.fetchCalibrations(startTime, endTime)
	if err != nil {
		return err
	}
	datumArray = append(datumArray, fetchDatumArray...)

	fetchDatumArray, err = t.fetchEGVs(startTime, endTime)
	if err != nil {
		return err
	}
	datumArray = append(datumArray, fetchDatumArray...)

	fetchDatumArray, err = t.fetchEvents(startTime, endTime)
	if err != nil {
		return err
	}
	datumArray = append(datumArray, fetchDatumArray...)

	if t.dataSetID == "" {
		for index := len(t.dataSource.DataSetIDs) - 1; index >= 0; index-- {
			dataSetID := t.dataSource.DataSetIDs[index]
			dataSet, dataSetErr := t.DataClient().GetDataSet(t.context, dataSetID)
			if dataSetErr != nil {
				return errors.Wrap(dataSetErr, "unable to get data set")
			}
			if dataSet.DataSetType == nil || *dataSet.DataSetType != data.DataSetTypeContinuous {
				continue
			}
			if dataSet.State != data.DataSetStateOpen {
				continue
			}

			// TODO: Is this data set okay for us?

			t.dataSetID = dataSetID
		}

		if t.dataSetID == "" {
			dataSetCreate := data.NewDataSetCreate()
			dataSetCreate.Client = &data.DataSetClient{
				Name:    DatasetClientName,
				Version: DatasetClientVersion,
			}
			dataSetCreate.DataSetType = data.DataSetTypeContinuous
			dataSetCreate.DeviceID = "multiple"
			dataSetCreate.DeviceManufacturers = []string{"Dexcom"}
			dataSetCreate.DeviceModel = "multiple"
			dataSetCreate.DeviceSerialNumber = "multiple"
			dataSetCreate.DeviceTags = []string{data.DeviceTagCGM}
			dataSetCreate.Time = time.Now()
			dataSetCreate.TimeProcessing = upload.TimeProcessingNone

			dataSet, dataSetErr := t.DataClient().CreateUserDataSet(t.context, t.providerSession.UserID, dataSetCreate)
			if dataSetErr != nil {
				return errors.Wrap(dataSetErr, "unable to create data set")
			}
			if err = t.updateDataSourceWithDataSetID(dataSet.UploadID); err != nil {
				return err
			}

			t.dataSetID = dataSet.UploadID
		}
	}

	// TODO: Run allDatum through Validate and Normalize once all Datum use structure/validator
	// TODO: identity fields for new data types (or perhaps not if we declare our own hash)

	sort.Sort(BySystemTime(datumArray))

	return t.storeDatumArray(datumArray)
}

func (t *TaskRunner) fetchDevices(startTime time.Time, endTime time.Time) ([]*dexcom.Device, error) {
	response, err := t.DexcomClient().GetDevices(t.context, startTime, endTime, t.tokenSource)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get devices")
	}

	if err = t.updateProviderSession(); err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}

	return response.Devices, nil
}

func (t *TaskRunner) fetchCalibrations(startTime time.Time, endTime time.Time) ([]data.Datum, error) {
	response, err := t.DexcomClient().GetCalibrations(t.context, startTime, endTime, t.tokenSource)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get calibrations")
	}

	if err = t.updateProviderSession(); err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}

	datumArray := []data.Datum{}
	for _, c := range response.Calibrations {
		datumArray = append(datumArray, translateCalibrationToDatum(c))
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEGVs(startTime time.Time, endTime time.Time) ([]data.Datum, error) {
	response, err := t.DexcomClient().GetEGVs(t.context, startTime, endTime, t.tokenSource)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get egvs")
	}

	if err = t.updateProviderSession(); err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}

	datumArray := []data.Datum{}
	for _, e := range response.EGVs {
		datumArray = append(datumArray, translateEGVToDatum(e, response.Unit, response.RateUnit))
	}

	return datumArray, nil
}

func (t *TaskRunner) fetchEvents(startTime time.Time, endTime time.Time) ([]data.Datum, error) {
	response, err := t.DexcomClient().GetEvents(t.context, startTime, endTime, t.tokenSource)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get events")
	}

	if err = t.updateProviderSession(); err != nil {
		return nil, err
	}

	t.validator.Validate(response)
	if err = t.validator.Error(); err != nil {
		return nil, err
	}

	datumArray := []data.Datum{}
	for _, e := range response.Events {
		switch e.EventType {
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

	return datumArray, nil
}

func (t *TaskRunner) storeDatumArray(datumArray []data.Datum) error {
	length := len(datumArray)
	for startIndex := 0; startIndex < length; startIndex += DataSetSize {
		endIndex := startIndex + DataSetSize
		if endIndex > length {
			endIndex = length
		}

		if err := t.DataClient().CreateDataSetsData(t.context, t.dataSetID, datumArray[startIndex:endIndex]); err != nil {
			return errors.Wrap(err, "unable to create data set data")
		}

		lastDatum := datumArray[endIndex-1]

		if err := t.updateDataSourceWithLatestDataTime(payloadSystemTime(lastDatum)); err != nil {
			return err
		}
	}

	return nil
}

func payloadSystemTime(datum data.Datum) time.Time {
	payload := datum.GetPayload()
	if payload == nil {
		return time.Time{}
	}
	systemTimeObject, ok := (*payload)["systemTime"]
	if !ok {
		return time.Time{}
	}
	systemTime, ok := systemTimeObject.(time.Time)
	if !ok {
		return time.Time{}
	}
	return systemTime
}

type BySystemTime []data.Datum

func (b BySystemTime) Len() int {
	return len(b)
}

func (b BySystemTime) Swap(left int, right int) {
	b[left], b[right] = b[right], b[left]
}

func (b BySystemTime) Less(left int, right int) bool {
	return payloadSystemTime(b[left]).Before(payloadSystemTime(b[right]))
}
