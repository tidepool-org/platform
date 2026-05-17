package webhook_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/oura"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
)

var _ = Describe("Webhook", func() {
	It("EventPath is expected", func() {
		Expect(ouraWebhook.EventPath).To(Equal("/event"))
	})

	Context("CallbackURLForEvent", func() {
		DescribeTable("return the expected results when the input",
			func(partnerURL string, eventType string, dataType string, expectedURL string) {
				Expect(ouraWebhook.CallbackURLForEvent(partnerURL, eventType, dataType)).To(Equal(expectedURL))
			},
			Entry("has a partner url that ends with a slash", "https://test.tidepool.org/v1/partners/oura/", oura.EventTypeCreate, oura.DataTypeDailyActivity, "https://test.tidepool.org/v1/partners/oura/event/create/daily_activity"),
			Entry("has an event type with a slash", "https://test.tidepool.org/v1/partners/oura", "with/slash", oura.DataTypeDailyActivity, "https://test.tidepool.org/v1/partners/oura/event/with%2Fslash/daily_activity"),
			Entry("has an data type with a slash", "https://test.tidepool.org/v1/partners/oura", oura.EventTypeCreate, "with/slash", "https://test.tidepool.org/v1/partners/oura/event/create/with%2Fslash"),
			Entry("is valid", "https://test.tidepool.org/v1/partners/oura", oura.EventTypeCreate, oura.DataTypeDailyActivity, "https://test.tidepool.org/v1/partners/oura/event/create/daily_activity"),
		)
	})

	Context("VerificationTokenForCallbackURL", func() {
		DescribeTable("return the expected results when the input",
			func(callbackURL string, partnerSecret string, expectedVerificationToken string) {
				Expect(ouraWebhook.VerificationTokenForCallbackURL(callbackURL, partnerSecret)).To(Equal(expectedVerificationToken))
			},
			Entry("has an empty callback url and secret", "", "", "e7ac0786668e0ff0f02b62bd04f45ff636fd82db63b1104601c975dc005f3a67"),
			Entry("has an empty callback url", "", "test-secret", "2024cd3527918eec9dd73d628320d5b830b1e280ce073bde14503a56d9d260ba"),
			Entry("has an empty secret", "https://test.tidepool.org/v1/partners/oura/event/create/daily_activity", "", "f213cb711a907c9a184f3fbd139f0008ec1d51b4130ec6310d51ffe8462aee08"),
			Entry("is valid", "https://test.tidepool.org/v1/partners/oura/event/create/daily_activity", "test-secret", "005470cbc365f3448d47046ae04c7126ee56510fd4ebe501c7b5bb9d200801d9"),
		)
	})
})
