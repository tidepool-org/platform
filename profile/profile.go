package profile

import "time"

type Profile struct {
	UserID       *string    `json:"userId,omitempty" bson:"userId,omitempty"`
	Value        *string    `json:"-" bson:"value,omitempty"`
	CreatedTime  *time.Time `json:"createdTime,omitempty" bson:"createdTime,omitempty"`
	ModifiedTime *time.Time `json:"modifiedTime,omitempty" bson:"modifiedTime,omitempty"`
	DeletedTime  *time.Time `json:"deletedTime,omitempty" bson:"deletedTime,omitempty"`
	Revision     *int       `json:"revision,omitempty" bson:"revision,omitempty"`

	// HACK: Pull out FullName while Value is encoded as a JSON string
	FullName *string `json:"fullName" bson:"-"`
}

type ProfileArray []*Profile
