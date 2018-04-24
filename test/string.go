package test

import "math/rand"

const (
	CharsetUppercase    = "ABCDEFGHIJKLMNOPQRSTUVWYXZ"
	CharsetLowercase    = "abcdefghijklmnopqrstuvwxyz"
	CharsetNumeric      = "1234567890"
	CharsetWhitespace   = " "
	CharsetSymbols      = "!\"#$%&'()*+,-./:;<=>@\\]^_`{|}~"
	CharsetAlpha        = CharsetUppercase + CharsetLowercase
	CharsetAlphaNumeric = CharsetUppercase + CharsetLowercase + CharsetNumeric
	CharsetText         = CharsetAlphaNumeric + CharsetWhitespace + CharsetSymbols
)

func NewString(length int, charset string) string {
	bytes := make([]byte, length)
	for index := range bytes {
		bytes[index] = charset[rand.Intn(len(charset))]
	}
	return string(bytes)
}

func NewVariableString(minimumLength int, maximumLength int, charset string) string {
	return NewString(minimumLength+rand.Intn(maximumLength-minimumLength+1), charset)
}

func NewText(minimumLength int, maximumLength int) string {
	return NewVariableString(minimumLength, maximumLength, CharsetText)
}
