package test

import "github.com/tidepool-org/platform/application"

type Runner struct {
	InitializeInvocations int
	InitializeInputs      []application.Provider
	InitializeStub        func(provider application.Provider) error
	InitializeOutputs     []error
	InitializeOutput      *error
	TerminateInvocations  int
	TerminateStub         func()
	RunInvocations        int
	RunStub               func() error
	RunOutputs            []error
	RunOutput             *error
}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Initialize(provider application.Provider) error {
	r.InitializeInvocations++
	r.InitializeInputs = append(r.InitializeInputs, provider)
	if r.InitializeStub != nil {
		return r.InitializeStub(provider)
	}
	if len(r.InitializeOutputs) > 0 {
		output := r.InitializeOutputs[0]
		r.InitializeOutputs = r.InitializeOutputs[1:]
		return output
	}
	if r.InitializeOutput != nil {
		return *r.InitializeOutput
	}
	panic("Initialize has no output")
}

func (r *Runner) Terminate() {
	r.TerminateInvocations++
	if r.TerminateStub != nil {
		r.TerminateStub()
	}
}

func (r *Runner) Run() error {
	r.RunInvocations++
	if r.RunStub != nil {
		return r.RunStub()
	}
	if len(r.RunOutputs) > 0 {
		output := r.RunOutputs[0]
		r.RunOutputs = r.RunOutputs[1:]
		return output
	}
	if r.RunOutput != nil {
		return *r.RunOutput
	}
	panic("Run has no output")
}

func (r *Runner) AssertOutputsEmpty() {
	if len(r.InitializeOutputs) > 0 {
		panic("InitializeOutputs is not empty")
	}
	if len(r.RunOutputs) > 0 {
		panic("RunOutputs is not empty")
	}
}
