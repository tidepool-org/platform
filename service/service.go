package service

import (
	"github.com/satori/go.uuid"
)

//GetUUID returns a new uuid
func GetUUID() string {
	return uuid.NewV4().String()
}
