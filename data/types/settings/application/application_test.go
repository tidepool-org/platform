package application_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesApplication "github.com/tidepool-org/platform/data/types/settings/application"
	dataTypesApplicationTest "github.com/tidepool-org/platform/data/types/settings/application/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "applicationSettings",
	}
}

var _ = Describe("Application", func() {
	It("Type is expected", func() {
		Expect(dataTypesApplication.Type).To(Equal("applicationSettings"))
	})

	It("NameLengthMaximum is expected", func() {
		Expect(dataTypesApplication.NameLengthMaximum).To(Equal(1000))
	})

	It("NameLengthMinimum is expected", func() {
		Expect(dataTypesApplication.NameLengthMinimum).To(Equal(1))
	})

	It("VersionLengthMaximum is expected", func() {
		Expect(dataTypesApplication.VersionLengthMaximum).To(Equal(1000))
	})

	It("VersionLengthMinimum is expected", func() {
		Expect(dataTypesApplication.VersionLengthMinimum).To(Equal(1))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := dataTypesApplication.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("applicationSettings"))
			Expect(datum.Name).To(BeNil())
			Expect(datum.Version).To(BeNil())
		})
	})

	Context("Application", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("return the expected results when the input",
				func(mutator func(datum *dataTypesApplication.Application), expectedErrors ...error) {
					datum := dataTypesApplicationTest.RandomApplication()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesApplication.Application) {},
				),
				Entry("type missing",
					func(datum *dataTypesApplication.Application) { datum.Type = "" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					func(datum *dataTypesApplication.Application) { datum.Type = "invalidType" },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "applicationSettings"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type applicationSettings",
					func(datum *dataTypesApplication.Application) { datum.Type = "applicationSettings" },
				),
				Entry("name missing",
					func(datum *dataTypesApplication.Application) { datum.Name = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/name", NewMeta()),
				),
				Entry("name invalid",
					func(datum *dataTypesApplication.Application) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotInRange(0, dataTypesApplication.NameLengthMinimum, dataTypesApplication.NameLengthMaximum), "/name", NewMeta()),
				),
				Entry("version missing",
					func(datum *dataTypesApplication.Application) { datum.Version = nil },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/version", NewMeta()),
				),
				Entry("version invalid",
					func(datum *dataTypesApplication.Application) { datum.Version = pointer.FromString("") },
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotInRange(0, dataTypesApplication.NameLengthMinimum, dataTypesApplication.NameLengthMaximum), "/version", NewMeta()),
				),
				Entry("multiple errors",
					func(datum *dataTypesApplication.Application) {
						datum.Name = nil
						datum.Version = nil
					},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/name", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/version", NewMeta()),
				),
			)
		})
	})
})
