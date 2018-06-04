package test

import (
	"fmt"
	"os"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	mgo "gopkg.in/mgo.v2"

	"github.com/tidepool-org/platform/test"
)

var (
	globalSession *mgo.Session
	nodeSession   *mgo.Session
	database      string
)

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	ssn, err := mgo.Dial(Address())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(ssn).ToNot(gomega.BeNil())
	globalSession = ssn
	return []byte(generateUniqueName("database"))
}, func(data []byte) {
	ssn, err := mgo.Dial(Address())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(ssn).ToNot(gomega.BeNil())
	nodeSession = ssn
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
	return fmt.Sprintf("test_%s_%s_%s", time.Now().Format("20060102150405"), test.NewString(16, test.CharsetNumeric), base)
}
