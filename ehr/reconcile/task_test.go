package reconcile_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/ehr/reconcile"
)

var _ = Describe("Task", func() {
	Describe("NewTaskCreate", func() {
		It("returns a task create", func() {
			create := reconcile.NewTaskCreate()
			Expect(create).ToNot(BeNil())
			Expect(create.Name).To(PointTo(Equal(reconcile.Type)))
			Expect(create.Type).To(Equal(reconcile.Type))
			Expect(create.AvailableTime).To(PointTo(BeTemporally("~", time.Now(), 3*time.Second)))
		})
	})
})
