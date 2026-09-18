// Package combined runs all platform HTTP APIs and their workers in one process.
package combined

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/tidepool-org/platform/application"
	authService "github.com/tidepool-org/platform/auth/service/service"
	blobService "github.com/tidepool-org/platform/blob/service"
	dataService "github.com/tidepool-org/platform/data/service/service"
	"github.com/tidepool-org/platform/service/server"
	taskService "github.com/tidepool-org/platform/task/service/service"
)

type hostedRunner interface {
	application.Runner
	SetServerFactory(server.Factory)
}

type component struct {
	name        string
	runner      hostedRunner
	endpoint    *endpoint
	done        chan struct{}
	initialized bool
}

type Application struct {
	*application.Application
	components []*component
	config     *server.Config
	server     *http.Server
	listener   net.Listener
	mu         sync.Mutex
	terminated bool
	stopped    chan struct{}
	running    sync.WaitGroup
}

func New() *Application {
	app := &Application{Application: application.New(), stopped: make(chan struct{})}
	app.add("auth", authService.New())
	app.add("blob", blobService.New())
	app.add("data", dataService.NewStandard())
	app.add("prescription", &prescriptionRunner{})
	app.add("task", taskService.New())
	return app
}

func (a *Application) add(name string, runner hostedRunner) {
	endpoint := newEndpoint()
	runner.SetServerFactory(endpoint.factory)
	a.components = append(a.components, &component{
		name: name, runner: runner, endpoint: endpoint, done: make(chan struct{}),
	})
}

func (a *Application) Initialize(provider application.Provider) error {
	if err := a.Application.Initialize(provider); err != nil {
		return err
	}
	a.config = server.NewConfig()
	if err := a.config.Load(provider.ConfigReporter().WithScopes("server")); err != nil {
		return err
	}
	if err := a.config.Validate(); err != nil {
		return err
	}
	a.server = &http.Server{Addr: a.config.Address}
	providers := make([]application.Provider, 0, len(a.components))
	consumerGroups := map[string]string{}
	for _, component := range a.components {
		componentProvider, err := application.NewNamedProvider(provider.Prefix(), component.name, "service")
		if err != nil {
			return fmt.Errorf("%s provider: %w", component.name, err)
		}
		if component.name == "auth" || component.name == "blob" || component.name == "data" {
			group := componentProvider.ConfigReporter().WithScopes("events").GetWithDefault("consumer_group", "")
			if group == "" {
				return fmt.Errorf("%s events consumer_group is required; configure its existing Kafka consumer group in the component's service events scope", component.name)
			}
			if owner, exists := consumerGroups[group]; exists {
				return fmt.Errorf("%s and %s must use different events consumer groups", owner, component.name)
			}
			consumerGroups[group] = component.name
		}
		providers = append(providers, componentProvider)
	}
	for i, component := range a.components {
		// Terminate also cleans up partially initialized components.
		component.initialized = true
		if err := component.runner.Initialize(providers[i]); err != nil {
			return fmt.Errorf("initialize %s: %w", component.name, err)
		}
	}
	return nil
}

func (a *Application) Run() error {
	// Serialize startup and shutdown: every Run has either reached Serve or
	// returned before Terminate may release a component's dependencies.
	a.mu.Lock()
	if a.terminated {
		a.mu.Unlock()
		return nil
	}
	if a.server == nil {
		a.mu.Unlock()
		return fmt.Errorf("combined service not initialized")
	}
	errs := make(chan error, len(a.components)+1)
	for _, component := range a.components {
		a.running.Add(1)
		go func() {
			defer a.running.Done()
			defer close(component.done)
			err := component.runner.Run()
			if err == nil {
				err = fmt.Errorf("stopped unexpectedly")
			}
			errs <- fmt.Errorf("%s: %w", component.name, err)
		}()
	}
	var sources []routeSource
	for _, component := range a.components {
		select {
		case <-component.endpoint.ready:
		case <-component.done:
		}
		sources = append(sources, routeSource{component.name, component.endpoint.router})
	}
	select {
	case err := <-errs:
		a.mu.Unlock()
		return err
	default:
	}
	handler, err := newHandler(a.VersionReporter().Long(), sources)
	if err != nil {
		a.mu.Unlock()
		return err
	}
	a.server.Handler = handler
	a.listener, err = net.Listen("tcp", a.config.Address)
	if err != nil {
		a.mu.Unlock()
		return err
	}
	a.running.Add(1)
	go func() {
		defer a.running.Done()
		if a.config.TLS {
			errs <- a.server.ServeTLS(a.listener, a.config.TLSCertificateFile, a.config.TLSKeyFile)
		} else {
			errs <- a.server.Serve(a.listener)
		}
	}()
	a.Logger().Infof("Serving combined platform API at %s", a.config.Address)
	a.mu.Unlock()
	select {
	case err := <-errs:
		select {
		case <-a.stopped:
			return nil
		default:
		}
		return err
	case <-a.stopped:
		return nil
	}
}

func (a *Application) Terminate() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.terminated {
		return
	}
	a.terminated = true
	close(a.stopped)
	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.config.Timeout)
		defer cancel()
		if err := a.server.Shutdown(ctx); err != nil {
			a.Logger().Errorf("Shutting down combined API: %v", err)
			_ = a.server.Close()
		}
	}
	for i := len(a.components) - 1; i >= 0; i-- {
		component := a.components[i]
		if component.initialized {
			component.runner.Terminate()
		}
		_ = component.endpoint.Shutdown()
	}
	a.running.Wait()
	a.Application.Terminate()
}
