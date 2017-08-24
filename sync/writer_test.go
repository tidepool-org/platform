package sync_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"

	"github.com/tidepool-org/platform/sync"
)

var _ = Describe("Writer", func() {
	Context("NewWriter", func() {
		It("returns an error if the writer is missng", func() {
			writer, err := sync.NewWriter(nil)
			Expect(err).To(MatchError("sync: writer is missing"))
			Expect(writer).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(sync.NewWriter(ioutil.Discard)).ToNot(BeNil())
		})
	})

	Context("with new writer", func() {
		var writer *sync.Writer

		BeforeEach(func() {
			var err error
			writer, err = sync.NewWriter(ioutil.Discard)
			Expect(err).ToNot(HaveOccurred())
			Expect(writer).ToNot(BeNil())
		})

		Context("Write", func() {
			It("returns successfully", func() {
				Expect(writer.Write([]byte("Writing Test!"))).To(Equal(13))
			})
		})
	})
})
