package data

type Inspector interface {
	GetProperty(key string) *string
	NewMissingPropertyError(key string) error
	NewInvalidPropertyError(key string, value string, allowedValues []string) error
}

type Factory interface {
	New(inspector Inspector) (Datum, error)
	Init(inspector Inspector) (Datum, error)
}
