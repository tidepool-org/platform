package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tidepool-org/platform/auth"
	authClient "github.com/tidepool-org/platform/auth/client"
	"github.com/tidepool-org/platform/platform"

	consentStore "github.com/tidepool-org/platform/consent/store/mongo"
	userClient "github.com/tidepool-org/platform/user/client"

	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"

	"github.com/tidepool-org/platform/pointer"

	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/log"

	"github.com/tidepool-org/platform/user"

	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tidepool-org/platform/application"
	"github.com/tidepool-org/platform/consent"
	"github.com/tidepool-org/platform/errors"
	migrationMongo "github.com/tidepool-org/platform/migration/mongo"
)

const gatekeeperDBName = "gatekeeper"
const seagullDBName = "seagull"
const consentDBName = "tidepool"

const permissionsCollectionName = "perms"
const profilesCollectionName = "seagull"
const consentRecordsCollectionName = "consent_records"

var (
	bddpOrganizations = []Organization{
		{
			Email: "bigdata+adces@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsADCES,
		},
		{
			Email: "bigdata+bt1@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsBeyondType1,
		},
		{
			Email: "bigdata+cwd@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsChildrenWithDiabetes,
		},
		{
			Email: "bigdata+cdn@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsTheDiabetesLink,
		},
		{
			Email: "bigdata+dyf1@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsDYF,
		},
		{
			Email: "bigdata+diabetessisters@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsDiabetesSisters,
		},
		{
			Email: "bigdata+diatribe@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsTheDiaTribeFoundation,
		},
		{
			Email: "bigdata+jdrf@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsBreakthroughT1D,
		},
		{
			Email: "bigdata+nsf@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsNightscoutFoundation,
		},
		{
			Email: "bigdata+t1dx@tidepool.org",
			Name:  consent.BigDataDonationProjectOrganizationsT1DExchange,
		},
	}
)

type Organization struct {
	Email string
	Name  consent.BigDataDonationProjectOrganization
}

type Permission struct {
	ID       primitive.ObjectID `bson:"_id"`
	SharerID string             `bson:"sharerId"`
	UserID   string             `bson:"userId"`
}

type SeagullDocument struct {
	ID     primitive.ObjectID `bson:"_id"`
	UserID string             `bson:"userId"`
	Value  string             `bson:"value"`
}

type Profile struct {
	FullName *string        `json:"fullName"`
	Patient  PatientProfile `json:"patient"`
}

type PatientProfile struct {
	Birthday      *string `json:"birthday"`
	IsOtherPerson bool    `json:"isOtherPerson"`
	FullName      *string `json:"fullName"`
}

func main() {
	application.RunAndExit(NewMigration())
}

type Migration struct {
	*migrationMongo.Migration

	consentRecordRepository consentStore.ConsentRecordRepository
	userClient              user.Client

	permissionsCollection *mongo.Collection
	profilesCollection    *mongo.Collection

	organizationsMap    map[string]Organization
	organizationUserIDs []string
}

func NewMigration() *Migration {
	return &Migration{
		Migration: migrationMongo.NewMigration(),
	}
}

func (m *Migration) Initialize(provider application.Provider) error {
	if err := m.Migration.Initialize(provider); err != nil {
		return err
	}

	m.CLI().Usage = "BACK-3857: Create consent records for each BDDP sharing relationship"
	m.CLI().Description = "BACK-3857"
	m.CLI().Authors = []cli.Author{
		{
			Name:  "Todd Kazakov",
			Email: "todd@tidepool.org",
		},
	}
	m.CLI().Action = func(ctx *cli.Context) error {
		if !m.ParseContext(ctx) {
			return nil
		}

		executionContext, cancel := context.WithCancel(context.Background())
		executionContext = log.NewContextWithLogger(executionContext, m.Logger())
		go func() {
			defer cancel()
			quitChannel := make(chan os.Signal, 1)
			signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
			<-quitChannel
		}()

		return m.execute(executionContext)
	}

	return nil
}

