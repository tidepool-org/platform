package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	toolMongo "github.com/tidepool-org/platform/tool/mongo"
)

const (
	OutputFlag = "output"
)

func main() {
	application.RunAndExit(NewTool())
}

type Manufacturers []string

type Tags []string

func (t Tags) Contains(tag string) bool {
	for _, tg := range t {
		if tg == tag {
			return true
		}
	}
	return false
}

type Device struct {
	Manufacturers    Manufacturers `json:"manufacturers,omitempty"`
	Model            string        `json:"model,omitempty"`
	Tags             Tags          `json:"tags,omitempty"`
	LatestUploadTime string        `json:"latestUploadTime,omitempty"`
	UploadCount      int           `json:"uploadCount,omitempty"`
}

type Devices []*Device

func (d Devices) Select(selector func(device *Device) bool) Devices {
	var devices Devices
	for _, device := range d {
		if selector(device) {
			devices = append(devices, device)
		}
	}
	return devices
}

type DevicesByLatestUploadTimeDescending Devices

func (d DevicesByLatestUploadTimeDescending) Len() int {
	return len(d)
}

func (d DevicesByLatestUploadTimeDescending) Less(left int, right int) bool {
	if d[right].LatestUploadTime == "" {
		return true
	} else if d[left].LatestUploadTime == "" {
		return false
	}
	if compare := strings.Compare(d[right].LatestUploadTime, d[left].LatestUploadTime); compare < 0 {
		return true
	} else if compare == 0 {
		return strings.Compare(d[right].Model, d[left].Model) <= 0
	} else {
		return false
	}
}

func (d DevicesByLatestUploadTimeDescending) Swap(left int, right int) {
	d[left], d[right] = d[right], d[left]
}

type TypeTuple struct {
	Type         string `bson:"type"`
	SubType      string `bson:"subType"`
	DeliveryType string `bson:"deliveryType"`
}

func (t TypeTuple) ResolvedType() string {
	if t.SubType != "" {
		return fmt.Sprintf("%s/%s", t.Type, t.SubType)
	} else if t.DeliveryType != "" {
		return fmt.Sprintf("%s/%s", t.Type, t.DeliveryType)
	}
	return t.Type
}

type TypeStats struct {
	Count      int    `bson:"count"`
	LatestTime string `bson:"latestTime"`
}

type User struct {
	UserID           string               `json:"userId"`
	Email            string               `json:"email"`
	EmailVerified    bool                 `json:"emailVerified"`
	TermsAccepted    string               `json:"termsAccepted,omitempty"`
	Roles            []string             `json:"roles,omitempty"`
	Name             string               `json:"name,omitempty"`
	BirthDate        string               `json:"birthDate,omitempty"`
	DiagnosisDate    string               `json:"diagnosisDate,omitempty"`
	Devices          Devices              `json:"devices,omitempty"`
	ActiveTypesStats map[string]TypeStats `json:"activeTypesStats,omitempty"`
}

type Tool struct {
	*toolMongo.Tool
	usersStore            *storeStructuredMongo.Store
	usersRepository       *storeStructuredMongo.Repository
	metadataStore         *storeStructuredMongo.Store
	metadataRepository    *storeStructuredMongo.Repository
	dataStore             *storeStructuredMongo.Store
	dataRepository        *storeStructuredMongo.Repository
	dataSetStore          *storeStructuredMongo.Store
	dataSetRepository     *storeStructuredMongo.Repository
	dataSourcesStore      *storeStructuredMongo.Store
	dataSourcesRepository *storeStructuredMongo.Repository
	output                string
}

func NewTool() *Tool {
	return &Tool{
		Tool: toolMongo.NewTool(),
	}
}

func (t *Tool) Initialize(provider application.Provider) error {
	if err := t.Tool.Initialize(provider); err != nil {
		return err
	}

	t.CLI().Usage = "Dump users"
	t.CLI().Authors = []cli.Author{
		{
			Name:  "Darin Krauss",
			Email: "darin@tidepool.org",
		},
	}
	t.CLI().Flags = append(t.CLI().Flags,
		cli.StringFlag{
			Name:  fmt.Sprintf("%s,%s", OutputFlag, "o"),
			Usage: "output file",
		},
	)
	t.CLI().Action = func(ctx *cli.Context) error {
		if !t.ParseContext(ctx) {
			return nil
		}
		return t.execute()
	}

	if err := t.initializeUsersRepository(); err != nil {
		return err
	}
	if err := t.initializeMetadataRepository(); err != nil {
		return err
	}
	if err := t.initializeDataRepository(); err != nil {
		return err
	}
	if err := t.initializeDataSetRepository(); err != nil {
		return err
	}
	if err := t.initializeDataSourcesRepository(); err != nil {
		return err
	}

	return nil
}

