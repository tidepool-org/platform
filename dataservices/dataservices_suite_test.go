package dataservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataservices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataservices Suite")
}
