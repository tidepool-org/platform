package test

import (
	"io/ioutil"
	"os"
)

func RandomTemporaryFile() *os.File {
	value, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	return value
}
