package middleware

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/tidepool-org/platform/crypto"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/service"
)

type FieldFunc func(newFields log.Fields, value string) log.Fields

type StringFieldFuncMap map[string]FieldFunc

type Header struct {
	HeaderFieldFuncs StringFieldFuncMap
}

const (
	_HeaderValueMaximumLength = 256
)

func NewRawFieldFunc(key string) FieldFunc {
	return func(fields log.Fields, value string) log.Fields {
		if value != "" {
			if len(value) > _HeaderValueMaximumLength {
				value = value[:_HeaderValueMaximumLength]
			}
			fields[key] = value
		}
		return fields
	}
}

func NewMD5FieldFunc(key string) FieldFunc {
	return func(fields log.Fields, value string) log.Fields {
		if value != "" {
			fields[key] = crypto.HashWithMD5(value)
		}
		return fields
	}
}

func NewHeader() (*Header, error) {
	return &Header{
		HeaderFieldFuncs: StringFieldFuncMap{},
	}, nil
}

func (h *Header) AddHeaderFieldFunc(header string, fieldFunc FieldFunc) {
	h.HeaderFieldFuncs[header] = fieldFunc
}

func (h *Header) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	return func(response rest.ResponseWriter, request *rest.Request) {
		if handler != nil && response != nil && request != nil {
			oldLogger := service.GetRequestLogger(request)

			defer func() {
				service.SetRequestLogger(request, oldLogger)
			}()

			if oldLogger != nil {
				newFields := log.Fields{}

				for header, fieldFunc := range h.HeaderFieldFuncs {
					if fieldFunc != nil {
						newFields = fieldFunc(newFields, request.Header.Get(header))
					}
				}

				if len(newFields) > 0 {
					service.SetRequestLogger(request, oldLogger.WithFields(newFields))
				}
			}

			handler(response, request)
		}
	}
}
