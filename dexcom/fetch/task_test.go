package fetch_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/tidepool-org/platform/dexcom/fetch"
	"github.com/tidepool-org/platform/task"
)

var _ = Describe("Task", func() {

	var getTask = func(retryCount int) *task.Task {
		return &task.Task{
			Data: map[string]interface{}{
				"retryCount": retryCount,
			},
		}
	}

	Context("SetErrorOrAllowTaskRetry", func() {
		DescribeTable("will not append error",
			func(setupFunc func() (*task.Task, int)) {
				tsk, startCount := setupFunc()
				Expect(tsk.Data["retryCount"]).To(Equal(startCount))
				fetch.SetErrorOrAllowTaskRetry(tsk, errors.New("some error"))
				Expect(tsk.HasError()).To(Equal(false))
				Expect(tsk.IsFailed()).To(Equal(false))
				Expect(tsk.Data["retryCount"]).To(Equal(startCount + 1))
			},
			Entry("if zero retries", func() (*task.Task, int) {
				return getTask(0), 0
			}),
			Entry("if one retry", func() (*task.Task, int) {
				return getTask(1), 1
			}),
			Entry("if two retries", func() (*task.Task, int) {
				return getTask(2), 2
			}),
			Entry("if three retries", func() (*task.Task, int) {
				return getTask(3), 3
			}),
		)
		DescribeTable("will append error",
			func(setupFunc func() (*task.Task, int)) {
				tsk, startCount := setupFunc()
				Expect(tsk.Data["retryCount"]).To(Equal(startCount))
				fetch.SetErrorOrAllowTaskRetry(tsk, errors.New("some error"))
				Expect(tsk.HasError()).To(Equal(true))
				Expect(tsk.IsFailed()).To(Equal(true))
			},
			Entry("when 4th retry", func() (*task.Task, int) {
				return getTask(4), 4
			}),
			Entry("more than 4 retries", func() (*task.Task, int) {
				return getTask(10), 10
			}),
		)
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
