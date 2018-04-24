package test

import (
	"math/rand"

	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/test"
)

func NewSourceParameter() string {
	return test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
}

func NewSourcePointer() string {
	sourcePointer := ""
	for index := 0; index <= rand.Intn(4); index++ {
		sourcePointer += "/" + test.NewVariableString(1, 8, test.CharsetAlphaNumeric)
	}
	return sourcePointer
}

func NewError() error {
	return errors.New(test.NewText(1, 64))
}

func NewMeta() interface{} {
	meta := map[string]interface{}{}
	for index := 0; index <= rand.Intn(2); index++ {
		meta[test.NewVariableString(1, 8, test.CharsetAlphaNumeric)] = test.NewText(1, 32)
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
