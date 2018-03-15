package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"time"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/validate"
)

var futureTime = time.Unix(4102444800, 0)

var _ = Describe("Base", func() {
	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			typ := testDataTypes.NewType()
			datum := types.New(typ)
			Expect(datum.Active).To(BeFalse())
			Expect(datum.Annotations).To(BeNil())
			Expect(datum.ArchivedDataSetID).To(BeNil())
			Expect(datum.ArchivedTime).To(BeNil())
			Expect(datum.ClockDriftOffset).To(BeNil())
			Expect(datum.ConversionOffset).To(BeNil())
			Expect(datum.CreatedTime).To(BeNil())
			Expect(datum.CreatedUserID).To(BeNil())
			Expect(datum.Deduplicator).To(BeNil())
			Expect(datum.DeletedTime).To(BeNil())
			Expect(datum.DeletedUserID).To(BeNil())
			Expect(datum.DeviceID).To(BeNil())
			Expect(datum.DeviceTime).To(BeNil())
			Expect(datum.GUID).To(BeNil())
			Expect(datum.ID).To(BeNil())
			Expect(datum.ModifiedTime).To(BeNil())
			Expect(datum.ModifiedUserID).To(BeNil())
			Expect(datum.Payload).To(BeNil())
			Expect(datum.SchemaVersion).To(Equal(0))
			Expect(datum.Source).To(BeNil())
			Expect(datum.Time).To(BeNil())
			Expect(datum.TimezoneOffset).To(BeNil())
			Expect(datum.Type).To(Equal(typ))
			Expect(datum.UploadID).To(BeNil())
			Expect(datum.UserID).To(BeNil())
			Expect(datum.Version).To(Equal(0))
		})
	})

	Context("with new datum", func() {
		var typ string
		var datum types.Base

		BeforeEach(func() {
			typ = testDataTypes.NewType()
			datum = types.New(typ)
		})

		Context("Meta", func() {
			It("returns the meta with type", func() {
				Expect(datum.Meta()).To(Equal(&types.Meta{Type: typ}))
			})
		})
	})

	Context("Base", func() {
		// Context("Parse", func() {
		// TODO
		// })

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *types.Base), expectedOrigins []structure.Origin, expectedErrors ...error) {
					datum := testDataTypes.NewBase()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, expectedOrigins, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *types.Base) {},
					structure.Origins(),
				),
				Entry("active true",
					func(datum *types.Base) { datum.Active = true },
					structure.Origins(),
				),
				Entry("active false",
					func(datum *types.Base) { datum.Active = false },
					structure.Origins(),
				),
				Entry("annotations missing",
					func(datum *types.Base) { datum.Annotations = nil },
					structure.Origins(),
				),
				Entry("annotations exist",
					func(datum *types.Base) { datum.Annotations = testData.NewBlobArray() },
					structure.Origins(),
				),
				Entry("archived data set id missing",
					func(datum *types.Base) { datum.ArchivedDataSetID = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/archivedDatasetId"),
				),
				Entry("archived data set id empty",
					func(datum *types.Base) { datum.ArchivedDataSetID = pointer.String("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivedDatasetId"),
				),
				Entry("archived data set id invalid",
					func(datum *types.Base) { datum.ArchivedDataSetID = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsDataSetIDNotValid("invalid"), "/archivedDatasetId"),
				),
				Entry("archived data set id valid",
					func(datum *types.Base) { datum.ArchivedDataSetID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("archived time missing; archived data set id missing",
					func(datum *types.Base) {
						datum.ArchivedDataSetID = nil
						datum.ArchivedTime = nil
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
				),
				Entry("archived time missing; archived data set id exists",
					func(datum *types.Base) {
						datum.ArchivedDataSetID = pointer.String(id.New())
						datum.ArchivedTime = nil
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/archivedDatasetId"),
				),
				Entry("archived time invalid",
					func(datum *types.Base) { datum.ArchivedTime = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/archivedTime"),
				),
				Entry("archived time not after created time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.String(time.Time{}.Format(time.RFC3339))
						datum.CreatedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("archived time not before now",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("clock drift offset missing",
					func(datum *types.Base) { datum.ClockDriftOffset = nil },
					structure.Origins(),
				),
				Entry("clock drift offset; out of range (lower)",
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.Int(-86400001) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-86400001, -86400000, 86400000), "/clockDriftOffset"),
				),
				Entry("clock drift offset; in range (lower)",
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.Int(-86400000) },
					structure.Origins(),
				),
				Entry("clock drift offset; in range (upper)",
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.Int(86400000) },
					structure.Origins(),
				),
				Entry("clock drift offset; out of range (upper)",
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.Int(86400001) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, -86400000, 86400000), "/clockDriftOffset"),
				),
				Entry("conversion offset missing",
					func(datum *types.Base) { datum.ConversionOffset = nil },
					structure.Origins(),
				),
				Entry("conversion offset exists",
					func(datum *types.Base) { datum.ConversionOffset = pointer.Int(testDataTypes.NewConversionOffset()) },
					structure.Origins(),
				),
				Entry("created user id missing",
					func(datum *types.Base) { datum.CreatedUserID = nil },
					structure.Origins(),
				),
				Entry("created user id empty",
					func(datum *types.Base) { datum.CreatedUserID = pointer.String("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdUserId"),
				),
				Entry("created user id invalid",
					func(datum *types.Base) { datum.CreatedUserID = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsUserIDNotValid("invalid"), "/createdUserId"),
				),
				Entry("created user id valid",
					func(datum *types.Base) { datum.CreatedUserID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("created time missing; created user id missing",
					func(datum *types.Base) {
						datum.CreatedTime = nil
						datum.CreatedUserID = nil
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
				),
				Entry("created time missing; created user id exists",
					func(datum *types.Base) {
						datum.CreatedTime = nil
						datum.CreatedUserID = pointer.String(id.New())
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/createdUserId"),
				),
				Entry("created time invalid",
					func(datum *types.Base) { datum.CreatedTime = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/createdTime"),
				),
				Entry("created time not before now",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.CreatedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("deleted user id missing",
					func(datum *types.Base) { datum.DeletedUserID = nil },
					structure.Origins(),
				),
				Entry("deleted user id empty",
					func(datum *types.Base) { datum.DeletedUserID = pointer.String("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deletedUserId"),
				),
				Entry("deleted user id invalid",
					func(datum *types.Base) { datum.DeletedUserID = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsUserIDNotValid("invalid"), "/deletedUserId"),
				),
				Entry("deleted user id valid",
					func(datum *types.Base) { datum.DeletedUserID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("deleted time missing; deleted user id missing",
					func(datum *types.Base) {
						datum.DeletedTime = nil
						datum.DeletedUserID = nil
					},
					structure.Origins(),
				),
				Entry("deleted time missing; deleted user id exists",
					func(datum *types.Base) {
						datum.DeletedTime = nil
						datum.DeletedUserID = pointer.String(id.New())
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/deletedUserId"),
				),
				Entry("deleted time invalid",
					func(datum *types.Base) { datum.DeletedTime = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/deletedTime"),
				),
				Entry("deleted time not after archived time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.String(time.Time{}.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("deleted time not after created time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.CreatedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.String(time.Time{}.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("deleted time not after modified time",
					func(datum *types.Base) {
						datum.DeletedTime = pointer.String(time.Time{}.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("deleted time not before now",
					func(datum *types.Base) {
						datum.DeletedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
				),
				Entry("deduplicator missing",
					func(datum *types.Base) { datum.Deduplicator = nil },
					structure.Origins(),
				),
				Entry("deduplicator invalid",
					func(datum *types.Base) { datum.Deduplicator.Name = "invalid" },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(validate.ErrorValueStringAsReverseDomainNotValid("invalid"), "/_deduplicator/name"),
				),
				Entry("deduplicator valid",
					func(datum *types.Base) { datum.Deduplicator = testData.NewDeduplicatorDescriptor() },
					structure.Origins(),
				),
				Entry("device id missing",
					func(datum *types.Base) { datum.DeviceID = nil },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deviceId"),
				),
				Entry("device id empty",
					func(datum *types.Base) { datum.DeviceID = pointer.String("") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceId"),
				),
				Entry("device id valid",
					func(datum *types.Base) { datum.DeviceID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("device time missing",
					func(datum *types.Base) { datum.DeviceTime = nil },
					structure.Origins(),
				),
				Entry("device time invalid",
					func(datum *types.Base) { datum.DeviceTime = pointer.String("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/deviceTime"),
				),
				Entry("device time valid",
					func(datum *types.Base) {
						datum.DeviceTime = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
					},
					structure.Origins(),
				),
				Entry("id missing",
					func(datum *types.Base) { datum.ID = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *types.Base) { datum.ID = pointer.String("") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id invalid",
					func(datum *types.Base) { datum.ID = pointer.String("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(id.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("id valid",
					func(datum *types.Base) { datum.ID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("modified user id missing",
					func(datum *types.Base) { datum.ModifiedUserID = nil },
					structure.Origins(),
				),
				Entry("modified user id empty",
					func(datum *types.Base) { datum.ModifiedUserID = pointer.String("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/modifiedUserId"),
				),
				Entry("modified user id invalid",
					func(datum *types.Base) { datum.ModifiedUserID = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsUserIDNotValid("invalid"), "/modifiedUserId"),
				),
				Entry("modified user id valid",
					func(datum *types.Base) { datum.ModifiedUserID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("modified time missing; modified user id missing",
					func(datum *types.Base) {
						datum.ArchivedTime = nil
						datum.ArchivedDataSetID = nil
						datum.ModifiedTime = nil
						datum.ModifiedUserID = nil
					},
					structure.Origins(),
				),
				Entry("modified time missing; modified user id exists",
					func(datum *types.Base) {
						datum.ArchivedTime = nil
						datum.ArchivedDataSetID = nil
						datum.ModifiedTime = nil
						datum.ModifiedUserID = pointer.String(id.New())
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/modifiedUserId"),
				),
				Entry("modified time missing; modified user id missing; archived time exists",
					func(datum *types.Base) {
						datum.ArchivedTime = datum.ModifiedTime
						datum.ModifiedTime = nil
						datum.ModifiedUserID = nil
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/modifiedTime"),
				),
				Entry("modified time missing; modified user id exists; archived time exists",
					func(datum *types.Base) {
						datum.ArchivedTime = datum.ModifiedTime
						datum.ModifiedTime = nil
						datum.ModifiedUserID = pointer.String(id.New())
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/modifiedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/modifiedUserId"),
				),
				Entry("modified time invalid",
					func(datum *types.Base) {
						datum.ModifiedTime = pointer.String("invalid")
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/modifiedTime"),
				),
				Entry("modified time not after archived time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(time.Time{}.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/modifiedTime"),
				),
				Entry("modified time not after created time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.CreatedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(time.Time{}.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/modifiedTime"),
				),
				Entry("modified time not before now",
					func(datum *types.Base) {
						datum.DeletedTime = pointer.String(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.String(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("schema version; out of range (lower)",
					func(datum *types.Base) { datum.SchemaVersion = 0 },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 3), "/_schemaVersion"),
				),
				Entry("schema version; in range (lower)",
					func(datum *types.Base) { datum.SchemaVersion = 1 },
					structure.Origins(),
				),
				Entry("schema version; in range (upper)",
					func(datum *types.Base) { datum.SchemaVersion = 3 },
					structure.Origins(),
				),
				Entry("schema version; out of range (upper)",
					func(datum *types.Base) { datum.SchemaVersion = 4 },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(4, 1, 3), "/_schemaVersion"),
				),
				Entry("source missing",
					func(datum *types.Base) { datum.Source = nil },
					structure.Origins(),
				),
				Entry("source invalid",
					func(datum *types.Base) { datum.Source = pointer.String("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "carelink"), "/source"),
				),
				Entry("source valid",
					func(datum *types.Base) { datum.Source = pointer.String("carelink") },
					structure.Origins(),
				),
				Entry("time missing",
					func(datum *types.Base) { datum.Time = nil },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
				),
				Entry("time invalid",
					func(datum *types.Base) { datum.Time = pointer.String("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/time"),
				),
				Entry("time valid",
					func(datum *types.Base) { datum.Time = pointer.String(test.NewTime().Format(time.RFC3339)) },
					structure.Origins(),
				),
				Entry("time zone offset missing",
					func(datum *types.Base) { datum.TimezoneOffset = nil },
					structure.Origins(),
				),
				Entry("time zone offset; out of range (lower)",
					func(datum *types.Base) { datum.TimezoneOffset = pointer.Int(-10081) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-10081, -10080, 10080), "/timezoneOffset"),
				),
				Entry("time zone offset; in range (lower)",
					func(datum *types.Base) { datum.TimezoneOffset = pointer.Int(-10080) },
					structure.Origins(),
				),
				Entry("time zone offset; in range (upper)",
					func(datum *types.Base) { datum.TimezoneOffset = pointer.Int(10080) },
					structure.Origins(),
				),
				Entry("time zone offset; out of range (upper)",
					func(datum *types.Base) { datum.TimezoneOffset = pointer.Int(10081) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(10081, -10080, 10080), "/timezoneOffset"),
				),
				Entry("type empty",
					func(datum *types.Base) { datum.Type = "" },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type valid",
					func(datum *types.Base) { datum.Type = id.New() },
					structure.Origins(),
				),
				Entry("upload id missing",
					func(datum *types.Base) { datum.UploadID = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/uploadId"),
				),
				Entry("upload id empty",
					func(datum *types.Base) { datum.UploadID = pointer.String("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/uploadId"),
				),
				Entry("upload id invalid",
					func(datum *types.Base) { datum.UploadID = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsDataSetIDNotValid("invalid"), "/uploadId"),
				),
				Entry("upload id valid",
					func(datum *types.Base) { datum.UploadID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("user id missing",
					func(datum *types.Base) { datum.UserID = nil },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/_userId"),
				),
				Entry("user id empty",
					func(datum *types.Base) { datum.UserID = pointer.String("") },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/_userId"),
				),
				Entry("user id invalid",
					func(datum *types.Base) { datum.UserID = pointer.String("invalid") },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsUserIDNotValid("invalid"), "/_userId"),
				),
				Entry("user id valid",
					func(datum *types.Base) { datum.UserID = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("version; out of range (lower)",
					func(datum *types.Base) { datum.Version = -1 },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/_version"),
				),
				Entry("version; in range (lower)",
					func(datum *types.Base) { datum.Version = 0 },
					structure.Origins(),
				),
				Entry("multiple errors with store origin",
					func(datum *types.Base) {
						datum.SchemaVersion = 0
						datum.UserID = nil
						datum.Version = -1
					},
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 3), "/_schemaVersion"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/_userId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/_version"),
				),
				Entry("multiple errors with internal origin",
					func(datum *types.Base) {
						datum.ArchivedDataSetID = pointer.String("invalid")
						datum.ArchivedTime = pointer.String("invalid")
						datum.CreatedTime = pointer.String("invalid")
						datum.CreatedUserID = pointer.String("invalid")
						datum.DeletedTime = pointer.String("invalid")
						datum.DeletedUserID = pointer.String("invalid")
						datum.Deduplicator.Name = "invalid"
						datum.ID = nil
						datum.ModifiedTime = pointer.String("invalid")
						datum.ModifiedUserID = pointer.String("invalid")
						datum.UploadID = nil
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsDataSetIDNotValid("invalid"), "/archivedDatasetId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/createdTime"),
					testErrors.WithPointerSource(data.ErrorValueStringAsUserIDNotValid("invalid"), "/createdUserId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/deletedTime"),
					testErrors.WithPointerSource(data.ErrorValueStringAsUserIDNotValid("invalid"), "/deletedUserId"),
					testErrors.WithPointerSource(validate.ErrorValueStringAsReverseDomainNotValid("invalid"), "/_deduplicator/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/modifiedTime"),
					testErrors.WithPointerSource(data.ErrorValueStringAsUserIDNotValid("invalid"), "/modifiedUserId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/uploadId"),
				),
				Entry("multiple errors with external origin",
					func(datum *types.Base) {
						datum.ClockDriftOffset = pointer.Int(-86400001)
						datum.DeviceID = nil
						datum.DeviceTime = pointer.String("invalid")
						datum.ID = pointer.String("")
						datum.Source = pointer.String("invalid")
						datum.Time = nil
						datum.TimezoneOffset = pointer.Int(-10081)
						datum.Type = ""
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-86400001, -86400000, 86400000), "/clockDriftOffset"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/deviceId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/deviceTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "carelink"), "/source"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-10081, -10080, 10080), "/timezoneOffset"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *types.Base), expectator func(datum *types.Base, expectedDatum *types.Base)) {
					for _, origin := range structure.Origins() {
						datum := testDataTypes.NewBase()
						mutator(datum)
						expectedDatum := testDataTypes.CloneBase(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("does not modify the datum",
					func(datum *types.Base) {},
					nil,
				),
			)

			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *types.Base), expectator func(datum *types.Base, expectedDatum *types.Base)) {
					datum := testDataTypes.NewBase()
					mutator(datum)
					expectedDatum := testDataTypes.CloneBase(datum)
					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					Expect(normalizer.Error()).To(BeNil())
					Expect(normalizer.Data()).To(BeEmpty())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("guid missing",
					func(datum *types.Base) { datum.GUID = nil },
					func(datum *types.Base, expectedDatum *types.Base) {
						Expect(datum.GUID).ToNot(BeNil())
						Expect(datum.GUID).ToNot(Equal(expectedDatum.GUID))
						expectedDatum.GUID = datum.GUID
					},
				),
				Entry("id missing",
					func(datum *types.Base) { datum.ID = nil },
					func(datum *types.Base, expectedDatum *types.Base) {
						Expect(datum.ID).ToNot(BeNil())
						Expect(datum.ID).ToNot(Equal(expectedDatum.ID))
						expectedDatum.ID = datum.ID
					},
				),
				Entry("default schema version",
					func(datum *types.Base) { datum.SchemaVersion = 0 },
					func(datum *types.Base, expectedDatum *types.Base) {
						Expect(datum.SchemaVersion).To(Equal(3))
						expectedDatum.SchemaVersion = datum.SchemaVersion
					},
				),
				Entry("all missing",
					func(datum *types.Base) {
						*datum = types.New("")
					},
					func(datum *types.Base, expectedDatum *types.Base) {
						Expect(datum.GUID).ToNot(BeNil())
						Expect(datum.GUID).ToNot(Equal(expectedDatum.GUID))
						Expect(datum.ID).ToNot(BeNil())
						Expect(datum.ID).ToNot(Equal(expectedDatum.ID))
						Expect(datum.SchemaVersion).To(Equal(3))
						expectedDatum.GUID = datum.GUID
						expectedDatum.ID = datum.ID
						expectedDatum.SchemaVersion = datum.SchemaVersion
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *types.Base), expectator func(datum *types.Base, expectedDatum *types.Base)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := testDataTypes.NewBase()
						mutator(datum)
						expectedDatum := testDataTypes.CloneBase(datum)
						normalizer := dataNormalizer.New()
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(origin))
						Expect(normalizer.Error()).To(BeNil())
						Expect(normalizer.Data()).To(BeEmpty())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					}
				},
				Entry("guid missing",
					func(datum *types.Base) { datum.GUID = nil },
					nil,
				),
				Entry("id missing",
					func(datum *types.Base) { datum.ID = nil },
					nil,
				),
				Entry("default schema version",
					func(datum *types.Base) { datum.SchemaVersion = 0 },
					nil,
				),
				Entry("all missing",
					func(datum *types.Base) {
						*datum = types.New("")
					},
					nil,
				),
			)
		})
	})

	Context("with new, initialized datum", func() {
		var datum *types.Base

		BeforeEach(func() {
			datum = testDataTypes.NewBase()
		})

		Context("IdentityFields", func() {
			It("returns error if user id is missing", func() {
				datum.UserID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if user id is empty", func() {
				datum.UserID = pointer.String("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("user id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if device id is missing", func() {
				datum.DeviceID = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("device id is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if device id is empty", func() {
				datum.DeviceID = pointer.String("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("device id is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if time is missing", func() {
				datum.Time = nil
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("time is missing"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if time is empty", func() {
				datum.Time = pointer.String("")
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("time is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns error if type is empty", func() {
				datum.Type = ""
				identityFields, err := datum.IdentityFields()
				Expect(err).To(MatchError("type is empty"))
				Expect(identityFields).To(BeEmpty())
			})

			It("returns the expected identity fields", func() {
				identityFields, err := datum.IdentityFields()
				Expect(err).ToNot(HaveOccurred())
				Expect(identityFields).To(Equal([]string{*datum.UserID, *datum.DeviceID, *datum.Time, datum.Type}))
			})
		})

		Context("GetPayload", func() {
			It("gets the payload", func() {
				Expect(datum.GetPayload()).To(Equal(datum.Payload))
			})
		})

		Context("SetUserID", func() {
			It("sets the user id", func() {
				userID := pointer.String(id.New())
				datum.SetUserID(userID)
				Expect(datum.UserID).To(Equal(userID))
			})
		})

		Context("SetDatasetID", func() {
			It("sets the data set id", func() {
				dataSetID := pointer.String(id.New())
				datum.SetDatasetID(dataSetID)
				Expect(datum.UploadID).To(Equal(dataSetID))
			})
		})

		Context("SetActive", func() {
			It("sets active to true", func() {
				datum.SetActive(true)
				Expect(datum.Active).To(BeTrue())
			})

			It("sets active to false", func() {
				datum.SetActive(false)
				Expect(datum.Active).To(BeFalse())
			})
		})

		Context("SetDeviceID", func() {
			It("sets the device id", func() {
				deviceID := pointer.String(id.New())
				datum.SetDeviceID(deviceID)
				Expect(datum.DeviceID).To(Equal(deviceID))
			})
		})

		Context("SetCreatedTime", func() {
			It("sets the created time", func() {
				createdTime := pointer.String(time.Now().Format(time.RFC3339))
				datum.SetCreatedTime(createdTime)
				Expect(datum.CreatedTime).To(Equal(createdTime))
			})
		})

		Context("SetCreatedUserID", func() {
			It("sets the created user id", func() {
				createdUserID := pointer.String(id.New())
				datum.SetCreatedUserID(createdUserID)
				Expect(datum.CreatedUserID).To(Equal(createdUserID))
			})
		})

		Context("SetModifiedTime", func() {
			It("sets the modified time", func() {
				modifiedTime := pointer.String(time.Now().Format(time.RFC3339))
				datum.SetModifiedTime(modifiedTime)
				Expect(datum.ModifiedTime).To(Equal(modifiedTime))
			})
		})

		Context("SetModifiedUserID", func() {
			It("sets the modified user id", func() {
				modifiedUserID := pointer.String(id.New())
				datum.SetModifiedUserID(modifiedUserID)
				Expect(datum.ModifiedUserID).To(Equal(modifiedUserID))
			})
		})

		Context("SetDeletedTime", func() {
			It("sets the deleted time", func() {
				deletedTime := pointer.String(time.Now().Format(time.RFC3339))
				datum.SetDeletedTime(deletedTime)
				Expect(datum.DeletedTime).To(Equal(deletedTime))
			})
		})

		Context("SetDeletedUserID", func() {
			It("sets the deleted user id", func() {
				deletedUserID := pointer.String(id.New())
				datum.SetDeletedUserID(deletedUserID)
				Expect(datum.DeletedUserID).To(Equal(deletedUserID))
			})
		})

		Context("DeduplicatorDescriptor", func() {
			It("gets the deduplicator descriptor", func() {
				Expect(datum.DeduplicatorDescriptor()).To(Equal(datum.Deduplicator))
			})
		})

		Context("SetDeduplicatorDescriptor", func() {
			It("sets the deduplicator descriptor", func() {
				deduplicatorDescriptor := testData.NewDeduplicatorDescriptor()
				datum.SetDeduplicatorDescriptor(deduplicatorDescriptor)
				Expect(datum.Deduplicator).To(Equal(deduplicatorDescriptor))
			})
		})
	})
})
