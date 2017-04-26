package profile

type Profile struct {
	ID    string `json:"-" bson:"_id,omitempty"`
	Value string `json:"-" bson:"value,omitempty"`

	FullName *string `json:"fullName" bson:"-"`
}
