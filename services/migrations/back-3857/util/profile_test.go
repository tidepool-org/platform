package util_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/services/migrations/back-3857/util"
)

var _ = Describe("Consent", func() {

	BeforeEach(func() {
	})

	Context("PopulateCreateFromSeagullDocumentValue", func() {
		var value string
		var create *consent.RecordCreate

		BeforeEach(func() {
			value = `{"profile":{"fullName":"Jill Jellyfish","patient":{"birthday":"2011-08-12","diagnosisDate":"2025-08-01","diagnosisType":"type1","isOtherPerson":true,"fullName":"James Jellyfish"}}}`
			create = &consent.RecordCreate{}
			create.GrantTime = time.Date(2025, 8, 29, 0, 0, 0, 0, time.UTC)
		})

		It("Should populate the create record correctly", func() {
			util.PopulateCreateFromSeagullDocumentValue(value, create, null.NewLogger())
			Expect(create.OwnerName).To(Equal("James Jellyfish"))
			Expect(create.ParentGuardianName).To(PointTo(Equal("Jill Jellyfish")))
			Expect(create.AgeGroup).To(Equal(consent.AgeGroupThirteenSeventeen))
		})
	})
})
