package v1_test

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("UsersDataDelete", func() {
	// 	Context("Unit Tests", func() {
	// 		var targetUserID string
	// 		var context *TestContext

	// 		BeforeEach(func() {
	// 			targetUserID = id.New()
	// 			context = NewTestContext()
	// 			context.RequestImpl.PathParams["user_id"] = targetUserID
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{true}
	// 			context.DataRepositoryImpl.DestroyDataForUserByIDOutputs = []error{nil}
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{nil}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{nil}
	// 		})

	// 		It("succeeds if authenticated as server", func() {
	// 			v1.UsersDataDelete(context)
	// 			Expect(context.DataRepositoryImpl.DestroyDataForUserByIDInputs).To(Equal([]string{targetUserID}))
	// 			Expect(context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDInputs).To(Equal([]string{targetUserID}))
	// 			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "users_data_delete", nil}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, []struct{}{}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if context is missing", func() {
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.DataRepositoryImpl.DestroyDataForUserByIDOutputs = []error{}
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.UsersDataDelete(nil) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if request is missing", func() {
	// 			context.RequestImpl = nil
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.DataRepositoryImpl.DestroyDataForUserByIDOutputs = []error{}
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.UsersDataDelete(nil) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if user id not provided as a parameter", func() {
	// 			delete(context.RequestImpl.PathParams, "user_id")
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{}
	// 			context.DataRepositoryImpl.DestroyDataForUserByIDOutputs = []error{}
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.UsersDataDelete(context)
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDMissing()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if not server", func() {
	// 			context.AuthDetailsImpl.IsServerOutputs = []bool{false}
	// 			context.DataRepositoryImpl.DestroyDataForUserByIDOutputs = []error{}
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.UsersDataDelete(context)
	// 			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if data store Repository is missing", func() {
	// 			context.DataRepositoryImpl = nil
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.UsersDataDelete(context) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if data store Repository delete data for user by id returns error", func() {
	// 			err := errors.New("other")
	// 			context.DataRepositoryImpl.DestroyDataForUserByIDOutputs = []error{err}
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.UsersDataDelete(context)
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete data for user by id", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if sync tasks store Repository is missing", func() {
	// 			context.SyncTaskRepositoryImpl = nil
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			Expect(func() { v1.UsersDataDelete(context) }).To(Panic())
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("responds with error if sync task store Repository delete sync tasks for user by id returns error", func() {
	// 			err := errors.New("other")
	// 			context.SyncTaskRepositoryImpl.DestroySyncTasksForUserByIDOutputs = []error{err}
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{}
	// 			v1.UsersDataDelete(context)
	// 			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete sync tasks for user by id", []interface{}{err}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("panics if metric client is missing", func() {
	// 			context.MetricClientImpl = nil
	// 			Expect(func() { v1.UsersDataDelete(context) }).To(Panic())
	// 			Expect(context.DataRepositoryImpl.DestroyDataForUserByIDInputs).To(Equal([]string{targetUserID}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})

	// 		It("logs and ignores if metric client record metric returns an error", func() {
	// 			context.MetricClientImpl.RecordMetricOutputs = []error{errors.New("other")}
	// 			v1.UsersDataDelete(context)
	// 			Expect(context.DataRepositoryImpl.DestroyDataForUserByIDInputs).To(Equal([]string{targetUserID}))
	// 			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "users_data_delete", nil}}))
	// 			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, []struct{}{}}}))
	// 			Expect(context.ValidateTest()).To(BeTrue())
	// 		})
	// 	})
})
