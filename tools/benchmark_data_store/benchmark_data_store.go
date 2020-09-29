package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"strconv"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/data"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataStoreMongo "github.com/tidepool-org/platform/data/store/mongo"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/origin"
	"github.com/tidepool-org/platform/pointer"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"github.com/tidepool-org/platform/tool"
)

const (
	AddressesFlag = "addresses"
	DatabaseFlag  = "database"
	TLSFlag       = "tls"
)

type BenchmarkInput struct {
	DataSetID *string         `json:"dataSetId,omitempty"`
	DeviceID  *string         `json:"deviceId,omitempty"`
	GroupID   *string         `json:"groupId,omitempty"`
	Limit     *int            `json:"limit,omitempty"`
	Selectors *data.Selectors `json:"selectors,omitempty"`
	Time      *string         `json:"time,omitempty"`
	UserID    *string         `json:"userId,omitempty"`
}

type Benchmark struct {
	Name  *string         `json:"name,omitempty"`
	Input *BenchmarkInput `json:"input,omitempty"`
}

type Benchmarks []*Benchmark

func main() {
	application.RunAndExit(NewTool())
}

type Tool struct {
	*tool.Tool
	config          *storeStructuredMongo.Config
	store           *dataStoreMongo.Store
	benchmarksFiles []string
	benchmarks      Benchmarks
}

func NewTool() *Tool {
	return &Tool{
		Tool: tool.New(),
	}
}

func (t *Tool) Initialize(provider application.Provider) error {
	if err := t.Tool.Initialize(provider); err != nil {
		return err
	}

	t.CLI().Usage = "Benchmark data store performance"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.StringSliceFlag{
			Name:  AddressesFlag,
			Usage: "addresses of store database server",
		},
		cli.StringFlag{
			Name:  DatabaseFlag,
			Usage: "store database name",
		},
		cli.BoolFlag{
			Name:  TLSFlag,
			Usage: "use TLS for store connection",
		},
	)

	t.CLI().Action = func(ctx *cli.Context) error {
		if !t.ParseContext(ctx) {
			return nil
		}
		return t.execute()
	}

	rand.Seed(time.Now().Unix())

	if err := t.initializeConfig(); err != nil {
		return err
	}

	return nil
}

func (t *Tool) Terminate() {
	t.terminateStore()
	t.terminateConfig()

	t.Tool.Terminate()
}

func (t *Tool) ParseContext(ctx *cli.Context) bool {
	if parsed := t.Tool.ParseContext(ctx); !parsed {
		return parsed
	}

	if ctx.IsSet(AddressesFlag) {
		t.config.Addresses = ctx.StringSlice(AddressesFlag)
	}
	if ctx.IsSet(DatabaseFlag) {
		t.config.Database = ctx.String(DatabaseFlag)
	}
	if ctx.IsSet(TLSFlag) {
		t.config.TLS = ctx.Bool(TLSFlag)
	}

	t.benchmarksFiles = ctx.Args()

	return true
}

func (t *Tool) execute() error {
	if err := t.initializeStore(); err != nil {
		return err
	}

	if err := t.loadBenchmarks(); err != nil {
		return err
	}
	if err := t.executeBenchmarks(log.NewContextWithLogger(context.Background(), t.Logger())); err != nil {
		return err
	}

	return nil
}

func (t *Tool) initializeConfig() error {
	t.Logger().Debug("Loading config")

	config := storeStructuredMongo.NewConfig(nil)
	if err := config.Load(); err != nil {
		return errors.Wrap(err, "unable to load config")
	}
	t.config = config

	return nil
}

func (t *Tool) terminateConfig() {
}

func (t *Tool) initializeStore() error {
	t.Logger().Debug("Creating store")

	params := storeStructuredMongo.Params{DatabaseConfig: t.config}
	store, err := dataStoreMongo.NewStore(params)
	if err != nil {
		return errors.Wrap(err, "unable to create store")
	}
	t.store = store

	return nil
}

func (t *Tool) terminateStore() {
	if t.store != nil {
		t.Logger().Debug("Destroying store")
		t.store.Terminate(nil)
		t.store = nil
	}
}

func (t *Tool) loadBenchmarks() error {
	t.Logger().Debug("Loading benchmarks")

	for _, benchmarksFile := range t.benchmarksFiles {
		benchmarks, err := t.loadBenchmarksFile(benchmarksFile)
		if err != nil {
			return err
		}
		t.benchmarks = append(t.benchmarks, benchmarks...)
	}
	return nil
}

