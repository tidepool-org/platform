package postprocess_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataTypes "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/food"
	dataWorkPostprocess "github.com/tidepool-org/platform/data/work/postprocess"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	summaryTypes "github.com/tidepool-org/platform/summary/types"
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

	Context("SummaryTypesForData", func() {
		newDatum := func(typ string, tm time.Time, active bool) data.Datum {
			datum := dataTypes.New(typ)
			datum.Time = pointer.FromTime(tm)
			datum.Active = active
			return &datum
		}

		DescribeTable("reports the summaries the data affects",
			func(datumArray func() data.Data, expected []string) {
				Expect(dataWorkPostprocess.SummaryTypesForData(datumArray())).To(ConsistOf(expected))
			},
			Entry("with no data", func() data.Data { return nil }, []string{}),
			Entry("with a nil datum", func() data.Data { return data.Data{nil} }, []string{}),
			Entry("with a type no summary is calculated from",
				func() data.Data { return data.Data{newDatum(food.Type, time.Now(), true)} }, []string{}),
			Entry("with a continuous glucose datum",
				func() data.Data { return data.Data{newDatum(continuous.Type, time.Now(), true)} },
				[]string{summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeContinuous}),
			Entry("with a self monitored glucose datum",
				func() data.Data { return data.Data{newDatum(selfmonitored.Type, time.Now(), true)} },
				[]string{summaryTypes.SummaryTypeBGM, summaryTypes.SummaryTypeContinuous}),
			Entry("with both glucose types, each once",
				func() data.Data {
					return data.Data{
						newDatum(continuous.Type, time.Now(), true),
						newDatum(selfmonitored.Type, time.Now(), true),
						newDatum(continuous.Type, time.Now().Add(-time.Hour), true),
					}
				},
				[]string{summaryTypes.SummaryTypeBGM, summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeContinuous}),
			// Data the deduplicator deactivates until the data set is closed affects nothing until then
			Entry("with an inactive glucose datum",
				func() data.Data { return data.Data{newDatum(continuous.Type, time.Now(), false)} }, []string{}),
			Entry("with a glucose datum without a time",
				func() data.Data {
					datum := dataTypes.New(continuous.Type)
					datum.Active = true
					return data.Data{&datum}
				}, []string{}),
			Entry("with a glucose datum older than two years",
				func() data.Data { return data.Data{newDatum(continuous.Type, time.Now().AddDate(0, -24, -1), true)} }, []string{}),
			Entry("with a glucose datum just under two years old",
				func() data.Data { return data.Data{newDatum(continuous.Type, time.Now().AddDate(0, -24, 1), true)} },
				[]string{summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeContinuous}),
			Entry("with a glucose datum more than a day in the future",
				func() data.Data { return data.Data{newDatum(continuous.Type, time.Now().Add(25*time.Hour), true)} }, []string{}),
			Entry("with a glucose datum just under a day in the future",
				func() data.Data { return data.Data{newDatum(continuous.Type, time.Now().Add(23*time.Hour), true)} },
				[]string{summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeContinuous}),
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
				Entry("with a user, a reason and summary types",
					map[string]any{"userId": nil, "reasons": []any{"DATA_ADDED"}, "summaryTypes": []any{"cgm", "con"}},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata:     userWork.Metadata{UserID: pointer.FromString(userID)},
							Reasons:      []string{"DATA_ADDED"},
							SummaryTypes: []string{"cgm", "con"},
						}
					},
				),
				Entry("with multiple reasons",
					map[string]any{"userId": nil, "reasons": []any{"LEGACY_DATA_ADDED", "UPLOAD_COMPLETED"}, "summaryTypes": []any{"bgm"}},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata:     userWork.Metadata{UserID: pointer.FromString(userID)},
							Reasons:      []string{"LEGACY_DATA_ADDED", "UPLOAD_COMPLETED"},
							SummaryTypes: []string{"bgm"},
						}
					},
				),
				// Work created before the summary types were reported, and the work of the legacy
				// ingestion service, report none and affect every summary
				Entry("without summary types",
					map[string]any{"userId": nil, "reasons": []any{"LEGACY_DATA_ADDED"}},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata:     userWork.Metadata{UserID: pointer.FromString(userID)},
							Reasons:      []string{"LEGACY_DATA_ADDED"},
							SummaryTypes: []string{"cgm", "bgm", "con"},
						}
					},
				),
				Entry("with nothing",
					map[string]any{},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{SummaryTypes: []string{"cgm", "bgm", "con"}}
					},
				),
				Entry("with reasons of the wrong type",
					map[string]any{"userId": nil, "reasons": true, "summaryTypes": []any{"cgm"}},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata:     userWork.Metadata{UserID: pointer.FromString(userID)},
							SummaryTypes: []string{"cgm"},
						}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/reasons"),
				),
				Entry("with summary types of the wrong type",
					map[string]any{"userId": nil, "reasons": []any{"DATA_ADDED"}, "summaryTypes": "cgm"},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata:     userWork.Metadata{UserID: pointer.FromString(userID)},
							Reasons:      []string{"DATA_ADDED"},
							SummaryTypes: []string{"cgm", "bgm", "con"},
						}
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray("cgm"), "/summaryTypes"),
				),
				Entry("with pending summary changes",
					map[string]any{
						"userId":                nil,
						"reasons":               []any{"DATA_ADDED"},
						"summaryTypes":          []any{"cgm"},
						"pendingSummaryUpdates": []any{"cgm"},
						"pendingSummaryDeletes": []any{"68ad3f439bd2caa1a5758a9c"},
					},
					func(userID string) *dataWorkPostprocess.Metadata {
						return &dataWorkPostprocess.Metadata{
							Metadata:              userWork.Metadata{UserID: pointer.FromString(userID)},
							Reasons:               []string{"DATA_ADDED"},
							SummaryTypes:          []string{"cgm"},
							PendingSummaryUpdates: []string{"cgm"},
							PendingSummaryDeletes: []string{"68ad3f439bd2caa1a5758a9c"},
						}
					},
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the metadata",
				func(mutator func(metadata *dataWorkPostprocess.Metadata), expectedErrors ...error) {
					datum := &dataWorkPostprocess.Metadata{
						Metadata:     userWork.Metadata{UserID: pointer.FromString(userID)},
						Reasons:      []string{dataWorkPostprocess.ReasonDataAdded},
						SummaryTypes: []string{summaryTypes.SummaryTypeCGM},
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
				Entry("succeeds with every summary type", func(datum *dataWorkPostprocess.Metadata) {
					datum.SummaryTypes = summaryTypes.AllSummaryTypes
				}),
				Entry("reports the summary types are missing",
					func(datum *dataWorkPostprocess.Metadata) { datum.SummaryTypes = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/summaryTypes"),
				),
				Entry("reports the summary types are empty",
					func(datum *dataWorkPostprocess.Metadata) { datum.SummaryTypes = []string{} },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/summaryTypes"),
				),
				Entry("reports a summary type is not valid",
					func(datum *dataWorkPostprocess.Metadata) { datum.SummaryTypes = []string{"invalid"} },
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", summaryTypes.AllSummaryTypes), "/summaryTypes/0"),
				),
				Entry("reports the summary types are duplicated",
					func(datum *dataWorkPostprocess.Metadata) {
						datum.SummaryTypes = []string{summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeCGM}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/summaryTypes/1"),
				),
				Entry("succeeds with pending summary changes", func(datum *dataWorkPostprocess.Metadata) {
					datum.PendingSummaryUpdates = []string{summaryTypes.SummaryTypeBGM, summaryTypes.SummaryTypeCGM}
					datum.PendingSummaryDeletes = []string{"68ad3f439bd2caa1a5758a9c"}
				}),
				Entry("reports a pending summary update the clinic service has no representation of",
					func(datum *dataWorkPostprocess.Metadata) {
						datum.PendingSummaryUpdates = []string{summaryTypes.SummaryTypeContinuous}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf(summaryTypes.SummaryTypeContinuous, []string{summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeBGM}), "/pendingSummaryUpdates/0"),
				),
				Entry("reports the pending summary updates are duplicated",
					func(datum *dataWorkPostprocess.Metadata) {
						datum.PendingSummaryUpdates = []string{summaryTypes.SummaryTypeCGM, summaryTypes.SummaryTypeCGM}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/pendingSummaryUpdates/1"),
				),
				Entry("reports a pending summary delete is empty",
					func(datum *dataWorkPostprocess.Metadata) {
						datum.PendingSummaryDeletes = []string{""}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/pendingSummaryDeletes/0"),
				),
			)
		})

		// The metadata is encoded when the work is created and decoded when it is processed
		It("is unchanged by being encoded and decoded", func() {
			datum := &dataWorkPostprocess.Metadata{
				Metadata:              userWork.Metadata{UserID: pointer.FromString(userID)},
				Reasons:               []string{dataWorkPostprocess.ReasonLegacyDataAdded, dataWorkPostprocess.ReasonUploadCompleted},
				SummaryTypes:          []string{summaryTypes.SummaryTypeBGM, summaryTypes.SummaryTypeContinuous},
				PendingSummaryUpdates: []string{summaryTypes.SummaryTypeCGM},
				PendingSummaryDeletes: []string{"68ad3f439bd2caa1a5758a9c"},
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
