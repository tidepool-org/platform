package test

import "github.com/tidepool-org/platform/user"

type Provider struct {
	UserClientInvocations int
	UserClientStub        func() user.Client
	UserClientOutputs     []user.Client
	UserClientOutput      *user.Client
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) UserClient() user.Client {
	p.UserClientInvocations++
	if p.UserClientStub != nil {
		return p.UserClientStub()
	}
	if len(p.UserClientOutputs) > 0 {
		output := p.UserClientOutputs[0]
		p.UserClientOutputs = p.UserClientOutputs[1:]
		return output
	}
	if p.UserClientOutput != nil {
		return *p.UserClientOutput
	}
	panic("UserClient has no output")
}

func (p *Provider) AssertOutputsEmpty() {
	if len(p.UserClientOutputs) > 0 {
		panic("UserClientOutputs is not empty")
	}
}
