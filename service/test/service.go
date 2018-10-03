package test

import (
	"github.com/onsi/gomega"

	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/config"
	testConfig "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/version"
)

func NewUserID() string {
	return test.NewString(10, test.CharsetHexidecimalLowercase)
}

type Service struct {
	*test.Mock
	VersionReporterInvocations int
	VersionReporterImpl        version.Reporter
	ConfigReporterInvocations  int
	ConfigReporterImpl         *testConfig.Reporter
	LoggerInvocations          int
	LoggerImpl                 log.Logger
	SecretInvocations          int
	SecretOutputs              []string
	AuthClientInvocations      int
	AuthClientImpl             *testAuth.Client
}

func NewService() *Service {
	versionReporter, _ := version.NewReporter(test.NewString(4, test.CharsetAlphaNumeric), test.NewString(8, test.CharsetAlphaNumeric), test.NewString(32, test.CharsetAlphaNumeric))
	return &Service{
		Mock:                test.NewMock(),
		VersionReporterImpl: versionReporter,
		ConfigReporterImpl:  testConfig.NewReporter(),
		LoggerImpl:          nullLog.NewLogger(),
		AuthClientImpl:      testAuth.NewClient(),
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
	s.Mock.Expectations()
	s.ConfigReporterImpl.Expectations()
	gomega.Expect(s.SecretOutputs).To(gomega.BeEmpty())
	s.AuthClientImpl.Expectations()
}
