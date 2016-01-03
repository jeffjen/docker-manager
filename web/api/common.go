package api

import (
	"errors"
	"net/http"
)

var (
	ErrMethodNotFound = errors.New("Method not Allowed")

	ErrBadRequest = errors.New("Bad Request")

	ErrNotImplemented = errors.New("Not Implemented")
)

func common(m string, r *http.Request) error {
	if r.Method != m {
		return ErrMethodNotFound
	}
	if err := r.ParseForm(); err != nil {
		return ErrBadRequest
	}
	return nil
}
