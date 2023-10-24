package v1_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
	datum.Active = true
	Expect(datum.GetType()).To(Equal(typ))
	return &datum
}

func NewOldDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Active = true
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, -24, -1))
	return datum
}

func NewNewDatum(typ string) *baseDatum.Base {
	datum := NewDatum(typ)
	datum.Active = true
	datum.Time = pointer.FromAny(time.Now().UTC().AddDate(0, 0, 2))
	return datum
}

var _ = Describe("DataSetsDataCreate", func() {
	Context("CheckDatumUpdatesSummary", func() {
		It("with non-summary type", func() {
			var updatesSummary map[string]struct{}
			datum := NewDatum(food.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(BeEmpty())
		})

		It("with too old summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewOldDatum(continuous.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with future summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewNewDatum(continuous.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with CGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(continuous.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(1))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeCGM))
		})

		It("with BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(selfmonitored.Type)

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(1))
			Expect(updatesSummary).To(HaveKey(types.SummaryTypeBGM))
		})

		It("with inactive BGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(selfmonitored.Type)
			datum.Active = false

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})

		It("with inactive CGM summary affecting record", func() {
			updatesSummary := make(map[string]struct{})
			datum := NewDatum(continuous.Type)
			datum.Active = false

			v1.CheckDatumUpdatesSummary(updatesSummary, datum)
			Expect(updatesSummary).To(HaveLen(0))
		})
	})
})
