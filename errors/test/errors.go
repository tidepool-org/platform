package test

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/onsi/gomega"
	"gopkg.in/mgo.v2/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/test"
)

func NewSourceParameter() string {
	return test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
}

func NewSourcePointer() string {
	sourcePointer := ""
	for index := 0; index <= rand.Intn(4); index++ {
		sourcePointer += "/" + test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)
	}
	return sourcePointer
}

func RandomError() error {
	return errors.New(test.RandomStringFromRange(1, 64))
}

func CloneError(err error) error {
	if err == nil {
		return nil
	}
	return err // TODO: Is there a reasonable way to clone an error?
}

func RandomSerializable() *errors.Serializable {
	return errors.NewSerializable(RandomError())
}

func CloneSerializable(serializable *errors.Serializable) *errors.Serializable {
	if serializable == nil {
		return nil
	}
	return errors.NewSerializable(CloneError(serializable.Error))
}

func NewObjectFromSerializable(serializable *errors.Serializable, objectFormat test.ObjectFormat) map[string]interface{} {
	if serializable == nil {
		return nil
	}

	object := map[string]interface{}{}

	switch objectFormat {
	case test.ObjectFormatBSON:
		if bites, err := bson.Marshal(serializable); err != nil {
			return nil
		} else if err = bson.Unmarshal(bites, &object); err != nil {
			return nil
		}
	case test.ObjectFormatJSON:
		if bites, err := json.Marshal(serializable); err != nil {
			return nil
		} else if err = json.Unmarshal(bites, &object); err != nil {
			return nil
		}
	default:
		return nil
	}

	return object
}

func NewMeta() interface{} {
	meta := map[string]interface{}{}
	for index := 0; index <= rand.Intn(2); index++ {
		meta[test.RandomStringFromRangeAndCharset(1, 8, test.CharsetAlphaNumeric)] = test.RandomStringFromRange(1, 32)
	}
	return meta
}

type source struct {
	parameter string
	pointer   string
}

func (s *source) Parameter() string {
	return s.parameter
}

func (s *source) Pointer() string {
	return s.pointer
}

func WithParameterSource(err error, parameter string) error {
	return errors.WithSource(err, &source{parameter: parameter})
}

func WithParameterSourceAndMeta(err error, parameter string, meta interface{}) error {
	return errors.WithMeta(errors.WithSource(err, &source{parameter: parameter}), meta)
}

func WithPointerSource(err error, pointer string) error {
	return errors.WithSource(err, &source{pointer: pointer})
}

func WithPointerSourceAndMeta(err error, pointer string, meta interface{}) error {
	return errors.WithMeta(errors.WithSource(err, &source{pointer: pointer}), meta)
}

func ExpectEqual(actualError error, expectedErrors ...error) {
	switch len(expectedErrors) {
	case 0:
		gomega.Expect(errors.Sanitize(actualError)).To(gomega.BeNil())
	case 1:
		gomega.Expect(errors.Sanitize(actualError)).To(gomega.Equal(errors.Sanitize(expectedErrors[0])))
	default:
		gomega.Expect(errors.Sanitize(actualError)).To(gomega.Equal(errors.Sanitize(errors.Append(expectedErrors...))))
	}
}

func ExpectErrorDetails(err error, code string, title string, detail string) {
	gomega.Expect(err).ToNot(gomega.BeNil())
	gomega.Expect(errors.Code(err)).To(gomega.Equal(code))
	gomega.Expect(errors.Cause(err)).To(gomega.Equal(err))
	bites, marshalErr := json.Marshal(errors.Sanitize(err))
	gomega.Expect(marshalErr).ToNot(gomega.HaveOccurred())
	gomega.Expect(bites).To(gomega.MatchJSON(fmt.Sprintf(`{"code": %q, "title": %q, "detail": %q}`, code, title, detail)))
}

func ExpectErrorJSON(err error, actualJSON []byte) {
	expectedJSON, err := json.Marshal(errors.Sanitize(err))
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(actualJSON).To(gomega.MatchJSON(expectedJSON))
}
