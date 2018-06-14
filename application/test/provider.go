package test

import (
	"github.com/tidepool-org/platform/config"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/version"
)

type Provider struct {
	VersionReporterInvocations int
	VersionReporterStub        func() version.Reporter
	VersionReporterOutputs     []version.Reporter
	VersionReporterOutput      *version.Reporter
	ConfigReporterInvocations  int
	ConfigReporterStub         func() config.Reporter
	ConfigReporterOutputs      []config.Reporter
	ConfigReporterOutput       *config.Reporter
	LoggerInvocations          int
	LoggerStub                 func() log.Logger
	LoggerOutputs              []log.Logger
	LoggerOutput               *log.Logger
	PrefixInvocations          int
	PrefixStub                 func() string
	PrefixOutputs              []string
	PrefixOutput               *string
	NameInvocations            int
	NameStub                   func() string
	NameOutputs                []string
	NameOutput                 *string
	UserAgentInvocations       int
	UserAgentStub              func() string
	UserAgentOutputs           []string
	UserAgentOutput            *string
}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) VersionReporter() version.Reporter {
	p.VersionReporterInvocations++
	if p.VersionReporterStub != nil {
		return p.VersionReporterStub()
	}
	if len(p.VersionReporterOutputs) > 0 {
		output := p.VersionReporterOutputs[0]
		p.VersionReporterOutputs = p.VersionReporterOutputs[1:]
		return output
	}
	if p.VersionReporterOutput != nil {
		return *p.VersionReporterOutput
	}
	panic("VersionReporter has no output")
}

func (p *Provider) ConfigReporter() config.Reporter {
	p.ConfigReporterInvocations++
	if p.ConfigReporterStub != nil {
		return p.ConfigReporterStub()
	}
	if len(p.ConfigReporterOutputs) > 0 {
		output := p.ConfigReporterOutputs[0]
		p.ConfigReporterOutputs = p.ConfigReporterOutputs[1:]
		return output
	}
	if p.ConfigReporterOutput != nil {
		return *p.ConfigReporterOutput
	}
	panic("ConfigReporter has no output")
}

func (p *Provider) Logger() log.Logger {
	p.LoggerInvocations++
	if p.LoggerStub != nil {
		return p.LoggerStub()
	}
	if len(p.LoggerOutputs) > 0 {
		output := p.LoggerOutputs[0]
		p.LoggerOutputs = p.LoggerOutputs[1:]
		return output
	}
	if p.LoggerOutput != nil {
		return *p.LoggerOutput
	}
	panic("Logger has no output")
}

func (p *Provider) Prefix() string {
	p.PrefixInvocations++
	if p.PrefixStub != nil {
		return p.PrefixStub()
	}
	if len(p.PrefixOutputs) > 0 {
		output := p.PrefixOutputs[0]
		p.PrefixOutputs = p.PrefixOutputs[1:]
		return output
	}
	if p.PrefixOutput != nil {
		return *p.PrefixOutput
	}
	panic("Prefix has no output")
}

func (p *Provider) Name() string {
	p.NameInvocations++
	if p.NameStub != nil {
		return p.NameStub()
	}
	if len(p.NameOutputs) > 0 {
		output := p.NameOutputs[0]
		p.NameOutputs = p.NameOutputs[1:]
		return output
	}
	if p.NameOutput != nil {
		return *p.NameOutput
	}
	panic("Name has no output")
}

func (p *Provider) UserAgent() string {
	p.UserAgentInvocations++
	if p.UserAgentStub != nil {
		return p.UserAgentStub()
	}
	if len(p.UserAgentOutputs) > 0 {
		output := p.UserAgentOutputs[0]
		p.UserAgentOutputs = p.UserAgentOutputs[1:]
		return output
	}
	if p.UserAgentOutput != nil {
		return *p.UserAgentOutput
	}
	panic("UserAgent has no output")
}

func (p *Provider) AssertOutputsEmpty() {
	if len(p.VersionReporterOutputs) > 0 {
		panic("VersionReporterOutputs is not empty")
	}
	if len(p.ConfigReporterOutputs) > 0 {
		panic("ConfigReporterOutputs is not empty")
	}
	if len(p.LoggerOutputs) > 0 {
		panic("LoggerOutputs is not empty")
	}
	if len(p.PrefixOutputs) > 0 {
		panic("PrefixOutputs is not empty")
	}
	if len(p.NameOutputs) > 0 {
		panic("NameOutputs is not empty")
	}
	if len(p.UserAgentOutputs) > 0 {
		panic("UserAgentOutputs is not empty")
	}
}