func (t *Tool) loadBenchmarksFile(benchmarksFile string) (Benchmarks, error) {
	t.Logger().Debugf("Loading benchmarks file %q", benchmarksFile)

	content, err := ioutil.ReadFile(benchmarksFile)
	if err != nil {
		return nil, errors.Newf("unable to read content from benchmarks file %q", benchmarksFile)
	}

	var benchmarks Benchmarks
	if err = json.Unmarshal(content, &benchmarks); err != nil {
		return nil, errors.Newf("unable to parse content from benchmarks file %q", benchmarksFile)
	}

	return benchmarks, nil
}

func (t *Tool) executeBenchmarks(ctx context.Context) error {
	t.Logger().Debug("Executing benchmarks")

	for _, benchmark := range t.benchmarks {
		if err := t.executeBenchmark(ctx, benchmark); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tool) executeBenchmark(ctx context.Context, benchmark *Benchmark) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if benchmark == nil {
		return errors.New("benchmark is missing")
	}
	if benchmark.Name == nil {
		return errors.New("benchmark function is missing")
	}

	now := time.Now()

	logger := t.Logger().WithField("benchmark", benchmark)
	logger.Debug("Executing benchmark")

	dataRepository := t.store.NewDataRepository()

	var err error
	switch *benchmark.Name {
	case "JellyfishMetaFindByInternalID":
		err = t.benchmarkJellyfishMetaFindByInternalID(ctx, dataRepository, benchmark.Input)
	case "JellyfishMetaFindBefore":
		err = t.benchmarkJellyfishMetaFindBefore(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaCreate":
		err = t.benchmarkPlatformMetaCreate(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaDeleteDataWithOrigin":
		err = t.benchmarkPlatformMetaDeleteDataWithOrigin(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaActivate":
		err = t.benchmarkPlatformMetaActivate(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaArchiveWithHashes":
		err = t.benchmarkPlatformMetaArchiveWithHashes(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaDeleteOtherData":
		err = t.benchmarkPlatformMetaDeleteOtherData(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaDeleteDataSet":
		err = t.benchmarkPlatformMetaDeleteDataSet(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaUnarchiveWithHashes":
		err = t.benchmarkPlatformMetaUnarchiveWithHashes(ctx, dataRepository, benchmark.Input)
	case "PlatformMetaDestroy":
		err = t.benchmarkPlatformMetaDestroy(ctx, dataRepository, benchmark.Input)
	case "PlatformDBActivateDataSetData":
		err = t.benchmarkPlatformDBActivateDataSetData(ctx, dataRepository, benchmark.Input)
	case "PlatformDBArchiveDataSetData":
		err = t.benchmarkPlatformDBArchiveDataSetData(ctx, dataRepository, benchmark.Input)
	case "PlatformDBArchiveDeviceDataUsingHashesFromDataSet":
		err = t.benchmarkPlatformDBArchiveDeviceDataUsingHashesFromDataSet(ctx, dataRepository, benchmark.Input)
	case "PlatformDBCreateDataSet":
		err = t.benchmarkPlatformDBCreateDataSet(ctx, dataRepository, benchmark.Input)
	case "PlatformDBCreateDataSetData":
		err = t.benchmarkPlatformDBCreateDataSetData(ctx, dataRepository, benchmark.Input)
	case "PlatformDBDeleteDataSet":
		err = t.benchmarkPlatformDBDeleteDataSet(ctx, dataRepository, benchmark.Input)
	case "PlatformDBDeleteDataSetData":
		err = t.benchmarkPlatformDBDeleteDataSetData(ctx, dataRepository, benchmark.Input)
	case "PlatformDBDeleteOtherDataSetData":
		err = t.benchmarkPlatformDBDeleteOtherDataSetData(ctx, dataRepository, benchmark.Input)
	case "PlatformDBDestroyDataForUserByID":
		err = t.benchmarkPlatformDBDestroyDataForUserByID(ctx, dataRepository, benchmark.Input)
	case "PlatformDBDestroyDataSetData":
		err = t.benchmarkPlatformDBDestroyDataSetData(ctx, dataRepository, benchmark.Input)
	case "PlatformDBDestroyDeletedDataSetData":
		err = t.benchmarkPlatformDBDestroyDeletedDataSetData(ctx, dataRepository, benchmark.Input)
	case "PlatformDBGetDataSet":
		err = t.benchmarkPlatformDBGetDataSet(ctx, dataRepository, benchmark.Input)
	case "PlatformDBGetDataSetByID":
		err = t.benchmarkPlatformDBGetDataSetByID(ctx, dataRepository, benchmark.Input)
	case "PlatformDBGetDataSetsForUserByID":
		err = t.benchmarkPlatformDBGetDataSetsForUserByID(ctx, dataRepository, benchmark.Input)
	case "PlatformDBListUserDataSets":
		err = t.benchmarkPlatformDBListUserDataSets(ctx, dataRepository, benchmark.Input)
	case "PlatformDBUnarchiveDeviceDataUsingHashesFromDataSet":
		err = t.benchmarkPlatformDBUnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataRepository, benchmark.Input)
	case "PlatformDBUpdateDataSet":
		err = t.benchmarkPlatformDBUpdateDataSet(ctx, dataRepository, benchmark.Input)
	case "TideWhispererAPIGetData":
		err = t.benchmarkTideWhispererAPIGetData(ctx, dataRepository, benchmark.Input)
	case "TideWhispererDBHasMedtronicDirectData":
		err = t.benchmarkTideWhispererDBHasMedtronicDirectData(ctx, dataRepository, benchmark.Input)
	case "TideWhispererDBGetDeviceData":
		err = t.benchmarkTideWhispererDBGetDeviceData(ctx, dataRepository, benchmark.Input)
	default:
		err = errors.New("benchmark name invalid")
	}

	logger.WithError(err).WithField("duration", time.Since(now)/time.Microsecond).Debug("Executed benchmark")

	if err != nil {
		err = errors.Wrapf(err, "failure executing benchmark %q", *benchmark.Name)
	}

	return err
}

func (t *Tool) benchmarkJellyfishMetaFindByInternalID(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	var internalBytes [12]byte

	binary.BigEndian.PutUint32(internalBytes[:], uint32(1470302050+rand.Int31n(65936775)))
	rand.Read(internalBytes[4:])
	internalID := hex.EncodeToString(internalBytes[:])

	selector := bson.M{
		"_id": internalID,
	}

	var results map[string]interface{}
	if err := session.(*dataStoreMongo.DataRepository).FindOne(context.Background(), selector).Decode(&results); err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	return nil
}

func (t *Tool) benchmarkJellyfishMetaFindBefore(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.DeviceID == nil {
		return errors.New("benchmark input device id is missing")
	}
	if input.GroupID == nil {
		return errors.New("benchmark input group id is missing")
	}
	if input.Time == nil {
		return errors.New("benchmark input time is missing")
	}

	selector := bson.M{
		"_active":  true,
		"deviceId": *input.DeviceID,
		"_groupId": *input.GroupID,
		"time": bson.M{
			"$lt": *input.Time,
		},
		"type": "basal",
	}

	var results map[string]interface{}
	opts := options.FindOne().SetSort(bson.M{"time": 1})
	if err := session.(*dataStoreMongo.DataRepository).FindOne(context.Background(), selector, opts).Decode(&results); err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData

func (t *Tool) benchmarkPlatformMetaCreate(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)
	if _, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData); err != nil {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData, DeleteDataSetData (selectors), DestroyDeletedDataSetData (selectors)

func (t *Tool) benchmarkPlatformMetaDeleteDataWithOrigin(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)

	dataSet, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData)
	if err != nil {
		return err
	}

	selectors := data.NewSelectors()
	for _, dataSetDatum := range preparedDataSetData {
		*selectors = append(*selectors, &data.Selector{Origin: &data.SelectorOrigin{ID: pointer.CloneString(dataSetDatum.GetOrigin().ID)}})
	}

	if err = session.DeleteDataSetData(ctx, dataSet, selectors); err != nil {
		return err
	}

	if err = session.DestroyDeletedDataSetData(ctx, dataSet, selectors); err != nil {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData, UpdateDataSet (state closed), ActivateDataSetData

func (t *Tool) benchmarkPlatformMetaActivate(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)

	dataSet, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData)
	if err != nil {
		return err
	}

	update := data.NewDataSetUpdate()
	update.State = pointer.FromString("closed")
	if _, err = session.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return err
	}

	if err = session.ActivateDataSetData(ctx, dataSet, nil); err != nil {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData, UpdateDataSet (state closed), ActivateDataSetData, ArchiveDeviceDataUsingHashesFromDataSet

func (t *Tool) benchmarkPlatformMetaArchiveWithHashes(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)

	dataSet, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData)
	if err != nil {
		return err
	}

	update := data.NewDataSetUpdate()
	update.State = pointer.FromString("closed")
	if _, err = session.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return err
	}

	if err = session.ActivateDataSetData(ctx, dataSet, nil); err != nil {
		return err
	}

	if err = session.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet); err != nil {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData, UpdateDataSet (state closed), ActivateDataSetData, DeleteOtherDataSetData

func (t *Tool) benchmarkPlatformMetaDeleteOtherData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)

	dataSet, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData)
	if err != nil {
		return err
	}

	update := data.NewDataSetUpdate()
	update.State = pointer.FromString("closed")
	if _, err = session.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return err
	}

	if err = session.ActivateDataSetData(ctx, dataSet, nil); err != nil {
		return err
	}

	if err = session.DeleteOtherDataSetData(ctx, dataSet); err != nil {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData, UpdateDataSet (state closed), DeleteDataSet

func (t *Tool) benchmarkPlatformMetaDeleteDataSet(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)

	dataSet, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData)
	if err != nil {
		return err
	}

	update := data.NewDataSetUpdate()
	update.State = pointer.FromString("closed")
	if _, err = session.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return err
	}

	if err = session.DeleteDataSet(ctx, dataSet); err != nil {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData, UpdateDataSet (state closed), DeleteDataSet, UnarchiveDeviceDataUsingHashesFromDataSet

func (t *Tool) benchmarkPlatformMetaUnarchiveWithHashes(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)

	dataSet, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData)
	if err != nil {
		return err
	}

	update := data.NewDataSetUpdate()
	update.State = pointer.FromString("closed")
	if _, err = session.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return err
	}

	if err = session.DeleteDataSet(ctx, dataSet); err != nil {
		return err
	}

	if err = session.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet); err != nil {
		return err
	}

	return nil
}

