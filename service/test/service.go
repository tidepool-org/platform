package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	authTest "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/config"
	configTest "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/version"
)

func NewUserID() string {
	return test.RandomStringFromRangeAndCharset(10, 10, test.CharsetHexadecimalLowercase)
}

type Service struct {
	VersionReporterInvocations int
	VersionReporterImpl        version.Reporter
	ConfigReporterInvocations  int
	ConfigReporterImpl         *configTest.Reporter
	LoggerInvocations          int
	LoggerImpl                 log.Logger
	SecretInvocations          int
	SecretOutputs              []string
	AuthClientInvocations      int
	AuthClientImpl             *authTest.Client
}

func NewService() *Service {
	versionReporter, _ := version.NewReporter(test.RandomStringFromRangeAndCharset(4, 4, test.CharsetAlphaNumeric), test.RandomStringFromRangeAndCharset(8, 8, test.CharsetAlphaNumeric), test.RandomStringFromRangeAndCharset(32, 32, test.CharsetAlphaNumeric))
	return &Service{
		VersionReporterImpl: versionReporter,
		ConfigReporterImpl:  configTest.NewReporter(),
		LoggerImpl:          nullLog.NewLogger(),
		AuthClientImpl:      authTest.NewClient(),
	}
}

func (s *Service) VersionReporter() version.Reporter {
	s.VersionReporterInvocations++

	return s.VersionReporterImpl
}

func (s *Service) ConfigReporter() config.Reporter {
	s.ConfigReporterInvocations++

	return s.ConfigReporterImpl
}

func (s *Service) Logger() log.Logger {
	s.LoggerInvocations++

	return s.LoggerImpl
}

func (s *Service) Secret() string {
	s.SecretInvocations++

	gomega.Expect(s.SecretOutputs).ToNot(gomega.BeEmpty())

	output := s.SecretOutputs[0]
	s.SecretOutputs = s.SecretOutputs[1:]
	return output
}

func (s *Service) AuthClient() auth.Client {
	s.AuthClientInvocations++

	return s.AuthClientImpl
}

func (s *Service) Expectations() {
	gomega.Expect(s.SecretOutputs).To(gomega.BeEmpty())
	s.AuthClientImpl.AssertOutputsEmpty()
}
