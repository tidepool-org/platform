package basal

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBasal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Basal Suite")
}
