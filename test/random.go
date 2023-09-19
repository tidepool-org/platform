package test

import "math/rand"

func RandomChoice[T any](values []T) T {
	return values[rand.Intn(len(values))]
}

func RandomArrayWithLength[T any](length int, generator func() T) []T {
	array := []T{}
	for index := 0; index < length; index++ {
		array = append(array, generator())
	}
	return array
}
