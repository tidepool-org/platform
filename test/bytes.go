package test

import "math/rand"

func MustBytes(value []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomBytes() []byte {
	return RandomBytesFromRange(RandomBytesLengthMinimum(), RandomBytesLengthMaximum())
}

func RandomBytesFromRange(minimumLength int, maximumLength int) []byte {
	if maximumLength < minimumLength {
		panic("RandomBytesFromRange: maximum length is not greater than or equal to minimum length")
	}
	result := make([]byte, RandomIntFromRange(minimumLength, maximumLength))
	length, err := rand.Read(result)
	if err != nil || length != len(result) {
		panic("RandomBytesFromRange: unable to read random bytes")
	}
	return result
}

func RandomBytesLengthMaximum() int {
	return 256
}

func RandomBytesLengthMinimum() int {
	return 1
}
