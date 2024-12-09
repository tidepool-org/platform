package origin_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/origin"
	originTest "github.com/tidepool-org/platform/origin/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Origin", func() {
	It("IDLengthMaximum is expected", func() {
		Expect(origin.IDLengthMaximum).To(Equal(100))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(origin.NameLengthMaximum).To(Equal(100))
	})

	It("TypeApplication is expected", func() {
		Expect(origin.TypeApplication).To(Equal("application"))
	})

	It("TypeDevice is expected", func() {
		Expect(origin.TypeDevice).To(Equal("device"))
	})

	It("TypeManual is expected", func() {
		Expect(origin.TypeManual).To(Equal("manual"))
	})

	It("TypeService is expected", func() {
		Expect(origin.TypeService).To(Equal("service"))
	})

	It("VersionLengthMaximum is expected", func() {
		Expect(origin.VersionLengthMaximum).To(Equal(100))
	})

	It("Types returns expected", func() {
		Expect(origin.Types()).To(Equal([]string{"application", "device", "manual", "service"}))
	})

	Context("Origin", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *origin.Origin)) {
				datum := originTest.RandomOrigin()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, originTest.NewObjectFromOrigin(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, originTest.NewObjectFromOrigin(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *origin.Origin) {},
			),
			Entry("empty",
				func(datum *origin.Origin) {
					*datum = *origin.NewOrigin()
				},
			),
			Entry("all",
				func(datum *origin.Origin) {
					datum.ID = pointer.FromString(originTest.RandomID())
					datum.Name = pointer.FromString(originTest.RandomName())
					datum.Payload = metadataTest.RandomMetadata()
					datum.Time = pointer.FromString(originTest.RandomTime())
					datum.Type = pointer.FromString(originTest.RandomType())
					datum.Version = pointer.FromString(originTest.RandomVersion())
				},
			),
		)

		Context("ParseOrigin", func() {
			It("returns nil when the object is missing", func() {
				Expect(origin.ParseOrigin(structureParser.NewObject(logTest.NewLogger(), nil))).To(BeNil())
			})

			It("returns new datum when the object is valid", func() {
				datum := originTest.RandomOrigin()
				object := originTest.NewObjectFromOrigin(datum, test.ObjectFormatJSON)
				parser := structureParser.NewObject(logTest.NewLogger(), &object)
				Expect(origin.ParseOrigin(parser)).To(Equal(datum))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})
		})

		Context("NewOrigin", func() {
			It("returns successfully with default values", func() {
				Expect(origin.NewOrigin()).To(Equal(&origin.Origin{}))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *origin.Origin), expectedErrors ...error) {
					expectedDatum := originTest.RandomOrigin()
					object := originTest.NewObjectFromOrigin(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &origin.Origin{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *origin.Origin) {},
				),
				Entry("multiple errors",
					func(object map[string]interface{}, expectedDatum *origin.Origin) {
						object["id"] = true
						object["name"] = true
						object["payload"] = true
						object["time"] = true
						object["type"] = true
						object["version"] = true
						expectedDatum.ID = nil
						expectedDatum.Name = nil
						expectedDatum.Payload = nil
						expectedDatum.Time = nil
						expectedDatum.Type = nil
						expectedDatum.Version = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/name"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/payload"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/time"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/type"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/version"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *origin.Origin), expectedErrors ...error) {
					datum := originTest.RandomOrigin()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *origin.Origin) {},
				),
				Entry("id missing",
					func(datum *origin.Origin) { datum.ID = nil },
				),
				Entry("id empty",
					func(datum *origin.Origin) { datum.ID = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id length; in range (upper)",
					func(datum *origin.Origin) { datum.ID = pointer.FromString(test.RandomStringFromRange(100, 100)) },
				),
				Entry("id length; out of range (upper)",
					func(datum *origin.Origin) { datum.ID = pointer.FromString(test.RandomStringFromRange(101, 101)) },
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/id"),
				),
				Entry("name missing",
					func(datum *origin.Origin) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *origin.Origin) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name length; in range (upper)",
					func(datum *origin.Origin) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("name length; out of range (upper)",
					func(datum *origin.Origin) {
						datum.Name = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/name"),
				),
				Entry("payload missing",
					func(datum *origin.Origin) { datum.Payload = nil },
				),
				Entry("payload invalid",
					func(datum *origin.Origin) { datum.Payload = metadata.NewMetadata() },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/payload"),
				),
				Entry("payload valid",
					func(datum *origin.Origin) { datum.Payload = metadataTest.RandomMetadata() },
				),
				Entry("time missing",
					func(datum *origin.Origin) { datum.Time = nil },
				),
				Entry("time empty",
					func(datum *origin.Origin) { datum.Time = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("", time.RFC3339Nano), "/time"),
				),
				Entry("time zero",
					func(datum *origin.Origin) {
						datum.Time = pointer.FromString(time.Time{}.Format(time.RFC3339Nano))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/time"),
				),
				Entry("time not zero",
					func(datum *origin.Origin) {
						datum.Time = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
					},
				),
				Entry("type missing",
					func(datum *origin.Origin) { datum.Type = nil },
				),
				Entry("type invalid",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"application", "device", "manual", "service"}), "/type"),
				),
				Entry("type application",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("application") },
				),
				Entry("type device",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("device") },
				),
				Entry("type manual",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("manual") },
				),
				Entry("type service",
					func(datum *origin.Origin) { datum.Type = pointer.FromString("service") },
				),
				Entry("version missing",
					func(datum *origin.Origin) { datum.Version = nil },
				),
				Entry("version empty",
					func(datum *origin.Origin) { datum.Version = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
				Entry("version length; in range (upper)",
					func(datum *origin.Origin) {
						datum.Version = pointer.FromString(test.RandomStringFromRange(100, 100))
					},
				),
				Entry("version length; out of range (upper)",
					func(datum *origin.Origin) {
						datum.Version = pointer.FromString(test.RandomStringFromRange(101, 101))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/version"),
				),
				Entry("multiple errors",
					func(datum *origin.Origin) {
						datum.ID = pointer.FromString("")
						datum.Name = pointer.FromString("")
						datum.Time = pointer.FromString("")
						datum.Type = pointer.FromString("invalid")
						datum.Version = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("", time.RFC3339Nano), "/time"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"application", "device", "manual", "service"}), "/type"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
			)
		})
	})
})
