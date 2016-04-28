package store

import "github.com/tidepool-org/platform/log"

// TODO: Consider adding Collection to NewSession

type Store interface {
	IsClosed() bool
	Close()
	GetStatus() interface{}
	NewSession(log log.Logger) (Session, error)
}

type Session interface {
	IsClosed() bool
	Close()
	Find(query Query, result interface{}) error
	FindAll(query Query, sort []string, filter Filter) Iterator
	Insert(d interface{}) error
	InsertAll(d ...interface{}) error
	Update(selector interface{}, d interface{}) error
	UpdateAll(selector interface{}, update interface{}) error
}

type Iterator interface {
	IsClosed() bool
	Close() error
	Next(result interface{}) bool
	All(result interface{}) error
	Err() error
}

type Query map[string]interface{}
type Filter map[string]bool
