package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/pvn/data/types"
	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Errors", func() {
	DescribeTable("all errors",
		func(err *service.Error, code string, title string, detail string) {
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(code))
			Expect(err.Title).To(Equal(title))
			Expect(err.Detail).To(Equal(detail))
		},
		Entry("ErrorValueMissing", types.ErrorValueMissing(), "value-missing", "value is missing", "Value is missing"),
		Entry("ErrorTypeInvalid", types.ErrorTypeInvalid("unknown"), "type-invalid", "type is invalid", "Type \"unknown\" is invalid"),
		Entry("ErrorSubTypeInvalid", types.ErrorSubTypeInvalid("unknown"), "sub-type-invalid", "sub type is invalid", "Sub type \"unknown\" is invalid"),
		Entry("ErrorDeliveryTypeInvalid", types.ErrorDeliveryTypeInvalid("unknown"), "delivery-type-invalid", "delivery type is invalid", "Delivery type \"unknown\" is invalid"),
	)
})
