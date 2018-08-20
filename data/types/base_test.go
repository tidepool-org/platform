package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"sort"
	"time"

	"github.com/tidepool-org/platform/data"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types"
	testDataTypesCommonAssociation "github.com/tidepool-org/platform/data/types/common/association/test"
	testDataTypesCommonLocation "github.com/tidepool-org/platform/data/types/common/location/test"
	testDataTypesCommonOrigin "github.com/tidepool-org/platform/data/types/common/origin/test"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	timeZone "github.com/tidepool-org/platform/time/zone"
	timeZoneTest "github.com/tidepool-org/platform/time/zone/test"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var futureTime = time.Unix(4102444800, 0)

var _ = Describe("Base", func() {
	Context("New", func() {
		It("creates a new datum with all values initialized", func() {
			typ := testDataTypes.NewType()
			datum := types.New(typ)
			Expect(datum.Active).To(BeFalse())
			Expect(datum.Annotations).To(BeNil())
			Expect(datum.Associations).To(BeNil())
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
			Expect(datum.Location).To(BeNil())
			Expect(datum.ModifiedTime).To(BeNil())
			Expect(datum.ModifiedUserID).To(BeNil())
			Expect(datum.Notes).To(BeNil())
			Expect(datum.Origin).To(BeNil())
			Expect(datum.Payload).To(BeNil())
			Expect(datum.SchemaVersion).To(Equal(0))
			Expect(datum.Source).To(BeNil())
			Expect(datum.Tags).To(BeNil())
			Expect(datum.Time).To(BeNil())
			Expect(datum.TimeZoneName).To(BeNil())
			Expect(datum.TimeZoneOffset).To(BeNil())
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
				Entry("associations missing",
					func(datum *types.Base) { datum.Associations = nil },
					structure.Origins(),
				),
				Entry("associations invalid",
					func(datum *types.Base) { (*datum.Associations)[0].Type = nil },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/associations/0/type"),
				),
				Entry("associations valid",
					func(datum *types.Base) { datum.Associations = testDataTypesCommonAssociation.NewAssociationArray() },
					structure.Origins(),
				),
				Entry("archived data set id missing",
					func(datum *types.Base) { datum.ArchivedDataSetID = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/archivedDatasetId"),
				),
				Entry("archived data set id empty",
					func(datum *types.Base) { datum.ArchivedDataSetID = pointer.FromString("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/archivedDatasetId"),
				),
				Entry("archived data set id invalid",
					func(datum *types.Base) { datum.ArchivedDataSetID = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/archivedDatasetId"),
				),
				Entry("archived data set id valid",
					func(datum *types.Base) { datum.ArchivedDataSetID = pointer.FromString(dataTest.RandomSetID()) },
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
						datum.ArchivedDataSetID = pointer.FromString(dataTest.RandomSetID())
						datum.ArchivedTime = nil
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/archivedDatasetId"),
				),
				Entry("archived time invalid",
					func(datum *types.Base) { datum.ArchivedTime = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/archivedTime"),
				),
				Entry("archived time not after created time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.FromString(time.Time{}.Format(time.RFC3339))
						datum.CreatedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("archived time not before now",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(futureTime.Format(time.RFC3339))
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
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.FromInt(-86400001) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-86400001, -86400000, 86400000), "/clockDriftOffset"),
				),
				Entry("clock drift offset; in range (lower)",
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.FromInt(-86400000) },
					structure.Origins(),
				),
				Entry("clock drift offset; in range (upper)",
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.FromInt(86400000) },
					structure.Origins(),
				),
				Entry("clock drift offset; out of range (upper)",
					func(datum *types.Base) { datum.ClockDriftOffset = pointer.FromInt(86400001) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(86400001, -86400000, 86400000), "/clockDriftOffset"),
				),
				Entry("conversion offset missing",
					func(datum *types.Base) { datum.ConversionOffset = nil },
					structure.Origins(),
				),
				Entry("conversion offset exists",
					func(datum *types.Base) { datum.ConversionOffset = pointer.FromInt(testDataTypes.NewConversionOffset()) },
					structure.Origins(),
				),
				Entry("created user id missing",
					func(datum *types.Base) { datum.CreatedUserID = nil },
					structure.Origins(),
				),
				Entry("created user id empty",
					func(datum *types.Base) { datum.CreatedUserID = pointer.FromString("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdUserId"),
				),
				Entry("created user id invalid",
					func(datum *types.Base) { datum.CreatedUserID = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/createdUserId"),
				),
				Entry("created user id valid",
					func(datum *types.Base) { datum.CreatedUserID = pointer.FromString(userTest.RandomID()) },
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
						datum.CreatedUserID = pointer.FromString(userTest.RandomID())
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/createdUserId"),
				),
				Entry("created time invalid",
					func(datum *types.Base) { datum.CreatedTime = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/createdTime"),
				),
				Entry("created time not before now",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.CreatedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(futureTime.Format(time.RFC3339))
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
					func(datum *types.Base) { datum.DeletedUserID = pointer.FromString("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deletedUserId"),
				),
				Entry("deleted user id invalid",
					func(datum *types.Base) { datum.DeletedUserID = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/deletedUserId"),
				),
				Entry("deleted user id valid",
					func(datum *types.Base) { datum.DeletedUserID = pointer.FromString(userTest.RandomID()) },
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
						datum.DeletedUserID = pointer.FromString(userTest.RandomID())
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/deletedUserId"),
				),
				Entry("deleted time invalid",
					func(datum *types.Base) { datum.DeletedTime = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/deletedTime"),
				),
				Entry("deleted time not after archived time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.FromString(time.Time{}.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("deleted time not after created time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.CreatedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.FromString(time.Time{}.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("deleted time not after modified time",
					func(datum *types.Base) {
						datum.DeletedTime = pointer.FromString(time.Time{}.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("deleted time not before now",
					func(datum *types.Base) {
						datum.DeletedTime = pointer.FromString(futureTime.Format(time.RFC3339))
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
					testErrors.WithPointerSource(net.ErrorValueStringAsReverseDomainNotValid("invalid"), "/_deduplicator/name"),
				),
				Entry("deduplicator valid",
					func(datum *types.Base) { datum.Deduplicator = testData.NewDeduplicatorDescriptor() },
					structure.Origins(),
				),
				Entry("device id missing",
					func(datum *types.Base) { datum.DeviceID = nil },
					structure.Origins(),
				),
				Entry("device id empty",
					func(datum *types.Base) { datum.DeviceID = pointer.FromString("") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceId"),
				),
				Entry("device id valid",
					func(datum *types.Base) { datum.DeviceID = pointer.FromString(testData.NewDeviceID()) },
					structure.Origins(),
				),
				Entry("device time missing",
					func(datum *types.Base) { datum.DeviceTime = nil },
					structure.Origins(),
				),
				Entry("device time invalid",
					func(datum *types.Base) { datum.DeviceTime = pointer.FromString("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/deviceTime"),
				),
				Entry("device time valid",
					func(datum *types.Base) {
						datum.DeviceTime = pointer.FromString(test.NewTime().Format("2006-01-02T15:04:05"))
					},
					structure.Origins(),
				),
				Entry("id missing",
					func(datum *types.Base) { datum.ID = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *types.Base) { datum.ID = pointer.FromString("") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id invalid",
					func(datum *types.Base) { datum.ID = pointer.FromString("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(data.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("id valid",
					func(datum *types.Base) { datum.ID = pointer.FromString(dataTest.RandomID()) },
					structure.Origins(),
				),
				Entry("location missing",
					func(datum *types.Base) { datum.Location = nil },
					structure.Origins(),
				),
				Entry("location invalid",
					func(datum *types.Base) {
						datum.Location.GPS = nil
						datum.Location.Name = nil
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/location/gps"),
				),
				Entry("location valid",
					func(datum *types.Base) { datum.Location = testDataTypesCommonLocation.NewLocation() },
					structure.Origins(),
				),
				Entry("modified user id missing",
					func(datum *types.Base) { datum.ModifiedUserID = nil },
					structure.Origins(),
				),
				Entry("modified user id empty",
					func(datum *types.Base) { datum.ModifiedUserID = pointer.FromString("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/modifiedUserId"),
				),
				Entry("modified user id invalid",
					func(datum *types.Base) { datum.ModifiedUserID = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/modifiedUserId"),
				),
				Entry("modified user id valid",
					func(datum *types.Base) { datum.ModifiedUserID = pointer.FromString(userTest.RandomID()) },
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
						datum.ModifiedUserID = pointer.FromString(userTest.RandomID())
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
						datum.ModifiedUserID = pointer.FromString(userTest.RandomID())
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/modifiedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueExists(), "/modifiedUserId"),
				),
				Entry("modified time invalid",
					func(datum *types.Base) {
						datum.ModifiedTime = pointer.FromString("invalid")
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/modifiedTime"),
				),
				Entry("modified time not after archived time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(time.Time{}.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/modifiedTime"),
				),
				Entry("modified time not after created time",
					func(datum *types.Base) {
						datum.ArchivedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.DeletedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.CreatedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(time.Time{}.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(time.Time{}, futureTime), "/modifiedTime"),
				),
				Entry("modified time not before now",
					func(datum *types.Base) {
						datum.DeletedTime = pointer.FromString(futureTime.Format(time.RFC3339))
						datum.ModifiedTime = pointer.FromString(futureTime.Format(time.RFC3339))
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/deletedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("notes missing",
					func(datum *types.Base) { datum.Notes = nil },
					structure.Origins(),
				),
				Entry("notes empty",
					func(datum *types.Base) { datum.Notes = pointer.FromStringArray([]string{}) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/notes"),
				),
				Entry("notes length; in range (upper)",
					func(datum *types.Base) { datum.Notes = pointer.FromStringArray(testDataTypes.NewNotes(100, 100)) },
					structure.Origins(),
				),
				Entry("notes length; out of range (upper)",
					func(datum *types.Base) { datum.Notes = pointer.FromStringArray(testDataTypes.NewNotes(101, 101)) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/notes"),
				),
				Entry("notes note empty",
					func(datum *types.Base) {
						datum.Notes = pointer.FromStringArray(append([]string{testDataTypes.NewNote(1, 1000), "", testDataTypes.NewNote(1, 1000), ""}, testDataTypes.NewNotes(0, 96)...))
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/notes/1"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/notes/3"),
				),
				Entry("notes note length; in range (upper)",
					func(datum *types.Base) {
						datum.Notes = pointer.FromStringArray(append([]string{testDataTypes.NewNote(1000, 1000), testDataTypes.NewNote(1, 1000), testDataTypes.NewNote(1000, 1000)}, testDataTypes.NewNotes(0, 97)...))
					},
					structure.Origins(),
				),
				Entry("notes note length; out of range (upper)",
					func(datum *types.Base) {
						datum.Notes = pointer.FromStringArray(append([]string{testDataTypes.NewNote(1001, 1001), testDataTypes.NewNote(1, 1000), testDataTypes.NewNote(1001, 1001)}, testDataTypes.NewNotes(0, 97)...))
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(1001, 1000), "/notes/0"),
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(1001, 1000), "/notes/2"),
				),
				Entry("origin missing",
					func(datum *types.Base) { datum.Origin = nil },
					structure.Origins(),
				),
				Entry("origin invalid",
					func(datum *types.Base) { datum.Origin.Name = pointer.FromString("") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/origin/name"),
				),
				Entry("origin valid",
					func(datum *types.Base) { datum.Origin = testDataTypesCommonOrigin.NewOrigin() },
					structure.Origins(),
				),
				Entry("payload missing",
					func(datum *types.Base) { datum.Payload = nil },
					structure.Origins(),
				),
				Entry("payload exists",
					func(datum *types.Base) { datum.Payload = testData.NewBlob() },
					structure.Origins(),
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
					func(datum *types.Base) { datum.Source = pointer.FromString("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "carelink"), "/source"),
				),
				Entry("source valid",
					func(datum *types.Base) { datum.Source = pointer.FromString("carelink") },
					structure.Origins(),
				),
				Entry("tags missing",
					func(datum *types.Base) { datum.Tags = nil },
					structure.Origins(),
				),
				Entry("tags empty",
					func(datum *types.Base) { datum.Tags = pointer.FromStringArray([]string{}) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/tags"),
				),
				Entry("tags length; in range (upper)",
					func(datum *types.Base) { datum.Tags = pointer.FromStringArray(testDataTypes.NewTags(100, 100)) },
					structure.Origins(),
				),
				Entry("tags length; out of range (upper)",
					func(datum *types.Base) { datum.Tags = pointer.FromStringArray(testDataTypes.NewTags(101, 101)) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/tags"),
				),
				Entry("tags tag empty",
					func(datum *types.Base) {
						datum.Tags = pointer.FromStringArray(append([]string{testDataTypes.NewTag(100, 100), ""}, testDataTypes.NewTags(0, 98)...))
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/tags/1"),
				),
				Entry("tags tag length; in range (upper)",
					func(datum *types.Base) {
						datum.Tags = pointer.FromStringArray(append([]string{testDataTypes.NewTag(100, 100)}, testDataTypes.NewTags(0, 99)...))
					},
					structure.Origins(),
				),
				Entry("tags tag length; out of range (upper)",
					func(datum *types.Base) {
						datum.Tags = pointer.FromStringArray(append([]string{testDataTypes.NewTag(101, 101)}, testDataTypes.NewTags(0, 99)...))
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/tags/0"),
				),
				Entry("tags tag duplicate",
					func(datum *types.Base) {
						tags := testDataTypes.NewTags(5, 99)
						datum.Tags = pointer.FromStringArray(append([]string{tags[4]}, tags...))
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/tags/5"),
				),
				Entry("time missing",
					func(datum *types.Base) { datum.Time = nil },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
				),
				Entry("time invalid",
					func(datum *types.Base) { datum.Time = pointer.FromString("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/time"),
				),
				Entry("time valid",
					func(datum *types.Base) { datum.Time = pointer.FromString(test.NewTime().Format(time.RFC3339)) },
					structure.Origins(),
				),
				Entry("time zone name missing",
					func(datum *types.Base) { datum.TimeZoneName = nil },
					structure.Origins(),
				),
				Entry("time zone name empty",
					func(datum *types.Base) { datum.TimeZoneName = pointer.FromString("") },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timezone"),
				),
				Entry("time zone name invalid",
					func(datum *types.Base) { datum.TimeZoneName = pointer.FromString("invalid") },
					structure.Origins(),
					testErrors.WithPointerSource(timeZone.ErrorValueStringAsNameNotValid("invalid"), "/timezone"),
				),
				Entry("time zone name valid",
					func(datum *types.Base) { datum.TimeZoneName = pointer.FromString(timeZoneTest.RandomName()) },
					structure.Origins(),
				),
				Entry("time zone offset missing",
					func(datum *types.Base) { datum.TimeZoneOffset = nil },
					structure.Origins(),
				),
				Entry("time zone offset; out of range (lower)",
					func(datum *types.Base) { datum.TimeZoneOffset = pointer.FromInt(-10081) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-10081, -10080, 10080), "/timezoneOffset"),
				),
				Entry("time zone offset; in range (lower)",
					func(datum *types.Base) { datum.TimeZoneOffset = pointer.FromInt(-10080) },
					structure.Origins(),
				),
				Entry("time zone offset; in range (upper)",
					func(datum *types.Base) { datum.TimeZoneOffset = pointer.FromInt(10080) },
					structure.Origins(),
				),
				Entry("time zone offset; out of range (upper)",
					func(datum *types.Base) { datum.TimeZoneOffset = pointer.FromInt(10081) },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(10081, -10080, 10080), "/timezoneOffset"),
				),
				Entry("type empty",
					func(datum *types.Base) { datum.Type = "" },
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/type"),
				),
				Entry("type valid",
					func(datum *types.Base) { datum.Type = test.NewVariableString(1, 16, test.CharsetAlphaNumeric) },
					structure.Origins(),
				),
				Entry("upload id missing",
					func(datum *types.Base) { datum.UploadID = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/uploadId"),
				),
				Entry("upload id empty",
					func(datum *types.Base) { datum.UploadID = pointer.FromString("") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/uploadId"),
				),
				Entry("upload id invalid",
					func(datum *types.Base) { datum.UploadID = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/uploadId"),
				),
				Entry("upload id valid",
					func(datum *types.Base) { datum.UploadID = pointer.FromString(dataTest.RandomSetID()) },
					structure.Origins(),
				),
				Entry("user id missing",
					func(datum *types.Base) { datum.UserID = nil },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/_userId"),
				),
				Entry("user id empty",
					func(datum *types.Base) { datum.UserID = pointer.FromString("") },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/_userId"),
				),
				Entry("user id invalid",
					func(datum *types.Base) { datum.UserID = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginStore},
					testErrors.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/_userId"),
				),
				Entry("user id valid",
					func(datum *types.Base) { datum.UserID = pointer.FromString(userTest.RandomID()) },
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
						datum.ArchivedDataSetID = pointer.FromString("invalid")
						datum.ArchivedTime = pointer.FromString("invalid")
						datum.CreatedTime = pointer.FromString("invalid")
						datum.CreatedUserID = pointer.FromString("invalid")
						datum.DeletedTime = pointer.FromString("invalid")
						datum.DeletedUserID = pointer.FromString("invalid")
						datum.Deduplicator.Name = "invalid"
						datum.ID = nil
						datum.ModifiedTime = pointer.FromString("invalid")
						datum.ModifiedUserID = pointer.FromString("invalid")
						datum.UploadID = nil
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/archivedDatasetId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/archivedTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/createdTime"),
					testErrors.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/createdUserId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/deletedTime"),
					testErrors.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/deletedUserId"),
					testErrors.WithPointerSource(net.ErrorValueStringAsReverseDomainNotValid("invalid"), "/_deduplicator/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", time.RFC3339), "/modifiedTime"),
					testErrors.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/modifiedUserId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/uploadId"),
				),
				Entry("multiple errors with external origin",
					func(datum *types.Base) {
						datum.ClockDriftOffset = pointer.FromInt(-86400001)
						datum.DeviceID = pointer.FromString("")
						datum.DeviceTime = pointer.FromString("invalid")
						datum.ID = pointer.FromString("")
						datum.Location.GPS = nil
						datum.Location.Name = nil
						datum.Notes = pointer.FromStringArray([]string{})
						datum.Origin.Name = pointer.FromString("")
						datum.Source = pointer.FromString("invalid")
						datum.Tags = pointer.FromStringArray([]string{})
						datum.Time = nil
						datum.TimeZoneName = pointer.FromString("")
						datum.TimeZoneOffset = pointer.FromInt(-10081)
						datum.Type = ""
					},
					structure.Origins(),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotInRange(-86400001, -86400000, 86400000), "/clockDriftOffset"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceId"),
					testErrors.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/deviceTime"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/location/gps"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/notes"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/origin/name"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalid", "carelink"), "/source"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/tags"),
					testErrors.WithPointerSource(structureValidator.ErrorValueNotExists(), "/time"),
					testErrors.WithPointerSource(structureValidator.ErrorValueEmpty(), "/timezone"),
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
				Entry("tags sorted",
					func(datum *types.Base) {},
					func(datum *types.Base, expectedDatum *types.Base) {
						sort.Strings(*expectedDatum.Tags)
					},
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
				Entry("id missing",
					func(datum *types.Base) { datum.ID = nil },
					func(datum *types.Base, expectedDatum *types.Base) {
						Expect(datum.ID).ToNot(BeNil())
						Expect(datum.ID).ToNot(Equal(expectedDatum.ID))
						expectedDatum.ID = datum.ID
						sort.Strings(*expectedDatum.Tags)
					},
				),
				Entry("default schema version",
					func(datum *types.Base) { datum.SchemaVersion = 0 },
					func(datum *types.Base, expectedDatum *types.Base) {
						Expect(datum.SchemaVersion).To(Equal(3))
						expectedDatum.SchemaVersion = datum.SchemaVersion
						sort.Strings(*expectedDatum.Tags)
					},
				),
				Entry("all missing",
					func(datum *types.Base) {
						*datum = types.New("")
					},
					func(datum *types.Base, expectedDatum *types.Base) {
						Expect(datum.ID).ToNot(BeNil())
						Expect(datum.ID).ToNot(Equal(expectedDatum.ID))
						Expect(datum.SchemaVersion).To(Equal(3))
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
				Entry("id missing",
					func(datum *types.Base) { datum.ID = nil },
					func(datum *types.Base, expectedDatum *types.Base) {
						sort.Strings(*expectedDatum.Tags)
					},
				),
				Entry("default schema version",
					func(datum *types.Base) { datum.SchemaVersion = 0 },
					func(datum *types.Base, expectedDatum *types.Base) {
						sort.Strings(*expectedDatum.Tags)
					},
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
				datum.UserID = pointer.FromString("")
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
				datum.DeviceID = pointer.FromString("")
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
				datum.Time = pointer.FromString("")
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
				userID := pointer.FromString(userTest.RandomID())
				datum.SetUserID(userID)
				Expect(datum.UserID).To(Equal(userID))
			})
		})

		Context("SetDataSetID", func() {
			It("sets the data set id", func() {
				dataSetID := pointer.FromString(dataTest.RandomSetID())
				datum.SetDataSetID(dataSetID)
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
				deviceID := pointer.FromString(testData.NewDeviceID())
				datum.SetDeviceID(deviceID)
				Expect(datum.DeviceID).To(Equal(deviceID))
			})
		})

		Context("SetCreatedTime", func() {
			It("sets the created time", func() {
				createdTime := pointer.FromString(time.Now().Format(time.RFC3339))
				datum.SetCreatedTime(createdTime)
				Expect(datum.CreatedTime).To(Equal(createdTime))
			})
		})

		Context("SetCreatedUserID", func() {
			It("sets the created user id", func() {
				createdUserID := pointer.FromString(userTest.RandomID())
				datum.SetCreatedUserID(createdUserID)
				Expect(datum.CreatedUserID).To(Equal(createdUserID))
			})
		})

		Context("SetModifiedTime", func() {
			It("sets the modified time", func() {
				modifiedTime := pointer.FromString(time.Now().Format(time.RFC3339))
				datum.SetModifiedTime(modifiedTime)
				Expect(datum.ModifiedTime).To(Equal(modifiedTime))
			})
		})

		Context("SetModifiedUserID", func() {
			It("sets the modified user id", func() {
				modifiedUserID := pointer.FromString(userTest.RandomID())
				datum.SetModifiedUserID(modifiedUserID)
				Expect(datum.ModifiedUserID).To(Equal(modifiedUserID))
			})
		})

		Context("SetDeletedTime", func() {
			It("sets the deleted time", func() {
				deletedTime := pointer.FromString(time.Now().Format(time.RFC3339))
				datum.SetDeletedTime(deletedTime)
				Expect(datum.DeletedTime).To(Equal(deletedTime))
			})
		})

		Context("SetDeletedUserID", func() {
			It("sets the deleted user id", func() {
				deletedUserID := pointer.FromString(userTest.RandomID())
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
