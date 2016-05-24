package pump_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPump(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/pump")
}