func (t *Tool) Terminate() {
	t.terminateDataSourcesRepository()
	t.terminateDataSetRepository()
	t.terminateDataRepository()
	t.terminateMetadataRepository()
	t.terminateUsersRepository()

	t.Tool.Terminate()
}

func (t *Tool) ParseContext(ctx *cli.Context) bool {
	if parsed := t.Tool.ParseContext(ctx); !parsed {
		return parsed
	}

	t.output = ctx.String(OutputFlag)

	return true
}

func (t *Tool) initializeUsersRepository() error {
	t.Logger().Debug("Creating users store")

	config := t.NewMongoConfig()
	config.Database = "user"
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return errors.Wrap(err, "unable to create users store")
	}
	t.usersStore = store

	t.Logger().Debug("Creating users repository")

	t.usersRepository = store.GetRepository("users")
	return nil
}

func (t *Tool) terminateUsersRepository() {
	if t.usersRepository != nil {
		t.Logger().Debug("Destroying users repository")
		t.usersRepository = nil
	}
	if t.usersStore != nil {
		t.Logger().Debug("Destroying users store")
		t.usersStore.Terminate(context.Background())
		t.usersStore = nil
	}
}

func (t *Tool) initializeMetadataRepository() error {
	t.Logger().Debug("Creating metadata store")

	config := t.NewMongoConfig()
	config.Database = "seagull"
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return errors.Wrap(err, "unable to create metadata store")
	}
	t.metadataStore = store

	t.Logger().Debug("Creating metadata repository")

	t.metadataRepository = store.GetRepository("seagull")
	return nil
}

func (t *Tool) terminateMetadataRepository() {
	if t.metadataRepository != nil {
		t.Logger().Debug("Destroying metadata repository")
		t.metadataRepository = nil
	}
	if t.metadataStore != nil {
		t.Logger().Debug("Destroying metadata store")
		t.metadataStore.Terminate(context.Background())
		t.metadataStore = nil
	}
}

func (t *Tool) initializeDataRepository() error {
	t.Logger().Debug("Creating data store")

	config := t.NewMongoConfig()
	config.Database = "data"
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return errors.Wrap(err, "unable to create data store")
	}
	t.dataStore = store

	t.Logger().Debug("Creating data repository")

	t.dataRepository = store.GetRepository("deviceData")
	return nil
}

func (t *Tool) terminateDataRepository() {
	if t.dataRepository != nil {
		t.Logger().Debug("Destroying data repository")
		t.dataRepository = nil
	}
	if t.dataStore != nil {
		t.Logger().Debug("Destroying data store")
		t.dataStore.Terminate(context.Background())
		t.dataStore = nil
	}
}

func (t *Tool) initializeDataSetRepository() error {
	t.Logger().Debug("Creating data set repository")

	t.dataSetRepository = t.dataStore.GetRepository("deviceDataSets")
	return nil
}

func (t *Tool) terminateDataSetRepository() {
	if t.dataSetRepository != nil {
		t.Logger().Debug("Destroying data set repository")
		t.dataSetRepository = nil
	}
}

func (t *Tool) initializeDataSourcesRepository() error {
	t.Logger().Debug("Creating data sources store")

	config := t.NewMongoConfig()
	config.Database = "tidepool"
	store, err := storeStructuredMongo.NewStore(config)
	if err != nil {
		return errors.Wrap(err, "unable to create data sources store")
	}
	t.dataSourcesStore = store

	t.Logger().Debug("Creating data sources repository")

	t.dataSourcesRepository = store.GetRepository("data_sources")
	return nil
}

func (t *Tool) terminateDataSourcesRepository() {
	if t.dataSourcesRepository != nil {
		t.Logger().Debug("Destroying data sources repository")
		t.dataSourcesRepository = nil
	}
	if t.dataSourcesStore != nil {
		t.Logger().Debug("Destroying data sources store")
		t.dataSourcesStore.Terminate(context.Background())
		t.dataSourcesStore = nil
	}
}

func (t *Tool) execute() error {
	var outputWriter io.Writer

	if t.output != "" {
		outputFile, err := os.Create(t.output)
		if err != nil {
			return errors.Wrap(err, "unable to create output file")
		}
		defer outputFile.Close()
		outputWriter = outputFile
	} else {
		outputWriter = os.Stdout
	}

	return t.iterateUsers(outputWriter)
}

