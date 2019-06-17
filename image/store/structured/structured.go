package structured

import (
	"context"
	"io"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/image"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/request"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

type Store interface {
	NewSession() Session
}

type Session interface {
	io.Closer

	List(ctx context.Context, userID string, filter *image.Filter, pagination *page.Pagination) (image.ImageArray, error)
	Create(ctx context.Context, userID string, metadata *image.Metadata) (*image.Image, error)
	DeleteAll(ctx context.Context, userID string) (bool, error)
	DestroyAll(ctx context.Context, userID string) (bool, error)

	Get(ctx context.Context, id string, condition *request.Condition) (*image.Image, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *Update) (*image.Image, error)
	Delete(ctx context.Context, id string, condition *request.Condition) (bool, error)
	Destroy(ctx context.Context, id string, condition *request.Condition) (bool, error)
}

type Update struct {
	Metadata          *image.Metadata    `json:"metadata,omitempty"`
	ContentID         *string            `json:"contentId,omitempty"`
	ContentIntent     *string            `json:"contentIntent,omitempty"`
	ContentAttributes *ContentAttributes `json:"contentAttributes,omitempty"`
	RenditionsID      *string            `json:"renditionsId,omitempty"`
	Rendition         *string            `json:"rendition,omitempty"`
}

func NewUpdate() *Update {
	return &Update{}
}

func (u *Update) Validate(validator structure.Validator) {
	contentIntentValidator := validator.String("contentIntent", u.ContentIntent)
	contentAttributesValidator := validator.WithReference("contentAttributes")
	renditionsIDValidator := validator.String("renditionsId", u.RenditionsID)
	renditionValidator := validator.String("rendition", u.Rendition)

	if u.Metadata != nil {
		u.Metadata.Validate(validator.WithReference("metadata"))
	}
	validator.String("contentId", u.ContentID).Using(image.ContentIDValidator)
	if u.ContentID != nil {
		contentIntentValidator.Exists().OneOf(image.ContentIntents()...)
		if u.ContentAttributes != nil {
			u.ContentAttributes.Validate(contentAttributesValidator)
		} else {
			contentAttributesValidator.ReportError(structureValidator.ErrorValueNotExists())
		}
		renditionsIDValidator.NotExists()
		renditionValidator.NotExists()
	} else {
		contentIntentValidator.NotExists()
		if u.ContentAttributes != nil {
			contentAttributesValidator.ReportError(structureValidator.ErrorValueExists())
		}
		renditionsIDValidator.Using(image.RenditionsIDValidator)
		if u.RenditionsID != nil {
			renditionValidator.Exists()
		}
		renditionValidator.NotEmpty()
	}
}

func (u *Update) IsEmpty() bool {
	return (u.Metadata == nil || u.Metadata.IsEmpty()) &&
		u.ContentID == nil && u.ContentIntent == nil && u.ContentAttributes == nil && u.RenditionsID == nil && u.Rendition == nil
}

type ContentAttributes struct {
	DigestMD5 *string `json:"digestMD5,omitempty"`
	MediaType *string `json:"mediaType,omitempty"`
	Width     *int    `json:"width,omitempty"`
	Height    *int    `json:"height,omitempty"`
	Size      *int    `json:"size,omitempty"`
}

func NewContentAttributes() *ContentAttributes {
	return &ContentAttributes{}
}

func (c *ContentAttributes) Validate(validator structure.Validator) {
	validator.String("digestMD5", c.DigestMD5).Exists().Using(crypto.Base64EncodedMD5HashValidator)
	validator.String("mediaType", c.MediaType).Exists().Using(image.MediaTypeValidator)
	validator.Int("width", c.Width).Exists().GreaterThan(0)
	validator.Int("height", c.Height).Exists().GreaterThan(0)
	validator.Int("size", c.Size).Exists().GreaterThan(0)
}
