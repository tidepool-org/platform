package json_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/json"
)

var _ = Describe("Logger", func() {
	Context("NewLogger", func() {
		It("returns an error if writer is missing", func() {
			logger, err := json.NewLogger(nil, log.DefaultLevelRanks(), log.DefaultLevel())
			Expect(err).To(MatchError("json: writer is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns an error if level ranks is missing", func() {
			logger, err := json.NewLogger(ioutil.Discard, nil, log.DefaultLevel())
			Expect(err).To(MatchError("log: level ranks is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(json.NewLogger(ioutil.Discard, log.DefaultLevelRanks(), log.DefaultLevel())).ToNot(BeNil())
		})
	})
})
