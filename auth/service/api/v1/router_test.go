package v1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gomock "github.com/golang/mock/gomock"

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
	var profileAccessor *user.MockUserProfileAccessor

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		svc, userAccessor, profileAccessor = serviceTest.NewMockedService(ctrl)
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
				var userProfile *user.UserProfile
				var userDetails *user.User

				JustBeforeEach(func() {
					app, err := rest.MakeRouter(rtr.Routes()...)
					Expect(err).ToNot(HaveOccurred())
					Expect(app).ToNot(BeNil())
					handlerFunc = app.AppFunc()
				})

				BeforeEach(func() {
					userID = userTest.RandomID()
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
						UserID:   pointer.FromString("abcdefghij"),
						Username: pointer.FromString("dev@tidepool.org"),
					}

					profileAccessor.EXPECT().FindUserProfile(gomock.Any(), userID).
						Return(userProfile, nil).AnyTimes()

					userAccessor.EXPECT().FindUserById(gomock.Any(), userID).
						Return(userDetails, nil).AnyTimes()
				})

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
						})

						It("it succeeds if the profile exists", func() {
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

						It("retrieves user's profile if this is a service to service request", func() {
							handlerFunc(res, req)
							Expect(res.WriteHeaderInputs).To(Equal([]int{http.StatusOK}))
						})
					})
				})
			})
		})
	})
})
