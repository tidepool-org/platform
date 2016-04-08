package device

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDeviceevent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deviceevent Suite")
}
