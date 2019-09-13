package status_test

import (
	"github.com/tidepool-org/platform/data/types/devicestatus/status"

	. "github.com/onsi/ginkgo"
)

func NewStatus() *status.TypeStatusArray {
	datum := *status.NewStatusArray()
	return &datum
}

func CloneStatusArray(datum *status.TypeStatusArray) *status.TypeStatusArray {
	if datum == nil {
		return nil
	}
	clone := status.NewStatusArray()
	return clone
}

var _ = Describe("Status", func() {

	Context("Status", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
		})

		Context("Normalize", func() {
		})
	})
})
