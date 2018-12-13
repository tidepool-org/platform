package test

func MustBytes(bytes []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return bytes
}

func MustString(str string, err error) string {
	if err != nil {
		panic(err)
	}
	return str
}
