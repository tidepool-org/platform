package suspend_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSuspend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/basal/suspend")
}
