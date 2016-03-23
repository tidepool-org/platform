package main_test

import (
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/onsi/gomega"

	"testing"
)

func TestDataservices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataservices Suite")
}
