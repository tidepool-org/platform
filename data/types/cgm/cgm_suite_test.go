package cgm_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCgm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/cgm")
}
