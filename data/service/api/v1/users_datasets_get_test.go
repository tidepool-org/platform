package v1_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("UsersDataSetsGet", func() {
	// 	Context("Unit Tests", func() {
	// 		var authUserID string
	// 		var targetUserID string
	// 		var uploads []*upload.Upload
	// 		var context *TestContext
	// 		var filter *store.Filter
	// 		var pagination *page.Pagination

	// 		BeforeEach(func() {
	// 			authUserID = id.New()
	// 			targetUserID = id.New()
	// 			uploads = []*upload.Upload{}
	// 			for i := 0; i < 3; i++ {
	// 				uploads = append(uploads, upload.New())
	// 			}
	// 			context = NewTestContext()
	// 			context.RequestImpl.PathParams["user_id"] = targetUserID
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{permission.Permissions{permission.Read: permission.Permission{}}, nil}}
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{{DataSets: uploads, Error: nil}}
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{false}
	// 			context.AuthDetailsImpl.UserIDOutputs = []string{authUserID}
	// 			filter = store.NewFilter()
	// 			pagination = page.NewPagination()
	// 		})

	// 		It("succeeds if authenticated as user, not server", func() {
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.DataRepositoryImpl.GetDataSetsForUserByIDInputs).To(Equal([]testDataStoreDEPRECATED.GetDataSetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if authenticated as server, not user", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{true}
	// 			context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.DataRepositoryImpl.GetDataSetsForUserByIDInputs).To(Equal([]testDataStoreDEPRECATED.GetDataSetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if deleted query parameter specified", func() {
	// 			filter.Deleted = true
	// 			context.RequestImpl.Request.URL.RawQuery = "deleted=true"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.DataRepositoryImpl.GetDataSetsForUserByIDInputs).To(Equal([]testDataStoreDEPRECATED.GetDataSetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if page query parameter specified", func() {
	// 			pagination.Page = 1
	// 			context.RequestImpl.Request.URL.RawQuery = "page=1"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.DataRepositoryImpl.GetDataSetsForUserByIDInputs).To(Equal([]testDataStoreDEPRECATED.GetDataSetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if size query parameter specified", func() {
	// 			pagination.Size = 10
	// 			context.RequestImpl.Request.URL.RawQuery = "size=10"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.DataRepositoryImpl.GetDataSetsForUserByIDInputs).To(Equal([]testDataStoreDEPRECATED.GetDataSetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if all query parameters specified", func() {
	// 			filter.Deleted = true
	// 			pagination.Page = 3
	// 			pagination.Size = 20
	// 			context.RequestImpl.Request.URL.RawQuery = "size=20&deleted=true&page=3"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.DataRepositoryImpl.GetDataSetsForUserByIDInputs).To(Equal([]testDataStoreDEPRECATED.GetDataSetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, uploads}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if context is missing", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			Expect(func() { v1.UsersDataSetsGet(nil) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if request is missing", func() {
	// 			context.RequestImpl = nil
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			Expect(func() { v1.UsersDataSetsGet(context) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user id not provided as a parameter", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			delete(context.RequestImpl.PathParams, "user_id")
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDMissing()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if user client is missing", func() {
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl = nil
	// 			Expect(func() { v1.UsersDataSetsGet(context) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user client get user permissions returns unauthorized error", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user client get user permissions returns any other error", func() {
	// 			err := errors.New("other")
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user client get user permissions does not return needed permissions", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{permission.Permissions{permission.Write: permission.Permission{}}, nil}}
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if deleted query parameter not a boolean", func() {
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.RequestImpl.Request.URL.RawQuery = "deleted=abc"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorTypeNotBoolean("").WithSourceParameter("deleted")}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if page query parameter not an integer", func() {
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.RequestImpl.Request.URL.RawQuery = "page=abc"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorTypeNotInteger("").WithSourceParameter("page")}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if page query parameter is less than minimum", func() {
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.RequestImpl.Request.URL.RawQuery = "page=-1"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorValueNotGreaterThanOrEqualTo(-1, 0).WithSourceParameter("page")}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if size query parameter not an integer", func() {
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.RequestImpl.Request.URL.RawQuery = "size=abc"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorTypeNotInteger("").WithSourceParameter("size")}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if size query parameter is less than minimum", func() {
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.RequestImpl.Request.URL.RawQuery = "size=0"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorValueNotInRange(0, 1, 100).WithSourceParameter("size")}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if size query parameter is greater than maximum", func() {
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{}
	// 			context.RequestImpl.Request.URL.RawQuery = "size=101"
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.RespondWithStatusAndErrorsInputs).To(Equal([]RespondWithStatusAndErrorsInput{{http.StatusBadRequest, []*service.Error{service.ErrorValueNotInRange(101, 1, 100).WithSourceParameter("size")}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if data store session is missing", func() {
	// 			context.DataRepositoryImpl = nil
	// 			Expect(func() { v1.UsersDataSetsGet(context) }).To(Panic())
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if data store session get data sets for user returns an error", func() {
	// 			err := errors.New("other")
	// 			context.DataRepositoryImpl.GetDataSetsForUserByIDOutputs = []testDataStoreDEPRECATED.GetDataSetsForUserByIDOutput{{DataSets: nil, Error: err}}
	// 			v1.UsersDataSetsGet(context)
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataRepositoryImpl.GetDataSetsForUserByIDInputs).To(Equal([]testDataStoreDEPRECATED.GetDataSetsForUserByIDInput{{UserID: targetUserID, Filter: filter, Pagination: pagination}}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get data sets for user", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})
	// 	})
})
