package task

import "github.com/tidepool-org/platform/auth"

type Client interface {
	GetStatus(ctx auth.Context) (*Status, error)
}

type Status struct {
	Version   string
	Server    interface{}
	TaskStore interface{}
}
