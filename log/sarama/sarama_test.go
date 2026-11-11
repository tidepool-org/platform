package sarama_test

import (
	"github.com/IBM/sarama"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	logSarama "github.com/tidepool-org/platform/log/sarama"
	logTest "github.com/tidepool-org/platform/log/test"
)

var _ = Describe("Sarama", func() {
	var logger *logTest.Logger

	BeforeEach(func() {
		logger = logTest.NewLogger()
	})

	Context("Setup", func() {
		It("returns an error if the logger is missing", func() {
			Expect(logSarama.Setup(nil)).To(MatchError("logger is missing"))
		})

		It("returns successfully", func() {
			Expect(logSarama.Setup(logger)).To(Succeed())
		})
	})

	Context("with setup", func() {
		BeforeEach(func() {
			Expect(logSarama.Setup(logger)).To(Succeed())
		})

		It("keeps the logger if the replacement is missing", func() {
			Expect(logSarama.Setup(nil)).To(MatchError("logger is missing"))
			sarama.Logger.Printf("producer/broker/%d starting up\n", 1)
			logger.AssertInfo("producer/broker/1 starting up")
		})

		It("Logger Print logs the operands at info", func() {
			sarama.Logger.Print("consumer/broker/", 1, " closed dead subscription to ", "topic", "\n")
			logger.AssertInfo("consumer/broker/1 closed dead subscription to topic")
		})

		It("Logger Printf logs the formatted message at info", func() {
			sarama.Logger.Printf("client/metadata fetching metadata for all topics from broker %s\n", "localhost:9092")
			logger.AssertInfo("client/metadata fetching metadata for all topics from broker localhost:9092")
		})

		It("Logger Println logs the operands separated by spaces at info", func() {
			sarama.Logger.Println("client/metadata", "retrying after", 250, "ms")
			logger.AssertInfo("client/metadata retrying after 250 ms")
		})

		It("DebugLogger Printf logs the formatted message at debug", func() {
			sarama.DebugLogger.Printf("producer/txnmgr [%s] committing transaction\n", "transactional-id")
			logger.AssertDebug("producer/txnmgr [transactional-id] committing transaction")
		})
	})
})
