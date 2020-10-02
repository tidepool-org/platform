package mongo_test

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	messageStore "github.com/tidepool-org/platform/message/store"
	messageStoreMongo "github.com/tidepool-org/platform/message/store/mongo"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
)

func NewMessage(groupID string, userID string) bson.M {
	timestamp := test.RandomTime()
	createdTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
	modifiedTime := test.RandomTimeFromRange(createdTime, time.Now())
	return bson.M{
		"groupid":      groupID,
		"userid":       userID,
		"guid":         test.RandomString(),
		"timestamp":    timestamp.Format(time.RFC3339Nano),
		"createdtime":  createdTime.Format(time.RFC3339Nano),
		"modifiedtime": modifiedTime.Format(time.RFC3339Nano),
		"messagetext":  "test",
	}
}

func NewMessages(groupID string, userID string) []interface{} {
	messages := []interface{}{}
	for count := 0; count < 2; count++ {
		messages = append(messages, NewMessage(groupID, userID))
	}
	parentMessage := NewMessage(groupID, userID)
	messages = append(messages, parentMessage)
	for count := 0; count < 2; count++ {
		message := NewMessage(groupID, userID)
		message["parentmessage"] = parentMessage["guid"]
		messages = append(messages, message)
	}
	return messages
}

func MarkMessagesDeleted(messages []interface{}) {
	for index, message := range messages {
		messages[index] = MarkMessageDeleted(message.(bson.M))
	}
}

func MarkMessageDeleted(message bson.M) bson.M {
	message["user"] = bson.M{
		"fullName": fmt.Sprintf("deleted user (%s)", message["userid"]),
	}
	delete(message, "userid")
	return message
}

func ValidateMessages(testMongoCollection *mongo.Collection, selector bson.M, expectedMessages []interface{}) {
	var actualMessages []bson.M
	opts := options.Find().SetProjection(bson.M{"_id": 0})
	cursor, err := testMongoCollection.Find(context.Background(), selector, opts)
	Expect(err).ToNot(HaveOccurred())
	Expect(cursor).ToNot(BeNil())
	Expect(cursor.All(context.Background(), &actualMessages)).To(Succeed())
	Expect(actualMessages).To(ConsistOf(expectedMessages...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *storeStructuredMongo.Config
	var mongoStore *messageStoreMongo.Store
	var repository messageStore.MessageRepository

	BeforeEach(func() {
		mongoConfig = storeStructuredMongoTest.NewConfig()
	})

	Context("New", func() {
		It("returns an error if unsuccessful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: nil}
			mongoStore, err = messageStoreMongo.NewStore(params)
			Expect(err).To(HaveOccurred())
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: mongoConfig}
			mongoStore, err = messageStoreMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: mongoConfig}
			mongoStore, err = messageStoreMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewMessageRepository", func() {
			It("returns a new repository", func() {
				repository = mongoStore.NewMessageRepository()
				Expect(repository).ToNot(BeNil())
			})
		})

		Context("with a new repository", func() {
			BeforeEach(func() {
				repository = mongoStore.NewMessageRepository()
				Expect(repository).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoCollection *mongo.Collection
				var messages []interface{}
				var ctx context.Context

				BeforeEach(func() {
					testMongoCollection = mongoStore.GetCollection("messages")
					messages = append(NewMessages(user.NewID(), user.NewID()), NewMessages(user.NewID(), user.NewID())...)
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					_, err := testMongoCollection.InsertMany(ctx, messages)
					Expect(err).ToNot(HaveOccurred())
				})

				Context("DeleteMessagesFromUser", func() {
					var deleteGroupID string
					var deleteUserID string
					var deleteUser *messageStore.User
					var deleteMessages []interface{}

					BeforeEach(func() {
						deleteGroupID = user.NewID()
						deleteUserID = user.NewID()
						deleteUser = &messageStore.User{
							ID:       deleteUserID,
							FullName: fmt.Sprintf("deleted user (%s)", deleteUserID),
						}
						deleteMessages = NewMessages(deleteGroupID, deleteUserID)
						messages = append(messages, NewMessages(deleteUserID, deleteGroupID)...)
					})

					JustBeforeEach(func() {
						_, err := testMongoCollection.InsertMany(ctx, deleteMessages)
						Expect(err).ToNot(HaveOccurred())
					})

					It("succeeds if it successfully removes messages", func() {
						Expect(repository.DeleteMessagesFromUser(ctx, deleteUser)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(repository.DeleteMessagesFromUser(nil, deleteUser)).To(MatchError("context is missing"))
					})

					It("returns an error if the user is missing", func() {
						Expect(repository.DeleteMessagesFromUser(ctx, nil)).To(MatchError("user is missing"))
					})

					It("returns an error if the user id is missing", func() {
						deleteUser.ID = ""
						Expect(repository.DeleteMessagesFromUser(ctx, deleteUser)).To(MatchError("user id is missing"))
					})

					It("has the correct stored messages", func() {
						ValidateMessages(testMongoCollection, bson.M{}, append(messages, deleteMessages...))
						Expect(repository.DeleteMessagesFromUser(ctx, deleteUser)).To(Succeed())
						MarkMessagesDeleted(deleteMessages)
						ValidateMessages(testMongoCollection, bson.M{}, append(messages, deleteMessages...))
					})
				})

				Context("DestroyMessagesForUserByID", func() {
					var destroyGroupID string
					var destroyUserID string
					var destroyMessages []interface{}

					BeforeEach(func() {
						destroyGroupID = user.NewID()
						destroyUserID = user.NewID()
						destroyMessages = NewMessages(destroyGroupID, destroyUserID)
						messages = append(messages, NewMessages(destroyUserID, destroyGroupID)...)
					})

					JustBeforeEach(func() {
						_, err := testMongoCollection.InsertMany(ctx, destroyMessages)
						Expect(err).ToNot(HaveOccurred())
					})

					It("succeeds if it successfully removes messages", func() {
						Expect(repository.DestroyMessagesForUserByID(ctx, destroyGroupID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(repository.DestroyMessagesForUserByID(nil, destroyGroupID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(repository.DestroyMessagesForUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("has the correct stored messages", func() {
						ValidateMessages(testMongoCollection, bson.M{}, append(messages, destroyMessages...))
						Expect(repository.DestroyMessagesForUserByID(ctx, destroyGroupID)).To(Succeed())
						ValidateMessages(testMongoCollection, bson.M{}, messages)
					})
				})
			})
		})
	})
})