// CreateDataSet, UpdateDataSet (set deduplicator), GetDataSetByID, CreateDataSetData, UpdateDataSet (state closed), ArchiveDataSetData, DestroyDataSetData

func (t *Tool) benchmarkPlatformMetaDestroy(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	preparedDataSet, preparedDataSetData := t.prepareDataSetWithData(input)

	dataSet, err := t.createDataSetWithData(ctx, session, preparedDataSet, preparedDataSetData)
	if err != nil {
		return err
	}

	update := data.NewDataSetUpdate()
	update.State = pointer.FromString("closed")
	if _, err = session.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return err
	}

	if err = session.ArchiveDataSetData(ctx, dataSet, nil); err != nil {
		return err
	}

	if err = session.DestroyDataSetData(ctx, dataSet, nil); err != nil {
		return err
	}

	return nil
}

func (t *Tool) benchmarkPlatformDBActivateDataSetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.ActivateDataSetData(ctx, dataSet, input.Selectors)
}

func (t *Tool) benchmarkPlatformDBArchiveDataSetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.ArchiveDataSetData(ctx, dataSet, input.Selectors)
}

func (t *Tool) benchmarkPlatformDBArchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.DeviceID = pointer.CloneString(input.DeviceID)
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.ArchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)
}

func (t *Tool) benchmarkPlatformDBCreateDataSet(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.CreatedUserID = pointer.CloneString(input.UserID)
	dataSet.DeviceID = pointer.CloneString(input.DeviceID)
	dataSet.ID = pointer.FromString(data.NewID())
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.CreateDataSet(ctx, dataSet)
}

