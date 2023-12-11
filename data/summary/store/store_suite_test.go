package store_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSummaryStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/summary/store")
}
