package json_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/test"
)

type WriteOutput struct {
	Count int
	Error error
}

type Writer struct {
	*test.Mock
	WriteInvocations int
	WriteInputs      [][]byte
	WriteOutputs     []WriteOutput
}

func NewWriter() *Writer {
	return &Writer{
		Mock: test.NewMock(),
	}
}

func (t *Writer) Write(bytes []byte) (int, error) {
	t.WriteInvocations++

	t.WriteInputs = append(t.WriteInputs, bytes)

	if len(t.WriteOutputs) == 0 {
		panic("Unexpected invocation of Write on Writer")
	}

	output := t.WriteOutputs[0]
	t.WriteOutputs = t.WriteOutputs[1:]
	return output.Count, output.Error
}

func (t *Writer) UnusedOutputsCount() int {
	return len(t.WriteOutputs)
}

var _ = Describe("JSON", func() {
	var writer *Writer

	BeforeEach(func() {
		writer = NewWriter()
	})

	AfterEach(func() {
		Expect(writer.UnusedOutputsCount()).To(Equal(0))
	})

	Context("NewSerializer", func() {
		It("returns an error if writer is missing", func() {
			serializer, err := json.NewSerializer(nil)
			Expect(err).To(MatchError("json: writer is missing"))
			Expect(serializer).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(json.NewSerializer(writer)).ToNot(BeNil())
		})
	})

	Context("with new serializer", func() {
		var serializer log.Serializer

		BeforeEach(func() {
			var err error
			serializer, err = json.NewSerializer(writer)
			Expect(err).ToNot(HaveOccurred())
			Expect(serializer).ToNot(BeNil())
		})

		Context("Serialize", func() {
			It("returns an error if fields are missing", func() {
				Expect(serializer.Serialize(nil)).To(MatchError("json: fields are missing"))
			})

			It("returns an error if unable to serialize fields", func() {
				Expect(serializer.Serialize(log.Fields{"a": func() {}})).To(MatchError("json: unable to serialize fields; json: unsupported type: func()"))
			})

			It("returns an error if an error is returned when writing bytes to writer", func() {
				writer.WriteOutputs = []WriteOutput{{Count: 0, Error: errors.New("test error")}}
				Expect(serializer.Serialize(log.Fields{})).To(MatchError("json: unable to write serialized field; test error"))
			})

			It("returns successfully after writing buffer with empty fields", func() {
				writer.WriteOutputs = []WriteOutput{{Count: 0, Error: nil}}
				Expect(serializer.Serialize(log.Fields{})).To(Succeed())
				Expect(writer.WriteInputs).To(Equal([][]byte{[]byte("{}\n")}))
			})

			It("returns successfully after writing buffer with non-empty fields", func() {
				writer.WriteOutputs = []WriteOutput{{Count: 0, Error: nil}}
				Expect(serializer.Serialize(log.Fields{"b": "right", "a": "left"})).To(Succeed())
				Expect(writer.WriteInputs).To(Equal([][]byte{[]byte("{\"a\":\"left\",\"b\":\"right\"}\n")}))
			})
		})
	})
})
