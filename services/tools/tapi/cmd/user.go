package cmd

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/tidepool-org/platform/services/tools/tapi/api"
	"github.com/tidepool-org/platform/user"
)

const (
	UserIDFlag   = "user-id"
	EmailFlag    = "email"
	PasswordFlag = "password"
	RoleFlag     = "role"
)

func UserCommands() cli.Commands {
	return cli.Commands{
		{
			Name:  "user",
			Usage: "user management",
			Subcommands: []cli.Command{
				{
					Name:  "get",
					Usage: "get a user by id or email",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  UserIDFlag,
							Usage: "`USERID` of the user to get",
						},
						cli.StringFlag{
							Name:  EmailFlag,
							Usage: "`EMAIL` of the user to get",
						},
					),
					Before: ensureNoArgs,
					Action: userGet,
				},
				{
					Name:  "find",
					Usage: "find all users matching the specified search criteria",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  RoleFlag,
							Usage: "find users matching the specified `ROLE`",
						},
					),
					Before: ensureNoArgs,
					Action: userFind,
				},
				{
					Name:  "add-role",
					Usage: "add the specified role to the user specified by id",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  UserIDFlag,
							Usage: "`USERID` of the user to update",
						},
						cli.StringSliceFlag{
							Name:  RoleFlag,
							Usage: "`ROLE` to add to the user",
						},
					),
					Before: ensureNoArgs,
					Action: userAddRoles,
				},
				{
					Name:  "remove-role",
					Usage: "remove the specified role from the user specified by id",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  UserIDFlag,
							Usage: "`USERID` of the user to update",
						},
						cli.StringSliceFlag{
							Name:  RoleFlag,
							Usage: "`ROLE` to remove from the user",
						},
					),
					Before: ensureNoArgs,
					Action: userRemoveRoles,
				},
				{
					Name:  "delete",
					Usage: "delete a user",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  UserIDFlag,
							Usage: "`USERID` of the user to delete",
						},
						cli.StringFlag{
							Name:  PasswordFlag,
							Usage: "`PASSWORD` of the user to delete (required if authenticated as the user being deleted)",
						},
					),
					Before: ensureNoArgs,
					Action: userDelete,
				},
				{
					Name:  "update-password",
					Usage: "update a user's password",
					Flags: CommandFlags(
						cli.StringFlag{
							Name:  UserIDFlag,
							Usage: "`USERID` of the user to update",
						},
					),
					Before: ensureNoArgs,
					Action: userUpdatePassword,
				},
			},
		},
	}
}

func userGet(c *cli.Context) error {
	var user *user.User
	var err error

	email := c.String(EmailFlag)
	if email != "" {
		if c.String(UserIDFlag) != "" {
			return errors.New("Must specified either EMAIL or USERID, but not both")
		}
		user, err = API(c).GetUserByEmail(email)
	} else {
		user, err = API(c).GetUserByID(c.String(UserIDFlag))
	}

	if err != nil {
		return err
	}

	return reportMessageWithJSON(c, user)
}

func userFind(c *cli.Context) error {
	if !c.IsSet(RoleFlag) {
		return errors.New("No search criteria specified")
	}

	role := c.String(RoleFlag)
	if role == "" {
		return errors.New("Role is missing")
	}

	query := &api.UsersQuery{}
	query.Role = &role
	users, err := API(c).FindUsers(query)
	if err != nil {
		return err
	}

	for _, user := range users {
		if err = reportMessageWithJSON(c, user); err != nil {
			return err
		}
	}

	return nil
}

func userAddRoles(c *cli.Context) error {
	updater, err := api.NewAddRolesUserUpdater(c.StringSlice(RoleFlag))
	if err != nil {
		return err
	}

	updateUser, err := API(c).ApplyUpdatersToUserByID(c.String(UserIDFlag), []api.UserUpdater{updater})
	if err != nil {
		return err
	}

	return reportMessageWithJSON(c, updateUser)
}

func userRemoveRoles(c *cli.Context) error {
	updater, err := api.NewRemoveRolesUserUpdater(c.StringSlice(RoleFlag))
	if err != nil {
		return err
	}

	updateUser, err := API(c).ApplyUpdatersToUserByID(c.String(UserIDFlag), []api.UserUpdater{updater})
	if err != nil {
		return err
	}

	return reportMessageWithJSON(c, updateUser)
}

func userDelete(c *cli.Context) error {
	userID := c.String(UserIDFlag)
	password := c.String(PasswordFlag)
	if password == "" && API(c).IsSessionUserID(userID) {
		var err error
		if password, err = readFromConsoleNoEcho("Password: "); err != nil {
			return err
		}
	}

	err := API(c).DeleteUserByID(userID, password)
	if err != nil {
		return err
	}

	return reportMessage(c, "User deleted.")
}

func userUpdatePassword(c *cli.Context) error {
	newPassword, err := readFromConsoleNoEcho("New Password: ")
	if err != nil {
		return err
	}

	newUserDetails := api.UserUpdates{Password: &newPassword}

	updateUser, err := API(c).UpdateUserByID(c.String(UserIDFlag), &newUserDetails)
	if err != nil {
		return err
	}

	return reportMessageWithJSON(c, updateUser)
}
