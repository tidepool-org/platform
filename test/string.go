package test

import (
	"math/rand"

	"github.com/onsi/gomega"
)

const (
	CharsetUppercase    = "ABCDEFGHIJKLMNOPQRSTUVWYXZ"
	CharsetLowercase    = "abcdefghijklmnopqrstuvwxyz"
	CharsetNumeric      = "1234567890"
	CharsetWhitespace   = " "
	CharsetSymbols      = "!\"#$%&'()*+,-./:;<=>@\\]^_`{|}~"
	CharsetAlphaNumeric = CharsetUppercase + CharsetLowercase + CharsetNumeric
	CharsetText         = CharsetAlphaNumeric + CharsetWhitespace + CharsetSymbols
)

func NewString(length int, charset string) string {
	gomega.Expect(charset).ToNot(gomega.BeEmpty())
	bytes := make([]byte, length)
	for index := range bytes {
		bytes[index] = charset[rand.Intn(len(charset))]
	}
	return string(bytes)
}

func NewVariableString(minimumLength int, maximumLength int, charset string) string {
	gomega.Expect(minimumLength).To(gomega.BeNumerically("<=", maximumLength))
	return NewString(minimumLength+rand.Intn(maximumLength-minimumLength+1), charset)
}

func NewText(minimumLength int, maximumLength int) string {
	return NewVariableString(minimumLength, maximumLength, CharsetText)
}
