package subscribe_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
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
	"github.com/tidepool-org/platform/pointer"
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
		Expect(ouraWebhookWorkSubscribe.FailingRetryDuration).To(Equal(10 * time.Minute))
	})

	It("FailingRetryDurationJitter is expected", func() {
		Expect(ouraWebhookWorkSubscribe.FailingRetryDurationJitter).To(Equal(time.Minute))
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
						mockOuraClient.EXPECT().ListSubscriptions(gomock.Any()).Return(nil, testErr)
						Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
					})

					Context("with list subscriptions successful", func() {
						var now time.Time
						var expirationTime time.Time
						var partnerURL string
						var partnerSecret string

						BeforeEach(func() {
							now = time.Now()
							expirationTime = now.Add(ouraWebhookWorkSubscribe.ExpirationTimeDurationMinimum)
							partnerURL = test.RandomString()
							partnerSecret = test.RandomString()
							mockOuraClient.EXPECT().PartnerURL().Return(partnerURL)
							mockOuraClient.EXPECT().PartnerSecret().Return(partnerSecret)
						})

						_subscription := func(dataType string, eventType string) *oura.Subscription {
							return &oura.Subscription{
								ID:             pointer.FromString(test.RandomString()),
								CallbackURL:    pointer.FromString(partnerURL + ouraWebhook.EventPath),
								DataType:       pointer.FromString(dataType),
								EventType:      pointer.FromString(eventType),
								ExpirationTime: pointer.FromString(test.RandomTimeAfter(expirationTime.Add(time.Minute)).Format(oura.SubscriptionExpirationTimeFormat)),
							}
						}

						_createSubscription := func(dataType string, eventType string) *oura.CreateSubscription {
							return &oura.CreateSubscription{
								CallbackURL:       pointer.FromString(partnerURL + ouraWebhook.EventPath),
								VerificationToken: pointer.FromString(partnerSecret),
								DataType:          pointer.FromString(dataType),
								EventType:         pointer.FromString(eventType),
							}
						}

						_updateSubscription := func(dataType string, eventType string) *oura.UpdateSubscription {
							return &oura.UpdateSubscription{
								CallbackURL:       pointer.FromString(partnerURL + ouraWebhook.EventPath),
								VerificationToken: pointer.FromString(partnerSecret),
								DataType:          pointer.FromString(dataType),
								EventType:         pointer.FromString(eventType),
							}
						}
						Context("without existing subscriptions", func() {
							BeforeEach(func() {
								mockOuraClient.EXPECT().ListSubscriptions(gomock.Any()).Return(oura.Subscriptions{}, nil)
							})

							It("fails if unable to create subscription", func() {
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().CreateSubscription(gomock.Any(), _createSubscription(oura.DataTypeDailyActivity, oura.EventTypeCreate)).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							It("returns successful after creating subscriptions for all event types and data types", func() {
								for _, dataType := range ouraWebhook.DataTypes() {
									for _, eventType := range oura.EventTypes() {
										mockOuraClient.EXPECT().CreateSubscription(gomock.Any(), _createSubscription(dataType, eventType)).Return(_subscription(dataType, eventType), nil)
									}
								}
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
									MatchAllFields(Fields{
										"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraWebhookWorkSubscribe.PendingAvailableDuration), time.Second),
										"ProcessingPriority":      Equal(0),
										"ProcessingTimeout":       Equal(int(ouraWebhookWorkSubscribe.ProcessingTimeout.Seconds())),
										"Metadata":                BeNil(),
									}),
								))
							})
						})

						Context("with existing subscriptions", func() {
							It("fails if unable to update subscription", func() {
								subscription := _subscription(oura.DataTypeDailyActivity, oura.EventTypeCreate)
								subscription.CallbackURL = pointer.FromString(test.RandomString())
								mockOuraClient.EXPECT().ListSubscriptions(gomock.Any()).Return(oura.Subscriptions{subscription}, nil)
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().UpdateSubscription(gomock.Any(), *subscription.ID, _updateSubscription(oura.DataTypeDailyActivity, oura.EventTypeCreate)).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							It("fails if unable to parse subscription expiration time", func() {
								subscription := _subscription(oura.DataTypeDailyActivity, oura.EventTypeCreate)
								subscription.ExpirationTime = pointer.FromString("invalid")
								mockOuraClient.EXPECT().ListSubscriptions(gomock.Any()).Return(oura.Subscriptions{subscription}, nil)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(fmt.Sprintf(`unable to parse expiration time of existing subscription with id "%s", data type "daily_activity", and event type "create"; parsing time "invalid" as "2006-01-02T15:04:05.999999999": cannot parse "invalid" as "2006"`, *subscription.ID))))
							})

							It("fails if unable to renew subscription", func() {
								subscription := _subscription(oura.DataTypeDailyActivity, oura.EventTypeCreate)
								subscription.ExpirationTime = pointer.FromString(test.RandomTimeFromRange(now, expirationTime).Format(oura.SubscriptionExpirationTimeFormat))
								mockOuraClient.EXPECT().ListSubscriptions(gomock.Any()).Return(oura.Subscriptions{subscription}, nil)
								testErr := errorsTest.RandomError()
								mockOuraClient.EXPECT().RenewSubscription(gomock.Any(), *subscription.ID).Return(nil, testErr)
								Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchFailingProcessResultError(MatchError(testErr)))
							})

							Context("with successful create, update, and renew", func() {
								var createSubscriptions oura.Subscriptions
								var updateSubscriptions oura.Subscriptions
								var renewSubscriptions oura.Subscriptions

								BeforeEach(func() {
									// All possible subscriptions, randomized
									subscriptions := oura.Subscriptions{}
									for _, dataType := range ouraWebhook.DataTypes() {
										for _, eventType := range oura.EventTypes() {
											subscriptions = append(subscriptions, _subscription(dataType, eventType))
										}
									}
									rand.Shuffle(len(subscriptions), func(i, j int) {
										subscriptions[i], subscriptions[j] = subscriptions[j], subscriptions[i]
									})

									// Remove zero to all as not existing
									for range test.RandomIntFromRange(0, len(subscriptions)) {
										index := test.RandomIntFromRange(0, len(subscriptions)-1)
										createSubscriptions = append(createSubscriptions, subscriptions[index])
										subscriptions = slices.Delete(subscriptions, index, index+1)
									}

									// Break into different existing subscription use cases
									for _, subscription := range subscriptions {
										switch test.RandomIntFromRange(0, 6) {
										case 0:
											subscription.CallbackURL = nil
											updateSubscriptions = append(updateSubscriptions, subscription)
										case 1:
											subscription.CallbackURL = pointer.FromString(test.RandomString())
											updateSubscriptions = append(updateSubscriptions, subscription)
										case 2:
											subscription.ExpirationTime = pointer.FromString(test.RandomTimeBefore(now).Format(oura.SubscriptionExpirationTimeFormat))
											renewSubscriptions = append(renewSubscriptions, subscription)
										case 3:
											subscription.ExpirationTime = pointer.FromString(test.RandomTimeFromRange(now, expirationTime).Format(oura.SubscriptionExpirationTimeFormat))
											renewSubscriptions = append(renewSubscriptions, subscription)
										default:
											// Existing, but valid (does not require update or renew)
										}
									}

									// All expectations
									for _, subscription := range createSubscriptions {
										mockOuraClient.EXPECT().CreateSubscription(gomock.Any(), _createSubscription(*subscription.DataType, *subscription.EventType)).Return(subscription, nil)
									}
									for _, subscription := range updateSubscriptions {
										mockOuraClient.EXPECT().UpdateSubscription(gomock.Any(), *subscription.ID, _updateSubscription(*subscription.DataType, *subscription.EventType)).Return(subscription, nil)
									}
									for _, subscription := range renewSubscriptions {
										mockOuraClient.EXPECT().RenewSubscription(gomock.Any(), *subscription.ID).Return(subscription, nil)
									}

									// Shuffle again
									rand.Shuffle(len(subscriptions), func(i, j int) {
										subscriptions[i], subscriptions[j] = subscriptions[j], subscriptions[i]
									})
									mockOuraClient.EXPECT().ListSubscriptions(gomock.Any()).Return(subscriptions, nil)
								})

								It("returns successful after creating subscriptions for all event types and data types", func() {
									Expect(processor.Process(ctx, wrk, mockProcessingUpdater)).To(workTest.MatchPendingProcessResult(
										MatchAllFields(Fields{
											"ProcessingAvailableTime": BeTemporally("~", time.Now().Add(ouraWebhookWorkSubscribe.PendingAvailableDuration), time.Second),
											"ProcessingPriority":      Equal(0),
											"ProcessingTimeout":       Equal(int(ouraWebhookWorkSubscribe.ProcessingTimeout.Seconds())),
											"Metadata":                BeNil(),
										}),
									))
								})
							})
						})
					})
				})
			})
		})
	})
})
