package sync_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	clinicsTest "github.com/tidepool-org/platform/clinics/test"
	"github.com/tidepool-org/platform/ehr/sync"
)

var _ = Describe("Task", func() {
	Describe("NewTaskCreate", func() {
		It("returns a task create", func() {
			clinic := clinicsTest.NewRandomClinic()
			create := sync.NewTaskCreate(*clinic.Id, sync.DefaultCadence)
			Expect(create).ToNot(BeNil())
			Expect(create.Name).To(PointTo(Equal(sync.TaskName(*clinic.Id))))
			Expect(create.Type).To(Equal(sync.Type))
			Expect(create.AvailableTime).ToNot(BeNil())
		})

		It("stores the clinic id in the data", func() {
			clinic := clinicsTest.NewRandomClinic()
			create := sync.NewTaskCreate(*clinic.Id, sync.DefaultCadence)
			Expect(create).ToNot(BeNil())

			extracted, err := sync.GetClinicId(create.Data)
			Expect(err).ToNot(HaveOccurred())
			Expect(extracted).To(Equal(*clinic.Id))
		})
	})

	Describe("GetCadence", func() {
		It("returns nil if data is nil", func() {
			result := sync.GetCadence(nil)
			Expect(result).To(BeNil())
		})

		It("returns nil if data doesn't have cadence", func() {
			result := sync.GetCadence(map[string]interface{}{})
			Expect(result).To(BeNil())
		})

		It("returns nil if cadence is invalid", func() {
			result := sync.GetCadence(map[string]interface{}{"cadence": "invalid"})
			Expect(result).To(BeNil())
		})

		It("returns the parses cadence correctly", func() {
			period := time.Duration(14) * time.Hour * 24
			result := sync.GetCadence(map[string]interface{}{"cadence": period.String()})
			Expect(result).To(PointTo(Equal(period)))
		})
	})
})
