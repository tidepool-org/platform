package source_test

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/data"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	metadataTest "github.com/tidepool-org/platform/metadata/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
	"github.com/tidepool-org/platform/user"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Source", func() {
	It("StateConnected is expected", func() {
		Expect(dataSource.StateConnected).To(Equal("connected"))
	})

	It("StateDisconnected is expected", func() {
		Expect(dataSource.StateDisconnected).To(Equal("disconnected"))
	})

	It("StateError is expected", func() {
		Expect(dataSource.StateError).To(Equal("error"))
	})

	It("States returns expected", func() {
		Expect(dataSource.States()).To(Equal([]string{"connected", "disconnected", "error"}))
	})

	Context("NewFilter", func() {
		It("returns successfully with default values", func() {
			Expect(dataSource.NewFilter()).To(Equal(&dataSource.Filter{}))
		})
	})

	Context("Filter", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataSource.Filter), expectedErrors ...error) {
					expectedDatum := dataSourceTest.RandomFilter()
					object := dataSourceTest.NewObjectFromFilter(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataSource.Filter{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {},
				),
				Entry("provider type invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerType"] = true
						expectedDatum.ProviderType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerType"),
				),
				Entry("provider type valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := authTest.RandomProviderTypes()
						object["providerType"] = valid
						expectedDatum.ProviderType = pointer.FromStringArray(valid)
					},
				),
				Entry("provider name invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerName"] = true
						expectedDatum.ProviderName = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerName"),
				),
				Entry("provider name valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := authTest.RandomProviderNames()
						object["providerName"] = valid
						expectedDatum.ProviderName = pointer.FromStringArray(valid)
					},
				),
				Entry("provider session id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerSessionId"] = true
						expectedDatum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerSessionId"),
				),
				Entry("provider session id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := authTest.RandomProviderSessionIDs()
						object["providerSessionId"] = valid
						expectedDatum.ProviderSessionID = pointer.FromStringArray(valid)
					},
				),
				Entry("provider external id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerExternalId"] = true
						expectedDatum.ProviderExternalID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerExternalId"),
				),
				Entry("provider external id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := authTest.RandomProviderExternalIDs()
						object["providerExternalId"] = valid
						expectedDatum.ProviderExternalID = pointer.FromStringArray(valid)
					},
				),
				Entry("state invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["state"] = true
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/state"),
				),
				Entry("state valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := dataSourceTest.RandomStates()
						object["state"] = valid
						expectedDatum.State = pointer.FromStringArray(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerType"] = true
						object["providerName"] = true
						object["providerSessionId"] = true
						object["providerExternalId"] = true
						object["state"] = true
						expectedDatum.ProviderType = nil
						expectedDatum.ProviderName = nil
						expectedDatum.ProviderSessionID = nil
						expectedDatum.ProviderExternalID = nil
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerName"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerExternalId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/state"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataSource.Filter), expectedErrors ...error) {
					datum := dataSourceTest.RandomFilter()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataSource.Filter) {},
				),
				Entry("provider type missing",
					func(datum *dataSource.Filter) { datum.ProviderType = nil },
				),
				Entry("provider type empty",
					func(datum *dataSource.Filter) {
						datum.ProviderType = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerType"),
				),
				Entry("provider type element empty",
					func(datum *dataSource.Filter) {
						datum.ProviderType = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", auth.ProviderTypes()), "/providerType/0"),
				),
				Entry("provider type element invalid",
					func(datum *dataSource.Filter) {
						datum.ProviderType = pointer.FromStringArray([]string{"invalid"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", auth.ProviderTypes()), "/providerType/0"),
				),
				Entry("provider type element duplicate",
					func(datum *dataSource.Filter) {
						providerType := authTest.RandomProviderType()
						datum.ProviderType = pointer.FromStringArray([]string{providerType, providerType})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/providerType/1"),
				),
				Entry("provider type valid",
					func(datum *dataSource.Filter) {
						datum.ProviderType = pointer.FromStringArray(authTest.RandomProviderTypes())
					},
				),
				Entry("provider name missing",
					func(datum *dataSource.Filter) { datum.ProviderName = nil },
				),
				Entry("provider name empty",
					func(datum *dataSource.Filter) {
						datum.ProviderName = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerName"),
				),
				Entry("provider name element empty",
					func(datum *dataSource.Filter) {
						datum.ProviderName = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerName/0"),
				),
				Entry("provider name element length in range (upper)",
					func(datum *dataSource.Filter) {
						datum.ProviderName = pointer.FromStringArray([]string{test.RandomStringFromRangeAndCharset(100, 100, test.CharsetAlphaNumeric)})
					},
				),
				Entry("provider name element length out of range (upper)",
					func(datum *dataSource.Filter) {
						datum.ProviderName = pointer.FromStringArray([]string{test.RandomStringFromRangeAndCharset(101, 101, test.CharsetAlphaNumeric)})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/providerName/0"),
				),
				Entry("provider name element duplicate",
					func(datum *dataSource.Filter) {
						providerName := authTest.RandomProviderName()
						datum.ProviderName = pointer.FromStringArray([]string{providerName, providerName})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/providerName/1"),
				),
				Entry("provider name valid",
					func(datum *dataSource.Filter) {
						datum.ProviderName = pointer.FromStringArray(authTest.RandomProviderNames())
					},
				),
				Entry("provider session id missing",
					func(datum *dataSource.Filter) { datum.ProviderSessionID = nil },
				),
				Entry("provider session id empty",
					func(datum *dataSource.Filter) {
						datum.ProviderSessionID = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("provider session id element empty",
					func(datum *dataSource.Filter) {
						datum.ProviderSessionID = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId/0"),
				),
				Entry("provider session id element invalid",
					func(datum *dataSource.Filter) {
						datum.ProviderSessionID = pointer.FromStringArray([]string{"invalid"})
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId/0"),
				),
				Entry("provider session id element duplicate",
					func(datum *dataSource.Filter) {
						providerSessionID := authTest.RandomProviderSessionID()
						datum.ProviderSessionID = pointer.FromStringArray([]string{providerSessionID, providerSessionID})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/providerSessionId/1"),
				),
				Entry("provider session id valid",
					func(datum *dataSource.Filter) {
						datum.ProviderSessionID = pointer.FromStringArray([]string{authTest.RandomProviderSessionID()})
					},
				),
				Entry("provider external id missing",
					func(datum *dataSource.Filter) { datum.ProviderExternalID = nil },
				),
				Entry("provider external id empty",
					func(datum *dataSource.Filter) {
						datum.ProviderExternalID = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
				),
				Entry("provider external id element empty",
					func(datum *dataSource.Filter) {
						datum.ProviderExternalID = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId/0"),
				),
				Entry("provider external id element length in range (upper)",
					func(datum *dataSource.Filter) {
						datum.ProviderExternalID = pointer.FromStringArray([]string{test.RandomStringFromRangeAndCharset(100, 100, test.CharsetAlphaNumeric)})
					},
				),
				Entry("provider external id element length out of range (upper)",
					func(datum *dataSource.Filter) {
						datum.ProviderExternalID = pointer.FromStringArray([]string{test.RandomStringFromRangeAndCharset(101, 101, test.CharsetAlphaNumeric)})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/providerExternalId/0"),
				),
				Entry("provider external id element duplicate",
					func(datum *dataSource.Filter) {
						providerExternalID := authTest.RandomProviderExternalID()
						datum.ProviderExternalID = pointer.FromStringArray([]string{providerExternalID, providerExternalID})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/providerExternalId/1"),
				),
				Entry("provider external id valid",
					func(datum *dataSource.Filter) {
						datum.ProviderExternalID = pointer.FromStringArray(authTest.RandomProviderExternalIDs())
					},
				),
				Entry("state missing",
					func(datum *dataSource.Filter) { datum.State = nil },
				),
				Entry("state empty",
					func(datum *dataSource.Filter) {
						datum.State = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/state"),
				),
				Entry("state element empty",
					func(datum *dataSource.Filter) {
						datum.State = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state/0"),
				),
				Entry("state element invalid",
					func(datum *dataSource.Filter) {
						datum.State = pointer.FromStringArray([]string{"invalid"})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataSource.States()), "/state/0"),
				),
				Entry("state element duplicate",
					func(datum *dataSource.Filter) {
						state := dataSourceTest.RandomState()
						datum.State = pointer.FromStringArray([]string{state, state})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/state/1"),
				),
				Entry("state valid",
					func(datum *dataSource.Filter) {
						datum.State = pointer.FromStringArray(dataSourceTest.RandomStates())
					},
				),
				Entry("multiple errors",
					func(datum *dataSource.Filter) {
						datum.ProviderType = pointer.FromStringArray([]string{})
						datum.ProviderName = pointer.FromStringArray([]string{})
						datum.ProviderSessionID = pointer.FromStringArray([]string{})
						datum.ProviderExternalID = pointer.FromStringArray([]string{})
						datum.State = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/state"),
				),
			)
		})

		Context("with new filter", func() {
			var filter *dataSource.Filter

			BeforeEach(func() {
				filter = dataSourceTest.RandomFilter()
			})

			Context("MutateRequest", func() {
				var req *http.Request

				BeforeEach(func() {
					req = testHttp.NewRequest()
				})

				It("returns an error when the request is missing", func() {
					errorsTest.ExpectEqual(filter.MutateRequest(nil), errors.New("request is missing"))
				})

				It("sets request query as expected", func() {
					Expect(filter.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(Equal(url.Values{
						"providerType":       *filter.ProviderType,
						"providerName":       *filter.ProviderName,
						"providerSessionId":  *filter.ProviderSessionID,
						"providerExternalId": *filter.ProviderExternalID,
						"state":              *filter.State,
					}))
				})

				It("does not set request query when the filter is empty", func() {
					filter.ProviderType = nil
					filter.ProviderName = nil
					filter.ProviderSessionID = nil
					filter.ProviderExternalID = nil
					filter.State = nil
					Expect(filter.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(BeEmpty())
				})
			})
		})
	})

	Context("NewCreate", func() {
		It("returns successfully with default values", func() {
			Expect(dataSource.NewCreate()).To(Equal(&dataSource.Create{}))
		})
	})

	Context("Create", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataSource.Create)) {
				datum := dataSourceTest.RandomCreate()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataSourceTest.NewObjectFromCreate(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataSource.Create) {},
			),
			Entry("empty",
				func(datum *dataSource.Create) { *datum = dataSource.Create{} },
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataSource.Create), expectedErrors ...error) {
					expectedDatum := dataSourceTest.RandomCreate()
					object := dataSourceTest.NewObjectFromCreate(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataSource.Create{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {},
				),
				Entry("provider type invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["providerType"] = true
						expectedDatum.ProviderType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerType"),
				),
				Entry("provider type valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						valid := authTest.RandomProviderType()
						object["providerType"] = valid
						expectedDatum.ProviderType = pointer.FromString(valid)
					},
				),
				Entry("provider name invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["providerName"] = true
						expectedDatum.ProviderName = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerName"),
				),
				Entry("provider name valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						valid := authTest.RandomProviderName()
						object["providerName"] = valid
						expectedDatum.ProviderName = pointer.FromString(valid)
					},
				),
				Entry("provider external id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["providerExternalId"] = true
						expectedDatum.ProviderExternalID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerExternalId"),
				),
				Entry("provider external id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						valid := authTest.RandomProviderExternalID()
						object["providerExternalId"] = valid
						expectedDatum.ProviderExternalID = pointer.FromString(valid)
					},
				),
				Entry("metadata invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["metadata"] = true
						expectedDatum.Metadata = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
				),
				Entry("metadata valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						valid := metadataTest.RandomMetadataMap()
						object["metadata"] = valid
						expectedDatum.Metadata = valid
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["providerType"] = true
						object["providerName"] = true
						object["providerSessionId"] = true
						object["providerExternalId"] = true
						object["metadata"] = true
						expectedDatum.ProviderType = nil
						expectedDatum.ProviderName = nil
						expectedDatum.ProviderExternalID = nil
						expectedDatum.Metadata = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerName"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerExternalId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataSource.Create), expectedErrors ...error) {
					datum := dataSourceTest.RandomCreate()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataSource.Create) {},
				),
				Entry("provider type missing",
					func(datum *dataSource.Create) { datum.ProviderType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerType"),
				),
				Entry("provider type empty",
					func(datum *dataSource.Create) {
						datum.ProviderType = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", auth.ProviderTypes()), "/providerType"),
				),
				Entry("provider type invalid",
					func(datum *dataSource.Create) {
						datum.ProviderType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", auth.ProviderTypes()), "/providerType"),
				),
				Entry("provider type valid",
					func(datum *dataSource.Create) {
						datum.ProviderType = pointer.FromString(authTest.RandomProviderType())
					},
				),
				Entry("provider name missing",
					func(datum *dataSource.Create) { datum.ProviderName = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerName"),
				),
				Entry("provider name empty",
					func(datum *dataSource.Create) {
						datum.ProviderName = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerName"),
				),
				Entry("provider name length in range (upper)",
					func(datum *dataSource.Create) {
						datum.ProviderName = pointer.FromString(test.RandomStringFromRangeAndCharset(100, 100, test.CharsetAlphaNumeric))
					},
				),
				Entry("provider name length out of range (upper)",
					func(datum *dataSource.Create) {
						datum.ProviderName = pointer.FromString(test.RandomStringFromRangeAndCharset(101, 101, test.CharsetAlphaNumeric))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/providerName"),
				),
				Entry("provider name valid",
					func(datum *dataSource.Create) {
						datum.ProviderName = pointer.FromString(authTest.RandomProviderName())
					},
				),
				Entry("provider external id missing",
					func(datum *dataSource.Create) { datum.ProviderExternalID = nil },
				),
				Entry("provider external id empty",
					func(datum *dataSource.Create) {
						datum.ProviderExternalID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
				),
				Entry("provider external id length in range (upper)",
					func(datum *dataSource.Create) {
						datum.ProviderExternalID = pointer.FromString(test.RandomStringFromRangeAndCharset(100, 100, test.CharsetAlphaNumeric))
					},
				),
				Entry("provider external id length out of range (upper)",
					func(datum *dataSource.Create) {
						datum.ProviderExternalID = pointer.FromString(test.RandomStringFromRangeAndCharset(101, 101, test.CharsetAlphaNumeric))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/providerExternalId"),
				),
				Entry("provider external id valid",
					func(datum *dataSource.Create) {
						datum.ProviderExternalID = pointer.FromString(authTest.RandomProviderExternalID())
					},
				),
				Entry("metadata invalid",
					func(datum *dataSource.Create) {
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataSource.MetadataLengthMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataSource.MetadataLengthMaximum), "/metadata"),
				),
				Entry("metadata valid",
					func(datum *dataSource.Create) {},
				),
				Entry("multiple errors",
					func(datum *dataSource.Create) {
						datum.ProviderType = nil
						datum.ProviderName = nil
						datum.ProviderExternalID = pointer.FromString("")
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataSource.MetadataLengthMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataSource.MetadataLengthMaximum), "/metadata"),
				),
			)
		})
	})

	Context("NewUpdate", func() {
		It("returns successfully with default values", func() {
			Expect(dataSource.NewUpdate()).To(Equal(&dataSource.Update{}))
		})
	})

	Context("Update", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataSource.Update)) {
				datum := dataSourceTest.RandomUpdate()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataSourceTest.NewObjectFromUpdate(datum, test.ObjectFormatJSON))
			},
			Entry("succeeds",
				func(datum *dataSource.Update) {},
			),
			Entry("empty",
				func(datum *dataSource.Update) { *datum = dataSource.Update{} },
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataSource.Update), expectedErrors ...error) {
					expectedDatum := dataSourceTest.RandomUpdate()
					object := dataSourceTest.NewObjectFromUpdate(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataSource.Update{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(dataSourceTest.MatchUpdate(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {},
				),
				Entry("provider session id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["providerSessionId"] = true
						expectedDatum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
				),
				Entry("provider session id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := authTest.RandomProviderSessionID()
						object["providerSessionId"] = valid
						expectedDatum.ProviderSessionID = pointer.FromString(valid)
					},
				),
				Entry("provider external id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["providerExternalId"] = true
						expectedDatum.ProviderExternalID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerExternalId"),
				),
				Entry("provider external id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := authTest.RandomProviderExternalID()
						object["providerExternalId"] = valid
						expectedDatum.ProviderExternalID = pointer.FromString(valid)
					},
				),
				Entry("state invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["state"] = true
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
				),
				Entry("state valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := dataSourceTest.RandomState()
						object["state"] = valid
						expectedDatum.State = pointer.FromString(valid)
					},
				),
				Entry("metadata invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["metadata"] = true
						expectedDatum.Metadata = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
				),
				Entry("metadata valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := metadataTest.RandomMetadataMap()
						object["metadata"] = metadataTest.NewObjectFromMetadataMap(valid, test.ObjectFormatJSON)
						expectedDatum.Metadata = valid
					},
				),
				Entry("error invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["error"] = true
						expectedDatum.Error = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/error"),
				),
				Entry("error valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := errorsTest.RandomSerializable()
						object["error"] = errorsTest.NewObjectFromSerializable(valid, test.ObjectFormatJSON)
						expectedDatum.Error = valid
					},
				),
				Entry("data set ids invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["dataSetIds"] = true
						expectedDatum.DataSetIDs = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/dataSetIds"),
				),
				Entry("data set ids valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := dataTest.RandomDataSetIDs()
						object["dataSetIds"] = valid
						expectedDatum.DataSetIDs = pointer.FromStringArray(valid)
					},
				),
				Entry("earliest data time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["earliestDataTime"] = true
						expectedDatum.EarliestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/earliestDataTime"),
				),
				Entry("earliest data time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["earliestDataTime"] = "invalid"
						expectedDatum.EarliestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := test.RandomTimeBeforeNow()
						object["earliestDataTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.EarliestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("latest data time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["latestDataTime"] = true
						expectedDatum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/latestDataTime"),
				),
				Entry("latest data time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["latestDataTime"] = "invalid"
						expectedDatum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/latestDataTime"),
				),
				Entry("latest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := test.RandomTimeBeforeNow()
						object["latestDataTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.LatestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("last import time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["lastImportTime"] = true
						expectedDatum.LastImportTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/lastImportTime"),
				),
				Entry("last import time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["lastImportTime"] = "invalid"
						expectedDatum.LastImportTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := test.RandomTimeBeforeNow()
						object["lastImportTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.LastImportTime = pointer.FromTime(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["providerSessionId"] = true
						object["providerExternalId"] = true
						object["state"] = true
						object["metadata"] = true
						object["error"] = true
						object["dataSetIds"] = true
						object["earliestDataTime"] = true
						object["latestDataTime"] = true
						object["lastImportTime"] = true
						expectedDatum.ProviderSessionID = nil
						expectedDatum.ProviderExternalID = nil
						expectedDatum.State = nil
						expectedDatum.Metadata = nil
						expectedDatum.Error = nil
						expectedDatum.DataSetIDs = nil
						expectedDatum.EarliestDataTime = nil
						expectedDatum.LatestDataTime = nil
						expectedDatum.LastImportTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerExternalId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/error"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/dataSetIds"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/earliestDataTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/latestDataTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/lastImportTime"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataSource.Update), expectedErrors ...error) {
					datum := dataSourceTest.RandomUpdate()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataSource.Update) {},
				),
				Entry("state missing; provider session id missing",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = nil
					},
				),
				Entry("state missing; provider session id empty",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state missing; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state missing; provider session id valid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state connected; provider session id missing",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString("connected")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
				),
				Entry("state connected; provider session id empty",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = pointer.FromString("connected")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("state connected; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = pointer.FromString("connected")
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("state connected; provider session id valid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString("connected")
					},
				),
				Entry("state disconnected; provider session id missing",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString("disconnected")
					},
				),
				Entry("state disconnected; provider session id empty",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = pointer.FromString("disconnected")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state disconnected; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = pointer.FromString("disconnected")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state disconnected; provider session id valid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString("disconnected")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state error; provider session id missing",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString("error")
					},
				),
				Entry("state error; provider session id empty",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = pointer.FromString("error")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state error; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = pointer.FromString("error")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state error; provider session id valid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString("error")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("provider external id missing",
					func(datum *dataSource.Update) {
						datum.ProviderExternalID = nil
					},
				),
				Entry("provider external id empty",
					func(datum *dataSource.Update) {
						datum.ProviderExternalID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
				),
				Entry("provider external id length in range (upper)",
					func(datum *dataSource.Update) {
						datum.ProviderExternalID = pointer.FromString(test.RandomStringFromRangeAndCharset(100, 100, test.CharsetAlphaNumeric))
					},
				),
				Entry("provider external id length out of range (upper)",
					func(datum *dataSource.Update) {
						datum.ProviderExternalID = pointer.FromString(test.RandomStringFromRangeAndCharset(101, 101, test.CharsetAlphaNumeric))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/providerExternalId"),
				),
				Entry("provider external id valid",
					func(datum *dataSource.Update) {
						datum.ProviderExternalID = pointer.FromString(authTest.RandomProviderExternalID())
					},
				),
				Entry("state missing",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = nil
					},
				),
				Entry("state empty",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state"),
				),
				Entry("state invalid",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataSource.States()), "/state"),
				),
				Entry("state connected",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString("connected")
					},
				),
				Entry("state disconnected",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString("disconnected")
					},
				),
				Entry("state error",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString("error")
					},
				),
				Entry("metadata invalid",
					func(datum *dataSource.Update) {
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataSource.MetadataLengthMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataSource.MetadataLengthMaximum), "/metadata"),
				),
				Entry("metadata valid",
					func(datum *dataSource.Update) {},
				),
				Entry("error missing",
					func(datum *dataSource.Update) {
						datum.Error = nil
					},
				),
				Entry("error valid",
					func(datum *dataSource.Update) {
						datum.Error = errorsTest.RandomSerializable()
					},
				),
				Entry("data set ids missing",
					func(datum *dataSource.Update) { datum.DataSetIDs = nil },
				),
				Entry("data set ids empty",
					func(datum *dataSource.Update) {
						datum.DataSetIDs = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds"),
				),
				Entry("data set ids element empty",
					func(datum *dataSource.Update) {
						datum.DataSetIDs = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds/0"),
				),
				Entry("data set ids element invalid",
					func(datum *dataSource.Update) {
						datum.DataSetIDs = pointer.FromStringArray([]string{"invalid"})
					},
					errorsTest.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/dataSetIds/0"),
				),
				Entry("data set ids element duplicate",
					func(datum *dataSource.Update) {
						dataSetID := dataTest.RandomDataSetID()
						datum.DataSetIDs = pointer.FromStringArray([]string{dataSetID, dataSetID})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/dataSetIds/1"),
				),
				Entry("data set ids valid",
					func(datum *dataSource.Update) {
						datum.DataSetIDs = pointer.FromStringArray([]string{dataTest.RandomDataSetID()})
					},
				),
				Entry("earliest data time missing",
					func(datum *dataSource.Update) { datum.EarliestDataTime = nil },
				),
				Entry("earliest data time zero",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/earliestDataTime"),
				),
				Entry("earliest data time after now",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(test.FutureFarTime())
						datum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time missing; latest data time missing",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time missing; latest data time zero",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/latestDataTime"),
				),
				Entry("earliest data time missing; latest data time after now",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/latestDataTime"),
				),
				Entry("earliest data time missing; latest data time valid",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
					},
				),
				Entry("earliest data time valid; latest data time missing",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time valid; latest data time before earliest data time",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time after now",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time valid",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.PastNearTime(), time.Now()))
					},
				),
				Entry("last import time missing",
					func(datum *dataSource.Update) { datum.LastImportTime = nil },
				),
				Entry("last import time zero",
					func(datum *dataSource.Update) {
						datum.LastImportTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/lastImportTime"),
				),
				Entry("last import time after now",
					func(datum *dataSource.Update) {
						datum.LastImportTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(datum *dataSource.Update) {
						datum.LastImportTime = pointer.FromTime(test.RandomTimeBeforeNow())
					},
				),
				Entry("multiple errors",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.ProviderExternalID = pointer.FromString("")
						datum.State = pointer.FromString("")
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataSource.MetadataLengthMaximum)}
						datum.DataSetIDs = pointer.FromStringArray([]string{})
						datum.EarliestDataTime = pointer.FromTime(time.Time{})
						datum.LatestDataTime = pointer.FromTime(time.Time{})
						datum.LastImportTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataSource.MetadataLengthMaximum), "/metadata"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/earliestDataTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/latestDataTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/lastImportTime"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataSource.Update), expectator func(datum *dataSource.Update, expectedDatum *dataSource.Update)) {
					datum := dataSourceTest.RandomUpdate()
					mutator(datum)
					expectedDatum := dataSourceTest.CloneUpdate(datum)
					normalizer := structureNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					Expect(normalizer.Normalize(datum)).ToNot(HaveOccurred())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum",
					func(datum *dataSource.Update) {},
					func(datum *dataSource.Update, expectedDatum *dataSource.Update) {},
				),
			)
		})

		Context("IsEmpty", func() {
			var datum *dataSource.Update

			BeforeEach(func() {
				datum = dataSource.NewUpdate()
			})

			It("returns true when no fields are specified", func() {
				Expect(datum.IsEmpty()).To(BeTrue())
			})

			It("returns false when provider session id is not nil", func() {
				datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when provider external id is not nil", func() {
				datum.ProviderExternalID = pointer.FromString(authTest.RandomProviderExternalID())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when state is not nil", func() {
				datum.State = pointer.FromString(dataSourceTest.RandomState())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when metadata is not nil", func() {
				datum.Metadata = metadataTest.RandomMetadataMap()
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when error is not nil", func() {
				datum.Error = errorsTest.RandomSerializable()
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when data set ids is not nil", func() {
				datum.DataSetIDs = pointer.FromStringArray(dataTest.RandomDataSetIDs())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when earliest data time is not nil", func() {
				datum.EarliestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when latest data time is not nil", func() {
				datum.LatestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when last import time is not nil", func() {
				datum.LastImportTime = pointer.FromTime(test.RandomTimeBeforeNow())
				Expect(datum.IsEmpty()).To(BeFalse())
			})

			It("returns false when all fields are not nil", func() {
				datum = dataSourceTest.RandomUpdate()
				Expect(datum.IsEmpty()).To(BeFalse())
			})
		})
	})

	Context("Source", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataSource.Source)) {
				datum := dataSourceTest.RandomSource()
				mutator(datum)
				test.ExpectSerializedObjectJSON(datum, dataSourceTest.NewObjectFromSource(datum, test.ObjectFormatJSON))
				test.ExpectSerializedObjectBSON(datum, dataSourceTest.NewObjectFromSource(datum, test.ObjectFormatBSON))
			},
			Entry("succeeds",
				func(datum *dataSource.Source) {},
			),
			Entry("empty",
				func(datum *dataSource.Source) { *datum = dataSource.Source{} },
			),
		)

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *dataSource.Source), expectedErrors ...error) {
					expectedDatum := dataSourceTest.RandomSource()
					object := dataSourceTest.NewObjectFromSource(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &dataSource.Source{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(dataSourceTest.MatchSource(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {},
				),
				Entry("id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["id"] = true
						expectedDatum.ID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
				),
				Entry("id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := dataSourceTest.RandomDataSourceID()
						object["id"] = valid
						expectedDatum.ID = pointer.FromString(valid)
					},
				),
				Entry("user id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["userId"] = true
						expectedDatum.UserID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
				),
				Entry("user id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := userTest.RandomUserID()
						object["userId"] = valid
						expectedDatum.UserID = pointer.FromString(valid)
					},
				),
				Entry("provider type invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["providerType"] = true
						expectedDatum.ProviderType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerType"),
				),
				Entry("provider type valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := authTest.RandomProviderType()
						object["providerType"] = valid
						expectedDatum.ProviderType = pointer.FromString(valid)
					},
				),
				Entry("provider name invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["providerName"] = true
						expectedDatum.ProviderName = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerName"),
				),
				Entry("provider name valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := authTest.RandomProviderName()
						object["providerName"] = valid
						expectedDatum.ProviderName = pointer.FromString(valid)
					},
				),
				Entry("provider session id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["providerSessionId"] = true
						expectedDatum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
				),
				Entry("provider session id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := authTest.RandomProviderSessionID()
						object["providerSessionId"] = valid
						expectedDatum.ProviderSessionID = pointer.FromString(valid)
					},
				),
				Entry("provider external id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["providerExternalId"] = true
						expectedDatum.ProviderExternalID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerExternalId"),
				),
				Entry("provider external id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := authTest.RandomProviderExternalID()
						object["providerExternalId"] = valid
						expectedDatum.ProviderExternalID = pointer.FromString(valid)
					},
				),
				Entry("state invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["state"] = true
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
				),
				Entry("state valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := dataSourceTest.RandomState()
						object["state"] = valid
						expectedDatum.State = pointer.FromString(valid)
					},
				),
				Entry("metadata invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["metadata"] = true
						expectedDatum.Metadata = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
				),
				Entry("metadata valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := metadataTest.RandomMetadataMap()
						object["metadata"] = metadataTest.NewObjectFromMetadataMap(valid, test.ObjectFormatJSON)
						expectedDatum.Metadata = valid
					},
				),
				Entry("error invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["error"] = true
						expectedDatum.Error = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/error"),
				),
				Entry("error valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := errorsTest.RandomSerializable()
						object["error"] = errorsTest.NewObjectFromSerializable(valid, test.ObjectFormatJSON)
						expectedDatum.Error = valid
					},
				),
				Entry("data set ids invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["dataSetIds"] = true
						expectedDatum.DataSetIDs = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/dataSetIds"),
				),
				Entry("data set ids valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := dataTest.RandomDataSetIDs()
						object["dataSetIds"] = valid
						expectedDatum.DataSetIDs = pointer.FromStringArray(valid)
					},
				),
				Entry("earliest data time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["earliestDataTime"] = true
						expectedDatum.EarliestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/earliestDataTime"),
				),
				Entry("earliest data time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["earliestDataTime"] = "invalid"
						expectedDatum.EarliestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeBeforeNow()
						object["earliestDataTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.EarliestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("latest data time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["latestDataTime"] = true
						expectedDatum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/latestDataTime"),
				),
				Entry("latest data time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["latestDataTime"] = "invalid"
						expectedDatum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/latestDataTime"),
				),
				Entry("latest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeBeforeNow()
						object["latestDataTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.LatestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("last import time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["lastImportTime"] = true
						expectedDatum.LastImportTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/lastImportTime"),
				),
				Entry("last import time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["lastImportTime"] = "invalid"
						expectedDatum.LastImportTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeBeforeNow()
						object["lastImportTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.LastImportTime = pointer.FromTime(valid)
					},
				),
				Entry("created time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["createdTime"] = true
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
				),
				Entry("created time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["createdTime"] = "invalid"
						expectedDatum.CreatedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/createdTime"),
				),
				Entry("created time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeBeforeNow()
						object["createdTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.CreatedTime = pointer.FromTime(valid)
					},
				),
				Entry("modified time invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["modifiedTime"] = true
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
				),
				Entry("modified time invalid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["modifiedTime"] = "invalid"
						expectedDatum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339Nano), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeBeforeNow()
						object["modifiedTime"] = valid.Format(time.RFC3339Nano)
						expectedDatum.ModifiedTime = pointer.FromTime(valid)
					},
				),
				Entry("revision invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
				Entry("revision valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := requestTest.RandomRevision()
						object["revision"] = valid
						expectedDatum.Revision = pointer.FromInt(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["id"] = true
						object["userId"] = true
						object["providerType"] = true
						object["providerName"] = true
						object["providerSessionId"] = true
						object["providerExternalId"] = true
						object["state"] = true
						object["metadata"] = true
						object["error"] = true
						object["dataSetIds"] = true
						object["earliestDataTime"] = true
						object["latestDataTime"] = true
						object["lastImportTime"] = true
						object["createdTime"] = true
						object["modifiedTime"] = true
						object["revision"] = true
						expectedDatum.ID = nil
						expectedDatum.UserID = nil
						expectedDatum.ProviderType = nil
						expectedDatum.ProviderName = nil
						expectedDatum.ProviderSessionID = nil
						expectedDatum.ProviderExternalID = nil
						expectedDatum.State = nil
						expectedDatum.Metadata = nil
						expectedDatum.Error = nil
						expectedDatum.DataSetIDs = nil
						expectedDatum.EarliestDataTime = nil
						expectedDatum.LatestDataTime = nil
						expectedDatum.LastImportTime = nil
						expectedDatum.CreatedTime = nil
						expectedDatum.ModifiedTime = nil
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/id"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/userId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerName"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerExternalId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotObject(true), "/metadata"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/error"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/dataSetIds"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/earliestDataTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/latestDataTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/lastImportTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/createdTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotTime(true), "/modifiedTime"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataSource.Source), expectedErrors ...error) {
					datum := dataSourceTest.RandomSource()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataSource.Source) {},
				),
				Entry("id missing",
					func(datum *dataSource.Source) {
						datum.ID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
				),
				Entry("id empty",
					func(datum *dataSource.Source) {
						datum.ID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/id"),
				),
				Entry("id invalid",
					func(datum *dataSource.Source) {
						datum.ID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(dataSource.ErrorValueStringAsIDNotValid("invalid"), "/id"),
				),
				Entry("id valid",
					func(datum *dataSource.Source) {
						datum.ID = pointer.FromString(dataSourceTest.RandomDataSourceID())
					},
				),
				Entry("user id missing",
					func(datum *dataSource.Source) {
						datum.UserID = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
				),
				Entry("user id empty",
					func(datum *dataSource.Source) {
						datum.UserID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/userId"),
				),
				Entry("user id invalid",
					func(datum *dataSource.Source) {
						datum.UserID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(user.ErrorValueStringAsIDNotValid("invalid"), "/userId"),
				),
				Entry("user id valid",
					func(datum *dataSource.Source) {
						datum.UserID = pointer.FromString(userTest.RandomUserID())
					},
				),
				Entry("provider type missing",
					func(datum *dataSource.Source) { datum.ProviderType = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerType"),
				),
				Entry("provider type empty",
					func(datum *dataSource.Source) {
						datum.ProviderType = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", auth.ProviderTypes()), "/providerType"),
				),
				Entry("provider type invalid",
					func(datum *dataSource.Source) {
						datum.ProviderType = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", auth.ProviderTypes()), "/providerType"),
				),
				Entry("provider type valid",
					func(datum *dataSource.Source) {
						datum.ProviderType = pointer.FromString(authTest.RandomProviderType())
					},
				),
				Entry("provider name missing",
					func(datum *dataSource.Source) { datum.ProviderName = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerName"),
				),
				Entry("provider name empty",
					func(datum *dataSource.Source) {
						datum.ProviderName = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerName"),
				),
				Entry("provider name length in range (upper)",
					func(datum *dataSource.Source) {
						datum.ProviderName = pointer.FromString(test.RandomStringFromRangeAndCharset(100, 100, test.CharsetAlphaNumeric))
					},
				),
				Entry("provider name length out of range (upper)",
					func(datum *dataSource.Source) {
						datum.ProviderName = pointer.FromString(test.RandomStringFromRangeAndCharset(101, 101, test.CharsetAlphaNumeric))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/providerName"),
				),
				Entry("provider name valid",
					func(datum *dataSource.Source) {
						datum.ProviderName = pointer.FromString(authTest.RandomProviderType())
					},
				),
				Entry("state missing; provider session id missing",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = nil
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state missing; provider session id empty",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state missing; provider session id invalid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = nil
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state missing; provider session id valid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state connected; provider session id missing",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
				),
				Entry("state connected; provider session id empty",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("state connected; provider session id invalid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("state connected; provider session id valid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
				),
				Entry("state disconnected; provider session id missing",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString(dataSource.StateDisconnected)
					},
				),
				Entry("state disconnected; provider session id empty",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = pointer.FromString(dataSource.StateDisconnected)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state disconnected; provider session id invalid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = pointer.FromString(dataSource.StateDisconnected)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state disconnected; provider session id valid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString(dataSource.StateDisconnected)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state error; provider session id missing",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = nil
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
				),
				Entry("state error; provider session id empty",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("state error; provider session id invalid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("invalid")
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("state error; provider session id valid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString(dataSource.StateConnected)
					},
				),
				Entry("provider external id missing",
					func(datum *dataSource.Source) {
						datum.ProviderExternalID = nil
					},
				),
				Entry("provider external id empty",
					func(datum *dataSource.Source) {
						datum.ProviderExternalID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
				),
				Entry("provider external id length in range (upper)",
					func(datum *dataSource.Source) {
						datum.ProviderExternalID = pointer.FromString(test.RandomStringFromRangeAndCharset(100, 100, test.CharsetAlphaNumeric))
					},
				),
				Entry("provider external id length out of range (upper)",
					func(datum *dataSource.Source) {
						datum.ProviderExternalID = pointer.FromString(test.RandomStringFromRangeAndCharset(101, 101, test.CharsetAlphaNumeric))
					},
					errorsTest.WithPointerSource(structureValidator.ErrorLengthNotLessThanOrEqualTo(101, 100), "/providerExternalId"),
				),
				Entry("provider external id valid",
					func(datum *dataSource.Source) {
						datum.ProviderExternalID = pointer.FromString(authTest.RandomProviderExternalID())
					},
				),
				Entry("state missing",
					func(datum *dataSource.Source) {
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state empty",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state"),
				),
				Entry("state invalid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataSource.States()), "/state"),
				),
				Entry("state valid",
					func(datum *dataSource.Source) {},
				),
				Entry("metadata invalid",
					func(datum *dataSource.Source) {
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataSource.MetadataLengthMaximum)}
					},
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataSource.MetadataLengthMaximum), "/metadata"),
				),
				Entry("metadata valid",
					func(datum *dataSource.Source) {},
				),
				Entry("error missing",
					func(datum *dataSource.Source) {
						datum.Error = nil
					},
				),
				Entry("error valid",
					func(datum *dataSource.Source) {
						datum.Error = errorsTest.RandomSerializable()
					},
				),
				Entry("data set ids missing",
					func(datum *dataSource.Source) { datum.DataSetIDs = nil },
				),
				Entry("data set ids empty",
					func(datum *dataSource.Source) {
						datum.DataSetIDs = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds"),
				),
				Entry("data set ids element empty",
					func(datum *dataSource.Source) {
						datum.DataSetIDs = pointer.FromStringArray([]string{""})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds/0"),
				),
				Entry("data set ids element invalid",
					func(datum *dataSource.Source) {
						datum.DataSetIDs = pointer.FromStringArray([]string{"invalid"})
					},
					errorsTest.WithPointerSource(data.ErrorValueStringAsSetIDNotValid("invalid"), "/dataSetIds/0"),
				),
				Entry("data set ids element duplicate",
					func(datum *dataSource.Source) {
						dataSetID := dataTest.RandomDataSetID()
						datum.DataSetIDs = pointer.FromStringArray([]string{dataSetID, dataSetID})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/dataSetIds/1"),
				),
				Entry("data set ids valid",
					func(datum *dataSource.Source) {
						datum.DataSetIDs = pointer.FromStringArray([]string{dataTest.RandomDataSetID()})
					},
				),
				Entry("earliest data time missing",
					func(datum *dataSource.Source) { datum.EarliestDataTime = nil },
				),
				Entry("earliest data time zero",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/earliestDataTime"),
				),
				Entry("earliest data time after now",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(test.FutureFarTime())
						datum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time missing; latest data time missing",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time missing; latest data time zero",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/latestDataTime"),
				),
				Entry("earliest data time missing; latest data time after now",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/latestDataTime"),
				),
				Entry("earliest data time missing; latest data time valid",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeBeforeNow())
					},
				),
				Entry("earliest data time valid; latest data time missing",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time valid; latest data time before earliest data time",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time after now",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time valid",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(test.PastNearTime())
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.PastNearTime(), time.Now()))
					},
				),
				Entry("last import time missing",
					func(datum *dataSource.Source) { datum.LastImportTime = nil },
				),
				Entry("last import time zero",
					func(datum *dataSource.Source) {
						datum.LastImportTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/lastImportTime"),
				),
				Entry("last import time after now",
					func(datum *dataSource.Source) {
						datum.LastImportTime = pointer.FromTime(test.FutureFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(datum *dataSource.Source) {
						datum.LastImportTime = pointer.FromTime(test.RandomTimeBeforeNow())
					},
				),
				Entry("created time missing",
					func(datum *dataSource.Source) { datum.CreatedTime = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
				),
				Entry("created time zero",
					func(datum *dataSource.Source) { datum.CreatedTime = pointer.FromTime(time.Time{}) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/createdTime"),
				),
				Entry("created time after now",
					func(datum *dataSource.Source) {
						datum.CreatedTime = pointer.FromTime(test.FutureFarTime())
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/createdTime"),
				),
				Entry("created time valid",
					func(datum *dataSource.Source) {
						datum.CreatedTime = pointer.FromTime(test.RandomTimeBeforeNow())
						datum.ModifiedTime = nil
					},
				),
				Entry("modified time missing",
					func(datum *dataSource.Source) { datum.ModifiedTime = nil },
				),
				Entry("modified time before created time",
					func(datum *dataSource.Source) {
						datum.CreatedTime = pointer.FromTime(test.PastNearTime())
						datum.ModifiedTime = pointer.FromTime(test.PastFarTime())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(test.PastFarTime(), test.PastNearTime()), "/modifiedTime"),
				),
				Entry("modified time after now",
					func(datum *dataSource.Source) { datum.ModifiedTime = pointer.FromTime(test.FutureFarTime()) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(test.FutureFarTime()), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(datum *dataSource.Source) {
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()))
					},
				),
				Entry("revision missing",
					func(datum *dataSource.Source) {
						datum.Revision = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/revision"),
				),
				Entry("revision out of range (lower)",
					func(datum *dataSource.Source) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
				Entry("revision in range (lower)",
					func(datum *dataSource.Source) {
						datum.Revision = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					func(datum *dataSource.Source) {
						datum.ID = nil
						datum.UserID = nil
						datum.ProviderType = nil
						datum.ProviderName = nil
						datum.ProviderSessionID = pointer.FromString("")
						datum.ProviderExternalID = pointer.FromString("")
						datum.State = nil
						datum.Metadata = map[string]any{"invalid": strings.Repeat("X", dataSource.MetadataLengthMaximum)}
						datum.DataSetIDs = pointer.FromStringArray([]string{})
						datum.EarliestDataTime = pointer.FromTime(time.Time{})
						datum.LatestDataTime = pointer.FromTime(time.Time{})
						datum.LastImportTime = pointer.FromTime(time.Time{})
						datum.CreatedTime = nil
						datum.ModifiedTime = pointer.FromTime(time.Time{})
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/id"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/userId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerExternalId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
					errorsTest.WithPointerSource(structureValidator.ErrorSizeNotLessThanOrEqualTo(4110, dataSource.MetadataLengthMaximum), "/metadata"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/dataSetIds"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/earliestDataTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/latestDataTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/lastImportTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/createdTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/modifiedTime"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
			)
		})

		Context("Normalize", func() {
			DescribeTable("normalizes the datum",
				func(mutator func(datum *dataSource.Source), expectator func(datum *dataSource.Source, expectedDatum *dataSource.Source)) {
					datum := dataSourceTest.RandomSource()
					mutator(datum)
					expectedDatum := dataSourceTest.CloneSource(datum)
					normalizer := structureNormalizer.New(logTest.NewLogger())
					Expect(normalizer).ToNot(BeNil())
					Expect(normalizer.Normalize(datum)).ToNot(HaveOccurred())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("does not modify the datum",
					func(datum *dataSource.Source) {},
					func(datum *dataSource.Source, expectedDatum *dataSource.Source) {},
				),
			)
		})

		Context("Sanitize", func() {
			var original *dataSource.Source
			var sanitized *dataSource.Source

			BeforeEach(func() {
				original = dataSourceTest.RandomSource()
				sanitized = dataSourceTest.CloneSource(original)
			})

			It("returns an error when the details are missing", func() {
				errorsTest.ExpectEqual(sanitized.Sanitize(nil), errors.New("unable to sanitize"))
			})

			It("does not modify the original when the details are server", func() {
				details := request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
				Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
				Expect(sanitized).To(Equal(original))
			})

			It("removes the provider session id and sanitizes the error when the details are user", func() {
				details := request.NewAuthDetails(request.MethodSessionToken, userTest.RandomUserID(), authTest.NewSessionToken())
				Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
				original.ProviderSessionID = nil
				original.Error.Error = errors.Sanitize(original.Error.Error)
				Expect(sanitized).To(Equal(original))
			})

			Context("with a wrapped unauthenticated error", func() {
				var unauthenticatedError error

				BeforeEach(func() {
					unauthenticatedError = request.ErrorUnauthenticated()
					original.Error.Error = errors.Wrap(unauthenticatedError, "wrapped unauthenticated error")
					sanitized = dataSourceTest.CloneSource(original)
				})

				It("promotes the unauthenticated error cause", func() {
					details := request.NewAuthDetails(request.MethodSessionToken, userTest.RandomUserID(), authTest.NewSessionToken())
					Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
					original.ProviderSessionID = nil
					original.Error.Error = errors.Sanitize(unauthenticatedError)
					Expect(sanitized).To(Equal(original))
				})
			})
		})

		Context("HasError", func() {
			It("returns false if the error wrapper is nil", func() {
				source := dataSourceTest.RandomSource()
				source.Error = nil
				Expect(source.HasError()).To(BeFalse())
			})

			It("returns false if the error is nil", func() {
				source := dataSourceTest.RandomSource()
				source.Error = &errors.Serializable{}
				Expect(source.HasError()).To(BeFalse())
			})

			It("returns true if the error is not nil", func() {
				testErr := errorsTest.RandomError()
				source := dataSourceTest.RandomSource()
				source.Error = &errors.Serializable{Error: testErr}
				Expect(source.HasError()).To(BeTrue())
			})
		})

		Context("GetError", func() {
			It("returns nil if the error wrapper is nil", func() {
				source := dataSourceTest.RandomSource()
				source.Error = nil
				Expect(source.GetError()).To(BeNil())
			})

			It("returns nil if the error is nil", func() {
				source := dataSourceTest.RandomSource()
				source.Error = &errors.Serializable{}
				Expect(source.GetError()).To(BeNil())
			})

			It("returns the error if the error is not nil", func() {
				testErr := errorsTest.RandomError()
				source := dataSourceTest.RandomSource()
				source.Error = &errors.Serializable{Error: testErr}
				Expect(source.GetError()).To(Equal(testErr))
			})
		})
	})

	Context("SourceArray", func() {
		Context("Sanitize", func() {
			var originals dataSource.SourceArray
			var sanitized dataSource.SourceArray

			BeforeEach(func() {
				originals = dataSourceTest.RandomSourceArray(1, 3)
				sanitized = dataSourceTest.CloneSourceArray(originals)
			})

			It("returns an error when the details are missing", func() {
				errorsTest.ExpectEqual(sanitized.Sanitize(nil), errors.New("unable to sanitize"))
			})

			It("does not modify the originals when the details are server", func() {
				details := request.NewAuthDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
				Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
				Expect(sanitized).To(Equal(originals))
			})

			It("removes the provider session id and sanitizes the error when the details are user", func() {
				details := request.NewAuthDetails(request.MethodSessionToken, userTest.RandomUserID(), authTest.NewSessionToken())
				Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
				for _, original := range originals {
					original.ProviderSessionID = nil
					original.Error.Error = errors.Sanitize(original.Error.Error)
				}
				Expect(sanitized).To(Equal(originals))
			})
		})
	})

	Context("NewID", func() {
		It("returns a string of 32 lowercase hexadecimal characters", func() {
			Expect(dataSource.NewID()).To(MatchRegexp("^[0-9a-f]{32}$"))
		})

		It("returns different IDs for each invocation", func() {
			Expect(dataSource.NewID()).ToNot(Equal(dataSource.NewID()))
		})
	})

	Context("IsValidID, IDValidator, and ValidateID", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(dataSource.IsValidID(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				dataSource.IDValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(dataSource.ValidateID(value), expectedErrors...)
			},
			Entry("is an empty string", "", structureValidator.ErrorValueEmpty()),
			Entry("has string length out of range (lower)", "0123456789abcdefghijklmnopqrstu", dataSource.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstu")),
			Entry("has string length in range", test.RandomStringFromRangeAndCharset(32, 32, test.CharsetLowercase+test.CharsetNumeric)),
			Entry("has string length out of range (upper)", "0123456789abcdefghijklmnopqrstuvw", dataSource.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstuvw")),
			Entry("has uppercase characters", "0123456789ABCDEFGHIJKLMNOPQRSTUV", dataSource.ErrorValueStringAsIDNotValid("0123456789ABCDEFGHIJKLMNOPQRSTUV")),
			Entry("has symbols", "0123456789!@#$%^abcdefghijklmnop", dataSource.ErrorValueStringAsIDNotValid("0123456789!@#$%^abcdefghijklmnop")),
			Entry("has whitespace", "0123456789      abcdefghijklmnop", dataSource.ErrorValueStringAsIDNotValid("0123456789      abcdefghijklmnop")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsIDNotValid with empty string", dataSource.ErrorValueStringAsIDNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as data source id`),
			Entry("is ErrorValueStringAsIDNotValid with non-empty string", dataSource.ErrorValueStringAsIDNotValid("0123456789abcdefghijklmnopqrstuv"), "value-not-valid", "value is not valid", `value "0123456789abcdefghijklmnopqrstuv" is not valid as data source id`),
		)
	})
})
