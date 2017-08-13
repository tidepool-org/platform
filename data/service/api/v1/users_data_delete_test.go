package v1_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("UsersDataDelete", func() {
	Context("Unit Tests", func() {
		var targetUserID string
		var context *TestContext

		BeforeEach(func() {
			targetUserID = id.New()
			context = NewTestContext()
			context.RequestImpl.PathParams["userid"] = targetUserID
			context.AuthDetailsImpl.IsServerOutputs = []bool{true}
			context.DataSessionImpl.DestroyDataForUserByIDOutputs = []error{nil}
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{nil}
			context.MetricClientImpl.RecordMetricOutputs = []error{nil}
		})

		It("succeeds if authenticated as server", func() {
			v1.UsersDataDelete(context)
			Expect(context.DataSessionImpl.DestroyDataForUserByIDInputs).To(Equal([]string{targetUserID}))
			Expect(context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDInputs).To(Equal([]string{targetUserID}))
			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "users_data_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, []struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if context is missing", func() {
			context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.DataSessionImpl.DestroyDataForUserByIDOutputs = []error{}
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.UsersDataDelete(nil) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if request is missing", func() {
			context.RequestImpl = nil
			context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.DataSessionImpl.DestroyDataForUserByIDOutputs = []error{}
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.UsersDataDelete(nil) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if user id not provided as a parameter", func() {
			delete(context.RequestImpl.PathParams, "userid")
			context.AuthDetailsImpl.IsServerOutputs = []bool{}
			context.DataSessionImpl.DestroyDataForUserByIDOutputs = []error{}
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.UsersDataDelete(context)
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{v1.ErrorUserIDMissing()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if not server", func() {
			context.AuthDetailsImpl.IsServerOutputs = []bool{false}
			context.DataSessionImpl.DestroyDataForUserByIDOutputs = []error{}
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.UsersDataDelete(context)
			Expect(context.RespondWithErrorInputs).To(Equal([]*service.Error{service.ErrorUnauthorized()}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if data store session is missing", func() {
			context.DataSessionImpl = nil
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.UsersDataDelete(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if data store session delete data for user by id returns error", func() {
			err := errors.New("other")
			context.DataSessionImpl.DestroyDataForUserByIDOutputs = []error{err}
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.UsersDataDelete(context)
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete data for user by id", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if sync tasks store session is missing", func() {
			context.SyncTasksSessionImpl = nil
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			Expect(func() { v1.UsersDataDelete(context) }).To(Panic())
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("responds with error if sync task store session delete sync tasks for user by id returns error", func() {
			err := errors.New("other")
			context.SyncTasksSessionImpl.DestroySyncTasksForUserByIDOutputs = []error{err}
			context.MetricClientImpl.RecordMetricOutputs = []error{}
			v1.UsersDataDelete(context)
			Expect(context.RespondWithInternalServerFailureInputs).To(Equal([]RespondWithInternalServerFailureInput{{"Unable to delete sync tasks for user by id", []interface{}{err}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("panics if metric client is missing", func() {
			context.MetricClientImpl = nil
			Expect(func() { v1.UsersDataDelete(context) }).To(Panic())
			Expect(context.DataSessionImpl.DestroyDataForUserByIDInputs).To(Equal([]string{targetUserID}))
			Expect(context.ValidateTest()).To(BeTrue())
		})

		It("logs and ignores if metric client record metric returns an error", func() {
			context.MetricClientImpl.RecordMetricOutputs = []error{errors.New("other")}
			v1.UsersDataDelete(context)
			Expect(context.DataSessionImpl.DestroyDataForUserByIDInputs).To(Equal([]string{targetUserID}))
			Expect(context.MetricClientImpl.RecordMetricInputs).To(Equal([]RecordMetricInput{{context, "users_data_delete", nil}}))
			Expect(context.RespondWithStatusAndDataInputs).To(Equal([]RespondWithStatusAndDataInput{{http.StatusOK, []struct{}{}}}))
			Expect(context.ValidateTest()).To(BeTrue())
		})
	})
})
