package combined

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"go.uber.org/fx"

	abbottAPI "github.com/tidepool-org/platform-plugin-abbott/abbott/service/api/v1"

	"github.com/tidepool-org/platform/auth"
	authService "github.com/tidepool-org/platform/auth/service/service"
	blobService "github.com/tidepool-org/platform/blob/service"
	"github.com/tidepool-org/platform/clinics"
	dataAPI "github.com/tidepool-org/platform/data/service/api"
	dataV1 "github.com/tidepool-org/platform/data/service/api/v1"
	"github.com/tidepool-org/platform/log"
	nullLog "github.com/tidepool-org/platform/log/null"
	"github.com/tidepool-org/platform/prescription"
	prescriptionApplication "github.com/tidepool-org/platform/prescription/application"
	prescriptionService "github.com/tidepool-org/platform/prescription/service"
	"github.com/tidepool-org/platform/service"
	serviceAPI "github.com/tidepool-org/platform/service/api"
	"github.com/tidepool-org/platform/status"
	taskService "github.com/tidepool-org/platform/task/service/service"
	"github.com/tidepool-org/platform/version"
)

type routeList []*rest.Route

func (r routeList) Routes() []*rest.Route { return r }

type routeTestService struct {
	service.Service
	name string
}

func (s routeTestService) Secret() string          { return s.name + "-secret" }
func (s routeTestService) Logger() log.Logger      { return nullLog.NewLogger() }
func (s routeTestService) AuthClient() auth.Client { return &unusedAuthClient{} }

// Embedding the interface makes any accidental dependency call fail the test.
// Routing probes must not need databases, tokens, Kafka, or remote services.
type unusedAuthClient struct{ auth.Client }

func registeredSources(t *testing.T) []routeSource {
	t.Helper()
	var sources []routeSource
	add := func(name string, routers []service.Router, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		var routes routeList
		for _, router := range routers {
			routes = append(routes, router.Routes()...)
		}
		sources = append(sources, routeSource{name, routes})
	}
	authRouters, err := authService.New().Routers()
	add("auth", authRouters, err)
	blobRouters, err := blobService.New().Routers()
	add("blob", blobRouters, err)

	base, err := serviceAPI.New(routeTestService{name: "data"})
	if err != nil {
		t.Fatal(err)
	}
	data := &dataAPI.Standard{API: base}
	if err := data.DEPRECATEDInitializeRouter(dataV1.Routes()); err != nil {
		t.Fatal(err)
	}
	sources = append(sources, routeSource{"data", routeList(data.Routes())})

	type prescriptionRouters struct {
		fx.In
		Routers []service.Router `group:"routers"`
	}
	var prescriptionRoutes prescriptionRouters
	graph := fx.New(
		fx.NopLogger, prescriptionApplication.Routers,
		fx.Provide(
			func() clinics.Client { return nil },
			func() prescriptionService.DeviceSettingsValidator { return nil },
			func() prescription.Service { return nil },
			func() version.Reporter { return nil },
			func() status.StoreStatusReporter { return nil },
		),
		fx.Populate(&prescriptionRoutes),
	)
	add("prescription", prescriptionRoutes.Routers, graph.Err())
	taskRouters, err := taskService.New().Routers()
	add("task", taskRouters, err)
	return sources
}

func routeKey(name string, route *rest.Route) string {
	return name + " " + route.HttpMethod + " " + route.PathExp
}

func TestRouteInventory(t *testing.T) {
	sources := registeredSources(t)
	app := New()
	var registered, hosted []string
	for _, source := range sources {
		registered = append(registered, source.name)
	}
	for _, component := range app.components {
		hosted = append(hosted, component.name)
	}
	if !reflect.DeepEqual(registered, hosted) {
		t.Fatalf("route coverage does not match the components started by the binary: %v vs %v", registered, hosted)
	}
	var actual []string
	for _, source := range sources {
		for _, route := range source.router.Routes() {
			actual = append(actual, routeKey(source.name, route))
		}
	}
	sort.Strings(actual)
	content, err := os.ReadFile("testdata/routes.txt")
	if err != nil {
		t.Fatalf("%v\nROUTE INVENTORY\n%s\nEND INVENTORY", err, strings.Join(actual, "\n"))
	}
	expected := strings.Split(strings.TrimSpace(string(content)), "\n")
	// Private plugin routes are tested too, without checking private API details
	// into the public repository. The public plugin has no additional routes.
	for _, route := range abbottAPI.Routes() {
		expected = append(expected, "data "+route.Method+" "+route.Path)
	}
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("API inventory changed; review testdata/routes.txt\nwant: %v\ngot: %v", expected, actual)
	}
}

