package errors_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataservices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dataservices/server/api/v1/errors")
}
