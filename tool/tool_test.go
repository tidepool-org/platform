package tool_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/application/version"
	_ "github.com/tidepool-org/platform/application/version/test"
	"github.com/tidepool-org/platform/tool"
)

var _ = Describe("Tool", func() {
	Context("New", func() {
		It("returns an error if the prefix is missing", func() {
			app, err := tool.New("")
			Expect(err).To(MatchError("application: prefix is missing"))
			Expect(app).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(tool.New("TIDEPOOL")).ToNot(BeNil())
		})
	})

	Context("with new tool", func() {
		var tuel *tool.Tool

		BeforeEach(func() {
			var err error
			tuel, err = tool.New("TIDEPOOL")
			Expect(err).ToNot(HaveOccurred())
			Expect(tuel).ToNot(BeNil())
		})

		Context("Initialize", func() {
			Context("with incorrectly specified version", func() {
				var versionBase string

				BeforeEach(func() {
					versionBase = version.Base
					version.Base = ""
				})

				AfterEach(func() {
					version.Base = versionBase
				})

				It("returns an error if the version is not specified correctly", func() {
					Expect(tuel.Initialize()).To(MatchError("application: unable to create version reporter; version: base is missing"))
				})
			})

			It("returns successfully", func() {
				Expect(tuel.Initialize()).To(Succeed())
			})
		})

		Context("Terminate", func() {
			It("returns without panic", func() {
				tuel.Terminate()
			})
		})

		Context("initialized", func() {
			BeforeEach(func() {
				Expect(tuel.Initialize()).To(Succeed())
			})

			AfterEach(func() {
				tuel.Terminate()
			})

			Context("CLI", func() {
				It("returns not nil", func() {
					Expect(tuel.CLI()).ToNot(BeNil())
				})
			})

			Context("Args", func() {
				It("returns nil", func() {
					Expect(tuel.Args()).To(BeNil())
				})
			})
		})
	})
})
