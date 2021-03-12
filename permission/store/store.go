package store

import (
	"io"
)

type Store interface {
	NewPermissionsSession() PermissionsSession
}

type PermissionsSession interface {
	io.Closer
}
