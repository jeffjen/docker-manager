package api

import (
	prov "github.com/jeffjen/podd/provider"

	"errors"
	"net/http"
)

var (
	AutoScaling prov.AutoScaling
)

var (
	ErrNoCandidateFound = errors.New("No candidate found")
)

type dir []http.Dir

func (d dir) Open(name string) (http.File, error) {
	for _, root := range d {
		if f, err := root.Open(name); err == nil {
			return f, nil
		}
	}
	return nil, ErrNoCandidateFound
}

func Dir(repo ...http.Dir) http.FileSystem {
	return dir(repo)
}
