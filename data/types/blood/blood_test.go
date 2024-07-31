package blood_test

import (
	"math"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood"
	dataTypesBloodTest "github.com/tidepool-org/platform/data/types/blood/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

const ExpectedTimeFormat = time.RFC3339Nano

var _ = Describe("Blood", func() {
	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			typ := dataTypesTest.NewType()
			datum := blood.New(typ)
			Expect(datum.Type).To(Equal(typ))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
		})
	})

	Context("with new datum", func() {
		var typ string
		var datum blood.Blood

		BeforeEach(func() {
			typ = dataTypesTest.NewType()
			datum = blood.New(typ)
		})

		Context("Meta", func() {
			It("returns the meta with delivery type", func() {
				Expect(datum.Meta()).To(Equal(&types.Meta{Type: typ}))
			})
		})
	})

	Context("Blood", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *blood.Blood), expectedErrors ...error) {
					datum := dataTypesBloodTest.NewBlood()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *blood.Blood) {},
				),
				Entry("type missing",
					func(datum *blood.Blood) { datum.Type = "" },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type exists",
					func(datum *blood.Blood) { datum.Type = dataTypesTest.NewType() },
				),
				Entry("units missing",
					func(datum *blood.Blood) { datum.Units = nil },
				),
				Entry("units exists",
					func(datum *blood.Blood) { datum.Units = pointer.FromString(dataTypesTest.NewType()) },
				),
				Entry("value missing",
					func(datum *blood.Blood) { datum.Value = nil },
				),
				Entry("value exists",
					func(datum *blood.Blood) {
						datum.Value = pointer.FromFloat64(test.RandomFloat64FromRange(-math.MaxFloat64, math.MaxFloat64))
					},
				),
			)
		})

		Context("IdentityFields", func() {
			var datum *blood.Blood

			BeforeEach(func() {
				datum = dataTypesBloodTest.NewBlood()
			})

			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.FromString("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if units is missing", func() {
				datum.Units = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("units is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if value is missing", func() {
				datum.Value = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("value is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, (*datum.Time).Format(ExpectedTimeFormat), datum.Type, *datum.Units, strconv.FormatFloat(*datum.Value, 'f', -1, 64)}))
			})
		})

		Context("LegacyIdentityFields", func() {
			It("returns the expected legacy identity fields", func() {
				datum := dataTypesBloodTest.NewBlood()
				datum.Type = "bg"
				datum.DeviceID = pointer.FromString("some-bg-device")
				t, err := time.Parse(types.TimeFormat, "2023-05-13T15:51:58Z")
				Expect(err).ToNot(HaveOccurred())
				datum.Time = pointer.FromTime(t)
				legacyIdentityFields, err := datum.LegacyIdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{"bg", "some-bg-device", "2023-05-13T15:51:58Z"}))
			})
		})

	})
})
