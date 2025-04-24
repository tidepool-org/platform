package unstructured

import (
	"context"
	"io"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/tidepool-org/platform/errors"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

type Store interface {
	Exists(ctx context.Context, userID string, id string) (bool, error)
	Put(ctx context.Context, userID string, id string, reader io.Reader, options *storeUnstructured.Options) error
	Get(ctx context.Context, userID string, id string) (io.ReadCloser, error)
	GetMany(ctx context.Context, userID string, ids ...string) ([]io.ReadCloser, error)
	Delete(ctx context.Context, userID string, id string) (bool, error)
	DeleteAll(ctx context.Context, userID string) error
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
	exists, err := s.store.Exists(ctx, asKey(userID, id, id))
	if err != nil {
		return false, errors.Wrap(err, "unable to exists blob")
	}
	return exists, nil
}

func (s *StoreImpl) Put(ctx context.Context, userID string, id string, reader io.Reader, options *storeUnstructured.Options) error {
	// QUESTION: Why is the id repeated twice for the key?
	if err := s.store.Put(ctx, asKey(userID, id, id), reader, options); err != nil {
		return errors.Wrap(err, "unable to put blob")
	}
	return nil
}

func (s *StoreImpl) Get(ctx context.Context, userID string, id string) (io.ReadCloser, error) {
	reader, err := s.store.Get(ctx, asKey(userID, id, id))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get blob")
	}
	return reader, nil
}

func (s *StoreImpl) GetMany(ctx context.Context, userID string, ids ...string) ([]io.ReadCloser, error) {
	group, ctx := errgroup.WithContext(ctx)
	// maxReaders is just some arbritrary limit to how many simultaneous
	// requests we make to the underlying unstructured Store (s3 usually in
	// this case) chosen to be more than 1 so that we aren't doing a slower
	// sequential reading of each id while also avoiding reading everything at
	// once.
	maxReaders := 4
	group.SetLimit(maxReaders)

	readers := make([]io.ReadCloser, len(ids))
	for i, id := range ids {
		i, id := i, id
		group.Go(func() error {
			reader, err := s.Get(ctx, userID, id)
			if err != nil {
				return errors.Wrapf(err, "unable to get blob, id: %v", id)
			}
			readers[i] = reader
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return readers, nil
}

func (s *StoreImpl) Delete(ctx context.Context, userID string, id string) (bool, error) {
	deleted, err := s.store.Delete(ctx, asKey(userID, id, id))
	if err != nil {
		return false, errors.Wrap(err, "unable to delete blob")
	}
	return deleted, nil
}

func (s *StoreImpl) DeleteAll(ctx context.Context, userID string) error {
	if err := s.store.DeleteDirectory(ctx, asKey(userID)); err != nil {
		return errors.Wrap(err, "unable to delete all blobs")
	}
	return nil
}

func asKey(parts ...string) string {
	return strings.Join(parts, "/")
}
