package store

type Store interface {
	IsClosed() bool
	Close()

	GetStatus() interface{}
}

type Session interface {
	IsClosed() bool
	Close()

	SetAgent(agent Agent)
}

type Agent interface {
	IsServer() bool
	UserID() string
}
