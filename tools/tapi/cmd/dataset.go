package cmd

import (
	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/tools/tapi/api"
)

const (
	DataSetIDFlag = "data-set-id"
	DeletedFlag   = "deleted"
	PageFlag      = "page"
	SizeFlag      = "size"
)

func DataSetCommands() cli.Commands {
	return cli.Commands{
		{
			Name:  "data-set",
			Usage: "data set management",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "list data sets",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  UserIDFlag,
							Usage: "`USERID` of the user to list data sets",
						},
						cli.BoolFlag{
							Name:  DeletedFlag,
							Usage: "include deleted data sets in the list",
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
					Action: dataSetList,
				},
				{
					Name:  "delete",
					Usage: "delete data set",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  DataSetIDFlag,
							Usage: "`DATASETID` of the data set to delete",
						},
					),
					Before: ensureNoArgs,
					Action: dataSetDelete,
				},
			},
		},
	}
}

func dataSetList(c *cli.Context) error {
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

	responseArray, err := API(c).ListDataSets(c.String(UserIDFlag), filter, pagination)
	if err != nil {
		return err
	}

	for _, dataSet := range responseArray.Data {
		if err = reportMessageWithJSON(c, dataSet); err != nil {
			return err
		}
	}

	return nil
}

func dataSetDelete(c *cli.Context) error {
	if err := API(c).DeleteDataSet(c.String(DataSetIDFlag)); err != nil {
		return err
	}

	return reportMessage(c, "Data set deleted.")
}