func (m *Migration) execute(ctx context.Context) error {
	m.Logger().Debug("Creating data store")

	timeout := time.Second * 10
	clientOptions := options.Client().
		ApplyURI(m.NewMongoConfig().AsConnectionString()).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return errors.Wrap(err, "connection options are invalid")
	}

	m.permissionsCollection = client.Database(gatekeeperDBName).Collection(permissionsCollectionName)
	m.profilesCollection = client.Database(seagullDBName).Collection(profilesCollectionName)

	externalConfig := authClient.NewExternalConfig()
	if err := externalConfig.Load(authClient.NewExternalEnvconfigLoader(nil)); err != nil {
		return err
	}
	authClnt, err := authClient.NewExternal(externalConfig, platform.AuthorizeAsService, "user_client", m.Logger())
	if err != nil {
		return err
	}
	if err := authClnt.Start(); err != nil {
		return err
	}
	ctx = auth.NewContextWithServerSessionTokenProvider(ctx, authClnt)

	m.userClient, err = userClient.NewDefaultClient(userClient.Params{
		ConfigReporter: m.ConfigReporter(),
		Logger:         m.Logger(),
		UserAgent:      "consent_migration",
	})
	m.consentRecordRepository = consentStore.ConsentRecordRepository{
		Repository: storeStructuredMongo.NewRepository(client.Database(consentDBName).Collection(consentRecordsCollectionName)),
	}

	orgs := make(map[string]Organization)
	m.organizationUserIDs = make([]string, 0, len(bddpOrganizations))
	for _, org := range bddpOrganizations {
		userID, err := m.resolveUserID(ctx, org.Email)
		if err != nil {
			m.Logger().WithError(err).Warnf("unable to resolve user id of %s organization", org.Email)
			continue
		}
		orgs[userID] = org
		m.organizationUserIDs = append(m.organizationUserIDs, userID)
	}

	bigdataUserID, err := m.resolveUserID(ctx, "bigdata@tidepool.org")
	if err != nil {
		return errors.Wrap(err, "unable to resolve bigdata user ID")
	}
	if err := m.ensureUsersShareWithBigdataAccount(ctx, bigdataUserID); err != nil {
		return errors.Wrap(err, "unable to ensure all organization users share with bigdata@tidepool.org")
	}

	return m.migrateOrganizationUsers(ctx, bigdataUserID, m.migrateUser)
}

func (m *Migration) ensureUsersShareWithBigdataAccount(ctx context.Context, bigdataUserID string) error {
	sort := bson.M{"_id": 1}
	batchSize := 1000
	opts := options.Find().SetSort(sort).SetLimit(int64(batchSize))

	for _, organizationUserID := range m.organizationUserIDs {
		for {
			id := primitive.NilObjectID
			selector := bson.M{
				"_id": bson.M{
					"$gt": id,
				},
				"userId": organizationUserID,
			}

			cursor, err := m.permissionsCollection.Find(ctx, selector, opts)
			if err != nil {
				return errors.Wrap(err, "error finding permissions")
			}

			var results []Permission
			if err := cursor.All(ctx, &results); err != nil {
				return errors.Wrap(err, "error decoding permissions")
			}

			for _, result := range results {
				bigdataShareSelector := bson.M{
					"sharerId": result.SharerID,
					"userId":   bigdataUserID,
				}
				insert := bson.M{
					"sharerId": result.SharerID,
					"userId":   bigdataUserID,
					"permissions": bson.M{
						"view": bson.M{},
					},
				}
				if m.DryRun() {
					if err := m.permissionsCollection.FindOne(ctx, bigdataShareSelector).Err(); err != nil {
						if errors.Is(err, mongo.ErrNoDocuments) {
							m.Logger().Infof("[DRY RUN] creating big data sharing relationship for user %s", result.SharerID)
						} else {
							m.Logger().WithError(err).Errorf("[DRY RUN] error finding big data sharing relationship for user %s", result.SharerID)
						}
					}
					continue
				}

				res, err := m.permissionsCollection.UpdateOne(ctx, bigdataShareSelector, bson.M{
					"$setOnInsert": insert,
				}, options.Update().SetUpsert(true))
				if err != nil {
					return errors.Wrapf(err, "error creating big data permission for user %s", result.SharerID)
				}
				if res.UpsertedCount > 0 {
					m.Logger().Infof("successfully created big data permission relationship for user %s", result.SharerID)
				}
			}

			if len(results) < batchSize {
				break
			}
			id = results[len(results)-1].ID
		}
	}

	return nil
}

func (m *Migration) migrateOrganizationUsers(ctx context.Context, organizationUserID string, migrate func(context.Context, Permission) error) error {
	sort := bson.M{"_id": 1}
	batchSize := 1000
	opts := options.Find().SetSort(sort).SetLimit(int64(batchSize))
	id := primitive.NilObjectID
	var successCount, errorCount int

	for {
		selector := bson.M{
			"_id": bson.M{
				"$gt": id,
			},
			"userId": organizationUserID,
		}

		cursor, err := m.permissionsCollection.Find(ctx, selector, opts)
		if err != nil {
			return errors.Wrap(err, "error finding permissions")
		}

		var results []Permission
		if err := cursor.All(ctx, &results); err != nil {
			return errors.Wrap(err, "error decoding permissions")
		}

		for _, result := range results {
			if err := migrate(ctx, result); err != nil {
				errorCount++
				m.Logger().WithError(err).Errorf("error migrating consent for user %s", result.SharerID)
			} else {
				successCount++
			}
		}

		if len(results) < batchSize {
			break
		}
		id = results[len(results)-1].ID
	}

	m.Logger().Infof("Success count: %d, error count: %d", successCount, errorCount)

	return nil
}

