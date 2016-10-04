package v1

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"net/http"

	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/profile"
	commonService "github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service"
)

type UsersDeleteParameters struct {
	Password string `json:"password,omitempty"`
}

func UsersDelete(serviceContext service.Context) {
	userID := serviceContext.Request().PathParam("userid")
	if userID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	var password *string
	if !serviceContext.AuthenticationDetails().IsServer() {
		authenticatedUserID := serviceContext.AuthenticationDetails().UserID()

		var permissions client.Permissions
		permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, authenticatedUserID, userID)
		if err != nil {
			if client.IsUnauthorizedError(err) {
				serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			} else {
				serviceContext.RespondWithInternalServerFailure("Unable to get user permissions", err)
			}
			return
		}
		if _, ok := permissions[client.OwnerPermission]; ok {
			var usersDeleteParameters UsersDeleteParameters
			if err = serviceContext.Request().DecodeJsonPayload(&usersDeleteParameters); err != nil {
				serviceContext.RespondWithError(commonService.ErrorJSONMalformed())
				return
			}
			password = &usersDeleteParameters.Password
		} else if _, ok = permissions[client.CustodianPermission]; !ok {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			return
		}
	}

	user, err := serviceContext.UserStoreSession().GetUserByID(userID)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to get user by id", err)
		return
	}
	if user == nil {
		serviceContext.RespondWithError(ErrorUserIDNotFound(userID))
		return
	}

	if password != nil {
		if !serviceContext.UserStoreSession().PasswordMatches(user, *password) {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			return
		}
	}

	messageUser := &messageStore.User{
		ID: userID,
	}

	if user.ProfileID != nil {
		var profile *profile.Profile
		profile, err = serviceContext.ProfileStoreSession().GetProfileByID(*user.ProfileID)
		if err != nil {
			serviceContext.RespondWithInternalServerFailure("Unable to get profile by id", err)
			return
		}
		if profile != nil && profile.FullName != nil {
			messageUser.FullName = *profile.FullName
		}
	}

	if err = serviceContext.MetricServicesClient().RecordMetric(serviceContext, "users_delete", map[string]string{"userId": userID}); err != nil {
		serviceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	if err = serviceContext.UserStoreSession().DeleteUser(user); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete user", err)
		return
	}

	if err = serviceContext.SessionStoreSession().DestroySessionsForUserByID(userID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy sessions for user by id", err)
		return
	}

	if err = serviceContext.PermissionStoreSession().DestroyPermissionsForUserByID(userID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy permissions for user by id", err)
		return
	}

	if err = serviceContext.NotificationStoreSession().DestroyNotificationsForUserByID(userID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy notifications for user by id", err)
		return
	}

	if err = serviceContext.DataServicesClient().DestroyDataForUserByID(serviceContext, userID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy data for user by id", err)
		return
	}

	if err = serviceContext.MessageStoreSession().DestroyMessagesForUserByID(userID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy messages for user by id", err)
		return
	}

	if err = serviceContext.MessageStoreSession().DeleteMessagesFromUser(messageUser); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete messages from user", err)
		return
	}

	if user.ProfileID != nil {
		if err = serviceContext.ProfileStoreSession().DestroyProfileByID(*user.ProfileID); err != nil {
			serviceContext.RespondWithInternalServerFailure("Unable to destroy profile by id", err)
			return
		}
	}

	if err = serviceContext.UserStoreSession().DestroyUserByID(userID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy user by id", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
