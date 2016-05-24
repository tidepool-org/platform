package calibration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCalibration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/device/calibration")
}
