package mongo

import (
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"

	"github.com/tidepool-org/platform/errors"
)

// plainRegistry decodes every embedded document as map[string]any, every array as []any and every datetime as
// time.Time wherever the Go type leaves the BSON type open, so a decoded document has one plain shape at every
// depth and marshals to JSON as a client sees it.
var plainRegistry = newPlainRegistry()

func newPlainRegistry() *bsoncodec.Registry {
	registry := bson.NewRegistry()
	registry.RegisterTypeMapEntry(bson.TypeEmbeddedDocument, reflect.TypeOf(map[string]any{}))
	registry.RegisterTypeMapEntry(bson.TypeArray, reflect.TypeOf([]any{}))
	registry.RegisterTypeMapEntry(bson.TypeDateTime, reflect.TypeOf(time.Time{}))
	return registry
}

// DecodeDocument returns a BSON document as a map of plain values: nested documents as map[string]any, arrays as
// []any and datetimes as time.Time at every depth.
func DecodeDocument(document []byte) (map[string]any, error) {
	result := map[string]any{}
	if err := UnmarshalDocument(document, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UnmarshalDocument decodes a BSON document into value, with plain values wherever value leaves the type open.
func UnmarshalDocument(document []byte, value any) error {
	decoder, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(document))
	if err != nil {
		return errors.Wrap(err, "unable to create decoder")
	}
	if err = decoder.SetRegistry(plainRegistry); err != nil {
		return errors.Wrap(err, "unable to set decoder registry")
	}
	if err = decoder.Decode(value); err != nil {
		return errors.Wrap(err, "unable to decode document")
	}
	return nil
}
