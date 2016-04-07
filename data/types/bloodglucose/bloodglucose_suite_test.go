package bloodglucose

import (
	"testing"

	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"
)

func TestBloodglucose(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bloodglucose Suite")
}
