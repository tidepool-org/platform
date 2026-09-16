package mongo

import (
	"context"
	stdErrors "errors"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/user"
)

// Config is the same as [storeStructuredMongo.Config] except with explicit seagull envconfig tags.
type Config struct {
	Scheme           string        `envconfig:"SEAGULL_TIDEPOOL_STORE_SCHEME"`
	Addresses        []string      `envconfig:"SEAGULL_TIDEPOOL_STORE_ADDRESSES" required:"true"`
	TLS              bool          `envconfig:"SEAGULL_TIDEPOOL_STORE_TLS" default:"true"`
	Database         string        `envconfig:"SEAGULL_TIDEPOOL_STORE_DATABASE" required:"true"`
	CollectionPrefix string        `envconfig:"SEAGULL_TIDEPOOL_STORE_COLLECTION_PREFIX"`
	Username         *string       `envconfig:"SEAGULL_TIDEPOOL_STORE_USERNAME" require:"true"`
	Password         *string       `envconfig:"SEAGULL_TIDEPOOL_STORE_PASSWORD" require:"true"`
	Timeout          time.Duration `envconfig:"SEAGULL_TIDEPOOL_STORE_TIMEOUT" default:"60s"`
	OptParams        *string       `envconfig:"SEAGULL_TIDEPOOL_STORE_OPT_PARAMS"`
}

var (
	_ Config = Config(storeStructuredMongo.Config{})
)

func NewConfig() (*Config, error) {
	c := Config{
		TLS:     true,
		Timeout: 30 * time.Second,
	}
	if err := envconfig.Process("", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// LegacySeagullProfileRepository accesses legacy seagull profiles while the
// seagll migration to keycloak is in progress.
type LegacySeagullProfileRepository struct {
	*storeStructuredMongo.Repository
}

func NewLegacySeagullProfileRepository(c *Config) (*LegacySeagullProfileRepository, error) {
	if c == nil {
		return nil, errors.New("config is missing")
	}
	storeConfig := storeStructuredMongo.Config(*c)

	store, err := storeStructuredMongo.NewStore(&storeConfig)
	if err != nil {
		return nil, err
	}
	return &LegacySeagullProfileRepository{
		store.GetRepository("seagull"),
	}, nil
}

func (p *LegacySeagullProfileRepository) EnsureIndexes() error {
	return nil
}

func (p *LegacySeagullProfileRepository) FindLegacyUserProfile(ctx context.Context, userID string) (*user.LegacyUserProfile, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	selector := bson.M{
		"userId": userID,
	}
	var doc user.LegacySeagullDocument
	if err := p.FindOne(ctx, selector).Decode(&doc); err != nil {
		if stdErrors.Is(err, mongo.ErrNoDocuments) {
			return nil, user.ErrUserProfileNotFound
		}
		return nil, err
	}

	return doc.ToLegacyProfile()
}

func (p *LegacySeagullProfileRepository) UpdateLegacyUserProfile(ctx context.Context, userID string, profile *user.LegacyUserProfile) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}
	if err := structureValidator.New(log.LoggerFromContext(ctx)).Validate(profile); err != nil {
		return err
	}
	var doc user.LegacySeagullDocument
	selector := bson.M{
		"userId": userID,
	}
	err := p.FindOne(ctx, selector).Decode(&doc)
	// A user can have no profile set - see seagull/lib/routes/seagullApi.js `if (err.statusCode == 404 && addIfNotThere)`
	if err != nil && !stdErrors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	hasExistingProfile := err == nil
	// We need to make a distinction b/t a seagull profile not existing (in which case we can upsert) versus a seagull profile actively being migrated, which is why we need to actually read the document.
	if hasExistingProfile && doc.IsMigrating() {
		return user.ErrUserProfileMigrationInProgress
	}

	// This will create a new value even if doc.Value is empty
	updatedValueRaw, err := user.AddProfileToSeagullValue(doc.Value, profile)
	if err != nil {
		return err
	}

	uopts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	uselector := bson.M{
		"userId": userID,
	}
	update := bson.M{
		"$set": bson.M{
			"value":  updatedValueRaw,
			"userId": userID, // Set because of possible upsert
		},
	}
	var updatedDoc user.LegacySeagullDocument
	err = p.FindOneAndUpdate(ctx, uselector, update, uopts).Decode(&updatedDoc)
	if err != nil {
		return err
	}
	// Handle case where a migration was started in between the start of this function and the update
	if updatedDoc.IsMigrating() {
		return user.ErrUserProfileMigrationInProgress
	}
	return nil
}
