package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/structure"
	"github.com/tidepool-org/platform/test"
)

type Base struct {
	*test.Mock
	SourceInvocations        int
	SourceOutputs            []structure.Source
	MetaInvocations          int
	MetaOutputs              []interface{}
	ErrorInvocations         int
	ErrorOutputs             []error
	ReportErrorInvocations   int
	ReportErrorInputs        []error
	WithSourceInvocations    int
	WithSourceInputs         []structure.Source
	WithSourceOutputs        []structure.Base
	WithMetaInvocations      int
	WithMetaInputs           []interface{}
	WithMetaOutputs          []structure.Base
	WithReferenceInvocations int
	WithReferenceInputs      []string
	WithReferenceOutputs     []structure.Base
}

func NewBase() *Base {
	return &Base{
		Mock: test.NewMock(),
	}
}

func (b *Base) Source() structure.Source {
	b.SourceInvocations++

	gomega.Expect(b.SourceOutputs).ToNot(gomega.BeEmpty())

	output := b.SourceOutputs[0]
	b.SourceOutputs = b.SourceOutputs[1:]
	return output
}

func (b *Base) Meta() interface{} {
	b.MetaInvocations++

	gomega.Expect(b.MetaOutputs).ToNot(gomega.BeEmpty())

	output := b.MetaOutputs[0]
	b.MetaOutputs = b.MetaOutputs[1:]
	return output
}

func (b *Base) Error() error {
	b.ErrorInvocations++

	gomega.Expect(b.ErrorOutputs).ToNot(gomega.BeEmpty())

	output := b.ErrorOutputs[0]
	b.ErrorOutputs = b.ErrorOutputs[1:]
	return output
}

func (b *Base) ReportError(err error) {
	b.ReportErrorInvocations++

	b.ReportErrorInputs = append(b.ReportErrorInputs, err)
}

func (b *Base) WithSource(source structure.Source) structure.Base {
	b.WithSourceInvocations++

	b.WithSourceInputs = append(b.WithSourceInputs, source)

	gomega.Expect(b.WithSourceOutputs).ToNot(gomega.BeEmpty())

	output := b.WithSourceOutputs[0]
	b.WithSourceOutputs = b.WithSourceOutputs[1:]
	return output
}

func (b *Base) WithMeta(meta interface{}) structure.Base {
	b.WithMetaInvocations++

	b.WithMetaInputs = append(b.WithMetaInputs, meta)

	gomega.Expect(b.WithMetaOutputs).ToNot(gomega.BeEmpty())

	output := b.WithMetaOutputs[0]
	b.WithMetaOutputs = b.WithMetaOutputs[1:]
	return output
}

func (b *Base) WithReference(reference string) structure.Base {
	b.WithReferenceInvocations++

	b.WithReferenceInputs = append(b.WithReferenceInputs, reference)

	gomega.Expect(b.WithReferenceOutputs).ToNot(gomega.BeEmpty())

	output := b.WithReferenceOutputs[0]
	b.WithReferenceOutputs = b.WithReferenceOutputs[1:]
	return output
}

func (b *Base) Expectations() {
	b.Mock.Expectations()
	gomega.Expect(b.SourceOutputs).To(gomega.BeEmpty())
	gomega.Expect(b.MetaOutputs).To(gomega.BeEmpty())
	gomega.Expect(b.ErrorOutputs).To(gomega.BeEmpty())
	gomega.Expect(b.WithSourceOutputs).To(gomega.BeEmpty())
	gomega.Expect(b.WithMetaOutputs).To(gomega.BeEmpty())
	gomega.Expect(b.WithReferenceOutputs).To(gomega.BeEmpty())
}
