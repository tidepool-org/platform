package test

import (
	"io/ioutil"
	"os"
)

func RandomTemporaryFile() *os.File {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	return file
}

func RandomTemporaryDirectory() string {
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	return directory
}
