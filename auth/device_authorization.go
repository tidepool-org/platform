package auth

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeviceAuthorizationPending    = "pending"
	DeviceAuthorizationExpired    = "expired"
	DeviceAuthorizationSuccessful = "successful"
	DeviceAuthorizationFailed     = "failed"

	LoopBundleID               = "org.tidepool.Loop"
	LoopBundleIDWithTeamPrefix = "75U4X84TEG.org.tidepool.Loop"

	DeviceAuthorizationExpirationDuration = time.Hour
)

type DeviceAuthorizationAccessor interface {
	GetUserDeviceAuthorization(ctx context.Context, userID string, id string) (*DeviceAuthorization, error)
	ListUserDeviceAuthorizations(ctx context.Context, userID string, pagination *page.Pagination) (DeviceAuthorizations, error)
	GetDeviceAuthorizationByToken(ctx context.Context, token string) (*DeviceAuthorization, error)
	CreateUserDeviceAuthorization(ctx context.Context, userID string, create *DeviceAuthorizationCreate) (*DeviceAuthorization, error)
	UpdateDeviceAuthorization(ctx context.Context, id string, update *DeviceAuthorizationUpdate) (*DeviceAuthorization, error)
}

type DeviceAuthorizationCreate struct {
	DevicePushToken string `json:"devicePushToken" bson:"devicePushToken"`
}

func NewDeviceAuthorizationCreate() *DeviceAuthorizationCreate {
	return &DeviceAuthorizationCreate{}
}

func (d *DeviceAuthorizationCreate) Parse(parser structure.ObjectParser) {
	if ptr := parser.String("devicePushToken"); ptr != nil {
		d.DevicePushToken = *ptr
	}
}

func (d *DeviceAuthorizationCreate) Validate(validator structure.Validator) {
	validator.String("devicePushToken", &d.DevicePushToken).LengthGreaterThanOrEqualTo(64)
}

type DeviceAuthorizationUpdate struct {
	BundleID         string `json:"bundleId" bson:"bundleId"`
	VerificationCode string `json:"verificationCode" bson:"verificationCode,omitempty"`
	DeviceCheckToken string `json:"deviceCheckToken" bson:"deviceCheckToken"`
	Status           string `json:"-" bson:"status"`
}

func (d *DeviceAuthorizationUpdate) Validate(validator structure.Validator) {
	// We should not validate the bundleId here, because the request will fail,
	// without persisting the failure in the database.
	validator.String("verificationCode", &d.VerificationCode).NotEmpty()
	validator.String("deviceCheckToken", &d.DeviceCheckToken).NotEmpty()
}

func (d *DeviceAuthorizationUpdate) Expire() {
	d.VerificationCode = ""
	d.Status = DeviceAuthorizationExpired
}

func (d *DeviceAuthorizationUpdate) IsExpired() bool {
	return d.Status == DeviceAuthorizationExpired
}

func NewDeviceAuthorizationUpdate() *DeviceAuthorizationUpdate {
	return &DeviceAuthorizationUpdate{}
}

type DeviceAuthorization struct {
	ID               string     `json:"id" bson:"id"`
	UserID           string     `json:"-" bson:"userId"`
	Token            string     `json:"-" bson:"token"`
	DevicePushToken  string     `json:"devicePushToken,omitempty" bson:"devicePushToken"`
	Status           string     `json:"status" bson:"status"`
	BundleID         string     `json:"bundleId,omitempty" bson:"bundleId,omitempty"`
	VerificationCode string     `json:"verificationCode,omitempty" bson:"verificationCode,omitempty"`
	DeviceCheckToken string     `json:"deviceCheckToken,omitempty" bson:"deviceCheckToken,omitempty"`
	CreatedTime      time.Time  `json:"createdTime" bson:"createdTime"`
	ExpirationTime   time.Time  `json:"expirationTime" bson:"expirationTime"`
	ModifiedTime     *time.Time `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
}

type DeviceAuthorizations []*DeviceAuthorization

func StatusTypes() []string {
	return []string{DeviceAuthorizationPending, DeviceAuthorizationExpired, DeviceAuthorizationSuccessful, DeviceAuthorizationFailed}
}

func NewDeviceAuthorizationID() string {
	// 8 bytes or 16 hex chars
	return id.Must(id.New(8))
}

func NewDeviceAuthorizationToken() string {
	// 16 bytes or 32 hex chars
	return id.Must(id.New(16))
}

func NewDeviceAuthorization(userID string, create *DeviceAuthorizationCreate) (*DeviceAuthorization, error) {
	if userID == "" {
		return nil, errors.New("user id is missing")
	}
	if create == nil {
		return nil, errors.New("create is missing")
	} else if err := structureValidator.New().Validate(create); err != nil {
		return nil, errors.Wrap(err, "create is invalid")
	}

	now := time.Now()
	return &DeviceAuthorization{
		ID:              NewDeviceAuthorizationID(),
		UserID:          userID,
		Token:           NewDeviceAuthorizationToken(),
		DevicePushToken: create.DevicePushToken,
		Status:          DeviceAuthorizationPending,
		CreatedTime:     now,
		ExpirationTime:  now.Add(DeviceAuthorizationExpirationDuration),
	}, nil
}

func (d *DeviceAuthorization) Validate(validator structure.Validator) {
	validator.String("id", &d.ID).Alphanumeric().LengthEqualTo(16)
	validator.String("userId", &d.UserID).Using(UserIDValidator)
	validator.String("token", &d.Token).Alphanumeric().LengthEqualTo(32)
	validator.String("devicePushToken", &d.DevicePushToken).LengthGreaterThanOrEqualTo(64)
	validator.String("status", &d.Status).OneOf(StatusTypes()...)

	validator.Time("createdTime", &d.CreatedTime).NotZero().BeforeNow(time.Second)
	validator.Time("modifiedTime", d.ModifiedTime).After(d.CreatedTime).BeforeNow(time.Second)
}

func (d *DeviceAuthorization) UpdateBundleID(bundleID string) error {
	if d.BundleID != "" {
		return errors.New("bundle id is already set")
	}
	if err := ValidateBundleID(bundleID); err != nil {
		return err
	}

	d.BundleID = bundleID
	return nil
}

func (d *DeviceAuthorization) UpdateStatus(status string) error {
	if d.IsCompleted() {
		return errors.New("cannot update status of a completed device authorization")
	}
	if err := ValidateStatus(status); err != nil {
		return err
	}

	d.Status = status
	return nil
}

func (d *DeviceAuthorization) IsCompleted() bool {
	return d.Status != "" && d.Status != DeviceAuthorizationPending
}

func (d *DeviceAuthorization) ShouldExpire() bool {
	return time.Now().After(d.ExpirationTime)
}

func ValidBundleIds() []string {
	return []string{LoopBundleID, LoopBundleIDWithTeamPrefix}
}

func arrayContains(arr []string, element string) bool {
	for _, a := range arr {
		if a == element {
			return true
		}
	}

	return false
}

func ValidateBundleID(bundleID string) error {
	if arrayContains(ValidBundleIds(), bundleID) {
		return nil
	}

	return errors.New("bundle id is not valid")
}

func ValidateStatus(status string) error {
	if arrayContains(StatusTypes(), status) {
		return nil
	}

	return errors.New("status is not valid")
}