func (t *Tool) benchmarkPlatformDBCreateDataSetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.CreateDataSetData(ctx, dataSet, t.generateRandomDataSetData(input.DeviceID))
}

func (t *Tool) benchmarkPlatformDBDeleteDataSet(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.DeleteDataSet(ctx, dataSet)
}

func (t *Tool) benchmarkPlatformDBDeleteDataSetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.DeleteDataSetData(ctx, dataSet, input.Selectors)
}

func (t *Tool) benchmarkPlatformDBDeleteOtherDataSetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.DeviceID = pointer.CloneString(input.DeviceID)
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.DeleteOtherDataSetData(ctx, dataSet)
}

func (t *Tool) benchmarkPlatformDBDestroyDataForUserByID(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.UserID == nil {
		return errors.New("benchmark input user id is missing")
	}

	return session.DestroyDataForUserByID(ctx, *input.UserID)
}

func (t *Tool) benchmarkPlatformDBDestroyDataSetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.DestroyDataSetData(ctx, dataSet, input.Selectors)
}

func (t *Tool) benchmarkPlatformDBDestroyDeletedDataSetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.DestroyDeletedDataSetData(ctx, dataSet, input.Selectors)
}

func (t *Tool) benchmarkPlatformDBGetDataSet(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.DataSetID == nil {
		return errors.New("benchmark input data set id is missing")
	}

	_, err := session.GetDataSet(ctx, *input.DataSetID)
	return err
}

