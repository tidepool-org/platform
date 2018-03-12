package upload_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types"
	testDataTypes "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/data/types/upload"
	testErrors "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testInternet "github.com/tidepool-org/platform/test/internet"
)

func NewMeta() interface{} {
	return &types.Meta{
		Type: "upload",
	}
}

func NewUpload() *upload.Upload {
	datum := upload.New()
	datum.Base = *testDataTypes.NewBase()
	datum.Type = "upload"
	datum.ByUser = pointer.String(id.New())
	datum.Client = NewClient()
	datum.ComputerTime = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
	datum.DataSetType = pointer.String(test.RandomStringFromArray(upload.DataSetTypes()))
	datum.DataState = pointer.String(test.RandomStringFromArray(upload.States()))
	datum.DeviceManufacturers = pointer.StringArray([]string{test.NewText(1, 16), test.NewText(1, 16)})
	datum.DeviceModel = pointer.String(test.NewText(1, 32))
	datum.DeviceSerialNumber = pointer.String(test.NewText(1, 16))
	datum.DeviceTags = pointer.StringArray(test.RandomStringArrayFromArray(1, len(upload.DeviceTags()), true, upload.DeviceTags()))
	datum.State = pointer.String(test.RandomStringFromArray(upload.States()))
	datum.TimeProcessing = pointer.String(upload.TimeProcessingUTCBootstrapping)
	datum.Timezone = pointer.String(test.NewTimeZone())
	datum.Version = pointer.String(testInternet.NewSemanticVersion())
	return datum
}

func CloneUpload(datum *upload.Upload) *upload.Upload {
	if datum == nil {
		return nil
	}
	clone := upload.New()
	clone.Base = *testDataTypes.CloneBase(&datum.Base)
	clone.ByUser = test.CloneString(datum.ByUser)
	clone.Client = CloneClient(datum.Client)
	clone.ComputerTime = test.CloneString(datum.ComputerTime)
	clone.DataSetType = test.CloneString(datum.DataSetType)
	clone.DataState = test.CloneString(datum.DataState)
	clone.DeviceManufacturers = test.CloneStringArray(datum.DeviceManufacturers)
	clone.DeviceModel = test.CloneString(datum.DeviceModel)
	clone.DeviceSerialNumber = test.CloneString(datum.DeviceSerialNumber)
	clone.DeviceTags = test.CloneStringArray(datum.DeviceTags)
	clone.State = test.CloneString(datum.State)
	clone.TimeProcessing = test.CloneString(datum.TimeProcessing)
	clone.Timezone = test.CloneString(datum.Timezone)
	clone.Version = test.CloneString(datum.Version)
	return clone
}

