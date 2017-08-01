package id

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func New() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}