func (m *Migration) migrateUser(ctx context.Context, share Permission) error {
	create, err := m.createRecordForUser(ctx, share)
	if err != nil {
		return errors.Wrapf(err, "error generating create record for user %s", share.SharerID)
	}

	pagination := page.NewPagination()
	pagination.Size = 1
	result, err := m.consentRecordRepository.ListConsentRecords(ctx, share.SharerID, &consent.RecordFilter{
		Type:    pointer.FromAny(create.Type),
		Version: pointer.FromAny(create.Version),
	}, pagination)
	if err != nil {
		return errors.Wrapf(err, "error listing consent records of user %s", share.SharerID)
	}
	if result.Count > 0 {
		m.Logger().Infof("skipping migration of user %s, because consent record already exists", share.SharerID)
		return nil
	}

	if err := m.migrate(ctx, share.SharerID, create); err != nil {
		return errors.Wrapf(err, "error migrating consent of user %s", share.SharerID)
	}

	m.Logger().Debugf("successfully migrated user %s", share.SharerID)
	return nil
}

func (m *Migration) createRecordForUser(ctx context.Context, share Permission) (*consent.RecordCreate, error) {
	c, err := m.permissionsCollection.Find(ctx, bson.M{
		"sharerId": share.SharerID,
		"userId": bson.M{
			"$in": m.organizationUserIDs,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "error finding organization permissions")
	}
	var organizationsForUser []Permission
	if err := c.All(ctx, &organizationsForUser); err != nil {
		return nil, errors.Wrap(err, "error decoding organizations permissions")
	}

	record := consent.NewRecordCreate()
	record.Type = "big_data_donation_project"
	record.Version = 1
	record.Metadata = &consent.RecordMetadata{
		SupportedOrganizations: getOrganizationNames(organizationsForUser, m.organizationsMap),
	}
	if earliest := getEarliestPermissionTime(organizationsForUser); !earliest.IsZero() {
		record.GrantTime = share.ID.Timestamp()
	}

	if err := m.populateAttributesFromUserProfile(ctx, share.SharerID, record); err != nil {
		return nil, errors.Wrapf(err, "error populating profile attributes")
	}

	return record, nil
}

func (m *Migration) populateAttributesFromUserProfile(ctx context.Context, userID string, create *consent.RecordCreate) error {
	document := SeagullDocument{}
	err := m.profilesCollection.FindOne(ctx, bson.M{"userId": userID}).Decode(&document)
	if err != nil {
		return errors.Wrapf(err, "unable to find profile of user %s", userID)
	}

	profile := Profile{}
	err = json.NewDecoder(strings.NewReader(document.Value)).Decode(&profile)
	if err != nil {
		return errors.Wrapf(err, "unable to decode profile of user %s", userID)
	}

	if profile.Patient.Birthday == nil {
		return errors.Newf("profile birthday is nil for user %s", userID)
	}
	if profile.FullName == nil {
		return errors.New("profile full name is nil for user " + userID)
	}

	// Determine approximate age at grant time
	birthday, err := time.Parse(time.DateOnly, *profile.Patient.Birthday)
	if err != nil {
		return errors.Wrapf(err, "unable to parse birthday for user %s", userID)
	}
	if age := yearsDifference(create.GrantTime, birthday); age >= 18 {
		create.AgeGroup = consent.AgeGroupEighteenOrOver
		create.GrantorType = consent.GrantorTypeOwner
		create.OwnerName = *profile.FullName
	} else {
		if profile.Patient.FullName == nil {
			return errors.Newf("patient full name is nil for user %s", userID)
		}

		if age > 13 {
			create.AgeGroup = consent.AgeGroupThirteenSeventeen
		} else {
			create.AgeGroup = consent.AgeGroupUnderThirteen
		}
		create.GrantorType = consent.GrantorTypeParentGuardian
		create.ParentGuardianName = profile.FullName
		create.OwnerName = *profile.Patient.FullName
	}

	return nil
}

func (m *Migration) resolveUserID(ctx context.Context, email string) (string, error) {
	usr, err := m.userClient.Get(ctx, email)
	if err != nil {
		return "", err
	}
	if usr == nil || usr.UserID == nil {
		return "", errors.New("user not found")
	}
	return *usr.UserID, nil
}

func (m *Migration) migrate(ctx context.Context, userID string, create *consent.RecordCreate) error {
	if m.DryRun() {
		m.Logger().Infof("[DRY RUN] migrating user %s", userID)
		return nil
	}
	created, err := m.consentRecordRepository.CreateConsentRecord(ctx, userID, create)
	if err != nil {
		return errors.Wrapf(err, "unable to create consent record for user %s", userID)
	}

	m.Logger().Infof("sucessfully created consent record for user %s", created.UserID)
	return nil
}

func getOrganizationNames(permissions []Permission, orgs map[string]Organization) []consent.BigDataDonationProjectOrganization {
	var names = make([]consent.BigDataDonationProjectOrganization, 0, len(permissions))
	for _, permission := range permissions {
		if org, ok := orgs[permission.UserID]; ok {
			names = append(names, org.Name)
		}
	}
	return names
}

func getEarliestPermissionTime(permissions []Permission) time.Time {
	var earliest time.Time
	for _, permission := range permissions {
		if earliest.IsZero() || permission.ID.Timestamp().Before(earliest) {
			earliest = permission.ID.Timestamp()
		}
	}
	return earliest
}

func yearsDifference(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}

	years := end.Year() - start.Year()
	if end.YearDay() < start.YearDay() {
		years--
	}

	return years
}
