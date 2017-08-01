package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tidepool-org/platform/id"
	messageStore "github.com/tidepool-org/platform/message/store"
	"github.com/tidepool-org/platform/profile"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/userservices/client"
	"github.com/tidepool-org/platform/userservices/service/api/v1"
)

var _ = Describe("UsersDelete", func() {
	Context("Unit Tests", func() {
		var authenticatedUserID string
		var targetUserID string
		var targetProfileID string
		var targetPassword string
		var targetFullName string
		var targetUser *user.User
		var targetProfile *profile.Profile
		var context *TestContext

		BeforeEach(func() {
			authenticatedUserID = id.New()
			targetUserID = id.New()
			targetProfileID = id.New()
			targetPassword = id.New()
			targetFullName = id.New()
			targetUser = &user.User{
				ProfileID: &targetProfileID,
			}
			targetProfile = &profile.Profile{
				FullName: &targetFullName,
			}
			context = NewTestContext()
		})

		WithDestroyingUser := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.UserStoreSessionImpl.DestroyUserByIDInputs).To(Equal([]string{targetUserID}))
				})

				Context("with destroying user", func() {
					BeforeEach(func() {
						context.UserStoreSessionImpl.DestroyUserByIDOutputs = []error{nil}
					})

					It("is successful", func() {
						v1.UsersDelete(context)
						Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
					})
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-destroying-user")
					context.UserStoreSessionImpl.DestroyUserByIDOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to destroy user by id", []interface{}{err}}}))
				})
			}
		}

		WithDestroyingProfile := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.ProfileStoreSessionImpl.DestroyProfileByIDInputs).To(Equal([]string{targetProfileID}))
				})

				Context("with destroying profile", func() {
					BeforeEach(func() {
						context.ProfileStoreSessionImpl.DestroyProfileByIDOutputs = []error{nil}
					})

					WithDestroyingUser(flags)()
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-destroying-profile")
					context.ProfileStoreSessionImpl.DestroyProfileByIDOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to destroy profile by id", []interface{}{err}}}))
				})
			}
		}

		WithDeletingMessagesFromUser := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					messageUser := &messageStore.User{
						ID: targetUserID,
					}
					if flags.IsSet("with-full-name") {
						messageUser.FullName = targetFullName
					}
					Expect(context.MessageStoreSessionImpl.DeleteMessagesFromUserInputs).To(Equal([]*messageStore.User{messageUser}))
				})

				Context("with deleting messages from user", func() {
					BeforeEach(func() {
						context.MessageStoreSessionImpl.DeleteMessagesFromUserOutputs = []error{nil}
					})

					if flags.IsSet("with-profile-id") {
						WithDestroyingProfile(flags)()
					} else {
						WithDestroyingUser(flags)()
					}
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-deleting-messages-from-user")
					context.MessageStoreSessionImpl.DeleteMessagesFromUserOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete messages from user", []interface{}{err}}}))
				})
			}
		}

		WithDestroyingMessages := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.MessageStoreSessionImpl.DestroyMessagesForUserByIDInputs).To(Equal([]string{targetUserID}))
				})

				Context("with destroying messages", func() {
					BeforeEach(func() {
						context.MessageStoreSessionImpl.DestroyMessagesForUserByIDOutputs = []error{nil}
					})

					WithDeletingMessagesFromUser(flags)()
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-destroying-messages")
					context.MessageStoreSessionImpl.DestroyMessagesForUserByIDOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to destroy messages for user by id", []interface{}{err}}}))
				})
			}
		}

		WithDestroyingData := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.DataServicesClientImpl.DestroyDataForUserByIDInputs).To(Equal([]DestroyDataForUserByIDInput{{context, targetUserID}}))
				})

				Context("with destroying data", func() {
					BeforeEach(func() {
						context.DataServicesClientImpl.DestroyDataForUserByIDOutputs = []error{nil}
					})

					WithDestroyingMessages(flags)()
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-destroying-data")
					context.DataServicesClientImpl.DestroyDataForUserByIDOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to destroy data for user by id", []interface{}{err}}}))
				})
			}
		}

		WithDestroyingNotifications := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.NotificationStoreSessionImpl.DestroyNotificationsForUserByIDInputs).To(Equal([]string{targetUserID}))
				})

				Context("with destroying notifications", func() {
					BeforeEach(func() {
						context.NotificationStoreSessionImpl.DestroyNotificationsForUserByIDOutputs = []error{nil}
					})

					WithDestroyingData(flags)()
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-destroying-notifications")
					context.NotificationStoreSessionImpl.DestroyNotificationsForUserByIDOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to destroy notifications for user by id", []interface{}{err}}}))
				})
			}
		}

		WithDestroyingPermissions := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.PermissionStoreSessionImpl.DestroyPermissionsForUserByIDInputs).To(Equal([]string{targetUserID}))
				})

				Context("with destroying permissions", func() {
					BeforeEach(func() {
						context.PermissionStoreSessionImpl.DestroyPermissionsForUserByIDOutputs = []error{nil}
					})

					WithDestroyingNotifications(flags)()
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-destroying-permissions")
					context.PermissionStoreSessionImpl.DestroyPermissionsForUserByIDOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to destroy permissions for user by id", []interface{}{err}}}))
				})
			}
		}

		WithDestroyingSessions := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.SessionStoreSessionImpl.DestroySessionsForUserByIDInputs).To(Equal([]string{targetUserID}))
				})

				Context("with destroying sessions", func() {
					BeforeEach(func() {
						context.SessionStoreSessionImpl.DestroySessionsForUserByIDOutputs = []error{nil}
					})

					WithDestroyingPermissions(flags)()
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-destroying-sessions")
					context.SessionStoreSessionImpl.DestroySessionsForUserByIDOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to destroy sessions for user by id", []interface{}{err}}}))
				})
			}
		}

		WithDeletingUser := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.UserStoreSessionImpl.DeleteUserInputs).To(Equal([]*user.User{targetUser}))
				})

				Context("with deleting user", func() {
					BeforeEach(func() {
						context.UserStoreSessionImpl.DeleteUserOutputs = []error{nil}
					})

					WithDestroyingSessions(flags)()
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-deleting-user")
					context.UserStoreSessionImpl.DeleteUserOutputs = []error{err}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete user", []interface{}{err}}}))
				})
			}
		}

		WithRecordingMetric := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.MetricServicesClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "users_delete", []map[string]string{{"userId": targetUserID}}}}))
				})

				Context("with recording metric", func() {
					BeforeEach(func() {
						context.MetricServicesClientImpl.RecordMetricOutputs = []error{nil}
					})

					WithDeletingUser(flags)()
				})

				Context("with recording metric returning an error", func() {
					BeforeEach(func() {
						context.MetricServicesClientImpl.RecordMetricOutputs = []error{errors.New("test-error-recording-metric")}
					})

					WithDeletingUser(flags)()
				})
			}
		}

		WithGetProfile := func(flags *TestFlags) func() {
			return func() {
				Context("with profile id", func() {
					AfterEach(func() {
						Expect(context.ProfileStoreSessionImpl.GetProfileByIDInputs).To(Equal([]string{targetProfileID}))
					})

					Context("with existing profile", func() {
						BeforeEach(func() {
							context.ProfileStoreSessionImpl.GetProfileByIDOutputs = []GetProfileByIDOutput{{targetProfile, nil}}
						})

						WithRecordingMetric(flags.Set("with-profile-id", "with-full-name"))()
					})

					Context("with existing profile without full name", func() {
						BeforeEach(func() {
							targetProfile.FullName = nil
							context.ProfileStoreSessionImpl.GetProfileByIDOutputs = []GetProfileByIDOutput{{targetProfile, nil}}
						})

						WithRecordingMetric(flags.Set("with-profile-id"))()
					})

					Context("with no existing profile", func() {
						BeforeEach(func() {
							context.ProfileStoreSessionImpl.GetProfileByIDOutputs = []GetProfileByIDOutput{{nil, nil}}
						})

						WithRecordingMetric(flags.Set("with-profile-id"))()
					})

					It("responds with failure if it returns error", func() {
						err := errors.New("test-error-getting-profile")
						context.ProfileStoreSessionImpl.GetProfileByIDOutputs = []GetProfileByIDOutput{{nil, err}}
						v1.UsersDelete(context)
						Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get profile by id", []interface{}{err}}}))
					})
				})

				Context("with no profile id", func() {
					BeforeEach(func() {
						targetUser.ProfileID = nil
					})

					WithRecordingMetric(flags)()
				})
			}
		}

		WithMatchingPassword := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.UserStoreSessionImpl.PasswordMatchesInputs).To(Equal([]PasswordMatchesInput{{targetUser, targetPassword}}))
				})

				Context("with matching password", func() {
					BeforeEach(func() {
						context.UserStoreSessionImpl.PasswordMatchesOutputs = []bool{true}
					})

					WithGetProfile(flags)()
				})

				It("responds with failure if it returns false", func() {
					context.UserStoreSessionImpl.PasswordMatchesOutputs = []bool{false}
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
				})
			}
		}

		WithoutClinicRole := func(flags *TestFlags) func() {
			return func() {
				Context("without clinic role", func() {
					if flags.IsSet("with-password") {
						WithMatchingPassword(flags)()
					} else {
						WithGetProfile(flags)()
					}
				})

				It("responds with failure if user has clinic role", func() {
					targetUser.Roles = []string{user.ClinicRole}
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
				})
			}
		}

		WithUserID := func(flags *TestFlags) func() {
			return func() {
				AfterEach(func() {
					Expect(context.UserStoreSessionImpl.GetUserByIDInputs).To(Equal([]string{targetUserID}))
				})

				Context("with existing user", func() {
					BeforeEach(func() {
						context.UserStoreSessionImpl.GetUserByIDOutputs = []GetUserByIDOutput{{targetUser, nil}}
					})

					WithoutClinicRole(flags)()
				})

				It("responds with failure if it returns no user", func() {
					context.UserStoreSessionImpl.GetUserByIDOutputs = []GetUserByIDOutput{{nil, nil}}
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDNotFound(targetUserID)}))
				})

				It("responds with failure if it returns error", func() {
					err := errors.New("test-error-getting-user")
					context.UserStoreSessionImpl.GetUserByIDOutputs = []GetUserByIDOutput{{nil, err}}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user by id", []interface{}{err}}}))
				})
			}
		}

		WithPassword := func(flags *TestFlags) func() {
			return func() {
				Context("with password", func() {
					BeforeEach(func() {
						context.RequestImpl.Body = ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"password": "%s"}`, targetPassword)))
					})

					WithUserID(flags.Set("with-password"))()
				})

				It("responds with failure if the request body is not parsable", func() {
					context.RequestImpl.Body = ioutil.NopCloser(strings.NewReader("{"))
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorJSONMalformed()}))
				})
			}
		}

		WithUserPermissions := func(flags *TestFlags) func() {
			return func() {
				BeforeEach(func() {
					context.AuthenticationDetailsImpl.UserIDOutputs = []string{authenticatedUserID}
				})

				AfterEach(func() {
					Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
				})

				Context("with owner permissions", func() {
					BeforeEach(func() {
						context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"root": client.Permission{}}, nil}}
					})

					WithPassword(flags)()
				})

				Context("with custodian permissions", func() {
					BeforeEach(func() {
						context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"custodian": client.Permission{}}, nil}}
					})

					WithUserID(flags)()
				})

				It("responds with failure if it returns other permissions", func() {
					context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"other": client.Permission{}}, nil}}
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
				})

				It("responds with failure if it returns empty permissions", func() {
					context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{}, nil}}
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
				})

				It("responds with failure if it returns no permissions", func() {
					context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, nil}}
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
				})

				It("responds with failure if it returns unauthorized error", func() {
					context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
				})

				It("responds with failure if it returns other error", func() {
					err := errors.New("test-error-getting-user-permissions")
					context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
					v1.UsersDelete(context)
					Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
				})
			}
		}

		AsServer := func(flags *TestFlags) func() {
			return func() {
				Context("as server", func() {
					BeforeEach(func() {
						context.AuthenticationDetailsImpl.IsServerOutputs = []bool{true}
					})

					WithUserID(flags)()
				})
			}
		}

		AsUser := func(flags *TestFlags) func() {
			return func() {
				Context("as user", func() {
					BeforeEach(func() {
						context.AuthenticationDetailsImpl.IsServerOutputs = []bool{false}
					})

					WithUserPermissions(flags)()
				})
			}
		}

		WithRequestParameter := func(flags *TestFlags) func() {
			return func() {
				Context("with request parameter", func() {
					BeforeEach(func() {
						context.RequestImpl.PathParams["userid"] = targetUserID
					})

					AsServer(flags)()
					AsUser(flags)()
				})

				It("responds with failure if the request parameter is missing", func() {
					v1.UsersDelete(context)
					Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDMissing()}))
				})
			}
		}

		Context("with valid test data", func() {
			AfterEach(func() {
				Expect(context.ValidateTest()).To(BeTrue())
			})

			WithRequestParameter(NewTestFlags())()
		})
	})
})
