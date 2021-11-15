package permission

import (
	"github.com/mdblp/go-json-rest/rest"
)

type Permission map[string]interface{}
type Permissions map[string]Permission

const Owner = "root"
const Custodian = "custodian"
const Write = "upload"
const Read = "view"

type Client interface {
	GetUserPermissions(req *rest.Request, targetUserID string) (bool, error)
}

func FixOwnerPermissions(permissions Permissions) Permissions {
	if ownerPermission, ok := permissions[Owner]; ok {
		if _, ok = permissions[Write]; !ok {
			permissions[Write] = ownerPermission
		}
		if _, ok = permissions[Read]; !ok {
			permissions[Read] = ownerPermission
		}
	}
	return permissions
}
