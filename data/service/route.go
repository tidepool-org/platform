package service

type Route struct {
	Method  string
	Path    string
	Handler HandlerFunc
}

func MakeRoute(method string, path string, handler HandlerFunc) Route {
	return Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}
}
