package cmd

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/tools/tapi/api"
)

const (
	DatasetIDFlag = "dataset-id"
	DeletedFlag   = "deleted"
	PageFlag      = "page"
	SizeFlag      = "size"
)

func DatasetCommands() cli.Commands {
	return cli.Commands{
		{
			Name:  "dataset",
			Usage: "dataset management",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "list datasets",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  UserIDFlag,
							Usage: "`USERID` of the user to list datasets",
						},
						cli.BoolFlag{
							Name:  DeletedFlag,
							Usage: "include deleted datasets in the list",
						},
						cli.IntFlag{
							Name:  PageFlag,
							Usage: "pagination `PAGE`",
						},
						cli.IntFlag{
							Name:  SizeFlag,
							Usage: "pagination `SIZE`",
						},
					),
					Before: ensureNoArgs,
					Action: datasetList,
				},
				{
					Name:  "delete",
					Usage: "delete dataset",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  DatasetIDFlag,
							Usage: "`DATASETID` of the dataset to delete",
						},
					),
					Before: ensureNoArgs,
					Action: datasetDelete,
				},
			},
		},
	}
}

func datasetList(c *cli.Context) error {
	var filter *api.Filter
	var pagination *api.Pagination

	if c.IsSet(DeletedFlag) {
		if filter == nil {
			filter = &api.Filter{}
		}
		deleted := c.Bool(DeletedFlag)
		filter.Deleted = &deleted
	}
	if c.IsSet(PageFlag) {
		if pagination == nil {
			pagination = &api.Pagination{}
		}
		page := c.Int(PageFlag)
		pagination.Page = &page
	}
	if c.IsSet(SizeFlag) {
		if pagination == nil {
			pagination = &api.Pagination{}
		}
		size := c.Int(SizeFlag)
		pagination.Size = &size
	}

	responseArray, err := API(c).ListDatasets(c.String(UserIDFlag), filter, pagination)
	if err != nil {
		return err
	}

	for _, dataset := range responseArray.Data {
		if err = reportMessageWithJSON(c, dataset); err != nil {
			return err
		}
	}

	return nil
}

func datasetDelete(c *cli.Context) error {
	if err := API(c).DeleteDataset(c.String(DatasetIDFlag)); err != nil {
		return err
	}

	return reportMessage(c, "Dataset deleted.")
}