var _ = Describe("Upload", func() {
	It("Type is expected", func() {
		Expect(upload.Type).To(Equal("upload"))
	})

	It("ComputerTimeFormat is expected", func() {
		Expect(upload.ComputerTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("DataSetTypeContinuous is expected", func() {
		Expect(upload.DataSetTypeContinuous).To(Equal("continuous"))
	})

	It("DataSetTypeNormal is expected", func() {
		Expect(upload.DataSetTypeNormal).To(Equal("normal"))
	})

	It("DeviceTagBGM is expected", func() {
		Expect(upload.DeviceTagBGM).To(Equal("bgm"))
	})

	It("DeviceTagCGM is expected", func() {
		Expect(upload.DeviceTagCGM).To(Equal("cgm"))
	})

	It("DeviceTagInsulinPump is expected", func() {
		Expect(upload.DeviceTagInsulinPump).To(Equal("insulin-pump"))
	})

	It("StateClosed is expected", func() {
		Expect(upload.StateClosed).To(Equal("closed"))
	})

	It("StateOpen is expected", func() {
		Expect(upload.StateOpen).To(Equal("open"))
	})

	It("TimeProcessingAcrossTheBoardTimezone is expected", func() {
		Expect(upload.TimeProcessingAcrossTheBoardTimezone).To(Equal("across-the-board-timezone"))
	})

	It("TimeProcessingNone is expected", func() {
		Expect(upload.TimeProcessingNone).To(Equal("none"))
	})

	It("TimeProcessingUTCBootstrapping is expected", func() {
		Expect(upload.TimeProcessingUTCBootstrapping).To(Equal("utc-bootstrapping"))
	})

	It("VersionLengthMinimum is expected", func() {
		Expect(upload.VersionLengthMinimum).To(Equal(5))
	})

	It("DataSetTypes returns expected", func() {
		Expect(upload.DataSetTypes()).To(Equal([]string{"continuous", "normal"}))
	})

	It("DeviceTags returns expected", func() {
		Expect(upload.DeviceTags()).To(Equal([]string{"bgm", "cgm", "insulin-pump"}))
	})

	It("States returns expected", func() {
		Expect(upload.States()).To(Equal([]string{"closed", "open"}))
	})

	It("TimeProcessings returns expected", func() {
		Expect(upload.TimeProcessings()).To(Equal([]string{"across-the-board-timezone", "none", "utc-bootstrapping"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := upload.New()
			Expect(datum).ToNot(BeNil())
			Expect(datum.Type).To(Equal("upload"))
			Expect(datum.ByUser).To(BeNil())
			Expect(datum.Client).To(BeNil())
			Expect(datum.ComputerTime).To(BeNil())
			Expect(datum.DataSetType).To(BeNil())
			Expect(datum.DataState).To(BeNil())
			Expect(datum.DeviceManufacturers).To(BeNil())
			Expect(datum.DeviceModel).To(BeNil())
			Expect(datum.DeviceSerialNumber).To(BeNil())
			Expect(datum.DeviceTags).To(BeNil())
			Expect(datum.State).To(BeNil())
			Expect(datum.TimeProcessing).To(BeNil())
			Expect(datum.Timezone).To(BeNil())
			Expect(datum.Version).To(BeNil())
		})
	})

	Context("Upload", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *upload.Upload), expectedOrigins []structure.Origin, expectedErrors ...error) {
					datum := NewUpload()
					mutator(datum)
					testDataTypes.ValidateWithExpectedOrigins(datum, expectedOrigins, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *upload.Upload) {},
					structure.Origins(),
				),
				Entry("type missing",
					func(datum *upload.Upload) { datum.Type = "" },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &types.Meta{}),
				),
				Entry("type invalid",
					func(datum *upload.Upload) { datum.Type = "invalidType" },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "upload"), "/type", &types.Meta{Type: "invalidType"}),
				),
				Entry("type upload",
					func(datum *upload.Upload) { datum.Type = "upload" },
					structure.Origins(),
				),
				Entry("by user missing",
					func(datum *upload.Upload) { datum.ByUser = nil },
					structure.Origins(),
				),
				Entry("by user empty",
					func(datum *upload.Upload) { datum.ByUser = pointer.String("") },
					structure.Origins(),
				),
				Entry("by user exists",
					func(datum *upload.Upload) { datum.ByUser = pointer.String(id.New()) },
					structure.Origins(),
				),
				Entry("client missing",
					func(datum *upload.Upload) { datum.Client = nil },
					structure.Origins(),
				),
				Entry("client invalid",
					func(datum *upload.Upload) {
						datum.Client.Name = nil
						datum.Client.Version = nil
					},
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/name", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/version", NewMeta()),
				),
				Entry("client valid",
					func(datum *upload.Upload) { datum.Client = NewClient() },
					structure.Origins(),
				),
				Entry("computer time missing",
					func(datum *upload.Upload) { datum.ComputerTime = nil },
					structure.Origins(),
				),
				Entry("computer time invalid",
					func(datum *upload.Upload) { datum.ComputerTime = pointer.String("invalid") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/computerTime", NewMeta()),
				),
				Entry("computer time valid",
					func(datum *upload.Upload) {
						datum.ComputerTime = pointer.String(test.NewTime().Format("2006-01-02T15:04:05"))
					},
					structure.Origins(),
				),
				Entry("data set type missing",
					func(datum *upload.Upload) { datum.DataSetType = nil },
					structure.Origins(),
				),
				Entry("data set type invalid",
					func(datum *upload.Upload) { datum.DataSetType = pointer.String("invalid") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"continuous", "normal"}), "/dataSetType", NewMeta()),
				),
				Entry("data set type normal",
					func(datum *upload.Upload) { datum.DataSetType = pointer.String("normal") },
					structure.Origins(),
				),
				Entry("data set type continuous",
					func(datum *upload.Upload) { datum.DataSetType = pointer.String("continuous") },
					structure.Origins(),
				),
				Entry("data state missing",
					func(datum *upload.Upload) { datum.DataState = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
				),
				Entry("data state invalid",
					func(datum *upload.Upload) { datum.DataState = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/dataState", NewMeta()),
				),
				Entry("data state open",
					func(datum *upload.Upload) { datum.DataState = pointer.String("open") },
					structure.Origins(),
				),
				Entry("data state closed",
					func(datum *upload.Upload) { datum.DataState = pointer.String("closed") },
					structure.Origins(),
				),
				Entry("device manufacturers missing",
					func(datum *upload.Upload) { datum.DeviceManufacturers = nil },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceManufacturers", NewMeta()),
				),
				Entry("device manufacturers empty",
					func(datum *upload.Upload) { datum.DeviceManufacturers = pointer.StringArray([]string{}) },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceManufacturers", NewMeta()),
				),
				Entry("device manufacturers element empty",
					func(datum *upload.Upload) {
						datum.DeviceManufacturers = pointer.StringArray([]string{test.NewText(1, 16), ""})
					},
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceManufacturers/1", NewMeta()),
				),
				Entry("device manufacturers single",
					func(datum *upload.Upload) {
						datum.DeviceManufacturers = pointer.StringArray([]string{test.NewText(1, 16)})
					},
					structure.Origins(),
				),
				Entry("device manufacturers multiple",
					func(datum *upload.Upload) {
						datum.DeviceManufacturers = pointer.StringArray([]string{test.NewText(1, 16), test.NewText(1, 16)})
					},
					structure.Origins(),
				),
				Entry("device model missing",
					func(datum *upload.Upload) { datum.DeviceModel = nil },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceModel", NewMeta()),
				),
				Entry("device model empty",
					func(datum *upload.Upload) { datum.DeviceModel = pointer.String("") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceModel", NewMeta()),
				),
				Entry("device model exists",
					func(datum *upload.Upload) { datum.DeviceModel = pointer.String(test.NewText(1, 32)) },
					structure.Origins(),
				),
				Entry("device serial number missing",
					func(datum *upload.Upload) { datum.DeviceSerialNumber = nil },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceSerialNumber", NewMeta()),
				),
				Entry("device serial number empty",
					func(datum *upload.Upload) { datum.DeviceSerialNumber = pointer.String("") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceSerialNumber", NewMeta()),
				),
				Entry("device serial number exists",
					func(datum *upload.Upload) { datum.DeviceSerialNumber = pointer.String(test.NewText(1, 16)) },
					structure.Origins(),
				),
				Entry("device tags missing",
					func(datum *upload.Upload) { datum.DeviceTags = nil },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceTags", NewMeta()),
				),
				Entry("device tags empty",
					func(datum *upload.Upload) { datum.DeviceTags = pointer.StringArray([]string{}) },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceTags", NewMeta()),
				),
				Entry("device tags elements single invalid",
					func(datum *upload.Upload) { datum.DeviceTags = pointer.StringArray([]string{"invalid"}) },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"bgm", "cgm", "insulin-pump"}), "/deviceTags/0", NewMeta()),
				),
				Entry("device tags elements single bgm",
					func(datum *upload.Upload) { datum.DeviceTags = pointer.StringArray([]string{"bgm"}) },
					structure.Origins(),
				),
				Entry("device tags elements single cgm",
					func(datum *upload.Upload) { datum.DeviceTags = pointer.StringArray([]string{"cgm"}) },
					structure.Origins(),
				),
				Entry("device tags elements single insulin-pump",
					func(datum *upload.Upload) { datum.DeviceTags = pointer.StringArray([]string{"insulin-pump"}) },
					structure.Origins(),
				),
				Entry("device tags elements multiple invalid",
					func(datum *upload.Upload) {
						datum.DeviceTags = pointer.StringArray([]string{"bgm", "invalid", "insulin-pump"})
					},
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"bgm", "cgm", "insulin-pump"}), "/deviceTags/1", NewMeta()),
				),
				Entry("device tags elements multiple valid",
					func(datum *upload.Upload) { datum.DeviceTags = pointer.StringArray([]string{"cgm", "insulin-pump"}) },
					structure.Origins(),
				),
				Entry("state missing",
					func(datum *upload.Upload) { datum.State = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
				),
				Entry("state invalid",
					func(datum *upload.Upload) { datum.State = pointer.String("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/state", NewMeta()),
				),
				Entry("state open",
					func(datum *upload.Upload) { datum.State = pointer.String("open") },
					structure.Origins(),
				),
				Entry("state closed",
					func(datum *upload.Upload) { datum.State = pointer.String("closed") },
					structure.Origins(),
				),
				Entry("time processing missing",
					func(datum *upload.Upload) { datum.TimeProcessing = nil },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/timeProcessing", NewMeta()),
				),
				Entry("time processing empty",
					func(datum *upload.Upload) { datum.TimeProcessing = pointer.String("") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("", []string{"across-the-board-timezone", "none", "utc-bootstrapping"}), "/timeProcessing", NewMeta()),
				),
				Entry("time processing invalid",
					func(datum *upload.Upload) { datum.TimeProcessing = pointer.String("invalid") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"across-the-board-timezone", "none", "utc-bootstrapping"}), "/timeProcessing", NewMeta()),
				),
				Entry("time processing across-the-board-timezone",
					func(datum *upload.Upload) { datum.TimeProcessing = pointer.String("across-the-board-timezone") },
					structure.Origins(),
				),
				Entry("time processing none",
					func(datum *upload.Upload) { datum.TimeProcessing = pointer.String("none") },
					structure.Origins(),
				),
				Entry("time processing utc-bootstrapping",
					func(datum *upload.Upload) { datum.TimeProcessing = pointer.String("utc-bootstrapping") },
					structure.Origins(),
				),
				Entry("timezone missing",
					func(datum *upload.Upload) { datum.Timezone = nil },
					structure.Origins(),
				),
				Entry("timezone empty",
					func(datum *upload.Upload) { datum.Timezone = pointer.String("") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/timezone", NewMeta()),
				),
				Entry("timezone exists",
					func(datum *upload.Upload) { datum.Timezone = pointer.String(test.NewTimeZone()) },
					structure.Origins(),
				),
				Entry("version missing",
					func(datum *upload.Upload) { datum.Version = nil },
					structure.Origins(),
				),
				Entry("version out of range (lower)",
					func(datum *upload.Upload) { datum.Version = pointer.String("1.23") },
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(4, 5), "/version", NewMeta()),
				),
				Entry("version in range (lower)",
					func(datum *upload.Upload) { datum.Version = pointer.String(testInternet.NewSemanticVersion()) },
					structure.Origins(),
				),
				Entry("multiple errors with store origin",
					func(datum *upload.Upload) {},
					[]structure.Origin{structure.OriginStore},
				),
				Entry("multiple errors with internal origin",
					func(datum *upload.Upload) {
						datum.DataState = pointer.String("invalid")
						datum.State = pointer.String("invalid")
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/dataState", NewMeta()),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/state", NewMeta()),
				),
				Entry("multiple errors with external origin",
					func(datum *upload.Upload) {
						datum.Type = "invalidType"
						datum.Client.Name = nil
						datum.Client.Version = nil
						datum.ComputerTime = pointer.String("invalid")
						datum.DataSetType = pointer.String("invalid")
						datum.DeviceManufacturers = nil
						datum.DeviceModel = nil
						datum.DeviceSerialNumber = nil
						datum.DeviceTags = nil
						datum.TimeProcessing = nil
						datum.Timezone = pointer.String("")
						datum.Version = pointer.String("1.23")
					},
					structure.Origins(),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "upload"), "/type", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/name", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/version", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringAsTimeNotValid("invalid", "2006-01-02T15:04:05"), "/computerTime", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"continuous", "normal"}), "/dataSetType", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceManufacturers", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceModel", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceSerialNumber", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/deviceTags", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/timeProcessing", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/timezone", &types.Meta{Type: "invalidType"}),
					testErrors.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(4, 5), "/version", &types.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *upload.Upload), expectator func(datum *upload.Upload, expectedDatum *upload.Upload)) {
					datum := NewUpload()
					mutator(datum)
					expectedDatum := CloneUpload(datum)
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
				Entry("does not modify the datum",
					func(datum *upload.Upload) {},
					func(datum *upload.Upload, expectedDatum *upload.Upload) {
						Expect(datum.DataSetType).ToNot(BeNil())
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceManufacturers)
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceTags)
					},
				),
				Entry("upload id missing",
					func(datum *upload.Upload) { datum.UploadID = nil },
					func(datum *upload.Upload, expectedDatum *upload.Upload) {
						Expect(datum.UploadID).ToNot(BeNil())
						Expect(*datum.UploadID).ToNot(BeEmpty())
						expectedDatum.UploadID = datum.UploadID
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceManufacturers)
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceTags)
					},
				),
				Entry("data set type missing",
					func(datum *upload.Upload) { datum.DataSetType = nil },
					func(datum *upload.Upload, expectedDatum *upload.Upload) {
						Expect(datum.DataSetType).ToNot(BeNil())
						Expect(*datum.DataSetType).To(Equal(upload.DataSetTypeNormal))
						expectedDatum.DataSetType = datum.DataSetType
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceManufacturers)
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceTags)
					},
				),
				Entry("all missing",
					func(datum *upload.Upload) {
						*datum = *upload.New()
						datum.Base = *testDataTypes.NewBase()
					},
					func(datum *upload.Upload, expectedDatum *upload.Upload) {
						Expect(datum.DataSetType).ToNot(BeNil())
						Expect(*datum.DataSetType).To(Equal(upload.DataSetTypeNormal))
						expectedDatum.DataSetType = datum.DataSetType
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceManufacturers)
						upload.SortAndDeduplicateStringArray(expectedDatum.DeviceTags)
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *upload.Upload), expectator func(datum *upload.Upload, expectedDatum *upload.Upload)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := NewUpload()
						mutator(datum)
						expectedDatum := CloneUpload(datum)
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
					func(datum *upload.Upload) {},
					nil,
				),
				Entry("data set type missing",
					func(datum *upload.Upload) { datum.DataSetType = nil },
					nil,
				),
				Entry("all missing",
					func(datum *upload.Upload) {
						*datum = *upload.New()
						datum.Base = *testDataTypes.NewBase()
					},
					nil,
				),
			)
		})

		Context("HasDeviceManufacturerOneOf", func() {
			var datum *upload.Upload

			BeforeEach(func() {
				datum = NewUpload()
			})

			It("returns false if datum device manufacturers is nil", func() {
				datum.DeviceManufacturers = nil
				Expect(datum.HasDeviceManufacturerOneOf([]string{"one", "two", "three"})).To(BeFalse())
			})

			DescribeTable("returns expected result",
				func(datumDeviceManufacturers []string, deviceManufacturers []string, expectedResult bool) {
					datum.DeviceManufacturers = pointer.StringArray(datumDeviceManufacturers)
					Expect(datum.HasDeviceManufacturerOneOf(deviceManufacturers)).To(Equal(expectedResult))
				},
				Entry("is nil datum array with nil search array", nil, nil, false),
				Entry("is nil datum array with empty search array", nil, []string{}, false),
				Entry("is nil datum array with single invalid search array", nil, []string{"one"}, false),
				Entry("is nil datum array with multiple invalid search array", nil, []string{"one", "three"}, false),
				Entry("is empty datum array with nil search array", []string{}, nil, false),
				Entry("is empty datum array with empty search array", []string{}, []string{}, false),
				Entry("is empty datum array with single invalid search array", []string{}, []string{"one"}, false),
				Entry("is empty datum array with multiple invalid search array", []string{}, []string{"one", "three"}, false),
				Entry("is single datum array with nil search array", []string{"two"}, nil, false),
				Entry("is single datum array with single search array", []string{"two"}, []string{}, false),
				Entry("is single datum array with single invalid search array", []string{"two"}, []string{"one"}, false),
				Entry("is single datum array with single valid search array", []string{"two"}, []string{"two"}, true),
				Entry("is single datum array with multiple invalid search array", []string{"two"}, []string{"one", "three"}, false),
				Entry("is single datum array with multiple invalid and valid search array", []string{"two"}, []string{"one", "two", "three", "four"}, true),
				Entry("is multiple datum array with nil search array", []string{"two", "four"}, nil, false),
				Entry("is multiple datum array with single search array", []string{"two", "four"}, []string{}, false),
				Entry("is multiple datum array with single invalid search array", []string{"two", "four"}, []string{"one"}, false),
				Entry("is multiple datum array with single valid search array", []string{"two", "four"}, []string{"two"}, true),
				Entry("is multiple datum array with multiple invalid search array", []string{"two", "four"}, []string{"one", "three"}, false),
				Entry("is multiple datum array with multiple valid search array", []string{"two", "four"}, []string{"two", "four"}, true),
				Entry("is multiple datum array with multiple invalid and valid search array", []string{"two", "four"}, []string{"one", "two", "three", "four"}, true),
			)
		})
	})

	Context("SortAndDeduplicateStringArray", func() {
		It("does nothing if string array is nil", func() {
			upload.SortAndDeduplicateStringArray(nil)
		})

		DescribeTable("returns expected result",
			func(stringArray []string, expectedStringArray []string) {
				upload.SortAndDeduplicateStringArray(&stringArray)
				Expect(stringArray).To(Equal(expectedStringArray))
			},
			Entry("empty", []string{}, []string{}),
			Entry("single element", []string{"alpha"}, []string{"alpha"}),
			Entry("multiple elements; in order", []string{"alpha", "beta", "charlie"}, []string{"alpha", "beta", "charlie"}),
			Entry("multiple elements; out of order", []string{"charlie", "alpha", "beta"}, []string{"alpha", "beta", "charlie"}),
			Entry("multiple elements; duplicates; in order", []string{"alpha", "alpha", "beta", "beta", "charlie", "charlie"}, []string{"alpha", "beta", "charlie"}),
			Entry("multiple elements; duplicates; out of order", []string{"charlie", "charlie", "beta", "beta", "alpha", "charlie", "beta", "alpha", "beta", "alpha", "beta", "beta"}, []string{"alpha", "beta", "charlie"}),
		)
	})
})
