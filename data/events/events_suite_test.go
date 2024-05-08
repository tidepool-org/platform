package events

import (
	"log/slog"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var _ = BeforeSuite(func() {
	slog.SetDefault(devNullSlogLogger(GinkgoT()))
})

// Cleaner is part of testing.T and FullGinkgoTInterface
type Cleaner interface {
	Cleanup(func())
}

func devNullSlogLogger(c Cleaner) *slog.Logger {
	f, err := os.Open(os.DevNull)
	Expect(err).To(Succeed())
	c.Cleanup(func() {
		Expect(f.Close()).To(Succeed())
	})
	return slog.New(slog.NewTextHandler(f, nil))
}
