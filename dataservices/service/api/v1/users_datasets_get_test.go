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
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{client.ViewPermission: client.Permission{}}, nil}}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{{uploads, nil}}
			context.IsAuthenticatedServerOutputs = []bool{false}
			context.AuthenticatedUserIDOutputs = []string{authenticatedUserID}
		})

		It("succeeds if authenticated as user, not server", func() {
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as server, not user", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.IsAuthenticatedServerOutputs = []bool{true}
			context.AuthenticatedUserIDOutputs = []string{}
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if context is missing", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{}
			context.IsAuthenticatedServerOutputs = []bool{}
			context.AuthenticatedUserIDOutputs = []string{}
			Expect(func() { v1.UsersDatasetsGet(nil) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if request is missing", func() {
			context.RequestImpl = nil
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{}
			context.IsAuthenticatedServerOutputs = []bool{}
			context.AuthenticatedUserIDOutputs = []string{}
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if not provided as a parameter", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{}
			context.IsAuthenticatedServerOutputs = []bool{}
			context.AuthenticatedUserIDOutputs = []string{}
			delete(context.RequestImpl.PathParams, "userid")
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDMissing()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if user services client is missing", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{}
			context.IsAuthenticatedServerOutputs = []bool{}
			context.AuthenticatedUserIDOutputs = []string{}
			context.UserServicesClientImpl = nil
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user services client get user permissions returns unauthorized error", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user services client get user permissions returns any other error", func() {
			err := errors.New("other")
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user services client get user permissions does not return needed permissions", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{client.UploadPermission: client.Permission{}}, nil}}
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if data store session is missing", func() {
			context.DataStoreSessionImpl = nil
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data store session get datasets for user returns an error", func() {
			err := errors.New("other")
			context.DataStoreSessionImpl.GetDatasetsForUserOutputs = []GetDatasetsForUserOutput{{nil, err}}
			v1.UsersDatasetsGet(context)
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserInputs).To(Equal([]string{targetUserID}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get datasets for user", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})
	})
})
