package structured

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

type Store interface {
	NewBlobRepository() BlobRepository
	NewDeviceLogsRepository() DeviceLogsRepository
}

type BlobRepository interface {
	List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.BlobArray, error)
	Create(ctx context.Context, userID string, create *Create) (*blob.Blob, error)
	DeleteAll(ctx context.Context, userID string) (bool, error)
	DestroyAll(ctx context.Context, userID string) (bool, error)

	Get(ctx context.Context, id string, condition *request.Condition) (*blob.Blob, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *Update) (*blob.Blob, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (bool, error)
	Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error)
}

type Create struct {
	MediaType *string
}

func NewCreate() *Create {
	return &Create{}
}

func (c *Create) Validate(validator structure.Validator) {
	validator.String("mediaType", c.MediaType).Using(net.MediaTypeValidator)
}

type Update struct {
	DigestMD5 *string
	MediaType *string
	Size      *int
	Status    *string
}

func NewUpdate() *Update {
	return &Update{}
}

func (u *Update) Validate(validator structure.Validator) {
	validator.String("digestMD5", u.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", u.MediaType).Using(net.MediaTypeValidator)
	validator.Int("size", u.Size).GreaterThanOrEqualTo(0)
	validator.String("status", u.Status).OneOf(blob.Statuses()...)
}

func (u *Update) IsEmpty() bool {
	return u.DigestMD5 == nil && u.MediaType == nil && u.Size == nil && u.Status == nil
}

type DeviceLogsRepository interface {
	List(ctx context.Context, userID string, filter *blob.DeviceLogsFilter, pagination *page.Pagination) (blob.DeviceLogsBlobArray, error)
	Get(ctx context.Context, deviceLogID string) (*blob.DeviceLogsBlob, error)
	Create(ctx context.Context, userID string, create *Create) (*blob.DeviceLogsBlob, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *DeviceLogsUpdate) (*blob.DeviceLogsBlob, error)
	Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error)
}

type DeviceLogsUpdate struct {
	DigestMD5 *string
	MediaType *string
	Size      *int
	StartAt   *time.Time
	EndAt     *time.Time
}

func NewDeviceLogsUpdate() *DeviceLogsUpdate {
	return &DeviceLogsUpdate{
		StartAt: &time.Time{},
		EndAt:   &time.Time{},
	}
}

func (u *DeviceLogsUpdate) Validate(validator structure.Validator) {
	validator.String("digestMD5", u.DigestMD5).Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", u.MediaType).Using(net.MediaTypeValidator)
	validator.Int("size", u.Size).GreaterThanOrEqualTo(0)
	validator.Time("startAt", u.StartAt).Exists()
	validator.Time("endAt", u.EndAt).Exists()
}

func (u *DeviceLogsUpdate) IsEmpty() bool {
	return u.DigestMD5 == nil && u.MediaType == nil && u.Size == nil && u.StartAt.IsZero() && u.EndAt.IsZero()
}
