package data_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/net"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Deduplicator", func() {
	Context("DeduplicatorDescriptor", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *data.DeduplicatorDescriptor)) {
				datum := dataTest.RandomDeduplicatorDescriptor()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataTest.NewObjectFromDeduplicatorDescriptor(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataTest.NewObjectFromDeduplicatorDescriptor(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *data.DeduplicatorDescriptor) {},
			),
			Entry("empty",
				func(datum *data.DeduplicatorDescriptor) { *datum = data.DeduplicatorDescriptor{} },
			),
		)

		Context("ParseDeduplicatorDescriptor", func() {
			// TODO
		})

		Context("ParseDeduplicatorDescriptorDEPRECATED", func() {
			// TODO
		})

		Context("NewDeduplicatorDescriptor", func() {
			It("returns successfully with default values", func() {
				Expect(data.NewDeduplicatorDescriptor()).To(Equal(&data.DeduplicatorDescriptor{}))
			})
		})

		Context("Parse", func() {
			// TODO
		})

		Context("ParseDEPRECATED", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *data.DeduplicatorDescriptor), expectedErrors ...error) {
					datum := dataTest.RandomDeduplicatorDescriptor()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *data.DeduplicatorDescriptor) {},
				),
				Entry("name missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Name = nil },
				),
				Entry("name empty",
					func(datum *data.DeduplicatorDescriptor) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name invalid",
					func(datum *data.DeduplicatorDescriptor) { datum.Name = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsReverseDomainNotValid("invalid"), "/name"),
				),
				Entry("name valid",
					func(datum *data.DeduplicatorDescriptor) {
						datum.Name = pointer.FromString(netTest.RandomReverseDomain())
					},
				),
				Entry("version missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Version = nil },
				),
				Entry("version empty",
					func(datum *data.DeduplicatorDescriptor) { datum.Version = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
				Entry("version invalid",
					func(datum *data.DeduplicatorDescriptor) { datum.Version = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsSemanticVersionNotValid("invalid"), "/version"),
				),
				Entry("version valid",
					func(datum *data.DeduplicatorDescriptor) {
						datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
					},
				),
				Entry("hash missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Hash = nil },
				),
				Entry("hash empty",
					func(datum *data.DeduplicatorDescriptor) { datum.Hash = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/hash"),
				),
				Entry("hash valid",
					func(datum *data.DeduplicatorDescriptor) { datum.Hash = pointer.FromString(test.RandomString()) },
				),
				Entry("multiple errors",
					func(datum *data.DeduplicatorDescriptor) {
						datum.Name = pointer.FromString("")
						datum.Version = pointer.FromString("")
						datum.Hash = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/hash"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *data.DeduplicatorDescriptor), expectator func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor)) {
					datum := dataTest.RandomDeduplicatorDescriptor()
					mutator(datum)
					expectedDatum := dataTest.CloneDeduplicatorDescriptor(datum)
					Expect(structureNormalizer.New(logTest.NewLogger()).Normalize(datum)).ToNot(HaveOccurred())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum",
					func(datum *data.DeduplicatorDescriptor) {},
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
				Entry("does not modify the datum; name missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Name = nil },
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
				Entry("does not modify the datum; version missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Version = nil },
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
				Entry("does not modify the datum; hash missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Hash = nil },
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
			)
		})

		Context("NormalizeDEPRECATED", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *data.DeduplicatorDescriptor), expectator func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor)) {
					datum := dataTest.RandomDeduplicatorDescriptor()
					mutator(datum)
					expectedDatum := dataTest.CloneDeduplicatorDescriptor(datum)
					normalizer := dataNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					datum.NormalizeDEPRECATED(normalizer)
					Expect(normalizer.Error()).ToNot(HaveOccurred())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum",
					func(datum *data.DeduplicatorDescriptor) {},
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
				Entry("does not modify the datum; name missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Name = nil },
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
				Entry("does not modify the datum; version missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Version = nil },
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
				Entry("does not modify the datum; hash missing",
					func(datum *data.DeduplicatorDescriptor) { datum.Hash = nil },
					func(datum *data.DeduplicatorDescriptor, expectedDatum *data.DeduplicatorDescriptor) {},
				),
			)
		})

		Context("with new deduplicator descriptor", func() {
			var datum *data.DeduplicatorDescriptor

			BeforeEach(func() {
				datum = dataTest.RandomDeduplicatorDescriptor()
			})

			Context("HasName", func() {
				It("return false if the name is missing", func() {
					datum.Name = nil
					Expect(datum.HasName()).To(BeFalse())
				})

				It("returns true if the name is empty", func() {
					datum.Name = pointer.FromString("")
					Expect(datum.HasName()).To(BeTrue())
				})

				It("returns true if the name exists", func() {
					datum.Name = pointer.FromString(netTest.RandomReverseDomain())
					Expect(datum.HasName()).To(BeTrue())
				})
			})

			Context("HasNameMatch", func() {
				var name string

				BeforeEach(func() {
					name = netTest.RandomReverseDomain()
				})

				It("return false if the name is missing", func() {
					datum.Name = nil
					Expect(datum.HasNameMatch(name)).To(BeFalse())
				})

				It("return false if the name is empty", func() {
					datum.Name = pointer.FromString("")
					Expect(datum.HasNameMatch(name)).To(BeFalse())
				})

				It("returns false if the name does not match", func() {
					datum.Name = pointer.FromString(netTest.RandomReverseDomain())
					Expect(datum.HasNameMatch(name)).To(BeFalse())
				})

				It("returns true if the name matches", func() {
					datum.Name = pointer.FromString(name)
					Expect(datum.HasNameMatch(name)).To(BeTrue())
				})
			})
		})
	})
})
