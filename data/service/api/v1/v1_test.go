package v1_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataServiceApiV1 "github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/data/summary/reporters"
	"github.com/tidepool-org/platform/request"
)

var _ = Describe("V1", func() {
	Context("Routes", func() {
		It("returns the correct routes", func() {
			Expect(dataServiceApiV1.Routes()).ToNot(BeEmpty())
		})
	})

	Context("GetPatientsWithRealtimeData request parse", func() {
		It("full request url", func() {
			loc := time.FixedZone("-0400", -4*60*60)
			startDate := time.Date(2024, 4, 25, 0, 0, 0, 0, loc)
			endDate := time.Date(2024, 5, 24, 23, 59, 59, 999000000, loc)

			filter := reporters.NewPatientRealtimeDaysFilter()
			url := "https://qa2.development.tidepool.org/v1/clinics/12345asdf/reports/realtime?startDate=2024-04-25T00%3A00%3A00.000-04%3A00&endDate=2024-05-24T23%3A59%3A59.999-04%3A00&patientFilters=%7B%22cgm.lastUploadDateTo%22%3A%222024-05-25T04%3A00%3A00.000Z%22%2C%22cgm.lastUploadDateFrom%22%3A%222024-04-25T04%3A00%3A00.000Z%22%2C%22tags%22%3A%5B%226577551586650af25e519385%22%2C%22657754bc86650af25e519372%22%5D%2C%22cgm.timeInLowPercent%22%3A%22%3E%3D0.04%22%2C%22cgm.timeInVeryLowPercent%22%3A%22%3E%3D0.01%22%2C%22cgm.timeInTargetPercent%22%3A%22%3C%3D0.7%22%2C%22cgm.timeInHighPercent%22%3A%22%3E%3D0.25%22%2C%22cgm.timeInVeryHighPercent%22%3A%22%3E%3D0.05%22%2C%22cgm.timeCGMUsePercent%22%3A%22%3E%3D0.7%22%7D"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).ToNot(HaveOccurred())

			err = request.DecodeRequestQuery(req, filter)
			Expect(err).ToNot(HaveOccurred())

			Expect(filter.StartTime).ToNot(BeNil())
			Expect(filter.StartTime.Equal(startDate)).To(BeTrue())

			Expect(filter.EndTime).ToNot(BeNil())
			Expect(filter.EndTime.Equal(endDate)).To(BeTrue())

			Expect(filter.PatientFilters).ToNot(BeNil())
			Expect(*filter.PatientFilters.CgmTimeInLowPercent).To(Equal(">=0.04"))
		})
	})

})
