package test

import "io/ioutil"

func RandomTemporaryDirectory() string {
	value, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	return value
}
