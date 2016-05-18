package scheduled_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestScheduled(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/basal/scheduled")
}
