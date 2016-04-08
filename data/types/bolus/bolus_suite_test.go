package bolus

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBolus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bolus Suite")
}
