package structured

import (
	"context"
	"io"

	"github.com/tidepool-org/platform/blob"
	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/net"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
)

type Store interface {
	NewSession() Session
}

type Session interface {
	io.Closer

	List(ctx context.Context, userID string, filter *blob.Filter, pagination *page.Pagination) (blob.Blobs, error)
	Create(ctx context.Context, userID string, create *Create) (*blob.Blob, error)
	Get(ctx context.Context, id string) (*blob.Blob, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *Update) (*blob.Blob, error)
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
