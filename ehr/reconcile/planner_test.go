package reconcile_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	api "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/ehr/reconcile"
	"github.com/tidepool-org/platform/ehr/sync"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Planner", func() {
	var authCtrl *gomock.Controller
	var clinicsCtrl *gomock.Controller
	var taskCtrl *gomock.Controller

	var clinicsClient *clinics.MockClient
	var logger log.Logger
	var planner *reconcile.Planner

	BeforeEach(func() {
		authCtrl = gomock.NewController(GinkgoT())
		clinicsCtrl = gomock.NewController(GinkgoT())
		taskCtrl = gomock.NewController(GinkgoT())
		clinicsClient = clinics.NewMockClient(clinicsCtrl)
		logger = null.NewLogger()
		planner = reconcile.NewPlanner(clinicsClient, logger)
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
				tsk, err := task.NewTask(sync.NewTaskCreate(*clinic.Id, sync.DefaultCadence))
				Expect(err).ToNot(HaveOccurred())
				Expect(tsk).ToNot(BeNil())
				tasks[*clinic.Id] = *tsk
			}
		})

		Describe("GetReconciliationPlan", func() {
			It("returns an empty plan when each clinic has a corresponding task", func() {
				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				setupEHRSettingsForClinics(clinicsClient, clinics)

				plan, err := planner.GetReconciliationPlan(context.Background(), tasks)
				Expect(err).ToNot(HaveOccurred())
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToDelete).To(BeEmpty())
			})

			It("returns a clinic creation task when a task for the clinic doesn't exist", func() {
				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				setupEHRSettingsForClinics(clinicsClient, clinics)
				delete(tasks, *clinics[0].Id)

				plan, err := planner.GetReconciliationPlan(context.Background(), tasks)
				Expect(err).ToNot(HaveOccurred())
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(HaveLen(1))
				Expect(plan.ToCreate[0].Name).To(PointTo(Equal(sync.TaskName(*clinics[0].Id))))
				Expect(plan.ToDelete).To(BeEmpty())
			})

			It("returns multiple clinic creation tasks when multiple clinics don't exist", func() {
				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				setupEHRSettingsForClinics(clinicsClient, clinics)
				delete(tasks, *clinics[1].Id)
				delete(tasks, *clinics[2].Id)

				plan, err := planner.GetReconciliationPlan(context.Background(), tasks)
				Expect(err).ToNot(HaveOccurred())
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(HaveLen(2))
				Expect(plan.ToCreate[0].Name).To(PointTo(Equal(sync.TaskName(*clinics[1].Id))))
				Expect(plan.ToCreate[1].Name).To(PointTo(Equal(sync.TaskName(*clinics[2].Id))))
				Expect(plan.ToDelete).To(BeEmpty())
			})

			It("returns a clinic for deletion when the task doesn't exist", func() {
				deleted := clinics[2]
				clinics = clinics[0:2]
				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				setupEHRSettingsForClinics(clinicsClient, clinics)

				plan, err := planner.GetReconciliationPlan(context.Background(), tasks)
				Expect(err).ToNot(HaveOccurred())
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToDelete).To(HaveLen(1))
				Expect(plan.ToDelete[0].Name).To(PointTo(Equal(sync.TaskName(*deleted.Id))))
			})

			It("returns multiple clinics for deletion when multiple tasks don't exist", func() {
				firstDeleted := clinics[1]
				secondDeleted := clinics[2]
				clinics = []api.Clinic{clinics[0]}

				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				setupEHRSettingsForClinics(clinicsClient, clinics)

				plan, err := planner.GetReconciliationPlan(context.Background(), tasks)
				Expect(err).ToNot(HaveOccurred())
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToDelete).To(HaveLen(2))
				Expect(
					[]string{*plan.ToDelete[0].Name, *plan.ToDelete[1].Name},
				).To(
					ConsistOf(sync.TaskName(*firstDeleted.Id), sync.TaskName(*secondDeleted.Id)),
				)
			})

			It("returns a task for deletion when the report settings are disabled", func() {
				settings := clinicsTest.NewRandomEHRSettings()
				settings.ScheduledReports.Cadence = api.DISABLED

				clinics = clinics[0:1]
				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				clinicsClient.EXPECT().GetEHRSettings(gomock.Any(), *clinics[0].Id).Return(settings, nil)

				tasks = map[string]task.Task{
					*clinics[0].Id: tasks[*clinics[0].Id],
				}

				plan, err := planner.GetReconciliationPlan(context.Background(), tasks)
				Expect(err).ToNot(HaveOccurred())
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToUpdate).To(BeEmpty())
				Expect(plan.ToDelete).To(HaveLen(1))
			})

			It("returns a task for update when the report cadence is different", func() {
				settings := clinicsTest.NewRandomEHRSettings()
				settings.ScheduledReports.Cadence = api.N7d

				clinics = clinics[0:1]
				clinicsClient.EXPECT().ListEHREnabledClinics(gomock.Any()).Return(clinics, nil)
				clinicsClient.EXPECT().GetEHRSettings(gomock.Any(), *clinics[0].Id).Return(settings, nil)

				tsk := tasks[*clinics[0].Id]
				tasks = map[string]task.Task{
					*clinics[0].Id: tsk,
				}

				plan, err := planner.GetReconciliationPlan(context.Background(), tasks)
				Expect(err).ToNot(HaveOccurred())
				Expect(plan).ToNot(BeNil())
				Expect(plan.ToCreate).To(BeEmpty())
				Expect(plan.ToDelete).To(BeEmpty())
				Expect(plan.ToUpdate).To(HaveLen(1))

				update, exists := plan.ToUpdate[tsk.ID]
				Expect(exists).To(BeTrue())
				Expect(update.Data).ToNot(BeNil())
				Expect((*update.Data)["cadence"]).To(Equal("168h0m0s"))
			})
		})
	})
})

func setupEHRSettingsForClinics(clinicsClient *clinics.MockClient, clinics []api.Clinic) {
	for _, clinic := range clinics {
		clinicsClient.EXPECT().GetEHRSettings(gomock.Any(), *clinic.Id).Return(clinicsTest.NewRandomEHRSettings(), nil)
	}
}
