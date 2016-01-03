package api

import (
	_ "github.com/Sirupsen/logrus"

	"net/http"
)

func ClusterList(w http.ResponseWriter, r *http.Request) {
	if err := common("GET", r); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, ErrNotImplemented.Error(), 403)
	return
}

func ClusterCreate(w http.ResponseWriter, r *http.Request) {
	if err := common("POST", r); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, ErrNotImplemented.Error(), 403)
	return
}
