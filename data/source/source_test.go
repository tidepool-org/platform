package source_test

import (
	"net/http"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/data"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataSourceTest "github.com/tidepool-org/platform/data/source/test"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
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

var futureTime = time.Unix(4102444800, 0)
var nearPastTime = time.Unix(1500000000, 0)
var farPastTime = time.Unix(1200000000, 0)

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
			filter := dataSource.NewFilter()
			Expect(filter).ToNot(BeNil())
			Expect(filter.ProviderType).To(BeNil())
			Expect(filter.ProviderName).To(BeNil())
			Expect(filter.ProviderSessionID).To(BeNil())
			Expect(filter.State).To(BeNil())
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
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {},
				),
				Entry("provider type missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						delete(object, "providerType")
						expectedDatum.ProviderType = nil
					},
				),
				Entry("provider type invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerType"] = true
						expectedDatum.ProviderType = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerType"),
				),
				Entry("provider type empty",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerType"] = []string{}
						expectedDatum.ProviderType = pointer.FromStringArray([]string{})
					},
				),
				Entry("provider type valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := authTest.RandomProviderTypes()
						object["providerType"] = valid
						expectedDatum.ProviderType = pointer.FromStringArray(valid)
					},
				),
				Entry("provider name missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						delete(object, "providerName")
						expectedDatum.ProviderName = nil
					},
				),
				Entry("provider name invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerName"] = true
						expectedDatum.ProviderName = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerName"),
				),
				Entry("provider name empty",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerName"] = []string{}
						expectedDatum.ProviderName = pointer.FromStringArray([]string{})
					},
				),
				Entry("provider name valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := authTest.RandomProviderNames()
						object["providerName"] = valid
						expectedDatum.ProviderName = pointer.FromStringArray(valid)
					},
				),
				Entry("provider session id missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						delete(object, "providerSessionId")
						expectedDatum.ProviderSessionID = nil
					},
				),
				Entry("provider session id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerSessionId"] = true
						expectedDatum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerSessionId"),
				),
				Entry("provider session id empty",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["providerSessionId"] = []string{}
						expectedDatum.ProviderSessionID = pointer.FromStringArray([]string{})
					},
				),
				Entry("provider session id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						valid := authTest.RandomProviderSessionIDs()
						object["providerSessionId"] = valid
						expectedDatum.ProviderSessionID = pointer.FromStringArray(valid)
					},
				),
				Entry("state missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						delete(object, "state")
						expectedDatum.State = nil
					},
				),
				Entry("state invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["state"] = true
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/state"),
				),
				Entry("state empty",
					func(object map[string]interface{}, expectedDatum *dataSource.Filter) {
						object["state"] = []string{}
						expectedDatum.State = pointer.FromStringArray([]string{})
					},
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
						object["state"] = true
						expectedDatum.ProviderType = nil
						expectedDatum.ProviderName = nil
						expectedDatum.ProviderSessionID = nil
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerName"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/state"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataSource.Filter), expectedErrors ...error) {
					datum := dataSourceTest.RandomFilter()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
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
						datum.ProviderName = pointer.FromStringArray(authTest.RandomProviderTypes())
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
						datum.State = pointer.FromStringArray([]string{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
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
						"providerType":      *filter.ProviderType,
						"providerName":      *filter.ProviderName,
						"providerSessionId": *filter.ProviderSessionID,
						"state":             *filter.State,
					}))
				})

				It("does not set request query when the filter is empty", func() {
					filter.ProviderType = nil
					filter.ProviderName = nil
					filter.ProviderSessionID = nil
					filter.State = nil
					Expect(filter.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(BeEmpty())
				})
			})
		})
	})

	Context("NewCreate", func() {
		It("returns successfully with default values", func() {
			create := dataSource.NewCreate()
			Expect(create).ToNot(BeNil())
			Expect(create.ProviderType).To(BeNil())
			Expect(create.ProviderName).To(BeNil())
			Expect(create.ProviderSessionID).To(BeNil())
			Expect(create.State).To(BeNil())
		})
	})

	Context("Create", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataSource.Create)) {
				datum := dataSourceTest.RandomCreate()
				mutator(datum)
				test.ExpectSerializedJSON(datum, dataSourceTest.NewObjectFromCreate(datum, test.ObjectFormatJSON))
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
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {},
				),
				Entry("provider type missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						delete(object, "providerType")
						expectedDatum.ProviderType = nil
					},
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
				Entry("provider name missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						delete(object, "providerName")
						expectedDatum.ProviderName = nil
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
				Entry("provider session id missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						delete(object, "providerSessionId")
						expectedDatum.ProviderSessionID = nil
					},
				),
				Entry("provider session id invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["providerSessionId"] = true
						expectedDatum.ProviderSessionID = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
				),
				Entry("provider session id valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						valid := authTest.RandomProviderSessionID()
						object["providerSessionId"] = valid
						expectedDatum.ProviderSessionID = pointer.FromString(valid)
					},
				),
				Entry("state missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						delete(object, "state")
						expectedDatum.State = nil
					},
				),
				Entry("state invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["state"] = true
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
				),
				Entry("state valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						valid := dataSourceTest.RandomState()
						object["state"] = valid
						expectedDatum.State = pointer.FromString(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *dataSource.Create) {
						object["providerType"] = true
						object["providerName"] = true
						object["providerSessionId"] = true
						object["state"] = true
						expectedDatum.ProviderType = nil
						expectedDatum.ProviderName = nil
						expectedDatum.ProviderSessionID = nil
						expectedDatum.State = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerType"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerName"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *dataSource.Create), expectedErrors ...error) {
					datum := dataSourceTest.RandomCreate()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
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
						datum.ProviderName = pointer.FromString(authTest.RandomProviderType())
					},
				),
				Entry("provider session id missing",
					func(datum *dataSource.Create) { datum.ProviderSessionID = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
				),
				Entry("provider session id empty",
					func(datum *dataSource.Create) {
						datum.ProviderSessionID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("provider session id invalid",
					func(datum *dataSource.Create) {
						datum.ProviderSessionID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("provider session id valid",
					func(datum *dataSource.Create) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
					},
				),
				Entry("state missing",
					func(datum *dataSource.Create) { datum.State = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state empty",
					func(datum *dataSource.Create) {
						datum.State = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state"),
				),
				Entry("state invalid",
					func(datum *dataSource.Create) {
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataSource.States()), "/state"),
				),
				Entry("state valid",
					func(datum *dataSource.Create) {
						datum.State = pointer.FromString(dataSourceTest.RandomState())
					},
				),
				Entry("multiple errors",
					func(datum *dataSource.Create) {
						datum.ProviderType = nil
						datum.ProviderName = nil
						datum.ProviderSessionID = nil
						datum.State = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerName"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
			)
		})
	})

	Context("NewUpdate", func() {
		It("returns successfully with default values", func() {
			update := dataSource.NewUpdate()
			Expect(update).ToNot(BeNil())
			Expect(update.ProviderSessionID).To(BeNil())
			Expect(update.State).To(BeNil())
			Expect(update.Error).To(BeNil())
			Expect(update.DataSetIDs).To(BeNil())
			Expect(update.EarliestDataTime).To(BeNil())
			Expect(update.LatestDataTime).To(BeNil())
			Expect(update.LastImportTime).To(BeNil())
		})
	})

	Context("Update", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataSource.Update)) {
				datum := dataSourceTest.RandomUpdate()
				mutator(datum)
				test.ExpectSerializedJSON(datum, dataSourceTest.NewObjectFromUpdate(datum, test.ObjectFormatJSON))
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
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					dataSourceTest.ExpectEqualUpdate(datum, expectedDatum)
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {},
				),
				Entry("provider session id missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						delete(object, "providerSessionId")
						expectedDatum.ProviderSessionID = nil
					},
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
				Entry("state missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						delete(object, "state")
						expectedDatum.State = nil
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
				Entry("error missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						delete(object, "error")
						expectedDatum.Error = nil
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
				Entry("data set ids missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						delete(object, "dataSetIds")
						expectedDatum.DataSetIDs = nil
					},
				),
				Entry("data set ids invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["dataSetIds"] = true
						expectedDatum.DataSetIDs = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/dataSetIds"),
				),
				Entry("data set ids empty",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["dataSetIds"] = []string{}
						expectedDatum.DataSetIDs = pointer.FromStringArray([]string{})
					},
				),
				Entry("data set ids valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := dataTest.RandomSetIDs()
						object["dataSetIds"] = valid
						expectedDatum.DataSetIDs = pointer.FromStringArray(valid)
					},
				),
				Entry("earliest data time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						delete(object, "earliestDataTime")
						expectedDatum.EarliestDataTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["earliestDataTime"] = valid.Format(time.RFC3339)
						expectedDatum.EarliestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("latest data time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						delete(object, "latestDataTime")
						expectedDatum.LatestDataTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/latestDataTime"),
				),
				Entry("latest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["latestDataTime"] = valid.Format(time.RFC3339)
						expectedDatum.LatestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("last import time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						delete(object, "lastImportTime")
						expectedDatum.LastImportTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["lastImportTime"] = valid.Format(time.RFC3339)
						expectedDatum.LastImportTime = pointer.FromTime(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *dataSource.Update) {
						object["providerSessionId"] = true
						object["state"] = true
						object["error"] = true
						object["dataSetIds"] = true
						object["earliestDataTime"] = true
						object["latestDataTime"] = true
						object["lastImportTime"] = true
						expectedDatum.ProviderSessionID = nil
						expectedDatum.State = nil
						expectedDatum.Error = nil
						expectedDatum.DataSetIDs = nil
						expectedDatum.EarliestDataTime = nil
						expectedDatum.LatestDataTime = nil
						expectedDatum.LastImportTime = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/providerSessionId"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
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
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *dataSource.Update) {},
				),
				Entry("state missing; provider session id missing",
					func(datum *dataSource.Update) {
						datum.State = nil
						datum.ProviderSessionID = nil
					},
				),
				Entry("state missing; provider session id empty",
					func(datum *dataSource.Update) {
						datum.State = nil
						datum.ProviderSessionID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("state missing; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.State = nil
						datum.ProviderSessionID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("state missing; provider session id valid",
					func(datum *dataSource.Update) {
						datum.State = nil
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
					},
				),
				Entry("state connected; provider session id missing",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("connected")
						datum.ProviderSessionID = nil
					},
				),
				Entry("state connected; provider session id empty",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("connected")
						datum.ProviderSessionID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("state connected; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("connected")
						datum.ProviderSessionID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("state connected; provider session id valid",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("connected")
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
					},
				),
				Entry("state disconnected; provider session id missing",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("disconnected")
						datum.ProviderSessionID = nil
					},
				),
				Entry("state disconnected; provider session id empty",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("disconnected")
						datum.ProviderSessionID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state disconnected; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("disconnected")
						datum.ProviderSessionID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state disconnected; provider session id valid",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("disconnected")
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueExists(), "/providerSessionId"),
				),
				Entry("state error; provider session id missing",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("error")
						datum.ProviderSessionID = nil
					},
				),
				Entry("state error; provider session id empty",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("error")
						datum.ProviderSessionID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("state error; provider session id invalid",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("error")
						datum.ProviderSessionID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("state error; provider session id valid",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("error")
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
					},
				),
				Entry("state missing",
					func(datum *dataSource.Update) {
						datum.State = nil
					},
				),
				Entry("state empty",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state"),
				),
				Entry("state invalid",
					func(datum *dataSource.Update) {
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataSource.States()), "/state"),
				),
				Entry("state connected",
					func(datum *dataSource.Update) {
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
						datum.State = pointer.FromString("error")
					},
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
						dataSetID := dataTest.RandomSetID()
						datum.DataSetIDs = pointer.FromStringArray([]string{dataSetID, dataSetID})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/dataSetIds/1"),
				),
				Entry("data set ids valid",
					func(datum *dataSource.Update) {
						datum.DataSetIDs = pointer.FromStringArray([]string{dataTest.RandomSetID()})
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
						datum.EarliestDataTime = pointer.FromTime(futureTime)
						datum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
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
						datum.LatestDataTime = pointer.FromTime(futureTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/latestDataTime"),
				),
				Entry("earliest data time missing; latest data time valid",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
					},
				),
				Entry("earliest data time valid; latest data time missing",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time valid; latest data time before earliest data time",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = pointer.FromTime(farPastTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(farPastTime, nearPastTime), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time after now",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = pointer.FromTime(futureTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time valid",
					func(datum *dataSource.Update) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeFromRange(nearPastTime, time.Now()).Truncate(time.Second))
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
						datum.LastImportTime = pointer.FromTime(futureTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(datum *dataSource.Update) {
						datum.LastImportTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
					},
				),
				Entry("multiple errors",
					func(datum *dataSource.Update) {
						datum.ProviderSessionID = pointer.FromString("")
						datum.State = pointer.FromString("")
						datum.DataSetIDs = pointer.FromStringArray([]string{})
						datum.EarliestDataTime = pointer.FromTime(time.Time{})
						datum.LatestDataTime = pointer.FromTime(time.Time{})
						datum.LastImportTime = pointer.FromTime(time.Time{})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state"),
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
					normalizer := structureNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					Expect(normalizer.Normalize(datum)).ToNot(HaveOccurred())
					if expectator != nil {
						expectator(datum, expectedDatum)
					}
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("modifies the datum",
					func(datum *dataSource.Update) {},
					func(datum *dataSource.Update, expectedDatum *dataSource.Update) {
						expectedDatum.EarliestDataTime = pointer.FromTime(expectedDatum.EarliestDataTime.UTC().Truncate(time.Second))
						expectedDatum.LatestDataTime = pointer.FromTime(expectedDatum.LatestDataTime.UTC().Truncate(time.Second))
						expectedDatum.LastImportTime = pointer.FromTime(expectedDatum.LastImportTime.UTC().Truncate(time.Second))
					},
				),
			)
		})

		Context("HasUpdates", func() {
			var update *dataSource.Update

			BeforeEach(func() {
				update = dataSource.NewUpdate()
			})

			It("returns false when no fields are specified", func() {
				Expect(update.HasUpdates()).To(BeFalse())
			})

			It("returns true when provider session id is not nil", func() {
				update.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
				Expect(update.HasUpdates()).To(BeTrue())
			})

			It("returns true when state is not nil", func() {
				update.State = pointer.FromString(dataSourceTest.RandomState())
				Expect(update.HasUpdates()).To(BeTrue())
			})

			It("returns true when error is not nil", func() {
				update.Error = errorsTest.RandomSerializable()
				Expect(update.HasUpdates()).To(BeTrue())
			})

			It("returns true when data set ids is not nil", func() {
				update.DataSetIDs = pointer.FromStringArray(dataTest.RandomSetIDs())
				Expect(update.HasUpdates()).To(BeTrue())
			})

			It("returns true when earliest data time is not nil", func() {
				update.EarliestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Millisecond))
				Expect(update.HasUpdates()).To(BeTrue())
			})

			It("returns true when latest data time is not nil", func() {
				update.LatestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Millisecond))
				Expect(update.HasUpdates()).To(BeTrue())
			})

			It("returns true when last import time is not nil", func() {
				update.LastImportTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Millisecond))
				Expect(update.HasUpdates()).To(BeTrue())
			})

			It("returns true when all fields are not nil", func() {
				update = dataSourceTest.RandomUpdate()
				Expect(update.HasUpdates()).To(BeTrue())
			})
		})
	})

	Context("Source", func() {
		DescribeTable("serializes the datum as expected",
			func(mutator func(datum *dataSource.Source)) {
				datum := dataSourceTest.RandomSource()
				mutator(datum)
				test.ExpectSerializedJSON(datum, dataSourceTest.NewObjectFromSource(datum, test.ObjectFormatJSON))
				test.ExpectSerializedBSON(datum, dataSourceTest.NewObjectFromSource(datum, test.ObjectFormatBSON))
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
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					dataSourceTest.ExpectEqualSource(datum, expectedDatum)
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {},
				),
				Entry("id missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "id")
						expectedDatum.ID = nil
					},
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
						valid := dataSourceTest.RandomID()
						object["id"] = valid
						expectedDatum.ID = pointer.FromString(valid)
					},
				),
				Entry("user id missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "userId")
						expectedDatum.UserID = nil
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
						valid := userTest.RandomID()
						object["userId"] = valid
						expectedDatum.UserID = pointer.FromString(valid)
					},
				),
				Entry("provider type missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "providerType")
						expectedDatum.ProviderType = nil
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
				Entry("provider name missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "providerName")
						expectedDatum.ProviderName = nil
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
				Entry("provider session id missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "providerSessionId")
						expectedDatum.ProviderSessionID = nil
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
				Entry("state missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "state")
						expectedDatum.State = nil
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
				Entry("error missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "error")
						expectedDatum.Error = nil
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
				Entry("data set ids missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "dataSetIds")
						expectedDatum.DataSetIDs = nil
					},
				),
				Entry("data set ids invalid type",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["dataSetIds"] = true
						expectedDatum.DataSetIDs = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotArray(true), "/dataSetIds"),
				),
				Entry("data set ids empty",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						object["dataSetIds"] = []string{}
						expectedDatum.DataSetIDs = pointer.FromStringArray([]string{})
					},
				),
				Entry("data set ids valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := dataTest.RandomSetIDs()
						object["dataSetIds"] = valid
						expectedDatum.DataSetIDs = pointer.FromStringArray(valid)
					},
				),
				Entry("earliest data time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "earliestDataTime")
						expectedDatum.EarliestDataTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["earliestDataTime"] = valid.Format(time.RFC3339)
						expectedDatum.EarliestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("latest data time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "latestDataTime")
						expectedDatum.LatestDataTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/latestDataTime"),
				),
				Entry("latest data time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["latestDataTime"] = valid.Format(time.RFC3339)
						expectedDatum.LatestDataTime = pointer.FromTime(valid)
					},
				),
				Entry("last import time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "lastImportTime")
						expectedDatum.LastImportTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["lastImportTime"] = valid.Format(time.RFC3339)
						expectedDatum.LastImportTime = pointer.FromTime(valid)
					},
				),
				Entry("created time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "createdTime")
						expectedDatum.CreatedTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/createdTime"),
				),
				Entry("created time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["createdTime"] = valid.Format(time.RFC3339)
						expectedDatum.CreatedTime = pointer.FromTime(valid)
					},
				),
				Entry("modified time missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "modifiedTime")
						expectedDatum.ModifiedTime = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorValueTimeNotParsable("invalid", time.RFC3339), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						valid := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second)
						object["modifiedTime"] = valid.Format(time.RFC3339)
						expectedDatum.ModifiedTime = pointer.FromTime(valid)
					},
				),
				Entry("revision missing",
					func(object map[string]interface{}, expectedDatum *dataSource.Source) {
						delete(object, "revision")
						expectedDatum.Revision = nil
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
						object["state"] = true
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
						expectedDatum.State = nil
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
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/state"),
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
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
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
						datum.ID = pointer.FromString(dataSourceTest.RandomID())
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
						datum.UserID = pointer.FromString(userTest.RandomID())
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
				Entry("provider session id missing",
					func(datum *dataSource.Source) { datum.ProviderSessionID = nil },
				),
				Entry("provider session id empty",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueEmpty(), "/providerSessionId"),
				),
				Entry("provider session id invalid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(auth.ErrorValueStringAsProviderSessionIDNotValid("invalid"), "/providerSessionId"),
				),
				Entry("provider session id valid",
					func(datum *dataSource.Source) {
						datum.ProviderSessionID = pointer.FromString(authTest.RandomProviderSessionID())
					},
				),
				Entry("state missing",
					func(datum *dataSource.Source) { datum.State = nil },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
				),
				Entry("state empty",
					func(datum *dataSource.Source) {
						datum.State = pointer.FromString("")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", dataSource.States()), "/state"),
				),
				Entry("state invalid",
					func(datum *dataSource.Source) {
						datum.State = pointer.FromString("invalid")
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", dataSource.States()), "/state"),
				),
				Entry("state valid",
					func(datum *dataSource.Source) {
						datum.State = pointer.FromString(dataSourceTest.RandomState())
					},
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
						dataSetID := dataTest.RandomSetID()
						datum.DataSetIDs = pointer.FromStringArray([]string{dataSetID, dataSetID})
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueDuplicate(), "/dataSetIds/1"),
				),
				Entry("data set ids valid",
					func(datum *dataSource.Source) {
						datum.DataSetIDs = pointer.FromStringArray([]string{dataTest.RandomSetID()})
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
						datum.EarliestDataTime = pointer.FromTime(futureTime)
						datum.LatestDataTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/earliestDataTime"),
				),
				Entry("earliest data time valid",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
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
						datum.LatestDataTime = pointer.FromTime(futureTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/latestDataTime"),
				),
				Entry("earliest data time missing; latest data time valid",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = nil
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
					},
				),
				Entry("earliest data time valid; latest data time missing",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = nil
					},
				),
				Entry("earliest data time valid; latest data time before earliest data time",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = pointer.FromTime(farPastTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(farPastTime, nearPastTime), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time after now",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = pointer.FromTime(futureTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/latestDataTime"),
				),
				Entry("earliest data time valid; latest data time valid",
					func(datum *dataSource.Source) {
						datum.EarliestDataTime = pointer.FromTime(nearPastTime)
						datum.LatestDataTime = pointer.FromTime(test.RandomTimeFromRange(nearPastTime, time.Now()).Truncate(time.Second))
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
						datum.LastImportTime = pointer.FromTime(futureTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/lastImportTime"),
				),
				Entry("last import time valid",
					func(datum *dataSource.Source) {
						datum.LastImportTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
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
						datum.CreatedTime = pointer.FromTime(futureTime)
						datum.ModifiedTime = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/createdTime"),
				),
				Entry("created time valid",
					func(datum *dataSource.Source) {
						datum.CreatedTime = pointer.FromTime(test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now()).Truncate(time.Second))
						datum.ModifiedTime = nil
					},
				),
				Entry("modified time missing",
					func(datum *dataSource.Source) { datum.ModifiedTime = nil },
				),
				Entry("modified time before created time",
					func(datum *dataSource.Source) {
						datum.CreatedTime = pointer.FromTime(nearPastTime)
						datum.ModifiedTime = pointer.FromTime(farPastTime)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotAfter(farPastTime, nearPastTime), "/modifiedTime"),
				),
				Entry("modified time after now",
					func(datum *dataSource.Source) { datum.ModifiedTime = pointer.FromTime(futureTime) },
					errorsTest.WithPointerSource(structureValidator.ErrorValueTimeNotBeforeNow(futureTime), "/modifiedTime"),
				),
				Entry("modified time valid",
					func(datum *dataSource.Source) {
						datum.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*datum.CreatedTime, time.Now()).Truncate(time.Second))
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
						datum.State = nil
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
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/state"),
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
					normalizer := structureNormalizer.New()
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
				Expect(sanitized.Sanitize(nil)).To(MatchError("unable to sanitize"))
			})

			It("does not modify the original when the details are server", func() {
				details := request.NewDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
				Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
				Expect(sanitized).To(Equal(original))
			})

			It("removes the provider session id and sanitizes the error when the details are user", func() {
				details := request.NewDetails(request.MethodSessionToken, userTest.RandomID(), authTest.NewSessionToken())
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
					details := request.NewDetails(request.MethodSessionToken, userTest.RandomID(), authTest.NewSessionToken())
					Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
					original.ProviderSessionID = nil
					original.Error.Error = errors.Sanitize(unauthenticatedError)
					Expect(sanitized).To(Equal(original))
				})
			})
		})
	})

	Context("Sources", func() {
		Context("Sanitize", func() {
			var originals dataSource.Sources
			var sanitized dataSource.Sources

			BeforeEach(func() {
				originals = dataSourceTest.RandomSources(1, 3)
				sanitized = dataSourceTest.CloneSources(originals)
			})

			It("returns an error when the details are missing", func() {
				Expect(sanitized.Sanitize(nil)).To(MatchError("unable to sanitize"))
			})

			It("does not modify the originals when the details are server", func() {
				details := request.NewDetails(request.MethodServiceSecret, "", authTest.NewSessionToken())
				Expect(sanitized.Sanitize(details)).ToNot(HaveOccurred())
				Expect(sanitized).To(Equal(originals))
			})

			It("removes the provider session id and sanitizes the error when the details are user", func() {
				details := request.NewDetails(request.MethodSessionToken, userTest.RandomID(), authTest.NewSessionToken())
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
		It("returns a string of 32 lowercase hexidecimal characters", func() {
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
