package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"net/http"

	"github.com/tidepool-org/platform/client"
	testDataDeduplicator "github.com/tidepool-org/platform/data/deduplicator/test"
	"github.com/tidepool-org/platform/data/service/api/v1"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service"
	userClient "github.com/tidepool-org/platform/user/client"
)

var _ = Describe("DatasetsDelete", func() {
	Context("Unit Tests", func() {
		var authUserID string
		var targetUserID string
		var targetUpload *upload.Upload
		var context *TestContext

		BeforeEach(func() {
			authUserID = id.New()
			targetUserID = id.New()
			targetUpload = upload.Init()
			targetUpload.UserID = targetUserID
			targetUpload.ByUser = id.New()
			context = NewTestContext()
			context.Context.RequestImpl.PathParams["datasetid"] = targetUpload.UploadID
			context.DataSessionImpl.GetDatasetByIDOutputs = []testDataStore.GetDatasetByIDOutput{{Dataset: targetUpload, Error: nil}}
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{false}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{authUserID}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{userClient.Permissions{"root": userClient.Permission{}}, nil}}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: false, Error: nil}}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{nil}
			context.MetricClientImpl.RecordMetricOutputs = []error{nil}
		})

		It("succeeds if authenticated as owner", func() {
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as custodian", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{userClient.Permissions{"custodian": userClient.Permission{}}, nil}}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as uploader and was the uploading user", func() {
			targetUpload.ByUser = authUserID
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{userClient.Permissions{"upload": userClient.Permission{}}, nil}}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("succeeds if authenticated as server", func() {
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{true}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if context is missing", func() {
			context.DataSessionImpl.GetDatasetByIDOutputs = []testDataStore.GetDatasetByIDOutput{}
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(nil) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if request is missing", func() {
			context.Context.RequestImpl = nil
			context.DataSessionImpl.GetDatasetByIDOutputs = []testDataStore.GetDatasetByIDOutput{}
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if dataset id not provided as a parameter", func() {
			delete(context.Context.RequestImpl.PathParams, "datasetid")
			context.DataSessionImpl.GetDatasetByIDOutputs = []testDataStore.GetDatasetByIDOutput{}
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorDatasetIDMissing()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if data store session is missing", func() {
			context.DataSessionImpl = nil
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data store session get dataset returns error", func() {
			err := errors.New("other")
			context.DataSessionImpl.GetDatasetByIDOutputs = []testDataStore.GetDatasetByIDOutput{{Dataset: nil, Error: err}}
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get dataset by id", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data store session get dataset returns no dataset", func() {
			context.DataSessionImpl.GetDatasetByIDOutputs = []testDataStore.GetDatasetByIDOutput{{Dataset: nil, Error: nil}}
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorDatasetIDNotFound(targetUpload.UploadID)}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user id is missing on dataset", func() {
			targetUpload.UserID = ""
			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user id from dataset", nil}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if user client is missing", func() {
			context.UserClientImpl = nil
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user client get user permissions returns unauthorized error", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user client get user permissions returns any other error", func() {
			err := errors.New("other")
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user client get user permissions does not return needed permissions", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{userClient.Permissions{"view": userClient.Permission{}}, nil}}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user client get user permissions returns upload permissions, but not user who uploaded", func() {
			context.UserClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{userClient.Permissions{"upload": userClient.Permission{}}, nil}}
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if data deduplicator factory is missing", func() {
			context.DataDeduplicatorFactoryImpl = nil
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data deduplicator factory is registered with dataset returns an error", func() {
			err := errors.New("other")
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: false, Error: err}}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to check if registered with dataset", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data deduplicator factory new registered deduplicator for data returns an error", func() {
			err := errors.New("other")
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: true, Error: nil}}
			context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDatasetOutput{{Deduplicator: nil, Error: err}}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDatasetInputs).To(Equal([]testDataDeduplicator.NewRegisteredDeduplicatorForDatasetInput{{Logger: context.LoggerImpl, DataSession: context.DataSessionImpl, Dataset: targetUpload}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to create registered deduplicator for dataset", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if deduplicator delete dataset returns an error", func() {
			deduplicatorImpl := testData.NewDeduplicator()
			err := errors.New("other")
			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetOutputs = []testDataDeduplicator.IsRegisteredWithDatasetOutput{{Is: true, Error: nil}}
			context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDatasetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDatasetOutput{{Deduplicator: deduplicatorImpl, Error: nil}}
			deduplicatorImpl.DeleteDatasetOutputs = []error{err}
			context.DataSessionImpl.DeleteDatasetOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDatasetInputs).To(Equal([]testDataDeduplicator.NewRegisteredDeduplicatorForDatasetInput{{Logger: context.LoggerImpl, DataSession: context.DataSessionImpl, Dataset: targetUpload}}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete dataset", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
			Expect(deduplicatorImpl.UnusedOutputsCount()).To(Equal(0))
		})

		It("responds with error if data store session delete dataset returns an error", func() {
			err := errors.New("other")
			context.DataSessionImpl.DeleteDatasetOutputs = []error{err}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete dataset", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if metric client is missing", func() {
			context.MetricClientImpl = nil
			Expect(func() { v1.DatasetsDelete(context) }).To(Panic())
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.UserClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("logs and ignores if metric client record metric returns an error", func() {
			context.MetricClientImpl.RecordMetricOutputs = []error{errors.New("other")}
			v1.DatasetsDelete(context)
			Expect(context.DataSessionImpl.GetDatasetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.DataSessionImpl.DeleteDatasetInputs).To(Equal([]*upload.Upload{targetUpload}))
			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "datasets_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})
	})
})
