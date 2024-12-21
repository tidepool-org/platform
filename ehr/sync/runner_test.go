package sync_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/tidepool-org/clinic/client"
	"go.uber.org/mock/gomock"

	"github.com/tidepool-org/platform/ehr/sync"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/task"
)

var _ = Describe("Runner", func() {
	var clinicsCtrl *gomock.Controller
	var clinicsClient *clinicsTest.MockClient
	var logger log.Logger

	BeforeEach(func() {
		clinicsCtrl = gomock.NewController(GinkgoT())
		clinicsClient = clinicsTest.NewMockClient(clinicsCtrl)
		logger = null.NewLogger()
	})

	AfterEach(func() {
		clinicsCtrl.Finish()
	})

	Describe("NewRunner", func() {
		It("returns successfully", func() {
			Expect(sync.NewRunner(clinicsClient, logger)).ToNot(BeNil())
		})
	})

	Describe("Run", func() {
		var tsk task.Task
		var clinic api.Clinic

		BeforeEach(func() {
			clinic = clinicsTest.NewRandomClinic()
			t, err := task.NewTask(context.Background(), sync.NewTaskCreate(*clinic.Id, sync.DefaultCadence))
			Expect(err).ToNot(HaveOccurred())
			Expect(t).ToNot(BeNil())
			tsk = *t
		})

		It("calls sync for the clinics service", func() {
			clinicsClient.EXPECT().SyncEHRData(gomock.Any(), *clinic.Id).Return(nil)

			runner, err := sync.NewRunner(clinicsClient, logger)

			Expect(err).ToNot(HaveOccurred())
			Expect(runner).ToNot(BeNil())
			runner.Run(context.Background(), &tsk)
		})
	})
})
