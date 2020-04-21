package application

import (
	"fmt"
	"go.uber.org/fx"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tidepool-org/platform/config"
	configEnv "github.com/tidepool-org/platform/config/env"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	logJson "github.com/tidepool-org/platform/log/json"
	"github.com/tidepool-org/platform/sync"
	"github.com/tidepool-org/platform/version"
)

var ProviderModule = fx.Provide(DefaultProvider)

type Provider interface {
	VersionReporter() version.Reporter
	ConfigReporter() config.Reporter
	Logger() log.Logger
	Prefix() string
	Name() string
	UserAgent() string
}

type ProviderResult struct {
	fx.Out

	Provider        Provider
	ConfigReporter  config.Reporter
	Logger          log.Logger
	UserAgent       string `name:"userAgent"`
	VersionReporter version.Reporter
}

type ProviderImpl struct {
	versionReporter version.Reporter
	configReporter  config.Reporter
	logger          log.Logger
	prefix          string
	name            string
	userAgent       string
}

func DefaultProvider() (*ProviderResult, error) {
	prvdr, err := NewProvider("TIDEPOOL", "service")
	if err != nil {
		return nil, err
	}

	return &ProviderResult{
		Provider:        prvdr,
		ConfigReporter:  prvdr.configReporter,
		Logger:          prvdr.logger,
		UserAgent:       prvdr.userAgent,
		VersionReporter: prvdr.versionReporter,
	}, nil
}

func NewProvider(prefix string, scopes ...string) (*ProviderImpl, error) {
	if prefix == "" {
		return nil, errors.New("prefix is missing")
	}

	versionReporter, err := NewVersionReporter()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create version reporter")
	}

	configReporter, err := configEnv.NewReporter(prefix)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create config reporter")
	}

	name := filepath.Base(os.Args[0])
	if strings.EqualFold(name, "debug") {
		name = configReporter.WithScopes(name).GetWithDefault("name", name)
	}

	configReporter = configReporter.WithScopes(name).WithScopes(scopes...)

	writer, err := sync.NewWriter(os.Stdout)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create writer")
	}

	level := configReporter.WithScopes("logger").GetWithDefault("level", "warn")

	logger, err := logJson.NewLogger(writer, log.DefaultLevelRanks(), log.Level(level))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create logger")
	}

	logger = logger.WithField("process", map[string]interface{}{
		"name":    name,
		"id":      os.Getpid(),
		"version": versionReporter.Short(),
	})

	logger.Infof("Logger level is %s", level)

	userAgent := fmt.Sprintf("%s-%s/%s", userAgentTitle(prefix), userAgentTitle(name), versionReporter.Base())

	return &ProviderImpl{
		versionReporter: versionReporter,
		configReporter:  configReporter,
		logger:          logger,
		prefix:          prefix,
		name:            name,
		userAgent:       userAgent,
	}, nil
}

func (p *ProviderImpl) VersionReporter() version.Reporter {
	return p.versionReporter
}

func (p *ProviderImpl) ConfigReporter() config.Reporter {
	return p.configReporter
}

func (p *ProviderImpl) Logger() log.Logger {
	return p.logger
}

func (p *ProviderImpl) Prefix() string {
	return p.prefix
}

func (p *ProviderImpl) Name() string {
	return p.name
}

func (p *ProviderImpl) UserAgent() string {
	return p.userAgent
}

var userAgentTitleExpression = regexp.MustCompile("[^a-zA-Z0-9]+")

func userAgentTitle(s string) string {
	return strings.Replace(strings.Title(strings.ToLower(userAgentTitleExpression.ReplaceAllString(s, " "))), " ", "", -1)
}
