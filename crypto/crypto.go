package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

func IsValidBase64EncodedMD5Hash(value string) bool {
	return ValidateBase64EncodedMD5Hash(value) == nil
}

func Base64EncodedMD5HashValidator(value string, errorReporter structure.ErrorReporter) {
	errorReporter.ReportError(ValidateBase64EncodedMD5Hash(value))
}

func ValidateBase64EncodedMD5Hash(value string) error {
	if value == "" {
		return structureValidator.ErrorValueEmpty()
	} else if bytes, err := base64.StdEncoding.DecodeString(value); err != nil {
		return ErrorValueStringAsBase64EncodedMD5HashNotValid(value)
	} else if len(bytes) != 16 {
		return ErrorValueStringAsBase64EncodedMD5HashNotValid(value)
	}
	return nil
}

func ErrorValueStringAsBase64EncodedMD5HashNotValid(value string) error {
	return errors.Preparedf(structureValidator.ErrorCodeValueNotValid, "value is not valid", "value %q is not valid as Base64 encoded MD5 hash", value)
}

func HexEncodedMD5Hash(sourceString string) string {
	md5Sum := md5.Sum([]byte(sourceString))
	return hex.EncodeToString(md5Sum[:])
}

func EncryptWithAES256UsingPassphrase(bytes []byte, passphrase []byte) (_ []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("unrecoverable encryption failure")
		}
	}()

	if len(bytes) == 0 {
		return nil, errors.New("bytes is missing")
	}
	if len(passphrase) == 0 {
		return nil, errors.New("passphrase is missing")
	}

	key, iv := passphraseToKey32AndIV16(passphrase)
	return encryptWithAES256(bytes, key, iv)
}

func DecryptWithAES256UsingPassphrase(bytes []byte, passphrase []byte) (_ []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("unrecoverable decryption failure")
		}
	}()

	if len(bytes) == 0 {
		return nil, errors.New("bytes is missing")
	}
	if len(passphrase) == 0 {
		return nil, errors.New("passphrase is missing")
	}

	key, iv := passphraseToKey32AndIV16(passphrase)
	return decryptWithAES256(bytes, key, iv)
}

func encryptWithAES256(bytes []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	paddedBytes := padBytesWithPKCS7(bytes)
	encryptedBytes := make([]byte, len(paddedBytes))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encryptedBytes, paddedBytes)
	return encryptedBytes, nil
}

func decryptWithAES256(bytes []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decryptedBytes := make([]byte, len(bytes))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(decryptedBytes, bytes)
	return unpadBytesWithPKCS7(decryptedBytes), nil
}

func hash(hashed []byte, password []byte) []byte {
	data := make([]byte, len(hashed)+len(password))
	copy(data, hashed)
	copy(data[len(hashed):], password)
	return generateMD5Sum(data)
}

func generateMD5Sum(data []byte) []byte {
	hash := md5.New()
	_, _ = hash.Write(data)
	return hash.Sum(nil)
}

func padBytesWithPKCS7(bytes []byte) []byte {
	overflowLength := len(bytes) % aes.BlockSize
	if overflowLength == 0 {
		return bytes
	}

	paddingLength := aes.BlockSize - overflowLength
	paddedBytes := make([]byte, len(bytes)+paddingLength)
	copy(paddedBytes, bytes)
	for i := 0; i < paddingLength; i++ {
		paddedBytes[len(bytes)+i] = byte(paddingLength)
	}
	return paddedBytes
}

func unpadBytesWithPKCS7(bytes []byte) []byte {
	overflowLength := int(bytes[len(bytes)-1])
	if overflowLength >= aes.BlockSize {
		return bytes
	}

	return bytes[:len(bytes)-overflowLength]
}

func passphraseToKey32AndIV16(passphrase []byte) ([]byte, []byte) {
	keyAndIV := make([]byte, 48)
	hashed := []byte{}
	for i := 0; i < 3; i++ {
		hashed = hash(hashed, passphrase)
		copy(keyAndIV[i*16:], hashed)
	}
	return keyAndIV[:32], keyAndIV[32:]
}
