package mongo_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"go.mongodb.org/mongo-driver/bson"
	bsonPrimitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/pointer"
	storeStructured "github.com/tidepool-org/platform/store/structured"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
	workStoreStructuredMongo "github.com/tidepool-org/platform/work/store/structured/mongo"
)

const processingTimeout = 300

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var store *workStoreStructuredMongo.Store
	var ctx context.Context
	var typ string

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
		typ = netTest.RandomReverseDomain()
	})

	AfterEach(func() {
		if store != nil {
			Expect(store.Terminate(context.Background())).ToNot(HaveOccurred())
			store = nil
		}
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			store, err = workStoreStructuredMongo.NewStore(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			Expect(store.EnsureIndexes()).To(Succeed())
		})

		Context("EnsureIndexes", func() {
			It("creates an index over processing timeout time restricted to processing work", func() {
				cursor, err := store.GetCollection("work").Indexes().List(ctx)
				Expect(err).ToNot(HaveOccurred())
				var indexes []storeStructuredMongoTest.MongoIndex
				Expect(cursor.All(ctx, &indexes)).To(Succeed())
				Expect(indexes).To(ContainElement(MatchFields(IgnoreExtras, Fields{
					"Key":                     Equal(storeStructuredMongoTest.MakeKeySlice("processingTimeoutTime")),
					"Name":                    Equal("ProcessingTimeoutTime"),
					"PartialFilterExpression": Equal(bson.D{{Key: "state", Value: work.StateProcessing}}),
				})))
			})
		})

		Context("ReapExpiredProcessing", func() {
			const reapGraceDuration = time.Minute

			It("returns an error when the grace duration is negative", func() {
				_, err := store.ReapExpiredProcessing(ctx, -time.Second)
				Expect(err).To(MatchError("grace duration is invalid"))
			})

			var collection *mongo.Collection
			var poll *work.Poll
			var serialID string
			var claimed *work.Work

			// Expires the processing timeout time of work directly, as the alternative is to wait
			// out both the processing timeout and the reap grace duration
			expireProcessingTimeoutTime := func(workID string, timeoutTime time.Time) {
				objectID, err := bsonPrimitive.ObjectIDFromHex(workID)
				Expect(err).ToNot(HaveOccurred())
				result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": bson.M{"processingTimeoutTime": timeoutTime}})
				Expect(err).ToNot(HaveOccurred())
				Expect(result.ModifiedCount).To(Equal(int64(1)))
			}

			BeforeEach(func() {
				collection = store.GetCollection("work")
				poll = &work.Poll{TypeQuantities: work.TypeQuantities{typ: 10}}
				serialID = typ + ":" + test.RandomString()

				created, err := store.Create(ctx, &work.Create{
					Type:              typ,
					SerialID:          pointer.FromString(serialID),
					ProcessingTimeout: processingTimeout,
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(created).ToNot(BeNil())

				polled, err := store.Poll(ctx, poll)
				Expect(err).ToNot(HaveOccurred())
				Expect(polled).To(HaveLen(1))
				claimed = polled[0]
				Expect(claimed.State).To(Equal(work.StateProcessing))
				Expect(claimed.ProcessingTimeoutTime).ToNot(BeNil())
			})

			Context("with work processing beyond the grace duration", func() {
				BeforeEach(func() {
					expireProcessingTimeoutTime(claimed.ID, time.Now().Add(-reapGraceDuration-time.Minute))
				})

				It("returns the work to failing with an immediate retry", func() {
					count, err := store.ReapExpiredProcessing(ctx, reapGraceDuration)
					Expect(err).ToNot(HaveOccurred())
					Expect(count).To(Equal(1))

					reaped, err := store.Get(ctx, claimed.ID, nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(reaped).ToNot(BeNil())
					Expect(reaped.State).To(Equal(work.StateFailing))
					Expect(reaped.FailingTime).ToNot(BeNil())
					Expect(reaped.FailingError).ToNot(BeNil())
					Expect(reaped.FailingError.Error).To(MatchError(ContainSubstring("processing timeout expired")))
					Expect(reaped.FailingRetryCount).To(PointTo(Equal(1)))
					Expect(reaped.FailingRetryTime).ToNot(BeNil())
					Expect(*reaped.FailingRetryTime).To(BeTemporally("<=", time.Now()))
					Expect(reaped.Revision).To(Equal(claimed.Revision + 1))
				})

				// State failing requires the processing timeout time to be absent and, as the work
				// was processing, the processing duration to be present
				It("clears the processing timeout time and records the processing duration", func() {
					_, err := store.ReapExpiredProcessing(ctx, reapGraceDuration)
					Expect(err).ToNot(HaveOccurred())

					reaped, err := store.Get(ctx, claimed.ID, nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(reaped).ToNot(BeNil())
					Expect(reaped.ProcessingTimeoutTime).To(BeNil())
					Expect(reaped.ProcessingDuration).ToNot(BeNil())
					Expect(*reaped.ProcessingDuration).To(BeNumerically(">=", 0))
				})

				It("allows the reaped work to be polled again", func() {
					Expect(store.ReapExpiredProcessing(ctx, reapGraceDuration)).To(Equal(1))

					polled, err := store.Poll(ctx, poll)
					Expect(err).ToNot(HaveOccurred())
					Expect(polled).To(HaveLen(1))
					Expect(polled[0].ID).To(Equal(claimed.ID))
					Expect(polled[0].State).To(Equal(work.StateProcessing))
				})

				// The original worker may still be running and report its completion, which is
				// conditional upon the revision it holds and must no longer be applied
				It("prevents the completion reported with the revision held before the reap", func() {
					Expect(store.ReapExpiredProcessing(ctx, reapGraceDuration)).To(Equal(1))

					condition := &storeStructured.Condition{Revision: pointer.FromInt(claimed.Revision)}
					updated, err := store.Update(ctx, claimed.ID, condition, &work.Update{
						State:         work.StateSuccess,
						SuccessUpdate: &work.SuccessUpdate{},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(updated).To(BeNil())

					deleted, err := store.Delete(ctx, claimed.ID, condition)
					Expect(err).ToNot(HaveOccurred())
					Expect(deleted).To(BeNil())

					unchanged, err := store.Get(ctx, claimed.ID, nil)
					Expect(err).ToNot(HaveOccurred())
					Expect(unchanged).ToNot(BeNil())
					Expect(unchanged.State).To(Equal(work.StateFailing))
				})

				It("unblocks work sharing the serial id of the reaped work", func() {
					sibling, err := store.Create(ctx, &work.Create{
						Type:              typ,
						SerialID:          pointer.FromString(serialID),
						ProcessingTimeout: processingTimeout,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(sibling).ToNot(BeNil())

					blocked, err := store.Poll(ctx, poll)
					Expect(err).ToNot(HaveOccurred())
					Expect(blocked).To(BeEmpty())

					Expect(store.ReapExpiredProcessing(ctx, reapGraceDuration)).To(Equal(1))

					polled, err := store.Poll(ctx, poll)
					Expect(err).ToNot(HaveOccurred())
					Expect(polled).To(HaveLen(1))
				})
			})

			It("does not reap work processing within the grace duration", func() {
				expireProcessingTimeoutTime(claimed.ID, time.Now().Add(-time.Second))

				count, err := store.ReapExpiredProcessing(ctx, reapGraceDuration)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(0))

				unchanged, err := store.Get(ctx, claimed.ID, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(unchanged).ToNot(BeNil())
				Expect(unchanged.State).To(Equal(work.StateProcessing))
				Expect(unchanged.Revision).To(Equal(claimed.Revision))
			})

			It("does not reap work that is not processing", func() {
				pending, err := store.Create(ctx, &work.Create{
					Type:                    typ,
					ProcessingAvailableTime: time.Now().Add(time.Hour),
					ProcessingTimeout:       processingTimeout,
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(pending).ToNot(BeNil())
				expireProcessingTimeoutTime(pending.ID, time.Now().Add(-reapGraceDuration-time.Minute))

				count, err := store.ReapExpiredProcessing(ctx, reapGraceDuration)
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(Equal(0))

				unchanged, err := store.Get(ctx, pending.ID, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(unchanged).ToNot(BeNil())
				Expect(unchanged.State).To(Equal(work.StatePending))
				Expect(unchanged.Revision).To(Equal(pending.Revision))
			})
		})

		Context("QueueSizes", func() {
			It("reports nothing with no work present", func() {
				Expect(store.QueueSizes(ctx)).To(BeEmpty())
			})

			It("reports the number of work items by type and state", func() {
				secondType := netTest.RandomReverseDomain()
				for workType, count := range map[string]int{typ: 3, secondType: 2} {
					for range count {
						created, err := store.Create(ctx, &work.Create{
							Type:              workType,
							ProcessingTimeout: processingTimeout,
						})
						Expect(err).ToNot(HaveOccurred())
						Expect(created).ToNot(BeNil())
					}
				}

				claimed, err := store.Poll(ctx, &work.Poll{TypeQuantities: work.TypeQuantities{typ: 1}})
				Expect(err).ToNot(HaveOccurred())
				Expect(claimed).To(HaveLen(1))

				Expect(store.QueueSizes(ctx)).To(ConsistOf(
					work.QueueSize{Type: typ, State: work.StatePending, Count: 2},
					work.QueueSize{Type: typ, State: work.StateProcessing, Count: 1},
					work.QueueSize{Type: secondType, State: work.StatePending, Count: 2},
				))
			})
		})

		Context("Poll", func() {
			// These work items intentionally share an identical processing available time and
			// processing priority so that only the identifier tie breaker in the Poll aggregation
			// gives them a total order. Without a total order the document reported first within a
			// serial id group is arbitrary, which allows a pending work item to be claimed while a
			// sibling sharing its serial id is still processing.
			Context("with multiple pending work items sharing a serial id and sort key", func() {
				const workCount = 10

				var serialID string
				var availableTime time.Time

				BeforeEach(func() {
					serialID = typ + ":" + test.RandomString()

					// Create in the future so every work item retains the exact same processing
					// available time, then wait for them to become available to poll
					availableTime = time.Now().Add(time.Second).UTC().Truncate(time.Millisecond)

					for range workCount {
						create := &work.Create{
							Type:                    typ,
							SerialID:                pointer.FromString(serialID),
							ProcessingAvailableTime: availableTime,
							ProcessingTimeout:       processingTimeout,
						}
						created, err := store.Create(ctx, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(created).ToNot(BeNil())
						Expect(created.ProcessingAvailableTime).To(BeTemporally("==", availableTime))
					}

					time.Sleep(time.Until(availableTime) + 100*time.Millisecond)
				})

				It("claims one work item and claims no further work item while it is processing", func() {
					poll := &work.Poll{TypeQuantities: work.TypeQuantities{typ: workCount}}

					claimed, err := store.Poll(ctx, poll)
					Expect(err).ToNot(HaveOccurred())
					Expect(claimed).To(HaveLen(1))
					Expect(claimed[0].State).To(Equal(work.StateProcessing))

					for index := range 20 {
						additional, err := store.Poll(ctx, poll)
						Expect(err).ToNot(HaveOccurred())
						Expect(additional).To(BeEmpty(), "poll %d claimed work while a work item sharing its serial id was processing", index)
					}
				})
			})

			// The serial id group exclusion must consider every member of the group, not only the
			// member that sorts first: a pending sibling with a higher processing priority sorts
			// ahead of a processing or failing member, which must still block the group.
			Context("with a work item claimed from a serial id group", func() {
				var poll *work.Poll
				var serialID string
				var claimed *work.Work

				createSibling := func(processingPriority int) *work.Work {
					sibling, err := store.Create(ctx, &work.Create{
						Type:               typ,
						SerialID:           pointer.FromString(serialID),
						ProcessingPriority: processingPriority,
						ProcessingTimeout:  processingTimeout,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(sibling).ToNot(BeNil())
					return sibling
				}

				updateToFailing := func(wrk *work.Work, retryTime time.Time) {
					updated, err := store.Update(ctx, wrk.ID, nil, &work.Update{
						State: work.StateFailing,
						FailingUpdate: &work.FailingUpdate{
							FailingError:      errors.Serializable{Error: errors.New("test failure")},
							FailingRetryCount: 1,
							FailingRetryTime:  retryTime,
						},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(updated).ToNot(BeNil())
					Expect(updated.State).To(Equal(work.StateFailing))
				}

				BeforeEach(func() {
					poll = &work.Poll{TypeQuantities: work.TypeQuantities{typ: 10}}
					serialID = typ + ":" + test.RandomString()

					createSibling(0)
					polled, err := store.Poll(ctx, poll)
					Expect(err).ToNot(HaveOccurred())
					Expect(polled).To(HaveLen(1))
					claimed = polled[0]
					Expect(claimed.State).To(Equal(work.StateProcessing))
				})

				It("claims nothing while the work item is processing, however a pending sibling sorts", func() {
					createSibling(1)

					for index := range 10 {
						polled, err := store.Poll(ctx, poll)
						Expect(err).ToNot(HaveOccurred())
						Expect(polled).To(BeEmpty(), "poll %d claimed work while a work item sharing its serial id was processing", index)
					}
				})

				It("claims nothing while the work item is failing with a future retry time, however a pending sibling sorts", func() {
					updateToFailing(claimed, time.Now().Add(time.Hour))
					createSibling(1)

					for index := range 10 {
						polled, err := store.Poll(ctx, poll)
						Expect(err).ToNot(HaveOccurred())
						Expect(polled).To(BeEmpty(), "poll %d claimed work while a work item sharing its serial id was failing with a future retry", index)
					}
				})

				It("claims the work item once its failing retry time has passed", func() {
					// The store clamps the failing retry time to no earlier than the update itself,
					// so requesting a past time yields a retry time that has passed by the poll
					updateToFailing(claimed, time.Now().Add(-time.Minute))

					polled, err := store.Poll(ctx, poll)
					Expect(err).ToNot(HaveOccurred())
					Expect(polled).To(HaveLen(1))
					Expect(polled[0].ID).To(Equal(claimed.ID))
					Expect(polled[0].State).To(Equal(work.StateProcessing))
				})
			})
		})

		Context("Update", func() {
			var created *work.Work

			BeforeEach(func() {
				var err error
				created, err = store.Create(ctx, &work.Create{
					Type:                    typ,
					GroupID:                 pointer.FromString(typ + ":" + test.RandomString()),
					SerialID:                pointer.FromString(typ + ":" + test.RandomString()),
					ProcessingAvailableTime: time.Now().Add(time.Hour),
					ProcessingTimeout:       processingTimeout,
					Metadata:                map[string]any{"reasons": []any{"DATA_ADDED"}},
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(created).ToNot(BeNil())
				Expect(created.State).To(Equal(work.StatePending))
				Expect(created.Revision).To(Equal(1))
			})

			// The producer coalesces repeated triggers by merging into the work item already
			// pending for a group, which requires a pending to pending update that revises both the
			// metadata and the processing available time while retaining the revision condition.
			Context("with a pending work item updated to pending", func() {
				var availableTime time.Time
				var update *work.Update

				BeforeEach(func() {
					availableTime = time.Now().Add(30 * time.Minute)
					update = &work.Update{
						State: work.StatePending,
						PendingUpdate: &work.PendingUpdate{
							ProcessingAvailableTime: availableTime,
							ProcessingPriority:      created.ProcessingPriority,
							ProcessingTimeout:       created.ProcessingTimeout,
							Metadata:                map[string]any{"reasons": []any{"DATA_ADDED", "UPLOAD_COMPLETED"}},
						},
					}
				})

				It("returns the updated work item with the revised metadata and processing available time", func() {
					updated, err := store.Update(ctx, created.ID, &storeStructured.Condition{Revision: pointer.FromInt(created.Revision)}, update)
					Expect(err).ToNot(HaveOccurred())
					Expect(updated).ToNot(BeNil())
					Expect(updated.State).To(Equal(work.StatePending))
					Expect(updated.Revision).To(Equal(created.Revision + 1))
					Expect(updated.Metadata).To(HaveKey("reasons"))
					Expect(updated.Metadata["reasons"]).To(ConsistOf("DATA_ADDED", "UPLOAD_COMPLETED"))
					Expect(updated.ProcessingAvailableTime).To(BeTemporally("~", availableTime, time.Millisecond))
					Expect(updated.ProcessingTimeout).To(Equal(created.ProcessingTimeout))
				})

				It("retains the pending time of the work item", func() {
					updated, err := store.Update(ctx, created.ID, &storeStructured.Condition{Revision: pointer.FromInt(created.Revision)}, update)
					Expect(err).ToNot(HaveOccurred())
					Expect(updated).ToNot(BeNil())
					// Millisecond tolerance as the stored time is truncated to the BSON date precision
					Expect(updated.PendingTime).To(BeTemporally("~", created.PendingTime, time.Millisecond))
				})

				It("returns nil when the revision condition no longer matches", func() {
					updated, err := store.Update(ctx, created.ID, &storeStructured.Condition{Revision: pointer.FromInt(created.Revision)}, update)
					Expect(err).ToNot(HaveOccurred())
					Expect(updated).ToNot(BeNil())

					stale, err := store.Update(ctx, created.ID, &storeStructured.Condition{Revision: pointer.FromInt(created.Revision)}, update)
					Expect(err).ToNot(HaveOccurred())
					Expect(stale).To(BeNil())
				})
			})
		})
	})
})
