package mailer_test

import (
	"context"

	"github.com/tidepool-org/go-common/events"

	"github.com/tidepool-org/platform/mailer"
	"github.com/tidepool-org/platform/mailer/test"

	"github.com/IBM/sarama"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mailer", func() {
	var (
		mailr      mailer.Mailer
		mockBroker *sarama.MockBroker
	)

	Describe("SendEmailTemplate", func() {
		BeforeEach(func() {
			t := GinkgoT()
			mockBroker = test.NewMockBroker(t)
			test.SetKafkaConfig(t, mockBroker)

			var err error
			mailr, err = mailer.Client()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			if mockBroker != nil {
				mockBroker.Close()
			}
		})

		It("should succeed", func() {
			Expect(mailr.SendEmailTemplate(context.Background(), events.SendEmailTemplateEvent{
				Recipient: "test@tidepool.org",
				Template:  "test_email",
			})).To(Succeed())
		})
	})
})
