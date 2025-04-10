package test

import (
	"context"

	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/devicetokens"
)

type DeviceTokenRepository struct {
	*authTest.DeviceTokenAccessor
	Documents []*devicetokens.Document
	Tokens    map[string][]*devicetokens.DeviceToken
	Error     error
}

func NewDeviceTokenRepository() *DeviceTokenRepository {
	return &DeviceTokenRepository{
		DeviceTokenAccessor: authTest.NewDeviceTokenAccessor(),
	}
}

func (r *DeviceTokenRepository) Expectations() {
	r.DeviceTokenAccessor.Expectations()
}

func (r *DeviceTokenRepository) GetAllByUserID(ctx context.Context, userID string) ([]*devicetokens.Document, error) {
	if r.Error != nil {
		return nil, r.Error
	}
	if tokens, ok := r.Tokens[userID]; ok {
		docs := make([]*devicetokens.Document, 0, len(tokens))
		for _, token := range tokens {
			docs = append(docs, &devicetokens.Document{DeviceToken: *token})
		}
		return docs, nil
	}
	return nil, nil
}

func (r *DeviceTokenRepository) Upsert(ctx context.Context, doc *devicetokens.Document) error {
	return nil
}

func (r *DeviceTokenRepository) EnsureIndexes() error {
	return nil
}
