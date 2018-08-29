package upload_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	testData "github.com/tidepool-org/platform/data/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/net"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Client", func() {
	Context("ParseClient", func() {
		// TODO
	})

	Context("NewClient", func() {
		It("is successful", func() {
			Expect(upload.NewClient()).To(Equal(&upload.Client{}))
		})
	})

	Context("Client", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *upload.Client), expectedErrors ...error) {
					datum := dataTypesUploadTest.NewClient()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *upload.Client) {},
				),
				Entry("name missing",
					func(datum *upload.Client) { datum.Name = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/name"),
				),
				Entry("name empty",
					func(datum *upload.Client) { datum.Name = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name invalid",
					func(datum *upload.Client) { datum.Name = pointer.FromString("org") },
					testErrors.WithPointerSource(net.ErrorValueStringAsReverseDomainNotValid("org"), "/name"),
				),
				Entry("name valid",
					func(datum *upload.Client) { datum.Name = pointer.FromString(netTest.RandomReverseDomain()) },
				),
				Entry("version missing",
					func(datum *upload.Client) { datum.Version = nil },
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/version"),
				),
				Entry("version empty",
					func(datum *upload.Client) { datum.Version = pointer.FromString("") },
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
				Entry("version invalid",
					func(datum *upload.Client) { datum.Version = pointer.FromString("1.2") },
					testErrors.WithPointerSource(net.ErrorValueStringAsSemanticVersionNotValid("1.2"), "/version"),
				),
				Entry("version valid",
					func(datum *upload.Client) { datum.Version = pointer.FromString(netTest.RandomSemanticVersion()) },
				),
				Entry("private missing",
					func(datum *upload.Client) { datum.Private = nil },
				),
				Entry("private exists",
					func(datum *upload.Client) { datum.Private = testData.NewBlob() },
				),
				Entry("multiple errors",
					func(datum *upload.Client) {
						datum.Name = nil
						datum.Version = nil
					},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/version"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *upload.Client)) {
					for _, origin := range structure.Origins() {
						datum := dataTypesUploadTest.NewClient()
						mutator(datum)
						expectedDatum := dataTypesUploadTest.CloneClient(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *upload.Client) {},
				),
				Entry("does not modify the datum; private missing",
					func(datum *upload.Client) { datum.Private = nil },
				),
				Entry("does not modify the datum; all missing",
					func(datum *upload.Client) { *datum = upload.Client{} },
				),
			)
		})
	})
})
