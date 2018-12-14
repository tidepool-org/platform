package unstructured

import (
	"context"
	"io"
	"strings"

	"github.com/tidepool-org/platform/errors"
	storeUnstructured "github.com/tidepool-org/platform/store/unstructured"
)

type Store interface {
	PutContent(ctx context.Context, userID string, imageID string, contentID string, contentIntent string, reader io.Reader, options *storeUnstructured.Options) error
	GetContent(ctx context.Context, userID string, imageID string, contentID string, contentIntent string) (io.ReadCloser, error)
	DeleteContent(ctx context.Context, userID string, imageID string, contentID string) error

	PutRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string, reader io.Reader, options *storeUnstructured.Options) error
	GetRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string) (io.ReadCloser, error)
	DeleteRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string) error

	Delete(ctx context.Context, userID string, imageID string) error
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

func (s *StoreImpl) PutContent(ctx context.Context, userID string, imageID string, contentID string, contentIntent string, reader io.Reader, options *storeUnstructured.Options) error {
	if err := s.store.Put(ctx, asKey(userID, imageID, "content", contentID, contentIntent), reader, options); err != nil {
		return errors.Wrap(err, "unable to put image content")
	}
	return nil
}

func (s *StoreImpl) GetContent(ctx context.Context, userID string, imageID string, contentID string, contentIntent string) (io.ReadCloser, error) {
	reader, err := s.store.Get(ctx, asKey(userID, imageID, "content", contentID, contentIntent))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get image content")
	}
	return reader, nil
}

func (s *StoreImpl) DeleteContent(ctx context.Context, userID string, imageID string, contentID string) error {
	if err := s.store.DeleteDirectory(ctx, asKey(userID, imageID, "content", contentID)); err != nil {
		return errors.Wrap(err, "unable to delete all image content")
	}
	return nil
}

func (s *StoreImpl) PutRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string, reader io.Reader, options *storeUnstructured.Options) error {
	if err := s.store.Put(ctx, asKey(userID, imageID, "content", contentID, "renditions", renditionsID, rendition), reader, options); err != nil {
		return errors.Wrap(err, "unable to put image rendition content")
	}
	return nil
}

func (s *StoreImpl) GetRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string, rendition string) (io.ReadCloser, error) {
	reader, err := s.store.Get(ctx, asKey(userID, imageID, "content", contentID, "renditions", renditionsID, rendition))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get image rendition content")
	}
	return reader, nil
}

func (s *StoreImpl) DeleteRenditionContent(ctx context.Context, userID string, imageID string, contentID string, renditionsID string) error {
	if err := s.store.DeleteDirectory(ctx, asKey(userID, imageID, "content", contentID, "renditions", renditionsID)); err != nil {
		return errors.Wrap(err, "unable to delete image rendition content")
	}
	return nil
}

func (s *StoreImpl) Delete(ctx context.Context, userID string, imageID string) error {
	if err := s.store.DeleteDirectory(ctx, asKey(userID, imageID)); err != nil {
		return errors.Wrap(err, "unable to delete image")
	}
	return nil
}

func (s *StoreImpl) DeleteAll(ctx context.Context, userID string) error {
	if err := s.store.DeleteDirectory(ctx, asKey(userID)); err != nil {
		return errors.Wrap(err, "unable to delete all images")
	}
	return nil
}

func asKey(parts ...string) string {
	return strings.Join(parts, "/")
}
