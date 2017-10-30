package env_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"syscall"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/config/env"
)

var _ = Describe("Reporter", func() {
	Context("NewConfig", func() {
		It("returns an error if prefix is missing", func() {
			reporter, err := env.NewReporter("")
			Expect(err).To(MatchError("prefix is missing"))
			Expect(reporter).To(BeNil())
		})

		DescribeTable("returns an error if prefix is invalid",
			func(prefix string) {
				reporter, err := env.NewReporter(prefix)
				Expect(err).To(MatchError("prefix is invalid"))
				Expect(reporter).To(BeNil())
			},
			Entry("is underscore", "_"),
			Entry("starts with underscore", "_TEST"),
			Entry("is number", "0"),
			Entry("starts with number", "0TEST"),
			Entry("is lowercase alpha", "a"),
			Entry("starts with lowercase alpha", "aTEST"),
			Entry("contains lowercase alpha", "TESTaTEST"),
			Entry("is non-alphanumeric", "."),
			Entry("starts with non-alphanumeric", ".TEST"),
			Entry("contains non-alphanumeric", "TEST.TEST"),
		)

		DescribeTable("returns a new config if prefix is valid",
			func(prefix string) {
				Expect(env.NewReporter(prefix)).ToNot(BeNil())
			},
			Entry("is uppercase alpha", "T"),
			Entry("starts with uppercase alpha", "TEST"),
			Entry("includes underscore", "TEST_TEST"),
			Entry("ends with underscore", "TEST_"),
			Entry("includes number", "TEST0TEST"),
			Entry("ends with number", "TEST0"),
		)
	})

	Context("with new config", func() {
		var reporter config.Reporter

		BeforeEach(func() {
			var err error
			reporter, err = env.NewReporter("TIDEPOOL_TEST")
			Expect(err).ToNot(HaveOccurred())
			Expect(reporter).ToNot(BeNil())
		})

		Context("String", func() {
			DescribeTable("returns expected values given environment variables",
				func(environmentKey string, environmentValue string, key string, expectedValue string, expectedFound bool) {
					Expect(syscall.Setenv(environmentKey, environmentValue)).To(Succeed())
					actualValue, actualFound := reporter.Get(key)
					Expect(syscall.Unsetenv(environmentKey)).To(Succeed())
					Expect(actualFound).To(Equal(expectedFound))
					Expect(actualValue).To(Equal(expectedValue))
				},
				Entry("joins parts with underscore", "TIDEPOOL_TEST_ALPHA", "dog", "ALPHA", "dog", true),
				Entry("converts to uppercase", "TIDEPOOL_TEST_BETA", "tester", "beta", "tester", true),
				Entry("replaces invalid characters with underscores", "TIDEPOOL_TEST_C_H_A_R_L_I_E", "brown", "C*H&A'R.L\"I?E", "brown", true),
				Entry("allows underscores", "TIDEPOOL_TEST_DEL_TA", "force", "DEL_TA", "force", true),
				Entry("allows empty value", "TIDEPOOL_TEST_ECHO", "", "ECHO", "", true),
				Entry("without match", "TIDEPOOL_TEST_FOXTROT", "dance", "NOTFOXTROT", "", false),
			)
		})

		Context("GetWithDefault", func() {
			It("returns the value if found", func() {
				Expect(syscall.Setenv("TIDEPOOL_TEST_GOLF", "bag")).To(Succeed())
				Expect(reporter.GetWithDefault("GOLF", "tee")).To(Equal("bag"))
				Expect(syscall.Unsetenv("TIDEPOOL_TEST_GOLF")).To(Succeed())
			})

			It("returns the value if found, even if empty", func() {
				Expect(syscall.Setenv("TIDEPOOL_TEST_HOTEL", "")).To(Succeed())
				Expect(reporter.GetWithDefault("HOTEL", "room")).To(Equal(""))
				Expect(syscall.Unsetenv("TIDEPOOL_TEST_HOTEL")).To(Succeed())
			})

			It("returns the default valuye if not found", func() {
				Expect(reporter.GetWithDefault("INDIA", "ink")).To(Equal("ink"))
			})
		})

		Context("Set", func() {
			It("sets the key with the value", func() {
				Expect(syscall.Unsetenv("TIDEPOOL_TEST_JULIETTE")).To(Succeed())
				reporter.Set("JULIETTE", "romeo")
				value, found := syscall.Getenv("TIDEPOOL_TEST_JULIETTE")
				Expect(found).To(BeTrue())
				Expect(value).To(Equal("romeo"))
				Expect(syscall.Unsetenv("TIDEPOOL_TEST_JULIETTE")).To(Succeed())
			})
		})

		Context("Delete", func() {
			It("deletes the key", func() {
				Expect(syscall.Setenv("TIDEPOOL_TEST_KILO", "meter")).To(Succeed())
				reporter.Delete("KILO")
				value, found := syscall.Getenv("TIDEPOOL_TEST_KILO")
				Expect(found).To(BeFalse())
				Expect(value).ToNot(Equal("meter"))
			})
		})

		Context("WithScopes", func() {
			DescribeTable("returns expected values given environment variables and scopes",
				func(environmentKey string, environmentValue string, scopes []string, key string, expectedValue string, expectedFound bool) {
					Expect(syscall.Setenv(environmentKey, environmentValue)).To(Succeed())
					actualValue, actualFound := reporter.WithScopes(scopes...).Get(key)
					Expect(syscall.Unsetenv(environmentKey)).To(Succeed())
					Expect(actualFound).To(Equal(expectedFound))
					Expect(actualValue).To(Equal(expectedValue))
				},
				Entry("joined exactly", "TIDEPOOL_TEST_ONE_TWO_THREE_EH", "AAA", []string{"ONE", "TWO", "THREE"}, "EH", "AAA", true),
				Entry("removes one scope", "TIDEPOOL_TEST_TWO_THREE_BEE", "BBB", []string{"ONE", "TWO", "THREE"}, "BEE", "BBB", true),
				Entry("removes two scopes", "TIDEPOOL_TEST_THREE_SEA", "CCC", []string{"ONE", "TWO", "THREE"}, "SEA", "CCC", true),
				Entry("does not return last scope", "TIDEPOOL_TEST_DEE", "DDD", []string{"ONE", "TWO", "THREE"}, "DEE", "", false),
				Entry("allows no scopes", "TIDEPOOL_TEST_EFF", "FFF", []string{}, "EFF", "FFF", true),
			)
		})
	})

	Context("GetKey", func() {
		DescribeTable("returns expected values given prefix, scopes, and key",
			func(prefix string, scopes []string, key string, expectedValue string) {
				Expect(env.GetKey(prefix, scopes, key)).To(Equal(expectedValue))
			},
			Entry("is as expected", "PREFIX", []string{"ONE", "TWO", "THREE"}, "KEY", "PREFIX_ONE_TWO_THREE_KEY"),
			Entry("is lowercase", "PREFIX", []string{"one", "two", "three"}, "key", "PREFIX_ONE_TWO_THREE_KEY"),
			Entry("has no prefix", "", []string{"one", "two", "three"}, "key", "_ONE_TWO_THREE_KEY"),
			Entry("has no scopes", "prefix", nil, "key", "PREFIX_KEY"),
			Entry("has empty scopes", "prefix", []string{}, "key", "PREFIX_KEY"),
			Entry("has an empty scope", "prefix", []string{""}, "key", "PREFIX__KEY"),
			Entry("has no key", "prefix", []string{"one", "two", "three"}, "", "PREFIX_ONE_TWO_THREE_"),
			Entry("has invalid characters", "pr!!ix", []string{"o$e", "t^o", "th*ee"}, "k~y", "PR__IX_O_E_T_O_TH_EE_K_Y"),
		)
	})
})
