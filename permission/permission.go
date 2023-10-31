package permission

import (
	"github.com/mdblp/go-json-rest/rest"
)

type Client interface {
	GetUserPermissions(req *rest.Request, targetUserID string) (bool, error)
	GetPatientPermissions(req *rest.Request) (bool, string, error)
}
