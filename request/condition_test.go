package request_test

import (
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Condition", func() {
	Context("NewCondition", func() {
		It("returns successfully with default values", func() {
			condition := request.NewCondition()
			Expect(condition).ToNot(BeNil())
			Expect(condition.Revision).To(BeNil())
		})
	})

	Context("Condition", func() {
		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]interface{}, expectedDatum *request.Condition), expectedErrors ...error) {
					expectedDatum := requestTest.RandomCondition()
					object := requestTest.NewObjectFromCondition(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &request.Condition{}
					errorsTest.ExpectEqual(structureParser.NewObject(&object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]interface{}, expectedDatum *request.Condition) {},
				),
				Entry("revision missing",
					func(object map[string]interface{}, expectedDatum *request.Condition) {
						delete(object, "revision")
						expectedDatum.Revision = nil
					},
				),
				Entry("revision invalid type",
					func(object map[string]interface{}, expectedDatum *request.Condition) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
				Entry("revision valid",
					func(object map[string]interface{}, expectedDatum *request.Condition) {
						valid := requestTest.RandomRevision()
						object["revision"] = valid
						expectedDatum.Revision = pointer.FromInt(valid)
					},
				),
				Entry("multiple",
					func(object map[string]interface{}, expectedDatum *request.Condition) {
						object["revision"] = true
						expectedDatum.Revision = nil
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/revision"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *request.Condition), expectedErrors ...error) {
					datum := requestTest.RandomCondition()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *request.Condition) {},
				),
				Entry("revision missing",
					func(datum *request.Condition) { datum.Revision = nil },
				),
				Entry("revision out of range (lower)",
					func(datum *request.Condition) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
				Entry("revision in range (lower)",
					func(datum *request.Condition) {
						datum.Revision = pointer.FromInt(0)
					},
				),
				Entry("multiple errors",
					func(datum *request.Condition) {
						datum.Revision = pointer.FromInt(-1)
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(-1, 0), "/revision"),
				),
			)
		})

		Context("with new condition", func() {
			var condition *request.Condition

			BeforeEach(func() {
				condition = requestTest.RandomCondition()
			})

			Context("MutateRequest", func() {
				var req *http.Request

				BeforeEach(func() {
					req = testHttp.NewRequest()
				})

				It("returns an error when the request is missing", func() {
					errorsTest.ExpectEqual(condition.MutateRequest(nil), errors.New("request is missing"))
				})

				It("sets request query as expected", func() {
					Expect(condition.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(Equal(url.Values{
						"revision": []string{strconv.Itoa(*condition.Revision)},
					}))
				})

				It("does not set request query when the condition is empty", func() {
					condition.Revision = nil
					Expect(condition.MutateRequest(req)).To(Succeed())
					Expect(req.URL.Query()).To(BeEmpty())
				})
			})
		})
	})
})
