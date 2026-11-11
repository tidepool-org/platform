// Package revision diffs and patches datums as BSON-shaped maps, so that a datum's history can be recorded as
// RFC 7386 JSON merge patches which Go and JavaScript clients can fold with standard libraries.
package revision

import (
	"reflect"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

const (
	// TagRevision is the struct tag that marks a datum field as outside the mutable projection. Its only value is
	// TagRevisionOmit.
	TagRevision     = "revision"
	TagRevisionOmit = "omit"
)

// serverOwnedKeys are the BSON keys of a stored datum that the server sets and a client never changes: every field
// of the base datum tagged `revision:"omit"`, and _id, which has no field. They are outside the mutable projection:
// never diffed, never patched, never set or unset by an update built here.
var serverOwnedKeys = mustOmittedKeys(reflect.TypeOf(types.Base{}), "_id")

func isServerOwnedKey(key string) bool {
	return slices.Contains(serverOwnedKeys, key)
}

// omittedKeys returns the BSON names of the struct's fields tagged `revision:"omit"`, followed by the extra keys. A
// revision tag with any other value is an error.
func omittedKeys(typ reflect.Type, extra ...string) ([]string, error) {
	if typ.Kind() != reflect.Struct {
		return nil, errors.Newf("type %s is not a struct", typ)
	}
	keys := []string{}
	for index := range typ.NumField() {
		field := typ.Field(index)
		value, tagged := field.Tag.Lookup(TagRevision)
		if !tagged {
			continue
		} else if value != TagRevisionOmit {
			return nil, errors.Newf("field %s.%s has revision tag %q, only %q is allowed", typ, field.Name, value, TagRevisionOmit)
		}
		name, _, _ := strings.Cut(field.Tag.Get("bson"), ",")
		if name == "" || name == "-" {
			return nil, errors.Newf("field %s.%s has no bson name", typ, field.Name)
		}
		keys = append(keys, name)
	}
	return append(keys, extra...), nil
}

func mustOmittedKeys(typ reflect.Type, extra ...string) []string {
	keys, err := omittedKeys(typ, extra...)
	if err != nil {
		panic(err)
	}
	return keys
}

// ToMap returns the datum as the map its stored document decodes to.
func ToMap(datum any) (map[string]any, error) {
	if datum == nil {
		return nil, errors.New("datum is missing")
	}
	bytes, err := bson.Marshal(datum)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal datum")
	}
	return storeStructuredMongo.DecodeDocument(bytes)
}

// Project returns a shallow copy of the document without the server-owned keys.
func Project(document map[string]any) map[string]any {
	projection := make(map[string]any, len(document))
	for key, value := range document {
		if !isServerOwnedKey(key) {
			projection[key] = value
		}
	}
	return projection
}

// Canonicalize returns a deep copy in which null members are dropped from every map, BSON map and array types are
// plain maps and slices, integers are int64, and times are UTC time.Time, so that two documents that decode
// differently but mean the same compare as equal, and a merge patch can express every stored value.
func Canonicalize(document map[string]any) map[string]any {
	return canonicalValue(document).(map[string]any)
}

func canonicalValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case map[string]any:
		return canonicalMap(typed)
	case primitive.M:
		return canonicalMap(typed)
	case primitive.D:
		result := make(map[string]any, len(typed))
		for _, element := range typed {
			if element.Value != nil {
				result[element.Key] = canonicalValue(element.Value)
			}
		}
		return result
	case []any:
		return canonicalSlice(typed)
	case primitive.A:
		return canonicalSlice(typed)
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case uint:
		return int64(typed)
	case uint8:
		return int64(typed)
	case uint16:
		return int64(typed)
	case uint32:
		return int64(typed)
	case uint64:
		return int64(typed)
	case float32:
		return float64(typed)
	case primitive.DateTime:
		return typed.Time().UTC()
	case time.Time:
		return typed.UTC()
	default:
		return value
	}
}

func canonicalMap(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, member := range value {
		if member != nil {
			result[key] = canonicalValue(member)
		}
	}
	return result
}

