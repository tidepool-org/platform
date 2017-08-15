package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log/null"
	baseMongo "github.com/tidepool-org/platform/store/mongo"
	testMongo "github.com/tidepool-org/platform/test/mongo"
	"github.com/tidepool-org/platform/user"
	"github.com/tidepool-org/platform/user/store"
	"github.com/tidepool-org/platform/user/store/mongo"
)

type TestAgent struct {
	TestIsServer bool
	TestUserID   string
}

func (t *TestAgent) IsServer() bool {
	return t.TestIsServer
}

func (t *TestAgent) UserID() string {
	return t.TestUserID
}

func NewUser(userID string) *user.User {
	email := fmt.Sprintf("%s@test.org", userID)
	return &user.User{
		ID:                userID,
		Email:             email,
		Emails:            []string{email},
		Roles:             []string{user.ClinicRole},
		TermsAcceptedTime: time.Now().UTC().Format(time.RFC3339),
		EmailVerified:     true,
		PasswordHash:      "1234567890",
		Hash:              id.New(),
		Private:           map[string]*user.IDHash{"meta": {ID: "meta-id", Hash: "meta-hash"}},
		CreatedTime:       time.Now().UTC().Format(time.RFC3339),
	}
}

func NewUsers() []interface{} {
	users := []interface{}{}
	users = append(users, NewUser(id.New()), NewUser(id.New()), NewUser(id.New()))
	return users
}

func ValidateUsers(testMongoCollection *mgo.Collection, selector bson.M, expectedUsers []interface{}) {
	var actualUsers []*user.User
	Expect(testMongoCollection.Find(selector).All(&actualUsers)).To(Succeed())
	Expect(actualUsers).To(ConsistOf(expectedUsers...))
}

