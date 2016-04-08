package bloodglucose

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBloodglucose(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bloodglucose Suite")
}
