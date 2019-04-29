package v1_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("DataSetsDelete", func() {
	// 	Context("Unit Tests", func() {
	// 		var authUserID string
	// 		var targetUserID string
	// 		var targetUpload *upload.Upload
	// 		var context *TestContext

	// 		BeforeEach(func() {
	// 			authUserID = id.New()
	// 			targetUserID = id.New()
	// 			targetUpload = upload.New()
	// 			targetUpload.UserID = targetUserID
	// 			targetUpload.ByUser = pointer.FromString(id.New())
	// 			context = NewTestContext()
	// 			context.Context.RequestImpl.PathParams["data_set_id"] = targetUpload.UploadID
	// 			context.DataSessionImpl.GetDataSetByIDOutputs = []testDataStoreDEPRECATED.GetDataSetByIDOutput{{DataSet: targetUpload, Error: nil}}
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{false}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{authUserID}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{permission.Permissions{"root": permission.Permission{}}, nil}}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: nil}}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{nil}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{nil}
	// 		})

	// 		It("succeeds if authenticated as owner", func() {
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataSessionImpl.DeleteDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "data_sets_delete", nil}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if authenticated as custodian", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{permission.Permissions{"custodian": permission.Permission{}}, nil}}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataSessionImpl.DeleteDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "data_sets_delete", nil}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if authenticated as uploader and was the uploading user", func() {
	// 			targetUpload.ByUser = pointer.FromString(authUserID)
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{permission.Permissions{"upload": permission.Permission{}}, nil}}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataSessionImpl.DeleteDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "data_sets_delete", nil}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("succeeds if authenticated as server", func() {
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{true}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataSessionImpl.DeleteDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "data_sets_delete", nil}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if context is missing", func() {
	// 			context.DataSessionImpl.GetDataSetByIDOutputs = []testDataStoreDEPRECATED.GetDataSetByIDOutput{}
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.DataSetsDelete(nil) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if request is missing", func() {
	// 			context.Context.RequestImpl = nil
	// 			context.DataSessionImpl.GetDataSetByIDOutputs = []testDataStoreDEPRECATED.GetDataSetByIDOutput{}
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.DataSetsDelete(context) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if data set id not provided as a parameter", func() {
	// 			delete(context.Context.RequestImpl.PathParams, "data_set_id")
	// 			context.DataSessionImpl.GetDataSetByIDOutputs = []testDataStoreDEPRECATED.GetDataSetByIDOutput{}
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorDataSetIDMissing()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if data store session is missing", func() {
	// 			context.DataSessionImpl = nil
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.DataSetsDelete(context) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if data store session get data set returns error", func() {
	// 			err := errors.New("other")
	// 			context.DataSessionImpl.GetDataSetByIDOutputs = []testDataStoreDEPRECATED.GetDataSetByIDOutput{{DataSet: nil, Error: err}}
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get data set by id", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if data store session get data set returns no data set", func() {
	// 			context.DataSessionImpl.GetDataSetByIDOutputs = []testDataStoreDEPRECATED.GetDataSetByIDOutput{{DataSet: nil, Error: nil}}
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorDataSetIDNotFound(targetUpload.UploadID)}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user id is missing on data set", func() {
	// 			targetUpload.UserID = ""
	// 			context.Context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.Context.AuthDetailsImpl.UserIDOutputs = []string{}
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user id from data set", nil}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if user client is missing", func() {
	// 			context.PermissionClientImpl = nil
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.DataSetsDelete(context) }).To(Panic())
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user client get user permissions returns unauthorized error", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, client.NewUnauthorizedError()}}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user client get user permissions returns any other error", func() {
	// 			err := errors.New("other")
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{nil, err}}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to get user permissions", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user client get user permissions does not return needed permissions", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{permission.Permissions{"view": permission.Permission{}}, nil}}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user client get user permissions returns upload permissions, but not user who uploaded", func() {
	// 			context.PermissionClientImpl.GetUserPermissionsOutputs = []GetUserPermissionsOutput{{permission.Permissions{"upload": permission.Permission{}}, nil}}
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if data deduplicator factory is missing", func() {
	// 			context.DataDeduplicatorFactoryImpl = nil
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.DataSetsDelete(context) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if data deduplicator factory is registered with data set returns an error", func() {
	// 			err := errors.New("other")
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: false, Error: err}}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to check if registered with data set", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if data deduplicator factory new registered deduplicator for data returns an error", func() {
	// 			err := errors.New("other")
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
	// 			context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDataSetOutput{{Deduplicator: nil, Error: err}}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDataSetInputs).To(Equal([]testDataDeduplicator.NewRegisteredDeduplicatorForDataSetInput{{Logger: context.LoggerImpl, DataSession: context.DataSessionImpl, DataSet: targetUpload}}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to create registered deduplicator for data set", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if deduplicator delete data set returns an error", func() {
	// 			deduplicatorImpl := testData.NewDeduplicator()
	// 			err := errors.New("other")
	// 			context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetOutputs = []testDataDeduplicator.IsRegisteredWithDataSetOutput{{Is: true, Error: nil}}
	// 			context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDataSetOutputs = []testDataDeduplicator.NewRegisteredDeduplicatorForDataSetOutput{{Deduplicator: deduplicatorImpl, Error: nil}}
	// 			deduplicatorImpl.DeleteDataSetOutputs = []error{err}
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.NewRegisteredDeduplicatorForDataSetInputs).To(Equal([]testDataDeduplicator.NewRegisteredDeduplicatorForDataSetInput{{Logger: context.LoggerImpl, DataSession: context.DataSessionImpl, DataSet: targetUpload}}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete data set", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 			Expect(deduplicatorImpl.UnusedOutputsCount()).To(Equal(0))
	// 		})

	// 		It("responds with error if data store session delete data set returns an error", func() {
	// 			err := errors.New("other")
	// 			context.DataSessionImpl.DeleteDataSetOutputs = []error{err}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataSessionImpl.DeleteDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete data set", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if metric client is missing", func() {
	// 			context.MetricClientImpl = nil
	// 			Expect(func() { v1.DataSetsDelete(context) }).To(Panic())
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.PermissionClientImpl.GetUserPermissionsInputs).To(Equal([]GetUserPermissionsInput{{context, authUserID, targetUserID}}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataSessionImpl.DeleteDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("logs and ignores if metric client record metric returns an error", func() {
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{errors.New("other")}
	// 			v1.DataSetsDelete(context)
	// 			Expect(context.DataSessionImpl.GetDataSetByIDInputs).To(Equal([]string{targetUpload.UploadID}))
	// 			Expect(context.DataDeduplicatorFactoryImpl.IsRegisteredWithDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.DataSessionImpl.DeleteDataSetInputs).To(Equal([]*upload.Upload{targetUpload}))
	// 			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "data_sets_delete", nil}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, struct{}{}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})
	// 	})
})
