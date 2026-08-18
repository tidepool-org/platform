package postprocess_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
	userWork "github.com/tidepool-org/platform/user/work"
)

var _ = Describe("Work", func() {
	It("Type is expected", func() {
		Expect(dataWorkPostprocess.Type).To(Equal("org.tidepool.data.upload.postprocess"))
	})

	It("ProcessingTimeout is expected", func() {
		Expect(dataWorkPostprocess.ProcessingTimeout).To(Equal(5 * time.Minute))
	})

	It("Reasons is expected", func() {
		Expect(dataWorkPostprocess.Reasons()).To(ConsistOf(
			"DATA_ADDED",
			"UPLOAD_COMPLETED",
			"LEGACY_DATA_ADDED",
			"SCHEMA_MIGRATION",
		))
	})

	Context("TriggersEHRSync", func() {
		DescribeTable("reports whether the reasons trigger a synchronization",
			func(reasons []string, expected bool) {
				Expect(dataWorkPostprocess.TriggersEHRSync(reasons)).To(Equal(expected))
			},
			Entry("with no reasons", nil, false),
			Entry("with data added", []string{dataWorkPostprocess.ReasonDataAdded}, false),
			Entry("with schema migration", []string{dataWorkPostprocess.ReasonSchemaMigration}, false),
			Entry("with upload completed", []string{dataWorkPostprocess.ReasonUploadCompleted}, true),
			Entry("with legacy data added", []string{dataWorkPostprocess.ReasonLegacyDataAdded}, true),
			Entry("with any reason triggering a synchronization",
				[]string{dataWorkPostprocess.ReasonDataAdded, dataWorkPostprocess.ReasonUploadCompleted}, true),
		)
	})

	Context("IDFromUserID", func() {
		var userID string

		BeforeEach(func() {
			userID = userTest.RandomUserID()
		})

		It("returns the identifier of the work for the user", func() {
			Expect(dataWorkPostprocess.IDFromUserID(userID)).To(Equal("org.tidepool.data.upload.postprocess:" + userID))
		})

		// The work is scoped to the user rather than to any data set, so that changes to any number
		// of data sets are coalesced into a single stream of work for the user
		It("returns the same identifier for the user however many times it is called", func() {
			Expect(dataWorkPostprocess.IDFromUserID(userID)).To(Equal(dataWorkPostprocess.IDFromUserID(userID)))
		})

		It("returns a different identifier for a different user", func() {
			Expect(dataWorkPostprocess.IDFromUserID(userID)).ToNot(Equal(dataWorkPostprocess.IDFromUserID(userTest.RandomUserID())))
		})
	})

	Context("Metadata", func() {
		var userID string

		BeforeEach(func() {
			userID = userTest.RandomUserID()
		})

		Context("Parse", func() {
			DescribeTable("parses the metadata",
				func(object map[string]any, expectedMetadata func(userID string) *dataWorkPostprocess.Metadata, expectedErrors ...error) {
					if userIDValue, ok := object["userId"]; ok && userIDValue == nil {
						object["userId"] = userID
					}
					result := &dataWorkPostprocess.Metadata{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
					Expect(result).To(Equal(expectedMetadata(userID)))
				},
				Entry("with a user and a reason",
					map[string]any{"userId": nil, "reasons": []any{"DATA_ADDED"}},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
							Reasons:  []string{"DATA_ADDED"},
						}
					},
				),
				Entry("with multiple reasons",
					map[string]any{"userId": nil, "reasons": []any{"LEGACY_DATA_ADDED", "UPLOAD_COMPLETED"}},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
							Reasons:  []string{"LEGACY_DATA_ADDED", "UPLOAD_COMPLETED"},
						}
					},
				),
				Entry("with nothing",
					map[string]any{},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{}
					},
				),
				Entry("with reasons of the wrong type",
					map[string]any{"userId": nil, "reasons": true},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
						}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/reasons"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the metadata",
				func(mutator func(metadata *dataWorkPostprocess.Metadata), expectedErrors ...error) {
					datum := &dataWorkPostprocess.Metadata{
						Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
						Reasons:  []string{dataWorkPostprocess.ReasonDataAdded},
					}
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds", func(datum *dataWorkPostprocess.Metadata) {}),
				Entry("succeeds with every reason", func(datum *dataWorkPostprocess.Metadata) {
					datum.Reasons = dataWorkPostprocess.Reasons()
				}),
				Entry("reports the user is not valid",
					func(datum *dataWorkPostprocess.Metadata) { datum.UserID = pointer.FromString("invalid") },
					errorsTest.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/userId"),
				),
				Entry("reports the reasons are missing",
					func(datum *dataWorkPostprocess.Metadata) { datum.Reasons = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/reasons"),
				),
				Entry("reports the reasons are empty",
					func(datum *dataWorkPostprocess.Metadata) { datum.Reasons = []string{} },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/reasons"),
				),
				Entry("reports a reason is not valid",
					func(datum *dataWorkPostprocess.Metadata) { datum.Reasons = []string{"INVALID"} },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("INVALID", dataWorkPostprocess.Reasons()), "/reasons/0"),
				),
				Entry("reports the reasons are duplicated",
					func(datum *dataWorkPostprocess.Metadata) {
						datum.Reasons = []string{dataWorkPostprocess.ReasonDataAdded, dataWorkPostprocess.ReasonDataAdded}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/reasons/1"),
				),
			)
		})

		// The metadata is encoded when the work is created and decoded when it is processed
		It("is unchanged by being encoded and decoded", func() {
			datum := &dataWorkPostprocess.Metadata{
				Metadata: userWork.Metadata{UserID: pointer.FromString(userID)},
				Reasons:  []string{dataWorkPostprocess.ReasonLegacyDataAdded, dataWorkPostprocess.ReasonUploadCompleted},
			}

			encoded, err := metadata.Encode(datum)
			Expect(err).ToNot(HaveOccurred())
			Expect(encoded).ToNot(BeEmpty())

			decoded, err := metadata.Decode[dataWorkPostprocess.Metadata](context.Background(), encoded)
			Expect(err).ToNot(HaveOccurred())
			Expect(decoded).To(Equal(datum))
		})
	})
})
