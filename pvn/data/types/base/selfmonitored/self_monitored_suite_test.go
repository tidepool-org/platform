package selfmonitored_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSelfMonitored(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/selfmonitored")
}
