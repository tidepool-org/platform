package request

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
)

type Sanitizable interface {
	Sanitize(details Details) error
}

type Responder struct {
	res rest.ResponseWriter
	req *rest.Request
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
		res: res,
		req: req,
	}, nil
}

func (r *Responder) SetCookie(cookie *http.Cookie) {
	if cookie != nil {
		http.SetCookie(r.res.(http.ResponseWriter), cookie)
	}
}

func (r *Responder) Redirect(statusCode int, url string, mutators ...ResponseMutator) {
	if err := r.mutateResponse(mutators); err != nil {
		r.InternalServerError(err)
	} else {
		http.Redirect(r.res.(http.ResponseWriter), r.req.Request, url, statusCode)
	}
}

func (r *Responder) Empty(statusCode int, mutators ...ResponseMutator) {
	// FUTURE: rest.ResponseWriter sets unnecessary/incorrect Content-Type and Content-Encoding headers
	if err := r.mutateResponse(mutators); err != nil {
		r.InternalServerError(err)
	} else {
		r.res.WriteHeader(statusCode)
	}
}

func (r *Responder) Bytes(statusCode int, bytes []byte, mutators ...ResponseMutator) {
	if err := r.mutateResponse(mutators); err != nil {
		r.InternalServerError(err)
	} else {
		r.res.WriteHeader(statusCode)
		if bytesWritten, writeErr := r.res.(http.ResponseWriter).Write(bytes); writeErr != nil {
			log.LoggerFromContext(r.req.Context()).WithError(writeErr).Error("Unable to write bytes")
		} else if bytesLength := len(bytes); bytesWritten != bytesLength {
			log.LoggerFromContext(r.req.Context()).WithFields(log.Fields{"bytesWritten": bytesWritten, "bytesLength": bytesLength}).Error("Bytes written does not equal bytes length")
		}
	}
}

func (r *Responder) String(statusCode int, str string, mutators ...ResponseMutator) {
	r.Bytes(statusCode, []byte(str), mutators...)
}

func (r *Responder) Reader(statusCode int, reader io.Reader, mutators ...ResponseMutator) {
	if reader == nil {
		r.InternalServerError(errors.New("reader is missing"))
	} else if err := r.mutateResponse(mutators); err != nil {
		r.InternalServerError(err)
	} else {
		r.res.WriteHeader(statusCode)
		if _, err = io.Copy(r.res.(io.Writer), reader); err != nil {
			log.LoggerFromContext(r.req.Context()).WithError(err).Error("Unable to copy bytes from reader")
		}
	}
}

func (r *Responder) Data(statusCode int, data interface{}, mutators ...ResponseMutator) {
	if data == nil {
		r.InternalServerError(errors.New("data is missing"))
	} else if sanitizeErr := r.sanitize(data); sanitizeErr != nil {
		r.InternalServerError(errors.Wrap(sanitizeErr, "unable to sanitize data"))
	} else if bytes, marshalErr := json.Marshal(data); marshalErr != nil {
		r.InternalServerError(errors.Wrap(marshalErr, "unable to serialize data"))
	} else {
		r.Bytes(statusCode, append(bytes, newLine...), append(mutators, NewHeaderMutator("Content-Type", "application/json; charset=utf-8"))...)
	}
}

func (r *Responder) Error(statusCode int, err error, mutators ...ResponseMutator) {
	if err == nil {
		r.InternalServerError(errors.New("error is missing"))
	} else {
		SetErrorToContext(r.req.Context(), err)
		if bytes, marshalErr := json.Marshal(errors.Sanitize(err)); marshalErr != nil {
			r.InternalServerError(errors.Wrap(marshalErr, "unable to serialize error"))
		} else {
			r.Bytes(statusCode, append(bytes, newLine...), append(mutators, NewHeaderMutator("Content-Type", "application/json; charset=utf-8"))...)
		}
	}
}

func (r *Responder) InternalServerError(err error, mutators ...ResponseMutator) {
	if err == nil {
		err = ErrorInternalServerError(errors.New("error is missing"))
	} else if !IsErrorInternalServerError(err) {
		err = ErrorInternalServerError(err)
	}
	r.Error(http.StatusInternalServerError, err, mutators...)
}

func (r *Responder) RespondIfError(err error, mutators ...ResponseMutator) bool {
	if err == nil {
		return false
	}
	if statusCode := StatusCodeForError(err); statusCode != http.StatusInternalServerError {
		r.Error(statusCode, err, mutators...)
	} else {
		r.InternalServerError(err, mutators...)
	}
	return true
}

func (r *Responder) mutateResponse(mutators []ResponseMutator) error {
	for _, mutator := range mutators {
		if err := mutator.MutateResponse(r.res.(http.ResponseWriter)); err != nil {
			return errors.Wrap(err, "unable to mutate response")
		}
	}
	return nil
}

func (r *Responder) sanitize(data interface{}) error {
	if sanitizable, ok := data.(Sanitizable); ok {
		return sanitizable.Sanitize(DetailsFromContext(r.req.Context()))
	}
	return nil
}

var newLine = []byte("\n")
