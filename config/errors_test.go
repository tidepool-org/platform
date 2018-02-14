package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

var _ = Describe("Errors", func() {
	DescribeTable("all errors",
		func(err error, code string, title string, detail string) {
			Expect(err).ToNot(BeNil())
			Expect(errors.Code(err)).To(Equal(code))
			Expect(errors.Cause(err)).To(Equal(err))
			bytes, bytesErr := json.Marshal(errors.Sanitize(err))
			Expect(bytesErr).ToNot(HaveOccurred())
			Expect(bytes).To(MatchJSON(fmt.Sprintf(`{"code": %q, "title": %q, "detail": %q}`, code, title, detail)))
		},
		Entry("is ErrorKeyNotFound", config.ErrorKeyNotFound("TEST"), "key-not-found", "key not found", "key \"TEST\" not found"),
	)
})
