package service

import "net/http"

type API interface {
	Handler() http.Handler
}
