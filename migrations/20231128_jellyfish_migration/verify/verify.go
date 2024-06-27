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
	mongoURI    string
	ref         string
	uploadOneID string
	uploadTwoID string
}

const MongoURIFlag = "uri"
const UploadIDOneFlag = "upload-id-one"
const UploadIDTwoFlag = "upload-id-two"
const ReferenceFlag = "reference"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	verifier := NewVerifier(ctx)
	verifier.RunAndExit()
	log.Println("finished verification")
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
		os.Exit(1)
	}

	m.CLI().Action = func(ctx *cli.Context) error {

		var err error
		m.client, err = mongo.Connect(m.ctx, options.Client().ApplyURI(strings.TrimSpace(m.config.mongoURI)))
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %w", err)
		}
		defer m.client.Disconnect(m.ctx)

		m.verificationUtil, err = utils.NewVerifier(
			m.ctx,
			m.client.Database("data").Collection("deviceData"),
		)

		if err != nil {
			return fmt.Errorf("unable to create verification utils : %w", err)
		}

		return m.verificationUtil.Verify(m.config.ref, m.config.uploadOneID, m.config.uploadTwoID)
	}

	if err := m.CLI().Run(os.Args); err != nil {
		if m.client != nil {
			m.client.Disconnect(m.ctx)
		}
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
			Name:        UploadIDOneFlag,
			Usage:       "uploadID of the first dataset",
			Destination: &m.config.uploadOneID,
			Required:    true,
		},
		cli.StringFlag{
			Name:        UploadIDTwoFlag,
			Usage:       "uploadID of the second dataset",
			Destination: &m.config.uploadTwoID,
			Required:    true,
		},
		cli.StringFlag{
			Name:        ReferenceFlag,
			Usage:       "comparison reference",
			Value:       "todo-reference",
			Destination: &m.config.ref,
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
