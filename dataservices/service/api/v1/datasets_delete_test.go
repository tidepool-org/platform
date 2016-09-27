package v1_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data/types/base/upload"
	"github.com/tidepool-org/platform/dataservices/service/api/v1"
	"github.com/tidepool-org/platform/service"
	"github.com/tidepool-org/platform/userservices/client"
)

var _ = Describe("DatasetsDelete", func() {
	Context("Unit Tests", func() {
		var authenticatedUserID string
		var targetUserID string
		var targetUpload *upload.Upload
		var context *TestContext

		BeforeEach(func() {
			authenticatedUserID = app.NewID()
			targetUserID = app.NewID()
			targetUpload = upload.Init()
			targetUpload.UserID = targetUserID
			targetUpload.ByUser = app.NewID()
			context = NewTestContext()
			context.RequestImpl.PathParams["datasetid"] = targetUpload.UploadID
			context.DataStoreSessionImpl.GetDatasetByIDOutputs = []GetDatasetByIDOutput{{targetUpload, nil}}
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{false}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{authenticatedUserID}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"root": client.Permission{}}, nil}}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{nil}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{nil}
		})

		It("succeeds if authenticated as owner", func() {
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricServicesClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as custodian", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"custodian": client.Permission{}}, nil}}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricServicesClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as uploader and was the uploading user", func() {
			targetUpload.ByUser = authenticatedUserID
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"upload": client.Permission{}}, nil}}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricServicesClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as server", func() {
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{true}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.DataStoreSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricServicesClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if context is missing", func() {
			context.DataStoreSessionImpl.GetDatasetByIDOutputs = []GetDatasetByIDOutput{}
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(nil) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if request is missing", func() {
			context.RequestImpl = nil
			context.DataStoreSessionImpl.GetDatasetByIDOutputs = []GetDatasetByIDOutput{}
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if dataset id not provided as a parameter", func() {
			delete(context.RequestImpl.PathParams, "datasetid")
			context.DataStoreSessionImpl.GetDatasetByIDOutputs = []GetDatasetByIDOutput{}
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorDatasetIDMissing()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if data store session is missing", func() {
			context.DataStoreSessionImpl = nil
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data store session get dataset returns error", func() {
			context.DataStoreSessionImpl.GetDatasetByIDOutputs = []GetDatasetByIDOutput{{nil, errors.New("other")}}
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorDatasetIDNotFound(targetUpload.UploadID)}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user id is missing on dataset", func() {
			targetUpload.UserID = ""
			context.AuthenticationDetailsImpl.IsServerOutputs = []bool{}
			context.AuthenticationDetailsImpl.UserIDOutputs = []string{}
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user id from dataset", nil}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if authentication details is missing", func() {
			context.AuthenticationDetailsImpl = nil
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if user services client is missing", func() {
			context.UserServicesClientImpl = nil
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user services client get user permissions returns unauthorized error", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user services client get user permissions returns any other error", func() {
			err := errors.New("other")
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user services client get user permissions does not return needed permissions", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"view": client.Permission{}}, nil}}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user services client get user permissions returns upload permissions, but not user who uploaded", func() {
			context.UserServicesClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{client.Permissions{"upload": client.Permission{}}, nil}}
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data store session delete dataset returns an error", func() {
			err := errors.New("other")
			context.DataStoreSessionImpl.DeleteDatasetOutputs = []error{err}
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete dataset", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if metric services client is missing", func() {
			context.MetricServicesClientImpl = nil
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserServicesClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authenticatedUserID, targetUserID}}))
			Expect(context.DataStoreSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("logs and ignores if metric services record metric returns an error", func() {
			context.MetricServicesClientImpl.RecordMetricOutputs = []error{errors.New("other")}
			v1.DatasetsDelete(context)
			Expect(context.DataStoreSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.DataStoreSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricServicesClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})
	})
})
