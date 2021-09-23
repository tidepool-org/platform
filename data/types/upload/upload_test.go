package upload_test

import (
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTypes "github.com/tidepool-org/platform/data/types"
	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	dataTypesUploadTest "github.com/tidepool-org/platform/data/types/upload/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

func NewMeta() interface{} {
	return &dataTypes.Meta{
		Type: "upload",
	}
}

var _ = Describe("Upload", func() {
	It("Type is expected", func() {
		Expect(dataTypesUpload.Type).To(Equal("upload"))
	})

	It("ComputerTimeFormat is expected", func() {
		Expect(dataTypesUpload.ComputerTimeFormat).To(Equal("2006-01-02T15:04:05"))
	})

	It("DataSetTypeContinuous is expected", func() {
		Expect(dataTypesUpload.DataSetTypeContinuous).To(Equal("continuous"))
	})

	It("DataSetTypeNormal is expected", func() {
		Expect(dataTypesUpload.DataSetTypeNormal).To(Equal("normal"))
	})

	It("DeviceTagBGM is expected", func() {
		Expect(dataTypesUpload.DeviceTagBGM).To(Equal("bgm"))
	})

	It("DeviceTagCGM is expected", func() {
		Expect(dataTypesUpload.DeviceTagCGM).To(Equal("cgm"))
	})

	It("DeviceTagInsulinPump is expected", func() {
		Expect(dataTypesUpload.DeviceTagInsulinPump).To(Equal("insulin-pump"))
	})

	It("StateClosed is expected", func() {
		Expect(dataTypesUpload.StateClosed).To(Equal("closed"))
	})

	It("StateOpen is expected", func() {
		Expect(dataTypesUpload.StateOpen).To(Equal("open"))
	})

	It("TimeProcessingAcrossTheBoardTimeZone is expected", func() {
		Expect(dataTypesUpload.TimeProcessingAcrossTheBoardTimeZone).To(Equal("across-the-board-timezone"))
	})

	It("TimeProcessingNone is expected", func() {
		Expect(dataTypesUpload.TimeProcessingNone).To(Equal("none"))
	})

	It("TimeProcessingUTCBootstrapping is expected", func() {
		Expect(dataTypesUpload.TimeProcessingUTCBootstrapping).To(Equal("utc-bootstrapping"))
	})

	It("VersionLengthMinimum is expected", func() {
		Expect(dataTypesUpload.VersionLengthMinimum).To(Equal(5))
	})

	It("DataSetTypes returns expected", func() {
		Expect(dataTypesUpload.DataSetTypes()).To(Equal([]string{"continuous", "normal"}))
	})

	It("DeviceTags returns expected", func() {
		Expect(dataTypesUpload.DeviceTags()).To(Equal([]string{"bgm", "cgm", "insulin-pump"}))
	})

	It("States returns expected", func() {
		Expect(dataTypesUpload.States()).To(Equal([]string{"closed", "open"}))
	})

	It("TimeProcessings returns expected", func() {
		Expect(dataTypesUpload.TimeProcessings()).To(Equal([]string{"across-the-board-timezone", "none", "utc-bootstrapping"}))
	})

	Context("New", func() {
		It("returns the expected datum with all values initialized", func() {
			datum := dataTypesUpload.New()
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
			Expect(datum.Version).To(BeNil())
		})
	})

	Context("Upload", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataTypesUpload.Upload), expectedOrigins []structure.Origin, expectedErrors ...error) {
					datum := dataTypesUploadTest.RandomUpload()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, expectedOrigins, expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataTypesUpload.Upload) {},
					structure.Origins(),
				),
				Entry("type missing",
					func(datum *dataTypesUpload.Upload) { datum.Type = "" },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/type", &dataTypes.Meta{}),
				),
				Entry("type invalid",
					func(datum *dataTypesUpload.Upload) { datum.Type = "invalidType" },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "upload"), "/type", &dataTypes.Meta{Type: "invalidType"}),
				),
				Entry("type upload",
					func(datum *dataTypesUpload.Upload) { datum.Type = "upload" },
					structure.Origins(),
				),
				Entry("by user missing",
					func(datum *dataTypesUpload.Upload) { datum.ByUser = nil },
					structure.Origins(),
				),
				Entry("by user empty",
					func(datum *dataTypesUpload.Upload) { datum.ByUser = pointer.FromString("") },
					structure.Origins(),
				),
				Entry("by user exists",
					func(datum *dataTypesUpload.Upload) { datum.ByUser = pointer.FromString(userTest.RandomID()) },
					structure.Origins(),
				),
				Entry("client missing",
					func(datum *dataTypesUpload.Upload) { datum.Client = nil },
					structure.Origins(),
				),
				Entry("client invalid",
					func(datum *dataTypesUpload.Upload) {
						datum.Client.Name = nil
						datum.Client.Version = nil
					},
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/name", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/version", NewMeta()),
				),
				Entry("client valid",
					func(datum *dataTypesUpload.Upload) { datum.Client = dataTypesUploadTest.NewClient() },
					structure.Origins(),
				),
				Entry("computer time missing",
					func(datum *dataTypesUpload.Upload) { datum.ComputerTime = nil },
					structure.Origins(),
				),
				Entry("computer time valid",
					func(datum *dataTypesUpload.Upload) {
						datum.ComputerTime = pointer.FromTime(test.RandomTime())
					},
					structure.Origins(),
				),
				Entry("data set type missing",
					func(datum *dataTypesUpload.Upload) { datum.DataSetType = nil },
					structure.Origins(),
				),
				Entry("data set type invalid",
					func(datum *dataTypesUpload.Upload) { datum.DataSetType = pointer.FromString("invalid") },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"continuous", "normal"}), "/dataSetType", NewMeta()),
				),
				Entry("data set type normal",
					func(datum *dataTypesUpload.Upload) { datum.DataSetType = pointer.FromString("normal") },
					structure.Origins(),
				),
				Entry("data set type continuous",
					func(datum *dataTypesUpload.Upload) { datum.DataSetType = pointer.FromString("continuous") },
					structure.Origins(),
				),
				Entry("data state missing",
					func(datum *dataTypesUpload.Upload) { datum.DataState = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
				),
				Entry("data state invalid",
					func(datum *dataTypesUpload.Upload) { datum.DataState = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/dataState", NewMeta()),
				),
				Entry("data state open",
					func(datum *dataTypesUpload.Upload) { datum.DataState = pointer.FromString("open") },
					structure.Origins(),
				),
				Entry("data state closed",
					func(datum *dataTypesUpload.Upload) { datum.DataState = pointer.FromString("closed") },
					structure.Origins(),
				),
				Entry("device manufacturers missing",
					func(datum *dataTypesUpload.Upload) { datum.DeviceManufacturers = nil },
					structure.Origins(),
				),
				Entry("device manufacturers empty",
					func(datum *dataTypesUpload.Upload) { datum.DeviceManufacturers = pointer.FromStringArray([]string{}) },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceManufacturers", NewMeta()),
				),
				Entry("device manufacturers element empty",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), ""})
					},
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceManufacturers/1", NewMeta()),
				),
				Entry("device manufacturers single",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16)})
					},
					structure.Origins(),
				),
				Entry("device manufacturers multiple",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), test.RandomStringFromRange(1, 16)})
					},
					structure.Origins(),
				),
				Entry("device manufacturers multiple duplicates",
					func(datum *dataTypesUpload.Upload) {
						duplicate := test.RandomStringFromRange(1, 16)
						datum.DeviceManufacturers = pointer.FromStringArray([]string{test.RandomStringFromRange(1, 16), duplicate, duplicate, test.RandomStringFromRange(1, 16)})
					},
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/deviceManufacturers/2", NewMeta()),
				),
				Entry("device model missing",
					func(datum *dataTypesUpload.Upload) { datum.DeviceModel = nil },
					structure.Origins(),
				),
				Entry("device model empty",
					func(datum *dataTypesUpload.Upload) { datum.DeviceModel = pointer.FromString("") },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceModel", NewMeta()),
				),
				Entry("device model exists",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceModel = pointer.FromString(test.RandomStringFromRange(1, 32))
					},
					structure.Origins(),
				),
				Entry("device serial number missing",
					func(datum *dataTypesUpload.Upload) { datum.DeviceSerialNumber = nil },
					structure.Origins(),
				),
				Entry("device serial number empty",
					func(datum *dataTypesUpload.Upload) { datum.DeviceSerialNumber = pointer.FromString("") },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceSerialNumber", NewMeta()),
				),
				Entry("device serial number exists",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceSerialNumber = pointer.FromString(test.RandomStringFromRange(1, 16))
					},
					structure.Origins(),
				),
				Entry("device tags missing",
					func(datum *dataTypesUpload.Upload) { datum.DeviceTags = nil },
					structure.Origins(),
				),
				Entry("device tags empty",
					func(datum *dataTypesUpload.Upload) { datum.DeviceTags = pointer.FromStringArray([]string{}) },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceTags", NewMeta()),
				),
				Entry("device tags elements single invalid",
					func(datum *dataTypesUpload.Upload) { datum.DeviceTags = pointer.FromStringArray([]string{"invalid"}) },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"bgm", "cgm", "insulin-pump"}), "/deviceTags/0", NewMeta()),
				),
				Entry("device tags elements single bgm",
					func(datum *dataTypesUpload.Upload) { datum.DeviceTags = pointer.FromStringArray([]string{"bgm"}) },
					structure.Origins(),
				),
				Entry("device tags elements single cgm",
					func(datum *dataTypesUpload.Upload) { datum.DeviceTags = pointer.FromStringArray([]string{"cgm"}) },
					structure.Origins(),
				),
				Entry("device tags elements single insulin-pump",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceTags = pointer.FromStringArray([]string{"insulin-pump"})
					},
					structure.Origins(),
				),
				Entry("device tags elements multiple invalid",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceTags = pointer.FromStringArray([]string{"bgm", "invalid", "insulin-pump"})
					},
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"bgm", "cgm", "insulin-pump"}), "/deviceTags/1", NewMeta()),
				),
				Entry("device tags elements multiple valid",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceTags = pointer.FromStringArray([]string{"cgm", "insulin-pump"})
					},
					structure.Origins(),
				),
				Entry("device tags elements multiple valid duplicates",
					func(datum *dataTypesUpload.Upload) {
						datum.DeviceTags = pointer.FromStringArray([]string{"cgm", "insulin-pump", "cgm", "insulin-pump"})
					},
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/deviceTags/2", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueDuplicate(), "/deviceTags/3", NewMeta()),
				),
				Entry("state missing",
					func(datum *dataTypesUpload.Upload) { datum.State = nil },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
				),
				Entry("state invalid",
					func(datum *dataTypesUpload.Upload) { datum.State = pointer.FromString("invalid") },
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/state", NewMeta()),
				),
				Entry("state open",
					func(datum *dataTypesUpload.Upload) { datum.State = pointer.FromString("open") },
					structure.Origins(),
				),
				Entry("state closed",
					func(datum *dataTypesUpload.Upload) { datum.State = pointer.FromString("closed") },
					structure.Origins(),
				),
				Entry("time processing missing",
					func(datum *dataTypesUpload.Upload) { datum.TimeProcessing = nil },
					structure.Origins(),
				),
				Entry("time processing invalid",
					func(datum *dataTypesUpload.Upload) { datum.TimeProcessing = pointer.FromString("invalid") },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"across-the-board-timezone", "none", "utc-bootstrapping"}), "/timeProcessing", NewMeta()),
				),
				Entry("time processing across-the-board-timezone",
					func(datum *dataTypesUpload.Upload) {
						datum.TimeProcessing = pointer.FromString("across-the-board-timezone")
					},
					structure.Origins(),
				),
				Entry("time processing none",
					func(datum *dataTypesUpload.Upload) { datum.TimeProcessing = pointer.FromString("none") },
					structure.Origins(),
				),
				Entry("time processing utc-bootstrapping",
					func(datum *dataTypesUpload.Upload) { datum.TimeProcessing = pointer.FromString("utc-bootstrapping") },
					structure.Origins(),
				),
				Entry("version missing",
					func(datum *dataTypesUpload.Upload) { datum.Version = nil },
					structure.Origins(),
				),
				Entry("version out of range (lower)",
					func(datum *dataTypesUpload.Upload) { datum.Version = pointer.FromString("1.23") },
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(4, 5), "/version", NewMeta()),
				),
				Entry("version in range (lower)",
					func(datum *dataTypesUpload.Upload) {
						datum.Version = pointer.FromString(netTest.RandomSemanticVersion())
					},
					structure.Origins(),
				),
				Entry("multiple errors with store origin",
					func(datum *dataTypesUpload.Upload) {},
					[]structure.Origin{structure.OriginStore},
				),
				Entry("multiple errors with internal origin",
					func(datum *dataTypesUpload.Upload) {
						datum.DataState = pointer.FromString("invalid")
						datum.State = pointer.FromString("invalid")
					},
					[]structure.Origin{structure.OriginInternal, structure.OriginStore},
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/dataState", NewMeta()),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"closed", "open"}), "/state", NewMeta()),
				),
				Entry("multiple errors with external origin",
					func(datum *dataTypesUpload.Upload) {
						datum.Type = "invalidType"
						datum.Client.Name = nil
						datum.Client.Version = nil
						datum.ComputerTime = nil
						datum.DataSetType = pointer.FromString("invalid")
						datum.DeviceManufacturers = pointer.FromStringArray([]string{})
						datum.DeviceModel = pointer.FromString("")
						datum.DeviceSerialNumber = pointer.FromString("")
						datum.DeviceTags = pointer.FromStringArray([]string{})
						datum.TimeProcessing = pointer.FromString("invalid")
						datum.Version = pointer.FromString("1.23")
					},
					structure.Origins(),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotEqualTo("invalidType", "upload"), "/type", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/name", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueNotExists(), "/client/version", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"continuous", "normal"}), "/dataSetType", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceManufacturers", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceModel", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceSerialNumber", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueEmpty(), "/deviceTags", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorValueStringNotOneOf("invalid", []string{"across-the-board-timezone", "none", "utc-bootstrapping"}), "/timeProcessing", &dataTypes.Meta{Type: "invalidType"}),
					errorsTest.WithPointerSourceAndMeta(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(4, 5), "/version", &dataTypes.Meta{Type: "invalidType"}),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum with origin external",
				func(mutator func(datum *dataTypesUpload.Upload), expectator func(datum *dataTypesUpload.Upload, expectedDatum *dataTypesUpload.Upload)) {
					datum := dataTypesUploadTest.RandomUpload()
					mutator(datum)
					expectedDatum := dataTypesUploadTest.CloneUpload(datum)
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
					func(datum *dataTypesUpload.Upload) {},
					func(datum *dataTypesUpload.Upload, expectedDatum *dataTypesUpload.Upload) {
						Expect(datum.DataSetType).ToNot(BeNil())
						sort.Strings(*expectedDatum.DeviceManufacturers)
						sort.Strings(*expectedDatum.DeviceTags)
					},
				),
				Entry("upload id missing",
					func(datum *dataTypesUpload.Upload) { datum.UploadID = nil },
					func(datum *dataTypesUpload.Upload, expectedDatum *dataTypesUpload.Upload) {
						Expect(datum.UploadID).ToNot(BeNil())
						Expect(*datum.UploadID).To(Equal(*datum.ID))
						expectedDatum.UploadID = datum.UploadID
						sort.Strings(*expectedDatum.DeviceManufacturers)
						sort.Strings(*expectedDatum.DeviceTags)
					},
				),
				Entry("data set type missing",
					func(datum *dataTypesUpload.Upload) { datum.DataSetType = nil },
					func(datum *dataTypesUpload.Upload, expectedDatum *dataTypesUpload.Upload) {
						Expect(datum.DataSetType).ToNot(BeNil())
						Expect(*datum.DataSetType).To(Equal(dataTypesUpload.DataSetTypeNormal))
						expectedDatum.DataSetType = datum.DataSetType
						sort.Strings(*expectedDatum.DeviceManufacturers)
						sort.Strings(*expectedDatum.DeviceTags)
					},
				),
				Entry("all missing",
					func(datum *dataTypesUpload.Upload) {
						*datum = *dataTypesUpload.New()
						datum.Base = *dataTypesTest.NewBase()
					},
					func(datum *dataTypesUpload.Upload, expectedDatum *dataTypesUpload.Upload) {
						Expect(datum.DataSetType).ToNot(BeNil())
						Expect(*datum.DataSetType).To(Equal(dataTypesUpload.DataSetTypeNormal))
						expectedDatum.DataSetType = datum.DataSetType
					},
				),
			)

			DescribeTable("normalizes the datum with origin internal/store",
				func(mutator func(datum *dataTypesUpload.Upload), expectator func(datum *dataTypesUpload.Upload, expectedDatum *dataTypesUpload.Upload)) {
					for _, origin := range []structure.Origin{structure.OriginInternal, structure.OriginStore} {
						datum := dataTypesUploadTest.RandomUpload()
						mutator(datum)
						expectedDatum := dataTypesUploadTest.CloneUpload(datum)
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
					func(datum *dataTypesUpload.Upload) {},
					nil,
				),
				Entry("data set type missing",
					func(datum *dataTypesUpload.Upload) { datum.DataSetType = nil },
					nil,
				),
				Entry("all missing",
					func(datum *dataTypesUpload.Upload) {
						*datum = *dataTypesUpload.New()
						datum.Base = *dataTypesTest.NewBase()
					},
					nil,
				),
			)
		})

		Context("with new upload", func() {
			var datum *dataTypesUpload.Upload

			BeforeEach(func() {
				datum = dataTypesUploadTest.RandomUpload()
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
