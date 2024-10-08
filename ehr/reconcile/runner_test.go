package reconcile_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/tidepool-org/clinic/client"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/clinics"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	taskTest "github.com/tidepool-org/platform/task/test"

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
	var taskClient *taskTest.MockClient
	var logger log.Logger

	BeforeEach(func() {
		authCtrl = gomock.NewController(GinkgoT())
		clinicsCtrl = gomock.NewController(GinkgoT())
		taskCtrl = gomock.NewController(GinkgoT())
		authClient = authTest.NewMockClient(authCtrl)
		clinicsClient = clinics.NewMockClient(clinicsCtrl)
		taskClient = taskTest.NewMockClient(taskCtrl)
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
				tsk, err := task.NewTask(sync.NewTaskCreate(*clinic.Id, sync.DefaultCadence))
				Expect(err).ToNot(HaveOccurred())
				Expect(tsk).ToNot(BeNil())
				tasks[*clinic.Id] = *tsk
			}
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
				setupEHRSettingsForClinics(clinicsClient, clinics)
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
