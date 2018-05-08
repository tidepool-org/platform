package service

import (
	"time"

	"github.com/ant0ine/go-json-rest/rest"
)

func AddDateHeader(response rest.ResponseWriter) {
	if response != nil {
		response.Header().Set("Date", time.Now().Format(time.RFC1123))
	}
}
