package mongo_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/user"
	userStore "github.com/tidepool-org/platform/user/store"
	userStoreMongo "github.com/tidepool-org/platform/user/store/mongo"
)

func NewUser(userID string) *user.User {
	email := fmt.Sprintf("%s@test.org", userID)
	createdTime := test.RandomTimeFromRange(test.RandomTimeMinimum(), time.Now())
	termsAcceptedTime := test.RandomTimeFromRange(createdTime, time.Now())
	return &user.User{
		ID:                userID,
		Email:             email,
		Emails:            []string{email},
		Roles:             []string{user.ClinicRole},
		TermsAcceptedTime: termsAcceptedTime.Format(time.RFC3339Nano),
		EmailVerified:     true,
		PasswordHash:      "1234567890",
		Hash:              test.RandomString(),
		Private:           map[string]*user.IDHash{"meta": {ID: "meta-id", Hash: "meta-hash"}},
		CreatedTime:       createdTime.Format(time.RFC3339Nano),
	}
}

func NewUsers() []interface{} {
	users := []interface{}{}
	users = append(users, NewUser(user.NewID()), NewUser(user.NewID()), NewUser(user.NewID()))
	return users
}

func ValidateUsers(testMongoCollection *mgo.Collection, selector bson.M, expectedUsers []interface{}) {
	var actualUsers []*user.User
	Expect(testMongoCollection.Find(selector).All(&actualUsers)).To(Succeed())
	Expect(actualUsers).To(ConsistOf(expectedUsers...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *userStoreMongo.Config
	var mongoStore *userStoreMongo.Store
	var mongoSession userStore.UsersSession

	BeforeEach(func() {
		mongoConfig = &userStoreMongo.Config{
			Config:       storeStructuredMongoTest.NewConfig(),
			PasswordSalt: "password-salt",
		}
	})

	AfterEach(func() {
		if mongoSession != nil {
			mongoSession.Close()
		}
		if mongoStore != nil {
			mongoStore.Close()
		}
	})

	Context("New", func() {
		It("returns an error if config is missing", func() {
			var err error
			mongoStore, err = userStoreMongo.NewStore(nil, logTest.NewLogger())
			Expect(err).To(MatchError("config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is missing", func() {
			var err error
			mongoConfig.Config = nil
			mongoStore, err = userStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is invalid", func() {
			var err error
			mongoConfig.Config.Addresses = nil
			mongoStore, err = userStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("config is invalid; addresses is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			var err error
			mongoConfig.PasswordSalt = ""
			mongoStore, err = userStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).To(MatchError("config is invalid; password salt is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if logger is missing", func() {
			var err error
			mongoStore, err = userStoreMongo.NewStore(mongoConfig, nil)
			Expect(err).To(MatchError("logger is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = userStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = userStoreMongo.NewStore(mongoConfig, logTest.NewLogger())
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewUsersSession", func() {
			It("returns a new session", func() {
				mongoSession = mongoStore.NewUsersSession()
				Expect(mongoSession).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewUsersSession()
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var users []interface{}
				var ctx context.Context

				BeforeEach(func() {
					testMongoSession = storeStructuredMongoTest.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.CollectionPrefix + "users")
					users = NewUsers()
					ctx = log.NewContextWithLogger(context.Background(), logTest.NewLogger())
				})

				JustBeforeEach(func() {
					Expect(testMongoCollection.Insert(users...)).To(Succeed())
				})

				AfterEach(func() {
					if testMongoSession != nil {
						testMongoSession.Close()
					}
				})

				Context("GetUserByID", func() {
					var getUserID string
					var getUser *user.User
					var getUserEmail string

					BeforeEach(func() {
						getUserID = user.NewID()
						getUser = NewUser(getUserID)
						getUserEmail = fmt.Sprintf("%s@test.org", getUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(getUser)).To(Succeed())
					})

					It("succeeds if it successfully gets the user", func() {
						usr, err := mongoSession.GetUserByID(ctx, getUserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(usr).ToNot(BeNil())
						Expect(usr.ID).To(Equal(getUserID))
						Expect(usr.Email).To(Equal(getUserEmail))
						Expect(usr.ProfileID).ToNot(BeNil())
						Expect(*usr.ProfileID).To(Equal("meta-id"))
					})

					It("succeeds even if two users exist with the same user id", func() {
						Expect(testMongoCollection.Insert(getUser)).To(Succeed())
						usr, err := mongoSession.GetUserByID(ctx, getUserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(usr).ToNot(BeNil())
						Expect(usr.ID).To(Equal(getUserID))
					})

					It("returns no error and no user if the user id is not found", func() {
						usr, err := mongoSession.GetUserByID(ctx, user.NewID())
						Expect(err).ToNot(HaveOccurred())
						Expect(usr).To(BeNil())
					})

					It("returns an error if the context is missing", func() {
						usr, err := mongoSession.GetUserByID(nil, getUserID)
						Expect(err).To(MatchError("context is missing"))
						Expect(usr).To(BeNil())
					})

					It("returns an error if the user id is missing", func() {
						usr, err := mongoSession.GetUserByID(ctx, "")
						Expect(err).To(MatchError("user id is missing"))
						Expect(usr).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						usr, err := mongoSession.GetUserByID(ctx, getUserID)
						Expect(err).To(MatchError("session closed"))
						Expect(usr).To(BeNil())
					})

					Context("with no private", func() {
						BeforeEach(func() {
							getUser.Private = nil
						})

						It("succeeds, but does not fill in the profile id", func() {
							usr, err := mongoSession.GetUserByID(ctx, getUserID)
							Expect(err).ToNot(HaveOccurred())
							Expect(usr).ToNot(BeNil())
							Expect(usr.ProfileID).To(BeNil())
						})
					})

					Context("with private, but no meta", func() {
						BeforeEach(func() {
							getUser.Private = map[string]*user.IDHash{"other": {ID: "other-id", Hash: "other-hash"}}
						})

						It("succeeds, but does not fill in the profile id", func() {
							usr, err := mongoSession.GetUserByID(ctx, getUserID)
							Expect(err).ToNot(HaveOccurred())
							Expect(usr).ToNot(BeNil())
							Expect(usr.ProfileID).To(BeNil())
						})
					})

					Context("with private and meta, but missing ID", func() {
						BeforeEach(func() {
							getUser.Private = map[string]*user.IDHash{"meta": {Hash: "meta-hash"}}
						})

						It("succeeds, but does not fill in the profile id", func() {
							usr, err := mongoSession.GetUserByID(ctx, getUserID)
							Expect(err).ToNot(HaveOccurred())
							Expect(usr).ToNot(BeNil())
							Expect(usr.ProfileID).To(BeNil())
						})
					})
				})

				Context("DeleteUser", func() {
					var deleteUserID string
					var deleteUser *user.User

					BeforeEach(func() {
						deleteUserID = user.NewID()
						deleteUser = NewUser(deleteUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(deleteUser)).To(Succeed())
					})

					It("succeeds if it successfully removes users", func() {
						Expect(mongoSession.DeleteUser(ctx, deleteUser)).To(Succeed())
						Expect(deleteUser.DeletedTime).ToNot(BeEmpty())
						Expect(deleteUser.DeletedUserID).To(BeEmpty())
					})

					It("returns an error if the user is missing", func() {
						Expect(mongoSession.DeleteUser(ctx, nil)).To(MatchError("user is missing"))
					})

					It("returns an error if the context is missing", func() {
						Expect(mongoSession.DeleteUser(nil, deleteUser)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						deleteUser.ID = ""
						Expect(mongoSession.DeleteUser(ctx, deleteUser)).To(MatchError("user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DeleteUser(ctx, deleteUser)).To(MatchError("session closed"))
					})

					It("has the correct stored users", func() {
						ValidateUsers(testMongoCollection, bson.M{}, append(users, deleteUser))
						Expect(mongoSession.DeleteUser(ctx, deleteUser)).To(Succeed())
						ValidateUsers(testMongoCollection, bson.M{}, append(users, deleteUser))
					})
				})

				Context("DestroyUserByID", func() {
					var destroyUserID string
					var destroyUser *user.User

					BeforeEach(func() {
						destroyUserID = user.NewID()
						destroyUser = NewUser(destroyUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyUser)).To(Succeed())
					})

					It("succeeds if it successfully removes users", func() {
						Expect(mongoSession.DestroyUserByID(ctx, destroyUserID)).To(Succeed())
					})

					It("returns an error if the context is missing", func() {
						Expect(mongoSession.DestroyUserByID(nil, destroyUserID)).To(MatchError("context is missing"))
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroyUserByID(ctx, "")).To(MatchError("user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyUserByID(ctx, destroyUserID)).To(MatchError("session closed"))
					})

					It("has the correct stored users", func() {
						ValidateUsers(testMongoCollection, bson.M{}, append(users, destroyUser))
						Expect(mongoSession.DestroyUserByID(ctx, destroyUserID)).To(Succeed())
						ValidateUsers(testMongoCollection, bson.M{}, users)
					})
				})
			})

			Context("PasswordMatches", func() {
				It("returns true if the passwords match", func() {
					usr := &user.User{
						ID:           "0cd1a5d68b",
						PasswordHash: "f4bbfc883178b79c184732c8aaa4e1e72a851ad1",
					}
					Expect(mongoSession.PasswordMatches(usr, "asdflknj237u9fsnkl")).To(BeTrue())
				})

				It("returns false if the passwords do not match", func() {
					usr := &user.User{
						ID:           "d23b0a8786",
						PasswordHash: "e8353f1aa1045a73ddeebd71febafee7d85768d8",
					}
					Expect(mongoSession.PasswordMatches(usr, "invalid-password")).To(BeFalse())
				})
			})

			Context("HashPassword", func() {
				DescribeTable("return correct password for",
					func(userID string, password string, expectedPasswordHash string) {
						Expect(mongoSession.(*userStoreMongo.UsersSession).HashPassword(userID, password)).To(Equal(expectedPasswordHash))
					},
					Entry("is example #1", "0cd1a5d68b", "asdflknj237u9fsnkl", "f4bbfc883178b79c184732c8aaa4e1e72a851ad1"),
					Entry("is example #2", "b52201f96b", "asdflknj237u9fsnkl", "eeeb9f6f8092012db6effb1b57fac0f41ea08156"),
					Entry("is example #3", "46267a83eb", "asdflknj237u9fsnkl", "a01c4b1f969837a5de28465db407d41bcea78d14"),
					Entry("is example #4", "982f600045", "asdflknj237u9fsnkl", "100801f42b3ca682dccf4bde05ee3a23749111a5"),
					Entry("is example #5", "a06176bed7", "asdflknj237u9fsnkl", "6c164ad4ff487ac912d5a71a9cece610bcdf2899"),
					Entry("is example #6", "d23b0a8786", "2938wdefjlr5tyu93", "e8353f1aa1045a73ddeebd71febafee7d85768d8"),
					Entry("is example #7", "a011c16df7", "2938wdefjlr5tyu93", "b5809a275e903a5c5605e49994295e0d208793eb"),
					Entry("is example #8", "8ea2d078f6", "2938wdefjlr5tyu93", "7561e37b64bc84791813038a8a8da176bec42e43"),
					Entry("is example #9", "6128ef12fc", "2938wdefjlr5tyu93", "545a5c015d9b080252c78e0ff5e1722cc266820f"),
					Entry("is example #10", "806d315a0b", "2938wdefjlr5tyu93", "2712f5216d763d55a497173c90482e0b2ed9f7d6"),
				)
			})
		})
	})
})
