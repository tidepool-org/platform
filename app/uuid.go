package app

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func NewUUID() string {
	return uuid.NewV4().String()
}

func NewID() string {
	return strings.Replace(NewUUID(), "-", "", -1)
}
