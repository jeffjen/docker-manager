package api

import (
	"errors"
	"net/http"
)

var (
	ErrNoCandidateFound = errors.New("No candidate found")
)

type Dir []http.Dir

func (d Dir) Open(name string) (http.File, error) {
	for _, root := range d {
		f, err := root.Open(name)
		if err == nil {
			return f, err
		}
	}
	return nil, ErrNoCandidateFound
}
