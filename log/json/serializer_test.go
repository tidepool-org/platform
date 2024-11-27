package json_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logJson "github.com/tidepool-org/platform/log/json"
)

type WriteOutput struct {
	BytesWritten int
	Error        error
}

type Writer struct {
	WriteInvocations int
	WriteInputs      [][]byte
	WriteOutputs     []WriteOutput
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Write(bites []byte) (int, error) {
	w.WriteInvocations++

	w.WriteInputs = append(w.WriteInputs, bites)

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
			serializer, err := logJson.NewSerializer(nil)
			Expect(err).To(MatchError("writer is missing"))
			Expect(serializer).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(logJson.NewSerializer(writer)).ToNot(BeNil())
		})
	})

	Context("with new serializer", func() {
		var serializer log.Serializer

		BeforeEach(func() {
			var err error
			serializer, err = logJson.NewSerializer(writer)
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