func (t *Tool) iterateUsers(writer io.Writer) error {
	t.Logger().Debug("Iterating users")

	userIndex := -1

	iter, err := t.usersRepository.Find(context.Background(), bson.M{})
	if err != nil {
		return errors.Wrap(err, "unable to find users")
	}

	var result struct {
		UserID        string   `bson:"userid"`
		Username      string   `bson:"username"`
		Authenticated bool     `bson:"authenticated"`
		TermsAccepted string   `bson:"termsAccepted"`
		Roles         []string `bson:"roles"`
	}
	for iter.Next(context.Background()) {
		if err := iter.Decode(result); err != nil {
			return errors.Wrap(err, "unable to iterate users")
		}
		userIndex++
		userID := result.UserID
		logger := t.Logger().WithFields(log.Fields{"userIndex": userIndex, "userId": userID})

		logger.Info("Dumping user")

		if userID == "" {
			logger.Warn("Missing user id in result from users query")
			continue
		}

		user := &User{
			UserID:        userID,
			Email:         result.Username,
			EmailVerified: result.Authenticated,
			TermsAccepted: timestampAsUTC(result.TermsAccepted),
			Roles:         result.Roles,
		}

		logger = logger.WithField("user", user)

		if err := t.getUserMetadata(userID, user, logger); err != nil {
			logger.WithError(err).Warn("Unable to get user metadata")
			continue
		}

		if email := strings.ToLower(user.Email); strings.HasSuffix(email, "@tidepool.org") || strings.HasSuffix(email, "@replacebg.org") {
			logger.Info("Filtered due to email domain")
			continue
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
			user.UserID,
			user.Name,
			user.Email,
			user.BirthDate,
		)
	}

	t.Logger().Debug("Iterated users")

	return nil
}

func (t *Tool) getUserMetadata(userID string, user *User, logger log.Logger) error {
	logger.Debug("Getting user metadata")

	var result []struct {
		Value string `bson:"value"`
	}
	opts := options.Find().SetLimit(2)
	cursor, err := t.metadataRepository.Find(context.Background(), bson.M{"userId": userID}, opts)
	if err != nil {
		return errors.Wrap(err, "unable to get user metadata")
	}
	if err := cursor.All(context.Background(), &result); err != nil {
		return errors.Wrap(err, "unable to get user metadata")
	} else if length := len(result); length == 0 {
		return errors.New("no user metadata found")
	} else if length > 1 {
		return errors.New("multiple user metadata found")
	}

	logger.Debug("Deserializing user metadata")

	var metadata struct {
		Profile *struct {
			FullName string `json:"fullName"`
			Patient  *struct {
				Birthday      string `json:"birthday"`
				DiagnosisDate string `json:"diagnosisDate"`
				IsOtherPerson bool   `json:"isOtherPerson"`
				FullName      string `json:"fullName"`
			} `json:"patient"`
		} `json:"profile"`
	}
	if err := json.Unmarshal([]byte(result[0].Value), &metadata); err != nil {
		logger.WithField("value", result[0].Value).Error("Unable to deserialize user metadata")
		return errors.Wrap(err, "unable to deserialize user metadata")
	}

	profile := metadata.Profile
	if profile == nil {
		return errors.New("user metadata missing profile")
	}

	user.Name = profile.FullName

	if patient := profile.Patient; patient != nil {
		if patient.IsOtherPerson {
			user.Name = patient.FullName
		}
		user.BirthDate = patient.Birthday
		user.DiagnosisDate = patient.DiagnosisDate
	}

	return nil
}

func (t *Tool) getUserDataDevicesDexcomAPI(userID string, user *User, logger log.Logger) error {
	logger.Debug("Getting user data devices Dexcom API")

	var result []struct {
		LatestDataTime time.Time `bson:"latestDataTime"`
	}
	opts := options.Find().SetLimit(2)
	cursor, err := t.metadataRepository.Find(context.Background(), bson.M{"userId": userID}, opts)
	if err != nil {
		return errors.Wrap(err, "unable to get user metadata")
	}
	if err := cursor.All(context.Background(), &result); err != nil {
		return err
	} else if length := len(result); length == 0 {
		return nil
	} else if length > 1 {
		return errors.New("multiple user metadata found")
	}

	device := &Device{
		Model:            "DexcomAPI",
		LatestUploadTime: timeAsUTC(result[0].LatestDataTime),
		UploadCount:      1,
	}
	user.Devices = append(user.Devices, device)

	return nil
}

func (t *Tool) getUserDataDevicesHealthKit(userID string, user *User, logger log.Logger) error {
	logger.Debug("Getting user data devices HealthKit")

	query := bson.M{
		"_userId": userID,
		"type":    "upload",
		"deviceModel": primitive.Regex{
			Pattern: "^(DexHealthKit_|HealthKit_)",
		},
	}
	if device, err := t.getUserDataDevice(query, logger); err != nil {
		return errors.Wrap(err, "unable to get user data devices HealthKit")
	} else if device != nil {
		device.Model = "HealthKit"
		user.Devices = append(user.Devices, device)
	}

	return nil
}

