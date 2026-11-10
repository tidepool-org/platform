package request_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cryptoTest "github.com/tidepool-org/platform/crypto/test"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

func generateNewUniqueKey[A any](f func() string, existingKeys map[string]A) string {
	maxLoops := 1000
	for range maxLoops {
		gen := f()
		if _, found := existingKeys[gen]; !found {
			return gen
		}
	}
	Fail(fmt.Sprintf("unable to generate a unique key after %d tries", maxLoops), 1)
	return ""
}

var _ = Describe("Parser", func() {
	Context("DecodeOption", func() {
		Context("IgnoreNotParsed", func() {
			It("returns false by default", func() {
				option := request.DecodeOption{}
				Expect(option.IgnoreNotParsed()).To(BeFalse())
			})

			It("returns true if set", func() {
				option := request.IgnoreNotParsed()
				Expect(option.IgnoreNotParsed()).To(BeTrue())
			})
		})
	})

	Context("DecodeOptions", func() {
		It("merges options", func() {
			resolvedOption := request.DecodeOptions([]request.DecodeOption{{}, request.IgnoreNotParsed(), {}})
			Expect(resolvedOption.IgnoreNotParsed()).To(BeTrue())
		})
	})

	Context("IgnoreNotParsed", func() {
		It("returns true", func() {
			option := request.IgnoreNotParsed()
			Expect(option.IgnoreNotParsed()).To(BeTrue())
		})
	})

	Context("with header", func() {
		var header http.Header
		var key string

		BeforeEach(func() {
			header = testHttp.RandomHeader()
			key = generateNewUniqueKey(testHttp.NewHeaderKey, header)
		})

		Context("ParseSingletonHeader", func() {
			var value string

			BeforeEach(func() {
				value = testHttp.NewHeaderValue()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{testHttp.NewHeaderValue(), value, testHttp.NewHeaderValue()}
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one", func() {
				header[key] = []string{value}
				result, err := request.ParseSingletonHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})

		Context("ParseDigestMD5Header", func() {
			var value string

			BeforeEach(func() {
				value = cryptoTest.RandomBase64EncodedMD5Hash()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{testHttp.NewHeaderValue(), fmt.Sprintf("MD5=%s", value), testHttp.NewHeaderValue()}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{cryptoTest.RandomBase64EncodedMD5Hash()}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value algorithm is not MD5", func() {
				header[key] = []string{fmt.Sprintf("SHA1=%s", value)}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value digest is not valid", func() {
				header[key] = []string{fmt.Sprintf("MD5=%s", test.RandomStringFromRangeAndCharset(20, 20, test.CharsetAlphaNumeric))}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{fmt.Sprintf("MD5=%s", value)}
				result, err := request.ParseDigestMD5Header(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})

		Context("ParseMediaTypeHeader", func() {
			var value string

			BeforeEach(func() {
				value = netTest.RandomMediaType()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{netTest.RandomMediaType(), value, netTest.RandomMediaType()}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{"/"}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{value}
				result, err := request.ParseMediaTypeHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})

		Context("ParseTimeHeader", func() {
			var value string
			var layout string

			BeforeEach(func() {
				layout = time.RFC3339
				value = time.Now().Format(layout)
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{time.Now().Add(1 * time.Hour).Format(layout), value}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{"abc"}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{value}
				result, err := request.ParseTimeHeader(header, key, layout)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(result.Format(layout)).To(Equal(value))
			})
		})

		Context("ParseIntHeader", func() {
			var value int

			BeforeEach(func() {
				value = test.RandomInt()
			})

			It("returns nil if the header is nil", func() {
				header = nil
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if the key is not present", func() {
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns nil if there are no values", func() {
				header[key] = []string{}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("returns an error if there is more than one value", func() {
				header[key] = []string{strconv.Itoa(test.RandomInt()), strconv.Itoa(value), strconv.Itoa(test.RandomInt())}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns an error if the value is not valid", func() {
				header[key] = []string{"abc"}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).To(MatchError(fmt.Sprintf("header %q is invalid", key)))
				Expect(result).To(BeNil())
			})

			It("returns the value if there is exactly one and it is valid", func() {
				header[key] = []string{strconv.Itoa(value)}
				result, err := request.ParseIntHeader(header, key)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(value))
			})
		})
	})
	Context("with context and source", func() {
		var ctx context.Context
		var source structure.Source

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			source = structure.NewPointerSource()
		})

		Context("ObjectInspectable", func() {
			It("inspects the object before parsing a stream", func() {
				parsable := &inspectableObject{}
				Expect(request.ParseStream(ctx, source, strings.NewReader(`{"value": "a"}`), parsable)).To(Succeed())
				Expect(parsable.inspected).To(Equal(map[string]any{"value": "a"}))
				Expect(parsable.inspectedBeforeParse).To(BeTrue())
				Expect(parsable.Value).To(Equal(pointer.FromString("a")))
			})

			It("inspects the object before parsing an object", func() {
				parsable := &inspectableObject{}
				Expect(request.ParseObject(ctx, source, map[string]any{"value": "a"}, parsable)).To(Succeed())
				Expect(parsable.inspected).To(Equal(map[string]any{"value": "a"}))
				Expect(parsable.inspectedBeforeParse).To(BeTrue())
				Expect(parsable.Value).To(Equal(pointer.FromString("a")))
			})

			It("returns the inspect error without parsing a stream", func() {
				err := errorsTest.RandomError()
				parsable := &inspectableObject{inspectErr: err}
				Expect(request.ParseStream(ctx, source, strings.NewReader(`{"value": "a"}`), parsable)).To(Equal(err))
				Expect(parsable.parsed).To(BeFalse())
			})

			It("returns the inspect error without parsing an object", func() {
				err := errorsTest.RandomError()
				parsable := &inspectableObject{inspectErr: err}
				Expect(request.ParseObject(ctx, source, map[string]any{"value": "a"}, parsable)).To(Equal(err))
				Expect(parsable.parsed).To(BeFalse())
			})

			It("does not inspect when the object is missing", func() {
				parsable := &inspectableObject{}
				Expect(request.ParseObject(ctx, source, nil, parsable)).To(MatchError("object is missing"))
				Expect(parsable.inspected).To(BeNil())
				Expect(parsable.parsed).To(BeFalse())
			})

			It("does not inspect a type that is not object parsable when parsing a stream", func() {
				parsable := &inspectableSimpleObject{}
				Expect(request.ParseStream(ctx, source, strings.NewReader(`{"value": "a"}`), parsable)).To(Succeed())
				Expect(parsable.inspected).To(BeFalse())
				Expect(parsable.Value).To(Equal("a"))
			})

			It("does not inspect a type that is not object parsable when parsing an object", func() {
				parsable := &inspectableSimpleObject{}
				Expect(request.ParseObject(ctx, source, map[string]any{"value": "a"}, parsable)).To(Succeed())
				Expect(parsable.inspected).To(BeFalse())
				Expect(parsable.Value).To(Equal("a"))
			})

			It("parses a missing object into a type that is not object parsable without error", func() {
				parsable := &inspectableSimpleObject{}
				Expect(request.ParseObject(ctx, source, nil, parsable)).To(Succeed())
				Expect(parsable.inspected).To(BeFalse())
				Expect(parsable.Value).To(BeEmpty())
			})
		})

		Context("ArrayInspectable", func() {
			It("inspects the array before parsing a stream", func() {
				parsable := &inspectableArray{}
				Expect(request.ParseStream(ctx, source, strings.NewReader(`["a"]`), parsable)).To(Succeed())
				Expect(parsable.inspected).To(Equal([]any{"a"}))
				Expect(parsable.inspectedBeforeParse).To(BeTrue())
				Expect(parsable.Value).To(Equal(pointer.FromString("a")))
			})

			It("inspects the array before parsing an array", func() {
				parsable := &inspectableArray{}
				Expect(request.ParseArray(ctx, source, []any{"a"}, parsable)).To(Succeed())
				Expect(parsable.inspected).To(Equal([]any{"a"}))
				Expect(parsable.inspectedBeforeParse).To(BeTrue())
				Expect(parsable.Value).To(Equal(pointer.FromString("a")))
			})

			It("returns the inspect error without parsing a stream", func() {
				err := errorsTest.RandomError()
				parsable := &inspectableArray{inspectErr: err}
				Expect(request.ParseStream(ctx, source, strings.NewReader(`["a"]`), parsable)).To(Equal(err))
				Expect(parsable.parsed).To(BeFalse())
			})

			It("returns the inspect error without parsing an array", func() {
				err := errorsTest.RandomError()
				parsable := &inspectableArray{inspectErr: err}
				Expect(request.ParseArray(ctx, source, []any{"a"}, parsable)).To(Equal(err))
				Expect(parsable.parsed).To(BeFalse())
			})

			It("does not inspect when the array is missing", func() {
				parsable := &inspectableArray{}
				Expect(request.ParseArray(ctx, source, nil, parsable)).To(MatchError("array is missing"))
				Expect(parsable.inspected).To(BeNil())
				Expect(parsable.parsed).To(BeFalse())
			})

			It("does not inspect a type that is not array parsable when parsing a stream", func() {
				parsable := &inspectableSimpleArray{}
				Expect(request.ParseStream(ctx, source, strings.NewReader(`["a"]`), parsable)).To(Succeed())
				Expect(parsable.inspected).To(BeFalse())
				Expect(parsable.values).To(Equal([]string{"a"}))
			})

			It("does not inspect a type that is not array parsable when parsing an array", func() {
				parsable := &inspectableSimpleArray{}
				Expect(request.ParseArray(ctx, source, []any{"a"}, parsable)).To(Succeed())
				Expect(parsable.inspected).To(BeFalse())
				Expect(parsable.values).To(Equal([]string{"a"}))
			})

			It("parses a missing array into a type that is not array parsable without error", func() {
				parsable := &inspectableSimpleArray{}
				Expect(request.ParseArray(ctx, source, nil, parsable)).To(Succeed())
				Expect(parsable.inspected).To(BeFalse())
				Expect(parsable.values).To(BeEmpty())
			})
		})
	})
})

type inspectableObject struct {
	Value                *string
	inspected            map[string]any
	inspectedBeforeParse bool
	inspectErr           error
	parsed               bool
}

func (i *inspectableObject) Inspect(object map[string]any) error {
	i.inspected = object
	i.inspectedBeforeParse = !i.parsed
	return i.inspectErr
}

func (i *inspectableObject) Parse(parser structure.ObjectParser) {
	i.parsed = true
	i.Value = parser.String("value")
}

type inspectableArray struct {
	Value                *string
	inspected            []any
	inspectedBeforeParse bool
	inspectErr           error
	parsed               bool
}

func (i *inspectableArray) Inspect(array []any) error {
	i.inspected = array
	i.inspectedBeforeParse = !i.parsed
	return i.inspectErr
}

func (i *inspectableArray) Parse(parser structure.ArrayParser) {
	i.parsed = true
	i.Value = parser.String(0)
}

// Implements ObjectInspectable, but not ObjectParsable, so is decoded directly from JSON and never inspected
type inspectableSimpleObject struct {
	Value     string `json:"value"`
	inspected bool
}

func (i *inspectableSimpleObject) Inspect(object map[string]any) error {
	i.inspected = true
	return nil
}

// Implements ArrayInspectable, but not ArrayParsable, so is decoded directly from JSON and never inspected
type inspectableSimpleArray struct {
	values    []string
	inspected bool
}

func (i *inspectableSimpleArray) Inspect(array []any) error {
	i.inspected = true
	return nil
}

func (i *inspectableSimpleArray) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &i.values)
}
