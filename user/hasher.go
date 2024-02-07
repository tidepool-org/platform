package user

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"time"
)

func generateUniqueHash(strings []string, length int) (string, error) {

	if len(strings) > 0 && length > 0 {

		hash := sha256.New()

		for i := range strings {
			hash.Write([]byte(strings[i]))
		}

		max := big.NewInt(9999999999)
		//add some randomness
		n, err := rand.Int(rand.Reader, max)

		if err != nil {
			return "", err
		}
		hash.Write([]byte(n.String()))
		//and use unix nano
		hash.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
		hashString := hex.EncodeToString(hash.Sum(nil))
		return string([]rune(hashString)[0:length]), nil
	}

	return "", errors.New("both strings and length are required")

}

func GeneratePasswordHash(id, pw, salt string) (string, error) {

	if salt == "" || id == "" {
		return "", errors.New("id and salt are required")
	}

	hash := sha1.New()
	if pw != "" {
		hash.Write([]byte(pw))
	}
	hash.Write([]byte(salt))
	hash.Write([]byte(id))
	pwHash := hex.EncodeToString(hash.Sum(nil))

	return pwHash, nil
}
