package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/request"

	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
	userService "github.com/tidepool-org/platform/user/service"
)

type UsersDeleteParameters struct {
	Password string `json:"password,omitempty"`
}

func UsersDelete(userServiceContext userService.Context) {
	ctx := userServiceContext.Request().Context()
	lgr := log.LoggerFromContext(ctx)

	targetUserID := userServiceContext.Request().PathParam("userId")
	if targetUserID == "" {
		userServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	var password *string
	if details := request.DetailsFromContext(ctx); !details.IsService() {
		authUserID := details.UserID()

		var permissions permission.Permissions
		permissions, err := userServiceContext.PermissionClient().GetUserPermissions(ctx, authUserID, targetUserID)
		if err != nil {
			if request.IsErrorUnauthorized(err) {
				userServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				userServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[permission.Owner]; ok {
			var usersDeleteParameters UsersDeleteParameters
			if err = userServiceContext.Request().DecodeJsonPayload(&usersDeleteParameters); err != nil {
				userServiceContext.RespondWithError(service.ErrorJSONMalformed())
				return
			}
			password = &usersDeleteParameters.Password
		} else if _, ok = permissions[permission.Custodian]; !ok {
			userServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	targetUser, err := userServiceContext.UsersSession().GetUserByID(ctx, targetUserID)
	if err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to get user by id", err)
		return
	}
	if targetUser == nil {
		userServiceContext.RespondWithError(ErrorUserIDNotFound(targetUserID))
		return
	}

	if targetUser.HasRole(user.ClinicRole) {
		userServiceContext.RespondWithError(service.ErrorUnauthorized())
		return
	}

	if password != nil {
		if !userServiceContext.UsersSession().PasswordMatches(targetUser, *password) {
			userServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	messageUser := &messageStore.User{
		ID: targetUserID,
	}

	if targetUser.ProfileID != nil {
		var profile *profile.Profile
		profile, err = userServiceContext.ProfilesSession().GetProfileByID(ctx, *targetUser.ProfileID)
		if err != nil {
			userServiceContext.RespondWithInternalServerFailure("Unable to get profile by id", err)
			return
		}
		if profile != nil && profile.FullName != nil {
			messageUser.FullName = *profile.FullName
		}
	}

	if err = userServiceContext.MetricClient().RecordMetric(ctx, "users_delete", map[string]string{"userId": targetUserID}); err != nil {
		lgr.WithError(err).Error("Unable to record metric")
	}

	if err = userServiceContext.UsersSession().DeleteUser(ctx, targetUser); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to delete user", err)
		return
	}

	if err = userServiceContext.SessionsSession().DestroySessionsForUserByID(ctx, targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy sessions for user by id", err)
		return
	}

	if err = userServiceContext.PermissionsSession().DestroyPermissionsForUserByID(ctx, targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy permissions for user by id", err)
		return
	}

	if err = userServiceContext.ConfirmationSession().DeleteUserConfirmations(ctx, targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy confirmations for user by id", err)
		return
	}

	if err = userServiceContext.DataClient().DestroyDataForUserByID(ctx, targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy data for user by id", err)
		return
	}

	if err = userServiceContext.MessagesSession().DestroyMessagesForUserByID(ctx, targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy messages for user by id", err)
		return
	}

	if err = userServiceContext.MessagesSession().DeleteMessagesFromUser(ctx, messageUser); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to delete messages from user", err)
		return
	}

	if targetUser.ProfileID != nil {
		if err = userServiceContext.ProfilesSession().DestroyProfileByID(ctx, *targetUser.ProfileID); err != nil {
			userServiceContext.RespondWithInternalServerFailure("Unable to destroy profile by id", err)
			return
		}
	}

	if err = userServiceContext.UsersSession().DestroyUserByID(ctx, targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy user by id", err)
		return
	}

	userServiceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