func canonicalSlice(value []any) []any {
	result := make([]any, len(value))
	for index, element := range value {
		result[index] = canonicalValue(element)
	}
	return result
}

// CreatePatch returns the RFC 7386 merge patch that turns from into to, empty when they are equal. Both must be
// canonical: values are compared structurally, and a null member in to cannot be expressed by a merge patch and is
// treated as absent.
func CreatePatch(from map[string]any, to map[string]any) map[string]any {
	patch, _ := diff(from, to)
	if patch == nil {
		return map[string]any{}
	}
	return patch.(map[string]any)
}

// diff returns the RFC 7386 merge patch that turns from into to and whether they differ. Objects are compared
// member by member; arrays and scalars are replaced whole.
func diff(from any, to any) (any, bool) {
	toMap, toIsMap := to.(map[string]any)
	if !toIsMap {
		if reflect.DeepEqual(from, to) {
			return nil, false
		}
		return deepCopy(to), true
	}
	fromMap, fromIsMap := from.(map[string]any)
	if !fromIsMap {
		fromMap = map[string]any{}
	}
	patch := map[string]any{}
	for key := range fromMap {
		if toValue, exists := toMap[key]; !exists || toValue == nil {
			patch[key] = nil
		}
	}
	for key, toValue := range toMap {
		if toValue == nil {
			continue
		}
		fromValue, exists := fromMap[key]
		if !exists {
			patch[key] = deepCopy(toValue)
		} else if memberPatch, changed := diff(fromValue, toValue); changed {
			patch[key] = memberPatch
		}
	}
	return patch, len(patch) > 0 || !fromIsMap
}

// Apply returns the target with the RFC 7386 merge patch applied. Neither argument is modified.
func Apply(target map[string]any, patch map[string]any) map[string]any {
	return merge(target, patch).(map[string]any)
}

// merge applies an RFC 7386 merge patch to any value: an object patch merges member by member, where a null member
// removes the key and a non-object target is replaced by an empty object first; any other patch replaces the target.
func merge(target any, patch any) any {
	patchMap, patchIsMap := patch.(map[string]any)
	if !patchIsMap {
		return deepCopy(patch)
	}
	result := map[string]any{}
	if targetMap, targetIsMap := target.(map[string]any); targetIsMap {
		for key, value := range targetMap {
			result[key] = deepCopy(value)
		}
	}
	for key, value := range patchMap {
		if value == nil {
			delete(result, key)
		} else {
			result[key] = merge(result[key], value)
		}
	}
	return result
}

// changedKeys returns the top-level keys a patch sets and the keys it removes, each sorted.
func changedKeys(patch map[string]any) (set []string, unset []string) {
	for key, value := range patch {
		if value == nil {
			unset = append(unset, key)
		} else {
			set = append(set, key)
		}
	}
	slices.Sort(set)
	slices.Sort(unset)
	return set, unset
}

// NewUpdate returns the Mongo update that applies a patch to a stored document: $set with the whole value of every
// top-level key the patch sets, read from result, the document after the patch, and $unset with the keys it
// removes. Either operator is left out when it has no keys.
func NewUpdate(patch map[string]any, result map[string]any) (bson.M, error) {
	setKeys, unsetKeys := changedKeys(patch)
	for _, key := range append(slices.Clone(setKeys), unsetKeys...) {
		if isServerOwnedKey(key) {
			return nil, errors.Newf("key %q is server owned", key)
		}
	}
	update := bson.M{}
	if len(setKeys) > 0 {
		set := bson.M{}
		for _, key := range setKeys {
			value, exists := result[key]
			if !exists {
				return nil, errors.Newf("key %q is missing from result", key)
			}
			set[key] = value
		}
		update["$set"] = set
	}
	if len(unsetKeys) > 0 {
		unset := bson.M{}
		for _, key := range unsetKeys {
			unset[key] = 1
		}
		update["$unset"] = unset
	}
	return update, nil
}

func deepCopy(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, member := range typed {
			result[key] = deepCopy(member)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, element := range typed {
			result[index] = deepCopy(element)
		}
		return result
	default:
		return value
	}
}
