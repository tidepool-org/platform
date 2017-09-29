package deduplicator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/id"
)

var _ = Describe("Hash", func() {
	Context("AssignDatasetDataIdentityHashes", func() {
		var testDataData []*testData.Datum
		var testDatasetData []data.Datum

		BeforeEach(func() {
			testDataData = []*testData.Datum{}
			testDatasetData = []data.Datum{}
			for i := 0; i < 3; i++ {
				testDatum := testData.NewDatum()
				testDataData = append(testDataData, testDatum)
				testDatasetData = append(testDatasetData, testDatum)
			}
		})

		AfterEach(func() {
			for _, testDataDatum := range testDataData {
				testDataDatum.Expectations()
			}
		})

		It("returns successfully if the data is nil", func() {
			Expect(deduplicator.AssignDatasetDataIdentityHashes(nil)).To(BeNil())
		})

		It("returns successfully if there is no data", func() {
			Expect(deduplicator.AssignDatasetDataIdentityHashes([]data.Datum{})).To(BeNil())
		})

		It("returns an error if any datum returns an error getting identity fields", func() {
			testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
			testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: errors.New("test error")}}
			hashes, err := deduplicator.AssignDatasetDataIdentityHashes(testDatasetData)
			Expect(err).To(MatchError("unable to gather identity fields for datum; test error"))
			Expect(hashes).To(BeNil())
		})

		It("returns an error if any datum returns no identity fields", func() {
			testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
			testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: nil}}
			hashes, err := deduplicator.AssignDatasetDataIdentityHashes(testDatasetData)
			Expect(err).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
			Expect(hashes).To(BeNil())
		})

		It("returns an error if any datum returns empty identity fields", func() {
			testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
			testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{}, Error: nil}}
			hashes, err := deduplicator.AssignDatasetDataIdentityHashes(testDatasetData)
			Expect(err).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
			Expect(hashes).To(BeNil())
		})

		It("returns an error if any datum returns any empty identity fields", func() {
			testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), id.New()}, Error: nil}}
			testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{id.New(), ""}, Error: nil}}
			hashes, err := deduplicator.AssignDatasetDataIdentityHashes(testDatasetData)
			Expect(err).To(MatchError("unable to generate identity hash for datum; identity field is empty"))
			Expect(hashes).To(BeNil())
		})

		Context("with identity fields", func() {
			BeforeEach(func() {
				testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{"test", "0"}, Error: nil}}
				testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{"test", "1"}, Error: nil}}
				testDataData[2].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{"test", "2"}, Error: nil}}
			})

			AfterEach(func() {
				Expect(testDataData[0].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: "GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg="}))
				Expect(testDataData[1].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: "+cywqM0rcj9REPt87Vfx2U+j9m57cB0XW2kmNZm5Ao8="}))
				Expect(testDataData[2].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: "dCPMoOxFVMbPvMkXMbyKeff8QmdBPu8hr/BVeHJhz78="}))
			})

			It("returns successfully", func() {
				hashes, err := deduplicator.AssignDatasetDataIdentityHashes(testDatasetData)
				Expect(err).ToNot(HaveOccurred())
				Expect(hashes).To(Equal([]string{
					"GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg=",
					"+cywqM0rcj9REPt87Vfx2U+j9m57cB0XW2kmNZm5Ao8=",
					"dCPMoOxFVMbPvMkXMbyKeff8QmdBPu8hr/BVeHJhz78=",
				}))
			})
		})
	})

	Context("GenerateIdentityHash", func() {
		It("returns an error if identity fields is missing", func() {
			hash, err := deduplicator.GenerateIdentityHash(nil)
			Expect(err).To(MatchError("identity fields are missing"))
			Expect(hash).To(Equal(""))
		})

		It("returns an error if identity fields is empty", func() {
			hash, err := deduplicator.GenerateIdentityHash([]string{})
			Expect(err).To(MatchError("identity fields are missing"))
			Expect(hash).To(Equal(""))
		})

		It("returns an error if an identity fields empty", func() {
			hash, err := deduplicator.GenerateIdentityHash([]string{"alpha", "", "charlie"})
			Expect(err).To(MatchError("identity field is empty"))
			Expect(hash).To(Equal(""))
		})

		It("successfully returns a hash with one identity field", func() {
			hash, err := deduplicator.GenerateIdentityHash([]string{"zero"})
			Expect(err).ToNot(HaveOccurred())
			Expect(hash).To(Equal("+RlOc/npRZ40UOoQoXnN93qvppW+7NO5NEqY0RFiIkM="))
		})

		It("successfully returns a hash with three identity fields", func() {
			hash, err := deduplicator.GenerateIdentityHash([]string{"alpha", "bravo", "charlie"})
			Expect(err).ToNot(HaveOccurred())
			Expect(hash).To(Equal("dO3wei6LXqnM+oEql2hguPTmyM0+QnmIZPyxzlvL2xY="))
		})

		It("successfully returns a hash with five identity fields", func() {
			hash, err := deduplicator.GenerateIdentityHash([]string{"one", "two", "three", "four", "five"})
			Expect(err).ToNot(HaveOccurred())
			Expect(hash).To(Equal("8HUIFZUXmOuySpngHvl+fJECoeELTiCRxwNxxgDzmVQ="))
		})
	})
})
