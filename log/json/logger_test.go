package json_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/json"
)

var _ = Describe("Logger", func() {
	Context("NewLogger", func() {
		It("returns an error if writer is missing", func() {
			logger, err := json.NewLogger(nil, log.DefaultLevels(), log.DefaultLevel())
			Expect(err).To(MatchError("json: writer is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns an error if levels is missing", func() {
			logger, err := json.NewLogger(os.Stdout, nil, log.DefaultLevel())
			Expect(err).To(MatchError("log: levels is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(json.NewLogger(os.Stdout, log.DefaultLevels(), log.DefaultLevel())).ToNot(BeNil())
		})
	})
})
