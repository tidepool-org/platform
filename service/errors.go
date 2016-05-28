package service

import (
	"net/http"
)

func ErrorInternalServerFailure() *Error {
	return &Error{
		Code:   "internal-server-failure",
		Status: http.StatusInternalServerError,
		Title:  "internal server failure",
		Detail: "Internal server failure",
	}
}
