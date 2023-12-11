package structured_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreStructuredTest "github.com/tidepool-org/platform/blob/store/structured/test"
	"github.com/tidepool-org/platform/crypto"
	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/net"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Structured", func() {
	Context("NewCreate", func() {
		It("returns successfully with default values", func() {
			Expect(blobStoreStructured.NewCreate()).To(Equal(&blobStoreStructured.Create{}))
		})
	})

	Context("with new create", func() {
		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *blobStoreStructured.Create), expectedErrors ...error) {
					datum := blobStoreStructuredTest.RandomCreate()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *blobStoreStructured.Create) {},
				),
				Entry("media type missing",
					func(datum *blobStoreStructured.Create) { datum.MediaType = nil },
				),
				Entry("media type empty",
					func(datum *blobStoreStructured.Create) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *blobStoreStructured.Create) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *blobStoreStructured.Create) {
						datum.MediaType = pointer.FromString(netTest.RandomMediaType())
					},
				),
			)
		})
	})

	Context("NewUpdate", func() {
		It("returns successfully with default values", func() {
			Expect(blobStoreStructured.NewUpdate()).To(Equal(&blobStoreStructured.Update{}))
		})
	})

	Context("with new update", func() {
		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *blobStoreStructured.Update), expectedErrors ...error) {
					datum := blobStoreStructuredTest.RandomUpdate()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *blobStoreStructured.Update) {},
				),
				Entry("digest MD5 missing",
					func(datum *blobStoreStructured.Update) { datum.DigestMD5 = nil },
				),
				Entry("digest MD5 empty",
					func(datum *blobStoreStructured.Update) { datum.DigestMD5 = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
				),
				Entry("digest MD5 invalid",
					func(datum *blobStoreStructured.Update) { datum.DigestMD5 = pointer.FromString("#") },
					errorsTest.WithPointerSource(crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("#"), "/digestMD5"),
				),
				Entry("digest MD5 valid",
					func(datum *blobStoreStructured.Update) {
						datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
					},
				),
				Entry("media type missing",
					func(datum *blobStoreStructured.Update) { datum.MediaType = nil },
				),
				Entry("media type empty",
					func(datum *blobStoreStructured.Update) { datum.MediaType = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
				),
				Entry("media type invalid",
					func(datum *blobStoreStructured.Update) { datum.MediaType = pointer.FromString("/") },
					errorsTest.WithPointerSource(net.ErrorValueStringAsMediaTypeNotValid("/"), "/mediaType"),
				),
				Entry("media type valid",
					func(datum *blobStoreStructured.Update) {
						datum.MediaType = pointer.FromString(netTest.RandomMediaType())
					},
				),
				Entry("size missing",
					func(datum *blobStoreStructured.Update) { datum.Size = nil },
				),
				Entry("size out of range (lower)",
					func(datum *blobStoreStructured.Update) { datum.Size = pointer.FromInt(-1) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/size"),
				),
				Entry("size in range (lower)",
					func(datum *blobStoreStructured.Update) { datum.Size = pointer.FromInt(0) },
				),
				Entry("status missing",
					func(datum *blobStoreStructured.Update) { datum.Status = nil },
				),
				Entry("status empty",
					func(datum *blobStoreStructured.Update) { datum.Status = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", blob.Statuses()), "/status"),
				),
				Entry("status invalid",
					func(datum *blobStoreStructured.Update) { datum.Status = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", blob.Statuses()), "/status"),
				),
				Entry("status created",
					func(datum *blobStoreStructured.Update) { datum.Status = pointer.FromString("created") },
				),
				Entry("status available",
					func(datum *blobStoreStructured.Update) { datum.Status = pointer.FromString("available") },
				),
				Entry("multiple errors",
					func(datum *blobStoreStructured.Update) {
						datum.DigestMD5 = pointer.FromString("")
						datum.MediaType = pointer.FromString("")
						datum.Size = pointer.FromInt(-1)
						datum.Status = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/digestMD5"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/size"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", blob.Statuses()), "/status"),
				),
			)

			Context("IsEmpty", func() {
				var datum *blobStoreStructured.Update

				BeforeEach(func() {
					datum = blobStoreStructured.NewUpdate()
				})

				It("returns true when no fields are specified", func() {
					Expect(datum.IsEmpty()).To(BeTrue())
				})

				It("returns false when the digest MD5 field is specified", func() {
					datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
					Expect(datum.IsEmpty()).To(BeFalse())
				})

				It("returns false when the media type field is specified", func() {
					datum.MediaType = pointer.FromString(netTest.RandomMediaType())
					Expect(datum.IsEmpty()).To(BeFalse())
				})

				It("returns false when the size field is specified", func() {
					datum.Size = pointer.FromInt(test.RandomIntFromRange(1, 100*1024*1024))
					Expect(datum.IsEmpty()).To(BeFalse())
				})

				It("returns false when the status field is specified", func() {
					datum.Status = pointer.FromString(test.RandomStringFromArray(blob.Statuses()))
					Expect(datum.IsEmpty()).To(BeFalse())
				})

				It("returns false when multiple fields are specified", func() {
					datum.DigestMD5 = pointer.FromString(cryptoTest.RandomBase64EncodedMD5Hash())
					datum.MediaType = pointer.FromString(netTest.RandomMediaType())
					datum.Size = pointer.FromInt(test.RandomIntFromRange(1, 100*1024*1024))
					datum.Status = pointer.FromString(test.RandomStringFromArray(blob.Statuses()))
					Expect(datum.IsEmpty()).To(BeFalse())
				})
			})
		})
	})
})
