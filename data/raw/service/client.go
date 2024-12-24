package service

import (
	"github.com/tidepool-org/platform/errors"
)

type Store interface{}

func NewClient(store Store) (*Client, error) {
	if store == nil {
		return nil, errors.New("store is missing")
	}
	return &Client{
		store: store,
	}, nil
}

type Client struct {
	store Store
}
