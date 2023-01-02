package appvalidate

import (
	"context"

	"github.com/kelseyhightower/envconfig"

	"github.com/tidepool-org/platform/errors"
	structValidator "github.com/tidepool-org/platform/structure/validator"
)

type ValidatorConfig struct {
	AppleAppID               string `envconfig:"TIDEPOOL_APPVALIDATION_APPLE_APP_ID" default:"org.tidepool.app"`
	UseProductionEnvironment bool   `envconfig:"TIDEPOOL_APPVALIDATION_USE_PRODUCTION" default:"false"`
	ChallengeSize            int    `envconfig:"TIDEPOOL_APPVALIDATION_CHALLENGE_SIZE" default:"12"`
}

// Validator is the "service" that performs every flow or action associated
// with attesting and asserting an iOS app's integrity.
// https://developer.apple.com/documentation/devicecheck/establishing_your_app_s_integrity
type Validator struct {
	repo          Repository
	generator     ChallengeGenerator
	isProduction  bool
	appleAppID    string
	challengeSize int
}

func NewValidatorConfig() (*ValidatorConfig, error) {
	cfg := &ValidatorConfig{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewValidator(r Repository, g ChallengeGenerator, cfg ValidatorConfig) (*Validator, error) {
	if cfg.AppleAppID == "" {
		return nil, errors.New("app id cannot be empty")
	}
	if cfg.ChallengeSize <= 0 {
		return nil, errors.New("challenge size must be a postive integer")
	}
	if r == nil {
		return nil, errors.New("repository is missing")
	}
	if g == nil {
		return nil, errors.New("challenge generator is missing")
	}
	return &Validator{
		repo:          r,
		generator:     g,
		appleAppID:    cfg.AppleAppID,
		isProduction:  cfg.UseProductionEnvironment,
		challengeSize: cfg.ChallengeSize,
	}, nil
}

func (v *Validator) CreateAttestChallenge(ctx context.Context, c *ChallengeCreate) (*ChallengeResult, error) {
	if err := structValidator.New().Validate(c); err != nil {
		return nil, err
	}

	// Once a request for an attestation challenge is requested, create the
	// backing AppValidation object to associate and track the state of this
	// attestation and future assertions.
	challenge, err := v.generator.GenerateChallenge(v.challengeSize)
	if err != nil {
		return nil, err
	}
	if challenge == "" {
		return nil, errors.New("empty challenge generated")
	}

	validation, err := NewAppValidation(challenge, c)
	if err != nil {
		return nil, err
	}
	if err := v.repo.Upsert(ctx, validation); err != nil {
		return nil, err
	}
	return &ChallengeResult{Challenge: challenge}, nil
}

func (v *Validator) CreateAssertChallenge(ctx context.Context, c *ChallengeCreate) (*ChallengeResult, error) {
	if err := structValidator.New().Validate(c); err != nil {
		return nil, err
	}

	filter := Filter{UserID: c.UserID, KeyID: c.KeyID}
	verified, err := v.repo.IsVerified(ctx, filter)
	if err != nil {
		return nil, err
	}
	// Can only create an assertion if already attested and that attestation
	// is verified.
	// https://developer.apple.com/documentation/devicecheck/establishing_your_app_s_integrity#3561591
	if !verified {
		return nil, errors.New("cannot request assertion if attestation is not verified")
	}

	challenge, err := v.generator.GenerateChallenge(v.challengeSize)
	if err != nil {
		return nil, err
	}
	if challenge == "" {
		return nil, errors.New("empty challenge generated")
	}

	update := AssertionUpdate{
		Challenge: challenge,
	}
	if err := v.repo.UpdateAssertion(ctx, filter, update); err != nil {
		return nil, err
	}

	return &ChallengeResult{Challenge: challenge}, nil
}
