package json_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logJson "github.com/tidepool-org/platform/log/json"
)

var _ = Describe("Logger", func() {
	Context("NewLogger", func() {
		It("returns an error if writer is missing", func() {
			logger, err := logJson.NewLogger(nil, log.DefaultLevelRanks(), log.DefaultLevel())
			Expect(err).To(MatchError("writer is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns an error if level ranks is missing", func() {
			logger, err := logJson.NewLogger(ioutil.Discard, nil, log.DefaultLevel())
			Expect(err).To(MatchError("level ranks is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(logJson.NewLogger(ioutil.Discard, log.DefaultLevelRanks(), log.DefaultLevel())).ToNot(BeNil())
		})
	})
})
