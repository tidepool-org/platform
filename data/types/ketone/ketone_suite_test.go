package ketone

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKetone(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ketone Suite")
}
