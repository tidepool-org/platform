package temporary_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTemporary(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/basal/temporary")
}
