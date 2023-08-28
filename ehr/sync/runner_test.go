package sync

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	api "github.com/tidepool-org/clinic/client"

	"github.com/tidepool-org/platform/clinics"
	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/task"
)

var _ = Describe("Runner", func() {
	var clinicsCtrl *gomock.Controller
	var clinicsClient *clinics.MockClient
	var logger log.Logger

	BeforeEach(func() {
		clinicsCtrl = gomock.NewController(GinkgoT())
		clinicsClient = clinics.NewMockClient(clinicsCtrl)
		logger = null.NewLogger()
	})

	AfterEach(func() {
		clinicsCtrl.Finish()
	})

	Describe("NewRunner", func() {
		It("returns successfully", func() {
			Expect(NewRunner(clinicsClient, logger)).ToNot(BeNil())
		})
	})

	Describe("Run", func() {
		var tsk task.Task
		var clinic api.Clinic

		BeforeEach(func() {
			clinic = clinicsTest.NewRandomClinic()
			t, err := task.NewTask(NewTaskCreate(*clinic.Id))
			Expect(err).ToNot(HaveOccurred())
			Expect(t).ToNot(BeNil())
			tsk = *t
		})

		It("calls sync for the clinics service", func() {
			clinicsClient.EXPECT().SyncEHRData(gomock.Any(), *clinic.Id).Return(nil)

			runner, err := NewRunner(clinicsClient, logger)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner).ToNot(BeNil())
			Expect(runner.Run(nil, &tsk)).To(BeTrue())
		})
	})
})
