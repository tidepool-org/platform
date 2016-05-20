package server

import "net/http"

type API interface {
	Handler() http.Handler
}
