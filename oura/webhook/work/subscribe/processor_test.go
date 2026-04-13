package subscribe_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.uber.org/mock/gomock"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/oura"
	ouraTest "github.com/tidepool-org/platform/oura/test"
	ouraWebhook "github.com/tidepool-org/platform/oura/webhook"
	ouraWebhookWorkSubscribe "github.com/tidepool-org/platform/oura/webhook/work/subscribe"
	ouraWebhookWorkSubscribeTest "github.com/tidepool-org/platform/oura/webhook/work/subscribe/test"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workBase "github.com/tidepool-org/platform/work/base"
	workTest "github.com/tidepool-org/platform/work/test"
)

var _ = Describe("processor", func() {
	It("PendingAvailableDuration is expected", func() {
		Expect(ouraWebhookWorkSubscribe.PendingAvailableDuration).To(Equal(24 * time.Hour))
	})

	It("FailingRetryDuration is expected", func() {
		Expect(ouraWebhookWorkSubscribe.FailingRetryDuration).To(Equal(1 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraWebhookWorkSubscribe.FailingRetryDurationJitter).To(Equal(10 * time.Second))
	})

	It("OverrideDisabled is expected", func() {
		Expect(ouraWebhookWorkSubscribe.OverrideDisabled).To(Equal("disabled"))
	})

	It("OverrideRenew is expected", func() {
		Expect(ouraWebhookWorkSubscribe.OverrideRenew).To(Equal("renew"))
	})

	It("OverrideReset is expected", func() {
		Expect(ouraWebhookWorkSubscribe.OverrideReset).To(Equal("reset"))
	})

	It("OverrideUpdate is expected", func() {
		Expect(ouraWebhookWorkSubscribe.OverrideUpdate).To(Equal("update"))
	})

	Context("Overrides", func() {
		It("returns expected data types", func() {
			Expect(ouraWebhookWorkSubscribe.Overrides()).To(Equal([]string{
				ouraWebhookWorkSubscribe.OverrideDisabled,
				ouraWebhookWorkSubscribe.OverrideRenew,
				ouraWebhookWorkSubscribe.OverrideReset,
				ouraWebhookWorkSubscribe.OverrideUpdate,
			}))
		})
	})

	Context("Metadata", func() {
		Context("MetadataKeyOverride", func() {
			It("returns expected value", func() {
				Expect(ouraWebhookWorkSubscribe.MetadataKeyOverride).To(Equal("override"))
			})
		})

		Context("Metadata", func() {
			DescribeTable("serializes the datum as expected",
				func(mutator func(datum *ouraWebhookWorkSubscribe.Metadata)) {
					datum := ouraWebhookWorkSubscribeTest.RandomMetadata(test.AllowOptional())
					mutator(datum)
					test.ExpectSerializedObjectJSON(datum, ouraWebhookWorkSubscribeTest.NewObjectFromMetadata(datum, test.ObjectFormatJSON))
					test.ExpectSerializedObjectBSON(datum, ouraWebhookWorkSubscribeTest.NewObjectFromMetadata(datum, test.ObjectFormatBSON))
				},
				Entry("succeeds",
					func(datum *ouraWebhookWorkSubscribe.Metadata) {},
				),
				Entry("empty",
					func(datum *ouraWebhookWorkSubscribe.Metadata) {
						*datum = ouraWebhookWorkSubscribe.Metadata{}
					},
				),
				Entry("all",
					func(datum *ouraWebhookWorkSubscribe.Metadata) {
						datum.Override = pointer.From(ouraWebhookWorkSubscribeTest.RandomOverride())
					},
				),
			)

			Context("Parse", func() {
				DescribeTable("parses the datum",
					func(mutator func(object map[string]any, expectedDatum *ouraWebhookWorkSubscribe.Metadata), expectedErrors ...error) {
						expectedDatum := ouraWebhookWorkSubscribeTest.RandomMetadata(test.AllowOptional())
						object := ouraWebhookWorkSubscribeTest.NewObjectFromMetadata(expectedDatum, test.ObjectFormatJSON)
						mutator(object, expectedDatum)
						result := &ouraWebhookWorkSubscribe.Metadata{}
						errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(result), expectedErrors...)
						Expect(result).To(Equal(expectedDatum))
					},
					Entry("succeeds",
						func(object map[string]any, expectedDatum *ouraWebhookWorkSubscribe.Metadata) {},
					),
					Entry("empty",
						func(object map[string]any, expectedDatum *ouraWebhookWorkSubscribe.Metadata) {
							clear(object)
							*expectedDatum = ouraWebhookWorkSubscribe.Metadata{}
						},
					),
					Entry("multiple errors",
						func(object map[string]any, expectedDatum *ouraWebhookWorkSubscribe.Metadata) {
							object["override"] = true
							expectedDatum.Override = nil
						},
						errorsTest.WithPointerSource(structureParser.ErrorTypeNotString(true), "/override"),
					),
				)
			})

			Context("Validate", func() {
				DescribeTable("validates the datum",
					func(mutator func(datum *ouraWebhookWorkSubscribe.Metadata), expectedErrors ...error) {
						datum := ouraWebhookWorkSubscribeTest.RandomMetadata(test.AllowOptional())
						mutator(datum)
						errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
					},
					Entry("succeeds",
						func(datum *ouraWebhookWorkSubscribe.Metadata) {},
					),
					Entry("override missing",
						func(datum *ouraWebhookWorkSubscribe.Metadata) {
							datum.Override = nil
						},
					),
					Entry("override empty",
						func(datum *ouraWebhookWorkSubscribe.Metadata) {
							datum.Override = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", ouraWebhookWorkSubscribe.Overrides()), "/override"),
					),
					Entry("override invalid",
						func(datum *ouraWebhookWorkSubscribe.Metadata) {
							datum.Override = pointer.From("invalid")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("invalid", ouraWebhookWorkSubscribe.Overrides()), "/override"),
					),
					Entry("override valid",
						func(datum *ouraWebhookWorkSubscribe.Metadata) {
							datum.Override = pointer.From(ouraWebhookWorkSubscribeTest.RandomOverride())
						},
					),
					Entry("multiple errors",
						func(datum *ouraWebhookWorkSubscribe.Metadata) {
							datum.Override = pointer.From("")
						},
						errorsTest.WithPointerSource(structureValidator.ErrorValueStringNotOneOf("", ouraWebhookWorkSubscribe.Overrides()), "/override"),
					),
				)
			})
		})
	})

	Context("with dependencies", func() {
		var ctx context.Context
		var mockController *gomock.Controller
		var mockWorkClient *workTest.MockClient
		var mockOuraClient *ouraTest.MockClient
		var dependencies ouraWebhookWorkSubscribe.Dependencies

		BeforeEach(func() {
			ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
			mockController, ctx = gomock.WithContext(ctx, GinkgoT())
			mockWorkClient = workTest.NewMockClient(mockController)
			mockOuraClient = ouraTest.NewMockClient(mockController)
			dependencies = ouraWebhookWorkSubscribe.Dependencies{
				Dependencies: workBase.Dependencies{
					WorkClient: mockWorkClient,
				},
				OuraClient: mockOuraClient,
			}
		})

		Context("NewProcessor", func() {
			It("returns an error if dependencies is invalid", func() {
				dependencies.WorkClient = nil
				processor, err := ouraWebhookWorkSubscribe.NewProcessor(dependencies)
				Expect(err).To(MatchError("dependencies is invalid; work client is missing"))
				Expect(processor).To(BeNil())
			})

			It("returns successfully", func() {
				processor, err := ouraWebhookWorkSubscribe.NewProcessor(dependencies)
				Expect(err).ToNot(HaveOccurred())
				Expect(processor).ToNot(BeNil())
			})

			Context("with processor", func() {
				var wrk *work.Work
				var mockProcessingUpdater *workTest.MockProcessingUpdater
				var processor *ouraWebhookWorkSubscribe.Processor

				BeforeEach(func() {
					wrkCreate, err := ouraWebhookWorkSubscribe.NewWorkCreate()
					Expect(err).ToNot(HaveOccurred())
					Expect(wrkCreate).ToNot(BeNil())
					wrk = workTest.NewWorkFromCreateWithState(wrkCreate, work.StateProcessing)
					mockProcessingUpdater = workTest.NewMockProcessingUpdater(mockController)
					processor, err = ouraWebhookWorkSubscribe.NewProcessor(dependencies)
					Expect(err).ToNot(HaveOccurred())
					Expect(processor).ToNot(BeNil())
				})

				Context("Process", func() {
					It("returns an error if unable to list existing subscriptions", func() {
						testErr := errorsTest.RandomError()
						mockOuraClient.EXPECT().ListSubscriptions(gomock.Not(gomock.Nil())).Return(nil, testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
					})

					Context("with list subscriptions successful", func() {
						var now time.Time
						var expirationTime time.Time
						var partnerURL string
						var partnerSecret string
						var knownSubscriptions oura.Subscriptions
						var unknownSubscriptions oura.Subscriptions
						var resetSubscriptions oura.Subscriptions
						var updateSubscriptions oura.Subscriptions
						var renewSubscriptions oura.Subscriptions
						var createSubscriptions oura.Subscriptions
						var deleteSubscriptions oura.Subscriptions
						var existingSubscriptions oura.Subscriptions
						var expectedMetadata map[string]any

						newSubscription := func(dataType string, eventType string) *oura.Subscription {
							return &oura.Subscription{
								ID:             pointer.From(ouraTest.RandomID()),
								CallbackURL:    pointer.From(partnerURL + ouraWebhook.EventPath),
								DataType:       pointer.From(dataType),
								EventType:      pointer.From(eventType),
								ExpirationTime: pointer.From(test.RandomTimeAfter(expirationTime.Add(time.Minute)).Format(oura.SubscriptionExpirationTimeFormat)),
							}
						}

						newCreateSubscription := func(dataType string, eventType string) *oura.CreateSubscription {
							return &oura.CreateSubscription{
								CallbackURL:       pointer.From(partnerURL + ouraWebhook.EventPath),
								VerificationToken: pointer.From(partnerSecret),
								DataType:          pointer.From(dataType),
								EventType:         pointer.From(eventType),
							}
						}

						newUpdateSubscription := func(dataType string, eventType string) *oura.UpdateSubscription {
							return &oura.UpdateSubscription{
								CallbackURL:       pointer.From(partnerURL + ouraWebhook.EventPath),
								VerificationToken: pointer.From(partnerSecret),
								DataType:          pointer.From(dataType),
								EventType:         pointer.From(eventType),
							}
						}

						BeforeEach(func() {
							now = time.Now().UTC()
							expirationTime = now.Add(ouraWebhookWorkSubscribe.ExpirationTimeDurationMinimum)
							partnerURL = test.RandomString()
							partnerSecret = test.RandomString()
							knownSubscriptions = oura.Subscriptions{}
							for _, dataType := range ouraWebhook.DataTypes() {
								for _, eventType := range oura.EventTypes() {
									knownSubscriptions = append(knownSubscriptions, newSubscription(dataType, eventType))
								}
							}
							unknownSubscriptions = oura.Subscriptions{newSubscription("unknown-1", "unknown-1"), newSubscription("unknown-2", "unknown-2")}
							resetSubscriptions = nil
							updateSubscriptions = nil
							renewSubscriptions = nil
							createSubscriptions = nil
							deleteSubscriptions = nil
							existingSubscriptions = nil
							expectedMetadata = map[string]any{}
							mockOuraClient.EXPECT().PartnerURL().Return(partnerURL).AnyTimes()
							mockOuraClient.EXPECT().PartnerSecret().Return(partnerSecret).AnyTimes()
						})

						JustBeforeEach(func() {
							mockOuraClient.EXPECT().ListSubscriptions(gomock.Not(gomock.Nil())).Return(existingSubscriptions, nil)
						})

						withResetSubscriptions := func(inner func()) {
							It("fails if unable to delete subscription", func() {
								failedSubscription := resetSubscriptions[0]
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().DeleteSubscription(gomock.Not(gomock.Nil()), *failedSubscription.ID).Return(testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							Context("with successful delete subscriptions", func() {
								JustBeforeEach(func() {
									for _, subscription := range resetSubscriptions {
										mockOuraClient.EXPECT().DeleteSubscription(gomock.Not(gomock.Nil()), *subscription.ID).Return(nil)
									}
								})

								inner()
							})
						}

						withCreateSubscriptions := func(inner func()) {
							It("fails if unable to create subscription", func() {
								failedSubscription := createSubscriptions[0]
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().CreateSubscription(gomock.Not(gomock.Nil()), newCreateSubscription(*failedSubscription.DataType, *failedSubscription.EventType)).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							It("fails if create subscription is missing", func() {
								failedSubscription := createSubscriptions[0]
								mockOuraClient.EXPECT().CreateSubscription(gomock.Not(gomock.Nil()), newCreateSubscription(*failedSubscription.DataType, *failedSubscription.EventType)).Return(nil, nil)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailedProcessResultError(MatchError(fmt.Sprintf("created subscription is missing with data type %q and event type %q", *failedSubscription.DataType, *failedSubscription.EventType))))
							})

							Context("with successful create subscriptions", func() {
								JustBeforeEach(func() {
									for _, subscription := range createSubscriptions {
										mockOuraClient.EXPECT().CreateSubscription(gomock.Not(gomock.Nil()), newCreateSubscription(*subscription.DataType, *subscription.EventType)).Return(subscription, nil)
									}
								})

								inner()
							})
						}

						withUpdateRenewAndCreateSubscriptions := func(inner func()) {
							expectSubscriptionsUntil := func(until oura.Subscriptions, known oura.Subscriptions, update oura.Subscriptions, renew oura.Subscriptions, create oura.Subscriptions) *oura.Subscription {
								for _, subscription := range known {
									if slices.Contains(until, subscription) {
										return subscription
									}
									if slices.Contains(update, subscription) {
										mockOuraClient.EXPECT().UpdateSubscription(gomock.Not(gomock.Nil()), *subscription.ID, newUpdateSubscription(*subscription.DataType, *subscription.EventType)).Return(subscription, nil)
									}
									if slices.Contains(renew, subscription) {
										mockOuraClient.EXPECT().RenewSubscription(gomock.Not(gomock.Nil()), *subscription.ID).Return(subscription, nil)
									}
									if slices.Contains(create, subscription) {
										mockOuraClient.EXPECT().CreateSubscription(gomock.Not(gomock.Nil()), newCreateSubscription(*subscription.DataType, *subscription.EventType)).Return(subscription, nil)
									}
								}
								return nil
							}

							Context("with update subscription failing", func() {
								var failedSubscription *oura.Subscription

								JustBeforeEach(func() {
									failedSubscription = expectSubscriptionsUntil(updateSubscriptions, knownSubscriptions, updateSubscriptions, renewSubscriptions, createSubscriptions)
									Expect(failedSubscription).ToNot(BeNil())
								})

								It("fails if unable to update subscription", func() {
									testErr := errorsTest.RandomError()
									mockOuraClient.EXPECT().UpdateSubscription(gomock.Not(gomock.Nil()), *failedSubscription.ID, newUpdateSubscription(*failedSubscription.DataType, *failedSubscription.EventType)).Return(nil, testErr)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
								})

								It("fails if updated subscription is missing", func() {
									mockOuraClient.EXPECT().UpdateSubscription(gomock.Not(gomock.Nil()), *failedSubscription.ID, newUpdateSubscription(*failedSubscription.DataType, *failedSubscription.EventType)).Return(nil, nil)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailedProcessResultError(MatchError(fmt.Sprintf("updated subscription is missing with id %q, data type %q, and event type %q", *failedSubscription.ID, *failedSubscription.DataType, *failedSubscription.EventType))))
								})
							})

							Context("with renew subscription failing", func() {
								var failedSubscription *oura.Subscription

								JustBeforeEach(func() {
									failedSubscription = expectSubscriptionsUntil(renewSubscriptions, knownSubscriptions, updateSubscriptions, renewSubscriptions, createSubscriptions)
									Expect(failedSubscription).ToNot(BeNil())
									if slices.Contains(updateSubscriptions, failedSubscription) {
										mockOuraClient.EXPECT().UpdateSubscription(gomock.Not(gomock.Nil()), *failedSubscription.ID, newUpdateSubscription(*failedSubscription.DataType, *failedSubscription.EventType)).Return(failedSubscription, nil)
									}
								})

								It("fails if unable to parse subscription expiration time", func() {
									failedSubscription.ExpirationTime = pointer.From("invalid")
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(fmt.Sprintf(`unable to parse expiration time of existing subscription with id %q, data type %q, and event type %q; parsing time "invalid" as "2006-01-02T15:04:05.999999999": cannot parse "invalid" as "2006"`, *failedSubscription.ID, *failedSubscription.DataType, *failedSubscription.EventType))))
								})

								It("fails if unable to renew subscription", func() {
									testErr := errorsTest.RandomError()
									mockOuraClient.EXPECT().RenewSubscription(gomock.Not(gomock.Nil()), *failedSubscription.ID).Return(nil, testErr)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
								})

								It("fails if renewed subscription is missing", func() {
									mockOuraClient.EXPECT().RenewSubscription(gomock.Not(gomock.Nil()), *failedSubscription.ID).Return(nil, nil)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailedProcessResultError(MatchError(fmt.Sprintf("renewed subscription is missing with id %q, data type %q, and event type %q", *failedSubscription.ID, *failedSubscription.DataType, *failedSubscription.EventType))))
								})
							})

							Context("with create subscription failing", func() {
								var failedSubscription *oura.Subscription

								JustBeforeEach(func() {
									failedSubscription = expectSubscriptionsUntil(createSubscriptions, knownSubscriptions, updateSubscriptions, renewSubscriptions, createSubscriptions)
									Expect(failedSubscription).ToNot(BeNil())
								})

								It("fails if unable to create subscription", func() {
									testErr := errorsTest.RandomError()
									mockOuraClient.EXPECT().CreateSubscription(gomock.Not(gomock.Nil()), newCreateSubscription(*failedSubscription.DataType, *failedSubscription.EventType)).Return(nil, testErr)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
								})

								It("fails if create subscription is missing", func() {
									mockOuraClient.EXPECT().CreateSubscription(gomock.Not(gomock.Nil()), newCreateSubscription(*failedSubscription.DataType, *failedSubscription.EventType)).Return(nil, nil)
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailedProcessResultError(MatchError(fmt.Sprintf("created subscription is missing with data type %q and event type %q", *failedSubscription.DataType, *failedSubscription.EventType))))
								})
							})

							Context("with successful update and renew subscriptions", func() {
								var failedSubscription *oura.Subscription

								JustBeforeEach(func() {
									failedSubscription = expectSubscriptionsUntil(nil, knownSubscriptions, updateSubscriptions, renewSubscriptions, createSubscriptions)
									Expect(failedSubscription).To(BeNil())
								})

								inner()
							})
						}

						withDeleteSubscriptions := func(inner func()) {
							It("fails if unable to delete subscription", func() {
								var failedSubscription *oura.Subscription
								for _, subscription := range existingSubscriptions {
									if slices.Contains(deleteSubscriptions, subscription) {
										failedSubscription = subscription
										break
									}
								}
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().DeleteSubscription(gomock.Not(gomock.Nil()), *failedSubscription.ID).Return(testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							Context("with successful delete subscriptions", func() {
								JustBeforeEach(func() {
									for _, subscription := range deleteSubscriptions {
										mockOuraClient.EXPECT().DeleteSubscription(gomock.Not(gomock.Nil()), *subscription.ID).Return(nil)
									}
								})

								inner()
							})
						}

						withSuccess := func() {
							It("returns successful", func() {
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
									MatchAllFields(Fields{
										"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraWebhookWorkSubscribe.PendingAvailableDuration), time.Second),
										"ProcessingPriority":      Equal(0),
										"ProcessingTimeout":       Equal(int(ouraWebhookWorkSubscribe.ProcessingTimeout.Seconds())),
										"Metadata":                Equal(expectedMetadata),
									}),
								))
							})
						}

						Context("without existing subscriptions", func() {
							Context("without override", func() {
								BeforeEach(func() {
									createSubscriptions = knownSubscriptions
								})

								withCreateSubscriptions(withSuccess)
							})

							Context("with disabled override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideDisabled}
									expectedMetadata = wrk.Metadata
								})

								withSuccess()
							})

							Context("with renew override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideRenew}
									createSubscriptions = knownSubscriptions
								})

								withCreateSubscriptions(withSuccess)
							})

							Context("with reset override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideReset}
									createSubscriptions = knownSubscriptions
								})

								withCreateSubscriptions(withSuccess)
							})

							Context("with update override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideUpdate}
									createSubscriptions = knownSubscriptions
								})

								withCreateSubscriptions(withSuccess)
							})
						})

						Context("with existing subscriptions", func() {
							BeforeEach(func() {
								// Start with all known subscriptions, randomized
								subscriptions := slices.Clone(knownSubscriptions)
								rand.Shuffle(len(subscriptions), func(i, j int) {
									subscriptions[i], subscriptions[j] = subscriptions[j], subscriptions[i]
								})

								// Break into segments to randomize which subscriptions are existing, update, renew, and create
								subscriptionSegments := make([]oura.Subscriptions, 6)
								start := 0
								for index, end := range slices.Sorted(slices.Values(append(rand.Perm(len(subscriptions) - 1)[:5], len(subscriptions)-1))) {
									subscriptionSegments[index] = subscriptions[start : end+1]
									start = end + 1
								}

								// Create
								createSubscriptions = slices.Clone(subscriptionSegments[0])

								// Update (missing callback URL)
								for _, subscription := range subscriptionSegments[1] {
									subscription.CallbackURL = nil
									updateSubscriptions = append(updateSubscriptions, subscription)
								}

								// Update (missing callback URL)
								for _, subscription := range subscriptionSegments[2] {
									subscription.CallbackURL = pointer.From(test.RandomString())
									updateSubscriptions = append(updateSubscriptions, subscription)
								}

								// Renew (expired)
								for index, subscription := range subscriptionSegments[3] {
									renewExpirationTime := now
									if index > 0 {
										renewExpirationTime = test.RandomTimeBefore(renewExpirationTime)
									}
									subscription.ExpirationTime = pointer.From(renewExpirationTime.Format(oura.SubscriptionExpirationTimeFormat))
									renewSubscriptions = append(renewSubscriptions, subscription)
								}

								// Renew (expires soon)
								for index, subscription := range subscriptionSegments[4] {
									renewExpirationTime := expirationTime.Add(-time.Minute)
									if index > 0 {
										renewExpirationTime = test.RandomTimeBefore(renewExpirationTime)
									}
									subscription.ExpirationTime = pointer.From(renewExpirationTime.Format(oura.SubscriptionExpirationTimeFormat))
									renewSubscriptions = append(renewSubscriptions, subscription)
								}

								// Delete final unknown subscriptions
								deleteSubscriptions = unknownSubscriptions

								// Existing subscription also adds update and renew subscriptions, randomized
								existingSubscriptions = slices.Clone(subscriptionSegments[5])
								existingSubscriptions = append(existingSubscriptions, updateSubscriptions...)
								existingSubscriptions = append(existingSubscriptions, renewSubscriptions...)
								existingSubscriptions = append(existingSubscriptions, deleteSubscriptions...)
								rand.Shuffle(len(existingSubscriptions), func(i, j int) {
									existingSubscriptions[i], existingSubscriptions[j] = existingSubscriptions[j], existingSubscriptions[i]
								})
							})

							// Assist with debugging failing tests
							AfterEach(func() {
								if CurrentSpecReport().Failed() {
									dumpSubscription := func(subscription *oura.Subscription) string {
										bites, _ := json.Marshal(subscription)
										return string(bites)
									}
									dumpSubscriptions := func(subscriptionsName string, subscriptions oura.Subscriptions) string {
										var dumpedSubscriptions []string
										for _, subscription := range subscriptions {
											dumpedSubscriptions = append(dumpedSubscriptions, dumpSubscription(subscription))
										}
										dumped := strings.Join(dumpedSubscriptions, "\n  ")
										if len(dumped) > 0 {
											dumped += "\n"
										}
										return fmt.Sprintf("%s: [%s]", subscriptionsName, dumped)
									}
									dumped := []string{
										fmt.Sprintf("now: %s", now.Format(time.RFC3339Nano)),
										fmt.Sprintf("expirationTime: %s", expirationTime.Format(time.RFC3339Nano)),
										dumpSubscriptions("createSubscriptions", createSubscriptions),
										dumpSubscriptions("updateSubscriptions", updateSubscriptions),
										dumpSubscriptions("renewSubscriptions", renewSubscriptions),
										dumpSubscriptions("deleteSubscriptions", deleteSubscriptions),
										dumpSubscriptions("existingSubscriptions", existingSubscriptions),
									}
									GinkgoWriter.Println(strings.Join(dumped, "\n"))
								}
							})

							Context("without override", func() {
								withUpdateRenewAndCreateSubscriptions(func() {
									withDeleteSubscriptions(withSuccess)
								})
							})

							Context("with disabled override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideDisabled}
									expectedMetadata = wrk.Metadata
									resetSubscriptions = existingSubscriptions
								})

								withResetSubscriptions(withSuccess)
							})

							Context("with renew override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideRenew}
									renewSubscriptions = slices.DeleteFunc(slices.Clone(existingSubscriptions), func(subscription *oura.Subscription) bool {
										return slices.Contains(deleteSubscriptions, subscription)
									})
								})

								withUpdateRenewAndCreateSubscriptions(func() {
									withDeleteSubscriptions(withSuccess)
								})
							})

							Context("with reset override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideReset}
									resetSubscriptions = slices.Clone(existingSubscriptions)
									createSubscriptions = slices.Clone(knownSubscriptions)
								})

								withResetSubscriptions(func() {
									withCreateSubscriptions(withSuccess)
								})
							})

							Context("with update override", func() {
								BeforeEach(func() {
									wrk.Metadata = map[string]any{ouraWebhookWorkSubscribe.MetadataKeyOverride: ouraWebhookWorkSubscribe.OverrideUpdate}
									updateSubscriptions = slices.DeleteFunc(slices.Clone(existingSubscriptions), func(subscription *oura.Subscription) bool {
										return slices.Contains(deleteSubscriptions, subscription)
									})
								})

								withUpdateRenewAndCreateSubscriptions(func() {
									withDeleteSubscriptions(withSuccess)
								})
							})
						})
					})
				})
			})
		})
	})
})
