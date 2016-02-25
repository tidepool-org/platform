package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDataservices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataservices Suite")
}
