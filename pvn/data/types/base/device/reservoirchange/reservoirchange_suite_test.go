package reservoirchange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestReservoirChange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/device/reservoirchange")
}
