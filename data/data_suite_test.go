package data_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Data Suite")
}
