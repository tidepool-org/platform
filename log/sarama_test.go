package log_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logtest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("NewSarama", func() {
	It("initializes a new sarama log adapter", func() {
		testLog := logtest.NewLogger()
		saramaLog := log.NewSarama(testLog)
		Expect(saramaLog).ToNot(Equal(nil))
	})

	It("implements Print", func() {
		testLog := logtest.NewLogger()
		saramaLog := log.NewSarama(testLog)
		Expect(saramaLog).ToNot(Equal(nil))

		saramaLog.Print("testing 1 2 3")

		testLog.AssertInfo("testing 1 2 3")
	})

	It("implements Printf", func() {
		testLog := logtest.NewLogger()
		saramaLog := log.NewSarama(testLog)
		Expect(saramaLog).ToNot(Equal(nil))

		saramaLog.Printf("testing %s", "4 5 6")

		testLog.AssertInfo("testing 4 5 6")
	})

	It("implements Println", func() {
		testLog := logtest.NewLogger()
		saramaLog := log.NewSarama(testLog)
		Expect(saramaLog).ToNot(Equal(nil))

		saramaLog.Println("testing 7 8 9")

		testLog.AssertInfo("testing 7 8 9")
	})
})
