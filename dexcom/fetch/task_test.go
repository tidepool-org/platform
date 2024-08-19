package fetch_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/dexcom/fetch"
	"github.com/tidepool-org/platform/log"

	logNull "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/task"
)

var _ = Describe("Task", func() {

	var getTask = func(retryCount int, hasFailed bool) *task.Task {
		tsk := &task.Task{
			Data: map[string]interface{}{
				"retryCount": retryCount,
			},
		}
		if hasFailed {
			tsk.SetFailed()
		}
		return tsk
	}

	Context("ErrorOrRetryTask", func() {
		DescribeTable("will append error, but not fail",
			func(setupFunc func() (*task.Task, int)) {
				tsk, startCount := setupFunc()
				Expect(tsk.IsFailed()).To(BeTrue())
				Expect(tsk.Data["retryCount"]).To(Equal(startCount))
				fetch.ErrorOrRetryTask(tsk, errors.New("some error"))
				Expect(tsk.HasError()).To(BeTrue())
				Expect(tsk.IsFailed()).To(BeFalse())
				Expect(tsk.Data["retryCount"]).To(Equal(startCount + 1))
			},
			Entry("if zero retries", func() (*task.Task, int) {
				retryCount := 0
				return getTask(retryCount, true), retryCount
			}),
			Entry("if one retry", func() (*task.Task, int) {
				retryCount := 1
				return getTask(retryCount, true), retryCount
			}),
			Entry("if two retries", func() (*task.Task, int) {
				retryCount := 2
				return getTask(retryCount, true), retryCount
			}),
		)
		DescribeTable("will append error and fail",
			func(setupFunc func() (*task.Task, int)) {
				tsk, startCount := setupFunc()
				Expect(tsk.IsFailed()).To(BeTrue())
				Expect(tsk.Data["retryCount"]).To(Equal(startCount))
				fetch.ErrorOrRetryTask(tsk, errors.New("some error"))
				Expect(tsk.HasError()).To(BeTrue())
				Expect(tsk.IsFailed()).To(BeTrue())
				Expect(tsk.Data["retryCount"]).To(Equal(startCount))
			},
			Entry("when 3rd retry", func() (*task.Task, int) {
				retryCount := 3
				return getTask(retryCount, true), retryCount
			}),
			Entry("more than 3 retries", func() (*task.Task, int) {
				retryCount := 10
				return getTask(retryCount, true), retryCount
			}),
		)
		DescribeTable("will set the error even if the task is not failed",
			func(setupFunc func() (*task.Task, int)) {
				tsk, startCount := setupFunc()
				Expect(tsk.IsFailed()).To(BeFalse())
				Expect(tsk.Data["retryCount"]).To(Equal(startCount))
				fetch.ErrorOrRetryTask(tsk, errors.New("some error"))
				Expect(tsk.HasError()).To(BeTrue())
				Expect(tsk.IsFailed()).To(BeFalse())
				Expect(tsk.Data["retryCount"]).To(Equal(startCount))
			},
			Entry("when 3rd retry", func() (*task.Task, int) {
				retryCount := 3
				return getTask(retryCount, false), retryCount
			}),
			Entry("more 1st retry", func() (*task.Task, int) {
				retryCount := 1
				return getTask(retryCount, false), retryCount
			}),
		)
	})

	Context("FailTask", func() {
		var logger log.Logger
		BeforeEach(func() {
			logger = logNull.NewLogger()
		})
		It("will set the task to have failed", func() {
			tsk := getTask(0, false)
			Expect(tsk.IsFailed()).To(BeFalse())
			fetch.FailTask(logger, tsk, errors.New("some error"))
			Expect(tsk.IsFailed()).To(BeTrue())
		})
		It("will not change the failure status if already set", func() {
			tsk := getTask(0, false)
			tsk.SetFailed()
			Expect(tsk.IsFailed()).To(BeTrue())
			fetch.FailTask(logger, tsk, errors.New("some error"))
			Expect(tsk.IsFailed()).To(BeTrue())
		})
	})

	Context("NewTaskCreate", func() {
		const providerID = "some-provider-id"
		const sourceID = "some-source-id"

		It("returns an error when provider session id not set", func() {
			tc, err := fetch.NewTaskCreate("", sourceID)
			Expect(err).ToNot(BeNil())
			Expect(tc).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("provider session id is missing"))
		})
		It("returns an error when data source id not set", func() {
			tc, err := fetch.NewTaskCreate(providerID, "")
			Expect(err).ToNot(BeNil())
			Expect(tc).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("data source id is missing"))
		})
		It("returns an initialised task create", func() {
			tc, err := fetch.NewTaskCreate(providerID, sourceID)
			Expect(err).To(BeNil())
			Expect(tc).ToNot(BeNil())
		})

		It("task has data initialised", func() {
			tc, _ := fetch.NewTaskCreate(providerID, sourceID)
			Expect(tc).ToNot(BeNil())
			Expect(tc.Type).To(Equal(fetch.Type))
			Expect(tc.Data["providerSessionId"]).To(Equal(providerID))
			Expect(tc.Data["dataSourceId"]).To(Equal(sourceID))
			Expect(tc.Data["retryCount"]).To(Equal(0))
		})

	})
})
