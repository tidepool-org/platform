package selfmonitored_test

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataBloodGlucoseTest "github.com/tidepool-org/platform/data/blood/glucose/test"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "smbg",
	}
}

func NewSelfMonitored(units *string) *selfmonitored.SelfMonitored {
	datum := selfmonitored.New()
	datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
	datum.Type = "smbg"
	datum.SubType = pointer.FromString(test.RandomStringFromArray(selfmonitored.SubTypes()))
	return datum
}

func CloneSelfMonitored(datum *selfmonitored.SelfMonitored) *selfmonitored.SelfMonitored {
	if datum == nil {
		return nil
	}
	clone := selfmonitored.New()
	clone.Glucose = *dataTypesBloodGlucoseTest.CloneGlucose(&datum.Glucose)
	clone.SubType = pointer.CloneString(datum.SubType)
	return clone
}

var _ = Describe("SelfMonitored", func() {
	It("Type is expected", func() {
		Expect(selfmonitored.Type).To(Equal("smbg"))
	})

	It("SubTypeLinked is expected", func() {
		Expect(selfmonitored.SubTypeLinked).To(Equal("linked"))
	})

	It("SubTypeManual is expected", func() {
		Expect(selfmonitored.SubTypeManual).To(Equal("manual"))
	})

	It("SubTypeScanned is expected", func() {
		Expect(selfmonitored.SubTypeScanned).To(Equal("scanned"))
	})

	It("SubTypes returns expected", func() {
		Expect(selfmonitored.SubTypes()).To(Equal([]string{"linked", "manual", "scanned"}))
	})

	Context("New", func() {
		It("returns the expected datum", func() {
			datum := selfmonitored.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("smbg"))
			Expect(datum.Units).To(BeNil())
			Expect(datum.Value).To(BeNil())
			Expect(datum.SubType).To(BeNil())
		})
	})

	Context("SelfMonitored", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectedErrors ...error) {
					datum := NewSelfMonitored(units)
					mutator(datum, units)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
				),
				Entry("type missing",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "smbg"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type smbg",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Type = "smbg" },
				),
				Entry("units missing; value missing",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units missing; value out of range (lower)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (lower)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value in range (upper)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(55.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units missing; value out of range (upper)",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", NewMeta()),
				),
				Entry("units invalid; value missing",
					pointer.FromString("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units invalid; value out of range (lower)",
					pointer.FromString("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (lower)",
					pointer.FromString("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(0.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value in range (upper)",
					pointer.FromString("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(55.0) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units invalid; value out of range (upper)",
					pointer.FromString("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"mmol/L", "mmol/l", "mg/dL", "mg/dl"}), "/units", NewMeta()),
				),
				Entry("units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/L; value out of range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/L; value in range (lower)",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/L; value in range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/L; value out of range (upper)",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(55.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mmol/l; value out of range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mmol/l; value in range (lower)",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mmol/l; value in range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(55.0) },
				),
				Entry("units mmol/l; value out of range (upper)",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(55.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(55.1, 0.0, 55.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value missing",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dL; value out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dL; value in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dL; value in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(1000.0) },
				),
				Entry("units mg/dL; value out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value missing",
					pointer.FromString("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", NewMeta()),
				),
				Entry("units mg/dl; value out of range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(-0.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(-0.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("units mg/dl; value in range (lower)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(0.0) },
				),
				Entry("units mg/dl; value in range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(1000.0) },
				),
				Entry("units mg/dl; value out of range (upper)",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = pointer.FromFloat64(1000.1) },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotInRange(1000.1, 0.0, 1000.0), "/value", NewMeta()),
				),
				Entry("sub type missing",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = nil },
				),
				Entry("sub type invalid",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("invalid") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"linked", "manual", "scanned"}), "/subType", NewMeta()),
				),
				Entry("sub type linked",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("linked") },
				),
				Entry("sub type manual",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("manual") },
				),
				Entry("sub type scanned",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("scanned") },
				),
				Entry("multiple errors",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) {
						datum.Type = ""
						datum.Value = nil
						datum.SubType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/units", &types.Meta{}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/value", &types.Meta{}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"linked", "manual", "scanned"}), "/subType", &types.Meta{}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectator func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string)) {
					for _, origin := range structure.Origins() {
						datum := NewSelfMonitored(units)
						mutator(datum, units)
						expectedDatum := CloneSelfMonitored(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())

						expectedDatum.RawValue = pointer.CloneFloat64(datum.RawValue)
						expectedDatum.RawUnits = pointer.CloneString(datum.RawUnits)
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units missing",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units missing; value missing",
					nil,
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units invalid",
					pointer.FromString("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units invalid; value missing",
					pointer.FromString("invalid"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; sub type missing",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = nil },
					nil,
				),
				Entry("does not modify the datum; sub type invalid",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("invalid") },
					nil,
				),
				Entry("does not modify the datum; sub type linked",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("linked") },
					nil,
				),
				Entry("does not modify the datum; sub type manual",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("manual") },
					nil,
				),
				Entry("does not modify the datum; sub type scanned",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.SubType = pointer.FromString("scanned") },
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectator func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string)) {
					datum := NewSelfMonitored(units)
					mutator(datum, units)
					expectedDatum := CloneSelfMonitored(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					expectedDatum.RawValue = pointer.CloneFloat64(datum.RawValue)
					expectedDatum.RawUnits = pointer.CloneString(datum.RawUnits)
					if expectator != nil {
						expectator(datum, expectedDatum, units)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("modifies the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dL; value missing",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
				Entry("modifies the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
						dataBloodGlucoseTest.ExpectNormalizedValue(datum.Value, expectedDatum.Value, units)
					},
				),
				Entry("modifies the datum; units mg/dl; value missing",
					pointer.FromString("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string) {
						dataBloodGlucoseTest.ExpectNormalizedUnits(datum.Units, expectedDatum.Units)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(units *string, mutator func(datum *selfmonitored.SelfMonitored, units *string), expectator func(datum *selfmonitored.SelfMonitored, expectedDatum *selfmonitored.SelfMonitored, units *string)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewSelfMonitored(units)
						mutator(datum, units)
						expectedDatum := CloneSelfMonitored(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						expectedDatum.RawValue = pointer.CloneFloat64(datum.RawValue)
						expectedDatum.RawUnits = pointer.CloneString(datum.RawUnits)
						if expectator != nil {
							expectator(datum, expectedDatum, units)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum; units mmol/L",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/L; value missing",
					pointer.FromString("mmol/L"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mmol/l",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mmol/l; value missing",
					pointer.FromString("mmol/l"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dL",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dL; value missing",
					pointer.FromString("mg/dL"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
				Entry("does not modify the datum; units mg/dl",
					pointer.FromString("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) {},
					nil,
				),
				Entry("does not modify the datum; units mg/dl; value missing",
					pointer.FromString("mg/dl"),
					func(datum *selfmonitored.SelfMonitored, units *string) { datum.Value = nil },
					nil,
				),
			)
		})

		Context("LegacyIdentityFields", func() {
			var datum *selfmonitored.SelfMonitored

			BeforeEach(func() {
				datum = NewSelfMonitored(pointer.FromString("mmol/l"))
			})

			It("returns error if device id is missing", func() {
				datum.DeviceID = nil
				identityFields, err := datum.LegacyIdentityFields()
				Expect(err).To(MatchError("device id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if device id is empty", func() {
				datum.DeviceID = pointer.FromString("")
				identityFields, err := datum.LegacyIdentityFields()
				Expect(err).To(MatchError("device id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if time is missing", func() {
				datum.Time = nil
				identityFields, err := datum.LegacyIdentityFields()
				Expect(err).To(MatchError("time is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if time is empty", func() {
				datum.Time = &time.Time{}
				identityFields, err := datum.LegacyIdentityFields()
				Expect(err).To(MatchError("time is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if value is missing", func() {
				datum.Value = nil
				identityFields, err := datum.LegacyIdentityFields()
				Expect(err).To(MatchError("value is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected legacy identity fields", func() {
				legacyIdentityFields, err := datum.LegacyIdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(legacyIdentityFields).To(Equal([]string{datum.Type, *datum.DeviceID, (*datum.Time).Format(types.LegacyFieldTimeFormat), strconv.FormatFloat(*datum.Value, 'f', -1, 64)}))
			})
		})
	})
})
