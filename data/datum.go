package data

type Datum interface {
	Init()

	Meta() interface{}

	Parse(parser ObjectParser) error
	Validate(validator Validator) error
	Normalize(normalizer Normalizer) error

	IdentityFields() ([]string, error)

	GetPayload() *map[string]interface{}

	SetUserID(userID string)
	SetDatasetID(datasetID string)
	SetActive(active bool)
	SetCreatedTime(createdTime string)
	SetCreatedUserID(createdUserID string)
	SetModifiedTime(modifiedTime string)
	SetModifiedUserID(modifiedUserID string)
	SetDeletedTime(deletedTime string)
	SetDeletedUserID(deletedUserID string)

	DeduplicatorDescriptor() *DeduplicatorDescriptor
	SetDeduplicatorDescriptor(deduplicatorDescriptor *DeduplicatorDescriptor)
}
