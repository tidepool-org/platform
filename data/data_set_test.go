package data_test

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("DataSet", func() {
	It("ComputerTimeFormat is expected", func() {
		Expect(data.ComputerTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("TimeFormat is expected", func() {
		Expect(data.TimeFormat).To(Equal(time.RFC3339Nano))
	})

	It("DeviceTimeFormat is expected", func() {
		Expect(data.DeviceTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("ClockDriftOffsetMaximum is expected", func() {
		Expect(data.ClockDriftOffsetMaximum).To(Equal(86400000))
	})

	It("ClockDriftOffsetMinimum is expected", func() {
		Expect(data.ClockDriftOffsetMinimum).To(Equal(-86400000))
	})

	It("DataSetTypeContinuous is expected", func() {
		Expect(data.DataSetTypeContinuous).To(Equal("continuous"))
	})

	It("DataSetTypeNormal is expected", func() {
		Expect(data.DataSetTypeNormal).To(Equal("normal"))
	})

	It("DataSetStateClosed is expected", func() {
		Expect(data.DataSetStateClosed).To(Equal("closed"))
	})

	It("DataSetStateOpen is expected", func() {
		Expect(data.DataSetStateOpen).To(Equal("open"))
	})

	It("DeviceTagBGM is expected", func() {
		Expect(data.DeviceTagBGM).To(Equal("bgm"))
	})

	It("DeviceTagCGM is expected", func() {
		Expect(data.DeviceTagCGM).To(Equal("cgm"))
	})

	It("DeviceTagInsulinPump is expected", func() {
		Expect(data.DeviceTagInsulinPump).To(Equal("insulin-pump"))
	})

	It("TimeProcessingAcrossTheBoardTimeZone is expected", func() {
		Expect(data.TimeProcessingAcrossTheBoardTimeZone).To(Equal("across-the-board-timezone"))
	})

	It("TimeProcessingNone is expected", func() {
		Expect(data.TimeProcessingNone).To(Equal("none"))
	})

	It("TimeProcessingUTCBootstrapping is expected", func() {
		Expect(data.TimeProcessingUTCBootstrapping).To(Equal("utc-bootstrapping"))
	})

	It("TimeZoneOffsetMaximum is expected", func() {
		Expect(data.TimeZoneOffsetMaximum).To(Equal(10080))
	})

	It("TimeZoneOffsetMinimum is expected", func() {
		Expect(data.TimeZoneOffsetMinimum).To(Equal(-10080))
	})

	It("VersionInternalMinimum is expected", func() {
		Expect(data.VersionInternalMinimum).To(Equal(0))
	})

	It("VersionLengthMinimum is expected", func() {
		Expect(data.VersionLengthMinimum).To(Equal(5))
	})

	It("DataSetTypes returns expected", func() {
		Expect(data.DataSetTypes()).To(Equal([]string{"continuous", "normal"}))
	})

	It("DataSetStates returns expected", func() {
		Expect(data.DataSetStates()).To(Equal([]string{"closed", "open"}))
	})

	It("DeviceTags returns expected", func() {
		Expect(data.DeviceTags()).To(Equal([]string{"bgm", "cgm", "insulin-pump"}))
	})

	It("TimeProcessings returns expected", func() {
		Expect(data.TimeProcessings()).To(Equal([]string{"across-the-board-timezone", "none", "utc-bootstrapping"}))
	})

	Context("DataSetClient", func() {
		Context("NewDataSetClient", func() {
			It("is successful", func() {
				Expect(data.NewDataSetClient()).To(Equal(&data.DataSetClient{}))
			})
		})

		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *data.DataSetClient), expectedErrors ...error) {
					datum := data.NewDataSetClient()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *data.DataSetClient) {},
				),
				Entry("name empty",
					func(datum *data.DataSetClient) { datum.Name = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
				),
				Entry("name valid",
					func(datum *data.DataSetClient) { datum.Name = pointer.FromString(netTest.RandomReverseDomain()) },
				),
				Entry("version empty",
					func(datum *data.DataSetClient) { datum.Version = pointer.FromString("") },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
				),
				Entry("version valid",
					func(datum *data.DataSetClient) { datum.Version = pointer.FromString(netTest.RandomSemanticVersion()) },
				),
				Entry("private missing",
					func(datum *data.DataSetClient) { datum.Private = nil },
				),
				Entry("private invalid",
					func(datum *data.DataSetClient) { datum.Private = map[string]any{} },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/private"),
				),
				Entry("private valid",
					func(datum *data.DataSetClient) { datum.Private = metadataTest.RandomMetadataMap() },
				),
				Entry("multiple errors",
					func(datum *data.DataSetClient) {
						datum.Name = pointer.FromString("")
						datum.Version = pointer.FromString("")
						datum.Private = map[string]any{}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/name"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/version"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/private"),
				),
			)
		})
	})

	Context("NewSetID", func() {
		It("returns a string of 32 lowercase hexadecimal characters", func() {
			Expect(data.NewSetID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(data.NewSetID()).ToNot(Equal(data.NewSetID()))
		})
	})

	Context("IsValidSetID, SetIDValidator, and ValidateSetID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(data.IsValidSetID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				data.SetIDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(data.ValidateSetID(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is first version with string length out of range (lower)", "upid_0123456789a", data.ErrorValueStringAsSetIDNotValid("upid_0123456789a")),
			Entry("is first version with string length in range", "upid_"+test.RandomStringFromRangeAndCharset(12, 12, test.CharsetHexidecimalLowercase)),
			Entry("is first version with uppercase characters", "upid_0123456789AB", data.ErrorValueStringAsSetIDNotValid("upid_0123456789AB")),
			Entry("is second version with string length in range", "upid_"+test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexidecimalLowercase)),
			Entry("is second version with uppercase characters", "upid_0123456789ABCDEF0123456789ABCDEF", data.ErrorValueStringAsSetIDNotValid("upid_0123456789ABCDEF0123456789ABCDEF")),
			Entry("is second version with string length out of range (upper)", "upid_0123456789abcdef0123456789abcdef0", data.ErrorValueStringAsSetIDNotValid("upid_0123456789abcdef0123456789abcdef0")),
			Entry("is third version with string length out of range (lower)", "0123456789abcdef0123456789abcde", data.ErrorValueStringAsSetIDNotValid("0123456789abcdef0123456789abcde")),
			Entry("is third version with string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetHexidecimalLowercase)),
			Entry("is third version with uppercase characters", "0123456789ABCDEF0123456789ABCDEF", data.ErrorValueStringAsSetIDNotValid("0123456789ABCDEF0123456789ABCDEF")),
			Entry("is third version with string length out of range (upper)", "0123456789abcdef0123456789abcdef0", data.ErrorValueStringAsSetIDNotValid("0123456789abcdef0123456789abcdef0")),
			Entry("has invalid prefix", "UPID_0123456789abcdef0123456789abcdef", data.ErrorValueStringAsSetIDNotValid("UPID_0123456789abcdef0123456789abcdef")),
			Entry("has symbols", "0123456789!@#$%^0123456789!@#$%^", data.ErrorValueStringAsSetIDNotValid("0123456789!@#$%^0123456789!@#$%^")),
			Entry("has whitespace", "0123456789      0123456789      ", data.ErrorValueStringAsSetIDNotValid("0123456789      0123456789      ")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsSetIDNotValid with empty string", data.ErrorValueStringAsSetIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as data set id`),
			Entry("is ErrorValueStringAsSetIDNotValid with non-empty string", data.ErrorValueStringAsSetIDNotValid("0123456789abcdefghijklmnopqrstuv"), "value-not-valid", "value is not valid", `value "0123456789abcdefghijklmnopqrstuv" is not valid as data set id`),
		)
	})

	Context("DataSet", func() {
		Context("New", func() {
			It("returns the expected datum with all values initialized", func() {
				datum := data.NewDataSet()
				Expect(datum.Active).To(BeFalse())
				Expect(datum.Annotations).To(BeNil())
				Expect(datum.ByUser).To(BeNil())
				Expect(datum.Client).To(BeNil())
				Expect(datum.ClockDriftOffset).To(BeNil())
				Expect(datum.ComputerTime).To(BeNil())
				Expect(datum.ConversionOffset).To(BeNil())
				Expect(datum.CreatedTime).To(BeNil())
				Expect(datum.CreatedUserID).To(BeNil())
				Expect(datum.DataSetType).ToNot(BeNil())
				Expect(*datum.DataSetType).To(Equal(data.DataSetTypeNormal))
				Expect(datum.DataState).To(BeNil())
				Expect(datum.Deduplicator).To(BeNil())
				Expect(datum.DeletedTime).To(BeNil())
				Expect(datum.DeletedUserID).To(BeNil())
				Expect(datum.DeviceID).To(BeNil())
				Expect(datum.DeviceManufacturers).To(BeNil())
				Expect(datum.DeviceModel).To(BeNil())
				Expect(datum.DeviceSerialNumber).To(BeNil())
				Expect(datum.DeviceTags).To(BeNil())
				Expect(datum.DeviceTime).To(BeNil())
				Expect(datum.ID).To(BeNil())
				Expect(datum.ModifiedTime).To(BeNil())
				Expect(datum.ModifiedUserID).To(BeNil())
				Expect(datum.Payload).To(BeNil())
				Expect(datum.Provenance).To(BeNil())
				Expect(datum.State).To(BeNil())
				Expect(datum.Time).To(BeNil())
				Expect(datum.TimeProcessing).To(BeNil())
				Expect(datum.TimeZoneName).To(BeNil())
				Expect(datum.TimeZoneOffset).To(BeNil())
				Expect(datum.Type).To(Equal("upload"))
				Expect(datum.UploadID).To(BeNil())
				Expect(datum.UserID).To(BeNil())
				Expect(datum.Version).To(BeNil())
				Expect(datum.VersionInternal).To(BeZero())

			})
		})

		Context("Upload", func() {
			Context("Parse", func() {
				// TODO
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *data.DataSet), expectedOrigins []structure.Origin, expectedErrors ...error) {
						datum := dataTest.RandomDataSet()
						mutator(datum)
						dataTypesTest.ValidateWithExpectedOrigins(datum, expectedOrigins, expectedErrors...)
					},
					Entry("succeeds",
						func(datum *data.DataSet) {},
						structure.Origins(),
					),
					Entry("type invalid",
						func(datum *data.DataSet) { datum.Type = "invalidType" },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "upload"), "/type"),
					),
					Entry("type upload",
						func(datum *data.DataSet) { datum.Type = "upload" },
						structure.Origins(),
					),
					Entry("by user missing",
						func(datum *data.DataSet) { datum.ByUser = nil },
						structure.Origins(),
					),
					Entry("by user empty",
						func(datum *data.DataSet) { datum.ByUser = pointer.FromString("") },
						structure.Origins(),
					),
					Entry("by user exists",
						func(datum *data.DataSet) { datum.ByUser = pointer.FromString(userTest.RandomID()) },
						structure.Origins(),
					),
					Entry("client missing",
						func(datum *data.DataSet) { datum.Client = nil },
						structure.Origins(),
					),
					Entry("client invalid",
						func(datum *data.DataSet) {
							datum.Client.Name = pointer.FromString("")
							datum.Client.Version = pointer.FromString("")
						},
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/client/name"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/client/version"),
					),
					Entry("client valid",
						func(datum *data.DataSet) { datum.Client = data.NewDataSetClient() },
						structure.Origins(),
					),
					Entry("computer time missing",
						func(datum *data.DataSet) { datum.ComputerTime = nil },
						structure.Origins(),
					),
					Entry("computer time invalid",
						func(datum *data.DataSet) { datum.ComputerTime = pointer.FromString("invalid") },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/computerTime"),
					),
					Entry("computer time valid",
						func(datum *data.DataSet) {
							datum.ComputerTime = pointer.FromString(test.RandomTime().Format("2006-01-02T15:04:05"))
						},
						structure.Origins(),
					),
					Entry("data set type missing",
						func(datum *data.DataSet) { datum.DataSetType = nil },
						structure.Origins(),
					),
					Entry("data set type invalid",
						func(datum *data.DataSet) { datum.DataSetType = pointer.FromString("invalid") },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"continuous", "normal"}), "/dataSetType"),
					),
					Entry("data set type normal",
						func(datum *data.DataSet) { datum.DataSetType = pointer.FromString("normal") },
						structure.Origins(),
					),
					Entry("data set type continuous",
						func(datum *data.DataSet) { datum.DataSetType = pointer.FromString("continuous") },
						structure.Origins(),
					),
					Entry("data state missing",
						func(datum *data.DataSet) { datum.DataState = nil },
						[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					),
					Entry("data state invalid",
						func(datum *data.DataSet) { datum.DataState = pointer.FromString("invalid") },
						[]structure.Origin{structure.OriginInternal, structure.OriginStore},
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/dataState"),
					),
					Entry("data state open",
						func(datum *data.DataSet) { datum.DataState = pointer.FromString("open") },
						structure.Origins(),
					),
					Entry("data state closed",
						func(datum *data.DataSet) { datum.DataState = pointer.FromString("closed") },
						structure.Origins(),
					),
					Entry("device manufacturers missing",
						func(datum *data.DataSet) { datum.DeviceManufacturers = nil },
						structure.Origins(),
					),
					Entry("device manufacturers empty",
						func(datum *data.DataSet) { datum.DeviceManufacturers = pointer.FromStringArray([]string{}) },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceManufacturers"),
					),
					Entry("device manufacturers element empty",
						func(datum *data.DataSet) {
							datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), ""})
						},
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceManufacturers/1"),
					),
					Entry("device manufacturers single",
						func(datum *data.DataSet) {
							datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16)})
						},
						structure.Origins(),
					),
					Entry("device manufacturers multiple",
						func(datum *data.DataSet) {
							datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), test.RandomStringFromRange(1, 16)})
						},
						structure.Origins(),
					),
					Entry("device manufacturers multiple duplicates",
						func(datum *data.DataSet) {
							duplicate := test.RandomStringFromRange(1, 16)
							datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), duplicate, duplicate, test.RandomStringFromRange(1, 16)})
						},
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/deviceManufacturers/2"),
					),
					Entry("device model missing",
						func(datum *data.DataSet) { datum.DeviceModel = nil },
						structure.Origins(),
					),
					Entry("device model empty",
						func(datum *data.DataSet) { datum.DeviceModel = pointer.FromString("") },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceModel"),
					),
					Entry("device model exists",
						func(datum *data.DataSet) {
							datum.DeviceModel = pointer.FromString(test.RandomStringFromRange(1, 32))
						},
						structure.Origins(),
					),
					Entry("device serial number missing",
						func(datum *data.DataSet) { datum.DeviceSerialNumber = nil },
						structure.Origins(),
					),
					Entry("device serial number empty",
						func(datum *data.DataSet) { datum.DeviceSerialNumber = pointer.FromString("") },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceSerialNumber"),
					),
					Entry("device serial number exists",
						func(datum *data.DataSet) {
							datum.DeviceSerialNumber = pointer.FromString(test.RandomStringFromRange(1, 16))
						},
						structure.Origins(),
					),
					Entry("device tags missing",
						func(datum *data.DataSet) { datum.DeviceTags = nil },
						structure.Origins(),
					),
					Entry("device tags empty",
						func(datum *data.DataSet) { datum.DeviceTags = pointer.FromStringArray([]string{}) },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceTags"),
					),
					Entry("device tags elements single invalid",
						func(datum *data.DataSet) { datum.DeviceTags = pointer.FromStringArray([]string{"invalid"}) },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"bgm", "cgm", "insulin-pump"}), "/deviceTags/0"),
					),
					Entry("device tags elements single bgm",
						func(datum *data.DataSet) { datum.DeviceTags = pointer.FromStringArray([]string{"bgm"}) },
						structure.Origins(),
					),
					Entry("device tags elements single cgm",
						func(datum *data.DataSet) { datum.DeviceTags = pointer.FromStringArray([]string{"cgm"}) },
						structure.Origins(),
					),
					Entry("device tags elements single insulin-pump",
						func(datum *data.DataSet) {
							datum.DeviceTags = pointer.FromStringArray([]string{"insulin-pump"})
						},
						structure.Origins(),
					),
					Entry("device tags elements multiple invalid",
						func(datum *data.DataSet) {
							datum.DeviceTags = pointer.FromStringArray([]string{"bgm", "invalid", "insulin-pump"})
						},
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"bgm", "cgm", "insulin-pump"}), "/deviceTags/1"),
					),
					Entry("device tags elements multiple valid",
						func(datum *data.DataSet) {
							datum.DeviceTags = pointer.FromStringArray([]string{"cgm", "insulin-pump"})
						},
						structure.Origins(),
					),
					Entry("device tags elements multiple valid duplicates",
						func(datum *data.DataSet) {
							datum.DeviceTags = pointer.FromStringArray([]string{"cgm", "insulin-pump", "cgm", "insulin-pump"})
						},
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/deviceTags/2"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/deviceTags/3"),
					),
					Entry("state missing",
						func(datum *data.DataSet) { datum.State = nil },
						[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					),
					Entry("state invalid",
						func(datum *data.DataSet) { datum.State = pointer.FromString("invalid") },
						[]structure.Origin{structure.OriginInternal, structure.OriginStore},
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/state"),
					),
					Entry("state open",
						func(datum *data.DataSet) { datum.State = pointer.FromString("open") },
						structure.Origins(),
					),
					Entry("state closed",
						func(datum *data.DataSet) { datum.State = pointer.FromString("closed") },
						structure.Origins(),
					),
					Entry("time processing missing",
						func(datum *data.DataSet) { datum.TimeProcessing = nil },
						structure.Origins(),
					),
					Entry("time processing invalid",
						func(datum *data.DataSet) { datum.TimeProcessing = pointer.FromString("invalid") },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"across-the-board-timezone", "none", "utc-bootstrapping"}), "/timeProcessing"),
					),
					Entry("time processing across-the-board-timezone",
						func(datum *data.DataSet) {
							datum.TimeProcessing = pointer.FromString("across-the-board-timezone")
						},
						structure.Origins(),
					),
					Entry("time processing none",
						func(datum *data.DataSet) { datum.TimeProcessing = pointer.FromString("none") },
						structure.Origins(),
					),
					Entry("time processing utc-bootstrapping",
						func(datum *data.DataSet) { datum.TimeProcessing = pointer.FromString("utc-bootstrapping") },
						structure.Origins(),
					),
					Entry("version missing",
						func(datum *data.DataSet) { datum.Version = nil },
						structure.Origins(),
					),
					Entry("version out of range (lower)",
						func(datum *data.DataSet) { datum.Version = pointer.FromString("1.23") },
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(4, 5), "/version"),
					),
					Entry("version in range (lower)",
						func(datum *data.DataSet) {
							datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
						},
						structure.Origins(),
					),
					Entry("multiple errors with store origin",
						func(datum *data.DataSet) {},
						[]structure.Origin{structure.OriginStore},
					),
					Entry("multiple errors with internal origin",
						func(datum *data.DataSet) {
							datum.DataState = pointer.FromString("invalid")
							datum.State = pointer.FromString("invalid")
						},
						[]structure.Origin{structure.OriginInternal, structure.OriginStore},
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/dataState"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/state"),
					),
					Entry("multiple errors with external origin",
						func(datum *data.DataSet) {
							datum.Client.Name = pointer.FromString("")
							datum.Client.Version = pointer.FromString("")
							datum.ComputerTime = pointer.FromString("invalid")
							datum.DataSetType = pointer.FromString("invalid")
							datum.DeviceManufacturers = pointer.FromStringArray([]string{})
							datum.DeviceModel = pointer.FromString("")
							datum.DeviceSerialNumber = pointer.FromString("")
							datum.DeviceTags = pointer.FromStringArray([]string{})
							datum.TimeProcessing = pointer.FromString("invalid")
							datum.Type = "invalidType"
							datum.Version = pointer.FromString("1.23")
						},
						structure.Origins(),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/client/name"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/client/version"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/computerTime"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"continuous", "normal"}), "/dataSetType"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceManufacturers"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceModel"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceSerialNumber"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/deviceTags"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"across-the-board-timezone", "none", "utc-bootstrapping"}), "/timeProcessing"),
						errorsTest.WithPointerSource(structureValidator.ErrorValueNotEqualTo("invalidType", "upload"), "/type"),
						errorsTest.WithPointerSource(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(4, 5), "/version"),
					),
				)
			})

			Context("Normalize", func() {
				DescribeTable("normalizes the datum with origin external",
					func(mutator func(datum *data.DataSet), expectator func(datum *data.DataSet, expectedDatum *data.DataSet)) {
						datum := dataTest.RandomDataSet()
						mutator(datum)
						expectedDatum := dataTest.CloneDataSet(datum)
						normalizer := structureNormalizer.New(logTest.NewLogger())
						Expect(normalizer).ToNot(BeNil())
						datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
						Expect(normalizer.Error()).To(BeNil())
						if expectator != nil {
							expectator(datum, expectedDatum)
						}
						Expect(datum).To(Equal(expectedDatum))
					},
					Entry("does not modify the datum",
						func(datum *data.DataSet) {},
						func(datum *data.DataSet, expectedDatum *data.DataSet) {
							Expect(datum.DataSetType).ToNot(BeNil())
							sort.Strings(*expectedDatum.DeviceManufacturers)
							sort.Strings(*expectedDatum.DeviceTags)
						},
					),
					Entry("upload id missing",
						func(datum *data.DataSet) { datum.UploadID = nil },
						func(datum *data.DataSet, expectedDatum *data.DataSet) {
							Expect(datum.UploadID).ToNot(BeNil())
							Expect(*datum.UploadID).To(Equal(*datum.ID))
							expectedDatum.UploadID = datum.UploadID
							sort.Strings(*expectedDatum.DeviceManufacturers)
							sort.Strings(*expectedDatum.DeviceTags)
						},
					),
					Entry("data set type missing",
						func(datum *data.DataSet) { datum.DataSetType = nil },
						func(datum *data.DataSet, expectedDatum *data.DataSet) {
							Expect(datum.DataSetType).ToNot(BeNil())
							Expect(*datum.DataSetType).To(Equal(data.DataSetTypeNormal))
							expectedDatum.DataSetType = datum.DataSetType
							sort.Strings(*expectedDatum.DeviceManufacturers)
							sort.Strings(*expectedDatum.DeviceTags)
						},
					),
					Entry("all missing",
						func(datum *data.DataSet) {
							*datum = *data.NewDataSet()
						},
						func(datum *data.DataSet, expectedDatum *data.DataSet) {
							Expect(datum.DataSetType).ToNot(BeNil())
							Expect(*datum.DataSetType).To(Equal(data.DataSetTypeNormal))
							expectedDatum.ID = datum.ID
							expectedDatum.DataSetType = datum.DataSetType
							expectedDatum.UploadID = datum.UploadID
						},
					),
				)

				DescribeTable("normalizes the datum with origin internal/store",
					func(mutator func(datum *data.DataSet), expectator func(datum *data.DataSet, expectedDatum *data.DataSet)) {
						for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
							datum := dataTest.RandomDataSet()
							mutator(datum)
							expectedDatum := dataTest.CloneDataSet(datum)
							normalizer := structureNormalizer.New(logTest.NewLogger())
							Expect(normalizer).ToNot(BeNil())
							datum.Normalize(normalizer.WithOrigin(origin))
							Expect(normalizer.Error()).To(BeNil())
							if expectator != nil {
								expectator(datum, expectedDatum)
							}
							Expect(datum).To(Equal(expectedDatum))
						}
					},
					Entry("does not modify the datum",
						func(datum *data.DataSet) {},
						nil,
					),
					Entry("data set type missing",
						func(datum *data.DataSet) { datum.DataSetType = nil },
						nil,
					),
					Entry("all missing",
						func(datum *data.DataSet) {
							*datum = *data.NewDataSet()
						},
						nil,
					),
				)
			})

			Context("with new upload", func() {
				var datum *data.DataSet

				BeforeEach(func() {
					datum = dataTest.RandomDataSet()
				})

				Context("HasDataSetTypeContinuous", func() {
					It("returns false if the data set type is missing", func() {
						datum.DataSetType = nil
						Expect(datum.HasDataSetTypeContinuous()).To(BeFalse())
					})

					It("returns true if the data set type is continuous", func() {
						datum.DataSetType = pointer.FromString("continuous")
						Expect(datum.HasDataSetTypeContinuous()).To(BeTrue())
					})

					It("returns false if the data set type is normal", func() {
						datum.DataSetType = pointer.FromString("normal")
						Expect(datum.HasDataSetTypeContinuous()).To(BeFalse())
					})
				})

				Context("HasDataSetTypeNormal", func() {
					It("returns true if the data set type is missing", func() {
						datum.DataSetType = nil
						Expect(datum.HasDataSetTypeNormal()).To(BeTrue())
					})

					It("returns false if the data set type is continuous", func() {
						datum.DataSetType = pointer.FromString("continuous")
						Expect(datum.HasDataSetTypeNormal()).To(BeFalse())
					})

					It("returns true if the data set type is normal", func() {
						datum.DataSetType = pointer.FromString("normal")
						Expect(datum.HasDataSetTypeNormal()).To(BeTrue())
					})
				})

				Context("HasDeduplicatorName", func() {
					It("returns false if the deduplicator is missing", func() {
						datum.Deduplicator = nil
						Expect(datum.HasDeduplicatorName()).To(BeFalse())
					})

					It("returns false if the deduplicator name is missing", func() {
						datum.Deduplicator.Name = nil
						Expect(datum.HasDeduplicatorName()).To(BeFalse())
					})

					It("returns true if the deduplicator name is empty", func() {
						datum.Deduplicator.Name = pointer.FromString("")
						Expect(datum.HasDeduplicatorName()).To(BeTrue())
					})

					It("returns true if the deduplicator name exists", func() {
						datum.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
						Expect(datum.HasDeduplicatorName()).To(BeTrue())
					})
				})

				Context("HasDeduplicatorNameMatch", func() {
					var name string

					BeforeEach(func() {
						name = netTest.RandomReverseDomain()
					})

					It("returns false if the deduplicator is missing", func() {
						datum.Deduplicator = nil
						Expect(datum.HasDeduplicatorNameMatch(name)).To(BeFalse())
					})

					It("returns false if the deduplicator name is missing", func() {
						datum.Deduplicator.Name = nil
						Expect(datum.HasDeduplicatorNameMatch(name)).To(BeFalse())
					})

					It("returns false if the deduplicator name is empty", func() {
						datum.Deduplicator.Name = pointer.FromString("")
						Expect(datum.HasDeduplicatorNameMatch(name)).To(BeFalse())
					})

					It("returns false if the deduplicator name does not match", func() {
						datum.Deduplicator.Name = pointer.FromString(netTest.RandomReverseDomain())
						Expect(datum.HasDeduplicatorNameMatch(name)).To(BeFalse())
					})

					It("returns true if the deduplicator name matches", func() {
						datum.Deduplicator.Name = pointer.FromString(name)
						Expect(datum.HasDeduplicatorNameMatch(name)).To(BeTrue())
					})
				})
			})
		})
	})
})
