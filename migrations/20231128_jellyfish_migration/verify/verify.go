package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Verify struct {
	ctx              context.Context
	cli              *cli.App
	config           *config
	client           *mongo.Client
	verificationUtil *DataVerify
}

type config struct {
	mongoURI          string
	findBlobs         bool
	verifyDeduped     bool
	verifyDevice      bool
	sameAccount       bool
	platformUploadID  string
	jellyfishUploadID string
	uploadIdDeduped   string
	deviceID          string
	userID            string
	dataTypes         string
}

const MongoURIFlag = "uri"
const PlatformUploadIDFlag = "upload-id-platform"
const JellyfishUploadIDFlag = "upload-id-jellyfish"
const SameAccountFlag = "same-account"
const FindBlobFlag = "find-blobs"
const VerifyDedupedFlag = "verify-deduped"
const VerifyDeviceFlag = "verify-device"
const UploadIDFlag = "upload-id"
const DeviceIDFlag = "device-id"
const UserIDFlag = "user-id"
const DataTypesFlag = "data-types"
const UseSubsetFlag = "use-subset"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	verifier := NewVerifier(ctx)
	verifier.RunAndExit()
}

func NewVerifier(ctx context.Context) *Verify {
	return &Verify{
		config: &config{sameAccount: false},
		ctx:    ctx,
		cli:    cli.NewApp(),
	}
}

func (m *Verify) RunAndExit() {
	if err := m.Initialize(); err != nil {
		log.Printf("error during Initialize [%s]", err.Error())
		os.Exit(1)
	}

	m.CLI().Action = func(ctx *cli.Context) error {

		var err error
		m.client, err = mongo.Connect(m.ctx, options.Client().ApplyURI(strings.TrimSpace(m.config.mongoURI)))
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %w", err)
		}
		defer m.client.Disconnect(m.ctx)
		log.Printf("using config %#v", m.config)

		if m.config.findBlobs {
			m.verificationUtil, err = NewDataVerify(
				m.ctx,
				m.client.Database("data").Collection("deviceDataSets"),
			)

			if err != nil {
				return fmt.Errorf("unable to create verification utils : %w", err)
			}
			return m.verificationUtil.WriteBlobIDs()
		}

		if m.config.verifyDeduped {
			m.verificationUtil, err = NewDataVerify(
				m.ctx,
				m.client.Database("data").Collection("deviceData"),
			)

			if err != nil {
				return fmt.Errorf("unable to create verification utils : %w", err)
			}
			return m.verificationUtil.VerifyDeduped(m.config.uploadIdDeduped, strings.Split(m.config.dataTypes, ","))
		}

		if m.config.verifyDevice {
			m.verificationUtil, err = NewDataVerify(
				m.ctx,
				m.client.Database("data").Collection("deviceData"),
			)

			if err != nil {
				return fmt.Errorf("unable to create verification utils : %w", err)
			}
			return m.verificationUtil.VerifyDeviceUploads(m.config.userID, m.config.deviceID, strings.Split(m.config.dataTypes, ","))
		}

		m.verificationUtil, err = NewDataVerify(
			m.ctx,
			m.client.Database("data").Collection("deviceData"),
		)

		if err != nil {
			return fmt.Errorf("unable to create verification utils : %w", err)
		}

		err = m.verificationUtil.VerifyUploadDifferences(m.config.platformUploadID, m.config.jellyfishUploadID, strings.Split(m.config.dataTypes, ","), m.config.sameAccount)
		if err != nil {
			log.Printf("error running verify : %s", err.Error())
		}
		return nil
	}

	if err := m.CLI().Run(os.Args); err != nil {
		if m.client != nil {
			m.client.Disconnect(m.ctx)
		}
		log.Printf("error during Run [%s]", err.Error())
		os.Exit(1)
	}
}

func (m *Verify) Initialize() error {
	m.CLI().Usage = "dataset verifictaion tool to compare dataset-a with dataset-b"
	m.CLI().Authors = []cli.Author{
		{
			Name:  "J H BATE",
			Email: "jamie@tidepool.org",
		},
	}
	m.CLI().Flags = append(m.CLI().Flags,
		cli.StringFlag{
			Name:        PlatformUploadIDFlag,
			Usage:       "uploadID of the first platform dataset",
			Destination: &m.config.platformUploadID,
			Required:    false,
		},
		cli.StringFlag{
			Name:        JellyfishUploadIDFlag,
			Usage:       "uploadID of the second jellyfish dataset",
			Destination: &m.config.jellyfishUploadID,
			Required:    false,
		},
		cli.BoolFlag{
			Name:        SameAccountFlag,
			Usage:       "the datasets are uploaded to the same account",
			Destination: &m.config.sameAccount,
			Required:    false,
		},
		cli.StringFlag{
			Name:        UploadIDFlag,
			Usage:       "uploadID of the dataset to check deduping of",
			Destination: &m.config.uploadIdDeduped,
			Required:    false,
		},
		cli.StringFlag{
			Name:        DeviceIDFlag,
			Usage:       "deviceID of the datasets to check",
			Destination: &m.config.deviceID,
			Required:    false,
		},
		cli.StringFlag{
			Name:        UserIDFlag,
			Usage:       "userID of the device to check",
			Destination: &m.config.userID,
			Required:    false,
		},
		cli.StringFlag{
			Name:        DataTypesFlag,
			Usage:       "comma seperated list of data types to compare",
			Destination: &m.config.dataTypes,
			Required:    false,
			Value:       strings.Join(DatasetTypes, ","),
		},
		cli.BoolFlag{
			Name:        FindBlobFlag,
			Usage:       "find all blobs for running data verifcation with",
			Destination: &m.config.findBlobs,
			Required:    false,
		},
		cli.BoolFlag{
			Name:        VerifyDedupedFlag,
			Usage:       "verify that a dataset has been deduplicated",
			Destination: &m.config.verifyDeduped,
			Required:    false,
		},
		cli.BoolFlag{
			Name:        VerifyDeviceFlag,
			Usage:       "verify a device datasets",
			Destination: &m.config.verifyDevice,
			Required:    false,
		},
		cli.StringFlag{
			Name:        MongoURIFlag,
			Usage:       "mongo connection URI",
			Destination: &m.config.mongoURI,
			Required:    false,
			//uri string comes from file called `uri`
			FilePath: "./uri",
		},
	)
	return nil
}

func (m *Verify) CLI() *cli.App {
	return m.cli
}
