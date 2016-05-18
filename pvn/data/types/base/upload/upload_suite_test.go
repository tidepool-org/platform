package upload_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUpload(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "data/types/base/upload")
}
