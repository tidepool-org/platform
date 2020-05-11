package test

import (
	"fmt"
	"os"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

var (
	globalSession *mgo.Session
	nodeSession   *mgo.Session
	database      string
)

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	session, err := mgo.Dial(Address())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(session).ToNot(gomega.BeNil())
	globalSession = session
	return []byte(generateUniqueName("database"))
}, func(data []byte) {
	session, err := mgo.Dial(Address())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(session).ToNot(gomega.BeNil())
	nodeSession = session
	database = string(data)
})

var _ = ginkgo.SynchronizedAfterSuite(func() {
	if nodeSession != nil {
		nodeSession.Close()
		nodeSession = nil
	}
}, func() {
	if globalSession != nil {
		globalSession.DB(database).DropDatabase()
		globalSession.Close()
		globalSession = nil
	}
})

func Address() string {
	return os.Getenv("TIDEPOOL_STORE_ADDRESSES")
}

func Session() *mgo.Session {
	return nodeSession
}

func Database() string {
	return database
}

func NewCollectionPrefix() string {
	return generateUniqueName("collection_")
}

func generateUniqueName(base string) string {
	return fmt.Sprintf("test_%s_%s_%s", time.Now().Format("20060102150405"), test.RandomStringFromRangeAndCharset(16, 16, test.CharsetNumeric), base)
}
