package auth

import (
	"context"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/page"
	"time"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	DeviceAuthorizationPending    = "pending"
	DeviceAuthorizationExpired    = "expired"
	DeviceAuthorizationSuccessful = "successful"
	DeviceAuthorizationFailed     = "failed"

	LoopBundleId               = "org.tidepool.Loop"
	LoopBundleIdWithTeamPrefix = "75U4X84TEG.org.tidepool.Loop"
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
	BundleId         string `json:"bundleId" bson:"bundleId"`
	VerificationCode string `json:"verificationCode" bson:"verificationCode"`
	DeviceCheckToken string `json:"deviceCheckToken" bson:"deviceCheckToken"`
	Status           string `json:"-" bson:"status"`
}

func (d *DeviceAuthorizationUpdate) Validate(validator structure.Validator) {
	// We should not validate the bundleId here, because it will fail the request,
	// but it will not persist the failure in the database.
	validator.String("verificationCode", &d.VerificationCode).NotEmpty()
	validator.String("deviceCheckToken", &d.DeviceCheckToken).NotEmpty()
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
	BundleId         string     `json:"bundleId,omitempty" bson:"bundleId,omitempty"`
	VerificationCode string     `json:"verificationCode,omitempty" bson:"verificationCode,omitempty"`
	DeviceCheckToken string     `json:"deviceCheckToken,omitempty" bson:"deviceCheckToken,omitempty"`
	CreatedTime      time.Time  `json:"createdTime" bson:"createdTime"`
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

	return &DeviceAuthorization{
		ID:              NewDeviceAuthorizationID(),
		UserID:          userID,
		Token:           NewDeviceAuthorizationToken(),
		DevicePushToken: create.DevicePushToken,
		Status:          DeviceAuthorizationPending,
		CreatedTime:     time.Now(),
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

func (d *DeviceAuthorization) UpdateBundleId(bundleId string) error {
	if d.BundleId != "" {
		return errors.New("bundle id is already set")
	}
	if err := ValidateBundleId(bundleId); err != nil {
		return err
	}

	d.BundleId = bundleId
	return nil
}

func ValidBundleIds() []string {
	return []string{LoopBundleId, LoopBundleIdWithTeamPrefix}
}

func arrayContains(arr []string, element string) bool {
	for _, a := range arr {
		if a == element {
			return true
		}
	}

	return false
}

func ValidateBundleId(bundleId string) error {
	if arrayContains(ValidBundleIds(), bundleId) {
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
