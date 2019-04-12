package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/test"
	userService "github.com/tidepool-org/platform/user/service"
)

var _ = Describe("PasswordHasher", func() {
	Context("PasswordHasherConfig", func() {
		Context("NewPasswordHasherConfig", func() {
			It("returns successfully with default values", func() {
				Expect(userService.NewPasswordHasherConfig()).To(Equal(&userService.PasswordHasherConfig{}))
			})
		})

		Context("with new password hasher config", func() {
			var salt string
			var passwordHasherConfig *userService.PasswordHasherConfig

			BeforeEach(func() {
				salt = test.RandomString()
				passwordHasherConfig = userService.NewPasswordHasherConfig()
				Expect(passwordHasherConfig).ToNot(BeNil())
			})

			Context("Load", func() {
				var configReporter *configTest.Reporter

				BeforeEach(func() {
					configReporter = configTest.NewReporter()
					configReporter.Config["salt"] = salt
				})

				It("returns an error if config reporter is missing", func() {
					errorsTest.ExpectEqual(passwordHasherConfig.Load(nil), errors.New("config reporter is missing"))
				})

				It("uses existing address if not set", func() {
					existingSalt := test.RandomString()
					passwordHasherConfig.Salt = existingSalt
					delete(configReporter.Config, "salt")
					Expect(passwordHasherConfig.Load(configReporter)).To(Succeed())
					Expect(passwordHasherConfig.Salt).To(Equal(existingSalt))
				})

				It("returns successfully and uses values from config reporter", func() {
					Expect(passwordHasherConfig.Load(configReporter)).To(Succeed())
					Expect(passwordHasherConfig.Salt).To(Equal(salt))
				})
			})

			Context("Validate", func() {
				BeforeEach(func() {
					passwordHasherConfig.Salt = salt
				})

				It("returns an error if the salt is missing", func() {
					passwordHasherConfig.Salt = ""
					errorsTest.ExpectEqual(passwordHasherConfig.Validate(), errors.New("salt is missing"))
				})

				It("returns success", func() {
					Expect(passwordHasherConfig.Validate()).To(Succeed())
					Expect(passwordHasherConfig.Salt).To(Equal(salt))
				})
			})
		})
	})

	Context("PasswordHasher", func() {
		var passwordHasherConfig *userService.PasswordHasherConfig

		BeforeEach(func() {
			passwordHasherConfig = userService.NewPasswordHasherConfig()
			passwordHasherConfig.Salt = test.RandomString()
		})

		Context("NewPasswordHasher", func() {
			It("returns an error if the config is missing", func() {
				passwordHasherConfig = nil
				passwordHasher, err := userService.NewPasswordHasher(passwordHasherConfig)
				errorsTest.ExpectEqual(err, errors.New("config is missing"))
				Expect(passwordHasher).To(BeNil())
			})

			It("returns an error if the config is invalid", func() {
				passwordHasherConfig.Salt = ""
				passwordHasher, err := userService.NewPasswordHasher(passwordHasherConfig)
				errorsTest.ExpectEqual(err, errors.New("config is invalid"))
				Expect(passwordHasher).To(BeNil())
			})

			It("returns successfully", func() {
				Expect(userService.NewPasswordHasher(passwordHasherConfig)).ToNot(BeNil())
			})
		})

		Context("with new password hasher", func() {
			var passwordHasher *userService.PasswordHasher

			BeforeEach(func() {
				passwordHasherConfig.Salt = "06rYrtzhwuSAd7heDZ1tHBxFq7Pysq4N"
				var err error
				passwordHasher, err = userService.NewPasswordHasher(passwordHasherConfig)
				Expect(err).ToNot(HaveOccurred())
				Expect(passwordHasher).ToNot(BeNil())
			})

			Context("HashPassword", func() {
				DescribeTable("hashes the password as expected",
					func(userID string, password string, expectedPasswordHash string) {
						Expect(passwordHasher.HashPassword(userID, password)).To(Equal(expectedPasswordHash))
					},
					Entry("succeeds", "4cc1fcabc9", "password", "ca4f66f8f4e0838327d03eb36904a19b3838847c"),
					Entry("succeeds", "8ffde4b919", "LцH7$lä\"qZ;ñ", "b24d8401a9392f46fc7e228508dcec5a7e7f18d0"),
					Entry("succeeds", "c9fdd05d6f", "🤣รก{לいにbX🐥(YW{ςc", "8a7548a17fc4a1e8aa7db9c163a00304f907ed06"),
					Entry("succeeds", "ec4fdc46db", "F]>!ロиÉfL👻*บς💪obÿ]s{", "90c408148db887a76b2c02acd011a72f7555784c"),
					Entry("succeeds", "faae8e1f1e", "ト(+c*ε(い", "7b3018cdc5600b10b8c86c35c42e442241c8d5dd"),
				)
			})
		})
	})
})