var _ = Describe("Mongo", func() {
	var mongoConfig *mongo.Config
	var mongoStore *mongo.Store
	var mongoSession store.Session

	BeforeEach(func() {
		mongoConfig = &mongo.Config{
			Config: &baseMongo.Config{
				Addresses:  []string{testMongo.Address()},
				Database:   testMongo.Database(),
				Collection: testMongo.NewCollectionName(),
				Timeout:    5 * time.Second,
			},
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
		It("returns an error if logger is missing", func() {
			var err error
			mongoStore, err = mongo.New(nil, mongoConfig)
			Expect(err).To(MatchError("mongo: logger is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is missing", func() {
			var err error
			mongoConfig.Config = nil
			mongoStore, err = mongo.New(null.NewLogger(), mongoConfig)
			Expect(err).To(MatchError("mongo: config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if base config is invalid", func() {
			var err error
			mongoConfig.Config.Addresses = nil
			mongoStore, err = mongo.New(null.NewLogger(), mongoConfig)
			Expect(err).To(MatchError("mongo: config is invalid; mongo: addresses is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if config is missing", func() {
			var err error
			mongoStore, err = mongo.New(null.NewLogger(), nil)
			Expect(err).To(MatchError("mongo: config is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns an error if config is invalid", func() {
			var err error
			mongoConfig.PasswordSalt = ""
			mongoStore, err = mongo.New(null.NewLogger(), mongoConfig)
			Expect(err).To(MatchError("mongo: config is invalid; mongo: password salt is missing"))
			Expect(mongoStore).To(BeNil())
		})

		It("returns a new store and no error if successful", func() {
			var err error
			mongoStore, err = mongo.New(null.NewLogger(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		BeforeEach(func() {
			var err error
			mongoStore, err = mongo.New(null.NewLogger(), mongoConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(mongoStore).ToNot(BeNil())
		})

		Context("NewSession", func() {
			It("returns a new session if no logger specified", func() {
				mongoSession = mongoStore.NewSession(nil)
				Expect(mongoSession).ToNot(BeNil())
				Expect(mongoSession.Logger()).ToNot(BeNil())
			})

			It("returns a new session if logger specified", func() {
				logger := null.NewLogger()
				mongoSession = mongoStore.NewSession(logger)
				Expect(mongoSession).ToNot(BeNil())
				Expect(mongoSession.Logger()).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			BeforeEach(func() {
				mongoSession = mongoStore.NewSession(null.NewLogger())
				Expect(mongoSession).ToNot(BeNil())
			})

			Context("with persisted data", func() {
				var testMongoSession *mgo.Session
				var testMongoCollection *mgo.Collection
				var users []interface{}

				BeforeEach(func() {
					testMongoSession = testMongo.Session().Copy()
					testMongoCollection = testMongoSession.DB(mongoConfig.Database).C(mongoConfig.Collection)
					users = NewUsers()
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
						getUserID = id.New()
						getUser = NewUser(getUserID)
						getUserEmail = fmt.Sprintf("%s@test.org", getUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(getUser)).To(Succeed())
					})

					It("succeeds if it successfully gets the user", func() {
						usr, err := mongoSession.GetUserByID(getUserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(usr).ToNot(BeNil())
						Expect(usr.ID).To(Equal(getUserID))
						Expect(usr.Email).To(Equal(getUserEmail))
						Expect(usr.ProfileID).ToNot(BeNil())
						Expect(*usr.ProfileID).To(Equal("meta-id"))
					})

					It("succeeds even if two users exist with the same user id", func() {
						Expect(testMongoCollection.Insert(getUser)).To(Succeed())
						usr, err := mongoSession.GetUserByID(getUserID)
						Expect(err).ToNot(HaveOccurred())
						Expect(usr).ToNot(BeNil())
						Expect(usr.ID).To(Equal(getUserID))
					})

					It("returns no error and no user if the user id is not found", func() {
						usr, err := mongoSession.GetUserByID(id.New())
						Expect(err).ToNot(HaveOccurred())
						Expect(usr).To(BeNil())
					})

					It("returns an error if the user id is missing", func() {
						usr, err := mongoSession.GetUserByID("")
						Expect(err).To(MatchError("mongo: user id is missing"))
						Expect(usr).To(BeNil())
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						usr, err := mongoSession.GetUserByID(getUserID)
						Expect(err).To(MatchError("mongo: session closed"))
						Expect(usr).To(BeNil())
					})

					Context("with no private", func() {
						BeforeEach(func() {
							getUser.Private = nil
						})

						It("succeeds, but does not fill in the profile id", func() {
							usr, err := mongoSession.GetUserByID(getUserID)
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
							usr, err := mongoSession.GetUserByID(getUserID)
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
							usr, err := mongoSession.GetUserByID(getUserID)
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
						deleteUserID = id.New()
						deleteUser = NewUser(deleteUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(deleteUser)).To(Succeed())
					})

					It("succeeds if it successfully removes users", func() {
						Expect(mongoSession.DeleteUser(deleteUser)).To(Succeed())
						Expect(deleteUser.DeletedTime).ToNot(BeEmpty())
						Expect(deleteUser.DeletedUserID).To(BeEmpty())
					})

					It("returns an error if the user is missing", func() {
						Expect(mongoSession.DeleteUser(nil)).To(MatchError("mongo: user is missing"))
					})

					It("returns an error if the user id is missing", func() {
						deleteUser.ID = ""
						Expect(mongoSession.DeleteUser(deleteUser)).To(MatchError("mongo: user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DeleteUser(deleteUser)).To(MatchError("mongo: session closed"))
					})

					It("has the correct stored users", func() {
						ValidateUsers(testMongoCollection, bson.M{}, append(users, deleteUser))
						Expect(mongoSession.DeleteUser(deleteUser)).To(Succeed())
						ValidateUsers(testMongoCollection, bson.M{}, append(users, deleteUser))
					})

					Context("with agent", func() {
						var agentUserID string

						BeforeEach(func() {
							agentUserID = id.New()
							mongoSession.SetAgent(&TestAgent{false, agentUserID})
						})

						It("succeeds if it successfully removes users", func() {
							Expect(mongoSession.DeleteUser(deleteUser)).To(Succeed())
							Expect(deleteUser.DeletedTime).ToNot(BeEmpty())
							Expect(deleteUser.DeletedUserID).To(Equal(agentUserID))
						})
					})
				})

				Context("DestroyUserByID", func() {
					var destroyUserID string
					var destroyUser *user.User

					BeforeEach(func() {
						destroyUserID = id.New()
						destroyUser = NewUser(destroyUserID)
					})

					JustBeforeEach(func() {
						Expect(testMongoCollection.Insert(destroyUser)).To(Succeed())
					})

					It("succeeds if it successfully removes users", func() {
						Expect(mongoSession.DestroyUserByID(destroyUserID)).To(Succeed())
					})

					It("returns an error if the user id is missing", func() {
						Expect(mongoSession.DestroyUserByID("")).To(MatchError("mongo: user id is missing"))
					})

					It("returns an error if the session is closed", func() {
						mongoSession.Close()
						Expect(mongoSession.DestroyUserByID(destroyUserID)).To(MatchError("mongo: session closed"))
					})

					It("has the correct stored users", func() {
						ValidateUsers(testMongoCollection, bson.M{}, append(users, destroyUser))
						Expect(mongoSession.DestroyUserByID(destroyUserID)).To(Succeed())
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
						Expect(mongoSession.(*mongo.Session).HashPassword(userID, password)).To(Equal(expectedPasswordHash))
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
