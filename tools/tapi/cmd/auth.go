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

import "github.com/urfave/cli"

const (
	TokenFlag = "token"
)

func AuthCommands() cli.Commands {
	return cli.Commands{
		{
			Name:   "server-login",
			Usage:  "authenticate and remember the new server session",
			Flags:  CommandFlags(),
			Before: ensureNoArgs,
			Action: authServerLogin,
		},
		{
			Name:  "login",
			Usage: "authenticate and remember the new user session",
			Flags: CommandFlags(
				cli.StringFlag{
					Name:  EmailFlag,
					Usage: "`EMAIL` for authentication",
				},
				cli.StringFlag{
					Name:  PasswordFlag,
					Usage: "`PASSWORD` for authentication",
				},
			),
			Before: ensureNoArgs,
			Action: authLogin,
		},
		{
			Name:   "logout",
			Usage:  "logout and forget the current session",
			Flags:  CommandFlags(),
			Before: ensureNoArgs,
			Action: authLogout,
		},
		{
			Name:   "whoami",
			Usage:  "display the current session",
			Flags:  CommandFlags(),
			Before: ensureNoArgs,
			Action: authWhoami,
		},
		{
			Name:  "check-session",
			Usage: "check a session for validity",
			Flags: CommandFlags(
				cli.StringFlag{
					Name:  TokenFlag,
					Usage: "`TOKEN` to check",
				},
			),
			Before: ensureNoArgs,
			Action: authCheckSession,
		},
	}
}

func authServerLogin(c *cli.Context) error {
	if err := API(c).ServerLogin(); err != nil {
		return err
	}

	return reportMessage(c, "Server session created.")
}

func authLogin(c *cli.Context) error {
	email := c.String(EmailFlag)
	if email == "" {
		var err error
		if email, err = readFromConsole("Email: "); err != nil {
			return err
		}
	}

	password := c.String(PasswordFlag)
	if password == "" {
		var err error
		if password, err = readFromConsoleNoEcho("Password: "); err != nil {
			return err
		}
	}

	_, err := API(c).Login(email, password)
	if err != nil {
		return err
	}

	return reportMessage(c, "Logged in.")
}

func authLogout(c *cli.Context) error {
	if err := API(c).Logout(); err != nil {
		return err
	}

	return reportMessage(c, "Logged out.")
}

func authWhoami(c *cli.Context) error {
	token, err := API(c).RefreshToken()
	if err != nil {
		return err
	}

	return reportMessageWithJSON(c, token)
}

func authCheckSession(c *cli.Context) error {
	checked, err := API(c).CheckToken(c.String(TokenFlag))
	if err != nil {
		return err
	}

	return reportMessageWithJSON(c, checked)
}
