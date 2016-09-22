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

	"github.com/tidepool-org/platform/version"
)

func VersionCommands(versionReporter version.Reporter) cli.Commands {
	return cli.Commands{
		{
			Name:   "version",
			Usage:  "print the version",
			Action: versionLong(versionReporter),
			Subcommands: []cli.Command{
				{
					Name:   "base",
					Usage:  "print the base version",
					Before: ensureNoArgs,
					Action: versionBase(versionReporter),
				},
				{
					Name:   "short-commit",
					Usage:  "print the short git commit",
					Before: ensureNoArgs,
					Action: versionShortCommit(versionReporter),
				},
				{
					Name:   "full-commit",
					Usage:  "print the full git commit",
					Before: ensureNoArgs,
					Action: versionFullCommit(versionReporter),
				},
				{
					Name:   "short",
					Usage:  "print the short version, including the short git commit",
					Before: ensureNoArgs,
					Action: versionShort(versionReporter),
				},
				{
					Name:   "long",
					Usage:  "print the long version, including the full git commit",
					Before: ensureNoArgs,
					Action: versionLong(versionReporter),
				},
			},
		},
	}
}

func versionBase(versionReporter version.Reporter) cli.ActionFunc {
	return func(c *cli.Context) error {
		return reportMessage(c, versionReporter.Base())
	}
}

func versionShortCommit(versionReporter version.Reporter) cli.ActionFunc {
	return func(c *cli.Context) error {
		return reportMessage(c, versionReporter.ShortCommit())
	}
}

func versionFullCommit(versionReporter version.Reporter) cli.ActionFunc {
	return func(c *cli.Context) error {
		return reportMessage(c, versionReporter.FullCommit())
	}
}

func versionShort(versionReporter version.Reporter) cli.ActionFunc {
	return func(c *cli.Context) error {
		return reportMessage(c, versionReporter.Short())
	}
}

func versionLong(versionReporter version.Reporter) cli.ActionFunc {
	return func(c *cli.Context) error {
		return reportMessage(c, versionReporter.Long())
	}
}
