package test

import "math/rand"

func MustString(value string, err error) string {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomString() string {
	return RandomStringFromRangeAndCharset(RandomStringLengthMinimum(), RandomStringLengthMaximum(), CharsetText)
}

func RandomStringFromArray(array []string) string {
	if len(array) == 0 {
		panic("RandomStringFromArray: array is empty")
	}
	return array[rand.Intn(len(array))]
}

func RandomStringFromCharset(charset string) string {
	return RandomStringFromRangeAndCharset(RandomStringLengthMinimum(), RandomStringLengthMaximum(), charset)
}

func RandomStringFromRange(minimumLength int, maximumLength int) string {
	return RandomStringFromRangeAndCharset(minimumLength, maximumLength, CharsetText)
}

func RandomStringFromRangeAndCharset(minimumLength int, maximumLength int, charset string) string {
	if maximumLength < minimumLength {
		panic("RandomStringFromRangeAndCharset: maximum length is not greater than or equal to minimum length")
	}
	if len(charset) == 0 {
		panic("RandomStringFromRangeAndCharset: charset is empty")
	}
	charsetRunes := []rune(charset)
	resultRunes := make([]rune, RandomIntFromRange(minimumLength, maximumLength))
	for index := range resultRunes {
		resultRunes[index] = charsetRunes[rand.Intn(len(charsetRunes))]
	}
	return string(resultRunes)
}

func RandomStringLengthMaximum() int {
	return 32
}

func RandomStringLengthMinimum() int {
	return 1
}

func NewObjectFromString(value string, objectFormat ObjectFormat) interface{} {
	return value
}
