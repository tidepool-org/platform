package test

import "math/rand"

const (
	CharsetUppercase            = "ABCDEFGHIJKLMNOPQRSTUVWYXZ"
	CharsetLowercase            = "abcdefghijklmnopqrstuvwxyz"
	CharsetNumeric              = "1234567890"
	CharsetWhitespace           = " "
	CharsetSymbols              = "!\"#$%&'()*+,-./:;<=>@\\]^_`{|}~"
	CharsetAlpha                = CharsetUppercase + CharsetLowercase
	CharsetAlphaNumeric         = CharsetUppercase + CharsetLowercase + CharsetNumeric
	CharsetText                 = CharsetAlphaNumeric + CharsetWhitespace + CharsetSymbols
	CharsetHexidecimalLowercase = CharsetNumeric + "abcdef"
)

func NewString(length int, charset string) string {
	bites := make([]byte, length)
	for index := range bites {
		bites[index] = charset[rand.Intn(len(charset))]
	}
	return string(bites)
}

func NewVariableString(minimumLength int, maximumLength int, charset string) string {
	return NewString(minimumLength+rand.Intn(maximumLength-minimumLength+1), charset)
}

func NewText(minimumLength int, maximumLength int) string {
	return NewVariableString(minimumLength, maximumLength, CharsetText)
}
