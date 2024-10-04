package reconcile_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	api "github.com/tidepool-org/clinic/client"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/task/tasktest"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/ehr/reconcile"
	"github.com/tidepool-org/platform/ehr/sync"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Runner", func() {
	var authCtrl *gomock.Controller
	var clinicsCtrl *gomock.Controller
	var taskCtrl *gomock.Controller

	var authClient *authTest.MockClient
	var clinicsClient *clinics.MockClient
	var taskClient *tasktest.MockClient
	var logger log.Logger

	BeforeEach(func() {
		authCtrl = gomock.NewController(GinkgoT())
		clinicsCtrl = gomock.NewController(GinkgoT())
		taskCtrl = gomock.NewController(GinkgoT())
		authClient = authTest.NewMockClient(authCtrl)
		clinicsClient = clinics.NewMockClient(clinicsCtrl)
		taskClient = tasktest.NewMockClient(taskCtrl)
		logger = null.NewLogger()
	})

	AfterEach(func() {
		authCtrl.Finish()
		clinicsCtrl.Finish()
		taskCtrl.Finish()
	})

	Context("With random data", func() {
		var clinics []api.Clinic
		var tasks map[string]task.Task

		BeforeEach(func() {
			clinics = test.RandomArrayWithLength(3, clinicsTest.NewRandomClinic)
			tasks = make(map[string]task.Task)
			for _, clinic := range clinics {
				clinic := clinic
				tsk, err := task.NewTask(sync.NewTaskCreate(*clinic.Id))
				Expect(err).ToNot(HaveOccurred())
				Expect(tsk).ToNot(BeNil())
				tasks[*clinic.Id] = *tsk
			}
		})

		Describe("GetReconciliationPlan", func() {
			It("returns an empty plan when each clinic has a corresponding task", func() {
				plan := reconcile.GetReconciliationPlan(tasks, clinics)
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToDelete).To(BeEmpty())
			})

			It("returns a clinic creation task when a task for the clinic doesn't exist", func() {
				delete(tasks, *clinics[0].Id)
				plan := reconcile.GetReconciliationPlan(tasks, clinics)
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(HaveLen(1))
				Expect(plan.ToCreate[0].Name).To(PointTo(Equal(sync.TaskName(*clinics[0].Id))))
				Expect(plan.ToDelete).To(BeEmpty())
			})

			It("returns multiple clinic creation tasks when multiple clinics don't exist", func() {
				delete(tasks, *clinics[1].Id)
				delete(tasks, *clinics[2].Id)
				plan := reconcile.GetReconciliationPlan(tasks, clinics)
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(HaveLen(2))
				Expect(plan.ToCreate[0].Name).To(PointTo(Equal(sync.TaskName(*clinics[1].Id))))
				Expect(plan.ToCreate[1].Name).To(PointTo(Equal(sync.TaskName(*clinics[2].Id))))
				Expect(plan.ToDelete).To(BeEmpty())
			})

			It("returns a clinic for deletion when the task doesn't exist", func() {
				deleted := clinics[2]
				clinics = clinics[0:2]
				plan := reconcile.GetReconciliationPlan(tasks, clinics)
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToDelete).To(HaveLen(1))
				Expect(plan.ToDelete[0].Name).To(PointTo(Equal(sync.TaskName(*deleted.Id))))
			})

			It("returns multiple clinics for deletion when multiple tasks don't exist", func() {
				firstDeleted := clinics[1]
				secondDeleted := clinics[2]
				clinics = []api.Clinic{clinics[0]}
				plan := reconcile.GetReconciliationPlan(tasks, clinics)
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToDelete).To(HaveLen(2))
				Expect(
					[]string{*plan.ToDelete[0].Name, *plan.ToDelete[1].Name},
				).To(
					ConsistOf(sync.TaskName(*firstDeleted.Id), sync.TaskName(*secondDeleted.Id)),
				)
			})
		})

		Describe("NewRunner", func() {
			It("returns successfully", func() {
				runner, err := reconcile.NewRunner(authClient, clinicsClient, taskClient, logger)
				Expect(err).ToNot(HaveOccurred())
				Expect(runner).ToNot(BeNil())
			})
		})

		Describe("Run", func() {
			It("works correctly", func() {
				runner, err := reconcile.NewRunner(authClient, clinicsClient, taskClient, logger)
				Expect(err).ToNot(HaveOccurred())
				Expect(runner).ToNot(BeNil())

				t, err := task.NewTask(reconcile.NewTaskCreate())
				Expect(err).ToNot(HaveOccurred())
				Expect(t).ToNot(BeNil())

				toBeDeleted := clinics[2]
				clinics = clinics[0:2]

				//taskCreate := sync.NewTaskCreate(*clinics[0].Id)
				delete(tasks, *clinics[0].Id)

				var tasksList task.Tasks
				for _, t := range tasks {
					t := t
					tasksList = append(tasksList, &t)
				}

				authClient.EXPECT().ServerSessionToken().Return("token", nil)
				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				taskClient.EXPECT().ListTasks(gomock.Any(), gomock.Any(), gomock.Any()).Return(tasksList, nil)
				taskClient.EXPECT().DeleteTask(gomock.Any(), gomock.Eq(tasks[*toBeDeleted.Id].ID)).Return(nil)
				taskClient.EXPECT().CreateTask(gomock.Any(), gomock.Any()).Return(nil, nil)
				runner.Run(context.Background(), t)
			})
		})
	})
})
