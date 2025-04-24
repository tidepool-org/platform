package errors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	logTest "github.com/tidepool-org/platform/log/test"
	structureNormalizer "github.com/tidepool-org/platform/structure/normalizer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Caller", func() {
	Context("GetCaller", func() {
		It("returns the Caller", func() {
			caller := errors.GetCaller(-1) // Hack to get caller from errors.go
			Expect(caller).ToNot(BeNil())
			Expect(caller.Package).To(Equal("github.com/tidepool-org/platform/errors"))
			Expect(caller.Function).To(Equal("GetCaller"))
			Expect(caller.File).To(Equal("errors/caller.go"))
			Expect(caller.Line).To(Equal(19))
		})
	})

	Context("Caller", func() {
		Context("PackageName", func() {
			It("returns the package name", func() {
				caller := errors.GetCaller(-1) // Hack to get caller from errors.go
				Expect(caller).ToNot(BeNil())
				Expect(caller.PackageName()).To(Equal("errors"))
			})
		})

		Context("FileName", func() {
			It("returns the file name", func() {
				caller := errors.GetCaller(-1) // Hack to get caller from errors.go
				Expect(caller).ToNot(BeNil())
				Expect(caller.FileName()).To(Equal("caller.go"))
			})
		})

		Context("Parse", func() {
			It("successefully parses the data", func() {
				object := &map[string]any{
					"package":  "test-package",
					"function": "test-function",
					"file":     "test-file",
					"line":     123,
				}
				parser := structureParser.NewObject(logTest.NewLogger(), object)
				caller := &errors.Caller{}
				caller.Parse(parser)
				Expect(caller.Package).To(Equal("test-package"))
				Expect(caller.Function).To(Equal("test-function"))
				Expect(caller.File).To(Equal("test-file"))
				Expect(caller.Line).To(Equal(123))
				Expect(parser.Error()).ToNot(HaveOccurred())
			})

			It("reports errors", func() {
				object := &map[string]any{
					"package":  true,
					"function": true,
					"file":     true,
					"line":     true,
				}
				parser := structureParser.NewObject(logTest.NewLogger(), object)
				caller := &errors.Caller{}
				caller.Parse(parser)
				Expect(caller.Package).To(BeZero())
				Expect(caller.Function).To(BeZero())
				Expect(caller.File).To(BeZero())
				Expect(caller.Line).To(BeZero())
				Expect(parser.Error()).To(MatchError("type is not string, but bool, type is not string, but bool, type is not string, but bool, type is not int, but bool"))
			})
		})

		Context("Validate", func() {
			It("successfully validates the caller", func() {
				validator := structureValidator.New(logTest.NewLogger())
				caller := &errors.Caller{
					Package:  "test-package",
					Function: "test-function",
					File:     "test-file",
					Line:     123,
				}
				caller.Validate(validator)
				Expect(validator.Error()).ToNot(HaveOccurred())
			})

			It("reports errors", func() {
				validator := structureValidator.New(logTest.NewLogger())
				caller := &errors.Caller{
					Package:  "",
					Function: "",
					File:     "",
					Line:     -1,
				}
				caller.Validate(validator)
				Expect(validator.Error()).To(MatchError("value is empty, value is empty, value is empty, value -1 is not greater than or equal to 0"))
			})
		})

		Context("Normalize", func() {
			It("successfully normalizes the caller", func() {
				normalizer := structureNormalizer.New(logTest.NewLogger())
				caller := &errors.Caller{
					Package:  "test-package",
					Function: "test-function",
					File:     "test-file",
					Line:     123,
				}
				caller.Normalize(normalizer)
				Expect(normalizer.Error()).ToNot(HaveOccurred())
			})
		})
	})
})
