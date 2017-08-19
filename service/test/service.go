package test

import (
	"github.com/tidepool-org/platform/auth"
	testAuth "github.com/tidepool-org/platform/auth/test"
	"github.com/tidepool-org/platform/config"
	testConfig "github.com/tidepool-org/platform/config/test"
	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/test"
	"github.com/tidepool-org/platform/version"
)

type Service struct {
	*test.Mock
	VersionReporterInvocations int
	VersionReporterImpl        version.Reporter
	ConfigReporterInvocations  int
	ConfigReporterImpl         *testConfig.Reporter
	LoggerInvocations          int
	LoggerImpl                 log.Logger
	AuthClientInvocations      int
	AuthClientImpl             *testAuth.Client
}

func NewService() *Service {
	versionReporter, _ := version.NewReporter(id.New(), id.New(), id.New())
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

func (s *Service) AuthClient() auth.Client {
	s.AuthClientInvocations++

	return s.AuthClientImpl
}

func (s *Service) UnusedOutputsCount() int {
	return s.AuthClientImpl.UnusedOutputsCount()
}
