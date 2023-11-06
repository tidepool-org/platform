package upload_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/metadata"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
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
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *upload.Client) {},
				),
				Entry("name missing",
					func(datum *upload.Client) { datum.Name = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/name"),
				),
				Entry("name empty",
					func(datum *upload.Client) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name invalid",
					func(datum *upload.Client) { datum.Name = pointer.FromString("org") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsReverseDomainNotValid("org"), "/name"),
				),
				Entry("name valid",
					func(datum *upload.Client) { datum.Name = pointer.FromString(netTest.RandomReverseDomain()) },
				),
				Entry("version missing",
					func(datum *upload.Client) { datum.Version = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/version"),
				),
				Entry("version empty",
					func(datum *upload.Client) { datum.Version = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
				Entry("version invalid",
					func(datum *upload.Client) { datum.Version = pointer.FromString("1.2") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsSemanticVersionNotValid("1.2"), "/version"),
				),
				Entry("version valid",
					func(datum *upload.Client) { datum.Version = pointer.FromString(netTest.RandomSemanticVersion()) },
				),
				Entry("private missing",
					func(datum *upload.Client) { datum.Private = nil },
				),
				Entry("private invalid",
					func(datum *upload.Client) { datum.Private = metadata.NewMetadata() },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/private"),
				),
				Entry("private valid",
					func(datum *upload.Client) { datum.Private = metadataTest.RandomMetadata() },
				),
				Entry("multiple errors",
					func(datum *upload.Client) {
						datum.Name = nil
						datum.Version = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/version"),
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
