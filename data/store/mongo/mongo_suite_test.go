package mongo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"testing"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/tidepool-org/platform/app"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/store/mongo")
}

func MongoTestAddress() string {
	return "127.0.0.1"
}

func MongoTestSession() *mgo.Session {
	return _nodeSession
}

func MongoTestDatabase() string {
	return _database
}

var (
	_globalSession *mgo.Session
	_nodeSession   *mgo.Session
	_database      string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	globalSession, err := mgo.Dial(MongoTestAddress())
	Expect(err).ToNot(HaveOccurred())
	Expect(globalSession).ToNot(BeNil())
	_globalSession = globalSession
	_database = NewTestSuiteID()
	return []byte(_database)
}, func(data []byte) {
	nodeSession, err := mgo.Dial(MongoTestAddress())
	Expect(err).ToNot(HaveOccurred())
	Expect(nodeSession).ToNot(BeNil())
	_nodeSession = nodeSession
	_database = string(data)
})

var _ = SynchronizedAfterSuite(func() {
	if _nodeSession != nil {
		_nodeSession.Close()
		_nodeSession = nil
	}
}, func() {
	if _globalSession != nil {
		_globalSession.DB(_database).DropDatabase()
		_globalSession.Close()
		_globalSession = nil
	}
})

func NewTestSuiteID() string {
	return fmt.Sprintf("platform-test-%s-%s", time.Now().UTC().Format("20060102150405"), app.NewID()[0:20])
}
