package main

import (
	"fmt"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
	"github.com/tidepool-org/platform/data/types/basal/temporary"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/metadata"
	"time"

	"context"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)
var (
	DBContextTimeout = time.Duration(20)*time.Second
)

func init() {
	orm.SetTableNameInflector(func(s string) string {
		return  s
	})
}

func NewDbContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), DBContextTimeout)
	return ctx
}

func connectToDatabase() *pg.DB {
	url := fmt.Sprintf("postgres://postgres@localhost:5432/postgres?sslmode=disable")
	opt, err := pg.ParseURL(url)
	if err != nil {
		panic(err)
	}

	db := pg.Connect(opt)
	fmt.Println("Trying to connect to db")

	ctx := NewDbContext()

	//db.AddQueryHook(dbLogger{})


	// Check if connection credentials are valid and PostgreSQL is up and running.
	if err := db.Ping(ctx); err != nil {
		fmt.Println("Error: ", err)
		return nil
	}
	fmt.Println("Connected successfully")

	return db
}

func main() {
	db := connectToDatabase()
	defer db.Close()

	//createUpload(db)
	//createCbg(db)
	createBasal(db)

}

/*
{
	"_id" : "02ti7rurmpu040j225n29qi2osku9cun",
	"computerTime" : "2017-03-31T22:44:51",
	"timeProcessing" : "none",
	"time" : "2017-03-31T22:44:51.000Z",
	"byUser" : "2d73f64242",
	"deviceTags" : [
		"cgm"
	],
	"uploadId" : "upid_6a158bb3947abb19ccd7df2b8d1a6222",
	"type" : "upload",
	"deviceManufacturers" : [
		"Dexcom"
	],
	"timezone" : "America/Los_Angeles",
	"timezoneOffset" : -420,
	"version" : "org.tidepool.blipnotes:1.1.2:312",
	"deviceId" : "HealthKit_DexG5_E753A756-6CA6-40D5-BB81-58DC04FA2DAB",
	"guid" : "14D8522A-9E05-4645-866E-3943ED3B3E6D",
	"deviceModel" : "HealthKit_DexG5",
	"deviceSerialNumber" : "",
	"_userId" : "2d73f64242",
	"_groupId" : "1bb0732297",
	"id" : "eenfha51upq6u24r89656srqhnhnerl7",
	"createdTime" : "2017-03-31T22:44:52.497Z",
	"_version" : 0,
	"_active" : true,
	"_schemaVersion" : 3
}
 */

func createUpload(db *pg.DB) {
	curTime := time.Now().Format(time.RFC3339)
	deviceManufacturers := []string{"Dexcom"}
	deviceTags := []string{"cgm", "pp"}
	uploadId := "upid_6a158bb3947abb19ccd7df2b8d1a6222"
	timeZoneName := "America/Los_Angeles";
	timeZoneOffset := -420

	deviceId := "HealthKit_DexG5_E753A756-6CA6-40D5-BB81-58DC04FA2DAB"
	guid := "14D8522A-9E05-4645-866E-3943ED3B3E6D"
	deviceModel := "HealthKit_DexG5"
	deviceSerialNumber := ""
	userId := "2d73f64242"
	//groupId := "1bb0732297"
	id := "eenfha51upq6u24r89656srqhnhnerl7"
	createdTime := "2017-03-31T22:44:52.497Z"
	version := "org.tidepool.blipnotes:1.1.2:312"

	upload := upload.Upload{
		Base: types.Base{
			CreatedTime: &createdTime,
			DeviceID: &deviceId,
			GUID: &guid,
			ID: &id,
			Time: &curTime,
		    UserID: &userId,
			UploadID: &uploadId,
			TimeZoneName: &timeZoneName,
			TimeZoneOffset: &timeZoneOffset,
		},
		DeviceModel: &deviceModel,
		DeviceManufacturers: &deviceManufacturers,
		DeviceSerialNumber: &deviceSerialNumber,
		DeviceTags: &deviceTags,
		Version: &version,
	}

	if err := db.Insert(&upload); err != nil {
		fmt.Println("Error: ", err)
	}
}

