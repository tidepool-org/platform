package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	gomock "go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ant0ine/go-json-rest/rest"

	authServiceApiV1 "github.com/tidepool-org/platform/auth/service/api/v1"
	serviceTest "github.com/tidepool-org/platform/auth/service/test"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	testRest "github.com/tidepool-org/platform/test/rest"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Router", func() {
	var ctrl *gomock.Controller
	var svc *serviceTest.Service
	var userAccessor *user.MockUserAccessor
	var profileAccessor *user.MockProfileAccessor
	var permsClient *permission.MockClient

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		svc, userAccessor, profileAccessor, permsClient = serviceTest.NewMockedService(ctrl)
	})

	Context("NewRouter", func() {
		It("returns an error if context is missing", func() {
			rtr, err := authServiceApiV1.NewRouter(nil)
			errorsTest.ExpectEqual(err, errors.New("service is missing"))
			Expect(rtr).To(BeNil())
		})

		It("returns successfully", func() {
			rtr, err := authServiceApiV1.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var rtr *authServiceApiV1.Router

		BeforeEach(func() {
			var err error
			rtr, err = authServiceApiV1.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(rtr.Routes()).ToNot(BeEmpty())
			})

			Context("Profile", func() {
				var res *testRest.ResponseWriter
				var req *rest.Request
				var ctx context.Context
				var handlerFunc rest.HandlerFunc
				var userID string
				var details request.AuthDetails
				var userProfile *user.Profile
				var userRoles []string
				var userDetails *user.User

				JustBeforeEach(func() {
					app, err := rest.MakeRouter(rtr.Routes()...)
					Expect(err).ToNot(HaveOccurred())
					Expect(app).ToNot(BeNil())
					handlerFunc = app.AppFunc()
				})

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					res = testRest.NewResponseWriter()
					res.HeaderOutput = &http.Header{}
					req = testRest.NewRequest()
					ctx = log.NewContextWithLogger(req.Context(), logTest.NewLogger())
					req.Request = req.WithContext(ctx)

					userProfile = &user.UserProfile{
						FullName:      "Some User Profile",
						Birthday:      "2001-02-03",
						DiagnosisDate: "2002-03-04",
						About:         "About me",
						MRN:           "11223344",
					}
					userDetails = &user.User{
						UserID:   pointer.FromString(userID),
						Username: pointer.FromString("dev@tidepool.org"),
					}
					userRoles = []string{user.RolePatient}

					userAccessor.EXPECT().
						FindUserById(gomock.Any(), userID).
						Return(userDetails, nil).AnyTimes()
				})

				Context("Legacy Profiles", func() {
					Context("GetProfile", func() {
						BeforeEach(func() {
							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/users/legacy/%s/profile", userID)
						})
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})
						AfterEach(func() {
							res.AssertOutputsEmpty()
						})

						Context("as service", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
								permsClient.EXPECT().
									HasMembershipRelationship(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasCustodianPermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasWritePermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
							})

							It("it succeeds if the profile exists", func() {
								profileAccessor.EXPECT().
									FindUserProfile(gomock.Any(), userID).
									Return(userProfile.ToLegacyProfile(userRoles), nil)

								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(userProfile.ToLegacyProfile(userRoles))).To(MatchJSON(res.WriteInputs[0]))
							})

							It("it includes the clinician object if user is a clinic", func() {
								userProfile = &user.UserProfile{
									FullName: "Some Clinician",
								}
								userRoles = []string{user.RoleClinic}
								profileAccessor.EXPECT().
									FindUserProfile(gomock.Any(), userID).
									Return(userProfile.ToLegacyProfile(userRoles), nil)
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(userProfile.ToLegacyProfile(userRoles))).To(MatchJSON(res.WriteInputs[0]))
								Expect(userProfile.ToLegacyProfile(userRoles).Clinic).NotTo(BeNil())
							})
						})

						Context("as user", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
							})

							It("retrieves user's own profile", func() {
								profileAccessor.EXPECT().
									FindUserProfile(gomock.Any(), userID).
									Return(userProfile.ToLegacyProfile(userRoles), nil)
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(userProfile.ToLegacyProfile(userRoles))).To(MatchJSON(res.WriteInputs[0]))
							})

							Context("other persons profile", func() {
								var otherPersonID string
								var otherProfile *user.LegacyUserProfile
								var otherDetails *user.User
								BeforeEach(func() {
									otherPersonID = userTest.RandomUserID()
									req.URL.Path = fmt.Sprintf("/v1/users/legacy/%s/profile", otherPersonID)
									otherProfile = &user.LegacyUserProfile{
										FullName: "Someone Else's Profile",
										Patient: &user.LegacyPatientProfile{
											Birthday:      "2002-03-04",
											DiagnosisDate: "2003-04-05",
											About:         "Not about me",
											MRN:           "11223346",
										},
									}
									otherDetails = &user.User{
										UserID:   pointer.FromString(otherPersonID),
										Username: pointer.FromString("dev+other@tidepool.org"),
									}
								})
								It("retrieves another person's profile if user has access", func() {
									permsClient.EXPECT().
										HasMembershipRelationship(gomock.Any(), userID, otherPersonID).
										Return(true, nil).AnyTimes()
									profileAccessor.EXPECT().
										FindUserProfile(gomock.Any(), otherPersonID).
										Return(otherProfile, nil).AnyTimes()
									userAccessor.EXPECT().
										FindUserById(gomock.Any(), otherPersonID).
										Return(otherDetails, nil).AnyTimes()
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(json.Marshal(otherProfile)).To(MatchJSON(res.WriteInputs[0]))
								})
								It("fails to retrieve another person's profile if user does not have access", func() {
									permsClient.EXPECT().
										HasMembershipRelationship(gomock.Any(), userID, otherPersonID).
										Return(false, nil).AnyTimes()
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									res.WriteOutputs = nil
								})
							})
						})
					})

					Context("UpdateProfile", func() {
						var updatedProfile *user.LegacyUserProfile
						BeforeEach(func() {
							req.Method = http.MethodPost
							req.URL.Path = fmt.Sprintf("/v1/users/legacy/%s/profile", userID)

							updatedProfile = &user.LegacyUserProfile{
								FullName: "Updated User Profile",
								Patient: &user.LegacyPatientProfile{
									Birthday:      "2000-01-02",
									DiagnosisDate: "2001-02-03",
									About:         "Updated info",
									MRN:           "11223345",
									Email:         "",
									Emails:        []string{},
								},
							}
							bites, err := json.Marshal(updatedProfile)

							Expect(err).ToNot(HaveOccurred())
							req.Body = io.NopCloser(bytes.NewReader(bites))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})
						AfterEach(func() {
							res.AssertOutputsEmpty()
						})

						Context("as service", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
								permsClient.EXPECT().
									HasMembershipRelationship(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasCustodianPermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasWritePermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								profileAccessor.EXPECT().
									UpdateUserProfile(gomock.Any(), userID, gomock.Any()).
									Return(nil)
							})

							It("succeeds", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(updatedProfile)).To(MatchJSON(res.WriteInputs[0]))
							})
						})

						Context("as user", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
							})

							It("successfully updates own profile", func() {
								profileAccessor.EXPECT().
									UpdateUserProfile(gomock.Any(), userID, gomock.Any()).
									Return(nil)
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(updatedProfile)).To(MatchJSON(res.WriteInputs[0]))
							})

							It("fails to update another person's profile that the user does not have custodian access to", func() {
								otherPersonID := userTest.RandomUserID()
								req.URL.Path = fmt.Sprintf("/v1/users/legacy/%s/profile", otherPersonID)
								permsClient.EXPECT().
									HasCustodianPermissions(gomock.Any(), userID, gomock.Not(userID)).
									Return(false, nil).AnyTimes()
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								res.WriteOutputs = nil
							})
						})
					})
				})

				Context("Non-legacy Profiles", func() {
					Context("GetProfile", func() {
						BeforeEach(func() {
							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/users/%s/profile", userID)
						})
						BeforeEach(func() {
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})
						AfterEach(func() {
							res.AssertOutputsEmpty()
						})

						Context("as service", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
								permsClient.EXPECT().
									HasMembershipRelationship(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasCustodianPermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasWritePermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
							})

							It("it succeeds if the profile exists", func() {
								profileAccessor.EXPECT().
									FindUserProfile(gomock.Any(), userID).
									Return(userProfile.ToLegacyProfile(userRoles), nil)
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(userProfile)).To(MatchJSON(res.WriteInputs[0]))
							})
						})

						Context("as user", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
							})

							It("retrieves user's own profile", func() {
								profileAccessor.EXPECT().
									FindUserProfile(gomock.Any(), userID).
									Return(userProfile.ToLegacyProfile(userRoles), nil)
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(userProfile)).To(MatchJSON(res.WriteInputs[0]))
							})

							Context("other persons profile", func() {
								var otherPersonID string
								var otherProfile *user.UserProfile
								var otherRoles []string
								var otherDetails *user.User
								BeforeEach(func() {
									otherPersonID = userTest.RandomUserID()
									req.URL.Path = fmt.Sprintf("/v1/users/%s/profile", otherPersonID)
									otherProfile = &user.UserProfile{
										FullName:      "Someone Else's Profile",
										Birthday:      "2002-03-04",
										DiagnosisDate: "2003-04-05",
										About:         "Not about me",
										MRN:           "11223346",
									}
									otherDetails = &user.User{
										UserID:   pointer.FromString(otherPersonID),
										Username: pointer.FromString("dev+other@tidepool.org"),
									}
									otherRoles = []string{user.RolePatient}
								})
								It("retrieves another person's profile if user has access", func() {
									permsClient.EXPECT().
										HasMembershipRelationship(gomock.Any(), userID, otherPersonID).
										Return(true, nil).AnyTimes()
									profileAccessor.EXPECT().
										FindUserProfile(gomock.Any(), otherPersonID).
										Return(otherProfile.ToLegacyProfile(otherRoles), nil).AnyTimes()
									userAccessor.EXPECT().
										FindUserById(gomock.Any(), otherPersonID).
										Return(otherDetails, nil).AnyTimes()
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
									Expect(json.Marshal(otherProfile)).To(MatchJSON(res.WriteInputs[0]))
								})
								It("fails to retrieve another person's profile if user does not have access", func() {
									permsClient.EXPECT().
										HasMembershipRelationship(gomock.Any(), userID, otherPersonID).
										Return(false, nil).AnyTimes()
									handlerFunc(res, req)
									Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
									res.WriteOutputs = nil
								})
							})
						})
					})

					Context("UpdateProfile", func() {
						var updatedProfile *user.UserProfile
						BeforeEach(func() {
							req.Method = http.MethodPost
							req.URL.Path = fmt.Sprintf("/v1/users/%s/profile", userID)

							updatedProfile = &user.UserProfile{
								FullName:      "Updated User Profile",
								Birthday:      "2000-01-02",
								DiagnosisDate: "2001-02-03",
								About:         "Updated info",
								MRN:           "11223345",
							}

							bites, err := json.Marshal(updatedProfile)

							Expect(err).ToNot(HaveOccurred())
							req.Body = io.NopCloser(bytes.NewReader(bites))
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}
						})
						AfterEach(func() {
							res.AssertOutputsEmpty()
						})

						Context("as service", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
								permsClient.EXPECT().
									HasMembershipRelationship(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasCustodianPermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()
								permsClient.EXPECT().
									HasWritePermissions(gomock.Any(), gomock.Any(), gomock.Any()).
									Return(true, nil).AnyTimes()

								profileAccessor.EXPECT().
									UpdateUserProfileV2(gomock.Any(), userID, updatedProfile).
									Return(nil).AnyTimes()
							})

							It("succeeds", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(updatedProfile)).To(MatchJSON(res.WriteInputs[0]))
							})
						})

						Context("as user", func() {
							BeforeEach(func() {
								details = request.NewAuthDetails(request.MethodSessionToken, userID, authTest.NewSessionToken())
								req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
								profileAccessor.EXPECT().
									UpdateUserProfileV2(gomock.Any(), userID, updatedProfile).
									Return(nil).AnyTimes()
							})

							It("successfully updates own profile", func() {
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
								Expect(json.Marshal(updatedProfile)).To(MatchJSON(res.WriteInputs[0]))
							})
							It("fails to update another person's profile that the user does not have custodian access to", func() {
								otherPersonID := userTest.RandomUserID()
								req.URL.Path = fmt.Sprintf("/v1/users/%s/profile", otherPersonID)
								permsClient.EXPECT().
									HasCustodianPermissions(gomock.Any(), userID, gomock.Not(userID)).
									Return(false, nil).AnyTimes()
								handlerFunc(res, req)
								Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusForbidden}))
								res.WriteOutputs = nil
							})
						})
					})
				})
			})

			Context("v1/users/:userId/users", func() {
				var res *testRest.ResponseWriter
				var req *rest.Request
				var ctx context.Context
				var handlerFunc rest.HandlerFunc
				var userID string
				var details request.AuthDetails
				var userProfile user.UserProfile
				var userLimitedProfile user.UserProfile
				var userRoles []string
				var userDetails *user.User
				var userLimitedDetails *user.User
				var shareeUserID string
				var shareeRoles []string
				var shareeProfile user.UserProfile
				var shareeDetails *user.User

				JustBeforeEach(func() {
					app, err := rest.MakeRouter(rtr.Routes()...)
					Expect(err).ToNot(HaveOccurred())
					Expect(app).ToNot(BeNil())
					handlerFunc = app.AppFunc()
				})

				BeforeEach(func() {
					userID = userTest.RandomUserID()
					shareeUserID = userTest.RandomUserID()
					res = testRest.NewResponseWriter()
					res.HeaderOutput = &http.Header{}
					req = testRest.NewRequest()
					ctx = log.NewContextWithLogger(req.Context(), logTest.NewLogger())
					req.Request = req.WithContext(ctx)
					req.Method = http.MethodGet
					req.URL.Path = fmt.Sprintf("/v1/users/%s/users", shareeUserID)
					res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

					userProfile = user.UserProfile{
						FullName:      "Some User Profile",
						Birthday:      "2001-02-03",
						DiagnosisDate: "2002-03-04",
						About:         "About me",
						MRN:           "11223344",
					}
					userLimitedProfile = user.UserProfile{
						FullName: "Some User Profile",
					}
					userRoles = []string{user.RolePatient}
					userDetails = &user.User{
						UserID:             pointer.FromString(userID),
						Username:           pointer.FromString("dev@tidepool.org"),
						EmailVerified:      pointer.FromBool(true),
						Roles:              &userRoles,
						Emails:             []string{"dev@tidepool.org"},
						Profile:            &userProfile,
						TrustorPermissions: &permission.Permission{},
					}
					userLimitedDetails = &user.User{
						UserID:             pointer.FromString(userID),
						Emails:             []string{"dev@tidepool.org"},
						Profile:            &userLimitedProfile,
						TrustorPermissions: &permission.Permission{},
						TrusteePermissions: &permission.Permission{},
					}

					shareeProfile = user.UserProfile{
						FullName:      "Someone Else's Profile",
						Birthday:      "2002-03-04",
						DiagnosisDate: "2003-04-05",
						About:         "Not about me",
						MRN:           "11223346",
					}
					shareeRoles = []string{user.RolePatient}
					shareeDetails = &user.User{
						UserID:             pointer.FromString(shareeUserID),
						Username:           pointer.FromString("sharee@tidepool.org"),
						EmailVerified:      pointer.FromBool(true),
						Roles:              &shareeRoles,
						Emails:             []string{"sharee@tidepool.org"},
						Profile:            &shareeProfile,
						TrustorPermissions: &permission.Permission{},
					}

					var s string
					userAccessor.EXPECT().
						FindUserById(gomock.Any(), gomock.AssignableToTypeOf(s)).
						DoAndReturn(
							func(ctx context.Context, id string) (*user.User, error) {
								switch id {
								case userID:
									return userDetails, nil
								case shareeUserID:
									return shareeDetails, nil
								}
								return nil, user.ErrUserNotFound
							}).AnyTimes()

					profileAccessor.EXPECT().
						FindUserProfile(gomock.Any(), gomock.AssignableToTypeOf(s)).
						DoAndReturn(
							func(ctx context.Context, id string) (*user.LegacyUserProfile, error) {
								switch id {
								case userID:
									return userProfile.ToLegacyProfile(userRoles), nil
								case shareeUserID:
									return shareeProfile.ToLegacyProfile(shareeRoles), nil
								}
								return nil, user.ErrUserProfileNotFound
							}).AnyTimes()

					permsClient.EXPECT().
						HasMembershipRelationship(gomock.Any(), shareeUserID, userID).
						Return(true, nil).AnyTimes()

				})
				AfterEach(func() {
					res.AssertOutputsEmpty()
				})

				Context("with full trust permissions", func() {
					BeforeEach(func() {
						permsClient.EXPECT().
							GroupsForUser(gomock.Any(), userID).
							Return(permission.Permissions{
								userID: permission.Permission{
									permission.Owner: map[string]any{},
								},
							}, nil).AnyTimes()
						permsClient.EXPECT().
							GroupsForUser(gomock.Any(), shareeUserID).
							Return(permission.Permissions{
								shareeUserID: permission.Permission{
									permission.Owner: map[string]any{},
								},
								userID: permission.Permission{
									permission.Read: map[string]any{},
								},
							}, nil).AnyTimes()

						permsClient.EXPECT().
							UsersInGroup(gomock.Any(), userID).
							Return(permission.Permissions{
								userID: permission.Permission{
									permission.Owner: map[string]any{},
								},
							}, nil).AnyTimes()
						permsClient.EXPECT().
							UsersInGroup(gomock.Any(), shareeUserID).
							Return(permission.Permissions{
								shareeUserID: permission.Permission{
									permission.Owner: map[string]any{},
								},
								userID: permission.Permission{
									permission.Read: map[string]any{},
								},
							}, nil).AnyTimes()
					})
					Context("as service", func() {
						BeforeEach(func() {
							details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
							req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
							permsClient.EXPECT().
								HasMembershipRelationship(gomock.Any(), gomock.Any(), gomock.Any()).
								Return(true, nil).AnyTimes()
						})
						It("returns sharer's user info w/ sharee.", func() {
							userResults := user.UserArray{
								userDetails,
							}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
						It("excludes self and returns empty if nothing shared.", func() {
							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/users/%s/users", userID)
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

							userResults := user.UserArray{}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
					})

					Context("as user", func() {
						BeforeEach(func() {
							details = request.NewAuthDetails(request.MethodSessionToken, shareeUserID, authTest.NewSessionToken())
							req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
							var s string
							permsClient.EXPECT().
								HasMembershipRelationship(gomock.Any(), gomock.AssignableToTypeOf(s), gomock.AssignableToTypeOf(s)).
								DoAndReturn(
									func(ctx context.Context, granteeID, grantorID string) (bool, error) {
										return granteeID == grantorID || (grantorID == userID && granteeID == shareeUserID), nil
									}).AnyTimes()
						})
						It("returns sharer's user info w/ sharee.", func() {
							userResults := user.UserArray{
								userDetails,
							}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
						It("excludes self and returns empty if nothing shared.", func() {
							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/users/%s/users", userID)
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

							userResults := user.UserArray{}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
					})
				})

				Context("with limited trust permissions", func() {
					BeforeEach(func() {
						permsClient.EXPECT().
							GroupsForUser(gomock.Any(), userID).
							Return(permission.Permissions{
								userID: permission.Permission{
									permission.Owner: map[string]any{},
								},
							}, nil).AnyTimes()
						permsClient.EXPECT().
							GroupsForUser(gomock.Any(), shareeUserID).
							Return(permission.Permissions{
								shareeUserID: permission.Permission{
									permission.Owner: map[string]any{},
								},
								userID: permission.Permission{},
							}, nil).AnyTimes()

						permsClient.EXPECT().
							UsersInGroup(gomock.Any(), userID).
							Return(permission.Permissions{
								userID: permission.Permission{
									permission.Owner: map[string]any{},
								},
							}, nil).AnyTimes()
						permsClient.EXPECT().
							UsersInGroup(gomock.Any(), shareeUserID).
							Return(permission.Permissions{
								shareeUserID: permission.Permission{
									permission.Owner: map[string]any{},
								},
								userID: permission.Permission{},
							}, nil).AnyTimes()
					})
					Context("as service", func() {
						BeforeEach(func() {
							details = request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
							req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
							permsClient.EXPECT().
								HasMembershipRelationship(gomock.Any(), gomock.Any(), gomock.Any()).
								Return(true, nil).AnyTimes()
						})
						It("returns full sharer details if service", func() {
							userResults := user.UserArray{
								userDetails,
							}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
						It("excludes self and returns empty if nothing shared.", func() {
							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/users/%s/users", userID)
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

							userResults := user.UserArray{}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
					})

					Context("as user", func() {
						BeforeEach(func() {
							details = request.NewAuthDetails(request.MethodSessionToken, shareeUserID, authTest.NewSessionToken())
							req.Request = req.WithContext(request.NewContextWithAuthDetails(req.Context(), details))
							var s string
							permsClient.EXPECT().
								HasMembershipRelationship(gomock.Any(), gomock.AssignableToTypeOf(s), gomock.AssignableToTypeOf(s)).
								DoAndReturn(
									func(ctx context.Context, granteeID, grantorID string) (bool, error) {
										return granteeID == grantorID || (grantorID == userID && granteeID == shareeUserID), nil
									}).AnyTimes()
						})
						It("returns sharer's limited user info w/ sharee.", func() {
							userResults := user.UserArray{
								userLimitedDetails,
							}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
						It("excludes self and returns empty if nothing shared.", func() {
							req.Method = http.MethodGet
							req.URL.Path = fmt.Sprintf("/v1/users/%s/users", userID)
							res.WriteOutputs = []testRest.WriteOutput{{BytesWritten: 0, Error: nil}}

							userResults := user.UserArray{}
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
							Expect(json.Marshal(userResults)).To(MatchJSON(res.WriteInputs[0]))
						})
					})
				})
			})
		})
	})
})
