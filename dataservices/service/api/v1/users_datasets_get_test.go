package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"net/http"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/dataservices/service/api/v1"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

var _ = Describe("UsersDatasetsGet", func() {
	Context("Unit Tests", func() {
		var authenticatedUserID string
		var targetUserID string
		var uploads []*upload.Upload
		var context *TestContext

		BeforeEach(func() {
			authenticatedUserID = app.NewID()
			targetUserID = app.NewID()
			for i := 0; i < 3; i++ {
				uploads = append(uploads, upload.New())
			}
			context = NewTestContext()
			context.RequestImpl.PathParams["userid"] = targetUserID
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{{uploads, nil}}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{client.ViewPermission: client.Permission{}}, nil}}
			context.SetAuthenticationInfo(&client.AuthenticationInfo{UserID: authenticatedUserID})
		})

		It("succeeds if authenticated as user, not server", func() {
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
		})

		It("succeeds if authenticated as server, not user", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.SetAuthenticationInfo(&client.AuthenticationInfo{IsServer: true})
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
		})

		It("panics if context is missing", func() {
			Expect(func() { v1.UsersDatasetsGet(nil) }).To(Panic())
		})

		It("panics if request is missing", func() {
			context.RequestImpl = nil
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
		})

		It("responds with error if not provided as a parameter", func() {
			delete(context.RequestImpl.PathParams, "userid")
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDMissing()}))
		})

		It("panics if user services client is missing", func() {
			context.UserServicesClientImpl = nil
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
		})

		It("responds with error if user services client get user permissions returns unauthorized error", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
		})

		It("responds with error if user services client get user permissions returns any other error", func() {
			err := errors.New("other")
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
		})

		It("responds with error if user services client get user permissions does not return needed permissions", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{client.UploadPermission: client.Permission{}}, nil}}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
		})

		It("panics if data store session is missing", func() {
			context.DataStoreSessionImpl = nil
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
		})

		It("responds with error if data store session get datasets for user returns an error", func() {
			err := errors.New("other")
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{{nil, err}}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserInputs).To(Equal([]string{targetUserID}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get datasets for user", []interface{}{err}}}))
		})
	})
})
