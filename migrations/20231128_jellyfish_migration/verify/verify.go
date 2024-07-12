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

	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
)

type Verify struct {
	ctx              context.Context
	cli              *cli.App
	config           *config
	client           *mongo.Client
	verificationUtil *utils.DataVerify
}

type config struct {
	mongoURI          string
	findBlobs         bool
	useSubset         bool
	platformUploadID  string
	jellyfishUploadID string
	dataTypes         string
}

const MongoURIFlag = "uri"
const PlatformUploadIDFlag = "upload-id-platform"
const JellyfishUploadIDFlag = "upload-id-jellyfish"
const FindBlobFlag = "find-blobs"
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
		config: &config{},
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

		if m.config.findBlobs {
			m.verificationUtil, err = utils.NewVerifier(
				m.ctx,
				m.client.Database("data").Collection("deviceDataSets"),
			)

			if err != nil {
				return fmt.Errorf("unable to create verification utils : %w", err)
			}

			ids, err := m.verificationUtil.FetchBlobIDs()
			if err != nil {
				return err
			}
			for i, v := range ids {
				log.Printf("%d - %v", i, v)
			}

			return nil
		}

		m.verificationUtil, err = utils.NewVerifier(
			m.ctx,
			m.client.Database("data").Collection("deviceData"),
		)

		if err != nil {
			return fmt.Errorf("unable to create verification utils : %w", err)
		}

		err = m.verificationUtil.Verify("ref", m.config.platformUploadID, m.config.jellyfishUploadID, strings.Split(m.config.dataTypes, ","), m.config.useSubset)
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
		cli.StringFlag{
			Name:        DataTypesFlag,
			Usage:       "comma seperated list of data types to compare",
			Destination: &m.config.dataTypes,
			Required:    false,
			Value:       strings.Join(utils.DatasetTypes, ","),
		},
		cli.BoolFlag{
			Name:        FindBlobFlag,
			Usage:       "find all blobs for running data verifcation with",
			Destination: &m.config.findBlobs,
			Required:    false,
		},
		cli.BoolFlag{
			Name:        UseSubsetFlag,
			Usage:       "use a subset of data to compare",
			Destination: &m.config.useSubset,
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