func (t *Tool) getUserDataDevicesOther(userID string, user *User, logger log.Logger) error {
	logger.Debug("Getting user data devices other")

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_userId": userID,
				"type":    "upload",
				"client.name": bson.M{
					"$ne": "org.tidepool.oauth.dexcom.fetch",
				},
				"deviceId": bson.M{
					"$not": primitive.Regex{
						Pattern: "^(DexHealthKit_|HealthKit_)",
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$deviceModel",
				"manufacturers": bson.M{
					"$addToSet": "$deviceManufacturers",
				},
				"tags": bson.M{
					"$addToSet": "$deviceTags",
				},
				"uploadCount": bson.M{
					"$sum": 1,
				},
				"latestUploadTime": bson.M{
					"$max": "$createdTime",
				},
			},
		},
	}
	iter, err := t.dataSetRepository.Aggregate(context.Background(), pipeline)
	if err != nil {
		return errors.Wrap(err, "unable to iterate user data devices other")
	}

	logger.Debug("Iterating user data devices other")

	var result struct {
		Model            string     `bson:"_id"`
		Manufacturers    [][]string `bson:"manufacturers"`
		Tags             [][]string `bson:"tags"`
		LatestUploadTime string     `bson:"latestUploadTime"`
		UploadCount      int        `bson:"uploadCount"`
	}
	for iter.Next(context.Background()) {
		err := iter.Decode(&result)
		if err != nil {
			return errors.Wrap(err, "unable to iterate user data devices other")
		}

		device := &Device{
			Model:            result.Model,
			Manufacturers:    mergeStringArrays(result.Manufacturers),
			Tags:             mergeStringArrays(result.Tags),
			LatestUploadTime: timestampAsUTC(result.LatestUploadTime),
			UploadCount:      result.UploadCount,
		}
		user.Devices = append(user.Devices, device)
	}

	return nil
}

func (t *Tool) getUserDataDevice(query interface{}, logger log.Logger) (*Device, error) {
	var result []struct {
		CreatedTime string `bson:"createdTime"`
	}
	opts := options.Find().SetSort(bson.M{"createdTime": -1}).SetLimit(2)
	cursor, err := t.dataSetRepository.Find(context.Background(), query, opts)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.Background(), &result); err != nil {
		return nil, err
	} else if len(result) == 0 {
		return nil, nil
	}

	uploadCount, err := t.dataSetRepository.CountDocuments(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return &Device{
		LatestUploadTime: timestampAsUTC(result[0].CreatedTime),
		UploadCount:      int(uploadCount),
	}, nil
}

func (t *Tool) getUserDataActiveTypesStats(userID string, user *User, logger log.Logger) error {
	logger.Debug("Getting user data active types stats")

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_active": true,
				"_userId": userID,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"type":         "$type",
					"subType":      "$subType",
					"deliveryType": "$deliveryType",
				},
				"count": bson.M{
					"$sum": 1,
				},
				"latestTime": bson.M{
					"$max": "$time",
				},
			},
		},
	}
	cursor, err := t.dataRepository.Aggregate(context.Background(), pipeline)
	if err != nil {
		return errors.Wrap(err, "unable to iterate data active types stats")
	}

	var result struct {
		TypeTuple TypeTuple `bson:"_id"`
		TypeStats `bson:",inline"`
	}
	for cursor.Next(context.Background()) {
		err := cursor.Decode(&result)
		if err != nil {
			return errors.Wrap(err, "unable to iterate data active types stats")
		}

		if user.ActiveTypesStats == nil {
			user.ActiveTypesStats = map[string]TypeStats{}
		}
		user.ActiveTypesStats[result.TypeTuple.ResolvedType()] = result.TypeStats
	}

	return nil
}

func timestampAsUTC(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	tm, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return ""
	}
	return timeAsUTC(tm)
}

func timeAsUTC(tm time.Time) string {
	return tm.Truncate(time.Second).UTC().Format(time.RFC3339Nano)
}

func stringInStringArray(str string, strArray []string) bool {
	for _, s := range strArray {
		if s == str {
			return true
		}
	}
	return false
}

func mergeStringArrays(strArrays [][]string) []string {
	switch len(strArrays) {
	case 0:
		return nil
	case 1:
		return strArrays[0]
	}

	strMap := map[string]interface{}{}
	for _, strArray := range strArrays {
		for _, str := range strArray {
			strMap[str] = nil
		}
	}

	var strArray []string
	for str := range strMap {
		strArray = append(strArray, str)
	}

	return strArray
}
