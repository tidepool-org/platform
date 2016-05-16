package combination_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCombination(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pvn/data/types/base/bolus/combination")
}
