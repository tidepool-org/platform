package v1

import (
	"net/http"

	"github.com/tidepool-org/platform/client"
	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
	userClient "github.com/tidepool-org/platform/user/client"
	userService "github.com/tidepool-org/platform/user/service"
)

type UsersDeleteParameters struct {
	Password string `json:"password,omitempty"`
}

func UsersDelete(userServiceContext userService.Context) {
	targetUserID := userServiceContext.Request().PathParam("userid")
	if targetUserID == "" {
		userServiceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	var password *string
	if !userServiceContext.AuthDetails().IsServer() {
		authUserID := userServiceContext.AuthDetails().UserID()

		var permissions userClient.Permissions
		permissions, err := userServiceContext.UserClient().GetUserPermissions(userServiceContext, authUserID, targetUserID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				userServiceContext.RespondWithError(service.ErrorUnauthorized())
			} else {
				userServiceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[userClient.OwnerPermission]; ok {
			var usersDeleteParameters UsersDeleteParameters
			if err = userServiceContext.Request().DecodeJsonPayload(&usersDeleteParameters); err != nil {
				userServiceContext.RespondWithError(service.ErrorJSONMalformed())
				return
			}
			password = &usersDeleteParameters.Password
		} else if _, ok = permissions[userClient.CustodianPermission]; !ok {
			userServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	targetUser, err := userServiceContext.UserStoreSession().GetUserByID(targetUserID)
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
		if !userServiceContext.UserStoreSession().PasswordMatches(targetUser, *password) {
			userServiceContext.RespondWithError(service.ErrorUnauthorized())
			return
		}
	}

	messageUser := &messageStore.User{
		ID: targetUserID,
	}

	if targetUser.ProfileID != nil {
		var profile *profile.Profile
		profile, err = userServiceContext.ProfileStoreSession().GetProfileByID(*targetUser.ProfileID)
		if err != nil {
			userServiceContext.RespondWithInternalServerFailure("Unable to get profile by id", err)
			return
		}
		if profile != nil && profile.FullName != nil {
			messageUser.FullName = *profile.FullName
		}
	}

	if err = userServiceContext.MetricClient().RecordMetric(userServiceContext, "users_delete", map[string]string{"userId": targetUserID}); err != nil {
		userServiceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	if err = userServiceContext.UserStoreSession().DeleteUser(targetUser); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to delete user", err)
		return
	}

	if err = userServiceContext.SessionStoreSession().DestroySessionsForUserByID(targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy sessions for user by id", err)
		return
	}

	if err = userServiceContext.PermissionStoreSession().DestroyPermissionsForUserByID(targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy permissions for user by id", err)
		return
	}

	if err = userServiceContext.ConfirmationStoreSession().DestroyConfirmationsForUserByID(targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy confirmations for user by id", err)
		return
	}

	if err = userServiceContext.DataClient().DestroyDataForUserByID(userServiceContext, targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy data for user by id", err)
		return
	}

	if err = userServiceContext.MessageStoreSession().DestroyMessagesForUserByID(targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy messages for user by id", err)
		return
	}

	if err = userServiceContext.MessageStoreSession().DeleteMessagesFromUser(messageUser); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to delete messages from user", err)
		return
	}

	if targetUser.ProfileID != nil {
		if err = userServiceContext.ProfileStoreSession().DestroyProfileByID(*targetUser.ProfileID); err != nil {
			userServiceContext.RespondWithInternalServerFailure("Unable to destroy profile by id", err)
			return
		}
	}

	if err = userServiceContext.UserStoreSession().DestroyUserByID(targetUserID); err != nil {
		userServiceContext.RespondWithInternalServerFailure("Unable to destroy user by id", err)
		return
	}

	userServiceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
