package crypto_test

import (
	"encoding/base64"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/crypto"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	structureTest "github.com/tidepool-org/platform/structure/test"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("Crypto", func() {
	Context("Base64EncodedMD5Hash", func() {
		DescribeTable("returns the expected result when the input",
			func(value string, expectedResult string) {
				Expect(crypto.Base64EncodedMD5Hash([]byte(value))).To(Equal(expectedResult))
			},
			Entry("is empty", "", "1B2M2Y8AsgTpgAmY7PhCfg=="),
			Entry("is not empty", "abcdefghijklmnopqrstuvwxyz", "w/zT12GS5AB9+0lsymfhOw=="),
			Entry("is whitespace", "        ", "e7Dt2Y8iQwoDtn+FPoPCyg=="),
			Entry("includes non-ASCII", "abcABC123 !\"#_😁😂😃ΨΪΫ", "+V8+B6cToORNU71pST2SeQ=="),
		)
	})

	Context("IsValidBase64EncodedMD5Hash, Base64EncodedMD5HashValidator, and ValidateBase64EncodedMD5Hash", func() {
		DescribeTable("return the expected results when the input",
			func(value string, expectedErrors ...error) {
				Expect(crypto.IsValidBase64EncodedMD5Hash(value)).To(Equal(len(expectedErrors) == 0))
				errorReporter := structureTest.NewErrorReporter()
				crypto.Base64EncodedMD5HashValidator(value, errorReporter)
				errorsTest.ExpectEqual(errorReporter.Error(), expectedErrors...)
				errorsTest.ExpectEqual(crypto.ValidateBase64EncodedMD5Hash(value), expectedErrors...)
			},
			Entry("is empty", "", structureValidator.ErrorValueEmpty()),
			Entry("is not valid Base64 encoded", "QUJDREVGSElKS0xNTk9QUQ=$", crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("QUJDREVGSElKS0xNTk9QUQ=$")),
			Entry("is valid Base64 encoded and byte length is out of range (lower)", "QUJDREVGSElKS0xNTk9Q", crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("QUJDREVGSElKS0xNTk9Q")),
			Entry("is valid Base64 encoded and byte length is in range", "QUJDREVGSElKS0xNTk9QUQ=="),
			Entry("is valid Base64 encoded and byte length is out of range (upper)", "QUJDREVGSElKS0xNTk9QUVI=", crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("QUJDREVGSElKS0xNTk9QUVI=")),
		)
	})

	Context("Errors", func() {
		DescribeTable("have expected details when error",
			errorsTest.ExpectErrorDetails,
			Entry("is ErrorValueStringAsBase64EncodedMD5HashNotValid with empty string", crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid(""), "value-not-valid", "value is not valid", `value "" is not valid as Base64 encoded MD5 hash`),
			Entry("is ErrorValueStringAsBase64EncodedMD5HashNotValid with non-empty string", crypto.ErrorValueStringAsBase64EncodedMD5HashNotValid("QUJDREVGSElKS0xNTk9QUQ=="), "value-not-valid", "value is not valid", `value "QUJDREVGSElKS0xNTk9QUQ==" is not valid as Base64 encoded MD5 hash`),
		)
	})

	Context("HexEncodedMD5Hash", func() {
		DescribeTable("returns the expected result when the input",
			func(value string, expectedResult string) {
				Expect(crypto.HexEncodedMD5Hash(value)).To(Equal(expectedResult))
			},
			Entry("is empty", "", "d41d8cd98f00b204e9800998ecf8427e"),
			Entry("is not empty", "abcdefghijklmnopqrstuvwxyz", "c3fcd3d76192e4007dfb496cca67e13b"),
			Entry("is whitespace", "        ", "7bb0edd98f22430a03b67f853e83c2ca"),
			Entry("includes non-ASCII", "abcABC123 !\"#_😁😂😃ΨΪΫ", "f95f3e07a713a0e44d53bd69493d9279"),
		)
	})

	Context("EncryptWithAES256UsingPassphrase", func() {
		It("returns an error if the bytes is missing", func() {
			encrypted, err := crypto.EncryptWithAES256UsingPassphrase(nil, []byte("secret"))
			Expect(err).To(MatchError("bytes is missing"))
			Expect(encrypted).To(BeNil())
		})

		It("returns an error if the bytes is empty", func() {
			encrypted, err := crypto.EncryptWithAES256UsingPassphrase([]byte{}, []byte("secret"))
			Expect(err).To(MatchError("bytes is missing"))
			Expect(encrypted).To(BeNil())
		})

		It("returns an error if the passphrase is missing", func() {
			encrypted, err := crypto.EncryptWithAES256UsingPassphrase([]byte("psZ5wJPUnU7Fqpuinhdz2m"), nil)
			Expect(err).To(MatchError("passphrase is missing"))
			Expect(encrypted).To(BeNil())
		})

		It("returns an error if the passphrase is empty", func() {
			encrypted, err := crypto.EncryptWithAES256UsingPassphrase([]byte("psZ5wJPUnU7Fqpuinhdz2m"), []byte{})
			Expect(err).To(MatchError("passphrase is missing"))
			Expect(encrypted).To(BeNil())
		})

		DescribeTable("is successful for",
			func(source string, passphrase string, expectedEncrypted string) {
				encrypted, err := crypto.EncryptWithAES256UsingPassphrase([]byte(source), []byte(passphrase))
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

	Context("DecryptWithAES256UsingPassphrase", func() {
		It("returns an error if the bytes is missing", func() {
			decrypted, err := crypto.DecryptWithAES256UsingPassphrase(nil, []byte("secret"))
			Expect(err).To(MatchError("bytes is missing"))
			Expect(decrypted).To(BeNil())
		})

		It("returns an error if the bytes is empty", func() {
			decrypted, err := crypto.DecryptWithAES256UsingPassphrase([]byte{}, []byte("secret"))
			Expect(err).To(MatchError("bytes is missing"))
			Expect(decrypted).To(BeNil())
		})

		It("returns an error if the passphrase is missing", func() {
			decrypted, err := crypto.DecryptWithAES256UsingPassphrase([]byte("psZ5wJPUnU7Fqpuinhdz2m"), nil)
			Expect(err).To(MatchError("passphrase is missing"))
			Expect(decrypted).To(BeNil())
		})

		It("returns an error if the passphrase is empty", func() {
			decrypted, err := crypto.DecryptWithAES256UsingPassphrase([]byte("psZ5wJPUnU7Fqpuinhdz2m"), []byte{})
			Expect(err).To(MatchError("passphrase is missing"))
			Expect(decrypted).To(BeNil())
		})

		DescribeTable("is successful for",
			func(source string, passphrase string, expectedDecrypted string) {
				sourceBytes, err := base64.StdEncoding.DecodeString(source)
				Expect(err).ToNot(HaveOccurred())
				decrypted, err := crypto.DecryptWithAES256UsingPassphrase(sourceBytes, []byte(passphrase))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(decrypted)).To(Equal(expectedDecrypted))
			},
			Entry("is example #1", "u9oBxQ7+8o5ohWLUp9RbpDduGW56uNAB8/ZJQcDo6Wx2pXY8snlvrmfyFHndhOem", "GUl4zbpWkd", "ibhjvB2DueXIVuKLV7QJIuHspsdDAsTWQmVyQHd"),
			Entry("is example #2", "Jx0fCMFp2tLSz37SZ7JOJQ==", "4ja4dk5tt5PJiW3zrvqF9NMur", "6JXA4IsvJxPnTO"),
			Entry("is example #3", "ERsnJlVRFZLjXLPhrKXSroYQPyvPZS6VW2KTvm60NUQ=", "k2YPlHpQwJJG4AzHr5U6", "jRSQI20ZFlTWlbB6ayzMY7EERB2h"),
			Entry("is example #4", "93AGPyJuhfBOySAK8MZ0A22aGodVkAcNszs4huUW7L4=", "8SrylkS5rnHzfAYR3Wf6aqJD2s1RF4p6mw", "afLjeGF9XASHASR3ZqFR6JWK8"),
			Entry("is example #5", "NUa40t5BBK+Q7Sz8IZwQCRrRLOh13ngRkRr6sz5/HYQ=", "JKD9MZZNtfBkE0bBXsGErEMEUH", "Z65Cj2eL49HJhuKNXxR7"),
			Entry("is example #6", "850qyryuP20rYoRe8KLknQ==", "rNcgal4yFmpwguudssyXoA8", "itTVH9y8hc35wwxy"),
			Entry("is example #7", "jcl9EsQfUixYsZOD+9/um+QntMJFheRtL2at5O3ThdU3/71R0VxnkFgGcpMNoiqf", "8JS12KIwfxZYIYHc9pleJKyB3ly7aROUqaeTUMP0", "AUw6ATb1VxHkN9G4v0TAeKm8ggxvNI6PfZM"),
			Entry("is example #8", "I54QVe/FTdjeYSX7rizbswyLN3MS8BofiYasnz0WbRo=", "OjhC5DbmzO4bADLFZjRxhqrLy", "gq78ziOpciZ0vmXXlv2"),
			Entry("is example #9", "NnBW8NHUV+8c7dgQJ8+2DA==", "qeEhPsLnA1kdikmxDVhYFPw9KwBK", "V5rkRY4I9s"),
			Entry("is example #10", "Y4NMFF1ZpceZ3/VVZ8/dtVrCqrcNDwuSalLi77S22XfVsZczPPs7nFkdpnYKO1uq", "zSYdWZ5nGz6jhwVn7HiutNHX0u", "MzgB9c7GWhn4fM1pamOenTKw9oF49GKZ5"),
			Entry("is example #11", "zxj30u3XPRfuLJYxQq28rdz5o2J1L4PDjxZMVcq6BRhw7kv09fV7zG7X8qp4FRgP", "Ie6OOqPcGhqgHjKzpW9O0Jiq84n5", "4rDsmjGsPBG3UttmiQO50bn6hCufP5Ij0OF5"),
			Entry("is example #12", "SWaiq3lbT6c8NGX1EJvbTJ7SBTuYYuYaKOQrNdUSaic=", "BBiDka0pnMqOXRlQfZqh8oZWPL4", "BgRdFhIrQfZTJS6fCA79V4gPAVUS"),
			Entry("is example #13", "D9PLoCA2FoJPnjY2XSlEt+U7Jx+DBFgzeexveLDnwVM=", "0DEQJ9YQJYvyCD", "3fccPOpsZ5wJPUnU7Fqpuinhdz2mKP"),
			Entry("is example #14", "FZV8f3mxa1TrNg/VA9fEy7GkpTovULFy2YuLjWiEwZUv+Bhzh1iPuapqu7ZR0xvq", "dBD6TmrHIqBNHmPjVFPFfI0dj", "TjIAqnR543zEHgFork2w7B7obzopAZyO9jw1W"),
			Entry("is example #15", "u0rcxso2Fppyo4BcmfCN+Q==", "syxE2r4re", "Lw1daYX1qrfiU"),
			Entry("is example #16", "fHKATdQ/XujFOBmlykTuW1EAXYq6jccnZ3j7lCb5fqNLW/yoCoVF1TzUiOOtXu7w", "lO4zVshpkou43eV", "tutCcOV9eDrHwhq1tMdEqiqGuRbLJyxZihp1R"),
			Entry("is example #17", "i6LPXHeVEvCuMdCfRy033XTxzxA7QHqmp9HlKus05Z0=", "0zsuYVmPJm2PQtBeT9VSugHgHVdjrW", "YJ5JmQ9mxHPBY7esS"),
			Entry("is example #18", "UhYigkOOGwmTFip4JgZ98WNhv0ws5yCRev+V9UnKNI8=", "Ju0CWtSepkpgumaRnxLkI8Ls", "qPAu8QZNMIdcA1AjMrD81IvD"),
			Entry("is example #19", "LA/XOwXV+5a3tp2MA1Nj7wqg+S4INUl/H1lxs8OSv+k=", "NI5lEQ14D1E95n091UiWhsGtiq5etG19I", "d2Uqx7KRxXRkc51i2oYGI"),
			Entry("is example #20", "c5cXDcgveTMfGbIsG6mwAo+pU/EBc62sYQ+TecRyAE58Ns4c3eqRPGq47R2TVXff", "o6CKPb9yHaAPpp8mxImH4kisdeTJ5sfBWAEXpRA", "yYPbaNOKyMewynTf2w0NUkhpsJbGwE2uIr6BcA"),
		)
	})
})
