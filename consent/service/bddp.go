package service

import (
	"context"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
)

type BigDataDonationProjectConfig struct {
	Disabled        bool   `envconfig:"TIDEPOOL_BIG_DATA_DONATION_PROJECT_SHARING_DISABLED"`
	RecipientUserID string `envconfig:"TIDEPOOL_BIG_DATA_DONATION_PROJECT_DATA_RECIPIENT_USER_ID"`
}

//go:generate mockgen -source=bddp.go -destination=test/bddp_mocks.go -package=test BigDataDonationProjectSharer
type BigDataDonationProjectSharer interface {
	Share(ctx context.Context, sharerUserID string) error
	Unshare(ctx context.Context, sharerUserID string) error
}

func NewBigDataDonationProjectSharer(config BigDataDonationProjectConfig, client permission.Client) (BigDataDonationProjectSharer, error) {
	if config.Disabled {
		return &disabledBigDataDonationProjectSharer{}, nil
	}
	if config.RecipientUserID == "" {
		return nil, errors.New("big data donation project data recipient user id is required")
	}

	return &bigDataDonationProjectSharer{
		config:           config,
		permissionClient: client,
	}, nil
}

type bigDataDonationProjectSharer struct {
	config           BigDataDonationProjectConfig
	permissionClient permission.Client
}

func (b *bigDataDonationProjectSharer) Share(ctx context.Context, sharerUserID string) error {
	return b.permissionClient.UpdateUserPermissions(ctx, sharerUserID, b.config.RecipientUserID, permission.Permissions{
		permission.Read: permission.Permission{},
	})
}

func (b *bigDataDonationProjectSharer) Unshare(ctx context.Context, sharerUserID string) error {
	return b.permissionClient.UpdateUserPermissions(ctx, sharerUserID, b.config.RecipientUserID, permission.Permissions{})
}

type disabledBigDataDonationProjectSharer struct{}

func (b *disabledBigDataDonationProjectSharer) Share(ctx context.Context, sharerUserID string) error {
	log.LoggerFromContext(ctx).WithFields(log.Fields{"sharerId": sharerUserID}).WithError(errors.New("sharing with big data donation project recipient is disabled")).Info("Share")
	return nil
}

func (b *disabledBigDataDonationProjectSharer) Unshare(ctx context.Context, sharerUserID string) error {
	log.LoggerFromContext(ctx).WithFields(log.Fields{"sharerId": sharerUserID}).WithError(errors.New("sharing with big data donation project recipient is disabled")).Info("Unshare")
	return nil
}
