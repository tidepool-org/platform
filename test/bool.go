package test

import "math/rand"

func MustBool(value bool, err error) bool {
	if err != nil {
		panic(err)
	}
	return value
}

func RandomBool() bool {
	return rand.Intn(2) == 0
}

func NewObjectFromBool(value bool, objectFormat ObjectFormat) interface{} {
	return value
}
