package calculator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCalculator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/bolus/calculator")
}
