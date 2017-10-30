package request

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type Sanitizable interface {
	Sanitize(details Details) error
}

type Responder struct {
	response rest.ResponseWriter
	request  *rest.Request
}

func MustNewResponder(res rest.ResponseWriter, req *rest.Request) *Responder {
	responder, err := NewResponder(res, req)
	if err != nil {
		panic(err)
	}

	return responder
}

func NewResponder(res rest.ResponseWriter, req *rest.Request) (*Responder, error) {
	if res == nil {
		return nil, errors.New("response is missing")
	}
	if req == nil {
		return nil, errors.New("request is missing")
	}

	return &Responder{
		response: res,
		request:  req,
	}, nil
}

func (r *Responder) SetCookie(cookie *http.Cookie) {
	if cookie != nil {
		http.SetCookie(r.response.(http.ResponseWriter), cookie)
	}
}

func (r *Responder) Redirect(statusCode int, url string) {
	http.Redirect(r.response.(http.ResponseWriter), r.request.Request, url, statusCode)
}

func (r *Responder) Empty(statusCode int) {
	r.response.WriteHeader(statusCode)
}

func (r *Responder) HTML(statusCode int, html string) {
	r.response.Header().Set("Content-Type", "text/html")
	r.response.WriteHeader(statusCode)
	r.response.(http.ResponseWriter).Write([]byte(html))
}

func (r *Responder) Data(statusCode int, data interface{}) {
	if data == nil {
		r.Error(http.StatusInternalServerError, errors.ErrorInternal(errors.New("data is missing")))
	} else if err := r.sanitize(data); err != nil {
		r.Error(http.StatusInternalServerError, errors.ErrorInternal(errors.Wrap(err, "unable to sanitize data")))
	} else {
		r.response.WriteHeader(statusCode)
		r.response.WriteJson(data)
	}
}

func (r *Responder) Error(statusCode int, err error) {
	if err == nil {
		err = errors.ErrorInternal(errors.New("error is missing"))
	}

	// service.SetRequestErrors(r.request, errs) // TODO:
	log.LoggerFromContext(r.request.Context()).WithError(err).Warn("Failure during request")

	r.response.WriteHeader(statusCode)
	r.response.WriteJson(errors.Sanitize(err))
}

func (r *Responder) sanitize(data interface{}) error {
	if sanitizable, ok := data.(Sanitizable); ok {
		return sanitizable.Sanitize(DetailsFromContext(r.request.Context()))
	}
	return nil
}
