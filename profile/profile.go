package profile

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

type Profile struct {
	ID    string `json:"-" bson:"_id,omitempty"`
	Value string `json:"-" bson:"value,omitempty"`

	FullName *string `json:"fullName" bson:"-"`
}
