package sync_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/ehr/sync"
)

var _ = Describe("Task", func() {
	Describe("NewTaskCreate", func() {
		It("returns a task create", func() {
			clinic := clinicsTest.NewRandomClinic()
			create := sync.NewTaskCreate(*clinic.Id)
			Expect(create).ToNot(BeNil())
			Expect(create.Name).To(PointTo(Equal(sync.TaskName(*clinic.Id))))
			Expect(create.Type).To(Equal(sync.Type))
			Expect(create.AvailableTime).ToNot(BeNil())
		})

		It("stores the clinic id in the data", func() {
			clinic := clinicsTest.NewRandomClinic()
			create := sync.NewTaskCreate(*clinic.Id)
			Expect(create).ToNot(BeNil())

			extracted, err := sync.GetClinicId(create.Data)
			Expect(err).ToNot(HaveOccurred())
			Expect(extracted).To(Equal(*clinic.Id))
		})
	})
})
