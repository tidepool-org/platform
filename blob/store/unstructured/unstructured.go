package unstructured

import (
	"context"
	"fmt"
	"io"

	"github.com/tidepool-org/platform/errors"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

type Store interface {
	Exists(ctx context.Context, userID string, id string) (bool, error)
	Put(ctx context.Context, userID string, id string, reader io.Reader, options *storeUnstructured.Options) error
	Get(ctx context.Context, userID string, id string) (io.ReadCloser, error)
	Delete(ctx context.Context, userID string, id string) (bool, error)
}

type StoreImpl struct {
	store storeUnstructured.Store
}

func NewStore(store storeUnstructured.Store) (*StoreImpl, error) {
	if store == nil {
		return nil, errors.New("store is missing")
	}

	return &StoreImpl{
		store: store,
	}, nil
}

func (s *StoreImpl) Exists(ctx context.Context, userID string, id string) (bool, error) {
	exists, err := s.store.Exists(ctx, asKey(userID, id))
	if err != nil {
		return false, errors.Wrap(err, "unable to exists blob")
	}
	return exists, nil
}

func (s *StoreImpl) Put(ctx context.Context, userID string, id string, reader io.Reader, options *storeUnstructured.Options) error {
	if err := s.store.Put(ctx, asKey(userID, id), reader, options); err != nil {
		return errors.Wrap(err, "unable to put blob")
	}
	return nil
}

func (s *StoreImpl) Get(ctx context.Context, userID string, id string) (io.ReadCloser, error) {
	reader, err := s.store.Get(ctx, asKey(userID, id))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get blob")
	}
	return reader, nil
}

func (s *StoreImpl) Delete(ctx context.Context, userID string, id string) (bool, error) {
	deleted, err := s.store.Delete(ctx, asKey(userID, id))
	if err != nil {
		return false, errors.Wrap(err, "unable to delete blob")
	}
	return deleted, nil
}

func asKey(userID string, id string) string {
	return fmt.Sprintf("%s/%s/%s", userID, id, id)
}
