package combined

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/application"
	serviceAPI "github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/service/server"
)

type stubRunner struct {
	factory             server.Factory
	server              server.Server
	provider            application.Provider
	initializationError error
	runError            error
	fail                chan error
	terminated          bool
	terminationOrder    *[]string
}

func (r *stubRunner) SetServerFactory(factory server.Factory) { r.factory = factory }
func (r *stubRunner) Initialize(provider application.Provider) error {
	r.provider = provider
	if r.initializationError != nil {
		return r.initializationError
	}
	api, err := serviceAPI.New(routeTestService{name: provider.Name()})
	if err != nil {
		return err
	}
	if err := api.InitializeRoutes(rest.Get("/"+provider.Name(), func(res rest.ResponseWriter, req *rest.Request) {
		res.WriteJson(map[string]string{"service": provider.Name()})
	})); err != nil {
		return err
	}
	r.server, err = r.factory(server.NewConfig(), provider.Logger(), api)
	return err
}
func (r *stubRunner) Run() error {
	if r.runError != nil {
		return r.runError
	}
	errs := make(chan error, 1)
	go func() { errs <- r.server.Serve() }()
	select {
	case err := <-errs:
		return err
	case err := <-r.fail:
		return err
	}
}
func (r *stubRunner) Terminate() {
	r.terminated = true
	if r.terminationOrder != nil {
		*r.terminationOrder = append(*r.terminationOrder, r.provider.Name())
	}
	if r.server != nil {
		_ = r.server.Shutdown()
	}
}

func testApplication(t *testing.T, runners ...*stubRunner) (*Application, application.Provider) {
	t.Helper()
	base, short, full := application.VersionBase, application.VersionShortCommit, application.VersionFullCommit
	application.VersionBase, application.VersionShortCommit, application.VersionFullCommit = "test", "short", "full"
	t.Cleanup(func() {
		application.VersionBase, application.VersionShortCommit, application.VersionFullCommit = base, short, full
	})
	t.Setenv("COMBINED_TEST_SERVER_ADDRESS", "127.0.0.1:0")
	t.Setenv("COMBINED_TEST_SERVER_TLS", "false")
	for _, name := range []string{"AUTH", "BLOB", "DATA"} {
		t.Setenv("COMBINED_TEST_"+name+"_SERVICE_EVENTS_CONSUMER_GROUP", "existing-"+name)
	}
	provider, err := application.NewNamedProvider("COMBINED_TEST", "server", "service")
	if err != nil {
		t.Fatal(err)
	}
	app := &Application{Application: application.New(), stopped: make(chan struct{})}
	for i, runner := range runners {
		app.add([]string{"auth", "blob", "data", "prescription", "task"}[i], runner)
	}
	t.Cleanup(app.Terminate)
	return app, provider
}

func runApplication(t *testing.T, app *Application) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- app.Run() }()
	t.Cleanup(func() {
		app.Terminate()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("graceful shutdown returned an error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("combined Run did not exit")
		}
	})
	return done
}

func TestAllComponentsUseOneListener(t *testing.T) {
	runners := []*stubRunner{{}, {}, {}, {}, {}}
	app, provider := testApplication(t, runners...)
	t.Setenv("COMBINED_TEST_AUTH_SERVICE_SECRET", "auth-specific")
	t.Setenv("COMBINED_TEST_BLOB_SERVICE_SECRET", "blob-specific")
	if err := app.Initialize(provider); err != nil {
		t.Fatal(err)
	}
	if got := runners[0].provider.ConfigReporter().GetWithDefault("secret", ""); got != "auth-specific" {
		t.Fatal("auth scope lost")
	}
	if got := runners[1].provider.ConfigReporter().GetWithDefault("secret", ""); got != "blob-specific" {
		t.Fatal("blob scope lost")
	}
	runApplication(t, app)
	select {
	case <-app.components[0].endpoint.ready:
	case <-time.After(5 * time.Second):
		t.Fatal("component did not start")
	}
	app.mu.Lock()
	address := app.listener.Addr().String()
	app.mu.Unlock()
	client := &http.Client{Timeout: 3 * time.Second}
	for _, name := range []string{"auth", "blob", "data", "prescription", "task"} {
		res, err := client.Get("http://" + address + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil || res.StatusCode != http.StatusOK || !strings.Contains(string(body), name) {
			t.Fatalf("%s: %s, %v", name, body, err)
		}
	}
	app.Terminate()
	for _, runner := range runners {
		if !runner.terminated {
			t.Fatal("component not terminated")
		}
	}
	if conn, err := net.DialTimeout("tcp", address, time.Second); err == nil {
		conn.Close()
		t.Fatal("shared listener is still open")
	}
}

func TestInitializationFailureCleansUpInReverseOrder(t *testing.T) {
	var order []string
	runners := []*stubRunner{{terminationOrder: &order}, {initializationError: errors.New("bad dependency"), terminationOrder: &order}, {}}
	app, provider := testApplication(t, runners...)
	if err := app.Initialize(provider); err == nil || !strings.Contains(err.Error(), "initialize blob") {
		t.Fatalf("unexpected error: %v", err)
	}
	app.Terminate()
	if strings.Join(order, ",") != "blob,auth" {
		t.Fatalf("cleanup order: %v", order)
	}
	if runners[2].provider != nil || runners[2].terminated {
		t.Fatal("uninitialized component was touched")
	}
}

func TestComponentFailureStopsApplication(t *testing.T) {
	for _, duringStartup := range []bool{true, false} {
		t.Run(map[bool]string{true: "startup", false: "running"}[duringStartup], func(t *testing.T) {
			runner := &stubRunner{fail: make(chan error, 1)}
			if duringStartup {
				runner.runError = errors.New("worker failed")
			}
			app, provider := testApplication(t, &stubRunner{}, runner)
			if err := app.Initialize(provider); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- app.Run() }()
			if !duringStartup {
				select {
				case <-app.components[1].endpoint.ready:
				case <-time.After(5 * time.Second):
					t.Fatal("not ready")
				}
				runner.fail <- errors.New("worker failed")
			}
			select {
			case err := <-done:
				if err == nil || !strings.Contains(err.Error(), "blob: worker failed") {
					t.Fatalf("failure was not propagated: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("failure was ignored")
			}
			app.Terminate()
		})
	}
}

func TestListenerFailureIsPropagated(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	app, provider := testApplication(t, &stubRunner{})
	t.Setenv("COMBINED_TEST_SERVER_ADDRESS", occupied.Addr().String())
	if err := app.Initialize(provider); err != nil {
		t.Fatal(err)
	}
	if err := app.Run(); err == nil {
		t.Fatal("binding failure was ignored")
	}
	app.Terminate()
}

func TestTerminateBeforeRun(t *testing.T) {
	app, provider := testApplication(t, &stubRunner{})
	if err := app.Initialize(provider); err != nil {
		t.Fatal(err)
	}
	app.Terminate()
	if err := app.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestConsumerGroupsMustBeExplicitAndDistinct(t *testing.T) {
	for _, group := range []string{"", "existing-AUTH"} {
		t.Run("blob="+group, func(t *testing.T) {
			runners := []*stubRunner{{}, {}}
			app, provider := testApplication(t, runners...)
			t.Setenv("COMBINED_TEST_BLOB_SERVICE_EVENTS_CONSUMER_GROUP", group)
			if err := app.Initialize(provider); err == nil || !strings.Contains(err.Error(), "consumer") {
				t.Fatalf("invalid consumer groups accepted: %v", err)
			}
			for _, runner := range runners {
				if runner.provider != nil {
					t.Fatal("components started before group validation")
				}
			}
		})
	}
}