func (t *Tool) benchmarkPlatformDBGetDataSetByID(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.DataSetID == nil {
		return errors.New("benchmark input data set id is missing")
	}

	_, err := session.GetDataSetByID(ctx, *input.DataSetID)
	return err
}

func (t *Tool) benchmarkPlatformDBGetDataSetsForUserByID(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.UserID == nil {
		return errors.New("benchmark input user id is missing")
	}

	_, err := session.GetDataSetsForUserByID(ctx, *input.UserID, nil, nil)
	return err
}

func (t *Tool) benchmarkPlatformDBListUserDataSets(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.UserID == nil {
		return errors.New("benchmark input user id is missing")
	}

	_, err := session.ListUserDataSets(ctx, *input.UserID, nil, nil)
	return err
}

func (t *Tool) benchmarkPlatformDBUnarchiveDeviceDataUsingHashesFromDataSet(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	dataSet := dataTypesUpload.New()
	dataSet.DeviceID = pointer.CloneString(input.DeviceID)
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.CloneString(input.DataSetID)
	return session.UnarchiveDeviceDataUsingHashesFromDataSet(ctx, dataSet)
}

func (t *Tool) benchmarkPlatformDBUpdateDataSet(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.DataSetID == nil {
		return errors.New("benchmark input user id is missing")
	}

	update := data.NewDataSetUpdate()
	update.Active = pointer.FromBool(true)
	_, err := session.UpdateDataSet(ctx, *input.DataSetID, update)
	return err
}

func (t *Tool) benchmarkTideWhispererAPIGetData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if err := t.benchmarkTideWhispererDBHasMedtronicDirectData(ctx, session, input); err != nil {
		return err
	}
	if err := t.benchmarkTideWhispererDBGetDeviceData(ctx, session, input); err != nil {
		return err
	}
	return nil
}

func (t *Tool) benchmarkTideWhispererDBHasMedtronicDirectData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.UserID == nil {
		return errors.New("benchmark input user id is missing")
	}

	selector := bson.M{
		"_userId": *input.UserID,
		"type":    "upload",
		"_state":  "closed",
		"_active": true,
		"deletedTime": bson.M{
			"$exists": false,
		},
		"deviceManufacturers": "Medtronic",
	}
	opts := options.Count().SetLimit(1)
	_, err := session.(*dataStoreMongo.DataRepository).CountDocuments(context.Background(), selector, opts)
	return err
}

