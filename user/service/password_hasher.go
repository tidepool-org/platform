package service

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/errors"
)

type PasswordHasherConfig struct {
	Salt string `json:"-"`
}

func NewPasswordHasherConfig() *PasswordHasherConfig {
	return &PasswordHasherConfig{}
}

func (p *PasswordHasherConfig) Load(configReporter config.Reporter) error {
	if configReporter == nil {
		return errors.New("config reporter is missing")
	}

	p.Salt = configReporter.GetWithDefault("salt", p.Salt)

	return nil
}

func (p *PasswordHasherConfig) Validate() error {
	if p.Salt == "" {
		return errors.New("salt is missing")
	}

	return nil
}

type PasswordHasher struct {
	salt string
}

func NewPasswordHasher(cfg *PasswordHasherConfig) (*PasswordHasher, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	} else if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &PasswordHasher{
		salt: cfg.Salt,
	}, nil
}

func (p *PasswordHasher) HashPassword(userID string, password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))
	hash.Write([]byte(p.salt))
	hash.Write([]byte(userID))
	return hex.EncodeToString(hash.Sum(nil))
}
