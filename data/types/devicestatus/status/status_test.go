package status_test

import (
	"github.com/tidepool-org/platform/data/types/devicestatus/status"

	. "github.com/onsi/ginkgo"
)

func NewStatus() *status.Array {
	datum := *status.NewStatusArray()
	return &datum
}

func CloneStatusArray(datum *status.Array) *status.Array {
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
