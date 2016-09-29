package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"encoding/base64"

	"github.com/tidepool-org/platform/app"
)

var _ = Describe("Crypto", func() {
	Context("EncryptWithAES256UsingPassphrase", func() {
		It("returns an error if the bytes is missing", func() {
			encrypted, err := app.EncryptWithAES256UsingPassphrase(nil, []byte("secret"))
			Expect(err).To(MatchError("app: bytes is missing"))
			Expect(encrypted).To(BeNil())
		})

		It("returns an error if the bytes is empty", func() {
			encrypted, err := app.EncryptWithAES256UsingPassphrase([]byte{}, []byte("secret"))
			Expect(err).To(MatchError("app: bytes is missing"))
			Expect(encrypted).To(BeNil())
		})

		It("returns an error if the passphrase is missing", func() {
			encrypted, err := app.EncryptWithAES256UsingPassphrase([]byte("psZ5wJPUnU7Fqpuinhdz2m"), nil)
			Expect(err).To(MatchError("app: passphrase is missing"))
			Expect(encrypted).To(BeNil())
		})

		It("returns an error if the passphrase is empty", func() {
			encrypted, err := app.EncryptWithAES256UsingPassphrase([]byte("psZ5wJPUnU7Fqpuinhdz2m"), []byte{})
			Expect(err).To(MatchError("app: passphrase is missing"))
			Expect(encrypted).To(BeNil())
		})

		DescribeTable("is successful for",
			func(source string, passphrase string, expectedEncrypted string) {
				encrypted, err := app.EncryptWithAES256UsingPassphrase([]byte(source), []byte(passphrase))
				Expect(err).ToNot(HaveOccurred())
				Expect(base64.StdEncoding.EncodeToString(encrypted)).To(Equal(expectedEncrypted))
			},
			Entry("is example #1", "ibhjvB2DueXIVuKLV7QJIuHspsdDAsTWQmVyQHd", "GUl4zbpWkd", "u9oBxQ7+8o5ohWLUp9RbpDduGW56uNAB8/ZJQcDo6Wx2pXY8snlvrmfyFHndhOem"),
			Entry("is example #2", "6JXA4IsvJxPnTO", "4ja4dk5tt5PJiW3zrvqF9NMur", "Jx0fCMFp2tLSz37SZ7JOJQ=="),
			Entry("is example #3", "jRSQI20ZFlTWlbB6ayzMY7EERB2h", "k2YPlHpQwJJG4AzHr5U6", "ERsnJlVRFZLjXLPhrKXSroYQPyvPZS6VW2KTvm60NUQ="),
			Entry("is example #4", "afLjeGF9XASHASR3ZqFR6JWK8", "8SrylkS5rnHzfAYR3Wf6aqJD2s1RF4p6mw", "93AGPyJuhfBOySAK8MZ0A22aGodVkAcNszs4huUW7L4="),
			Entry("is example #5", "Z65Cj2eL49HJhuKNXxR7", "JKD9MZZNtfBkE0bBXsGErEMEUH", "NUa40t5BBK+Q7Sz8IZwQCRrRLOh13ngRkRr6sz5/HYQ="),
			Entry("is example #6", "itTVH9y8hc35wwxy", "rNcgal4yFmpwguudssyXoA8", "850qyryuP20rYoRe8KLknQ=="),
			Entry("is example #7", "AUw6ATb1VxHkN9G4v0TAeKm8ggxvNI6PfZM", "8JS12KIwfxZYIYHc9pleJKyB3ly7aROUqaeTUMP0", "jcl9EsQfUixYsZOD+9/um+QntMJFheRtL2at5O3ThdU3/71R0VxnkFgGcpMNoiqf"),
			Entry("is example #8", "gq78ziOpciZ0vmXXlv2", "OjhC5DbmzO4bADLFZjRxhqrLy", "I54QVe/FTdjeYSX7rizbswyLN3MS8BofiYasnz0WbRo="),
			Entry("is example #9", "V5rkRY4I9s", "qeEhPsLnA1kdikmxDVhYFPw9KwBK", "NnBW8NHUV+8c7dgQJ8+2DA=="),
			Entry("is example #10", "MzgB9c7GWhn4fM1pamOenTKw9oF49GKZ5", "zSYdWZ5nGz6jhwVn7HiutNHX0u", "Y4NMFF1ZpceZ3/VVZ8/dtVrCqrcNDwuSalLi77S22XfVsZczPPs7nFkdpnYKO1uq"),
			Entry("is example #11", "4rDsmjGsPBG3UttmiQO50bn6hCufP5Ij0OF5", "Ie6OOqPcGhqgHjKzpW9O0Jiq84n5", "zxj30u3XPRfuLJYxQq28rdz5o2J1L4PDjxZMVcq6BRhw7kv09fV7zG7X8qp4FRgP"),
			Entry("is example #12", "BgRdFhIrQfZTJS6fCA79V4gPAVUS", "BBiDka0pnMqOXRlQfZqh8oZWPL4", "SWaiq3lbT6c8NGX1EJvbTJ7SBTuYYuYaKOQrNdUSaic="),
			Entry("is example #13", "3fccPOpsZ5wJPUnU7Fqpuinhdz2mKP", "0DEQJ9YQJYvyCD", "D9PLoCA2FoJPnjY2XSlEt+U7Jx+DBFgzeexveLDnwVM="),
			Entry("is example #14", "TjIAqnR543zEHgFork2w7B7obzopAZyO9jw1W", "dBD6TmrHIqBNHmPjVFPFfI0dj", "FZV8f3mxa1TrNg/VA9fEy7GkpTovULFy2YuLjWiEwZUv+Bhzh1iPuapqu7ZR0xvq"),
			Entry("is example #15", "Lw1daYX1qrfiU", "syxE2r4re", "u0rcxso2Fppyo4BcmfCN+Q=="),
			Entry("is example #16", "tutCcOV9eDrHwhq1tMdEqiqGuRbLJyxZihp1R", "lO4zVshpkou43eV", "fHKATdQ/XujFOBmlykTuW1EAXYq6jccnZ3j7lCb5fqNLW/yoCoVF1TzUiOOtXu7w"),
			Entry("is example #17", "YJ5JmQ9mxHPBY7esS", "0zsuYVmPJm2PQtBeT9VSugHgHVdjrW", "i6LPXHeVEvCuMdCfRy033XTxzxA7QHqmp9HlKus05Z0="),
			Entry("is example #18", "qPAu8QZNMIdcA1AjMrD81IvD", "Ju0CWtSepkpgumaRnxLkI8Ls", "UhYigkOOGwmTFip4JgZ98WNhv0ws5yCRev+V9UnKNI8="),
			Entry("is example #19", "d2Uqx7KRxXRkc51i2oYGI", "NI5lEQ14D1E95n091UiWhsGtiq5etG19I", "LA/XOwXV+5a3tp2MA1Nj7wqg+S4INUl/H1lxs8OSv+k="),
			Entry("is example #20", "yYPbaNOKyMewynTf2w0NUkhpsJbGwE2uIr6BcA", "o6CKPb9yHaAPpp8mxImH4kisdeTJ5sfBWAEXpRA", "c5cXDcgveTMfGbIsG6mwAo+pU/EBc62sYQ+TecRyAE58Ns4c3eqRPGq47R2TVXff"),
		)
	})
})
