package task_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/pointer"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/task"
	taskTest "github.com/tidepool-org/platform/task/test"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Task", func() {
	Context("NewID", func() {
		It("returns a string of 32 lowercase hexadecimal characters", func() {
			Expect(task.NewID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(task.NewID()).ToNot(Equal(task.NewID()))
		})
	})

	Context("IsValidID, IDValidator, and ValidateID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(task.IsValidID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				task.IDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(task.ValidateID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "0123456789abcdef0123456789abcde", task.ErrorValueStringAsIDNotValid("0123456789abcdef0123456789abcde")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexadecimalLowercase)),
			Entry("has string length out of range (upper)", "0123456789abcdef0123456789abcdef0", task.ErrorValueStringAsIDNotValid("0123456789abcdef0123456789abcdef0")),
			Entry("has uppercase characters", "0123456789ABCDEF0123456789abcdef", task.ErrorValueStringAsIDNotValid("0123456789ABCDEF0123456789abcdef")),
			Entry("has symbols", "0123456789$%^&*(0123456789abcdef", task.ErrorValueStringAsIDNotValid("0123456789$%^&*(0123456789abcdef")),
			Entry("has whitespace", "0123456789      0123456789abcdef", task.ErrorValueStringAsIDNotValid("0123456789      0123456789abcdef")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsIDNotValid with empty string", task.ErrorValueStringAsIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as task id`),
			Entry("is ErrorValueStringAsIDNotValid with non-empty string", task.ErrorValueStringAsIDNotValid("0123456789abcdef0123456789abcdef"), "value-not-valid", "value is not valid", `value "0123456789abcdef0123456789abcdef" is not valid as task id`),
		)
	})

	Context("Task", func() {
		Context("NewTask", func() {
			var ctx context.Context
			var create *task.TaskCreate

			BeforeEach(func() {
				ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				create = taskTest.RandomTaskCreate(test.AllowOptionals())
			})

			It("returns an error when the create is missing", func() {
				result, err := task.NewTask(ctx, nil)
				errorsTest.ExpectEqual(err, errors.New("create is missing"))
				Expect(result).To(BeNil())
			})

			It("returns an error when the create is invalid", func() {
				create.Type = ""
				result, err := task.NewTask(ctx, create)
				errorsTest.ExpectEqual(err, errors.New("create is invalid"))
				Expect(result).To(BeNil())
			})

			It("returns a new pending task using the current time as the available time when the create does not specify one", func() {
				create.AvailableTime = nil

				result := test.Must(task.NewTask(ctx, create))
				Expect(result).To(PointTo(MatchAllFields(Fields{
					"ID":            MatchRegexp("^[0-9a-f]{32}$"),
					"Name":          Equal(create.Name),
					"Type":          Equal(create.Type),
					"Data":          Equal(create.Data),
					"AvailableTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
					"DeadlineTime":  BeNil(),
					"State":         Equal(task.TaskStatePending),
					"Error":         BeNil(),
					"RunTime":       BeNil(),
					"Duration":      BeNil(),
					"CreatedTime":   BeTemporally("~", time.Now(), time.Second),
					"ModifiedTime":  BeNil(),
					"Revision":      Equal(1),
					"StateLock":     BeNil(),
				})))
			})

			It("returns a new pending task using the current time as the available time when the create specifies one in the past", func() {
				create.AvailableTime = pointer.From(test.RandomTimeBeforeNow())

				result := test.Must(task.NewTask(ctx, create))
				Expect(result.AvailableTime).To(PointTo(BeTemporally("~", time.Now(), time.Second)))
			})

			It("returns a new pending task using the specified available time when it is in the future", func() {
				create.AvailableTime = pointer.From(test.RandomTimeAfterNow())

				result := test.Must(task.NewTask(ctx, create))
				Expect(result.AvailableTime).To(PointTo(Equal(*create.AvailableTime)))
			})
		})

		Context("LogFields", func() {
			It("returns the id, type, and state as log fields", func() {
				tsk := taskTest.RandomTask()
				Expect(tsk.LogFields()).To(Equal(log.Fields{
					"id":        tsk.ID,
					"type":      tsk.Type,
					"state":     tsk.State,
					"revision":  tsk.Revision,
					"stateLock": tsk.StateLock,
				}))
			})
		})
	})
})