func preparedSources(t *testing.T, originals []routeSource, probe bool) []routeSource {
	t.Helper()
	var sources []routeSource
	for _, source := range originals {
		api, err := serviceAPI.New(routeTestService{name: source.name})
		if err != nil {
			t.Fatal(err)
		}
		var routes []*rest.Route
		for _, original := range source.router.Routes() {
			route := *original
			if probe {
				key := routeKey(source.name, original)
				route.Func = func(res rest.ResponseWriter, req *rest.Request) {
					body, err := io.ReadAll(req.Body)
					if err != nil {
						t.Error(err)
					}
					res.Header().Set("X-Reached-Route", key)
					res.WriteJson(map[string]interface{}{"params": req.PathParams, "query": req.URL.Query().Get("probe"), "body": string(body)})
				}
			}
			routes = append(routes, &route)
		}
		if err := api.InitializeRoutes(routes...); err != nil {
			t.Fatal(err)
		}
		if err := api.InitializeMiddleware(); err != nil {
			t.Fatal(err)
		}
		sources = append(sources, routeSource{source.name, api})
	}
	return sources
}

var pathParameter = regexp.MustCompile(`[:#*]([A-Za-z][A-Za-z0-9_]*)`)

func TestEveryRouteIsReachable(t *testing.T) {
	originals := registeredSources(t)
	handler, err := newHandler("test-version", preparedSources(t, originals, true))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	for _, source := range originals {
		for _, route := range source.router.Routes() {
			t.Run(routeKey(source.name, route), func(t *testing.T) {
				params := map[string]string{}
				for _, match := range pathParameter.FindAllStringSubmatch(route.PathExp, -1) {
					params[match[1]] = "test-" + match[1]
				}
				path := route.MakePath(params)
				if path == "/status" {
					path += "/" + source.name
				}
				owner := source.name
				if path == "/v1/metrics" {
					owner = "auth"
				}
				req, err := http.NewRequest(route.HttpMethod, server.URL+path+"?probe=kept", strings.NewReader("request-body"))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set(auth.TidepoolServiceSecretHeaderKey, owner+"-secret")
				req.Header.Set("Accept-Encoding", "gzip")
				res, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer res.Body.Close()
				if res.StatusCode != http.StatusOK || res.Header.Get("X-Reached-Route") != routeKey(owner, route) {
					t.Fatalf("route not reached: status %d, selected %q", res.StatusCode, res.Header.Get("X-Reached-Route"))
				}
				if res.Header.Get("Content-Encoding") != "gzip" {
					t.Fatal("component gzip middleware was lost")
				}
				reader, err := gzip.NewReader(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				defer reader.Close()
				var payload struct {
					Params      map[string]string
					Query, Body string
				}
				if err := json.NewDecoder(reader).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(payload.Params, params) || payload.Query != "kept" || payload.Body != "request-body" {
					t.Fatalf("request changed in dispatch: %+v, expected params %v", payload, params)
				}
			})
		}
	}
}

func TestCombinedAuthenticationAndHealth(t *testing.T) {
	handler, err := newHandler("test-version", preparedSources(t, registeredSources(t), false))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"/v1/consents", "/v1/device_logs/test/content", "/v1/users/test/datasets",
		"/v1/clinics/test/prescriptions", "/v1/tasks",
	} {
		t.Run(path, func(t *testing.T) {
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
			if res.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", res.Code, res.Body)
			}
		})
	}
	for _, source := range registeredSources(t) {
		for _, route := range source.router.Routes() {
			path := pathParameter.ReplaceAllString(route.PathExp, "test")
			if path == "/status" {
				path += "/" + source.name
			}
			req := httptest.NewRequest(route.HttpMethod, path, nil)
			req.Header.Set(auth.TidepoolServiceSecretHeaderKey, "wrong-secret")
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != http.StatusForbidden {
				t.Errorf("%s: invalid credentials returned %d", routeKey(source.name, route), res.Code)
			}
		}
	}
	for path, expected := range map[string]int{"/status": 200, "/v1/metrics": 200, "/not-an-api": 404, "/v1/tasks/test/unknown": 404} {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Code != expected {
			t.Errorf("%s: got %d, want %d", path, res.Code, expected)
		}
		if path == "/status" && !strings.Contains(res.Body.String(), `"prescription"`) {
			t.Error("health is missing a component")
		}
		if path == "/v1/metrics" && !strings.Contains(res.Body.String(), "go_goroutines") {
			t.Error("combined metrics not served")
		}
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPatch, "/v1/tasks", nil))
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unknown method got %d", res.Code)
	}
}

func TestRouteCollisionsFail(t *testing.T) {
	noop := func(rest.ResponseWriter, *rest.Request) {}
	for _, paths := range [][2]string{{"/same", "/same"}, {"/users/:id", "/users/:userId"}, {"/v1/metrics", "/v1/metrics"}} {
		t.Run(fmt.Sprint(paths), func(t *testing.T) {
			_, err := newHandler("test", []routeSource{
				{"one", routeList{rest.Get(paths[0], noop)}},
				{"two", routeList{rest.Get(paths[1], noop)}},
			})
			if err == nil {
				t.Fatal("conflicting routes were accepted")
			}
		})
	}
}
