package test

func MustBytes(bytes []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return bytes
}
