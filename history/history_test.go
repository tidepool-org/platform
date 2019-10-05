package history_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/tidepool-org/platform/history"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"

	dataTypesTest "github.com/tidepool-org/platform/data/types/test"
	"github.com/tidepool-org/platform/structure"
)

type HistoryTest struct {
	history.History

	ref []byte
}

func (h *HistoryTest) Validate(validator structure.Validator) {
	h.History.Validate(validator, h.ref)
}

func RandomHistory() *history.History {
	datum := history.New()
	datum.Time = pointer.FromString(test.RandomTime().Format(time.RFC3339Nano))
	datum.Changes = RandomJSONPatchArray()
	return datum
}

func RandomHistoryTest() *HistoryTest {
	datum := RandomHistory()
	historyTest := HistoryTest{*datum, PatchObjects()[0]}
	return &historyTest
}

var _ = Describe("History", func() {
	Context("History", func() {
		Context("Parse", func() {
			// TODO
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *HistoryTest), expectedErrors ...error) {
					datum := RandomHistoryTest()
					mutator(datum)
					dataTypesTest.ValidateWithExpectedOrigins(datum, structure.Origins(), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *HistoryTest) {},
				),
			)

		})
	})
})
