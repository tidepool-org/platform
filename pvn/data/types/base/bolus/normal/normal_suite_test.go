package normal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNormal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/bolus/normal")
}
