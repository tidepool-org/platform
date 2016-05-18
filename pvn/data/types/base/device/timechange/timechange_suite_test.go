package timechange_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTimeChange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/device/timechange")
}
