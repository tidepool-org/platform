package id

import (
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var Expression = regexp.MustCompile("^[0-9a-f]{32}$")

func New() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}
