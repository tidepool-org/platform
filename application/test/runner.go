package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/test"
)

type Runner struct {
	*test.Mock
	InitializeInvocations int
	InitializeOutputs     []error
	TerminateInvocations  int
	RunInvocations        int
	RunOutputs            []error
}

func NewRunner() *Runner {
	return &Runner{
		Mock: test.NewMock(),
	}
}

func (r *Runner) Initialize() error {
	r.InitializeInvocations++

	gomega.Expect(r.InitializeOutputs).ToNot(gomega.BeEmpty())

	output := r.InitializeOutputs[0]
	r.InitializeOutputs = r.InitializeOutputs[1:]
	return output
}

func (r *Runner) Terminate() {
	r.TerminateInvocations++
}

func (r *Runner) Run() error {
	r.RunInvocations++

	gomega.Expect(r.RunOutputs).ToNot(gomega.BeEmpty())

	output := r.RunOutputs[0]
	r.RunOutputs = r.RunOutputs[1:]
	return output
}

func (r *Runner) Expectations() {
	r.Mock.Expectations()
	gomega.Expect(r.InitializeOutputs).To(gomega.BeEmpty())
	gomega.Expect(r.RunOutputs).To(gomega.BeEmpty())
}
