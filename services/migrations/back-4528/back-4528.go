package main

import (
	"context"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/errors"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

// deprecatedTasksSelector selects the tasks of the retired summary update and migration runners,
// of types org.tidepool.summary.update.* and org.tidepool.summary.migrate.*
var deprecatedTasksSelector = bson.M{"type": bson.M{"$regex": `^org\.tidepool\.summary\.`}}

func main() {
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*migrationMongo.Migration
}

func NewMigration() *Migration {
	return &Migration{
		Migration: migrationMongo.NewMigration(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Migration.Initialize(provider); err != nil {
		return err
	}

	m.CLI().Usage = "BACK-4528: Delete the tasks of the retired summary runners"
	m.CLI().Description = "BACK-4528: The summary update and migration task runners are replaced by upload postprocess work,\n" +
		"   and the task service no longer runs or creates their tasks. This migration deletes the tasks left behind.\n" +
		"   Run it once the release removing the runners is deployed, as the task service of an earlier release\n" +
		"   recreates the tasks when it starts."
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Todd Kazakov",
			Email: "todd@tidepool.org",
		},
	}

	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}
		return m.execute()
	}

	return nil
}

func (m *Migration) execute() error {
	ctx := context.Background()

	// The task service stores its tasks in the database of the default store configuration, which
	// the migration loads as well
	store, err := storeStructuredMongo.NewStore(m.NewMongoConfig())
	if err != nil {
		return errors.Wrap(err, "unable to create store")
	}
	defer store.Terminate(ctx)

	count, err := deleteDeprecatedTasks(ctx, store.GetRepository("tasks"), m.DryRun())
	if err != nil {
		return err
	}

	if m.DryRun() {
		m.Logger().Infof("Found %d deprecated tasks to delete", count)
	} else {
		m.Logger().Infof("Deleted %d deprecated tasks", count)
	}
	return nil
}

// deleteDeprecatedTasks deletes the deprecated tasks, or on a dry run only counts them, and reports
// how many there were
func deleteDeprecatedTasks(ctx context.Context, repository *storeStructuredMongo.Repository, dryRun bool) (int64, error) {
	if dryRun {
		count, err := repository.CountDocuments(ctx, deprecatedTasksSelector)
		if err != nil {
			return 0, errors.Wrap(err, "unable to count deprecated tasks")
		}
		return count, nil
	}

	result, err := repository.DeleteMany(ctx, deprecatedTasksSelector)
	if err != nil {
		return 0, errors.Wrap(err, "unable to delete deprecated tasks")
	}
	return result.DeletedCount, nil
}
