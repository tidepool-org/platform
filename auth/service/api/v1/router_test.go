package v1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authServiceApiV1 "github.com/tidepool-org/platform/auth/service/api/v1"
	serviceTest "github.com/tidepool-org/platform/auth/service/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
)

var _ = Describe("Router", func() {
	var svc *serviceTest.Service

	BeforeEach(func() {
		svc = serviceTest.NewService()
	})

	Context("NewRouter", func() {
		It("returns an error if context is missing", func() {
			rtr, err := authServiceApiV1.NewRouter(nil)
			errorsTest.ExpectEqual(err, errors.New("service is missing"))
			Expect(rtr).To(BeNil())
		})

		It("returns successfully", func() {
			rtr, err := authServiceApiV1.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})
	})

	Context("with new router", func() {
		var rtr *authServiceApiV1.Router

		BeforeEach(func() {
			var err error
			rtr, err = authServiceApiV1.NewRouter(svc)
			Expect(err).ToNot(HaveOccurred())
			Expect(rtr).ToNot(BeNil())
		})

		Context("Routes", func() {
			It("returns the expected routes", func() {
				Expect(rtr.Routes()).ToNot(BeEmpty())
			})
		})
	})
})