/*
{
	"_id" : "0007dpsd89jea1bln7oqtmecq135s0e9",
	"uploadId" : "upid_f386bccbaa8dfde4f7bcecb4a343df2d",
	"value" : 4.88465823212007,
	"units" : "mmol/L",
	"deviceId" : "HealthKit_DexG5_5B42B589-8BDA-4DBC-8DBA-CCF9FF7BAC69",
	"guid" : "3903E89B-2E97-4314-ABA3-B78641E384F3",
	"type" : "cbg",
	"time" : "2016-04-02T18:33:56.000Z",
	"payload" : {
		"Trend Rate" : 0.2,
		"Status" : "IN_RANGE",
		"Trend Arrow" : "Flat",
		"Transmitter Time" : "2016-04-02T18:33:49.000Z",
		"HKDeviceName" : "10386270000221"
	},
	"_userId" : "c405d00c22",
	"_groupId" : "8c539f31e4",
	"id" : "sig3cpjifetq1j59uvo2up6pp3msg0ce",
	"createdTime" : "2017-10-24T18:34:41.144Z",
	"_version" : 0,
	"_active" : true,
	"_schemaVersion" : 3
}
 */
func createCbg(db *pg.DB) {
	curTime := time.Now().Format(time.RFC3339)
	value := 4.88465823212007
	units := "mmol/L"
	uploadId := "upid_f386bccbaa8dfde4f7bcecb4a343df2d"

	deviceId := "HealthKit_DexG5_E753A756-6CA6-40D5-BB81-58DC04FA2DAB"
	guid := "3903E89B-2E97-4314-ABA3-B78641E384F3"
	userId := "c405d00c22"
	//groupId := "8c539f31e4"
	id := "sig3cpjifetq1j59uvo2up6pp3msg0ce"
	createdTime := "2017-10-24T18:34:41.144Z"

	payload := metadata.Metadata {
		"Trend Rate" : 0.2,
		"Status" : "IN_RANGE",
		"Trend Arrow" : "Flat",
		"Transmitter Time" : "2016-04-02T18:33:49.000Z",
		"HKDeviceName" : "10386270000221",
	}

	cbg := continuous.Continuous{
		Glucose: glucose.Glucose{
			blood.Blood{
				Base: types.Base{
					CreatedTime: &createdTime,
					DeviceID: &deviceId,
					GUID: &guid,
					ID: &id,
					Payload: &payload,
					Time: &curTime,
					UserID: &userId,
					UploadID: &uploadId,
				},
				Value: &value,
				Units: &units,
			},
		},
	}
	if err := db.Insert(&cbg); err != nil {
		fmt.Println("Error: ", err)
	}
}

/*
{
	"_id" : "0009fbj2ij8594l25tf9no79o1rk3si5_0",
	"time" : "2017-04-26T13:05:41.000Z",
	"timezoneOffset" : -420,
	"clockDriftOffset" : 0,
	"conversionOffset" : 0,
	"deviceTime" : "2017-04-26T06:05:41",
	"source" : "carelink",
	"type" : "basal",
	"deliveryType" : "temp",
	"rate" : 1.5,
	"duration" : 1800000,
	"suppressed" : {
		"time" : "2017-04-26T13:00:41.000Z",
		"timezoneOffset" : -420,
		"clockDriftOffset" : 0,
		"conversionOffset" : 0,
		"deviceTime" : "2017-04-26T06:00:41",
		"source" : "carelink",
		"type" : "basal",
		"deliveryType" : "temp",
		"rate" : 1.475,
		"duration" : 1800000,
		"suppressed" : {
			"time" : "2017-04-26T12:45:42.000Z",
			"timezoneOffset" : -420,
			"clockDriftOffset" : 0,
			"conversionOffset" : 0,
			"deviceTime" : "2017-04-26T05:45:42",
			"source" : "carelink",
			"type" : "basal",
			"deliveryType" : "scheduled",
			"scheduleName" : "standard",
			"rate" : 0.8,
			"duration" : 6258000,
			"payload" : {
				"rawSeqNums" : [
					1985
				],
				"rawUploadId" : 55748370
			},
			"jsDate" : "2017-04-26T05:45:42.000Z",
			"deviceId" : "Paradigm Revel - 723-=-55748370",
			"index" : 4253
		},
		"payload" : {
			"rawSeqNums" : [
				1984
			],
			"rawUploadId" : 55748370
		},
		"jsDate" : "2017-04-26T06:00:41.000Z",
		"deviceId" : "Paradigm Revel - 723-=-55748370",
		"index" : 4254
	},
	"payload" : {
		"rawSeqNums" : [
			1983
		],
		"rawUploadId" : 55748370
	},
	"deviceId" : "Paradigm Revel - 723-=-55748370",
	"uploadId" : "upid_88bd54ea5da1",
	"guid" : "4a11d5fa-baae-4601-aa78-c4f6fe65ece4",
	"_userId" : "dc85192a07",
	"_groupId" : "4803e6e091",
	"id" : "cfli193puvcrfi4jb8mk8ia7huljh1g0",
	"createdTime" : "2017-05-23T19:12:51.024Z",
	"_version" : 0,
	"_active" : false,
	"_schemaVersion" : 3,
	"_archivedTime" : 1495566771027
}

 */

