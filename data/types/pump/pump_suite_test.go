package pump

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPump(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pump Suite")
}
