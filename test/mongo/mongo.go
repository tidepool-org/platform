package mongo

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"fmt"
	"math/rand"
	"time"

	mgo "gopkg.in/mgo.v2"
)

var (
	_globalSession *mgo.Session
	_nodeSession   *mgo.Session
	_database      string
)

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	globalSession, err := mgo.Dial(Address())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(globalSession).ToNot(gomega.BeNil())
	_globalSession = globalSession
	return []byte(generateUniqueName("database"))
}, func(data []byte) {
	nodeSession, err := mgo.Dial(Address())
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(nodeSession).ToNot(gomega.BeNil())
	_nodeSession = nodeSession
	_database = string(data)
})

var _ = ginkgo.SynchronizedAfterSuite(func() {
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

func Address() string {
	return "127.0.0.1"
}

func Session() *mgo.Session {
	return _nodeSession
}

func Database() string {
	return _database
}

func NewCollectionPrefix() string {
	return generateUniqueName("collection_")
}

func generateUniqueName(base string) string {
	return fmt.Sprintf("test_%s_%08x%08x_%s", time.Now().UTC().Format("20060102150405"), rand.Uint32(), rand.Uint32(), base)
}
