package v1_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	v1 "github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/data/summary/types"
	baseDatum "github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/pointer"
)

func NewDatum(typ string) *baseDatum.Base {
	datum := baseDatum.New(typ)
	datum.Time = pointer.FromAny(time.Now().UTC())
	Expect(datum.GetType()).To(Equal(typ))
	return &datum
}

func NewOldDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, -24, -1))
	return datum
}

func NewNewDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, 0, 2))
	return datum
}

var _ = Describe("DataSetsDataCreate", func() {
	Context("CheckDatumUpdatesSummary", func() {
		It("with non-summary type", func() {
			var updatesSummary map[string]struct{}
			var datum data.Datum = NewDatum(food.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(BeEmpty())
		})

		It("with too old summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			var datum data.Datum = NewOldDatum(continuous.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with future summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			var datum data.Datum = NewNewDatum(continuous.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with CGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			var datum data.Datum = NewDatum(continuous.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(1))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeCGM))
		})

		It("with BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			var datum data.Datum = NewDatum(selfmonitored.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(1))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeBGM))
		})
	})
})
