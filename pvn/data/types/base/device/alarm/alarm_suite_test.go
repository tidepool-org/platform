package alarm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAlarm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/device/alarm")
}
