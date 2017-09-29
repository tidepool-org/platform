package user

type Permission map[string]interface{}
type Permissions map[string]Permission

const OwnerPermission = "root"
const CustodianPermission = "custodian"
const UploadPermission = "upload"
const ViewPermission = "view"
