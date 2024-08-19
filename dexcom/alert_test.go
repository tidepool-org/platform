package dexcom_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/dexcom/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Alert", func() {

	It("AlertStates returns expected", func() {
		Expect(dexcom.AlertStates()).To(Equal([]string{"unknown", "inactive", "activeSnoozed", "activeAlarming"}))
		Expect(dexcom.AlertStates()).To(Equal([]string{
			dexcom.AlertStateUnknown,
			dexcom.AlertStateInactive,
			dexcom.AlertStateActiveSnoozed,
			dexcom.AlertStateActiveAlarming,
		}))
	})

	It("AlertNames returns expected", func() {
		Expect(dexcom.AlertNames()).To(Equal([]string{"unknown", "high", "low", "rise", "fall", "outOfRange", "urgentLow", "urgentLowSoon", "noReadings", "fixedLow"}))
		Expect(dexcom.AlertNames()).To(Equal([]string{
			dexcom.AlertNameUnknown,
			dexcom.AlertNameHigh,
			dexcom.AlertNameLow,
			dexcom.AlertNameRise,
			dexcom.AlertNameFall,
			dexcom.AlertNameOutOfRange,
			dexcom.AlertNameUrgentLow,
			dexcom.AlertNameUrgentLowSoon,
			dexcom.AlertNameNoReadings,
			dexcom.AlertNameFixedLow,
		}))
	})

	Describe("Validate", func() {
		var getTestAlert = func() *dexcom.Alert {
			return test.RandomAlert()
		}
		DescribeTable("requires",
			func(setupAlertFunc func() *dexcom.Alert) {
				testAlert := setupAlertFunc()
				validator := validator.New()
				testAlert.Validate(validator)
				Expect(validator.Error()).To(HaveOccurred())
			},
			Entry("systemTime to be set", func() *dexcom.Alert {
				Skip("systemTime will occassionally be unset that kills the whole upload process")
				alert := getTestAlert()
				alert.SystemTime = nil
				return alert
			}),
			Entry("displayTime to be set", func() *dexcom.Alert {
				Skip("displayTime will occassionally be unset that kills the whole upload process")
				alert := getTestAlert()
				alert.DisplayTime = nil
				return alert
			}),
			Entry("id to be set", func() *dexcom.Alert {
				alert := getTestAlert()
				alert.ID = nil
				return alert
			}),
			Entry("transmitterGeneration to be set", func() *dexcom.Alert {
				alert := getTestAlert()
				alert.TransmitterGeneration = nil
				return alert
			}),

			Entry("alertName to be set", func() *dexcom.Alert {
				alert := getTestAlert()
				alert.AlertName = nil
				return alert
			}),
			Entry("alertName to be set to an allowed value", func() *dexcom.Alert {
				alert := getTestAlert()
				alert.AlertName = pointer.FromString("other")
				return alert
			}),
			Entry("alertState to be set", func() *dexcom.Alert {
				alert := getTestAlert()
				alert.AlertState = nil
				return alert
			}),
			Entry("alertState to be set to an allowed value", func() *dexcom.Alert {
				alert := getTestAlert()
				alert.AlertState = pointer.FromString("other")
				return alert
			}),
			Entry("displayDevice to be set", func() *dexcom.Alert {
				alert := getTestAlert()
				alert.DisplayDevice = nil
				return alert
			}),
		)
	})
})
