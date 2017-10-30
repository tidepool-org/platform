package store

type Store interface {
	IsClosed() bool
	Close()

	Status() interface{}
}

type Session interface {
	IsClosed() bool
	Close()

	EnsureIndexes() error
}
