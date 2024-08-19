package devlog

import (
	"bytes"
	"fmt"
	"log"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	plog "github.com/tidepool-org/platform/log"
)

var _ = Describe("devlog", func() {
	Context("Serialize", func() {
		It("pulls out the time field", func() {
			stamp := time.Now().Truncate(time.Microsecond).Format(time.RFC3339Nano)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}

			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).ToNot(ContainSubstring("time=" + stamp))
		})

		It("displays caller info on error level logs", func() {
			stamp := time.Now().Format(time.Stamp)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"

			fields := plog.Fields{"message": message, "time": stamp}

			fields["caller"] = &errors.Caller{File: "foo.go", Line: 42}
			fields["level"] = string(plog.ErrorLevel)
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring("caller=foo.go:42"))
		})

		It("pulls out the message field", func() {
			stamp := time.Now().Format(time.Stamp)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}

			fields["extra"] = "field"
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring(stamp + " ?? " + message + ":"))
		})

		It("falls back to ?? in the level isn't recognized", func() {
			stamp := time.Now().Format(time.Stamp)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}

			fields["level"] = "some"
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring(stamp + " ?? " + message))
		})

		It("abbreviates the debug level", func() {
			stamp := time.Now().Format(time.Stamp)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}

			fields["level"] = "debug"
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring(stamp + " DD " + message))
		})

		It("handles structs", func() {
			stamp := time.Now().Truncate(time.Microsecond).Format(time.RFC3339Nano)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}

			datum := struct{ foo string }{foo: "bar"}
			fields["datum"] = datum
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("datum=%+v", datum)))
		})

		It("handles ints", func() {
			stamp := time.Now().Truncate(time.Microsecond).Format(time.RFC3339Nano)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}
			value := 42

			fields["int"] = value
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring("int=42"))
		})

		It("handles bools", func() {
			stamp := time.Now().Truncate(time.Microsecond).Format(time.RFC3339Nano)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}

			value := true
			fields["bool"] = value
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring("bool=true"))
		})

		It("handles floats", func() {
			stamp := time.Now().Truncate(time.Microsecond).Format(time.RFC3339Nano)
			buf := &bytes.Buffer{}
			s := &serializer{Logger: log.New(buf, "", 0)}
			message := "this is a test"
			fields := plog.Fields{"message": message, "time": stamp}

			value := 24.2
			fields["float"] = value
			Expect(s.Serialize(fields)).To(Succeed())
			Expect(buf.String()).To(ContainSubstring("float=24.2"))
		})
	})
})
