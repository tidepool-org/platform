package test

func MustBytes(bites []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return bites
}

func MustString(str string, err error) string {
	if err != nil {
		panic(err)
	}
	return str
}
