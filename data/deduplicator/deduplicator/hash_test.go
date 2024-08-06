package deduplicator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	dataDeduplicatorDeduplicator "github.com/tidepool-org/platform/data/deduplicator/deduplicator"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"

	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataTest "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/blood/glucose/selfmonitored"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
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
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(nil, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent)).To(Succeed())
		})

		It("returns successfully when there is no data", func() {
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(nil, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(Succeed())
		})

		It("returns successfully when there is no data", func() {
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(data.Data{}, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent)).To(Succeed())
		})

		It("returns successfully when there is no data", func() {
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(data.Data{}, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(Succeed())
		})

		It("returns an error when any datum returns an error getting identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: nil, Error: errors.New("test error")}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent)).To(MatchError("unable to gather identity fields for datum; test error"))
		})

		It("returns an error when any datum returns an error getting legacy identity fields", func() {
			dataSetDataTest[0].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[0].GetTypeOutputs = []string{"test-type"}
			dataSetDataTest[1].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: nil, Error: errors.New("test error")}}
			dataSetDataTest[1].GetTypeOutputs = []string{"test-type"}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(MatchError("unable to gather legacy identity fields for datum *test.Datum; test error"))
		})

		It("returns an error when any datum returns no identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: nil, Error: nil}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent)).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
		})

		It("returns an error when any datum returns no legacy identity fields", func() {
			dataSetDataTest[0].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[0].GetTypeOutputs = []string{"test-type"}
			dataSetDataTest[1].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: nil, Error: nil}}
			dataSetDataTest[1].GetTypeOutputs = []string{"test-type"}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(MatchError("unable to generate legacy identity hash for datum *test.Datum; identity fields are missing"))
		})

		It("returns an error when any datum returns empty identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{}, Error: nil}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent)).To(MatchError("unable to generate identity hash for datum; identity fields are missing"))
		})

		It("returns an error when any datum returns empty legacy identity fields", func() {
			dataSetDataTest[0].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[0].GetTypeOutputs = []string{"test-type"}
			dataSetDataTest[1].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{}, Error: nil}}
			dataSetDataTest[1].GetTypeOutputs = []string{"test-type"}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(MatchError("unable to generate legacy identity hash for datum *test.Datum; identity fields are missing"))
		})

		It("returns an error when any datum returns any empty identity fields", func() {
			dataSetDataTest[0].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[1].IdentityFieldsOutputs = []dataTest.IdentityFieldsOutput{{IdentityFields: []string{userTest.RandomID(), ""}, Error: nil}}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent)).To(MatchError("unable to generate identity hash for datum; identity field is empty"))
		})

		It("returns an error when any datum returns any empty legacy identity fields", func() {
			dataSetDataTest[0].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{userTest.RandomID(), dataTest.NewDeviceID()}, Error: nil}}
			dataSetDataTest[0].GetTypeOutputs = []string{"test-type"}
			dataSetDataTest[1].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{userTest.RandomID(), ""}, Error: nil}}
			dataSetDataTest[1].GetTypeOutputs = []string{"test-type"}
			Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(MatchError("unable to generate legacy identity hash for datum *test.Datum; identity field is empty"))
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
				Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionCurrent)).To(Succeed())
			})
		})

		Context("with legacy identity fields", func() {
			BeforeEach(func() {
				dataSetDataTest[0].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{"test", "0"}, Error: nil}}
				dataSetDataTest[0].GetTypeOutputs = []string{"test-type"}
				dataSetDataTest[1].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{"test", "1"}, Error: nil}}
				dataSetDataTest[1].GetTypeOutputs = []string{"test-type"}
				dataSetDataTest[2].LegacyIdentityFieldsOutputs = []dataTest.LegacyIdentityFieldsOutput{{LegacyIdentityFields: []string{"test", "2"}, Error: nil}}
				dataSetDataTest[2].GetTypeOutputs = []string{"test-type"}
			})

			AfterEach(func() {
				Expect(dataSetDataTest[0].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: pointer.FromString("1i2gupee95mf7ooatr0cveo9qr6dt2i6")}))
				Expect(dataSetDataTest[1].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: pointer.FromString("f0mq6bk3d79aua8lre8ti8bosqkaf10d")}))
				Expect(dataSetDataTest[2].DeduplicatorDescriptorValue).To(Equal(&data.DeduplicatorDescriptor{Hash: pointer.FromString("sknnju4it96cqsbcmuh3g1navg1kdvp5")}))
			})

			It("returns successfully", func() {
				Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(dataSetData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(Succeed())
			})
		})

		Context("with legacy identity fields", func() {
			var smbgData data.Data
			BeforeEach(func() {
				var newSMBG = func() data.Datum {
					datum := selfmonitored.New()
					datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(pointer.FromString("mg/dl"))
					datum.Type = "smbg"
					datum.Value = pointer.FromFloat64(150)
					datum.SubType = pointer.FromString(test.RandomStringFromArray(selfmonitored.SubTypes()))

					normalizer := dataNormalizer.New()
					Expect(normalizer).ToNot(BeNil())
					datum.Normalize(normalizer.WithOrigin(structure.OriginExternal))
					return datum
				}

				for i := 0; i < 10; i++ {
					smbgData = append(smbgData, newSMBG())
				}

			})

			It("returns successfully", func() {
				Expect(dataDeduplicatorDeduplicator.AssignDataSetDataIdentityHashes(smbgData, dataDeduplicatorDeduplicator.DeviceDeactivateHashVersionLegacy)).To(Succeed())
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
	Context("GenerateLegacyIdentityHash", func() {
		It("returns an error when identity fields is missing", func() {
			hash, err := dataDeduplicatorDeduplicator.GenerateLegacyIdentityHash(nil, false)
			Expect(err).To(MatchError("identity fields are missing"))
			Expect(hash).To(BeEmpty())
		})

		It("returns an error when identity fields is empty", func() {
			hash, err := dataDeduplicatorDeduplicator.GenerateLegacyIdentityHash([]string{}, false)
			Expect(err).To(MatchError("identity fields are missing"))
			Expect(hash).To(BeEmpty())
		})

		It("returns an error when an identity fields empty", func() {
			hash, err := dataDeduplicatorDeduplicator.GenerateLegacyIdentityHash([]string{"alpha", "", "charlie"}, false)
			Expect(err).To(MatchError("identity field is empty"))
			Expect(hash).To(BeEmpty())
		})

		DescribeTable("hash from legacy identity tests",
			func(fields []string, expectedHash string, expectedErr error) {
				actualHash, actualErr := dataDeduplicatorDeduplicator.GenerateLegacyIdentityHash(fields, true)
				if expectedErr != nil {
					Expect(actualErr).To(Equal(expectedErr))
				} else {
					Expect(actualHash).To(Equal(expectedHash))
					Expect(actualErr).To(BeNil())
				}
			},
			// Entry("smbg id", []string{"smbg", "tools", "2014-06-11T11:12:43.029Z", "5.550747991045533"}, "e2ihon9nqcro96c4uugb4ftdnr07nqok", nil),
			// Entry("smbg id", []string{"smbg", "tools", "2014-06-11T17:57:01.703Z", "4.5"}, "c14eds071pp5gsirfmgmsclbcahs8th0", nil),
			// Entry("smbg id", []string{"smbg", "tools", "2015-07-04T10:13:00.000Z", "4.9"}, "rk2htms97m7hipdu5lrso7ufd3pedm6n", nil),
			// Entry("smbg id", []string{"smbg", "tools", "2015-07-04T10:13:00.000Z", "4.8"}, "urrkdln86rl4vhqckps6gnupg5njqk6n", nil),

			Entry("smbg id", []string{"smbg", "tandemCIQ1002717664049", "2023-05-13T15:51:58.000Z", "12.211645580300173"}, "o5d5o0q9g1s70prg609osop8s4sg182f", nil),
			Entry("smbg id", []string{"smbg", "tandemCIQ1002717664049", "2023-05-12T15:21:16.000Z", "7.382494828090559"}, "pr2b3dii5hddtlo3i973t8rat2blm3n0", nil),
			Entry("smbg id", []string{"smbg", "tandemCIQ1002717664049", "2023-03-27T15:38:59.000Z", "7.49350978791147"}, "dotp56ahq7m8ehesfep0rimmo6rooaht", nil),
			Entry("smbg id", []string{"smbg", "tandemCIQ1002717664049", "2023-02-06T07:38:17.000Z", "15.153542015554306"}, "bo06rjosvq671bi55fj6gbu95nu2lrrt", nil),
			Entry("smbg id", []string{"smbg", "tandemCIQ1002717664049", "2023-02-06T07:16:27.000Z", "13.821362497703378"}, "nj1elbn2ql6qpeecbn1r5c7qoqvvc1oj", nil),

			// Entry("cbg id", []string{"cbg", "tools", "2014-06-11T11:12:43.029Z"}, "eb12p6h892pmd0hhccpt2r17muc407o0", nil),
			// Entry("cbg id", []string{"cbg", "tools", "2014-06-11T17:57:01.703Z"}, "ha2ogn1kenqqhseed504sqnanhnclg5s", nil),
			// Entry("cbg id", []string{"cbg", "tools", "2014-06-12T11:12:43.029Z"}, "i922lobl3kron3t81pjap31anopkspvb", nil),
			// Entry("cbg id", []string{"cbg", "DexHealthKit_Dexcom:com.dexcom.Share2:3.0.4.17", "2015-12-21T11:23:08Z"}, "nsikjhfaprplpq78hc7di2lu5qpt1e3k", nil),

			// Entry("basal id", []string{"basal", "scheduled", "tools", "2014-06-11T00:00:00.000Z"}, "kmm427pfbrc6rugtmbuli8j4q61u17uk", nil),
			// Entry("basal id", []string{"basal", "scheduled", "tools", "2014-06-11T06:00:00.000Z"}, "cjou7vscvp8ogv34d6vejootulqfn3jd", nil),
			// Entry("basal id", []string{"basal", "temp", "tools", "2014-06-11T09:00:00.000Z"}, "tn33bjb0241j9qh4jg9vdnf1g6k1g9r8", nil),
			// Entry("basal id", []string{"basal", "scheduled", "tools", "2014-06-11T19:00:00.000Z"}, "kftn188l8rjuvma3qkd3iqg34t0plajp", nil),
		)
	})
})
