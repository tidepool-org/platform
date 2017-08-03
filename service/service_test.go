package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
	_ "github.com/tidepool-org/platform/version/test"
)

var _ = Describe("Service", func() {
	Context("New", func() {
		It("returns an error if the name is missing", func() {
			app, err := service.New("", "TIDEPOOL")
			Expect(err).To(MatchError("application: name is missing"))
			Expect(app).To(BeNil())
		})

		It("returns an error if the prefix is missing", func() {
			app, err := service.New("test", "")
			Expect(err).To(MatchError("application: prefix is missing"))
			Expect(app).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(service.New("test", "TIDEPOOL")).ToNot(BeNil())
		})
	})
})
