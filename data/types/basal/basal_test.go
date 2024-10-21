package basal_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/basal"
	dataTypesBasalTest "github.com/tidepool-org/platform/data/types/basal/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const ExpectedTimeFormat = time.RFC3339Nano

var _ = Describe("Basal", func() {
	It("Type is expected", func() {
		Expect(basal.Type).To(Equal("basal"))
	})

	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			deliveryType := dataTypesTest.NewType()
			datum := basal.New(deliveryType)
			Expect(datum.Type).To(Equal("basal"))
			Expect(datum.DeliveryType).To(Equal(deliveryType))
		})
	})

	Context("with new datum", func() {
		var deliveryType string
		var datum basal.Basal

		BeforeEach(func() {
			deliveryType = dataTypesTest.NewType()
			datum = basal.New(deliveryType)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&basal.Meta{Type: "basal", DeliveryType: deliveryType}))
			})
		})
	})

	Context("Basal", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *basal.Basal), expectedErrors ...error) {
					datum := dataTypesBasalTest.RandomBasal()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *basal.Basal) {},
				),
				Entry("type missing",
					func(datum *basal.Basal) { datum.Type = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type invalid",
					func(datum *basal.Basal) { datum.Type = "invalid" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
				),
				Entry("type basal",
					func(datum *basal.Basal) { datum.Type = "basal" },
				),
				Entry("delivery type missing",
					func(datum *basal.Basal) { datum.DeliveryType = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deliveryType"),
				),
				Entry("delivery type valid",
					func(datum *basal.Basal) { datum.DeliveryType = dataTypesTest.NewType() },
				),
				Entry("multiple errors",
					func(datum *basal.Basal) {
						datum.Type = "invalid"
						datum.DeliveryType = ""
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "basal"), "/type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deliveryType"),
				),
			)
		})

		Context("IdentityFields", func() {
			var datum *basal.Basal

			BeforeEach(func() {
				datum = dataTypesBasalTest.RandomBasal()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.FromString("")
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if delivery type is empty", func() {
				datum.DeliveryType = ""
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).To(MatchError("delivery type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields(types.IdentityFieldsVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, (*datum.Time).Format(ExpectedTimeFormat), datum.Type, datum.DeliveryType}))
			})
		})
		Context("Legacy IdentityFields", func() {
			var datum *basal.Basal

			BeforeEach(func() {
				datum = dataTypesBasalTest.RandomBasal()
			})

			It("returns error if delivery type is empty", func() {
				datum.DeliveryType = ""
				identityFields, err := datum.IdentityFields(types.LegacyIdentityFieldsVersion)
				Expect(err).To(MatchError("delivery type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected legacy identity fields", func() {
				datum.DeviceID = pointer.FromString("some-device")
				t, err := time.Parse(types.TimeFormat, "2023-05-13T15:51:58Z")
				Expect(err).ToNot(HaveOccurred())
				datum.Time = pointer.FromTime(t)
				datum.DeliveryType = "some-delivery"
				legacyIdentityFields, err := datum.IdentityFields(types.LegacyIdentityFieldsVersion)
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{"basal", "some-delivery", "some-device", "2023-05-13T15:51:58.000Z"}))
			})
		})
	})

	Context("ParseDeliveryType", func() {
		// TODO
	})
})
