package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"net/http"

	"github.com/tidepool-org/platform/client"
	"github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/data/store"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
)

var _ = Describe("UsersDatasetsGet", func() {
	Context("Unit Tests", func() {
		var authUserID string
		var targetUserID string
		var uploads []*upload.Upload
		var context *TestContext
		var filter *store.Filter
		var pagination *store.Pagination

		BeforeEach(func() {
			authUserID = id.New()
			targetUserID = id.New()
			uploads = []*upload.Upload{}
			for i := 0; i < 3; i++ {
				uploads = append(uploads, upload.Init())
			}
			context = NewTestContext()
			context.RequestImpl.PathParams["userid"] = targetUserID
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{userClient.Permissions{userClient.ViewPermission: userClient.Permission{}}, nil}}
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{{Datasets: uploads, Error: nil}}
			context.AuthDetailsImpl.IsServerOutputs = []bool{false}
			context.AuthDetailsImpl.UserIDOutputs = []string{authUserID}
			filter = store.NewFilter()
			pagination = store.NewPagination()
		})

		It("succeeds if authenticated as user, not server", func() {
			v1.UsersDatasetsGet(context)
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserByIDInputs).To(Equal([]testDataStore.GetDatasetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as server, not user", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.AuthDetailsImpl.IsServerOutputs = []bool{true}
			context.AuthDetailsImpl.UserIDOutputs = []string{}
			v1.UsersDatasetsGet(context)
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserByIDInputs).To(Equal([]testDataStore.GetDatasetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if deleted query parameter specified", func() {
			filter.Deleted = true
			context.RequestImpl.Request.URL.RawQuery = "deleted=true"
			v1.UsersDatasetsGet(context)
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserByIDInputs).To(Equal([]testDataStore.GetDatasetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if page query parameter specified", func() {
			pagination.Page = 1
			context.RequestImpl.Request.URL.RawQuery = "page=1"
			v1.UsersDatasetsGet(context)
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserByIDInputs).To(Equal([]testDataStore.GetDatasetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if size query parameter specified", func() {
			pagination.Size = 10
			context.RequestImpl.Request.URL.RawQuery = "size=10"
			v1.UsersDatasetsGet(context)
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserByIDInputs).To(Equal([]testDataStore.GetDatasetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if all query parameters specified", func() {
			filter.Deleted = true
			pagination.Page = 3
			pagination.Size = 20
			context.RequestImpl.Request.URL.RawQuery = "size=20&deleted=true&page=3"
			v1.UsersDatasetsGet(context)
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserByIDInputs).To(Equal([]testDataStore.GetDatasetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if context is missing", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.AuthDetailsImpl.UserIDOutputs = []string{}
			Expect(func() { v1.UsersDatasetsGet(nil) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if request is missing", func() {
			context.RequestImpl = nil
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.AuthDetailsImpl.UserIDOutputs = []string{}
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user id not provided as a parameter", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.AuthDetailsImpl.UserIDOutputs = []string{}
			delete(context.RequestImpl.PathParams, "userid")
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDMissing()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if user client is missing", func() {
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl = nil
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user client get user permissions returns unauthorized error", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			v1.UsersDatasetsGet(context)
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user client get user permissions returns any other error", func() {
			err := errors.New("other")
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			v1.UsersDatasetsGet(context)
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user client get user permissions does not return needed permissions", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{userClient.Permissions{userClient.UploadPermission: userClient.Permission{}}, nil}}
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			v1.UsersDatasetsGet(context)
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if deleted query parameter not a boolean", func() {
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.RequestImpl.Request.URL.RawQuery = "deleted=abc"
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorTypeNotBoolean("").WithSourceParameter("deleted")}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if page query parameter not an integer", func() {
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.RequestImpl.Request.URL.RawQuery = "page=abc"
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorTypeNotInteger("").WithSourceParameter("page")}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if page query parameter is less than minimum", func() {
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.RequestImpl.Request.URL.RawQuery = "page=-1"
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorValueNotGreaterThanOrEqualTo(-1, 0).WithSourceParameter("page")}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if size query parameter not an integer", func() {
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.RequestImpl.Request.URL.RawQuery = "size=abc"
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorTypeNotInteger("").WithSourceParameter("size")}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if size query parameter is less than minimum", func() {
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.RequestImpl.Request.URL.RawQuery = "size=0"
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorValueNotInRange(0, 1, 100).WithSourceParameter("size")}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if size query parameter is greater than maximum", func() {
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{}
			context.RequestImpl.Request.URL.RawQuery = "size=101"
			v1.UsersDatasetsGet(context)
			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorValueNotInRange(101, 1, 100).WithSourceParameter("size")}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if data store session is missing", func() {
			context.DataStoreSessionImpl = nil
			Expect(func() { v1.UsersDatasetsGet(context) }).To(Panic())
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data store session get datasets for user returns an error", func() {
			err := errors.New("other")
			context.DataStoreSessionImpl.GetDatasetsForUserByIDOutputs = []testDataStore.GetDatasetsForUserByIDOutput{{Datasets: nil, Error: err}}
			v1.UsersDatasetsGet(context)
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.GetDatasetsForUserByIDInputs).To(Equal([]testDataStore.GetDatasetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get datasets for user", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})
	})
})
