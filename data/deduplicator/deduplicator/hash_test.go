package deduplicator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Hash", func() {
	Context("AssignDataSetDataIdentityHashes", func() {
		var dataSetDataTest []*dataTest.Datum
		var dataSetData data.Data

		BeforeEach(func() {
			dataSetDataTest = []*dataTest.Datum{}
			dataSetData = data.Data{}
			for i := 0; i < 3; i++ {
				datum := dataTest.NewDatum()
				dataSetDataTest = append(dataSetDataTest, datum)
				dataSetData = append(dataSetData, datum)
			}
		})

		AfterEach(func() {
			for _, datum := range dataSetDataTest {
				datum.Expectations()
			}
		})

		It("returns successfully when the data is nil", func() {
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(nil, "")).To(Succeed())
		})

		It("returns successfully when there is no data", func() {
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(data.Data{}, "")).To(Succeed())
		})

		It("returns an error when any datum returns an error getting identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomUserID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: nil, Error: errors.New("test error")}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, "")).To(MatchError("unable to gather identity fields for datum; test error"))
		})

		It("returns an error when any datum returns no identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomUserID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: nil, Error: nil}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, "")).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
		})

		It("returns an error when any datum returns empty identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomUserID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{}, Error: nil}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, "")).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
		})

		It("returns an error when any datum returns any empty identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomUserID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomUserID(), ""}, Error: nil}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, "")).To(MatchError("unable to generate identity hash for datum; identity field is empty"))
		})

		Context("with identity fields", func() {
			BeforeEach(func() {
				dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{"test", "0"}, Error: nil}}
				dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{"test", "1"}, Error: nil}}
				dataSetDataTest[2].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{"test", "2"}, Error: nil}}
			})

			AfterEach(func() {
				Expect(dataSetDataTest[0].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: pointer.FromString("GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg=")}))
				Expect(dataSetDataTest[1].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: pointer.FromString("+cywqM0rcj9REPt87Vfx2U+j9m57cB0XW2kmNZm5Ao8=")}))
				Expect(dataSetDataTest[2].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: pointer.FromString("dCPMoOxFVMbPvMkXMbyKeff8QmdBPu8hr/BVeHJhz78=")}))
			})

			It("returns successfully", func() {
				Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, "")).To(Succeed())
			})
		})
	})

	Context("GenerateIdentityHash", func() {
		It("returns an error when identity fields is missing", func() {
			hash, err := dataDeduplicatorDeduplicator.GenerateIdentityHash(nil)
			Expect(err).To(MatchError("identity fields are missing"))
			Expect(hash).To(BeEmpty())
		})

		It("returns an error when identity fields is empty", func() {
			hash, err := dataDeduplicatorDeduplicator.GenerateIdentityHash([]string{})
			Expect(err).To(MatchError("identity fields are missing"))
			Expect(hash).To(BeEmpty())
		})

		It("returns an error when an identity fields empty", func() {
			hash, err := dataDeduplicatorDeduplicator.GenerateIdentityHash([]string{"alpha", "", "charlie"})
			Expect(err).To(MatchError("identity field is empty"))
			Expect(hash).To(BeEmpty())
		})

		It("successfully returns a hash with one identity field", func() {
			Expect(dataDeduplicatorDeduplicator.GenerateIdentityHash([]string{"zero"})).To(Equal("+RlOc/npRZ40UOoQoXnN93qvppW+7NO5NEqY0RFiIkM="))
		})

		It("successfully returns a hash with three identity fields", func() {
			Expect(dataDeduplicatorDeduplicator.GenerateIdentityHash([]string{"alpha", "bravo", "charlie"})).To(Equal("dO3wei6LXqnM+oEql2hguPTmyM0+QnmIZPyxzlvL2xY="))
		})

		It("successfully returns a hash with five identity fields", func() {
			Expect(dataDeduplicatorDeduplicator.GenerateIdentityHash([]string{"one", "two", "three", "four", "five"})).To(Equal("8HUIFZUXmOuySpngHvl+fJECoeELTiCRxwNxxgDzmVQ="))
		})
	})
})
