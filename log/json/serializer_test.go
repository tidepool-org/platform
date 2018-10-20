package json_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/test"
)

type WriteOutput struct {
	BytesWritten int
	Error        error
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

func (w *Writer) Write(bytes []byte) (int, error) {
	w.WriteInvocations++

	w.WriteInputs = append(w.WriteInputs, bytes)

	if len(w.WriteOutputs) == 0 {
		panic("Unexpected invocation of Write on Writer")
	}

	output := w.WriteOutputs[0]
	w.WriteOutputs = w.WriteOutputs[1:]
	return output.BytesWritten, output.Error
}

func (w *Writer) UnusedOutputsCount() int {
	return len(w.WriteOutputs)
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
			Expect(err).To(MatchError("writer is missing"))
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
				Expect(serializer.Serialize(nil)).To(MatchError("fields are missing"))
			})

			It("returns an error if unable to serialize fields", func() {
				Expect(serializer.Serialize(log.Fields{"a": func() {}})).To(MatchError("unable to serialize fields; json: unsupported type: func()"))
			})

			It("returns an error if an error is returned when writing bytes to writer", func() {
				writer.WriteOutputs = []WriteOutput{{BytesWritten: 0, Error: errors.New("test error")}}
				Expect(serializer.Serialize(log.Fields{})).To(MatchError("unable to write serialized field; test error"))
			})

			It("returns successfully after writing buffer with empty fields", func() {
				writer.WriteOutputs = []WriteOutput{{BytesWritten: 0, Error: nil}}
				Expect(serializer.Serialize(log.Fields{})).To(Succeed())
				Expect(writer.WriteInputs).To(Equal([][]byte{[]byte("{}\n")}))
			})

			It("returns successfully after writing buffer with non-empty fields", func() {
				writer.WriteOutputs = []WriteOutput{{BytesWritten: 0, Error: nil}}
				Expect(serializer.Serialize(log.Fields{"b": "right", "a": "left"})).To(Succeed())
				Expect(writer.WriteInputs).To(Equal([][]byte{[]byte("{\"a\":\"left\",\"b\":\"right\"}\n")}))
			})
		})
	})
})
