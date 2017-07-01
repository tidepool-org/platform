package store

import "github.com/tidepool-org/platform/log"

type Store interface {
	IsClosed() bool
	Close()

	GetStatus() interface{}
}

type Session interface {
	IsClosed() bool
	Close()

	Logger() log.Logger

	SetAgent(agent Agent)
}

type Agent interface {
	IsServer() bool
	UserID() string
}
