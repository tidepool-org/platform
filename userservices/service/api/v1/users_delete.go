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
	targetUserID := serviceContext.Request().PathParam("userid")
	if targetUserID == "" {
		serviceContext.RespondWithError(ErrorUserIDMissing())
		return
	}

	var password *string
	if !serviceContext.AuthenticationDetails().IsServer() {
		authenticatedUserID := serviceContext.AuthenticationDetails().UserID()

		var permissions client.Permissions
		permissions, err := serviceContext.UserServicesClient().GetUserPermissions(serviceContext, authenticatedUserID, targetUserID)
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

	targetUser, err := serviceContext.UserStoreSession().GetUserByID(targetUserID)
	if err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to get user by id", err)
		return
	}
	if targetUser == nil {
		serviceContext.RespondWithError(ErrorUserIDNotFound(targetUserID))
		return
	}

	if password != nil {
		if !serviceContext.UserStoreSession().PasswordMatches(targetUser, *password) {
			serviceContext.RespondWithError(commonService.ErrorUnauthorized())
			return
		}
	}

	messageUser := &messageStore.User{
		ID: targetUserID,
	}

	if targetUser.ProfileID != nil {
		var profile *profile.Profile
		profile, err = serviceContext.ProfileStoreSession().GetProfileByID(*targetUser.ProfileID)
		if err != nil {
			serviceContext.RespondWithInternalServerFailure("Unable to get profile by id", err)
			return
		}
		if profile != nil && profile.FullName != nil {
			messageUser.FullName = *profile.FullName
		}
	}

	if err = serviceContext.MetricServicesClient().RecordMetric(serviceContext, "users_delete", map[string]string{"userId": targetUserID}); err != nil {
		serviceContext.Logger().WithError(err).Error("Unable to record metric")
	}

	if err = serviceContext.UserStoreSession().DeleteUser(targetUser); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete user", err)
		return
	}

	if err = serviceContext.SessionStoreSession().DestroySessionsForUserByID(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy sessions for user by id", err)
		return
	}

	if err = serviceContext.PermissionStoreSession().DestroyPermissionsForUserByID(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy permissions for user by id", err)
		return
	}

	if err = serviceContext.NotificationStoreSession().DestroyNotificationsForUserByID(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy notifications for user by id", err)
		return
	}

	if err = serviceContext.DataServicesClient().DestroyDataForUserByID(serviceContext, targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy data for user by id", err)
		return
	}

	if err = serviceContext.MessageStoreSession().DestroyMessagesForUserByID(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy messages for user by id", err)
		return
	}

	if err = serviceContext.MessageStoreSession().DeleteMessagesFromUser(messageUser); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to delete messages from user", err)
		return
	}

	if targetUser.ProfileID != nil {
		if err = serviceContext.ProfileStoreSession().DestroyProfileByID(*targetUser.ProfileID); err != nil {
			serviceContext.RespondWithInternalServerFailure("Unable to destroy profile by id", err)
			return
		}
	}

	if err = serviceContext.UserStoreSession().DestroyUserByID(targetUserID); err != nil {
		serviceContext.RespondWithInternalServerFailure("Unable to destroy user by id", err)
		return
	}

	serviceContext.RespondWithStatusAndData(http.StatusOK, struct{}{})
}
