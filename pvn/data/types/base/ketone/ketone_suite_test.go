package ketone_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKetone(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/ketone")
}