func createBasal(db *pg.DB) {
	deviceId := "Paradigm Revel - 723-=-55748370"
	uploadId := "upid_88bd54ea5da1"
	guid := "4a11d5fa-baae-4601-aa78-c4f6fe65ece4"
	userId := "dc85192a07"
	//_groupId := "4803e6e091"
	id := "cfli193puvcrfi4jb8mk8ia7huljh1g0"
	createdTime := "2017-05-23T19:12:51.024Z"
	//_version := 0
	active := false
	//archivedTime := 1495566771027
	curTime := time.Now().Format(time.RFC3339)
	timezoneOffset := -420
	clockDriftOffset := 0
	conversionOffset := 0
	deviceTime := "2017-04-26T06:05:41"
	source := "carelink"
	deliveryType := "temp"
	rate := 1.5
	duration := 1800000

	payload := metadata.Metadata {
		"rawSeqNums" : []int{1983},
		"rawUploadId" : 55748370,
	}

	basal := temporary.Temporary{
		Basal : basal.Basal{
			Base: types.Base{
				CreatedTime: &createdTime,
				//ArchivedTime: &archivedTime,
				DeviceID: &deviceId,
				GUID: &guid,
				ID: &id,
				Time: &curTime,
				UserID: &userId,
				UploadID: &uploadId,
				//TimeZoneName: &timeZoneName,
				TimeZoneOffset: &timezoneOffset,
				Active: active,
				ClockDriftOffset: &clockDriftOffset,
				ConversionOffset: &conversionOffset,
				DeviceTime: &deviceTime,
				Source: &source,
				Payload: &payload,
			},
			DeliveryType: deliveryType,
		},
		Rate: &rate,
		Duration: &duration,
	}

	if err := db.Insert(&basal); err != nil {
		fmt.Println("Basal Error: ", err)
	}
}

/*
{
	"_id" : "0019svb1cfk1bmge4m2qml32g5d9fh0f",
	"time" : "2016-08-08T20:24:51.000Z",
	"timezoneOffset" : -420,
	"clockDriftOffset" : 0,
	"conversionOffset" : 0,
	"deviceTime" : "2016-08-08T13:24:51",
	"deviceId" : "InsOmn-130130814",
	"type" : "bolus",
	"subType" : "normal",
	"normal" : 11.4,
	"payload" : {
		"logIndices" : [
			6873
		]
	},
	"uploadId" : "upid_03df4e06a055",
	"guid" : "bd993e4b-068e-4999-afa3-3070def06127",
	"_userId" : "c8b5078db0",
	"_groupId" : "4ff6ff34e0",
	"id" : "be23fre4bm1q80lse2tcvamh33n0euls",
	"createdTime" : "2016-09-30T18:11:06.280Z",
	"_version" : 0,
	"_active" : true,
	"_schemaVersion" : 3
}
 */
