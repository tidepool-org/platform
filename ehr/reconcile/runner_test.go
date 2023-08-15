package reconcile_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	api "github.com/tidepool-org/clinic/client"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/ehr/reconcile"
	"github.com/tidepool-org/platform/ehr/sync"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Runner", func() {
	Describe("GetReconciliationPlan", func() {
		var clinics []api.Clinic
		var tasks map[string]task.Task

		BeforeEach(func() {
			clinics = test.RandomArrayWithLength(3, clinicsTest.NewRandomClinic)
			tasks = make(map[string]task.Task)
			for _, clinic := range clinics {
				tsk, err := task.NewTask(sync.NewTaskCreate(*clinic.Id))
				Expect(err).ToNot(HaveOccurred())
				Expect(tsk).ToNot(BeNil())
				tasks[*clinic.Id] = *tsk
			}
		})

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

		It("returns a multiple clinic creation tasks when multiple clinics don't exist", func() {
			delete(tasks, *clinics[0].Id)
			delete(tasks, *clinics[1].Id)
			plan := reconcile.GetReconciliationPlan(tasks, clinics)
			Expect(plan).ToNot(BeNil())
			Expect(plan.ToCreate).To(HaveLen(2))
			Expect(plan.ToCreate[0].Name).To(PointTo(Equal(sync.TaskName(*clinics[0].Id))))
			Expect(plan.ToCreate[1].Name).To(PointTo(Equal(sync.TaskName(*clinics[1].Id))))
			Expect(plan.ToDelete).To(BeEmpty())
		})
	})
})