func (t *Tool) benchmarkTideWhispererDBGetDeviceData(ctx context.Context, session dataStore.DataRepository, input *BenchmarkInput) error {
	if input.UserID == nil {
		return errors.New("benchmark input user id is missing")
	}

	selector := bson.M{
		"_userId": input.UserID,
		"_active": true,
	}

	// FUTURE: Consider adding some/all of these options

	// &type=deviceEvent
	// if len(p.Types) > 0 && p.Types[0] != "" {
	// 	selector["type"] = bson.M{"$in": p.Types}
	// }

	// &type=status
	// if len(p.SubTypes) > 0 && p.SubTypes[0] != "" {
	// 	selector["subType"] = bson.M{"$in": p.SubTypes}
	// }

	// ?startDate=2018-07-13T00:00:00.000Z&endDate=2018-09-08T00:38:50.000Z
	// if p.Date.Start != "" && p.Date.End != "" {
	// 	selector["time"] = bson.M{"$gte": p.Date.Start, "$lte": p.Date.End}
	// } else if p.Date.Start != "" {
	// 	selector["time"] = bson.M{"$gte": p.Date.Start}
	// } else if p.Date.End != "" {
	// 	selector["time"] = bson.M{"$lte": p.Date.End}
	// }

	// ?carelink=true
	// if !p.Carelink {
	// 	selector["source"] = bson.M{"$ne": "carelink"}
	// }

	// ?dexcom=true
	// if !p.Dexcom && p.DexcomDataSource != nil {
	// 	dexcomQuery := []bson.M{
	// 		{"type": bson.M{"$ne": "cbg"}},
	// 		{"uploadId": bson.M{"$in": p.DexcomDataSource["dataSetIds"]}},
	// 	}
	// 	if earliestDataTime, ok := p.DexcomDataSource["earliestDataTime"].(time.Time); ok {
	// 		dexcomQuery = append(dexcomQuery, bson.M{"time": bson.M{"$lt": earliestDataTime.Format(time.RFC3339Nano)}})
	// 	}
	// 	if latestDataTime, ok := p.DexcomDataSource["latestDataTime"].(time.Time); ok {
	// 		dexcomQuery = append(dexcomQuery, bson.M{"time": bson.M{"$gt": latestDataTime.Format(time.RFC3339Nano)}})
	// 	}
	// 	selector["$or"] = dexcomQuery
	// }

	selectFields := bson.M{"_id": 0, "_userId": 0, "_groupId": 0, "_version": 0, "_active": 0, "_schemaVersion": 0, "createdTime": 0, "modifiedTime": 0}
	opts := options.Find().SetProjection(selectFields)
	if input.Limit != nil {
		opts = opts.SetLimit(int64(*input.Limit))
	}
	iter, err := session.(*dataStoreMongo.DataRepository).Find(context.Background(), selector, opts)
	if err != nil {
		return errors.New("cannot create query on data repository")
	}

	var results bson.Raw
	for iter.Next(context.Background()) {
		iter.Decode(&results)
	}
	return nil
}

func (t *Tool) prepareDataSetWithData(input *BenchmarkInput) (*dataTypesUpload.Upload, data.Data) {
	dataSet := dataTypesUpload.New()
	dataSet.CreatedUserID = pointer.CloneString(input.UserID)
	dataSet.DeviceID = pointer.CloneString(input.DeviceID)
	dataSet.ID = pointer.FromString(data.NewID())
	dataSet.UserID = pointer.CloneString(input.UserID)
	dataSet.UploadID = pointer.FromString(data.NewSetID())
	return dataSet, t.generateRandomDataSetData(input.DeviceID)
}

func (t *Tool) createDataSetWithData(ctx context.Context, session dataStore.DataRepository, dataSet *dataTypesUpload.Upload, dataSetData data.Data) (*dataTypesUpload.Upload, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	}
	if dataSetData == nil {
		return nil, errors.New("data set data is missing")
	}

	if err := session.CreateDataSet(ctx, dataSet); err != nil {
		return nil, err
	}

	update := data.NewDataSetUpdate()
	update.Active = pointer.FromBool(true)
	update.Deduplicator = dataSet.Deduplicator
	if _, err := session.UpdateDataSet(ctx, *dataSet.UploadID, update); err != nil {
		return nil, err
	}

	if _, err := session.GetDataSetByID(ctx, *dataSet.UploadID); err != nil {
		return nil, err
	}

	if err := session.CreateDataSetData(ctx, dataSet, dataSetData); err != nil {
		return nil, err
	}

	return session.GetDataSetByID(ctx, *dataSet.UploadID)
}

func (t *Tool) generateRandomDataSetData(deviceID *string) data.Data {
	dataSetData := make(data.Data, 2000)
	for index := range dataSetData {
		dataSetData[index] = t.generateRandomDataSetDatum(deviceID)
	}
	return dataSetData
}

func (t *Tool) generateRandomDataSetDatum(deviceID *string) data.Datum {
	origin := &origin.Origin{
		ID: pointer.FromString(strconv.Itoa(rand.Int())),
	}
	return &dataTypes.Base{
		DeviceID: pointer.CloneString(deviceID),
		ID:       pointer.FromString(data.NewID()),
		Origin:   origin,
		Time:     pointer.FromString(time.Now().Add(-timeYear).Add(time.Duration(rand.Int63n(int64(2 * timeYear)))).Format(time.RFC3339Nano)),
		Type:     "benchmark",
	}
}

const timeYear = 365 * 24 * time.Hour
