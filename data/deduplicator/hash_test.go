package deduplicator_test

import (
	"errors"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	testDataStore "github.com/tidepool-org/platform/data/store/test"
	testData "github.com/tidepool-org/platform/data/test"
	"github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/log"
)

var _ = Describe("Hash", func() {
	Context("NewFactory", func() {
		It("returns a new factory", func() {
			Expect(deduplicator.NewHashFactory()).ToNot(BeNil())
		})
	})

	Context("with a new factory", func() {
		var testFactory deduplicator.Factory
		var userID string
		var testDataset *upload.Upload

		BeforeEach(func() {
			var err error
			testFactory, err = deduplicator.NewHashFactory()
			Expect(err).ToNot(HaveOccurred())
			Expect(testFactory).ToNot(BeNil())
			userID = app.NewID()
			testDataset = upload.Init()
			Expect(testDataset).ToNot(BeNil())
			testDataset.UserID = userID
			testDataset.GroupID = app.NewID()
			testDataset.DeviceID = app.StringAsPointer(app.NewID())
			testDataset.DeviceManufacturers = app.StringArrayAsPointer([]string{"Medtronic"})
		})

		Context("CanDeduplicateDataset", func() {
			It("returns an error if the dataset is missing", func() {
				can, err := testFactory.CanDeduplicateDataset(nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(can).To(BeFalse())
			})

			Context("with deduplicator", func() {
				BeforeEach(func() {
					testDataset.Deduplicator = &data.DeduplicatorDescriptor{Name: "hash"}
				})

				It("returns false if the deduplicator name is empty", func() {
					testDataset.Deduplicator.Name = ""
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
				})

				It("returns false if the deduplicator name does not match", func() {
					testDataset.Deduplicator.Name = "not-hash"
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
				})

				It("returns true if the deduplicator name does match", func() {
					Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
				})
			})

			It("returns false if the dataset id is missing", func() {
				testDataset.UploadID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the user id is missing", func() {
				testDataset.UserID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns false if the group id is missing", func() {
				testDataset.GroupID = ""
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeFalse())
			})

			It("returns true if dataset id, user id, and group id are specified", func() {
				Expect(testFactory.CanDeduplicateDataset(testDataset)).To(BeTrue())
			})
		})

		Context("NewDeduplicator", func() {
			It("returns an error if the logger is missing", func() {
				testDeduplicator, err := testFactory.NewDeduplicator(nil, testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: logger is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the data store session is missing", func() {
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), nil, testDataset)
				Expect(err).To(MatchError("deduplicator: data store session is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset is missing", func() {
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), nil)
				Expect(err).To(MatchError("deduplicator: dataset is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset id is missing", func() {
				testDataset.UploadID = ""
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset user id is missing", func() {
				testDataset.UserID = ""
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset user id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns an error if the dataset group id is missing", func() {
				testDataset.GroupID = ""
				testDeduplicator, err := testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)
				Expect(err).To(MatchError("deduplicator: dataset group id is missing"))
				Expect(testDeduplicator).To(BeNil())
			})

			It("returns a new deduplicator upon success", func() {
				Expect(testFactory.NewDeduplicator(log.NewNull(), testDataStore.NewSession(), testDataset)).ToNot(BeNil())
			})
		})

		Context("with a new deduplicator", func() {
			var testDataStoreSession *testDataStore.Session
			var testDeduplicator data.Deduplicator

			BeforeEach(func() {
				var err error
				testDataStoreSession = testDataStore.NewSession()
				testDeduplicator, err = testFactory.NewDeduplicator(log.NewNull(), testDataStoreSession, testDataset)
				Expect(err).ToNot(HaveOccurred())
				Expect(testDeduplicator).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(testDataStoreSession.UnusedOutputsCount()).To(Equal(0))
			})

			Context("InitializeDataset", func() {
				It("returns an error if there is an error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{errors.New("test error")}
					err := testDeduplicator.InitializeDataset()
					Expect(err).To(MatchError("deduplicator: unable to initialize dataset; test error"))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(testDeduplicator.InitializeDataset()).To(Succeed())
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})

				It("sets the dataset deduplicator if there is no error", func() {
					testDataStoreSession.UpdateDatasetOutputs = []error{nil}
					Expect(testDeduplicator.InitializeDataset()).To(Succeed())
					Expect(testDataset.DeduplicatorDescriptor()).To(Equal(&data.DeduplicatorDescriptor{Name: "hash"}))
					Expect(testDataStoreSession.UpdateDatasetInputs).To(ConsistOf(testDataset))
				})
			})

			Context("AddDataToDataset", func() {
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
						Expect(testDataDatum.UnusedOutputsCount()).To(Equal(0))
					}
				})

				It("returns an error if the dataset is missing", func() {
					err := testDeduplicator.AddDataToDataset(nil)
					Expect(err).To(MatchError("deduplicator: dataset data is missing"))
				})

				It("returns successfully if there is no data", func() {
					Expect(testDeduplicator.AddDataToDataset([]data.Datum{})).To(Succeed())
				})

				It("returns an error if any datum returns an error getting identity fields", func() {
					testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{app.NewID(), app.NewID()}, Error: nil}}
					testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: errors.New("test error")}}
					err := testDeduplicator.AddDataToDataset(testDatasetData)
					Expect(err).To(MatchError("deduplicator: unable to gather identity fields for datum; test error"))
				})

				It("returns an error if any datum returns no identity fields", func() {
					testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{app.NewID(), app.NewID()}, Error: nil}}
					testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: nil, Error: nil}}
					err := testDeduplicator.AddDataToDataset(testDatasetData)
					Expect(err).To(MatchError("deduplicator: unable to generate identity hash for datum; deduplicator: identity fields are missing"))
				})

				It("returns an error if any datum returns empty identity fields", func() {
					testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{app.NewID(), app.NewID()}, Error: nil}}
					testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{}, Error: nil}}
					err := testDeduplicator.AddDataToDataset(testDatasetData)
					Expect(err).To(MatchError("deduplicator: unable to generate identity hash for datum; deduplicator: identity fields are missing"))
				})

				It("returns an error if any datum returns any empty identity fields", func() {
					testDataData[0].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{app.NewID(), app.NewID()}, Error: nil}}
					testDataData[1].IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{app.NewID(), ""}, Error: nil}}
					err := testDeduplicator.AddDataToDataset(testDatasetData)
					Expect(err).To(MatchError("deduplicator: unable to generate identity hash for datum; deduplicator: identity field is empty"))
				})

				Context("with identity fields", func() {
					BeforeEach(func() {
						for index, testDataDatum := range testDataData {
							testDataDatum.IdentityFieldsOutputs = []testData.IdentityFieldsOutput{{IdentityFields: []string{"test", strconv.Itoa(index)}, Error: nil}}
						}
						testDataStoreSession.FindDatasetDataDeduplicatorHashesOutputs = []testDataStore.FindDatasetDataDeduplicatorHashesOutput{{
							Hashes: []string{
								"GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg=",
							},
							Error: nil,
						}}
					})

					AfterEach(func() {
						for _, testDataDatum := range testDataData {
							Expect(testDataDatum.SetDeduplicatorDescriptorInputs).To(HaveLen(1))
							Expect(testDataDatum.SetDeduplicatorDescriptorInputs[0].Name).To(Equal("hash"))
							Expect(testDataDatum.SetDeduplicatorDescriptorInputs[0].Hash).ToNot(BeEmpty())
						}
						Expect(testDataStoreSession.FindDatasetDataDeduplicatorHashesInputs).To(Equal([]testDataStore.FindDatasetDataDeduplicatorHashesInput{{
							UserID: userID,
							Hashes: []string{
								"GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg=",
								"+cywqM0rcj9REPt87Vfx2U+j9m57cB0XW2kmNZm5Ao8=",
								"dCPMoOxFVMbPvMkXMbyKeff8QmdBPu8hr/BVeHJhz78=",
							},
						}}))
					})

					It("returns an error if finding deduplicator hashes returns an error", func() {
						testDataStoreSession.FindDatasetDataDeduplicatorHashesOutputs = []testDataStore.FindDatasetDataDeduplicatorHashesOutput{{Hashes: nil, Error: errors.New("test error")}}
						err := testDeduplicator.AddDataToDataset(testDatasetData)
						Expect(err).To(MatchError("deduplicator: unable to find existing identity hashes; test error"))
					})

					Context("with deduplicator descriptor", func() {
						BeforeEach(func() {
							testDataData[0].DeduplicatorDescriptorOutputs = []*data.DeduplicatorDescriptor{{
								Name: "hash",
								Hash: "GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg=",
							}}
							testDataData[1].DeduplicatorDescriptorOutputs = []*data.DeduplicatorDescriptor{{
								Name: "hash",
								Hash: "+cywqM0rcj9REPt87Vfx2U+j9m57cB0XW2kmNZm5Ao8=",
							}}
							testDataData[2].DeduplicatorDescriptorOutputs = []*data.DeduplicatorDescriptor{{
								Name: "hash",
								Hash: "dCPMoOxFVMbPvMkXMbyKeff8QmdBPu8hr/BVeHJhz78=",
							}}
						})

						It("returns success if finding deduplicator hashes returns all hashes", func() {
							testDataStoreSession.FindDatasetDataDeduplicatorHashesOutputs = []testDataStore.FindDatasetDataDeduplicatorHashesOutput{{
								Hashes: []string{
									"GRp47M02cMlAzSn7oJTQ2LC9eb1Qd6mIPO1U8GeuoYg=",
									"+cywqM0rcj9REPt87Vfx2U+j9m57cB0XW2kmNZm5Ao8=",
									"dCPMoOxFVMbPvMkXMbyKeff8QmdBPu8hr/BVeHJhz78=",
								},
								Error: nil,
							}}
							Expect(testDeduplicator.AddDataToDataset(testDatasetData)).To(Succeed())
						})

						Context("with creating dataset data", func() {
							AfterEach(func() {
								Expect(testDataStoreSession.CreateDatasetDataInputs).To(ConsistOf(testDataStore.CreateDatasetDataInput{
									Dataset: testDataset,
									DatasetData: []data.Datum{
										testDataData[1],
										testDataData[2],
									},
								}))
							})

							It("returns an error if there is an error with CreateDatasetDataInput", func() {
								testDataStoreSession.CreateDatasetDataOutputs = []error{errors.New("test error")}
								err := testDeduplicator.AddDataToDataset(testDatasetData)
								Expect(err).To(MatchError("deduplicator: unable to add data to dataset; test error"))
							})

							It("returns successfully if there is no error", func() {
								testDataStoreSession.CreateDatasetDataOutputs = []error{nil}
								Expect(testDeduplicator.AddDataToDataset(testDatasetData)).To(Succeed())
							})
						})
					})
				})
			})

			Context("FinalizeDataset", func() {
				It("returns an error if there is an error on activate", func() {
					uploadID := app.NewID()
					testDataset.UploadID = uploadID
					testDataStoreSession.ActivateDatasetDataOutputs = []error{errors.New("test error")}
					err := testDeduplicator.FinalizeDataset()
					Expect(err).To(MatchError(`deduplicator: unable to activate data in dataset with id "` + uploadID + `"; test error`))
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
				})

				It("returns successfully if there is no error", func() {
					testDataStoreSession.ActivateDatasetDataOutputs = []error{nil}
					Expect(testDeduplicator.FinalizeDataset()).To(Succeed())
					Expect(testDataStoreSession.ActivateDatasetDataInputs).To(ConsistOf(testDataset))
				})
			})
		})
	})
})
