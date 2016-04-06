package device

import (
	"testing"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

func TestDeviceevent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deviceevent Suite")
}
