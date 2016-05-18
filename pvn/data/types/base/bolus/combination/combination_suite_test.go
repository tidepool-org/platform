package combination_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCombination(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/bolus/combination")
}
