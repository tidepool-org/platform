package store

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

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
