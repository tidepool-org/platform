package test

import (
	"context"
	"maps"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/work"
)

// Implementation of work.Provider.
type Provider struct {
	ctx    context.Context
	Fields log.Fields
}

func NewProvider(ctx context.Context) *Provider {
	return &Provider{
		ctx:    ctx,
		Fields: log.Fields{},
	}
}

func (p *Provider) Context() context.Context {
	return p.ctx
}

func (p *Provider) AddFieldToContext(key string, value any) {
	p.Fields[key] = value
}

func (p *Provider) AddFieldsToContext(fields log.Fields) {
	maps.Copy(p.Fields, fields)
}

func (p *Provider) Failing(err error) *work.ProcessResult {
	return work.NewProcessResultFailing(work.FailingUpdate{FailingError: errors.Serializable{Error: err}})
}

func (p *Provider) Failed(err error) *work.ProcessResult {
	return work.NewProcessResultFailed(work.FailedUpdate{FailedError: errors.Serializable{Error: err}})
}

type MockMetadata struct {
	Mock *string `json:"mock,omitempty" bson:"mock,omitempty"`
	Any  any     `json:"any,omitempty" bson:"any,omitempty"` // Used to test encoding errors
}

func (m *MockMetadata) Parse(parser structure.ObjectParser) {
	m.Mock = parser.String("mock")
}

func (m *MockMetadata) Validate(validator structure.Validator) {
	validator.String("mock", m.Mock).NotEmpty()
}

func (m *MockMetadata) AsObject() map[string]any {
	object := map[string]any{}
	if m.Mock != nil {
		object["mock"] = *m.Mock
	}
	return object
}

func RandomMockMetadata(options ...test.Option) *MockMetadata {
	return &MockMetadata{
		Mock: test.RandomOptional(test.RandomString, options...),
	}
}

func CloneMockMetadata(datum *MockMetadata) *MockMetadata {
	if datum == nil {
		return nil
	}
	return &MockMetadata{
		Mock: pointer.Clone(datum.Mock),
	}
}

func NewObjectFromMockMetadata(datum *MockMetadata, objectFormat test.ObjectFormat) map[string]any {
	if datum == nil {
		return nil
	}
	object := map[string]any{}
	if datum.Mock != nil {
		object["mock"] = test.NewObjectFromString(*datum.Mock, objectFormat)
	}
	return object
}
