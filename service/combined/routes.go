package combined

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/service"
)

type routeSource struct {
	name   string
	router service.Router
}

// newHandler mounts the actual component routes and their middleware in one
// router. It does not proxy requests or maintain a second list of API prefixes.
func newHandler(version string, sources []routeSource) (http.Handler, error) {
	var routes []*rest.Route
	owners := map[string]string{}
	names := make([]string, 0, len(sources))
	seenNames := map[string]bool{}
	for _, source := range sources {
		if source.name == "" || seenNames[source.name] || source.router == nil {
			return nil, fmt.Errorf("invalid or duplicate route source %q", source.name)
		}
		seenNames[source.name] = true
		names = append(names, source.name)
		for _, original := range source.router.Routes() {
			if original == nil || original.Func == nil {
				return nil, fmt.Errorf("%s has an invalid route", source.name)
			}
			route := *original
			route.HttpMethod = strings.ToUpper(route.HttpMethod)
			if route.HttpMethod == http.MethodGet && route.PathExp == "/status" {
				route.PathExp = "/status/" + source.name
			}
			normalizeParameters(&route)
			key := route.HttpMethod + " " + route.PathExp
			if owner, found := owners[key]; found {
				// All three metrics handlers use prometheus.DefaultGatherer,
				// which now contains the metrics from the entire process.
				if key == "GET /v1/metrics" && sharedMetrics(owner) && sharedMetrics(source.name) && owner != source.name {
					continue
				}
				return nil, fmt.Errorf("route %s is registered by both %s and %s", key, owner, source.name)
			}
			owners[key] = source.name
			routes = append(routes, &route)
		}
	}

	routes = append(routes, rest.Get("/status", func(res rest.ResponseWriter, req *rest.Request) {
		_ = res.WriteJson(struct {
			Version  string   `json:"version"`
			Services []string `json:"services"`
		}{version, names})
	}))
	router, err := rest.MakeRouter(routes...)
	if err != nil {
		return nil, fmt.Errorf("unable to combine API routes: %w", err)
	}
	api := rest.NewApi()
	api.SetApp(router)
	return api.MakeHandler(), nil
}

func sharedMetrics(name string) bool {
	return name == "auth" || name == "data" || name == "task"
}

var parameterPattern = regexp.MustCompile(`[:#*]([A-Za-z][A-Za-z0-9_]*)`)

// go-json-rest requires shared path placeholders to have identical names.
// For example auth uses :id where data uses :providerSessionId. Normalize the
// routing index, then restore each handler's own names before its middleware.
func normalizeParameters(route *rest.Route) {
	var names []string
	route.PathExp = parameterPattern.ReplaceAllStringFunc(route.PathExp, func(match string) string {
		name := "p" + strconv.Itoa(len(names))
		names = append(names, match[1:])
		return match[:1] + name
	})
	if len(names) == 0 {
		return
	}
	handler := route.Func
	route.Func = func(res rest.ResponseWriter, req *rest.Request) {
		original := req.PathParams
		defer func() { req.PathParams = original }()
		req.PathParams = make(map[string]string, len(names))
		for i, name := range names {
			req.PathParams[name] = original["p"+strconv.Itoa(i)]
		}
		handler(res, req)
	}
}
