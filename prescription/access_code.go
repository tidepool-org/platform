package prescription

import (
	"math/rand"
)

const (
	accessCodeLength = 6
	characters       = "ABCDEFGHJKLMNPQRSTUVWXYZ123456789"
)

func GenerateAccessCode() string {
	b := make([]byte, accessCodeLength)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}
