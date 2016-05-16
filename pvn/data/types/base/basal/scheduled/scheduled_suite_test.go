package scheduled_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSuspend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pvn/data/types/base/basal/scheduled")
}
