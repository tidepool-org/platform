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
	r.writeRaw("text/html", statusCode, []byte(html))
}

func (r *Responder) Data(statusCode int, data interface{}) {
	if data == nil {
		r.Error(http.StatusInternalServerError, errors.ErrorInternal(errors.New("data is missing")))
	} else if err := r.sanitize(data); err != nil {
		r.Error(http.StatusInternalServerError, errors.ErrorInternal(errors.Wrap(err, "unable to sanitize data")))
	} else {
		r.writeJSON(statusCode, data)
	}
}

func (r *Responder) Error(statusCode int, err error) {
	if err == nil {
		err = errors.ErrorInternal(errors.New("error is missing"))
	}

	// service.SetRequestErrors(r.request, errs) // TODO:
	log.LoggerFromContext(r.request.Context()).WithError(err).Warn("Failure during request")

	r.writeJSON(statusCode, errors.Sanitize(err))
}

func (r *Responder) sanitize(data interface{}) error {
	if sanitizable, ok := data.(Sanitizable); ok {
		return sanitizable.Sanitize(DetailsFromContext(r.request.Context()))
	}
	return nil
}

func (r *Responder) writeJSON(statusCode int, object interface{}) {
	r.response.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.response.WriteHeader(statusCode)
	if err := r.response.WriteJson(object); err != nil {
		log.LoggerFromContext(r.request.Context()).WithError(err).Error("Unable to write JSON")
	} else if _, err = r.response.(http.ResponseWriter).Write(_NewLine); err != nil {
		log.LoggerFromContext(r.request.Context()).WithError(err).Error("Unable to write new line")
	}
}

func (r *Responder) writeRaw(contentType string, statusCode int, bytes []byte) {
	r.response.Header().Set("Content-Type", contentType)
	r.response.WriteHeader(statusCode)
	if _, err := r.response.(http.ResponseWriter).Write(bytes); err != nil {
		log.LoggerFromContext(r.request.Context()).WithError(err).Error("Unable to write bytes")
	}
}

var _NewLine = []byte("\n")
