package test

import (
	"fmt"
	"os"
	"time"

	"github.com/tidepool-org/platform/test"
)

func Address() string {
	return os.Getenv("TIDEPOOL_STORE_ADDRESSES")
}

func Database() string {
	return generateUniqueName("database")
}

func NewCollectionPrefix() string {
	return generateUniqueName("collection_")
}

func generateUniqueName(base string) string {
	return fmt.Sprintf("test_%s_%s_%s", time.Now().Format("20060102150405"), test.RandomStringFromRangeAndCharset(4, 4, test.CharsetNumeric), base)
}
